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

	"github.com/golang/protobuf/proto"
	"github.com/labstack/echo"
	"github.com/prometheus/prometheus/prompb"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
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
	if p.consumerFunc == nil {
		return ctx.NoContent(http.StatusOK)
	}
	req := ctx.Request()
	buf, err := ReadBody(req)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("read body err: %s", err))
	}

	var wr prompb.WriteRequest
	err = proto.Unmarshal(buf, &wr)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("unmarshal body err: %s", err))
	}
	ms, err := convertToMetrics(wr)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("convertToMetrics err: %s", err))
	}
	p.consumerFunc(ms)

	return ctx.NoContent(http.StatusOK)
}

func (p *provider) RegisterConsumer(consumer model.ObservableDataConsumerFunc) {
	p.consumerFunc = consumer
}

func (p *provider) ComponentID() model.ComponentID {
	return model.ComponentID(providerName)
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
