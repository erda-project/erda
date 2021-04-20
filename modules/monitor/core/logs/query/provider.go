// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package query

import (
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/cassandra"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/gocql/gocql"
)

type define struct{}

func (d *define) Services() []string { return []string{"logs-query"} }
func (d *define) Dependencies() []string {
	return []string{"http-server", "cassandra"}
}
func (d *define) Summary() string     { return "logs store" }
func (d *define) Description() string { return d.Summary() }
func (d *define) Config() interface{} { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	Cassandra cassandra.SessionConfig `file:"cassandra"`
	Download  struct {
		TimeSpan time.Duration `file:"time_span" default:"5m"`
	} `file:"download"`
}

type provider struct {
	Cfg              *config
	Logger           logs.Logger
	session          *gocql.Session
	checkOrgCluster  func(ctx httpserver.Context) (string, error)
	getApplicationID func(ctx httpserver.Context) (string, error)
}

func (p *provider) Init(ctx servicehub.Context) error {
	cassandra := ctx.Service("cassandra").(cassandra.Interface)
	session, err := cassandra.Session(&p.Cfg.Cassandra)
	if err != nil {
		return fmt.Errorf("fail to create cassandra session: %s", err)
	}
	p.session = session
	routes := ctx.Service("http-server", interceptors.Recover(p.Logger)).(httpserver.Router)
	err = p.intRoutes(routes)
	if err != nil {
		return fmt.Errorf("fail to init routes: %s", err)
	}
	return nil
}

func init() {
	servicehub.RegisterProvider("logs-query", &define{})
}
