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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/dbengine"
)

// PublishItemH5Targets H5的目标版本记录表
type PublishItemH5Targets struct {
	dbengine.BaseModel
	H5VersionID      uint64 `gorm:"column:h5_version_id"`
	TargetVersion    string `gorm:"column:target_version"`
	TargetBuildID    string `gorm:"column:target_build_id"`
	TargetMobileType string `gorm:"column:target_mobile_type"`
}

// TableName 设置模型对应数据库表名称
func (PublishItemH5Targets) TableName() string {
	return "dice_publish_item_h5_targets"
}

// CreateH5Targets 创建h5目标app版本关系
func (client *DBClient) CreateH5Targets(target *PublishItemH5Targets) error {
	return client.Create(target).Error
}

// GetH5VersionsByTarget 根据移动应用版本获取对应的H5版本
func (client *DBClient) GetH5VersionsByTarget(itemID uint64, mobileType apistructs.ResourceType, appVersion, packageName string) ([]*PublishItemVersion, error) {
	var h5versions []*PublishItemVersion
	if err := client.Table("dice_publish_item_versions").Joins("inner join dice_publish_item_h5_targets on dice_publish_item_versions.id = dice_publish_item_h5_targets.h5_version_id").
		Where("dice_publish_item_h5_targets.target_version = ?", appVersion).
		Where("dice_publish_item_h5_targets.target_mobile_type = ?", mobileType).
		Where("dice_publish_item_versions.publish_item_id = ?", itemID).
		Where("dice_publish_item_versions.package_name = ?", packageName).
		Scan(&h5versions).Error; err != nil {
		return nil, err
	}

	return h5versions, nil
}

// GetTargetsByH5Version 返回H5包版本的目标版本信息
func (client *DBClient) GetTargetsByH5Version(versionID uint64) ([]PublishItemH5Targets, error) {
	var targetRelations []PublishItemH5Targets
	if err := client.Table("dice_publish_item_h5_targets").Where("h5_version_id = ?", versionID).
		Find(&targetRelations).Error; err != nil {
		return nil, err
	}

	return targetRelations, nil
}
