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

package block

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/mysql"
)

type pconfig struct {
	Tables struct {
		SystemBlock string `file:"system_block" default:"sp_dashboard_block_system"`
		UserBlock   string `file:"user_block" default:"sp_dashboard_block"`
	} `file:"tables"`
}

type provider struct {
	Cfg *pconfig
	Log logs.Logger
	db  *DB
}

// Init .
func (p *provider) Init(ctx servicehub.Context) error {
	if len(p.Cfg.Tables.SystemBlock) > 0 {
		tableSystemBlock = p.Cfg.Tables.SystemBlock
	}
	if len(p.Cfg.Tables.UserBlock) > 0 {
		tableBlock = p.Cfg.Tables.UserBlock
	}
	p.db = newDB(ctx.Service("mysql").(mysql.Interface).DB())
	return nil
}

func init() {
	servicehub.Register("dataview-v1", &servicehub.Spec{
		Services:     []string{"chart-block"},
		Dependencies: []string{"mysql"},
		Description:  "chart block",
		ConfigFunc:   func() interface{} { return &pconfig{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
