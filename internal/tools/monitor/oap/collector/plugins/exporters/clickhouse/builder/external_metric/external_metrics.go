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

package external_metric

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/creator"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model/odata"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/typeconvert"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins/exporters/clickhouse/builder"
)

const (
	chTableCreator = "clickhouse.table.creator@external_metric"
	chTableLoader  = "clickhouse.table.loader@external_metric"
)

type Builder struct {
	logger  logs.Logger
	client  clickhouse.Conn
	Creator creator.Interface
	Loader  loader.Interface
	cfg     *builder.BuilderConfig
	batchC  chan []*metric.Metric
}

func (bu *Builder) BuildBatch(ctx context.Context, sourceBatch interface{}) ([]driver.Batch, error) {
	items, ok := sourceBatch.([]*metric.Metric)
	if !ok {
		return nil, fmt.Errorf("soureBatch<%T> must be []*metric.Metric", sourceBatch)
	}

	batches, err := bu.buildBatches(ctx, items)
	if err != nil {
		return nil, fmt.Errorf("failed buildBatches: %w", err)
	}
	return batches, nil
}

func NewBuilder(ctx servicehub.Context, logger logs.Logger, cfg *builder.BuilderConfig) (*Builder, error) {
	bu := &Builder{
		cfg:    cfg,
		logger: logger,
	}

	ch, err := builder.GetClickHouseInf(ctx, odata.ExternalMetricType)
	if err != nil {
		return nil, fmt.Errorf("get clickhouse interface: %w", err)
	}
	bu.client = ch.Client()

	if svc, ok := ctx.Service(chTableCreator).(creator.Interface); !ok {
		return nil, fmt.Errorf("service %q must existed", chTableCreator)
	} else {
		bu.Creator = svc
	}

	if svc, ok := ctx.Service(chTableLoader).(loader.Interface); !ok {
		return nil, fmt.Errorf("service %q must existed", chTableLoader)
	} else {
		bu.Loader = svc
	}

	return bu, nil
}

func (bu *Builder) buildBatches(ctx context.Context, items []*metric.Metric) ([]driver.Batch, error) {
	tableBatch := make(map[string]driver.Batch)
	for _, data := range items {
		table, err := bu.getOrCreateTenantTable(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("cannot get tenant table: %w", err)
		}

		if _, ok := tableBatch[table]; !ok {
			// nolint
			batch, err := bu.client.PrepareBatch(ctx, "INSERT INTO "+table, driver.WithReleaseConnection())
			if err != nil {
				return nil, err
			}
			tableBatch[table] = batch
		}
		batch := tableBatch[table]

		numFieldKeys := getNumFieldKeysBuf()
		numFieldValues := getNumFieldValuesBuf()
		strFieldKeys := getStrFieldKeysBuf()
		strFieldValues := getStrFieldValuesBuf()
		tagKeys := getTagKeysBuf()
		tagValues := getTagValuesBuf()

		pairs := getPairBuf()
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
		putpPairBuf(pairs)

		for k, v := range data.Fields {
			switch vv := v.(type) {
			case string:
				strFieldKeys = append(strFieldKeys, k)
				strFieldValues = append(strFieldValues, vv)
			default:
				vfloat64, err := typeconvert.ToFloat64(v)
				if err != nil {
					continue
				}
				numFieldKeys = append(numFieldKeys, k)
				numFieldValues = append(numFieldValues, vfloat64)
			}
		}
		// ignore empty
		if len(numFieldKeys) == 0 && len(strFieldKeys) == 0 {
			putNumFieldKeysBuf(numFieldKeys)
			putNumFieldValuesBuf(numFieldValues)
			putStrFieldKeysBuf(strFieldKeys)
			putStrFieldValuesBuf(strFieldValues)
			putTagKeysBuf(tagKeys)
			putTagValuesBuf(tagValues)
			continue
		}

		err1 := batch.AppendStruct(&metric.TableMetrics{
			OrgName:           data.OrgName,
			TenantId:          data.Tags[bu.cfg.TenantIdKey],
			MetricGroup:       data.Name,
			Timestamp:         time.Unix(0, data.Timestamp),
			NumberFieldKeys:   numFieldKeys,
			NumberFieldValues: numFieldValues,
			StringFieldKeys:   strFieldKeys,
			StringFieldValues: strFieldValues,
			TagKeys:           tagKeys,
			TagValues:         tagValues,
		})
		putNumFieldKeysBuf(numFieldKeys)
		putNumFieldValuesBuf(numFieldValues)
		putStrFieldKeysBuf(strFieldKeys)
		putStrFieldValuesBuf(strFieldValues)
		putTagKeysBuf(tagKeys)
		putTagValuesBuf(tagValues)
		if err1 != nil {
			if abortErr := batch.Abort(); abortErr != nil {
				bu.logger.Errorf("failed to abort external metric batch, err: %v", abortErr)
			}
			return nil, fmt.Errorf("batch append: %w", err1)
		}
	}
	batches := make([]driver.Batch, 0, len(tableBatch))
	for _, d := range tableBatch {
		batches = append(batches, d)
	}
	return batches, nil
}

