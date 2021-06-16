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
	"net/http"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/http/encoding"
	"github.com/erda-project/erda-proto-go/msp/configcenter/pb"
	instancedb "github.com/erda-project/erda/modules/msp/instance/db"
	mperm "github.com/erda-project/erda/modules/msp/instance/permission"
	"github.com/erda-project/erda/pkg/common/apis"
	perm "github.com/erda-project/erda/pkg/common/permission"
)

type config struct {
}

// +provider
type provider struct {
	Cfg                 *config
	Log                 logs.Logger
	Register            transport.Register `autowired:"service-register" optional:"true"`
	DB                  *gorm.DB           `autowired:"mysql-client"`
	Perm                perm.Interface     `autowired:"permission"`
	MPerm               mperm.Interface    `autowired:"msp.permission"`
	configCenterService *configCenterService
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.configCenterService = &configCenterService{
		p:                p,
		instanceTenantDB: &instancedb.InstanceTenantDB{DB: p.DB},
		instanceDB:       &instancedb.InstanceDB{DB: p.DB},
	}
	if p.Register != nil {
		type ConfigCenterService = pb.ConfigCenterServiceServer
		pb.RegisterConfigCenterServiceImp(p.Register, p.configCenterService, apis.Options(),
			transport.WithHTTPOptions(transhttp.WithEncoder(func(rw http.ResponseWriter, r *http.Request, data interface{}) error {
				// compatibility with api "/api/tmc/config/tenants/{tenantID}/groups/{groupID}" response
				if resp, ok := data.(*pb.GetGroupPropertiesResponse); ok && resp != nil {
					m := make(map[string]interface{})
					for _, item := range resp.Data {
						m[item.Group] = item.Properties
					}
					data = m
				}
				return encoding.EncodeResponse(rw, r, data)
			})),
			p.Perm.Check(
				perm.Method(ConfigCenterService.GetGroups, perm.ScopeProject, "config-center_group", perm.ActionGet, p.MPerm.TenantToProjectID("", "TenantID")),
				perm.Method(ConfigCenterService.GetGroupProperties, perm.ScopeProject, "config-center_properties", perm.ActionGet, p.MPerm.TenantToProjectID("", "TenantID")),
				perm.Method(ConfigCenterService.SaveGroupProperties, perm.ScopeProject, "config-center_properties", perm.ActionUpdate, p.MPerm.TenantToProjectID("", "TenantID")),
			),
		)
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
