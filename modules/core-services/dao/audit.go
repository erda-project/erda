// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dao

import (
	"encoding/json"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/model"
)

// CreateAudit 创建审计
func (client *DBClient) CreateAudit(audit *model.Audit) error {
	return client.Create(audit).Error
}

// BatchCreateAudit 批量传教审计
func (client *DBClient) BatchCreateAudit(audits []model.Audit) error {
	return client.BulkInsert(audits)
}

// GetAuditsByParam 通过参数查询成员
func (client *DBClient) GetAuditsByParam(param *apistructs.AuditsListRequest) (int, []model.Audit, error) {
	var audits []model.Audit
	var total int
	db := client.Table("dice_audit").Where("deleted = 0").Where("start_time >= ? AND start_time <= ?", param.StartAt, param.EndAt)

	if !param.Sys {
		db = db.Where("org_id = ?", param.OrgID)
	} else {
		db = db.Where("scope_type = 'sys'")
	}

	if param.UserID != nil {
		db = db.Where("user_id in ( ? )", param.UserID)
	}

	if param.FDPProjectID != "" {
		db = db.Where("fdp_project_id = ?", param.FDPProjectID)
	} else {
		db = db.Where("fdp_project_id = ? or fdp_project_id is NULL", "")
	}

	if err := db.Order("start_time DESC").Offset((param.PageNo - 1) * param.PageSize).Limit(param.PageSize).
		Find(&audits).Offset(0).Limit(-1).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	return total, audits, nil
}

// GetAuditSettings 从 dice_org 获取审计事件清理周期
func (client *DBClient) GetAuditSettings() ([]model.AuditSettings, error) {
	var auditSettings []model.AuditSettings
	if err := client.Table("dice_org").Select("id, config").Find(&auditSettings).Error; err != nil {
		return nil, err
	}

	return auditSettings, nil
}

// DeleteAuditsByTimeAndOrg 软删除某个企业的审计事件
func (client *DBClient) DeleteAuditsByTimeAndOrg(startTime time.Time, orgIDs []uint64) error {
	// var audit model.Audit
	return client.Table("dice_audit").Where("org_id in ( ? )", orgIDs).Where("start_time <= ?", startTime).
		Where("scope_type != 'sys'").Update("deleted", "1").Error
}

// DeleteAuditsByTimeAndSys 软删除系统级别的审计事件
func (client *DBClient) DeleteAuditsByTimeAndSys(startTime time.Time) error {
	// var audit model.Audit
	return client.Table("dice_audit").Where("start_time <= ?", startTime).Where("scope_type = 'sys'").
		Update("deleted", "1").Error
}

// ArchiveAuditsByTimeAndOrg 归档某个企业的审计事件
func (client *DBClient) ArchiveAuditsByTimeAndOrg() error {
	// 在审计历史表创建
	if err := client.Table("dice_audit_history").
		Exec("INSERT INTO `dice_audit_history` SELECT * FROM `dice_audit` Where deleted = '1'").Error; err != nil {
		return err
	}
	// 删除审计表数据
	if err := client.Table("dice_audit").Where("deleted = '1'").Delete(model.Audit{}).Error; err != nil {
		return err
	}

	return nil
}

// InitOrgAuditInterval 初始化企业的审计事件清理周期
func (client *DBClient) InitOrgAuditInterval(orgIDs []uint64) error {
	var orgs []model.Org
	if err := client.Table("dice_org").Where("id in ( ? )", orgIDs).Find(&orgs).Error; err != nil {
		return err
	}

	for _, v := range orgs {
		config := &v.Config
		config.AuditInterval = -7
		cfg, err := json.Marshal(config)
		if err != nil {
			return err
		}
		if err := client.Table("dice_org").Where("id = ?", v.ID).Update("config", string(cfg)).Error; err != nil {
			return err
		}
	}

	return nil
}

// UpdateAuditCleanCron 修改企业事件清理周期
func (client *DBClient) UpdateAuditCleanCron(orgID, interval int64) error {
	var org model.Org
	if err := client.Table("dice_org").Where("id = ( ? )", orgID).Find(&org).Error; err != nil {
		return err
	}
	config := &org.Config
	config.AuditInterval = interval
	cfg, err := json.Marshal(config)
	if err != nil {
		return err
	}
	if err := client.Table("dice_org").Where("id = ?", org.ID).Update("config", string(cfg)).Error; err != nil {
		return err
	}

	return nil
}

// GetAuditCleanCron 获取企业事件清理周期
func (client *DBClient) GetAuditCleanCron(orgID int64) (*model.Org, error) {
	var org model.Org
	if err := client.Table("dice_org").Where("id = ( ? )", orgID).Find(&org).Error; err != nil {
		return nil, err
	}

	return &org, nil
}
