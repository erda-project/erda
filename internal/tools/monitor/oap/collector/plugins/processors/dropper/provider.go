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

package dropper

import (
	"strings"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/apps/msp/apm/trace"
	"github.com/erda-project/erda/internal/tools/monitor/core/log"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/erda-project/erda/internal/tools/monitor/core/profile"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model/odata"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins"
)

var providerName = plugins.WithPrefixProcessor("dropper")

type config struct {
	MetricPrefix string `file:"metric_prefix"`
}

type provider struct {
	Cfg *config
	Log logs.Logger
}

var _ model.Processor = (*provider)(nil)

func (p *provider) ComponentClose() error {
	return nil
}

func (p *provider) ComponentConfig() interface{} {
	return p.Cfg
}

func (p *provider) ProcessMetric(item *metric.Metric) (*metric.Metric, error) {
	if len(p.Cfg.MetricPrefix) > 0 && strings.HasPrefix(item.Name, p.Cfg.MetricPrefix) {
		return nil, nil
	}
	return item, nil
}

func (p *provider) ProcessLog(item *log.Log) (*log.Log, error) { return item, nil }

func (p *provider) ProcessSpan(item *trace.Span) (*trace.Span, error) { return item, nil }

func (p *provider) ProcessRaw(item *odata.Raw) (*odata.Raw, error) { return item, nil }

func (p *provider) ProcessProfile(*profile.ProfileIngest) (*profile.Output, error) {
	return &profile.Output{}, nil
}

func (p *provider) Init(ctx servicehub.Context) error {
	return nil
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services: []string{
			providerName,
		},
		Description: "help to drop item by prefix name",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
