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

	"github.com/erda-project/erda/modules/dop/model"
)

// UpdateOrAddEnvConfig 更新环境变量配置
func (client *DBClient) UpdateOrAddEnvConfig(config *model.ConfigItem) error {
	return client.Save(config).Error
}

// UpdateEnvConfig 更新环境变量配置
func (client *DBClient) SoftDeleteEnvConfig(config *model.ConfigItem) error {
	return client.Model(config).Update("is_deleted", IsDeleteValue).Error
}

// GetEnvConfigByID 根据 ID 获取 配置信息
func (client *DBClient) GetEnvConfigByID(configID int64) (*model.ConfigItem, error) {
	configItem := &model.ConfigItem{}
	if err := client.Where("id = ?", configID).Where("is_deleted = ?", "N").
		First(configItem).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return configItem, nil
}

// GetEnvConfigsByNamespaceID 根据 namespaceID 获取所有配置信息
func (client *DBClient) GetEnvConfigsByNamespaceID(namespaceID int64) ([]model.ConfigItem, error) {
	var configItems []model.ConfigItem
	if err := client.Where("namespace_id = ?", namespaceID).Where("is_deleted = ?", "N").
		Find(&configItems).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return configItems, nil
}

// GetEnvConfigByKey 根据 namespaceID, key 获取某个配置信息
func (client *DBClient) GetEnvConfigByKey(namespaceID int64, key string) (*model.ConfigItem, error) {
	configItem := &model.ConfigItem{}
	if err := client.Where("namespace_id = ?", namespaceID).Where("is_deleted = ?", "N").
		Where("item_key = ?", key).Find(configItem).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return configItem, nil
}
