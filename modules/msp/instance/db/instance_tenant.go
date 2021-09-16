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

	"github.com/erda-project/erda/pkg/database/gormutil"
)

// InstanceTenantDB .
type InstanceTenantDB struct {
	*gorm.DB
}

func (db *InstanceTenantDB) query() *gorm.DB {
	return db.Table(TableInstanceTenant).Where("`is_deleted`='N'")
}

func (db *InstanceTenantDB) GetByFields(fields map[string]interface{}) (*InstanceTenant, error) {
	query := db.query()
	query, err := gormutil.GetQueryFilterByFields(query, instanceTenantFieldColumns, fields)
	if err != nil {
		return nil, err
	}
	var list []*InstanceTenant
	if err := query.Limit(1).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) <= 0 {
		return nil, nil
	}
	return list[0], nil
}

func (db *InstanceTenantDB) GetByID(id string) (*InstanceTenant, error) {
	return db.GetByFields(map[string]interface{}{
		"ID": id,
	})
}

func (db *InstanceTenantDB) GetByTenantGroup(group string) ([]*InstanceTenant, error) {
	if len(group) <= 0 {
		return nil, nil
	}
	var list []*InstanceTenant
	if err := db.query().
		Where("tenant_group=?", group).Order("update_time DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (db *InstanceTenantDB) GetClusterNameByTenantGroup(group string) (string, error) {
	if len(group) <= 0 {
		return "", nil
	}
	var list []*InstanceTenant
	if err := db.query().
		Where("tenant_group=?", group).Limit(1).Find(&list).Error; err != nil {
		return "", err
	}
	if len(list) <= 0 {
		return "", nil
	}
	return list[0].Az, nil
}

func (db *InstanceTenantDB) GetByEngineAndTenantGroup(engine string, tenantGroup string) (*InstanceTenant, error) {
	return db.GetByFields(map[string]interface{}{
		"Engine":      engine,
		"TenantGroup": tenantGroup,
	})
}
