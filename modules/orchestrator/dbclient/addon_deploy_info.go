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

package dbclient

import (
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
)

// AddonDeploy 平台组件发布信息
type AddonDeploy struct {
	ID           uint64    `gorm:"primary_key"`      // 唯一Id
	AddonName    string    `gorm:"type:varchar(64)"` // addon名称
	Version      string    `gorm:"type:varchar(32)"` // 版本
	DeployStatus string    `gorm:"type:varchar(32)"` // 发布状态
	DeployMode   string    `gorm:"type:varchar(32)"` // 发布方式，upgrade或rollback
	Deleted      string    `gorm:"column:is_deleted"`
	CreatedAt    time.Time `gorm:"column:create_time"`
	UpdatedAt    time.Time `gorm:"column:update_time"`
}

// TableName 数据库表名
func (AddonDeploy) TableName() string {
	return "tb_addon_deploy_info"
}

// CreateAddonDeploy insert AddonDeploy
func (db *DBClient) CreateAddonDeploy(addonDeploy *AddonDeploy) error {
	return db.Create(addonDeploy).Error
}

// UpdateAddonDeploy update AddonDeploy
func (db *DBClient) UpdateAddonDeploy(addonDeploy *AddonDeploy) error {
	if err := db.Save(addonDeploy).Error; err != nil {
		return errors.Wrapf(err, "failed to update addonDeploy info, id: %v", addonDeploy.ID)
	}
	return nil
}

// GetDeployByAddonName 根据addonName获取AddonDeploy信息
func (db *DBClient) GetDeployByAddonName(addonName string) (*[]AddonDeploy, error) {
	var addonDeploys []AddonDeploy
	if err := db.
		Where("addon_name = ?", addonName).
		Where("is_deleted = ?", apistructs.AddonNotDeleted).
		First(&addonDeploys).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get addon deploys info, addon_name : %s",
			addonName)
	}
	return &addonDeploys, nil
}

// GetDeployById 根据id获取addonDeploy信息
func (db *DBClient) GetDeployById(id int64) (*AddonDeploy, error) {
	var addonDeploy AddonDeploy
	if err := db.
		Where("id = ?", id).
		Where("is_deleted = ?", apistructs.AddonNotDeleted).
		First(&addonDeploy).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get addon deploy info, id : %d",
			id)
	}
	return &addonDeploy, nil
}
