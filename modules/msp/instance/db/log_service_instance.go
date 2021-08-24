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

import "github.com/jinzhu/gorm"

type LogServiceInstanceDB struct {
	*gorm.DB
}

func (db *LogServiceInstanceDB) AddOrUpdateEsUrls(instanceId, esUrls, esConfig string) error {
	var instance LogServiceInstance
	resp := db.Where("instance_id=?", instanceId).Limit(1).Find(&instance)

	if resp.RecordNotFound() {
		instance = LogServiceInstance{InstanceID: instanceId, EsUrls: esUrls, EsConfig: esConfig}
		if db.Dialect().GetName() == "mysql" {
			return db.Set("gorm:insert_modifier", "IGNORE").Save(&instance).Error
		}
		return db.Save(&instance).Error
	}

	if instance.EsUrls == esUrls && instance.EsConfig == esConfig {
		return nil
	}
	instance.EsUrls = esUrls
	instance.EsConfig = esConfig

	return db.Save(&instance).Error
}

func (db *LogServiceInstanceDB) GetFirst() (*LogServiceInstance, error) {
	var instance LogServiceInstance

	resp := db.First(&instance)

	if resp.RecordNotFound() {
		return nil, nil
	}

	return &instance, resp.Error
}
