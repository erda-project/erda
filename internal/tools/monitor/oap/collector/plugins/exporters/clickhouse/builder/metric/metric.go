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

package metric

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
	"github.com/erda-project/erda/internal/tools/monitor/core/settings/retention-strategy"
	tablepkg "github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/creator"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model/odata"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/typeconvert"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins/exporters/clickhouse/builder"
)

const (
	chTableCreator   = "clickhouse.table.creator@metric"
	chTableLoader    = "clickhouse.table.loader@metric"
	chTableRetention = "storage-retention-strategy@metric"
)

type Builder struct {
	logger    logs.Logger
	client    clickhouse.Conn
	Creator   creator.Interface
	Loader    loader.Interface
	Retention retention.Interface
	cfg       *builder.BuilderConfig
	batchC    chan []*metric.Metric
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

	ch, err := builder.GetClickHouseInf(ctx, odata.MetricType)
	if err != nil {
		return nil, fmt.Errorf("get clickhouse interface: %w", err)
	}
	bu.client = ch.Client()

	if svc, ok := ctx.Service(chTableCreator).(creator.Interface); !ok {
		return nil, fmt.Errorf("service %q must existed", chTableCreator)
	} else {
		bu.Creator = svc
	}

	if svc, ok := ctx.Service(chTableRetention).(retention.Interface); !ok {
		return nil, fmt.Errorf("service %q must existed", chTableRetention)
	} else {
		bu.Retention = svc
	}

	if svc, ok := ctx.Service(chTableLoader).(loader.Interface); !ok {
		return nil, fmt.Errorf("service %q must existed", chTableLoader)
	} else {
		bu.Loader = svc
	}

	return bu, nil
}

func (bu *Builder) buildBatches(ctx context.Context, items []*metric.Metric) ([]driver.Batch, error) {
	metaBatch, err := bu.client.PrepareBatch(ctx, "INSERT INTO "+bu.Loader.Database()+".metrics_meta", driver.WithReleaseConnection())
	if err != nil {
		return nil, fmt.Errorf("prepare metrics_meta batch: %w", err)
	}
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

		numFieldKeys := getnumFieldKeysBuf()
		numFieldValues := getnumFieldValuesBuf()
		strFieldKeys := getstrFieldKeysBuf()
		strFieldValues := getstrFieldValuesBuf()
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
			putnumFieldKeysBuf(numFieldKeys)
			putnumFieldValuesBuf(numFieldValues)
			putstrFieldKeysBuf(strFieldKeys)
			putstrFieldValuesBuf(strFieldValues)
			puttagKeysBuf(tagKeys)
			puttagValuesBuf(tagValues)
			continue
		}

		err1 := metaBatch.AppendStruct(&metric.TableMetricsMeta{
			OrgName:         data.OrgName,
			TenantId:        data.Tags[bu.cfg.TenantIdKey],
			MetricGroup:     data.Name,
			Timestamp:       time.Unix(0, data.Timestamp),
			NumberFieldKeys: numFieldKeys,
			StringFieldKeys: strFieldKeys,
			TagKeys:         tagKeys,
		})
		err2 := batch.AppendStruct(&metric.TableMetrics{
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
		putnumFieldKeysBuf(numFieldKeys)
		putnumFieldValuesBuf(numFieldValues)
		putstrFieldKeysBuf(strFieldKeys)
		putstrFieldValuesBuf(strFieldValues)
		puttagKeysBuf(tagKeys)
		puttagValuesBuf(tagValues)
		if err1 != nil || err2 != nil {
			_ = batch.Abort()
			return nil, fmt.Errorf("batch append: %w", err)
		}
	}
	batches := make([]driver.Batch, 0, len(tableBatch)+1)
	for _, d := range tableBatch {
		batches = append(batches, d)
	}
	batches = append(batches, metaBatch)
	return batches, nil
}

func (bu *Builder) getOrCreateTenantTable(ctx context.Context, data *metric.Metric) (string, error) {
	key := bu.Retention.GetConfigKey(data.Name, data.Tags)
	ttl := bu.Retention.GetTTL(key)
	var (
		wait  <-chan error
		table string
	)
	if len(key) > 0 {
		wait, table = bu.Creator.Ensure(ctx, data.Tags[lib.OrgNameKey], key, tablepkg.FormatTTLToDays(ttl))
	} else {
		wait, table = bu.Creator.Ensure(ctx, data.Tags[lib.OrgNameKey], "", tablepkg.FormatTTLToDays(ttl))
	}
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

func getnumFieldKeysBuf() []string {
	v := numFieldKeysBuf.Get()
	if v == nil {
		return []string{}
	}
	return v.([]string)
}
func putnumFieldKeysBuf(buf []string) {
	buf = buf[:0]
	numFieldKeysBuf.Put(buf)
}

func getnumFieldValuesBuf() []float64 {
	v := numFieldValuesBuf.Get()
	if v == nil {
		return []float64{}
	}
	return v.([]float64)
}
func putnumFieldValuesBuf(buf []float64) {
	buf = buf[:0]
	numFieldValuesBuf.Put(buf)
}

func getstrFieldKeysBuf() []string {
	v := strFieldKeysBuf.Get()
	if v == nil {
		return []string{}
	}
	return v.([]string)
}
func putstrFieldKeysBuf(buf []string) {
	buf = buf[:0]
	strFieldKeysBuf.Put(buf)
}

func getstrFieldValuesBuf() []string {
	v := strFieldValuesBuf.Get()
	if v == nil {
		return []string{}
	}
	return v.([]string)
}
func putstrFieldValuesBuf(buf []string) {
	buf = buf[:0]
	strFieldValuesBuf.Put(buf)
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
