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
	"encoding/json"
	"errors"

	"github.com/jinzhu/gorm"
)

type InstanceTenantDb struct {
	*gorm.DB
}

func (db *InstanceTenantDb) QueryTkByTenantGroup(tenantGroup string) (string, error) {
	var tenantInfo InstanceTenant
	err := db.
		Select("*").
		Where("tenant_group = ?", tenantGroup).
		Where("engine = ?", "monitor").
		Where("is_deleted = ?", "N").
		Order("create_time", false).
		Limit(1).
		Find(&tenantInfo).
		Error
	if err != nil {
		return "", err
	}
	var config map[string]interface{}
	err = json.Unmarshal([]byte(tenantInfo.Config), &config)
	if err != nil {
		return "", err
	}
	tk := config["TERMINUS_KEY"]
	if tk == nil {
		return "", errors.New("tk not exist")
	}
	return tk.(string), nil
}
