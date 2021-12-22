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
	"time"

	"embed"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	componentprotocol "github.com/erda-project/erda-infra/providers/component-protocol"
	"github.com/erda-project/erda-infra/providers/component-protocol/protocol"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/admin/dao"
	"github.com/erda-project/erda/modules/admin/manager"
	"github.com/erda-project/erda/modules/admin/services/workbench"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/conf"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
)

//go:embed component-protocol/scenarios
var scenarioFS embed.FS

type Config struct {
	Debug bool `default:"false" env:"DEBUG" desc:"enable debug logging"`
}

type provider struct {
	Config Config

	Log      logs.Logger
	Protocol componentprotocol.Interface
	CPTran   i18n.I18n       `autowired:"i18n@cp"`
	Tran     i18n.Translator `translator:"common"`
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
	p.Log.Info("init admin")

	p.Log.Info("init component-protocol")
	bdl := bundle.New(
		bundle.WithHepa(),
		bundle.WithOrchestrator(),
		bundle.WithEventBox(),
		bundle.WithGittar(),
		bundle.WithDOP(),
		bundle.WithMSP(),
		bundle.WithPipeline(),
		bundle.WithMonitor(),
		bundle.WithCollector(),
		bundle.WithHTTPClient(httpclient.New(
			httpclient.WithTimeout(time.Second*15, time.Duration(conf.BundleTimeoutSecond())*time.Second), // bundle 默认 (time.Second, time.Second*3)
		)),
		bundle.WithKMS(),
		bundle.WithCoreServices(),
		bundle.WithHTTPClient(
			httpclient.New(
				httpclient.WithTimeout(time.Second, time.Second*90),
				httpclient.WithEnableAutoRetry(false),
			)),
	)

	p.Protocol.SetI18nTran(p.CPTran)
	wb := workbench.New(workbench.WithBundle(bdl))
	p.Protocol.WithContextValue(types.Workbench, wb)
	p.Protocol.WithContextValue(types.GlobalCtxKeyBundle, bdl)
	protocol.MustRegisterProtocolsFromFS(scenarioFS)
	p.Log.Info("init component-protocol done")

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
