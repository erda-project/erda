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

package project

import (
	"github.com/jinzhu/gorm"

	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/i18n"
	tenantpb "github.com/erda-project/erda-proto-go/msp/tenant/pb"
	pb "github.com/erda-project/erda-proto-go/msp/tenant/project/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/msp/instance/db/monitor"
	"github.com/erda-project/erda/modules/msp/tenant/db"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
}

// +provider
type provider struct {
	Cfg            *config
	Log            logs.Logger
	Register       transport.Register
	projectService *projectService
	bdl            *bundle.Bundle
	I18n           i18n.Translator              `autowired:"i18n" translator:"msp-i18n"`
	DB             *gorm.DB                     `autowired:"mysql-client"`
	TenantServer   tenantpb.TenantServiceServer `autowired:"erda.msp.tenant.TenantService"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bdl = bundle.New(bundle.WithMSP())
	p.projectService = &projectService{
		p:            p,
		MSPProjectDB: &db.MSPProjectDB{DB: p.DB},
		MSPTenantDB:  &db.MSPTenantDB{DB: p.DB},
		MonitorDB:    &monitor.MonitorDB{DB: p.DB},
	}
	if p.Register != nil {
		pb.RegisterProjectServiceImp(p.Register, p.projectService, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.msp.tenant.project.ProjectService" || ctx.Type() == pb.ProjectServiceServerType() || ctx.Type() == pb.ProjectServiceHandlerType():
		return p.projectService
	}
	return p
}

func init() {
	servicehub.Register("erda.msp.tenant.project", &servicehub.Spec{
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
