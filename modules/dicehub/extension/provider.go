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
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"

	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/http/encoding"
	pb "github.com/erda-project/erda-proto-go/core/dicehub/extension/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dicehub/extension/db"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
	ExtensionMenu map[string][]string `file:"extension_menu" env:"EXTENSION_MENU"`
}

const FilePath = "/app/extensions-init"

// +provider
type provider struct {
	Cfg              *config
	Log              logs.Logger
	Register         transport.Register `autowired:"service-register" required:"true"`
	DB               *gorm.DB           `autowired:"mysql-client"`
	extensionService *extensionService
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
	go func() {
		err := p.extensionService.InitExtension(FilePath)
		if err != nil {
			panic(err)
		}
		logrus.Infoln("End init extension")
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
