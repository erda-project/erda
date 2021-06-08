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

package menu

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/msp/menu/pb"
	"github.com/erda-project/erda/bundle"
	instancedb "github.com/erda-project/erda/modules/msp/instance/db"
	db "github.com/erda-project/erda/modules/msp/menu/db"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/jinzhu/gorm"
)

type config struct {
}

// +provider
type provider struct {
	Cfg         *config
	Log         logs.Logger
	Register    transport.Register
	DB          *gorm.DB `autowired:"mysql-client"`
	menuService *menuService
	bdl         *bundle.Bundle
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bdl = bundle.New(bundle.WithScheduler())

	if p.Register != nil {
		p.menuService = &menuService{
			p:                p,
			db:               &db.MenuConfigDB{DB: p.DB},
			instanceTenantDB: &instancedb.InstanceTenantDB{DB: p.DB},
			instanceDB:       &instancedb.InstanceDB{DB: p.DB},
			bdl:              p.bdl,
		}
		pb.RegisterMenuServiceImp(p.Register, p.menuService, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.msp.menu.MenuService" || ctx.Type() == pb.MenuServiceServerType() || ctx.Type() == pb.MenuServiceHandlerType():
		return p.menuService
	}
	return p
}

func init() {
	servicehub.Register("erda.msp.menu", &servicehub.Spec{
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
