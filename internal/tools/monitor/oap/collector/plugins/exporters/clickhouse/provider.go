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
	"github.com/erda-project/erda/internal/apps/msp/apm/trace"
	"github.com/erda-project/erda/internal/tools/monitor/core/log"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/erda-project/erda/internal/tools/monitor/core/profile"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model/odata"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins/exporters/clickhouse/builder"
	externalmetric "github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins/exporters/clickhouse/builder/external_metric"
	logStore "github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins/exporters/clickhouse/builder/log"
	metricstore "github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins/exporters/clickhouse/builder/metric"
	profilebuilder "github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins/exporters/clickhouse/builder/profile"
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

var _ model.Exporter = (*provider)(nil)

// +provider
type provider struct {
	Cfg     *config
	Log     logs.Logger
	storage *Storage
}

func (p *provider) ComponentClose() error {
	return p.storage.Close()
}

func (p *provider) ExportRaw(items ...*odata.Raw) error { return nil }
func (p *provider) ExportLog(items ...*log.Log) error {
	p.storage.WriteBatchAsync(items)
	return nil
}
func (p *provider) ExportProfile(items ...*profile.Output) error {
	p.storage.WriteBatchAsync(items)
	return nil
}

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
	case odata.ProfileType:
	case odata.ExternalMetricType:
	case odata.LogType:
	default:
		return fmt.Errorf("invalid builder for data_type: %q", p.Cfg.BuilderCfg.DataType)
	}

	var batchBuilder BatchBuilder
	switch dt {
	case odata.SpanType:
		tmp, err := span.NewBuilder(ctx, p.Log.Sub("span-builder"), p.Cfg.BuilderCfg)
		if err != nil {
			return fmt.Errorf("span build: %w", err)
		}
		batchBuilder = tmp
	case odata.MetricType:
		tmp, err := metricstore.NewBuilder(ctx, p.Log.Sub("metric-builder"), p.Cfg.BuilderCfg)
		if err != nil {
			return fmt.Errorf("metrics build: %w", err)
		}
		batchBuilder = tmp
	case odata.ProfileType:
		tmp, err := profilebuilder.NewBuilder(ctx, p.Log.Sub("profile-builder"), p.Cfg.BuilderCfg)
		if err != nil {
			return fmt.Errorf("profile build: %w", err)
		}
		batchBuilder = tmp
	case odata.ExternalMetricType:
		tmp, err := externalmetric.NewBuilder(ctx, p.Log.Sub("external-metric-builder"), p.Cfg.BuilderCfg)
		if err != nil {
			return fmt.Errorf("external metrics build: %w", err)
		}
		batchBuilder = tmp
	case odata.LogType:
		tmp, err := logStore.NewBuilder(ctx, p.Log.Sub("log-builder"), p.Cfg.BuilderCfg)
		if err != nil {
			return fmt.Errorf("log build: %w", err)
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
