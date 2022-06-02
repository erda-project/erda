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
	"strings"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/modules/apps/msp/apm/trace"
	"github.com/erda-project/erda/modules/tools/monitor/oap/collector/lib"
	"github.com/erda-project/erda/pkg/strutil"
)

type Config struct {
	RetryNum      int    `file:"retry_num" default:"5" ENV:"EXPORTER_CH_SPAN_RETRY_NUM"`
	SeriesTagKeys string `file:"series_tag_keys"`
}

type WriteSpan struct {
	db                  string
	retryCnt            int
	highCardinalityKeys map[string]struct{}
	logger              logs.Logger
	client              clickhouse.Conn
	itemsCh             chan []*trace.Span
}

func NewWriteSpan(cli clickhouse.Conn, logger logs.Logger, db string, cfg *Config) *WriteSpan {
	hk := make(map[string]struct{})
	for _, k := range strings.Split(cfg.SeriesTagKeys, ",") {
		hk[strings.TrimSpace(k)] = struct{}{}
	}

	return &WriteSpan{
		db:                  db,
		retryCnt:            cfg.RetryNum,
		highCardinalityKeys: hk,
		logger:              logger,
		client:              cli,
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

		err = ws.enrichBatch(metaBatch, seriesBatch, items)
		if err != nil {
			ws.logger.Errorf("failed enrichBatch: %w", err)
			continue
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

func (ws *WriteSpan) enrichBatch(metaBatch driver.Batch, seriesBatch driver.Batch, items []*trace.Span) error {
	for _, data := range items {
		var seriesTags map[string]string
		if len(ws.highCardinalityKeys) > 0 {
			seriesTags = make(map[string]string, len(ws.highCardinalityKeys))
		}

		// enrich metabatch
		metas := getMetaBuf()
		for k, v := range data.Tags {
			if len(ws.highCardinalityKeys) > 0 {
				if _, ok := ws.highCardinalityKeys[k]; ok {
					seriesTags[k] = v
					continue
				}
			}
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
					_ = metaBatch.Abort()
					putMetaBuf(metas)
					return fmt.Errorf("metaBatch append: %w", err)
				}
			}
		}
		putMetaBuf(metas)

		// enrich series batch
		series := getSeriesBuf()
		series.OrgName = data.OrgName
		series.TraceId = data.TraceId
		series.SpanId = data.SpanId
		series.ParentSpanId = data.ParentSpanId
		series.StartTime = data.StartTime
		series.EndTime = data.EndTime
		series.SeriesID = sid
		series.Tags = seriesTags
		err := seriesBatch.AppendStruct(&series)
		if err != nil { // TODO. Data may lost when encounter error
			_ = seriesBatch.Abort()
			putSeriesBuf(series)
			return fmt.Errorf("seriesBatch: %w", err)
		}
		putSeriesBuf(series)
	}
	return nil
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
	buf.Tags = nil
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
