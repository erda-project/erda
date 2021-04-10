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

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/cassandra"
	"github.com/gocql/gocql"
)

type define struct{}

func (d *define) Services() []string     { return []string{"trace-query"} }
func (d *define) Dependencies() []string { return []string{"cassandra"} }
func (d *define) Summary() string        { return "trace query" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{}    { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	Cassandra cassandra.SessionConfig `file:"cassandra"`
}

type provider struct {
	Cfg              *config
	Log              logs.Logger
	cassandraSession *gocql.Session
}

func (p *provider) Init(ctx servicehub.Context) error {
	c := ctx.Service("cassandra").(cassandra.Interface)
	session, err := c.Session(&p.Cfg.Cassandra)
	p.cassandraSession = session
	if err != nil {
		return fmt.Errorf("fail to create cassandra session: %s", err)
	}
	return nil
}

// Start .
func (p *provider) Start() error {
	return nil
}

func (p *provider) Close() error {
	return nil
}

func init() {
	servicehub.RegisterProvider("trace-query", &define{})
}
