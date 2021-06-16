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

package registercenter

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/msp/registercenter/pb"
	"github.com/erda-project/erda/pkg/common/apis"
	perm "github.com/erda-project/erda/pkg/common/permission"
)

type config struct{}

// +provider
type provider struct {
	Cfg                   *config
	Log                   logs.Logger
	Register              transport.Register `autowired:"service-register" optional:"true"`
	DB                    *gorm.DB           `autowired:"mysql-client"`
	Perm                  perm.Interface     `autowired:"permission"`
	registerCenterService *registerCenterService
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.registerCenterService = &registerCenterService{
		p: p,
	}
	if p.Register != nil {
		type RegisterCenterService pb.RegisterCenterServiceServer
		pb.RegisterRegisterCenterServiceImp(p.Register, p.registerCenterService, apis.Options(), p.Perm.Check(
			perm.Method(RegisterCenterService.ListInterface, perm.ScopeProject, "registercenter", perm.ActionGet, perm.FieldValue("projectID")),
			perm.Method(RegisterCenterService.GetHTTPServices, perm.ScopeProject, "registercenter", perm.ActionGet, perm.FieldValue("projectID")),
			perm.Method(RegisterCenterService.EnableHTTPService, perm.ScopeProject, "registercenter", perm.ActionUpdate, perm.FieldValue("projectID")),
		))
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.msp.registercenter.RegisterCenterService" || ctx.Type() == pb.RegisterCenterServiceServerType() || ctx.Type() == pb.RegisterCenterServiceHandlerType():
		return p.registerCenterService
	}
	return p
}

func init() {
	servicehub.Register("erda.msp.registercenter", &servicehub.Spec{
		Services: pb.ServiceNames(),
		Types:    pb.Types(),
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
