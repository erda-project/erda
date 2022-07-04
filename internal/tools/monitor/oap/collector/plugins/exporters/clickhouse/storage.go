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
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib"
)

type BatchBuilder interface {
	BuildBatch(ctx context.Context, sourceBatch interface{}) (batches []driver.Batch, err error)
}

type StorageConfig struct {
	CurrencyNum int `file:"currency_num" default:"3"`
	RetryNum    int `file:"retry_num" default:"5"`
}

type Storage struct {
	logger          logs.Logger
	cfg             *StorageConfig
	currencyLimiter chan struct{}
	batchCh         chan interface{}
	sqlBuilder      BatchBuilder
}

func (st *Storage) Start(ctx context.Context) error {
	maxprocs := lib.AvailableCPUs()
	st.batchCh = make(chan interface{}, maxprocs)
	for i := 0; i < maxprocs; i++ {
		go st.handleBatch(ctx)
	}

	return nil
}

func (st *Storage) WriteBatchAsync(batch interface{}) {
	st.batchCh <- batch
}

func (st *Storage) handleBatch(ctx context.Context) {
	// TODO. st.batchCh close handle
	for items := range st.batchCh {
		batches, err := st.sqlBuilder.BuildBatch(ctx, items)
		if err != nil {
			st.logger.Errorf("construct batch: %s", err)
			continue
		}
		for i := range batches {
			st.sendBatch(ctx, batches[i])
		}
	}
}

func (st *Storage) sendBatch(ctx context.Context, b driver.Batch) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	backoffDelay := time.Second
	maxBackoffDelay := 30 * time.Second
	backoffSleep := func() {
		time.Sleep(backoffDelay)
		backoffDelay *= 2
		if backoffDelay > maxBackoffDelay {
			backoffDelay = maxBackoffDelay
		}
	}

	st.currencyLimiter <- struct{}{}
	go func() {
		for i := 0; i < st.cfg.RetryNum; i++ {
			if err := b.Send(); err != nil {
				st.logger.Errorf("send bactch to clickhouse: %s, retry after: %s", err, backoffDelay)
				backoffSleep()
				continue
			} else {
				break
			}
		}
		<-st.currencyLimiter
	}()
}
