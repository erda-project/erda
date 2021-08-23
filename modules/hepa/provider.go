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

package hepa

import (
	"context"
	"net/http"

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
	srv := &http.Server{
		Addr: p.Cfg.Server.ListenAddr,
	}
	errChan := make(chan error, 0)
	go func() {
		err = server.Start(srv)
		errChan <- err
	}()
	select {
	case <-ctx.Done():
		if err := srv.Shutdown(context.Background()); err != nil {
			logrus.Fatal("Server Shutdown:", err)
		}
		return nil
	case err := <-errChan:
		return err
	}
}

func init() {
	servicehub.Register("hepa", &servicehub.Spec{
		Services:    []string{"hepa"},
		Description: "hepa",
		ConfigFunc:  func() interface{} { return &myCfg{} },
		Creator:     func() servicehub.Provider { return &provider{} },
	})
}
