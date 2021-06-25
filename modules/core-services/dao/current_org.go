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
	"github.com/erda-project/erda/modules/core-services/model"
)

// CreateCurrentOrg 添加用户当前所属企业
func (client *DBClient) CreateCurrentOrg(orgRelation *model.CurrentOrg) error {
	return client.Create(orgRelation).Error
}

// UpdateCurrentOrg 更新用户当前所属企业
func (client *DBClient) UpdateCurrentOrg(userID string, orgID int64) error {
	var currentOrg model.CurrentOrg
	if err := client.Where("user_id = ?", userID).Find(&currentOrg).Error; err != nil {
		return err
	}
	currentOrg.OrgID = orgID
	return client.Save(&currentOrg).Error
}

// DeleteCurrentOrg 删除当前用户所属企业
func (client *DBClient) DeleteCurrentOrg(userID string) error {
	return client.Where("user_id = ?", userID).Delete(&model.CurrentOrg{}).Error
}

// GetCurrentOrgByUser 根据userID获取当前所属企业ID
func (client *DBClient) GetCurrentOrgByUser(userID string) (int64, error) {
	var currentOrg model.CurrentOrg
	if err := client.Where("user_id = ?", userID).Find(&currentOrg).Error; err != nil {
		return 0, err
	}
	return currentOrg.OrgID, nil
}
