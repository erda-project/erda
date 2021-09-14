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

package extension

import (
	"context"
	"net/http"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"

	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/http/encoding"
	election "github.com/erda-project/erda-infra/providers/etcd-election"
	pb "github.com/erda-project/erda-proto-go/core/dicehub/extension/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dicehub/extension/db"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
	ExtensionMenu        map[string][]string `file:"extension_menu" env:"EXTENSION_MENU"`
	ExtensionSources     string              `file:"extension_sources" env:"EXTENSION_SOURCES"`
	ExtensionSourcesCron string              `file:"extension_sources_cron" env:"EXTENSION_SOURCES_CRON"`
}

const FilePath = "/app/extensions-init"

// +provider
type provider struct {
	Cfg                   *config
	Log                   logs.Logger
	Register              transport.Register `autowired:"service-register" required:"true"`
	DB                    *gorm.DB           `autowired:"mysql-client"`
	InitExtensionElection election.Interface `autowired:"etcd-election@initExtension"`
	extensionService      *extensionService
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.newExtensionService()
	if p.Register != nil {
		pb.RegisterExtensionServiceImp(p.Register, p.extensionService, apis.Options(),
			transport.WithHTTPOptions(
				transhttp.WithDecoder(func(r *http.Request, data interface{}) error {
					lang := r.URL.Query().Get("lang")
					if lang != "" {
						r.Header.Set("lang", lang)
					}
					return encoding.DecodeRequest(r, data)
				}),
			))
	}

	var once sync.Once
	p.InitExtensionElection.OnLeader(func(ctx context.Context) {
		once.Do(func() {
			err := p.extensionService.InitExtension(FilePath, false)
			if err != nil {
				panic(err)
			}
			logrus.Infoln("End init extension")
		})
	})

	go func() {
		c := cron.New()

		err := c.AddFunc(p.Cfg.ExtensionSourcesCron, func() {
			go p.extensionService.TimedTaskSynchronizationExtensions()
		})

		if err != nil {
			logrus.Errorf("error to add cron task %v", err)
		} else {
			c.Start()
		}
	}()
	return nil
}

func (p *provider) newExtensionService() {
	p.extensionService = &extensionService{
		p:             p,
		db:            &db.ExtensionConfigDB{DB: p.DB},
		bdl:           bundle.New(bundle.WithCoreServices()),
		extensionMenu: p.Cfg.ExtensionMenu,
	}
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.dicehub.extension.ExtensionService" || ctx.Type() == pb.ExtensionServiceServerType() || ctx.Type() == pb.ExtensionServiceHandlerType():
		return p.extensionService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.dicehub.extension", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                pb.Types(),
		OptionalDependencies: []string{"service-register"},
		Description:          "",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
