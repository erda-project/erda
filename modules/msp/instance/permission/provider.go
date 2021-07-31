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

package permission

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	instancedb "github.com/erda-project/erda/modules/msp/instance/db"
	monitordb "github.com/erda-project/erda/modules/msp/instance/db/monitor"
	tenantdb "github.com/erda-project/erda/modules/msp/tenant/db"
)

// +provider
type provider struct {
	Log              logs.Logger
	DB               *gorm.DB `autowired:"mysql-client"`
	instanceTenantDB *instancedb.InstanceTenantDB
	tmcDB            *instancedb.TmcDB
	monitorDB        *monitordb.MonitorDB
	MSPTenantDB      *tenantdb.MSPTenantDB
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.instanceTenantDB = &instancedb.InstanceTenantDB{DB: p.DB}
	p.tmcDB = &instancedb.TmcDB{DB: p.DB}
	p.monitorDB = &monitordb.MonitorDB{DB: p.DB}
	p.MSPTenantDB = &tenantdb.MSPTenantDB{DB: p.DB}
	return nil
}

func init() {
	servicehub.Register("msp.permission", &servicehub.Spec{
		Services: []string{"msp.permission"},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
