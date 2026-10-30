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
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib"
)

type BatchBuilder interface {
	BuildBatch(ctx context.Context, sourceBatch interface{}) ([]driver.Batch, error)
}

type StorageConfig struct {
	CurrencyNum int `file:"currency_num" default:"3"`
	RetryNum    int `file:"retry_num" default:"5"`
}

var errStorageClosed = errors.New("clickhouse storage closed")

type Storage struct {
	logger     logs.Logger
	cfg        *StorageConfig
	batchCh    chan interface{}
	sqlBuilder BatchBuilder
	ctx        context.Context
	done       chan struct{}
	workerWg   sync.WaitGroup
	closeOnce  sync.Once
	mu         sync.RWMutex
}

func (st *Storage) Start(ctx context.Context) error {
	if st.cfg.CurrencyNum < 1 || st.cfg.RetryNum < 1 {
		return fmt.Errorf("clickhouse currency_num and retry_num must be positive")
	}
	st.ctx = ctx
	st.done = make(chan struct{})
	st.batchCh = make(chan interface{}, min(lib.AvailableCPUs(), st.cfg.CurrencyNum))
	for i := 0; i < st.cfg.CurrencyNum; i++ {
		st.workerWg.Add(1)
		go st.handleBatch()
	}
	return nil
}

func (st *Storage) WriteBatchAsync(batch interface{}) error {
	st.mu.RLock()
	defer st.mu.RUnlock()
	select {
	case <-st.done:
		return errStorageClosed
	default:
	}
	if err := st.ctx.Err(); err != nil {
		return err
	}
	select {
	case st.batchCh <- batch:
		return nil
	case <-st.done:
		return errStorageClosed
	case <-st.ctx.Done():
		return st.ctx.Err()
	}
}

func (st *Storage) handleBatch() {
	defer st.workerWg.Done()
	for {
		select {
		case <-st.ctx.Done():
			return
		case items, ok := <-st.batchCh:
			if !ok {
				return
			}
			batches, err := st.sqlBuilder.BuildBatch(st.ctx, items)
			if err != nil {
				st.logger.Errorf("construct batch: %s", err)
				continue
			}
			for _, batch := range batches {
				st.sendBatch(batch)
			}
		}
	}
}

func (st *Storage) sendBatch(b driver.Batch) {
	if b.IsSent() {
		return
	}
	start := time.Now()
	backoffDelay := time.Second
	const maxBackoffDelay = 30 * time.Second
	for i := 0; i < st.cfg.RetryNum; i++ {
		if st.ctx.Err() != nil {
			st.abortBatch(b)
			return
		}
		// Native batches also set IsSent on failure; retry according to Send's error.
		if err := b.Send(); err != nil {
			st.logger.Warnf("clickhouse send failed attempt=%d rows=%d err=%s", i+1, b.Rows(), err)
		} else {
			return
		}
		if i+1 == st.cfg.RetryNum {
			break
		}
		if !st.waitWithContext(backoffDelay + time.Duration(rand.Int63n(int64(backoffDelay/2)))) {
			st.abortBatch(b)
			return
		}
		backoffDelay = min(backoffDelay*2, maxBackoffDelay)
	}
	st.logger.Warnf("clickhouse send exhausted rows=%d dur=%s attempts=%d", b.Rows(), time.Since(start), st.cfg.RetryNum)
	st.abortBatch(b)
}

func (st *Storage) Close() error {
	st.closeOnce.Do(func() {
		if st.done == nil {
			return
		}
		// Unblock queued writers before taking the lock that protects the channel.
		close(st.done)
		st.mu.Lock()
		close(st.batchCh)
		st.mu.Unlock()
		st.workerWg.Wait()
	})
	return nil
}

func (st *Storage) waitWithContext(d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-st.ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (st *Storage) abortBatch(b driver.Batch) {
	if !b.IsSent() {
		if err := b.Abort(); err != nil {
			st.logger.Errorf("abort batch: %s", err)
		}
	}
}
