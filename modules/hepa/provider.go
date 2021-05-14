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

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	_ "github.com/erda-project/erda-infra/providers/health"
	"github.com/erda-project/erda/modules/hepa/common"
	"github.com/erda-project/erda/modules/hepa/config"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
	"github.com/erda-project/erda/modules/hepa/server"
	"github.com/erda-project/erda/modules/hepa/ver"
)

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
	logrus.Infof("server conf: %+v", config.ServerConf)
	logrus.Infof("log conf: %+v", config.LogConf)
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
	servicehub.Register("hepa", &servicehub.Spec{
		Services:    []string{"hepa"},
		Description: "hepa",
		ConfigFunc:  func() interface{} { return &myCfg{} },
		Creator:     func() servicehub.Provider { return &provider{} },
	})
}
