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

package collector

import (
	"strings"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/authentication"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins"
)

var providerName = plugins.WithPrefixReceiver("collector")

type config struct {
	MetadataKeyOfTopic string `file:"metadata_key_of_topic"`
	Auth               struct {
		Username string `file:"username"`
		Password string `file:"password"`
		Force    bool   `file:"force"`
		Skip     bool   `file:"skip"`
	}
}

var _ model.Receiver = (*provider)(nil)

// +provider
type provider struct {
	Cfg       *config
	Log       logs.Logger
	Router    httpserver.Router `autowired:"http-router"`
	Validator authentication.Validator

	auth     *Authenticator
	consumer model.ObservableDataConsumerFunc
}

type skipValidator struct{}

func (skipValidator) Validate(scope string, scopeId string, token string) bool {
	return true
}

func (p *provider) ComponentClose() error {
	return nil
}

func (p *provider) ComponentConfig() interface{} {
	return p.Cfg
}

func (p *provider) RegisterConsumer(consumer model.ObservableDataConsumerFunc) {
	p.consumer = consumer
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	if p.Cfg.Auth.Skip {
		p.Validator = skipValidator{}
	} else {
		p.Validator = ctx.Service("erda.oap.collector.authentication.Validator").(authentication.Validator)
	}
	p.auth = NewAuthenticator(
		WithLogger(p.Log),
		WithValidator(p.Validator),
		WithConfig(p.Cfg),
	)

	p.Router.POST("/api/v1/collect/logs/:source", p.collectLogs, p.auth.keyAuth())
	p.Router.POST("/api/v1/collect/:metric", p.collectMetric, p.auth.keyAuth())

	p.Router.POST("/collect/:metric", p.collectMetric, p.auth.basicAuth())
	p.Router.POST("/collect/logs/:source", p.collectLogs, p.auth.basicAuth())
	p.Router.POST("/collect/logs-all", p.collectLogsAll, p.auth.basicAuth())

	// profile
	p.Router.POST("/ingest", p.collectProfile, p.auth.basicAuth())
	return nil
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services:    []string{providerName},
		Description: "here is description of erda.oap.collector.receiver.collector",
		DependenciesFunc: func(hub *servicehub.Hub) (list []string) {
			hub.ForeachServices(func(service string) bool {
				if strings.HasPrefix(service, "erda.oap.collector.authentication.") {
					list = append(list, service)
				}
				return true
			})
			return list
		},
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
