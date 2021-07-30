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

package dataview

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/core/monitor/dataview/pb"
	"github.com/erda-project/erda/modules/core/monitor/dataview/db"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
	Tables struct {
		SystemBlock string `file:"system_block" default:"sp_dashboard_block_system"`
		UserBlock   string `file:"user_block" default:"sp_dashboard_block"`
	} `file:"tables"`
}

// +provider
type provider struct {
	Cfg             *config
	Log             logs.Logger
	Register        transport.Register `autowired:"service-register" optional:"true"`
	DB              *gorm.DB           `autowired:"mysql-client"`
	dataViewService *dataViewService
}

func (p *provider) Init(ctx servicehub.Context) error {
	if len(p.Cfg.Tables.SystemBlock) > 0 {
		db.TableSystemView = p.Cfg.Tables.SystemBlock
	}
	if len(p.Cfg.Tables.UserBlock) > 0 {
		db.TableCustomView = p.Cfg.Tables.UserBlock
	}
	p.dataViewService = &dataViewService{
		p:      p,
		sys:    &db.SystemViewDB{DB: p.DB},
		custom: &db.CustomViewDB{DB: p.DB},
	}
	if p.Register != nil {
		pb.RegisterDataViewServiceImp(p.Register, p.dataViewService, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.monitor.dataview.DataViewService" || ctx.Type() == pb.DataViewServiceServerType() || ctx.Type() == pb.DataViewServiceHandlerType():
		return p.dataViewService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.monitor.dataview", &servicehub.Spec{
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
