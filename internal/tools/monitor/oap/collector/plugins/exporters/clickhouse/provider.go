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
	"strings"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/clickhouse"
	"github.com/erda-project/erda/internal/apps/msp/apm/trace"
	"github.com/erda-project/erda/internal/tools/monitor/core/log"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model/odata"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins/exporters/clickhouse/builder"
	metricstore "github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins/exporters/clickhouse/builder/metric"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins/exporters/clickhouse/builder/span"
)

type config struct {
	Keypass    map[string][]string `file:"keypass"`
	Keydrop    map[string][]string `file:"keydrop"`
	Keyinclude []string            `file:"keyinclude"`
	Keyexclude []string            `file:"keyexclude"`

	StorageCfg *StorageConfig         `file:"storage"`
	BuilderCfg *builder.BuilderConfig `file:"builder"`
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger
	ch  clickhouse.Interface

	storage *Storage
}

func (p *provider) ExportRaw(items ...*odata.Raw) error { return nil }
func (p *provider) ExportLog(items ...*log.Log) error   { return nil }

func (p *provider) ExportMetric(items ...*metric.Metric) error {
	p.storage.WriteBatchAsync(items)
	return nil
}

func (p *provider) ExportSpan(items ...*trace.Span) error {
	p.storage.WriteBatchAsync(items)
	return nil
}

func (p *provider) ComponentConfig() interface{} {
	return p.Cfg
}

func (p *provider) Connect() error {
	return nil
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	dt := odata.DataType(strings.ToUpper(p.Cfg.BuilderCfg.DataType))
	switch dt {
	case odata.SpanType:
	case odata.MetricType:
	default:
		return fmt.Errorf("invalid builder for data_type: %q", p.Cfg.BuilderCfg.DataType)
	}

	var batchBuilder BatchBuilder
	switch dt {
	case odata.SpanType:
		batchBuilder = span.NewBuilder(p.ch.Client(), p.Log.Sub("span-builder"), p.Cfg.BuilderCfg)
	case odata.MetricType:
		tmp, err := metricstore.NewBuilder(ctx, p.Log.Sub("metric-builder"), p.Cfg.BuilderCfg)
		if err != nil {
			return fmt.Errorf("metrics build: %w", err)
		}
		batchBuilder = tmp
	default:
		return fmt.Errorf("invalid data_type: %q", p.Cfg.BuilderCfg.DataType)
	}
	p.storage = &Storage{
		cfg:             p.Cfg.StorageCfg,
		logger:          p.Log.Sub("storage"),
		currencyLimiter: make(chan struct{}, p.Cfg.StorageCfg.CurrencyNum),
		sqlBuilder:      batchBuilder,
	}
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	return p.storage.Start(ctx)
}

func init() {
	servicehub.Register("erda.oap.collector.exporter.clickhouse", &servicehub.Spec{
		Services: []string{
			"erda.oap.collector.exporter.clickhouse",
		},
		Description:  "here is description of erda.oap.collector.exporter.clickhouse",
		Dependencies: []string{"clickhouse", "clickhouse.table.initializer"},
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
