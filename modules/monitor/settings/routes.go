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

package settings

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/jinzhu/gorm"
)

func (p *provider) intRoutes(routes httpserver.Router) error {
	routes.GET("/api/global/settings", p.getGlobalSetting)
	routes.PUT("/api/global/settings", p.setGlobalSetting)
	routes.PUT("/api/config/register", p.registerMonitorConfig)
	return nil
}

// configItem .
type configItem struct {
	Key   string      `json:"key"`
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
	Unit  string      `json:"unit"`
}

// configGroup .
type configGroup struct {
	Key   string        `json:"key"`
	Name  string        `json:"name"`
	Items []*configItem `json:"items"`
}

// globalSetting 面向用户的配置
type globalSetting struct {
	OrgID     int    `gorm:"column:org_id"`
	OrgName   string `gorm:"column:org_name"`
	Namespace string `gorm:"column:namespace"`
	Group     string `gorm:"column:group"`
	Type      string `gorm:"column:type"`
	Unit      string `gorm:"column:unit"`
	Key       string `gorm:"column:key"`
	Value     string `gorm:"column:value"`
}

const (
	globalSettingTableName    = "sp_monitor_global_settings"
	globalSettingInsertUpdate = "INSERT INTO `sp_monitor_global_settings`" +
		"(`org_id`,`org_name`,`namespace`,`group`,`key`,`type`,`value`,`unit`) " +
		"VALUES(?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `org_name`=VALUES(`org_name`),`type`=VALUES(`type`),`value`=VALUES(`value`),`unit`=VALUES(`unit`);"
)

func (p *provider) getGlobalSetting(req *http.Request, r struct {
	OrgID     int64  `query:"org_id" validate:"gt=0"`
	Workspace string `query:"workspace"`
}) interface{} {
	var list []*globalSetting
	if len(r.Workspace) > 0 {
		r.Workspace = strings.ToLower(r.Workspace)
		if err := p.db.Table(globalSettingTableName).Where("`org_id`=? AND `namespace`=?", r.OrgID, r.Workspace).
			Find(&list).Error; err != nil {
			p.L.Errorf("fail to load %s: %s", globalSettingTableName, err)
			return api.Errors.Internal(fmt.Sprintf("fail to load settings"), err.Error())
		}
	} else {
		if err := p.db.Table(globalSettingTableName).Where("`org_id`=?", r.OrgID).
			Find(&list).Error; err != nil {
			p.L.Errorf("fail to load settings: %s", err)
			return api.Errors.Internal(fmt.Sprintf("fail to load settings"), err.Error())
		}
	}
	lang := api.Language(req)
	cfg := p.getDefaultConfig(lang, r.Workspace)
	for _, item := range list {
		ns := cfg[item.Namespace]
		if ns == nil {
			ns := make(map[string]map[string]*configItem)
			cfg[item.Namespace] = ns
		}
		cg := ns[item.Group]
		if cg == nil {
			cg = make(map[string]*configItem)
			ns[item.Group] = cg
		}
		cg[item.Key] = &configItem{
			Key:   item.Key,
			Name:  item.Key,
			Type:  item.Type,
			Value: item.Value,
			Unit:  item.Unit,
		}
	}
	result := make(map[string][]*configGroup)
	for ns, groups := range cfg {
		nscfg := p.cfgMap[ns]
		for group, gcfg := range groups {
			cg := &configGroup{
				Key:  group,
				Name: group,
			}
			sort.Slice(cg.Items, func(i, j int) bool {
				return cg.Items[i].Key < cg.Items[j].Key
			})
			for _, item := range gcfg {
				cg.Items = append(cg.Items, item)
			}
			if nscfg != nil {
				cd := nscfg[group]
				if cd != nil && cd.convert != nil {
					cg = cd.convert(lang, ns, cg)
				}
			}
			result[ns] = append(result[ns], cg)
		}
		sort.Slice(result[ns], func(i, j int) bool {
			return result[ns][i].Key < result[ns][j].Key
		})
	}
	return api.Success(result)
}

