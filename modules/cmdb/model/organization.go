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

package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

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
