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

	"github.com/erda-project/erda/apistructs"
)

// AddonNode addon node信息
type AddonNode struct {
	ID         string `gorm:"type:varchar(64)"`
	InstanceID string `gorm:"type:varchar(64)"` // AddonInstance 主键
	Namespace  string `gorm:"type:text"`
	NodeName   string
	CPU        float64
	Mem        uint64
	Deleted    string    `gorm:"column:is_deleted"` // Y: 已删除 N: 未删除
	CreatedAt  time.Time `gorm:"column:create_time"`
	UpdatedAt  time.Time `gorm:"column:update_time"`
}

// TableName 数据库表名
func (AddonNode) TableName() string {
	return "tb_middle_node"
}

// CreateAddonNode insert addonNode
func (db *DBClient) CreateAddonNode(addonNode *AddonNode) error {
	return db.Create(addonNode).Error
}

// GetAddonNodesByInstanceID 根据instanceID获取addonNode信息
func (db *DBClient) GetAddonNodesByInstanceID(instanceID string) (*[]AddonNode, error) {
	// 获取匹配搜索结果总量
	var addonNodes []AddonNode
	if err := db.
		Where("instance_id = ?", instanceID).
		Where("is_deleted = ?", apistructs.AddonNotDeleted).
		Find(&addonNodes).Error; err != nil {
		return nil, err
	}
	return &addonNodes, nil
}

// GetAddonNodesByInstanceIDs 根据instanceID列表获取addonNode信息
func (db *DBClient) GetAddonNodesByInstanceIDs(instanceIDs []string) (*[]AddonNode, error) {
	// 获取匹配搜索结果总量
	var addonNodes []AddonNode
	if err := db.
		Where("instance_id in (?)", instanceIDs).
		Where("is_deleted = ?", apistructs.AddonNotDeleted).
		Find(&addonNodes).Error; err != nil {
		return nil, err
	}
	return &addonNodes, nil
}
