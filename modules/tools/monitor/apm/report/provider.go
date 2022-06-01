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

package report

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type provider struct {
	Cfg    *config
	Log    logs.Logger
	db     *gorm.DB
	ctx    servicehub.Context
	bundle *bundle.Bundle
}

type define struct{}

func (d *define) Services() []string     { return []string{"apm-report"} }
func (d *define) Dependencies() []string { return []string{"http-server", "mysql"} }
func (d *define) Summary() string        { return "apm-report api" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{}    { return &config{} }

func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct{}

func (report *provider) Init(ctx servicehub.Context) (err error) {
	report.ctx = ctx

	// mysql
	report.db = ctx.Service("mysql").(mysql.Interface).DB()

	// bundle
	hc := httpclient.New(httpclient.WithTimeout(time.Second, time.Second*60))
	report.bundle = bundle.New(
		bundle.WithHTTPClient(hc),
		bundle.WithCoreServices(),
	)

	// http server
	routes := ctx.Service("http-server", interceptors.Recover(report.Log)).(httpserver.Router)
	return report.initRoutes(routes)
}

func init() {
	servicehub.RegisterProvider("apm-report", &define{})
}
