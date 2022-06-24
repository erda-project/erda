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
	"github.com/erda-project/erda/internal/apps/msp/apm/trace"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/uintset"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins/exporters/clickhouse/builder"
	"github.com/erda-project/erda/pkg/strutil"
)

const activeCacheDuration = 24 * time.Hour

type Builder struct {
	db                  string
	logger              logs.Logger
	client              clickhouse.Conn
	sidSet              *uintset.Uint64Set
	cfg                 *builder.BuilderConfig
	highCardinalityKeys map[string]struct{}
}

func NewBuilder(cli clickhouse.Conn, logger logs.Logger, cfg *builder.BuilderConfig) *Builder {
	hk := make(map[string]struct{})
	for _, k := range strings.Split(cfg.SeriesTagKeys, ",") {
		hk[strings.TrimSpace(k)] = struct{}{}
	}

	return &Builder{
		db:                  cfg.Database,
		cfg:                 cfg,
		highCardinalityKeys: hk,
		logger:              logger,
		client:              cli,
	}
}

func (st *Builder) Start(ctx context.Context) error {
	err := st.initSeriesIDSet()
	if err != nil {
		return fmt.Errorf("initSeriesIDSet: %w", err)
	}

	go st.syncSeriesIDSet(ctx)
	return nil
}

func (st *Builder) BuildBatch(ctx context.Context, sourceBatch interface{}) ([]driver.Batch, error) {
	items, ok := sourceBatch.([]*trace.Span)
	if !ok {
		return nil, fmt.Errorf("soureBatch<%T> must be []*trace.Span", sourceBatch)
	}
	// nolint
	metaBatch, err := st.client.PrepareBatch(ctx, "INSERT INTO "+st.db+"."+trace.CH_TABLE_META)
	if err != nil {
		return nil, fmt.Errorf("prepare metaBatch: %s", err)

	}
	// nolint
	seriesBatch, err := st.client.PrepareBatch(ctx, "INSERT INTO "+st.db+"."+trace.CH_TABLE_SERIES)
	if err != nil {
		return nil, fmt.Errorf("prepare seriesBatch: %s", err)
	}

	err = st.enrichBatch(metaBatch, seriesBatch, items)
	if err != nil {
		return nil, fmt.Errorf("failed enrichBatch: %w", err)
	}

	return []driver.Batch{metaBatch, seriesBatch}, nil
}

func (st *Builder) initSeriesIDSet() error {
	now := time.Now()
retry:
	// nolint
	rows, err := st.client.Query(context.TODO(), fmt.Sprintf("SELECT distinct(series_id) FROM %s.spans_meta_all WHERE create_at > fromUnixTimestamp64Nano(cast(%d,'Int64'))", st.db, now.Add(-activeCacheDuration).UnixNano()))
	if err != nil {
		if exp, ok := err.(*clickhouse.Exception); ok && validExp(exp) {
			time.Sleep(time.Second)
			goto retry
		} else {
			return err
		}
	}

	sids := make([]uint64, 0, 1024)
	for rows.Next() {
		var sid uint64
		if err := rows.Scan(&sid); err != nil {
			return fmt.Errorf("scan: %w", err)
		}
		sids = append(sids, sid)
	}
	st.sidSet.AddBatch(sids)
	return rows.Close()
}

func (st *Builder) syncSeriesIDSet(ctx context.Context) {
	ticker := time.NewTicker(st.cfg.TTLCheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			st.sidSet.CleanOldPart()
		}
	}
}

func (st *Builder) enrichBatch(metaBatch driver.Batch, seriesBatch driver.Batch, items []*trace.Span) error {
	for _, data := range items {
		var seriesTags map[string]string
		if len(st.highCardinalityKeys) > 0 {
			seriesTags = make(map[string]string, len(st.highCardinalityKeys))
		}

		// enrich metabatch
		metas := getMetaBuf()
		for k, v := range data.Tags {
			if len(st.highCardinalityKeys) > 0 {
				if _, ok := st.highCardinalityKeys[k]; ok {
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

		if !st.sidSet.Has(sid) {
			for i := range metas {
				metas[i].SeriesID = sid
				err := metaBatch.AppendStruct(&metas[i])
				if err != nil { // TODO. Data may lost when encounter error
					_ = metaBatch.Abort()
					putMetaBuf(metas)
					return fmt.Errorf("metaBatch append: %w", err)
				}
			}
			st.sidSet.Add(sid)
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
			return fmt.Errorf("seriesBatch append: %w", err)
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

func validExp(exp *clickhouse.Exception) bool {
	switch exp.Code {
	case 81, 60: // db or table not existed
		return true
	}
	return false
}
