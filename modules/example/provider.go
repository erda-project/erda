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

// Author: recallsong
// Email: songruiguo@qq.com

package example

import (
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/httpserver"
)

// define Represents the definition of provider and provides some information
type define struct{}

// Declare what services the provider provides
func (d *define) Service() []string { return []string{"example"} }

// Declare which services the provider depends on
func (d *define) Dependencies() []string { return []string{"http-server"} }

// Describe information about this provider
func (d *define) Description() string { return "example" }

// Return an instance representing the configuration
func (d *define) Config() interface{} { return &config{} }

// Return a provider creator
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{
			closeCh: make(chan struct{}),
		}
	}
}

type config struct {
	Message  string        `file:"message" flag:"msg" default:"hi" desc:"message for example"`
	Interval time.Duration `file:"interval" flag:"interval" default:"3s" desc:"interval to print message"`
}

type provider struct {
	C       *config     // auto inject this field
	L       logs.Logger // auto inject this field
	closeCh chan struct{}
}

func (p *provider) Init(ctx servicehub.Context) error {
	// register some apis
	routes := ctx.Service("http-server").(httpserver.Router)
	routes.GET("/hello/:name",
		func(params struct {
			Name    string `param:"name"`
			Message string `query:"message" validate:"required"`
		}) interface{} {
			// return api.Errors.Internal("example error")
			return api.Success(fmt.Sprintf("hello %s, %s", params.Name, params.Message))
		},
	)
	// http://localhost:8080/hello/recallsong?message=good
	// response:
	// {
	//     "success": true,
	//     "data": "hello recallsong, good"
	// }
	return nil
}

func (p *provider) Start() error {
	p.L.Info("example provider is running...")
	tick := time.Tick(10 * time.Second)
	for {
		select {
		case <-tick:
			// do something
			p.L.Info("message: ", p.C.Message)
		case <-p.closeCh:
			return nil
		}
	}
}

func (p *provider) Close() error {
	p.L.Info("example provider is closing...")
	close(p.closeCh)
	return nil
}

func init() {
	servicehub.RegisterProvider("example", &define{})
}
