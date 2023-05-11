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

package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/pkg/strutil"
)

// BaseModel common info for all models
type BaseModel struct {
	ID        int64     `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Org 企业资源模型
type Org struct {
	BaseModel
	Name           string
	DisplayName    string
	Desc           string
	Logo           string
	Locale         string
	OpenFdp        bool   `json:"openFdp"`
	UserID         string `gorm:"column:creator"` // 所属用户Id
	Config         OrgConfig
	BlockoutConfig BlockoutConfig
	Type           string
	Status         string // TODO deprecated 待admin下线后删除
	IsPublic       bool
}

// TableName 设置模型对应数据库表名称
func (Org) TableName() string {
	return "dice_org"
}

type BlockoutConfig struct {
	BlockDEV   bool `json:"blockDev"`
	BlockTEST  bool `json:"blockTest"`
	BlockStage bool `json:"blockStage"`
	BlockProd  bool `json:"blockProd"`
}

type OrgConfig struct {
	EnablePersonalMessageEmail bool   `json:"enablePersonalMessageEmail"`
	EnableMS                   bool   `json:"enableMs"`
	SMTPHost                   string `json:"smtpHost"`
	SMTPUser                   string `json:"smtpUser"`
	SMTPPassword               string `json:"smtpPassword"`
	SMTPPort                   int64  `json:"smtpPort"`
	SMTPIsSSL                  bool   `json:"smtpIsSSL"`
	SMSKeyID                   string `json:"smsKeyID"`
	SMSKeySecret               string `json:"smsKeySecret"`
	SMSSignName                string `json:"smsSignName"`
	SMSMonitorTemplateCode     string `json:"smsMonitorTemplateCode"` // 监控单独的短信模版
	VMSKeyID                   string `json:"vmsKeyID"`
	VMSKeySecret               string `json:"vmsKeySecret"`
	VMSMonitorTtsCode          string `json:"vmsMonitorTtsCode"`          // 监控单独的语音模版
	VMSMonitorCalledShowNumber string `json:"vmsMonitorCalledShowNumber"` // 监控单独的被叫显号
	AuditInterval              int64  `json:"auditInterval"`

	// 开关：制品是否可以跨集群部署
	EnableReleaseCrossCluster bool `json:"enableReleaseCrossCluster"`
	EnableAI                  bool `json:"enableAI"`
}

func (cfg OrgConfig) Value() (driver.Value, error) {
	if b, err := json.Marshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to marshal orgConfig, err: %v", err)
	} else {
		return string(b), nil
	}
}

func (cfg *OrgConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return errors.New("invalid scan source for OrgConfig")
	}
	if len(v) == 0 {
		return nil
	}
	if err := json.Unmarshal(v, cfg); err != nil {
		return fmt.Errorf("failed to unmarshal orgConfig, err: %v", err)
	}
	return nil
}

func (cfg BlockoutConfig) Value() (driver.Value, error) {
	b, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal BlockoutConfig, err: %v", err)
	}
	return string(b), nil
}

func (cfg *BlockoutConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return errors.New("invalid scan source for BlockoutConfig")
	}
	if len(v) == 0 {
		return nil
	}
	if err := json.Unmarshal(v, cfg); err != nil {
		return fmt.Errorf("failed to unmarshal BlockoutConfig, err: %v", err)
	}
	return nil
}

// CreateOrg 创建企业
func (client *DBClient) CreateOrg(org *Org) error {
	return client.Create(org).Error
}

// UpdateOrg 更新企业元信息，不可更改企业名称
func (client *DBClient) UpdateOrg(org *Org) error {
	return client.Save(org).Error
}

// DeleteOrg 删除企业
func (client *DBClient) DeleteOrg(orgID int64) error {
	return client.Where("id = ?", orgID).Delete(&Org{}).Error
}

// GetOrg 根据orgID获取企业信息
func (client *DBClient) GetOrg(orgID int64) (Org, error) {
	var org Org
	if err := client.Where("id = ?", orgID).Find(&org).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return org, ErrNotFoundOrg
		}
		return org, err
	}
	fmt.Printf("----------------------------------org: %+v-----------------------", org)
	return org, nil
}

// GetOrgByName 根据 orgName 获取企业信息
func (client *DBClient) GetOrgByName(orgName string) (*Org, error) {
	var org Org
	if err := client.Where("name = ?", orgName).First(&org).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrNotFoundOrg
		}
		return nil, err
	}
	return &org, nil
}

// GetOrgsByParam 获取企业列表
func (client *DBClient) GetOrgsByParam(name string, pageNum, pageSize int) (int, []Org, error) {
	var (
		orgs  []Org
		total int
	)
	if name == "" {
		if err := client.Order("updated_at DESC").Offset((pageNum - 1) * pageSize).Limit(pageSize).
			Find(&orgs).Error; err != nil {
			return 0, nil, err
		}
		// 获取总量
		if err := client.Model(&Org{}).Count(&total).Error; err != nil {
			return 0, nil, err
		}
	} else {
		if err := client.Where("name LIKE ?", "%"+name+"%").Or("display_name LIKE ?", "%"+name+"%").Order("updated_at DESC").
			Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&orgs).Error; err != nil {
			return 0, nil, err
		}
		// 获取总量
		if err := client.Model(&Org{}).Where("name LIKE ?", "%"+name+"%").Or("display_name LIKE ?", "%"+name+"%").
			Count(&total).Error; err != nil {
			return 0, nil, err
		}
	}
	return total, orgs, nil
}

// GetPublicOrgsByParam Get public orgs list
func (client *DBClient) GetPublicOrgsByParam(name string, pageNum, pageSize int) (int, []Org, error) {
	var (
		orgs  []Org
		total int
	)
	if err := client.Where("is_public = ?", 1).Where("name LIKE ? OR display_name LIKE ?", "%"+name+"%", "%"+name+"%").Order("updated_at DESC").
		Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&orgs).Error; err != nil {
		return 0, nil, err
	}
	if err := client.Model(&Org{}).Where("is_public = ?", 1).Where("name LIKE ? OR display_name LIKE ?", "%"+name+"%", "%"+name+"%").
		Count(&total).Error; err != nil {
		return 0, nil, err
	}
	return total, orgs, nil
}

// GetOrgsByUserID 根据userID获取企业列表
func (client *DBClient) GetOrgsByUserID(userID string) ([]Org, error) {
	var orgs []Org
	if err := client.Where("creator = ?", userID).Order("updated_at DESC").
		Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

// GetOrgsByIDsAndName 根据企业ID列表 & 企业名称获取企业列表
func (client *DBClient) GetOrgsByIDsAndName(orgIDs []int64, name string, pageNo, pageSize int) (
	int, []Org, error) {
	var (
		total int
		orgs  []Org
	)
	if err := client.Where("id in (?)", orgIDs).
		Where("name LIKE ? OR display_name LIKE ?", strutil.Concat("%", name, "%"), strutil.Concat("%", name, "%")).Order("updated_at DESC").
		Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&orgs).Error; err != nil {
		return 0, nil, err
	}
	// 获取总量
	if err := client.Model(&Org{}).Where("id in (?)", orgIDs).
		Where("name LIKE ? OR display_name LIKE ?", strutil.Concat("%", name, "%"), strutil.Concat("%", name, "%")).Count(&total).Error; err != nil {
		return 0, nil, err
	}
	return total, orgs, nil
}

// GetOrgList 获取所有企业列表(仅供内部用户使用)
func (client *DBClient) GetOrgList() ([]Org, error) {
	var orgs []Org
	if err := client.Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}
