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

import "github.com/erda-project/erda/pkg/database/dbengine"

// IssuePropertyValue 事件属性值表
type IssuePropertyValue struct {
	dbengine.BaseModel

	PropertyID int64  `gorm:"column:property_id"`
	Value      string `gorm:"column:value"`
	Name       string `gorm:"column:name"`
}

func (IssuePropertyValue) TableName() string {
	return "dice_issue_property_value"
}

func (client *DBClient) CreateIssuePropertyValue(value *IssuePropertyValue) error {
	return client.Create(value).Error
}

func (client *DBClient) CreateIssuePropertyValues(values []IssuePropertyValue) error {
	return client.BulkInsert(values)
}

func (client *DBClient) DeleteIssuePropertyValue(id int64) error {
	return client.Where("id = ?", id).Delete(&IssuePropertyValue{}).Error
}

func (client *DBClient) DeleteIssuePropertyValuesByPropertyID(propertyID int64) error {
	return client.Table("dice_issue_property_value").Where("property_id = ?", propertyID).Delete(&IssuePropertyValue{}).Error
}

func (client *DBClient) UpdateIssuePropertyValue(property *IssuePropertyValue) error {
	return client.Save(property).Error
}

func (client *DBClient) GetIssuePropertyValues(orgID int64) ([]IssuePropertyValue, error) {
	var propertyValues []IssuePropertyValue
	if err := client.Where("property_id = ?", orgID).Find(&propertyValues).Error; err != nil {
		return nil, err
	}
	return propertyValues, nil
}

func (client *DBClient) GetIssuePropertyValue(id int64) (*IssuePropertyValue, error) {
	var propertyValues IssuePropertyValue
	if err := client.Where("id = ?", id).Find(&propertyValues).Error; err != nil {
		return nil, err
	}
	return &propertyValues, nil
}
