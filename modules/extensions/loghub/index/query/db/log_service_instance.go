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

import (
	"github.com/jinzhu/gorm"
)

// LogServiceInstanceTable .
const LogServiceInstanceTable = "sp_log_service_instance"

// LogServiceInstance .
type LogServiceInstance struct {
	ID         int    `gorm:"column:id;primary_key"`
	InstanceID string `gorm:"column:instance_id"`
	EsUrls     string `gorm:"column:es_urls"`
	EsConfig   string `gorm:"column:es_config"`
}

// TableName .
func (LogServiceInstance) TableName() string { return LogServiceInstanceTable }

// LogServiceInstanceDB .
type LogServiceInstanceDB struct {
	*gorm.DB
}

func (db *LogServiceInstanceDB) GetFirst() (*LogServiceInstance, error) {
	var instance LogServiceInstance

	resp := db.First(&instance)

	if resp.RecordNotFound() {
		return nil, nil
	}

	return &instance, resp.Error
}
