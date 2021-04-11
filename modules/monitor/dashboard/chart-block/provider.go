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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda/modules/monitor/common/db"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq"
	"github.com/erda-project/erda/modules/monitor/utils"
)

type define struct{}

func (d *define) Services() []string { return []string{"chart-block"} }
func (d *define) Dependencies() []string {
	return []string{"http-server", "mysql"}
}
func (d *define) Summary() string     { return "chart block" }
func (d *define) Description() string { return d.Summary() }
func (d *define) Config() interface{} { return &pconfig{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider { return &provider{} }
}

type pconfig struct {
	PresetDashboards string `file:"preset_dashboards"`
	Tables           struct {
		SystemBlock string `file:"system_block" default:"sp_dashboard_block_system"`
		UserBlock   string `file:"user_block" default:"sp_dashboard_block"`
	} `file:"tables"`
}

type provider struct {
	Cfg       *pconfig
	Log       logs.Logger
	metricq   metricq.Queryer
	db        *DB
	authDb    *db.DB
	presetMap map[string][]string
}

// Init .
func (p *provider) Init(ctx servicehub.Context) error {
	if len(p.Cfg.Tables.SystemBlock) > 0 {
		tableSystemBlock = p.Cfg.Tables.SystemBlock
	}
	if len(p.Cfg.Tables.UserBlock) > 0 {
		tableBlock = p.Cfg.Tables.UserBlock
	}

	p.authDb = db.New(ctx.Service("mysql").(mysql.Interface).DB())

	p.db = newDB(ctx.Service("mysql").(mysql.Interface).DB())
	if len(p.Cfg.PresetDashboards) > 0 {
		if err := p.preLoadPresetDashboard(); err != nil {
			return err
		}
		// compatibility for version 3.17
		if err := p.removeOldPresetDashboard(); err != nil {
			return err
		}
	}
	routes := ctx.Service("http-server", interceptors.Recover(p.Log)).(httpserver.Router)
	return p.intRoutes(routes)
}

func (p *provider) getOrCreatePresetDashboard(d *DashboardBlockQuery) (res []*UserBlock, err error) {
	if len(p.Cfg.PresetDashboards) <= 0 {
		return
	}
	presets, ok := p.presetMap[d.Scope]
	if !ok {
		return
	}
	var userBlocks []*UserBlock
	for _, fp := range presets {
		buf, err := ioutil.ReadFile(filepath.Join(p.Cfg.PresetDashboards, fp))
		if err != nil {
			return nil, err
		}
		dashboard := UserBlock{}
		err = json.Unmarshal(buf, &dashboard)
		if err != nil {
			return nil, err
		}
		dashboard.ID = utils.GetMD5Base64WithLegth([]byte(fp+"_"+d.Scope+"_"+d.ScopeID), 64)
		dashboard.Scope = d.Scope // always same
		dashboard.ScopeID = d.ScopeID
		// dynamic add filter org_name for org, terminus_key for micro_service
		q, err := url.ParseQuery("r_scopeId=" + url.QueryEscape(dashboard.ScopeID))
		if err != nil {
			return nil, err
		}
		dashboard.ViewConfig.replaceWithQuery(q)
		block := UserBlock{ID: dashboard.ID, Scope: dashboard.Scope, ScopeID: dashboard.ScopeID}
		if err := p.db.Where(&block).FirstOrCreate(&dashboard).Error; err != nil {
			return nil, err
		}
		userBlocks = append(userBlocks, &block)
	}
	return userBlocks, nil
}

func (p *provider) preLoadPresetDashboard() error {
	if len(p.Cfg.PresetDashboards) <= 0 {
		return nil
	}
	p.presetMap = make(map[string][]string)
	ds, err := os.Stat(p.Cfg.PresetDashboards)
	if err != nil {
		return err
	}
	if !ds.IsDir() {
		return fmt.Errorf("path of preset_dashboards must be directory")
	}
	if err := filepath.Walk(p.Cfg.PresetDashboards, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		buf, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		dashboard := UserBlock{}
		err = json.Unmarshal(buf, &dashboard)
		if err != nil {
			return err
		}
		if v, ok := p.presetMap[dashboard.Scope]; !ok {
			p.presetMap[dashboard.Scope] = []string{info.Name()}
		} else {
			p.presetMap[dashboard.Scope] = append(v, info.Name())
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to load preset dashboard. err=%s", err)
	}
	return nil
}

func (p *provider) removeOldPresetDashboard() error {
	if err := p.db.Where("id like ?", "aliyun-metrics-dashboard-1%").Delete(UserBlock{}).Error; err != nil {
		return err
	}
	return nil
}

func init() {
	servicehub.RegisterProvider("chart-block", &define{})
}
