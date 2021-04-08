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
	"net/http"
	"reflect"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
)

var clientType = reflect.TypeOf((*MetricQuery)(nil)).Elem()

type config struct {
	Endpoint string        `file:"endpoint" default:"http://localhost:7076" desc:"collector host"`
	Timeout  time.Duration `file:"timeout" default:"10s"`
}

type define struct{}

func (d *define) Services() []string { return []string{"metricq-client"} }
func (d *define) Types() []reflect.Type {
	return []reflect.Type{
		clientType,
	}
}
func (d *define) Summary() string { return "metricq-client" }

func (d *define) Description() string { return d.Summary() }

func (d *define) Config() interface{} { return &config{} }

func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type provider struct {
	Cfg    *config
	Log    logs.Logger
	Client *queryClient
}

// Init .
func (p *provider) Init(ctx servicehub.Context) error {
	client := &queryClient{
		endpoint:   p.Cfg.Endpoint,
		timeout:    p.Cfg.Timeout,
		httpClient: &http.Client{},
	}
	p.Client = client
	return nil
}

// Provide .
func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	return p.Client
}

func init() {
	servicehub.RegisterProvider("metricq-client", &define{})
}
