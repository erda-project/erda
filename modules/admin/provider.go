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

package admin

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/admin/dao"
	"github.com/erda-project/erda/modules/admin/manager"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
)

type Config struct {
	Debug bool `default:"false" env:"DEBUG" desc:"enable debug logging"`
}

type provider struct {
	Config Config
}

func init() {
	servicehub.Register("admin", &servicehub.Spec{
		Services:    []string{"admin"},
		Description: "erda platform admin",
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}

func (p *provider) Init(ctx servicehub.Context) error {

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000000000",
	})
	logrus.SetOutput(os.Stdout)

	if p.Config.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	return nil
}

func (p *provider) Run(ctx context.Context) error {
	var (
		dbClient *dao.DBClient
		err      error
	)
	if dbClient, err = dao.Open(); err != nil {
		logrus.Fatal(err)
	}
	defer dbClient.Close()

	etcdStore, err := etcd.New()
	if err != nil {
		return err
	}

	return p.RunServer(dbClient, etcdStore)
}

func (p *provider) RunServer(dbClient *dao.DBClient, etcdStore *etcd.Store) error {
	admin := manager.NewAdminManager(
		manager.WithDB(dbClient),
		manager.WithBundle(manager.NewBundle()),
		manager.WithETCDStore(etcdStore),
	)
	server, err := p.NewServer(admin.Routers())
	if err != nil {
		logrus.Fatal(err)
	}
	return server.ListenAndServe()
}

func (p *provider) NewServer(endpoints []httpserver.Endpoint) (*httpserver.Server, error) {
	server := httpserver.New(":9095")
	server.Router().UseEncodedPath()
	server.RegisterEndpoint(endpoints)
	return server, nil
}
