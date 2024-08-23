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
	"sort"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/apps/msp/apm/trace"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/creator"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model/odata"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins/exporters/clickhouse/builder"
)

const (
	chTableCreator = "clickhouse.table.creator@span"
)

type Builder struct {
	logger  logs.Logger
	client  clickhouse.Conn
	Creator creator.Interface
	cfg     *builder.BuilderConfig
}

func NewBuilder(ctx servicehub.Context, logger logs.Logger, cfg *builder.BuilderConfig) (*Builder, error) {
	bu := &Builder{
		cfg:    cfg,
		logger: logger,
	}

	ch, err := builder.GetClickHouseInf(ctx, odata.SpanType)
	if err != nil {
		return nil, fmt.Errorf("get clickhouse interface: %w", err)
	}
	bu.client = ch.Client()

	if svc, ok := ctx.Service(chTableCreator).(creator.Interface); !ok {
		return nil, fmt.Errorf("service %q must existed", chTableCreator)
	} else {
		bu.Creator = svc
	}

	return bu, nil
}

func (bu *Builder) BuildBatch(ctx context.Context, sourceBatch interface{}) ([]driver.Batch, error) {
	items, ok := sourceBatch.([]*trace.Span)
	if !ok {
		return nil, fmt.Errorf("soureBatch<%T> must be []*trace.Span", sourceBatch)
	}
	// nolint
	batches, err := bu.buildBatches(ctx, items)
	if err != nil {
		return nil, fmt.Errorf("failed buildBatches: %w", err)
	}
	return batches, nil
}

func (bu *Builder) buildBatches(ctx context.Context, items []*trace.Span) ([]driver.Batch, error) {
	tablebatch := make(map[string]driver.Batch)
	for _, data := range items {
		table, err := bu.getOrCreateTenantTable(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("cannot get tenant table: %w", err)
		}
		if _, ok := tablebatch[table]; !ok {
			// nolint
			batch, err := bu.client.PrepareBatch(ctx, "INSERT INTO "+table, driver.WithReleaseConnection())
			if err != nil {
				return nil, err
			}
			tablebatch[table] = batch
		}
		batch := tablebatch[table]

		tagKeys := gettagKeysBuf()
		tagValues := gettagValuesBuf()

		pairs := getpairsBuf()
		for k, v := range data.Tags {
			pairs = append(pairs, odata.Tag{Key: k, Value: v})
		}
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Key < pairs[j].Key
		})
		for i := range pairs {
			tagKeys = append(tagKeys, pairs[i].Key)
			tagValues = append(tagValues, pairs[i].Value)
		}
		putpairsBuf(pairs)

		err = batch.AppendStruct(&trace.TableSpan{
			OrgName:       data.OrgName,
			TenantId:      data.Tags[bu.cfg.TenantIdKey],
			TraceId:       data.TraceId,
			SpanId:        data.SpanId,
			ParentSpanId:  data.ParentSpanId,
			OperationName: data.OperationName,
			StartTime:     time.Unix(0, data.StartTime),
			EndTime:       time.Unix(0, data.EndTime),
			TagKeys:       tagKeys,
			TagValues:     tagValues,
		})
		puttagKeysBuf(tagKeys)
		puttagValuesBuf(tagValues)
		if err != nil {
			_ = batch.Abort()
			return nil, fmt.Errorf("batch append: %w", err)
		}
	}
	batches := make([]driver.Batch, 0, len(tablebatch))
	for _, d := range tablebatch {
		batches = append(batches, d)
	}
	return batches, nil
}

func (bu *Builder) getOrCreateTenantTable(ctx context.Context, data *trace.Span) (string, error) {
	wait, table := bu.Creator.Ensure(ctx, data.Tags[lib.OrgNameKey], "", 0)
	if wait != nil {
		select {
		case <-wait:
		case <-ctx.Done():
			return table, nil
		}
	}
	return table, nil
}

var (
	tagKeysBuf, tagValuesBuf sync.Pool
	pairsBuf                 sync.Pool
)

func getpairsBuf() []odata.Tag {
	v := pairsBuf.Get()
	if v == nil {
		return []odata.Tag{}
	}
	return v.([]odata.Tag)
}

func putpairsBuf(buf []odata.Tag) {
	buf = buf[:0]
	pairsBuf.Put(buf)
}

func gettagKeysBuf() []string {
	v := tagKeysBuf.Get()
	if v == nil {
		return []string{}
	}
	return v.([]string)
}
func puttagKeysBuf(buf []string) {
	buf = buf[:0]
	tagKeysBuf.Put(buf)
}

func gettagValuesBuf() []string {
	v := tagValuesBuf.Get()
	if v == nil {
		return []string{}
	}
	return v.([]string)
}
func puttagValuesBuf(buf []string) {
	buf = buf[:0]
	tagValuesBuf.Put(buf)
}
