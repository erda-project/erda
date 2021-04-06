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

// AddonInstanceExtra 存储addon实例额外信息
type AddonInstanceExtra struct {
	ID         string    `gorm:"type:varchar(64)"` // 唯一Id
	InstanceID string    `gorm:"type:varchar(64)"` // addon名称
	Field      string    `gorm:"type:varchar(32)"` // 属性名称
	Value      string    `gorm:"type:text"`        // 属性值
	Deleted    string    `gorm:"column:is_deleted"`
	CreatedAt  time.Time `gorm:"column:create_time"`
	UpdatedAt  time.Time `gorm:"column:update_time"`
}

// TableName 数据库表名
func (AddonInstanceExtra) TableName() string {
	return "tb_middle_instance_extra"
}

// CreateAddonInstanceExtra insert AddonExtra
func (db *DBClient) CreateAddonInstanceExtra(addonInstanceExtra *AddonInstanceExtra) error {
	return db.Create(addonInstanceExtra).Error
}

// UpdateAddonInstanceExtra update AddonInstanceExtra
func (db *DBClient) UpdateAddonInstanceExtra(addonInstanceExtra *AddonInstanceExtra) error {
	if err := db.
		Save(addonInstanceExtra).Error; err != nil {
		return errors.Wrapf(err, "failed to update addonInstanceExtra info, id: %v", addonInstanceExtra.ID)
	}
	return nil
}

// GetByInstanceIDAndField 根据addonName、field获取AddonExtra信息
func (db *DBClient) GetByInstanceIDAndField(instanceID, field string) (*AddonInstanceExtra, error) {
	var addonInstanceExtra AddonInstanceExtra
	if err := db.
		Where("instance_id = ?", instanceID).
		Where("field = ?", field).
		Where("is_deleted = ?", apistructs.AddonNotDeleted).
		First(&addonInstanceExtra).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get addon addonInstanceExtra info, instanceID : %s, field : %s",
			instanceID, field)
	}
	return &addonInstanceExtra, nil
}

// GetByAddonName 根据addonName获取AddonExtra信息
func (db *DBClient) GetInstanceExtraByAddonName(addonName string) (*[]AddonExtra, error) {
	var addonExtras []AddonExtra
	if err := db.
		Where("addon_name = ?", addonName).
		Where("is_deleted = ?", apistructs.AddonNotDeleted).
		Find(&addonExtras).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get addon addonExtras info, addon_name : %s",
			addonName)
	}
	return &addonExtras, nil
}
