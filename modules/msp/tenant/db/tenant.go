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

	"github.com/erda-project/erda/pkg/common/errors"
)

// MSPTenantDB msp_tenant
type MSPTenantDB struct {
	*gorm.DB
}

func (db *MSPTenantDB) db() *gorm.DB {
	return db.Table(TableMSPTenant)
}

func (db *MSPTenantDB) InsertTenant(tenant *MSPTenant) (*MSPTenant, *errors.DatabaseError) {
	result := db.db().Create(tenant)

	if result.Error != nil {
		return nil, errors.NewDatabaseError(result.Error)
	}
	value := result.Value.(*MSPTenant)
	return value, nil
}

func (db *MSPTenantDB) QueryTenant(tenantID string) (*MSPTenant, error) {
	tenant := MSPTenant{}
	err := db.db().Where("`id` = ?", tenantID).Find(&tenant).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (db *MSPTenantDB) QueryTenantByProjectIDAndWorkspace(projectID, workspace string) (*MSPTenant, error) {
	tenant := MSPTenant{}
	err := db.db().Where("`related_project_id` = ?", projectID).Where("`related_workspace` = ?", workspace).Find(&tenant).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (db *MSPTenantDB) QueryTenantByProjectID(projectID string) ([]*MSPTenant, error) {
	var tenants []*MSPTenant
	err := db.db().Where("`related_project_id` = ?", projectID).Find(&tenants).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return tenants, nil
}

func (db *MSPTenantDB) DeleteTenantByTenantID(tenantId string) (*MSPTenant, error) {
	tenant, err := db.QueryTenant(tenantId)
	if err != nil {
		return nil, err
	}
	tenant.IsDeleted = true
	err = db.Model(&tenant).Update(&tenant).Error
	if err != nil {
		return nil, err
	}
	return tenant, err
}
