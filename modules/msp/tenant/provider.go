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

package tenant

import (
	"github.com/jinzhu/gorm"

	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	pb "github.com/erda-project/erda-proto-go/msp/tenant/pb"
	"github.com/erda-project/erda/modules/msp/instance/db/monitor"
	"github.com/erda-project/erda/modules/msp/tenant/db"
	"github.com/erda-project/erda/pkg/common/apis"
	perm "github.com/erda-project/erda/pkg/common/permission"
)

type config struct {
}

// +provider
type provider struct {
	Cfg           *config
	Log           logs.Logger
	Register      transport.Register
	tenantService *tenantService
	DB            *gorm.DB       `autowired:"mysql-client"`
	Perm          perm.Interface `autowired:"permission"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.tenantService = &tenantService{
		p:            p,
		MSPTenantDB:  &db.MSPTenantDB{DB: p.DB},
		MSPProjectDB: &db.MSPProjectDB{DB: p.DB},
		MonitorDB:    &monitor.MonitorDB{DB: p.DB},
	}
	if p.Register != nil {
		type TenantService pb.TenantServiceServer
		pb.RegisterTenantServiceImp(p.Register, p.tenantService, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.msp.tenant.TenantService" || ctx.Type() == pb.TenantServiceServerType() || ctx.Type() == pb.TenantServiceHandlerType():
		return p.tenantService
	}
	return p
}

func init() {
	servicehub.Register("erda.msp.tenant", &servicehub.Spec{
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
