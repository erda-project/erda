// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clickhouse

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
)

type testBatchBuilder func(context.Context, interface{}) ([]driver.Batch, error)

func (f testBatchBuilder) BuildBatch(ctx context.Context, items interface{}) ([]driver.Batch, error) {
	return f(ctx, items)
}

type testBatch struct {
	driver.Batch
	send     func() error
	sent     bool
	aborted  bool
	attempts int
}

func (b *testBatch) Send() error {
	b.attempts++
	b.sent = true
	if b.send != nil {
		return b.send()
	}
	return nil
}

func (b *testBatch) IsSent() bool { return b.sent }
func (b *testBatch) Rows() int    { return 1 }
func (b *testBatch) Abort() error {
	b.aborted, b.sent = true, true
	return nil
}

func waitStorageTest(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for storage")
	}
}

func TestStorageCloseDrainsConcurrentWrites(t *testing.T) {
	const workers = 2
	entered := make(chan struct{}, workers)
	release := make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	var accepted, sent, active atomic.Int32
	st := &Storage{
		cfg:    &StorageConfig{CurrencyNum: workers, RetryNum: 1},
		logger: logrusx.New(),
		sqlBuilder: testBatchBuilder(func(context.Context, interface{}) ([]driver.Batch, error) {
			return []driver.Batch{&testBatch{send: func() error {
				if n := active.Add(1); n > workers {
					t.Errorf("sending concurrency %d exceeds %d", n, workers)
				}
				defer active.Add(-1)
				select {
				case entered <- struct{}{}:
				default:
				}
				<-release
				sent.Add(1)
				return nil
			}}}, nil
		}),
	}
	require.NoError(t, st.Start(context.Background()))
	// Always release blocked sends before cleanup, including on a failed assertion.
	t.Cleanup(func() {
		unblock()
		require.NoError(t, st.Close())
	})
	for i := 0; i < workers; i++ {
		require.NoError(t, st.WriteBatchAsync(i))
		accepted.Add(1)
	}
	for i := 0; i < workers; i++ {
		select {
		case <-entered:
		case <-time.After(5 * time.Second):
			t.Fatal("workers did not start sending")
		}
	}
	for i := 0; i < cap(st.batchCh); i++ {
		require.NoError(t, st.WriteBatchAsync(i))
		accepted.Add(1)
	}
	var writers sync.WaitGroup
	for i := 0; i < 64; i++ {
		writers.Add(1)
		go func() {
			defer writers.Done()
			err := st.WriteBatchAsync(nil)
			if err == nil {
				accepted.Add(1)
			} else if !errors.Is(err, errStorageClosed) {
				t.Errorf("unexpected write error: %v", err)
			}
		}()
	}
	closed := make(chan struct{})
	go func() {
		defer close(closed)
		if err := st.Close(); err != nil {
			t.Errorf("Close failed: %v", err)
		}
	}()
	writesDone := make(chan struct{})
	go func() {
		writers.Wait()
		close(writesDone)
	}()
	waitStorageTest(t, writesDone)
	require.ErrorIs(t, st.WriteBatchAsync(nil), errStorageClosed)
	select {
	case <-closed:
		t.Fatal("Close returned while sends were blocked")
	default:
	}
	unblock()
	waitStorageTest(t, closed)
	require.Equal(t, accepted.Load(), sent.Load())
	require.NoError(t, st.Close())
}

func TestStorageSendRetriesSentBatch(t *testing.T) {
	b := &testBatch{}
	b.send = func() error {
		if b.attempts == 1 {
			return errors.New("temporary send failure")
		}
		return nil
	}
	st := &Storage{
		cfg: &StorageConfig{RetryNum: 2}, logger: logrusx.New(), ctx: context.Background(),
	}
	st.sendBatch(b)
	require.Equal(t, 2, b.attempts)
	require.False(t, b.aborted)
}

func TestStorageSendExhaustionDoesNotBackoff(t *testing.T) {
	b := &testBatch{send: func() error { return errors.New("send failure") }}
	st := &Storage{
		cfg: &StorageConfig{RetryNum: 1}, logger: logrusx.New(), ctx: context.Background(),
	}
	done := make(chan struct{})
	go func() {
		st.sendBatch(b)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("last failed attempt must not wait for another retry")
	}
	require.Equal(t, 1, b.attempts)
}

func TestStorageSendCancellation(t *testing.T) {
	for _, beforeSend := range []bool{true, false} {
		t.Run(map[bool]string{true: "before send", false: "during backoff"}[beforeSend], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			b := &testBatch{send: func() error {
				cancel()
				return errors.New("send failure")
			}}
			st := &Storage{cfg: &StorageConfig{RetryNum: 5}, logger: logrusx.New(), ctx: ctx}
			if beforeSend {
				cancel()
			}
			st.sendBatch(b)
			if beforeSend {
				require.Zero(t, b.attempts)
				require.True(t, b.aborted)
			} else {
				require.Equal(t, 1, b.attempts)
			}
		})
	}
}

func TestStorageContinuesAfterBuildError(t *testing.T) {
	var builds int
	b := &testBatch{}
	st := &Storage{
		cfg: &StorageConfig{CurrencyNum: 1, RetryNum: 1}, logger: logrusx.New(),
		sqlBuilder: testBatchBuilder(func(context.Context, interface{}) ([]driver.Batch, error) {
			builds++
			if builds == 1 {
				return nil, errors.New("invalid batch")
			}
			return []driver.Batch{b}, nil
		}),
	}
	require.NoError(t, st.Start(context.Background()))
	t.Cleanup(func() { require.NoError(t, st.Close()) })
	require.NoError(t, st.WriteBatchAsync(nil))
	require.NoError(t, st.WriteBatchAsync(nil))
	require.NoError(t, st.Close())
	require.Equal(t, 2, builds)
	require.Equal(t, 1, b.attempts)
}

func TestStorageRejectsCanceledWrites(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	st := &Storage{cfg: &StorageConfig{CurrencyNum: 1, RetryNum: 1}}
	require.NoError(t, st.Start(ctx))
	cancel()
	require.ErrorIs(t, st.WriteBatchAsync(nil), context.Canceled)
	require.NoError(t, st.Close())
}

func TestStorageInvalidConfig(t *testing.T) {
	for _, cfg := range []StorageConfig{
		{CurrencyNum: 0, RetryNum: 1},
		{CurrencyNum: -1, RetryNum: 1},
		{CurrencyNum: 1, RetryNum: 0},
		{CurrencyNum: 1, RetryNum: -1},
	} {
		st := &Storage{cfg: &cfg}
		require.Error(t, st.Start(context.Background()))
		require.NoError(t, st.Close())
	}
}

func TestProviderReturnsStorageErrors(t *testing.T) {
	st := &Storage{cfg: &StorageConfig{CurrencyNum: 1, RetryNum: 1}}
	require.NoError(t, st.Start(context.Background()))
	require.NoError(t, st.Close())
	p := &provider{storage: st}
	require.ErrorIs(t, p.ExportLog(), errStorageClosed)
	require.ErrorIs(t, p.ExportMetric(), errStorageClosed)
	require.ErrorIs(t, p.ExportSpan(), errStorageClosed)
	require.ErrorIs(t, p.ExportProfile(), errStorageClosed)
}
