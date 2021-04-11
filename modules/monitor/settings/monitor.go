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
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/pkg/router"
	"github.com/jinzhu/gorm"
	"github.com/recallsong/go-utils/conv"
	"github.com/recallsong/go-utils/encoding/md5x"
	"github.com/recallsong/go-utils/reflectx"
)

func (p *provider) monitorConfigMap(ns string) *configDefine {
	metricDays, logDays := 8, 7
	ttl := os.Getenv("METRIC_INDEX_TTL")
	if len(ttl) > 0 {
		d, err := time.ParseDuration(ttl)
		if err != nil {
			p.L.Errorf("fail to parse metric ttl: %s", err)
		} else {
			metricDays = int(math.Ceil(d.Hours() / 24))
		}
	}
	ttl = os.Getenv("LOG_TTL")
	if len(ttl) > 0 {
		sed, err := strconv.ParseInt(ttl, 10, 64)
		if err != nil {
			p.L.Errorf("fail to parse log ttl: %s", err)
		} else {
			const daySec = float64(24 * 60 * 60)
			logDays = int(math.Ceil(float64(sed) / daySec))
		}
	}
	cd := &configDefine{
		handler: p.updateMonitorConfig,
	}
	if ns == "general" {
		cd.defaults = map[string]func(lang i18n.LanguageCodes) *configItem{
			"metrics_ttl": func(lang i18n.LanguageCodes) *configItem {
				return &configItem{
					Key:   "metrics_ttl",
					Name:  p.t.Text(lang, "base") + " " + p.t.Text(lang, "metrics_ttl"),
					Type:  "number",
					Value: metricDays,
					Unit:  p.t.Text(lang, "days"),
				}
			},
		}
		cd.convert = func(lang i18n.LanguageCodes, ns string, cg *configGroup) *configGroup {
			cg.Name = p.t.Text(lang, cg.Key)
			for _, item := range cg.Items {
				item.Name = p.t.Text(lang, item.Key)
				if item.Key == "metrics_ttl" {
					item.Name = p.t.Text(lang, "base") + " " + item.Name
				}
				item.Value = getValue(item.Type, item.Value)
			}
			return cg
		}
	} else {
		cd.defaults = map[string]func(lang i18n.LanguageCodes) *configItem{
			"logs_ttl": func(lang i18n.LanguageCodes) *configItem {
				return &configItem{
					Key:   "logs_ttl",
					Name:  p.t.Text(lang, "logs_ttl"),
					Type:  "number",
					Value: logDays,
					Unit:  p.t.Text(lang, "days"),
				}
			},
			"metrics_ttl": func(lang i18n.LanguageCodes) *configItem {
				return &configItem{
					Key:   "metrics_ttl",
					Name:  p.t.Text(lang, "metrics_ttl"),
					Type:  "number",
					Value: metricDays,
					Unit:  p.t.Text(lang, "days"),
				}
			},
		}
		cd.convert = func(lang i18n.LanguageCodes, ns string, cg *configGroup) *configGroup {
			cg.Name = p.t.Text(lang, cg.Key)
			for _, item := range cg.Items {
				item.Name = p.t.Text(lang, item.Key)
				if item.Key == "metrics_ttl" {
					item.Name = p.t.Text(lang, "app") + " " + item.Name
				}
				item.Value = getValue(item.Type, item.Value)
			}
			return cg
		}
	}
	return cd
}

const (
	monitorConfigRegisterTableName    = "sp_monitor_config_register"
	monitorConfigRegisterInsertUpdate = "INSERT INTO `sp_monitor_config_register`" +
		"(`scope`,`scope_id`,`namespace`,`type`,`names`,`filters`,`enable`,`update_time`,`desc`,`hash`) " +
		"VALUES(?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `update_time`=VALUES(`update_time`),`enable`=VALUES(`enable`)"
	monitorConfigInsertUpdate = "INSERT INTO `sp_monitor_config`" +
		"(`org_id`,`org_name`,`type`,`names`,`filters`,`config`,`create_time`,`update_time`,`enable`,`key`,`hash`) " +
		"VALUES(?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `names`=VALUES(`names`),`filters`=VALUES(`filters`),`config`=VALUES(`config`),`update_time`=VALUES(`update_time`),`enable`=VALUES(`enable`),`key`=VALUES(`key`);"
)

// monitorConfigRegister .
type monitorConfigRegister struct {
	Scope      string    `json:"scope" gorm:"column:scope"`
	ScopeID    string    `json:"scope_id" gorm:"column:scope_id"`
	Namespace  string    `json:"namespace" gorm:"column:namespace"`
	Type       string    `json:"type" gorm:"column:type"`
	Names      string    `json:"names" gorm:"column:names"`
	Filters    string    `json:"filters" gorm:"column:filters"`
	Enable     bool      `json:"enable" gorm:"column:enable"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
	Desc       string    `json:"desc" gorm:"column:desc"`
	Hash       string    `json:"hash" gorm:"column:hash"`
}

// monitorConfig .
type monitorConfig struct {
	OrgID      int       `json:"org_id" gorm:"column:org_id"`
	OrgName    string    `json:"org_name" gorm:"column:org_name"`
	Type       string    `json:"type" gorm:"column:type"`
	Names      string    `json:"names" gorm:"column:names"`
	Filters    string    `json:"filters" gorm:"column:filters"`
	Config     string    `json:"config" gorm:"column:config"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
	Enable     bool      `json:"enable" gorm:"column:enable"`
	Key        string    `json:"key" gorm:"column:key"`
}

