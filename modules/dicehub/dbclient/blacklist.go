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

// PublishItemBlackList 发布内容黑名单
type PublishItemBlackList struct {
	dbengine.BaseModel
	UserID         string
	PublishItemID  uint64
	PublishItemKey string
	UserName       string
	Operator       string
	DeviceNo       string
}

// TableName 设置模型对应数据库表名称
func (PublishItemBlackList) TableName() string {
	return "dice_publish_items_blacklist"
}

// CreateBlacklist 添加黑名单
func (client *DBClient) CreateBlacklist(blacklist *PublishItemBlackList) error {
	return client.Create(blacklist).Error
}

// CreateBlacklist 移除出黑名单
func (client *DBClient) DeleteBlacklist(blacklist *PublishItemBlackList) error {
	return client.Delete(blacklist).Error
}

// GetBlacklistByID 根据ID查询
func (client *DBClient) GetBlacklistByID(id uint64) (*PublishItemBlackList, error) {
	var blacklist PublishItemBlackList
	if err := client.
		Where("id = ?", id).
		Find(&blacklist).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &blacklist, nil
}

// GetBlacklists 根据publishItemKey查询黑名单
func (client *DBClient) GetBlacklists(pageNo, pageSize, publishItemID uint64) (uint64, *[]PublishItemBlackList, error) {
	var blacklists []PublishItemBlackList
	var total uint64
	query := client.Model(&PublishItemBlackList{}).Where("publish_item_id = ?", publishItemID)
	err := query.Count(&total).Error
	if err != nil {
		return 0, nil, err
	}
	err = query.Order("created_at desc").
		Offset((pageNo - 1) * pageSize).
		Limit(pageSize).Find(&blacklists).Error
	if err != nil {
		return 0, nil, err
	}
	return total, &blacklists, nil
}

// GetBlacklistByUserIDAndDeviceNo 根据用户ID，设备号，publishItemKey查询
func (client *DBClient) GetBlacklistByUserID(userID string, publishItemID uint64) ([]*PublishItemBlackList, error) {
	var blacklist []*PublishItemBlackList
	query := client.Model(&PublishItemBlackList{})
	query = query.Where("publish_item_id = ?", publishItemID)
	query = query.Where("user_id = ?", userID)

	if err := query.Find(&blacklist).Error; err != nil {
		return nil, err
	}
	return blacklist, nil
}

// GetBlacklistByDeviceNo 根据设备号，publishItemKey查询
func (client *DBClient) GetBlacklistByDeviceNo(publishItemID uint64, deviceNo string) ([]*PublishItemBlackList, error) {
	var blacklist []*PublishItemBlackList
	query := client.Model(&PublishItemBlackList{})
	query = query.Where("publish_item_id = ?", publishItemID)
	query = query.Where("device_no = ?", deviceNo)

	if err := query.Find(&blacklist).Error; err != nil {
		return nil, err
	}
	return blacklist, nil
}