func (bu *Builder) getOrCreateTenantTable(ctx context.Context, data *metric.Metric) (string, error) {
	var (
		wait  <-chan error
		table string
	)
	wait, table = bu.Creator.Ensure(ctx, data.Tags[lib.OrgNameKey], "", int64(time.Hour))
	if wait != nil {
		select {
		case <-wait: // ignore
		case <-ctx.Done():
			return table, nil
		}
	}
	return table, nil
}

var (
	numFieldKeysBuf, numFieldValuesBuf, strFieldKeysBuf, strFieldValuesBuf, tagKeysBuf, tagValuesBuf sync.Pool
	pairsBuf                                                                                         sync.Pool
)

func getPairBuf() []odata.Tag {
	v := pairsBuf.Get()
	if v == nil {
		return []odata.Tag{}
	}
	return v.([]odata.Tag)
}

func putpPairBuf(buf []odata.Tag) {
	buf = buf[:0]
	pairsBuf.Put(buf)
}

func getNumFieldKeysBuf() []string {
	v := numFieldKeysBuf.Get()
	if v == nil {
		return []string{}
	}
	return v.([]string)
}
func putNumFieldKeysBuf(buf []string) {
	buf = buf[:0]
	numFieldKeysBuf.Put(buf)
}

func getNumFieldValuesBuf() []float64 {
	v := numFieldValuesBuf.Get()
	if v == nil {
		return []float64{}
	}
	return v.([]float64)
}
func putNumFieldValuesBuf(buf []float64) {
	buf = buf[:0]
	numFieldValuesBuf.Put(buf)
}

func getStrFieldKeysBuf() []string {
	v := strFieldKeysBuf.Get()
	if v == nil {
		return []string{}
	}
	return v.([]string)
}
func putStrFieldKeysBuf(buf []string) {
	buf = buf[:0]
	strFieldKeysBuf.Put(buf)
}

func getStrFieldValuesBuf() []string {
	v := strFieldValuesBuf.Get()
	if v == nil {
		return []string{}
	}
	return v.([]string)
}
func putStrFieldValuesBuf(buf []string) {
	buf = buf[:0]
	strFieldValuesBuf.Put(buf)
}

func getTagKeysBuf() []string {
	v := tagKeysBuf.Get()
	if v == nil {
		return []string{}
	}
	return v.([]string)
}
func putTagKeysBuf(buf []string) {
	buf = buf[:0]
	tagKeysBuf.Put(buf)
}

func getTagValuesBuf() []string {
	v := tagValuesBuf.Get()
	if v == nil {
		return []string{}
	}
	return v.([]string)
}
func putTagValuesBuf(buf []string) {
	buf = buf[:0]
	tagValuesBuf.Put(buf)
}
