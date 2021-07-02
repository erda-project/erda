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

package metric

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	monitordb "github.com/erda-project/erda/modules/msp/instance/db/monitor"
	mperm "github.com/erda-project/erda/modules/msp/instance/permission"
)

type config struct {
	MonitorURL string `file:"monitor_url"`
}

// +provider
type provider struct {
	Cfg           *config
	Log           logs.Logger
	DB            *gorm.DB          `autowired:"mysql-client"`
	Router        httpserver.Router `autowired:"http-router"`
	MPerm         mperm.Interface   `autowired:"msp.permission"`
	db            *monitordb.MonitorDB
	compatibleTKs map[string][]string
}

func (p *provider) Init(ctx servicehub.Context) (err error) {
	p.db = &monitordb.MonitorDB{DB: p.DB}
	err = p.loadCompatibleTKs()
	if err != nil {
		return err
	}
	return p.initRoutes(p.Router)
}

func init() {
	servicehub.Register("erda.msp.apm.metric", &servicehub.Spec{
		Services:   []string{"erda.msp.apm.metric"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
