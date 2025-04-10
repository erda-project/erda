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

package promremotewrite

import (
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo"

	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/receivercurrentlimiter"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/protoparser/promremotewrite"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
)

var providerName = plugins.WithPrefixReceiver("prometheus-remote-write")

type config struct {
	RemoteWriteUrl string `file:"remote_write_url" default:"/api/v1/prometheus-remote-write" desc:"remote write url"`
	GroupMetrics   struct {
		MinSize        int     `file:"min_size" default:"10" desc:"min size of group metrics"`
		RetentionRatio float64 `file:"retention_ratio" default:"0.5" desc:"retention ratio of group metrics"`
		GroupTagName   string  `file:"group_tag_name" default:"collector_group" desc:"group by which tag"`
	} `file:"group_metrics"`
}

var _ model.Receiver = (*provider)(nil)

// +provider
type provider struct {
	Cfg    *config
	Log    logs.Logger
	Router httpserver.Router `autowired:"http-router"`

	consumerFunc     model.ObservableDataConsumerFunc
	groupMetricsChan chan *metric.Metric
}

func (p *provider) ComponentClose() error {
	return nil
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	p.Log.Infof("group metric min_size: %v, retention_ratio: %v", p.Cfg.GroupMetrics.MinSize, p.Cfg.GroupMetrics.RetentionRatio)
	p.groupMetricsChan = make(chan *metric.Metric, 1000)
	go promremotewrite.DealGroupMetrics(ctx, promremotewrite.GroupMetricsOptions{
		MinSize:        p.Cfg.GroupMetrics.MinSize,
		RetentionRatio: p.Cfg.GroupMetrics.RetentionRatio,
		GroupTagName:   p.Cfg.GroupMetrics.GroupTagName,
		MetricsChan:    p.groupMetricsChan,
		Callback: func(record *metric.Metric) error {
			return p.consumerFunc(record)
		},
	})
	p.Router.POST(p.Cfg.RemoteWriteUrl, p.prwHandler)
	return nil
}

func (p *provider) prwHandler(ctx echo.Context) error {
	err := receivercurrentlimiter.Do(func() error {
		return promremotewrite.ParseStream(ctx.Request().Body, p.groupMetricsChan)
	})
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			p.Log.Errorf("close body error: %v", err)
		}
	}(ctx.Request().Body)

	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("parse stream: %s", err))
	}
	return ctx.NoContent(http.StatusOK)
}

func (p *provider) RegisterConsumer(consumer model.ObservableDataConsumerFunc) {
	p.consumerFunc = consumer
}

func (p *provider) ComponentConfig() interface{} {
	return p.Cfg
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services: []string{
			providerName,
		},
		Description: "here is description of prometheus-remote-write",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