func (p *provider) updateMonitorConfig(tx *gorm.DB, orgid int, orgName, ns, group string, keys map[string]interface{}) error {
	if ns == "general" {
		ns = ""
	}
	orgID := strconv.Itoa(orgid)
	key := md5x.SumString(orgID + "/" + ns).String16()
	for k, v := range keys {
		var typ string
		if k == "metrics_ttl" {
			typ = "metric"
		} else if k == "logs_ttl" {
			typ = "log"
		} else {
			continue
		}
		days := conv.ToInt64(v, -1)
		if days < 0 {
			return fmt.Errorf("invalid value %v for key %s", v, k)
		}
		var list []*monitorConfigRegister
		err := tx.Table(monitorConfigRegisterTableName).
			Where("`scope`='org' AND (`scope_id`=? OR `scope_id`='') AND `namespace`=? AND `type`=?", orgID, ns, typ).
			Find(&list).Error
		if err != nil {
			return fmt.Errorf("fail to get register monitor config: %s", err)
		}
		err = p.syncMonitorConfig(tx, orgid, orgID, orgName, ns, typ, key, list, days)
		if err != nil {
			return fmt.Errorf("fail to get sync monitor config: %s", err)
		}
	}
	return nil
}

func (p *provider) syncMonitorConfig(tx *gorm.DB, orgid int, orgID, orgName, env, typ, key string, list []*monitorConfigRegister, days int64) (err error) {
	cfg := getConfigFromDays(days)
	now := time.Now()
	for _, item := range list {
		if len(item.ScopeID) <= 0 {
			item.Filters, err = insertOrgFilter(typ, orgID, orgName, item.Filters)
			if err != nil {
				p.L.Error(err)
				return err
			}
		}
		hash := md5x.SumString(orgID + "," + env + "," + typ + "," + item.Names + item.Filters).String()
		// Update the actual configuration table
		err = tx.Exec(monitorConfigInsertUpdate, orgid, orgName, typ, item.Names, item.Filters, cfg, now, now, 1, key, hash).Error
		if err != nil {
			p.L.Error(err)
			return err
		}
	}
	return nil
}

func insertOrgFilter(typ, orgID, orgName, filters string) (string, error) {
	var fs []*router.KeyValue
	err := json.Unmarshal(reflectx.StringToBytes(filters), &fs)
	if err != nil {
		return filters, fmt.Errorf("invalid monitor filters: %s, %s", filters, err)
	}
	if typ == "metric" {
		fs = append([]*router.KeyValue{
			{
				Key:   "org_name",
				Value: orgName,
			},
		}, fs...)
	} else if typ == "log" {
		fs = append([]*router.KeyValue{
			{
				Key:   "dice_org_name",
				Value: orgName,
			},
		}, fs...)
	}
	byts, _ := json.Marshal(fs)
	return string(byts), nil
}

func getConfigFromDays(days int64) string {
	c := struct {
		TTL string `json:"ttl"`
	}{
		TTL: time.Duration(days * 24 * int64(time.Hour)).String(),
	}
	byts, _ := json.Marshal(&c)
	return string(byts)
}

func (p *provider) registerMonitorConfig(r *http.Request, list []*monitorConfigRegister) interface{} {
	desc := r.FormValue("desc")
	now := time.Now()
	tx := p.db.Begin()

	orgIDs := make(map[string]map[string]map[string][]*monitorConfigRegister)
	for _, item := range list {
		if item == nil {
			continue
		}
		if len(item.Desc) <= 0 {
			item.Desc = desc
		}
		if len(item.Scope) <= 0 || len(item.Type) <= 0 {
			tx.Rollback()
			return api.Errors.InvalidParameter("invalid scope or type")
		}
		item.Namespace = strings.ToLower(item.Namespace)
		item.Hash = md5x.SumString(item.Scope + "," + item.ScopeID + "," + item.Type + "," + item.Namespace + "," + item.Names + item.Filters).String()
		item.UpdateTime = now
		err := tx.Exec(monitorConfigRegisterInsertUpdate, item.Scope, item.ScopeID, item.Namespace, item.Type, item.Names, item.Filters,
			item.Enable, item.UpdateTime, item.Desc, item.Hash).Error
		if err != nil {
			tx.Rollback()
			return api.Errors.Internal(err)
		}
		if item.Scope == "org" {
			envs, ok := orgIDs[item.ScopeID]
			if !ok {
				envs = make(map[string]map[string][]*monitorConfigRegister)
				orgIDs[item.ScopeID] = envs
			}
			typs, ok := envs[item.Namespace]
			if !ok {
				typs = make(map[string][]*monitorConfigRegister)
				envs[item.Namespace] = typs
			}
			typs[item.Type] = append(typs[item.Type], item)
		}
	}
	for orgID, envs := range orgIDs {
		orgid, err := strconv.Atoi(orgID)
		if err != nil {
			continue
		}
		for env, typs := range envs {
			if env == "general" {
				env = ""
			}
			key := md5x.SumString(orgID + "/" + env).String16()
			for typ, list := range typs {
				var ck string
				if typ == "metric" {
					ck = "metrics_ttl"
				} else if typ == "log" {
					ck = "logs_ttl"
				} else {
					continue
				}
				var gs globalSetting
				err := tx.Table(globalSettingTableName).
					Where("`org_id`=? AND `namespace`=? AND `group`='monitor' AND `key`=?", orgid, env, ck).
					Find(&gs).Error
				if err != nil {
					if gorm.IsRecordNotFoundError(err) {
						continue
					}
					tx.Rollback()
					return api.Errors.Internal(err)
				}
				days, err := strconv.ParseInt(gs.Value, 10, 64)
				if err != nil {
					p.L.Errorf("fail to parse metric ttl to int: %s", err)
					continue
				}
				err = p.syncMonitorConfig(tx, orgid, orgID, gs.OrgName, env, typ, key, list, days)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("fail to get sync monitor config: %s", err)
				}
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success("OK")
}
