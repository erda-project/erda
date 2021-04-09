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

package hepa

import (
	"context"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	_ "github.com/erda-project/erda-infra/providers/health"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/hepa/common"
	"github.com/erda-project/erda/modules/hepa/config"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
	"github.com/erda-project/erda/modules/hepa/server"
	"github.com/erda-project/erda/modules/hepa/ver"
)

// define Represents the definition of provider and provides some information
type define struct{}

// Declare what services the provider provides
func (d *define) Service() []string { return []string{"hepa"} }

// Declare which services the provider depends on
func (d *define) Dependencies() []string { return []string{} }

// Describe information about this provider
func (d *define) Description() string { return "hepa" }

// Return an instance representing the configuration
func (d *define) Config() interface{} { return &myCfg{} }

// Return a provider creator
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type myCfg struct {
	Log    config.LogConfig
	Server config.ServerConfig
}

type provider struct {
	Cfg *myCfg      // auto inject this field
	Log logs.Logger // auto inject this field
}

func (p *provider) Init(ctx servicehub.Context) error {
	config.ServerConf = &p.Cfg.Server
	config.LogConf = &p.Cfg.Log
	common.InitLogger()
	orm.Init()
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	server.CreateSingleton(common.AccessLog)
	gwCtl, err := server.NewGatewayController()
	if err != nil {
		return err
	}
	gwCtl.Register()
	openapiCtl, err := server.NewOpenapiController()
	if err != nil {
		return err
	}
	openapiCtl.Register()
	logrus.Info(ver.String())

	return server.Start(p.Cfg.Server.ListenAddr)
}

func init() {
	servicehub.RegisterProvider("hepa", &define{})
}
