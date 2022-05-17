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
	db       string
	retryCnt int
	logger   logs.Logger
	client   clickhouse.Conn
}

func NewWriteSpan(cfg Config) *WriteSpan {
	return &WriteSpan{
		db:       cfg.Database,
		retryCnt: cfg.Retry,
		logger:   cfg.Logger,
		client:   cfg.Client,
	}
}

func (ws *WriteSpan) WriteAll(ctx context.Context, items []*trace.Span) error {
	metaBuf, seriesBuf := getMetaBuf(), getseriesBuf()
	defer func() {
		putMetaBuf(metaBuf)
		putseriesBuf(seriesBuf)
	}()

	for i := range items {
		span := items[i]
		if v, ok := span.Tags[trace.OrgNameKey]; ok {
			span.OrgName = v
		} else {
			return fmt.Errorf("must have %q", trace.OrgNameKey)
		}
		series, metas, sid := splitSpan(span)
		seriesBuf = append(seriesBuf, series)

		if !seriesIDExisted(sid) {
			metaBuf = append(metaBuf, metas...)
		}
	}

	metaBatch, seriesBatch, err := ws.buildBatch(ctx, metaBuf, seriesBuf)
	if err != nil {
		return fmt.Errorf("cannot build batch: %w", err)
	}

	if metaBatch != nil {
		go ws.sendBatch(ctx, metaBatch)
	}
	if seriesBatch != nil {
		go ws.sendBatch(ctx, seriesBatch)
	}
	return nil
}

func splitSpan(data *trace.Span) (*trace.Series, []*trace.Meta, uint64) {
	metas := make([]*trace.Meta, 0, len(data.Tags))
	for k, v := range data.Tags {
		metas = append(metas, &trace.Meta{Key: k, Value: v, OrgName: data.OrgName, CreateAt: data.EndTime})
	}
	sort.Slice(metas, func(i, j int) bool {
		return metas[i].Key < metas[j].Key
	})
	sid := hashTagsList(metas)

	for i := range metas {
		metas[i].SeriesID = sid
	}

	series := &trace.Series{
		OrgName:      data.OrgName,
		TraceId:      data.TraceId,
		SpanId:       data.SpanId,
		ParentSpanId: data.ParentSpanId,
		StartTime:    data.StartTime,
		EndTime:      data.EndTime,
		SeriesID:     sid,
	}
	return series, metas, sid
}

func (ws *WriteSpan) buildBatch(
	ctx context.Context,
	metaBuf []*trace.Meta,
	seriesBuf []*trace.Series,
) (metaBatch driver.Batch, seriesBatch driver.Batch, err error) {
	m, n := len(metaBuf), len(seriesBuf)
	if m > 0 {
		b, err := ws.client.PrepareBatch(ctx, "INSERT INTO "+ws.db+"."+trace.CH_TABLE_META)
		if err != nil {
			return nil, nil, err
		}
		for i := 0; i < m; i++ {
			if err := b.AppendStruct(metaBuf[i]); err != nil {
				_ = b.Abort()
				return nil, nil, err
			}
		}
		metaBatch = b
	}
	if n > 0 {
		b, err := ws.client.PrepareBatch(ctx, "INSERT INTO "+ws.db+"."+trace.CH_TABLE_SERIES)
		if err != nil {
			return nil, nil, err
		}
		for i := 0; i < n; i++ {
			if err := b.AppendStruct(seriesBuf[i]); err != nil {
				_ = b.Abort()
				return nil, nil, err
			}
		}
		seriesBatch = b
	}
	return
}

// TODO. Graceful close
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

func hashTagsList(list []*trace.Meta) uint64 {
	h := fnv.New64a()
	for _, item := range list {
		h.Write(strutil.NoCopyStringToBytes(item.Key))
		h.Write(strutil.NoCopyStringToBytes("\n"))
		h.Write(strutil.NoCopyStringToBytes(item.Value))
		h.Write(strutil.NoCopyStringToBytes("\n"))
	}
	return h.Sum64()
}

var metaBufPool sync.Pool

func getMetaBuf() []*trace.Meta {
	v := metaBufPool.Get()
	if v == nil {
		return []*trace.Meta{}
	}
	return v.([]*trace.Meta)
}

func putMetaBuf(buf []*trace.Meta) {
	buf = buf[:0]
	metaBufPool.Put(buf)
}

var seriesBufPool sync.Pool

func getseriesBuf() []*trace.Series {
	v := seriesBufPool.Get()
	if v == nil {
		return []*trace.Series{}
	}
	return v.([]*trace.Series)
}

func putseriesBuf(buf []*trace.Series) {
	buf = buf[:0]
	seriesBufPool.Put(buf)
}

// currency limiter
var currencyLimiter chan struct{}

func InitCurrencyLimiter(currency int) {
	currencyLimiter = make(chan struct{}, currency)
}

// seriesIDMap
var (
	seriesIDMap map[uint64]struct{}
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

	rows, err := client.Query(context.TODO(), "select distinct(series_id) from "+db+".spans_meta")
	if err != nil {
		return err
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
