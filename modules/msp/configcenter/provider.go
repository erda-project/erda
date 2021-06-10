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

package configcenter

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/msp/configcenter/pb"
	instancedb "github.com/erda-project/erda/modules/msp/instance/db"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
}

// +provider
type provider struct {
	Cfg                 *config
	Log                 logs.Logger
	Register            transport.Register `autowired:"service-register" optional:"true"`
	DB                  *gorm.DB           `autowired:"mysql-client"`
	configCenterService *configCenterService
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.configCenterService = &configCenterService{
		p:                p,
		instanceTenantDB: &instancedb.InstanceTenantDB{DB: p.DB},
		instanceDB:       &instancedb.InstanceDB{DB: p.DB},
	}
	if p.Register != nil {
		pb.RegisterConfigCenterServiceImp(p.Register, p.configCenterService, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.msp.configcenter.ConfigCenterService" || ctx.Type() == pb.ConfigCenterServiceServerType() || ctx.Type() == pb.ConfigCenterServiceHandlerType():
		return p.configCenterService
	}
	return p
}

func init() {
	servicehub.Register("erda.msp.configcenter", &servicehub.Spec{
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
