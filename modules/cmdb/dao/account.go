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

	"github.com/erda-project/erda/modules/cmdb/model"
)

// CreateCloudAccount 创建云账号
func (client *DBClient) CreateAccount(account *model.CloudAccount) error {
	return client.Create(account).Error
}

// UpdateCloudAccount 更新云账号
func (client *DBClient) UpdateAccount(account *model.CloudAccount) error {
	return client.Save(account).Error
}

// DeleteCloudAccount 删除云账号
func (client *DBClient) DeleteAccount(orgID, accountID int64) error {
	return client.Where("org_id = ?", orgID).Where("id = ?", accountID).Delete(&model.CloudAccount{}).Error
}

// GetAccountsByOrg 根据 OrgID 获取云账号列表
func (client *DBClient) GetAccountsByOrgID(orgID int64) ([]model.CloudAccount, error) {
	var accounts []model.CloudAccount
	if err := client.Where("org_id = ?", orgID).Find(&accounts).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return accounts, nil
		}
		return accounts, err
	}
	return accounts, nil
}

// GetAccountByName 根据 OrgID & Name 获取云账号
func (client *DBClient) GetAccountByName(orgID int64, name string) (model.CloudAccount, error) {
	var account model.CloudAccount
	if err := client.Where("org_id = ?", orgID).Where("name = ?", name).
		Find(&account).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return account, nil
		}
		return account, err
	}
	return account, nil
}

// GetAccountByID 根据 ID 获取云账号
func (client *DBClient) GetAccountByID(orgID, accountID int64) (model.CloudAccount, error) {
	var account model.CloudAccount
	if err := client.Where("org_id = ?", orgID).Where("id = ?", accountID).Find(&account).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return account, nil
		}
		return account, err
	}
	return account, nil
}
