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

package span

import (
	"context"
	"fmt"
	"hash/fnv"
	"sort"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/erda-project/erda/modules/oap/collector/lib"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/modules/msp/apm/trace"
	"github.com/erda-project/erda/pkg/strutil"
)

type Config struct {
	Database string
	Retry    int
	Logger   logs.Logger
	Client   clickhouse.Conn
}

type WriteSpan struct {
	db          string
	retryCnt    int
	logger      logs.Logger
	client      clickhouse.Conn
	seriesBatch driver.Batch

	itemsCh chan []*trace.Span
}

func NewWriteSpan(cfg Config) *WriteSpan {
	return &WriteSpan{
		db:       cfg.Database,
		retryCnt: cfg.Retry,
		logger:   cfg.Logger,
		client:   cfg.Client,
	}
}

func (ws *WriteSpan) Start(ctx context.Context) {
	maxprocs := lib.AvailableCPUs()
	ws.itemsCh = make(chan []*trace.Span, maxprocs)
	for i := 0; i < maxprocs; i++ {
		go ws.handleSpan(ctx)
	}
}

func (ws *WriteSpan) handleSpan(ctx context.Context) {
begin:
	for items := range ws.itemsCh {
		// nolint
		metaBatch, err := ws.client.PrepareBatch(ctx, "INSERT INTO "+ws.db+"."+trace.CH_TABLE_META)
		if err != nil {
			ws.logger.Errorf("prepare metaBatch: %s", err)
			continue
		}
		// nolint
		seriesBatch, err := ws.client.PrepareBatch(ctx, "INSERT INTO "+ws.db+"."+trace.CH_TABLE_SERIES)
		if err != nil {
			ws.logger.Errorf("prepare seriesBatch: %s", err)
			continue
		}

		for _, data := range items {
			metas := getMetaBuf()
			for k, v := range data.Tags {
				metas = append(metas, trace.Meta{Key: k, Value: v, OrgName: data.OrgName, CreateAt: data.EndTime})
			}
			sort.Slice(metas, func(i, j int) bool {
				return metas[i].Key < metas[j].Key
			})
			sid := hashTagsList(metas)

			if !seriesIDExisted(sid) {
				for i := range metas {
					metas[i].SeriesID = sid
					err := metaBatch.AppendStruct(&metas[i])
					if err != nil { // TODO. Data may lost when encounter error
						ws.logger.Errorf("metaBatch append: %s", err)
						_ = metaBatch.Abort()
						putMetaBuf(metas)
						goto begin
					}
				}
			}
			putMetaBuf(metas)

			series := getSeriesBuf()
			series.OrgName = data.OrgName
			series.TraceId = data.TraceId
			series.SpanId = data.SpanId
			series.ParentSpanId = data.ParentSpanId
			series.StartTime = data.StartTime
			series.EndTime = data.EndTime
			series.SeriesID = sid
			err = seriesBatch.AppendStruct(&series)
			if err != nil { // TODO. Data may lost when encounter error
				ws.logger.Errorf("seriesBatch: %s", err)
				_ = seriesBatch.Abort()
				putSeriesBuf(series)
				goto begin
			}
			putSeriesBuf(series)
		}

		ws.sendBatch(ctx, metaBatch)
		ws.sendBatch(ctx, seriesBatch)
	}
}

func (ws *WriteSpan) AddBatch(items []*trace.Span) {
	ws.itemsCh <- items
}

func (ws *WriteSpan) sendBatch(ctx context.Context, b driver.Batch) {
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

	currencyLimiter <- struct{}{}
	go func() {
		for i := 0; i < ws.retryCnt; i++ {
			if err := b.Send(); err != nil {
				ws.logger.Errorf("send bactch to clickhouse: %s, retry after: %s\n", err, backoffDelay)
				backoffSleep()
				continue
			} else {
				break
			}
		}
		<-currencyLimiter
	}()
}

func hashTagsList(list []trace.Meta) uint64 {
	h := fnv.New64a()
	for _, item := range list {
		h.Write(strutil.NoCopyStringToBytes(item.Key))
		h.Write(strutil.NoCopyStringToBytes("\n"))
		h.Write(strutil.NoCopyStringToBytes(item.Value))
		h.Write(strutil.NoCopyStringToBytes("\n"))
	}
	return h.Sum64()
}

var (
	metaBufPool, seriesBufPool sync.Pool
)

func getMetaBuf() []trace.Meta {
	v := metaBufPool.Get()
	if v == nil {
		return []trace.Meta{}
	}
	return v.([]trace.Meta)
}

func putMetaBuf(buf []trace.Meta) {
	buf = buf[:0]
	metaBufPool.Put(buf)
}

func getSeriesBuf() trace.Series {
	v := seriesBufPool.Get()
	if v == nil {
		return trace.Series{}
	}
	return v.(trace.Series)
}

func putSeriesBuf(buf trace.Series) {
	buf.SeriesID = 0
	buf.EndTime = 0
	buf.StartTime = 0
	buf.OrgName = ""
	buf.TraceId = ""
	buf.SpanId = ""
	buf.ParentSpanId = ""
	seriesBufPool.Put(buf)
}

// currency limiter
var currencyLimiter chan struct{}

func InitCurrencyLimiter(currency int) {
	currencyLimiter = make(chan struct{}, currency)
}

// seriesIDMap
var (
	seriesIDMap map[uint64]struct{} // TODO. Need refresh with certain interval
	simMutex    *sync.RWMutex
)

func seriesIDExisted(sid uint64) bool {
	simMutex.RLock()
	_, ok := seriesIDMap[sid]
	simMutex.RUnlock()
	if ok {
		return true
	}

	simMutex.Lock()
	seriesIDMap[sid] = struct{}{}
	simMutex.Unlock()
	return false
}

func InitSeriesIDMap(client clickhouse.Conn, db string) error {
	simMutex = &sync.RWMutex{}

retry:
	// nolint
	rows, err := client.Query(context.TODO(), "select distinct(series_id) from "+db+".spans_meta_all")
	if err != nil {
		if exp, ok := err.(*clickhouse.Exception); ok && validExp(exp) {
			time.Sleep(time.Second)
			goto retry
		} else {
			return err
		}
	}

	seriesIDMap = make(map[uint64]struct{}, 50000)
	for rows.Next() {
		var sid uint64
		if err := rows.Scan(&sid); err != nil {
			return fmt.Errorf("scan: %w", err)
		}
		seriesIDMap[sid] = struct{}{}
	}
	return rows.Close()
}

func validExp(exp *clickhouse.Exception) bool {
	switch exp.Code {
	case 81, 60: // db or table not existed
		return true
	}
	return false
}
