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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreatePublisher 创建Publisher
func (client *DBClient) CreatePublisher(publisher *model.Publisher) error {
	return client.Create(publisher).Error
}

// UpdatePublisher 更新Publisher
func (client *DBClient) UpdatePublisher(publisher *model.Publisher) error {
	return client.Save(publisher).Error
}

// DeletePublisher 删除Publisher
func (client *DBClient) DeletePublisher(publisherID int64) error {
	return client.Where("id = ?", publisherID).Delete(&model.Publisher{}).Error
}

// GetPublisherByID 根据publisherID获取Publisher信息
func (client *DBClient) GetPublisherByID(publisherID int64) (model.Publisher, error) {
	var publisher model.Publisher
	if err := client.Where("id = ?", publisherID).Find(&publisher).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return publisher, ErrNotFoundPublisher
		}
		return publisher, err
	}
	return publisher, nil
}

// GetPublishersByOrgIDAndName 根据orgID与名称获取Publisher列表
func (client *DBClient) GetPublishersByOrgIDAndName(orgID int64, params *apistructs.PublisherListRequest) (
	int, []model.Publisher, error) {
	var (
		publishers []model.Publisher
		total      int
	)
	db := client.Where("org_id = ?", orgID)
	if params.Name != "" {
		db = db.Where("name = ?", params.Name)
	}
	if params.Query != "" {
		db = db.Where("name LIKE ?", strutil.Concat("%", params.Query, "%"))
	}
	db = db.Order("updated_at DESC")
	if err := db.Offset((params.PageNo - 1) * params.PageSize).Limit(params.PageSize).
		Find(&publishers).Error; err != nil {
		return 0, nil, err
	}

	// 获取总量
	db = client.Model(&model.Publisher{}).Where("org_id = ?", orgID)
	if params.Name != "" {
		db = db.Where("name = ?", params.Name)
	}
	if params.Query != "" {
		db = db.Where("name LIKE ?", strutil.Concat("%", params.Query, "%"))
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	return total, publishers, nil
}

// GetPublishersByIDs 根据publisherIDs获取Publisher列表
func (client *DBClient) GetPublishersByIDs(publisherIDs []int64, params *apistructs.PublisherListRequest) (
	int, []model.Publisher, error) {
	var (
		total      int
		publishers []model.Publisher
	)
	db := client.Where("id in (?)", publisherIDs)
	if params.Name != "" {
		db = db.Where("name = ?", params.Name)
	}
	if params.Query != "" {
		db = db.Where("name LIKE ?", strutil.Concat("%", params.Query, "%"))
	}
	db = db.Order("updated_at DESC")
	if err := db.Offset((params.PageNo - 1) * params.PageSize).Limit(params.PageSize).
		Find(&publishers).Error; err != nil {
		return 0, nil, err
	}

	// 获取总量
	db = client.Model(&model.Publisher{}).Where("id in (?)", publisherIDs)
	if params.Name != "" {
		db = db.Where("name = ?", params.Name)
	}
	if params.Query != "" {
		db = db.Where("name LIKE ?", strutil.Concat("%", params.Query, "%"))
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	return total, publishers, nil
}

// GetPublisherByOrgAndName 根据orgID & Publisher名称 获取Publisher
func (client *DBClient) GetPublisherByOrgAndName(orgID int64, name string) (*model.Publisher, error) {
	var publisher model.Publisher
	if err := client.Where("org_id = ?", orgID).
		Where("name = ?", name).Find(&publisher).Error; err != nil {
		return nil, err
	}
	return &publisher, nil
}

// GetPublishersByIDs 获取企业的Publisher个数
func (client *DBClient) GetOrgPublishersCount(orgID uint64) (int64, error) {
	var count int64
	if err := client.Model(&model.Publisher{}).Where("org_id = ?", orgID).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// GetPublisherByOrgID 根据orgID 获取Publisher
func (client *DBClient) GetPublisherByOrgID(orgID int64) (*model.Publisher, error) {
	var publisher model.Publisher
	if err := client.Where("org_id = ?", orgID).First(&publisher).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrNotFoundPublisher
		}
		return nil, err
	}
	return &publisher, nil
}
