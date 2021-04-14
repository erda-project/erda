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

package example

import (
	"context"
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/httpserver"
)

type config struct {
	Message  string        `file:"message" flag:"msg" default:"hi" desc:"message for example"`
	Interval time.Duration `file:"interval" flag:"interval" default:"3s" desc:"interval to print message"`
}

type provider struct {
	Cfg *config     // auto inject this field
	Log logs.Logger // auto inject this field
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

func (p *provider) Run(ctx context.Context) error {
	p.Log.Info("example provider is running...")
	tick := time.Tick(p.Cfg.Interval)
	for {
		select {
		case <-tick:
			// do something
			p.Log.Info("message: ", p.Cfg.Message)
		case <-ctx.Done():
			return nil
		}
	}
}

func init() {
	servicehub.Register("erda.example.http", &servicehub.Spec{
		Services:     []string{"erda.example.http"},
		Dependencies: []string{"http-server"},
		Description:  "example provider",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
