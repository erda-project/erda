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
