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

package polling

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPoll_Success(t *testing.T) {
	callCount := 0
	result, err := Poll(context.Background(), Config{
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		Timeout:         1 * time.Second,
	}, func(ctx context.Context) Result {
		callCount++
		if callCount >= 3 {
			return Result{Done: true, Data: "success"}
		}
		return Result{Done: false}
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 3, callCount)
}

func TestPoll_Timeout(t *testing.T) {
	callCount := 0
	_, err := Poll(context.Background(), Config{
		InitialInterval: 50 * time.Millisecond,
		MaxInterval:     100 * time.Millisecond,
		Timeout:         150 * time.Millisecond,
	}, func(ctx context.Context) Result {
		callCount++
		return Result{Done: false} // Never succeeds
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "polling timeout")
	assert.GreaterOrEqual(t, callCount, 1)
}

func TestPoll_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	callCount := 0

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := Poll(ctx, Config{
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     100 * time.Millisecond,
		Timeout:         0, // No timeout
	}, func(ctx context.Context) Result {
		callCount++
		return Result{Done: false}
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
}

func TestPoll_PermanentError(t *testing.T) {
	expectedErr := errors.New("permanent error")
	result, err := Poll(context.Background(), Config{
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		Timeout:         1 * time.Second,
	}, func(ctx context.Context) Result {
		return Result{Done: true, Err: expectedErr}
	})

	assert.ErrorIs(t, err, expectedErr)
	assert.Nil(t, result)
}

func TestPoll_ExponentialBackoff(t *testing.T) {
	timestamps := []time.Time{}
	callCount := 0

	Poll(context.Background(), Config{
		InitialInterval: 20 * time.Millisecond,
		MaxInterval:     100 * time.Millisecond,
		Timeout:         500 * time.Millisecond,
		Multiplier:      2.0,
	}, func(ctx context.Context) Result {
		timestamps = append(timestamps, time.Now())
		callCount++
		if callCount >= 5 {
			return Result{Done: true, Data: "done"}
		}
		return Result{Done: false}
	})

	// Verify intervals are increasing (with some tolerance for timing)
	if len(timestamps) >= 3 {
		interval1 := timestamps[1].Sub(timestamps[0])
		interval2 := timestamps[2].Sub(timestamps[1])
		// Second interval should be roughly double the first (with tolerance)
		assert.Greater(t, interval2, interval1)
	}
}

func TestPoll_DefaultMultiplier(t *testing.T) {
	callCount := 0
	result, err := Poll(context.Background(), Config{
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		Timeout:         1 * time.Second,
		Multiplier:      0, // Should use default 2.0
	}, func(ctx context.Context) Result {
		callCount++
		if callCount >= 2 {
			return Result{Done: true, Data: "success"}
		}
		return Result{Done: false}
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 500*time.Millisecond, cfg.InitialInterval)
	assert.Equal(t, 5*time.Second, cfg.MaxInterval)
	assert.Equal(t, 2*time.Minute, cfg.Timeout)
	assert.Equal(t, 2.0, cfg.Multiplier)
}