func (p *provider) setGlobalSetting(r *http.Request, params struct {
	OrgID int `query:"org_id" validate:"gt=0"`
}, configs map[string][]*configGroup) interface{} {
	orgName, err := p.getOrgName(params.OrgID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	// orgName := "terminus"
	defCfg := p.getDefaultConfig(api.Language(r), "")
	tx := p.db.Begin()
	for ns, groups := range configs {
		ns = strings.ToLower(ns)
		nscfg := p.cfgMap[ns]
		nsdef := defCfg[ns]
		if nscfg == nil || nsdef == nil {
			continue
		}
		for _, group := range groups {
			if group == nil {
				continue
			}
			gcfg := nscfg[group.Key]
			gdef := nsdef[group.Key]
			if gcfg == nil || gcfg.handler == nil || gdef == nil {
				continue
			}
			cfg := map[string]interface{}{}
			for _, item := range group.Items {
				cdef := gdef[item.Key]
				if cdef == nil {
					continue
				}
				cfg[item.Key] = item.Value
				err := tx.Exec(globalSettingInsertUpdate, params.OrgID, orgName, ns, group.Key, item.Key, cdef.Type, fmt.Sprint(item.Value), cdef.Unit).Error
				if err != nil {
					tx.Rollback()
					return api.Errors.Internal(err)
				}
			}
			err := gcfg.handler(tx, params.OrgID, orgName, ns, group.Key, cfg)
			if err != nil {
				tx.Rollback()
				return api.Errors.Internal(err)
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success("OK")
}

type configDefine struct {
	handler  func(tx *gorm.DB, orgID int, orgName, ns, group string, keys map[string]interface{}) error
	defaults map[string]func(lang i18n.LanguageCodes) *configItem
	convert  func(lang i18n.LanguageCodes, ns string, gs *configGroup) *configGroup
}

func (p *provider) initConfigMap() {
	p.cfgMap = map[string]map[string]*configDefine{
		"dev": {
			"monitor": p.monitorConfigMap("dev"),
		},
		"test": {
			"monitor": p.monitorConfigMap("test"),
		},
		"staging": {
			"monitor": p.monitorConfigMap("staging"),
		},
		"prod": {
			"monitor": p.monitorConfigMap("prod"),
		},
		"general": {
			"monitor": p.monitorConfigMap("general"),
		},
	}
}

func (p *provider) getDefaultConfig(lang i18n.LanguageCodes, ns string) map[string]map[string]map[string]*configItem {
	result := map[string]map[string]map[string]*configItem{}
	if len(ns) > 0 {
		cfg := p.cfgMap[ns]
		if cfg == nil {
			return nil
		}
		nscfg := map[string]map[string]*configItem{}
		for group, cfg := range cfg {
			gcfg := map[string]*configItem{}
			for key, fn := range cfg.defaults {
				gcfg[key] = fn(lang)
			}
			nscfg[group] = gcfg
		}
		result[ns] = nscfg
	} else {
		for ns, cfg := range p.cfgMap {
			nscfg := map[string]map[string]*configItem{}
			for group, cfg := range cfg {
				gcfg := map[string]*configItem{}
				for key, fn := range cfg.defaults {
					gcfg[key] = fn(lang)
				}
				nscfg[group] = gcfg
			}
			result[ns] = nscfg
		}
	}
	return result
}

func getValue(typ string, value interface{}) interface{} {
	switch typ {
	case "number":
		switch val := value.(type) {
		case string:
			v, err := strconv.Atoi(val)
			if err == nil {
				return v
			}
		}
	}
	return value
}

func (p *provider) getOrgName(id int) (string, error) {
	resp, err := p.bundle.GetOrg(id)
	if err != nil {
		return "", fmt.Errorf("fail to get orgName: %s", err)
	}
	if resp == nil {
		return "", fmt.Errorf("org %d not exist", id)
	}
	return resp.Name, nil
}
