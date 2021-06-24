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

package dao

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/modules/core-services/model"
)

// CreateFavoritedResource 创建收藏关系
func (client *DBClient) CreateFavoritedResource(resource *model.FavoritedResource) error {
	return client.Create(resource).Error
}

// GetFavoritedResource 查询收藏关系
func (client *DBClient) GetFavoritedResource(target string, targetID uint64, userID string) (*model.FavoritedResource, error) {
	var resource model.FavoritedResource
	if err := client.Where("target = ?", target).
		Where("target_id = ?", targetID).
		Where("user_id = ?", userID).Find(&resource).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &resource, nil
}

// GetFavoritedResourcesByUser 查询收藏关系
func (client *DBClient) GetFavoritedResourcesByUser(target, userID string) ([]model.FavoritedResource, error) {
	var resources []model.FavoritedResource
	if err := client.Where("target = ?", target).
		Where("user_id = ?", userID).Find(&resources).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return resources, nil
}

// DeleteFavoritedResource 删除收藏关系
func (client *DBClient) DeleteFavoritedResource(id uint64) error {
	return client.Where("id = ?", id).Delete(&model.FavoritedResource{}).Error
}

// DeleteFavoritedResourcesByTarget 根据 target 删除收藏关系
func (client *DBClient) DeleteFavoritedResourcesByTarget(target string, targetID uint64) error {
	return client.Where("target = ?", target).Where("target_id = ?", targetID).Delete(&model.FavoritedResource{}).Error
}
