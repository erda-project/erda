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

// AddonInstanceRelation 存储addon实例依赖关系
type AddonInstanceRelation struct {
	ID                string    `gorm:"type:varchar(64)"` // 唯一Id
	OutsideInstanceID string    `gorm:"type:varchar(64)"` // addon实例Id
	InsideInstanceID  string    `gorm:"type:varchar(32)"` // addon实例依赖Id
	Deleted           string    `gorm:"column:is_deleted"`
	CreatedAt         time.Time `gorm:"column:create_time"`
	UpdatedAt         time.Time `gorm:"column:update_time"`
}

// TableName 数据库表名
func (AddonInstanceRelation) TableName() string {
	return "tb_middle_instance_relation"
}

// CreateAddonInstanceRelation insert AddonInstanceRelation
func (db *DBClient) CreateAddonInstanceRelation(addonInstanceRelation *AddonInstanceRelation) error {
	return db.Create(addonInstanceRelation).Error
}

// UpdateAddonInstanceRelation update AddonInstanceRelation
func (db *DBClient) UpdateAddonInstanceRelation(addonInstanceRelation *AddonInstanceRelation) error {
	if err := db.
		Save(addonInstanceRelation).Error; err != nil {
		return errors.Wrapf(err, "failed to update addonInstanceRelation info, id: %v", addonInstanceRelation.ID)
	}
	return nil
}

// GetByInstanceIDAndField 根据addonName、field获取AddonExtra信息
func (db *DBClient) GetByOutSideInstanceID(instanceID string) (*[]AddonInstanceRelation, error) {
	var addonInstanceRelations []AddonInstanceRelation
	if err := db.
		Where("outside_instance_id = ?", instanceID).
		Where("is_deleted = ?", apistructs.AddonNotDeleted).
		Find(&addonInstanceRelations).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get addon addonInstanceRelations info, instanceID : %s",
			instanceID)
	}
	return &addonInstanceRelations, nil
}

// GetByInSideInstanceID 根据addonName、field获取AddonExtra信息
func (db *DBClient) GetByInSideInstanceID(instanceID string) (*AddonInstanceRelation, error) {
	var addonInstanceRelations AddonInstanceRelation
	if err := db.
		Where("inside_instance_id = ?", instanceID).
		Where("is_deleted = ?", apistructs.AddonNotDeleted).
		First(&addonInstanceRelations).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get addon addonInstanceRelations info, instanceID : %s",
			instanceID)
	}
	return &addonInstanceRelations, nil
}
