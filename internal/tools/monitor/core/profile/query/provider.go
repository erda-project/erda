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

package query

import (
	"github.com/prometheus/client_golang/prometheus"
	storageconfig "github.com/pyroscope-io/pyroscope/pkg/config"
	"github.com/pyroscope-io/pyroscope/pkg/health"
	"github.com/pyroscope-io/pyroscope/pkg/service"
	"github.com/pyroscope-io/pyroscope/pkg/storage"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	mysql "github.com/erda-project/erda-infra/providers/mysql/v2"
)

type config struct {
	MaxNodesRender        int `file:"max_nodes_render" env:"MAX_NODES_RENDER" default:"8192"`
	MaxNodesSerialization int `file:"max_nodes_serialization" env:"MAX_NODES_SERIALIZATION" default:"2048"`
}

type provider struct {
	Cfg      *config
	Log      logs.Logger
	Register transport.Register `autowired:"service-register" optional:"true"`
	DB       mysql.Interface

	st                 *storage.Storage
	applicationService service.ApplicationMetadataService
}

func (p *provider) Init(ctx servicehub.Context) error {
	logger := logrus.StandardLogger()
	appMetadataSvc := service.NewApplicationMetadataService(p.DB.DB())
	p.applicationService = appMetadataSvc
	st, err := storage.New(storage.NewConfig(&storageconfig.Server{
		MaxNodesSerialization: 8192,
	}).WithInMemory(), logger, prometheus.DefaultRegisterer, new(health.Controller), p.applicationService)
	if err != nil {
		return err
	}
	p.st = st
	routes := ctx.Service("http-server", interceptors.Recover(p.Log)).(httpserver.Router)
	routes.GET("/api/profile/render", p.render)
	routes.GET("/api/profile/render-diff", p.renderDiff)
	routes.GET("/api/profile/apps", p.listApps)
	routes.GET("/api/profile/labels", p.getLabels)
	routes.GET("/api/profile/label-values", p.getLabelValues)
	return nil
}

func init() {
	servicehub.Register("erda.core.monitor.profile.render", &servicehub.Spec{
		Services:             []string{"erda.core.monitor.profile.render"},
		OptionalDependencies: []string{"service-register"},
		Description:          "profile render api",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
