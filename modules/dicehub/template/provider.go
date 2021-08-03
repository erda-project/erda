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

package template

import (
	"github.com/jinzhu/gorm"

	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	pb "github.com/erda-project/erda-proto-go/core/dicehub/template/pb"
	"github.com/erda-project/erda/modules/dicehub/template/db"
)

type config struct {
}

// +provider
type provider struct {
	Cfg             *config
	Log             logs.Logger
	Register        transport.Register `autowired:"service-register" required:"true"`
	DB              *gorm.DB           `autowired:"mysql-client"`
	templateService *templateService
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.templateService = &templateService{
		p: p,
		db: &db.TemplateDB{
			DB: p.DB,
		},
	}
	if p.Register != nil {
		pb.RegisterTemplateServiceImp(p.Register, p.templateService)
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.dicehub.template.TemplateService" || ctx.Type() == pb.TemplateServiceServerType() || ctx.Type() == pb.TemplateServiceHandlerType():
		return p.templateService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.dicehub.template", &servicehub.Spec{
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
