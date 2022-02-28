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

package fluentbit

import (
	"fmt"
	"net/http"

	"github.com/buger/jsonparser"
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/oap/collector/authentication"
	"github.com/erda-project/erda/modules/oap/collector/common"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/core/model/odata"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
	"github.com/labstack/echo"
)

var providerName = plugins.WithPrefixReceiver("fluent-bit")

type config struct {
	FLBKeyMappings flbKeyMappings `file:"flb_key_mappings"`
}

// This is the key mappings config of erda's internal model to fluent-bit's output format
type flbKeyMappings struct {
	TimeUnixNano string `file:"time_unix_nano" default:"time"`
	Name         string `file:"name" default:"source"`
	Content      string `file:"log" default:"log"`
	Severity     string `file:"severity" default:"level"`
	Stream       string `file:"stream" default:"stream"`
	Attributes   string `file:"attributes" default:"tags"`
	// parse kubernetes's metadata from `kubernetes` field
	Kubernetes string `file:"kubernetes" default:"kubernetes"`
}

// +provider
type provider struct {
	Cfg       *config
	Log       logs.Logger
	Router    httpserver.Router        `autowired:"http-router"`
	Validator authentication.Validator `autowired:"erda.oap.collector.authentication.Validator"`

	consumerFunc model.ObservableDataConsumerFunc
}

func (p *provider) ComponentID() model.ComponentID {
	return model.ComponentID(providerName)
}

func (p *provider) RegisterConsumer(consumer model.ObservableDataConsumerFunc) {
	p.consumerFunc = consumer
	return
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	p.Router.POST("/api/v1/collect/fluent-bit", p.flbHandler)
	return nil
}

func (p *provider) flbHandler(ctx echo.Context) error {
	if p.consumerFunc == nil {
		return ctx.NoContent(http.StatusOK)
	}
	req := ctx.Request()
	buf, err := common.ReadBody(req)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("read body err: %s", err))
	}

	_, err = jsonparser.ArrayEach(buf, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			p.Log.Errorf("jsonparser err: %s", err)
			return
		}
		lg, err := parseItem(value, p.Cfg.FLBKeyMappings)
		if err != nil {
			p.Log.Errorf("parseItem err: %s", err)
			return
		}
		p.consumerFunc(odata.NewLog(lg))
	})
	if err != nil {
		return fmt.Errorf("parser err: %w", err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services: []string{
			providerName,
		},
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
