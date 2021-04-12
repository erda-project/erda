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
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/pkg/dbengine"
)

// PublishItemErase 数据擦除
type PublishItemErase struct {
	dbengine.BaseModel
	PublishItemID  uint64
	PublishItemKey string
	DeviceNo       string
	Operator       string
	EraseStatus    string
}

// TableName 设置模型对应数据库表名称
func (PublishItemErase) TableName() string {
	return "dice_publish_items_erase"
}

// CreateErase 添加数据擦除
func (client *DBClient) CreateErase(erase *PublishItemErase) error {
	return client.Create(erase).Error
}

// UpdateErase 更新数据擦除状态
func (client *DBClient) UpdateErase(erase *PublishItemErase) error {
	return client.Save(erase).Error
}

// DeleteErase 移除数据擦除
func (client *DBClient) DeleteErase(erase *PublishItemErase) error {
	return client.Delete(erase).Error
}

// GetErases 根据artifactID查询数据擦除
func (client *DBClient) GetErases(pageNo, pageSize, artifactID uint64) (uint64, *[]PublishItemErase, error) {
	var erases []PublishItemErase
	var total uint64
	query := client.Model(&PublishItemErase{}).Where("publish_item_id = ?", artifactID)
	err := query.Count(&total).Error
	if err != nil {
		return 0, nil, err
	}
	err = query.Order("created_at desc").
		Offset((pageNo - 1) * pageSize).
		Limit(pageSize).Find(&erases).Error
	if err != nil {
		return 0, nil, err
	}
	return total, &erases, nil
}

// GetEraseByDeviceNo 根据设备号，publishItemID查询
func (client *DBClient) GetEraseByDeviceNo(publishItemID uint64, deviceNo string) (*PublishItemErase, error) {
	var erase PublishItemErase
	if err := client.
		Where("publish_item_id = ?", publishItemID).
		Where("device_no = ?", deviceNo).
		Find(&erase).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &erase, nil
}
