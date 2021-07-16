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

package monitor

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/msp/instance/db"
	"github.com/erda-project/erda/modules/msp/instance/db/monitor"
	"github.com/erda-project/erda/modules/msp/resource/deploy/handlers"
)

type config struct {
	TaCollectUrl string `file:"ta_collect_url" default:"//analytics.terminus.io/collect"`
	TaStaticUrl  string `file:"ta_static_url" default:"//analytics.terminus.io/ta.js"`
}

// +provider
type provider struct {
	*handlers.DefaultDeployHandler
	Cfg       *config
	Log       logs.Logger
	DB        *gorm.DB `autowired:"mysql-client"`
	MonitorDb *monitor.MonitorDB
	ProjectDb *db.ProjectDB
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.DefaultDeployHandler = handlers.NewDefaultHandler(p.DB, p.Log)
	p.MonitorDb = &monitor.MonitorDB{DB: p.DB}
	p.ProjectDb = &db.ProjectDB{DB: p.DB}
	return nil
}

func init() {
	servicehub.Register("erda.msp.resource.deploy.handlers.monitor", &servicehub.Spec{
		Services: []string{
			"erda.msp.resource.deploy.handlers.monitor",
		},
		Description: "erda.msp.resource.deploy.handlers.monitor",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
