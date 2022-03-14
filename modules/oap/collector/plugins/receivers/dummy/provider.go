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

package dummy

import (
	"context"
	"encoding/json"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	lpb "github.com/erda-project/erda-proto-go/oap/logs/pb"
	mpb "github.com/erda-project/erda-proto-go/oap/metrics/pb"
	tpb "github.com/erda-project/erda-proto-go/oap/trace/pb"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/core/model/odata"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
)

var providerName = plugins.WithPrefixReceiver("dummy")

type config struct {
	// you can use the sample in testdata/
	MetricSample string        `file:"metric_sample"`
	TraceSample  string        `file:"trace_sample"`
	LogSample    string        `file:"log_sample"`
	Rate         time.Duration `file:"rate" default:"10s"`
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger

	consumerFunc model.ObservableDataConsumerFunc
}

func (p *provider) ComponentConfig() interface{} {
	return p.Cfg
}

func (p *provider) RegisterConsumer(consumer model.ObservableDataConsumerFunc) {
	p.consumerFunc = consumer
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	if p.Cfg.MetricSample != "" {
		go p.dummyMetrics(ctx)
	}

	if p.Cfg.TraceSample != "" {
		go p.dummyTraces(ctx)
	}

	if p.Cfg.LogSample != "" {
		go p.dummyLogs(ctx)
	}
	return nil
}

func (p *provider) dummyMetrics(ctx context.Context) {
	ticker := time.NewTicker(p.Cfg.Rate)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
		item := mpb.Metric{}
		err := json.Unmarshal([]byte(p.Cfg.MetricSample), &item)
		if err != nil {
			p.Log.Errorf("unmarshal MetricSample err: %s", err)
		}
		if p.consumerFunc != nil {
			p.consumerFunc(odata.NewMetric(&item))
		}
	}
}

func (p *provider) dummyTraces(ctx context.Context) {
	ticker := time.NewTicker(p.Cfg.Rate)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
		item := tpb.Span{}
		err := json.Unmarshal([]byte(p.Cfg.TraceSample), &item)
		if err != nil {
			p.Log.Errorf("unmarshal TraceSample err: %s", err)
		}
		if p.consumerFunc != nil {
			p.consumerFunc(odata.NewSpan(&item))
		}
	}
}

func (p *provider) dummyLogs(ctx context.Context) {
	ticker := time.NewTicker(p.Cfg.Rate)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
		item := lpb.Log{}
		err := json.Unmarshal([]byte(p.Cfg.MetricSample), &item)
		if err != nil {
			p.Log.Errorf("unmarshal LogSample err: %s", err)
		}
		item.TimeUnixNano = uint64(time.Now().Nanosecond())

		if p.consumerFunc != nil {
			p.consumerFunc(odata.NewLog(&item))
		}
	}
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services: []string{
			providerName,
		},
		Description: "dummy receiver for debug&test",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
