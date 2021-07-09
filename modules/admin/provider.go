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
