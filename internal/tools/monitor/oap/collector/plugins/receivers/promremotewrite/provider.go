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
	"net/http"

	"github.com/labstack/echo"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/protoparser/promremotewrite"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
)

var providerName = plugins.WithPrefixReceiver("prometheus-remote-write")

type config struct {
}

// +provider
type provider struct {
	Cfg    *config
	Log    logs.Logger
	Router httpserver.Router `autowired:"http-router"`

	consumerFunc model.ObservableDataConsumerFunc
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	p.Router.POST("/api/v1/prometheus-remote-write", p.prwHandler)
	return nil
}

func (p *provider) prwHandler(ctx echo.Context) error {
	if err := promremotewrite.ParseStream(ctx.Request().Body, func(record *metric.Metric) error {
		return p.consumerFunc(lib.ConsumerTimeout, record)
	}); err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("parse stream: %s", err))
	}
	defer ctx.Request().Body.Close()

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
