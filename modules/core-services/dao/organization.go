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
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateOrg 创建企业
func (client *DBClient) CreateOrg(org *model.Org) error {
	return client.Create(org).Error
}

// UpdateOrg 更新企业元信息，不可更改企业名称
func (client *DBClient) UpdateOrg(org *model.Org) error {
	return client.Save(org).Error
}

// DeleteOrg 删除企业
func (client *DBClient) DeleteOrg(orgID int64) error {
	return client.Where("id = ?", orgID).Delete(&model.Org{}).Error
}

// GetOrg 根据orgID获取企业信息
func (client *DBClient) GetOrg(orgID int64) (model.Org, error) {
	var org model.Org
	if err := client.Where("id = ?", orgID).Find(&org).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return org, ErrNotFoundOrg
		}
		return org, err
	}
	return org, nil
}

// GetOrgByName 根据 orgName 获取企业信息
func (client *DBClient) GetOrgByName(orgName string) (*model.Org, error) {
	var org model.Org
	if err := client.Where("name = ?", orgName).First(&org).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrNotFoundOrg
		}
		return nil, err
	}
	return &org, nil
}

// GetOrgsByParam 获取企业列表
func (client *DBClient) GetOrgsByParam(name string, pageNum, pageSize int) (int, []model.Org, error) {
	var (
		orgs  []model.Org
		total int
	)
	if name == "" {
		if err := client.Order("updated_at DESC").Offset((pageNum - 1) * pageSize).Limit(pageSize).
			Find(&orgs).Error; err != nil {
			return 0, nil, err
		}
		// 获取总量
		if err := client.Model(&model.Org{}).Count(&total).Error; err != nil {
			return 0, nil, err
		}
	} else {
		if err := client.Where("name LIKE ?", "%"+name+"%").Or("display_name LIKE ?", "%"+name+"%").Order("updated_at DESC").
			Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&orgs).Error; err != nil {
			return 0, nil, err
		}
		// 获取总量
		if err := client.Model(&model.Org{}).Where("name LIKE ?", "%"+name+"%").Or("display_name LIKE ?", "%"+name+"%").
			Count(&total).Error; err != nil {
			return 0, nil, err
		}
	}
	return total, orgs, nil
}

// Get public orgs list
func (client *DBClient) GetPublicOrgsByParam(name string, pageNum, pageSize int) (int, []model.Org, error) {
	var (
		orgs  []model.Org
		total int
	)
	if err := client.Where("is_public = ?", 1).Where("name LIKE ? OR display_name LIKE ?", "%"+name+"%", "%"+name+"%").Order("updated_at DESC").
		Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&orgs).Error; err != nil {
		return 0, nil, err
	}
	if err := client.Model(&model.Org{}).Where("is_public = ?", 1).Where("name LIKE ? OR display_name LIKE ?", "%"+name+"%", "%"+name+"%").
		Count(&total).Error; err != nil {
		return 0, nil, err
	}
	return total, orgs, nil
}

// GetOrgsByUserID 根据userID获取企业列表
func (client *DBClient) GetOrgsByUserID(userID string) ([]model.Org, error) {
	var orgs []model.Org
	if err := client.Where("creator = ?", userID).Order("updated_at DESC").
		Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

// GetOrgsByIDsAndName 根据企业ID列表 & 企业名称获取企业列表
func (client *DBClient) GetOrgsByIDsAndName(orgIDs []int64, name string, pageNo, pageSize int) (
	int, []model.Org, error) {
	var (
		total int
		orgs  []model.Org
	)
	if err := client.Where("id in (?)", orgIDs).
		Where("name LIKE ? OR display_name LIKE ?", strutil.Concat("%", name, "%"), strutil.Concat("%", name, "%")).Order("updated_at DESC").
		Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&orgs).Error; err != nil {
		return 0, nil, err
	}
	// 获取总量
	if err := client.Model(&model.Org{}).Where("id in (?)", orgIDs).
		Where("name LIKE ? OR display_name LIKE ?", strutil.Concat("%", name, "%"), strutil.Concat("%", name, "%")).Count(&total).Error; err != nil {
		return 0, nil, err
	}
	return total, orgs, nil
}

// GetOrgList 获取所有企业列表(仅供内部用户使用)
func (client *DBClient) GetOrgList() ([]model.Org, error) {
	var orgs []model.Org
	if err := client.Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}
