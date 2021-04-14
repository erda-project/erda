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

package db

import (
	"github.com/jinzhu/gorm"
)

type NotifyConfig struct {
	gorm.Model
	NotifyID  string `gorm:"column:notify_id"`
	Metadata  string `gorm:"column:metadata"`
	Behavior  string `gorm:"column:behavior"`
	Templates string `gorm:"column:templates"`
	Scope     string `gorm:"column:scope"`
	ScopeID   string `gorm:"column:scope_id"`
}

func (c *NotifyConfig) TableName() string {
	return "sp_notify_user_define"
}

func (n *NotifyDB) GetAllUserDefineNotify(scope, scopeID string) (*[]NotifyConfig, error) {
	allNotifies := make([]NotifyConfig, 0)
	err := n.DB.Model(&allNotifies).Where("scope = ?", scope).
		Where("scope_id = ?", scopeID).Find(&allNotifies).Error
	return &allNotifies, err
}

func (n *NotifyDB) CreateUserDefineNotifyTemplate(customize *NotifyConfig) error {
	err := n.DB.Create(customize).Error
	return err
}

func (n *NotifyDB) GetUserDefine(notifyId string) (*NotifyConfig, error) {
	var userDefine NotifyConfig
	err := n.DB.Model(&NotifyConfig{}).Where("notify_id = ?", notifyId).Find(&userDefine).Error
	return &userDefine, err
}

func (n *NotifyDB) GetAllUserDefineTemplates() (*[]NotifyConfig, error) {
	allTemplates := make([]NotifyConfig, 0)
	err := n.DB.Model(&NotifyConfig{}).Find(&allTemplates).Error
	return &allTemplates, err
}
