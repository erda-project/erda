// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dao

import (
	"strconv"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

// IssueProperty 事件属性表
type IssueProperty struct {
	dbengine.BaseModel

	ScopeType         string `gorm:"column:scope_type"`          // 系统管理员(sys)/企业(org)/项目(project)/应用(app)
	ScopeID           int64  `gorm:"column:scope_id"`            // 企业ID/项目ID/应用ID
	OrgID             int64  `gorm:"column:org_id"`              // 冗余 OrgID，方便删除企业时删除所有相关issue
	Required          bool   `gorm:"column:required"`            // 是否是必填项
	PropertyType      string `gorm:"column:property_type"`       // 属性的的类型（单选，多选，文本）
	PropertyName      string `gorm:"column:property_name"`       // 属性的名称
	DisplayName       string `gorm:"column:display_name"`        // 属性的展示名称
	PropertyIssueType string `gorm:"column:property_issue_type"` // 事件类型
	Relation          int64  `gorm:"column:relation"`            // 关联公用事件
	Index             int64  `gorm:"column:index"`               // 字段排序级
}

func (IssueProperty) TableName() string {
	return "dice_issue_property"
}

func (client *DBClient) CreateIssueProperty(property *IssueProperty) error {
	return client.Create(property).Error
}

func (client *DBClient) DeleteIssueProperty(orgId int64, propertyIssueType string, id int64, index int64) error {
	return client.Where(&IssueProperty{}).Where("id = ?", id).Delete(&IssueProperty{}).Error
}

func (client *DBClient) UpdateIssuePropertiesIndex(properties []IssueProperty) error {
	for i := range properties {
		property := properties[i]
		if err := client.Save(&property).Error; err != nil {
			return err
		}
	}
	return nil
}

func (client *DBClient) UpdateIssueProperty(property *IssueProperty) error {
	return client.Save(property).Error
}

func (client *DBClient) MigrateOrgCustomFileds(orgID int64, propertyNames []string) error {
	tx := client.Begin()
	var properties []IssueProperty
	// Obtain all custom fields of COMMON type that need to be migrated
	tx.Table("dice_issue_property").Where("org_id = ?", orgID).Where("scope_id = ?", orgID).Where("property_issue_type = 'COMMON'").Where("property_name IN (?)", propertyNames).
		Find(&properties)
	if tx.Error != nil {
		tx.Rollback()
		return tx.Error
	}
	for _, commonProperty := range properties {
		var threeTypeProperties []IssueProperty
		// Starting from the first one, obtain data on the three types of properties that can be associated with each commonProperty (EQUIREMENT, TASK, BUG), which may be 0-3
		tx.Table("dice_issue_property").Where("property_name = ?", commonProperty.PropertyName).
			Where("org_id = ?", orgID).Where("scope_id = ?", orgID).Where("property_issue_type != 'COMMON' ").Find(&threeTypeProperties)
		if tx.Error != nil {
			tx.Rollback()
			return tx.Error
		}
		// Obtain the project ID of the issue associated with the relationship
		var projectIDs []IssuePropertyRelation
		// Is the map record common created? If it is, do not create it
		existCommon := make(map[string]IssueProperty, 0)
		for _, relationshipType := range threeTypeProperties {
			// Obtain all project IDs that reference this type
			tx.Table("dice_issue_property_relation").Where("property_id = ?", relationshipType.ID).Group("project_id").Find(&projectIDs)
			if tx.Error != nil {
				tx.Rollback()
				return tx.Error
			}
			// create
			for _, commonAndThree := range projectIDs {
				// Create a map, don't repeat it anymore
				var resultCommon IssueProperty
				existCommonKey := relationshipType.PropertyName + ":" + "project" + ":" + strconv.FormatInt(commonAndThree.ProjectID, 10)
				if v, ok := existCommon[existCommonKey]; ok {
					resultCommon = v
				} else {
					// If the common does not exist, create a new common
					creatCommon := IssueProperty{
						ScopeType:         "project",
						ScopeID:           commonAndThree.ProjectID,
						OrgID:             commonProperty.OrgID,
						Required:          commonProperty.Required,
						PropertyType:      commonProperty.PropertyType,
						PropertyName:      commonProperty.PropertyName,
						DisplayName:       commonProperty.DisplayName,
						PropertyIssueType: commonProperty.PropertyIssueType,
						Relation:          commonProperty.Relation,
						Index:             commonProperty.Index,
					}
					tx.Table("dice_issue_property").Create(&creatCommon).Scan(&resultCommon)
					if tx.Error != nil {
						tx.Rollback()
						return tx.Error
					}
					existCommon[existCommonKey] = resultCommon
				}

				// Create this relationship under each project
				creatResultRelationshipType := IssueProperty{
					ScopeType:         "project",
					ScopeID:           commonAndThree.ProjectID,
					OrgID:             commonProperty.OrgID,
					Required:          commonProperty.Required,
					PropertyType:      commonProperty.PropertyType,
					PropertyName:      commonProperty.PropertyName,
					DisplayName:       commonProperty.DisplayName,
					PropertyIssueType: relationshipType.PropertyIssueType,
					Relation:          int64(resultCommon.ID),
					Index:             commonProperty.Index,
				}
				var resultRelationshipType IssueProperty
				tx.Table("dice_issue_property").Create(&creatResultRelationshipType).Scan(&resultRelationshipType)
				if tx.Error != nil {
					tx.Rollback()
					return tx.Error
				}
				// Update Relationship ID
				tx.Table("dice_issue_property_relation").Where("org_id = ?", resultRelationshipType.OrgID).Where("project_id = ?", commonAndThree.ProjectID).
					Where("property_id = ?", relationshipType.ID).Updates(IssuePropertyRelation{PropertyID: int64(resultRelationshipType.ID)})
				if tx.Error != nil {
					tx.Rollback()
					return tx.Error
				}
				// Because the property_id of the value connection is from the organizational level to the project level, there is a new property_id, so a new value needs to be created
				if commonProperty.PropertyType == pb.PropertyTypeEnum_Select.String() || commonProperty.PropertyType == pb.PropertyTypeEnum_MultiSelect.String() {
					// First, obtain the current organizational level information, which is the original value information
					var beforePropertyValue []IssuePropertyValue
					tx.Table("dice_issue_property_value").Where("property_id = ?", commonProperty.ID).Find(&beforePropertyValue)
					if tx.Error != nil {
						tx.Rollback()
						return tx.Error
					}
					for _, v := range beforePropertyValue {
						creatPropertyValue := IssuePropertyValue{
							PropertyID: int64(resultCommon.ID),
							Value:      v.Value,
							Name:       v.Name,
						}
						// Create a new one and modify a relation relationship
						var currentValue IssuePropertyValue
						tx.Table("dice_issue_property_value").Create(&creatPropertyValue).Scan(&currentValue)
						if tx.Error != nil {
							tx.Rollback()
							return tx.Error
						}
						// The old one called before is found. If it is found, it will be updated to the new value_id. If it cannot be found, it will be ignored
						tx.Table("dice_issue_property_relation").Where("org_id = ?", commonAndThree.OrgID).Where("project_id = ?", commonAndThree.ProjectID).
							Where("property_value_id = ?", v.ID).Updates(IssuePropertyRelation{PropertyValueID: int64(currentValue.ID)})
						if tx.Error != nil {
							tx.Rollback()
							return tx.Error
						}
					}
				}
			}
		}
	}
	tx.Commit()
	return nil
}

func (client *DBClient) GetIssueProperties(req pb.GetIssuePropertyRequest) ([]IssueProperty, error) {
	if req.ScopeType == "" {
		req.ScopeType = string(apistructs.ProjectScope)
	}
	var properties []IssueProperty
	var propertiesProject []IssueProperty
	db := client.Where(IssueProperty{}).Where("org_id = ?", req.OrgID)
	if req.PropertyIssueType != "" {
		db = db.Where("property_issue_type = ?", req.PropertyIssueType)
	}
	str := "%" + req.PropertyName + "%"
	if req.PropertyName != "" {
		db = db.Where("property_name LIKE ?", str)
	}
	if err := db.Where("scope_type = ?", string(apistructs.OrgScope)).Order("index").Find(&properties).Error; err != nil {
		return nil, err
	}
	if req.ScopeType != string(apistructs.ProjectScope) {
		return properties, nil
	}

	// propertiesProject 项目级 properties
	db = db.Where("scope_type = ?", string(apistructs.ProjectScope))
	if req.ScopeID != "" {
		db = db.Where("scope_id = ?", req.ScopeID)
	}
	if err := db.Order("index").Find(&propertiesProject).Error; err != nil {
		return nil, err
	}

	if req.OnlyCurrentScopeType && req.ScopeType == string(apistructs.ProjectScope) {
		return propertiesProject, nil
	}
	// 优先级：项目 > 企业级，当有重复的字段时，项目覆盖企业的字段；
	properties = nameConflict(properties, propertiesProject)
	return properties, nil
}

// nameConflict 重名覆盖函数用于解决自定义事项名称相同冲突问题，后一个属性将覆盖前一个,没有则添加
func nameConflict(properties ...[]IssueProperty) []IssueProperty {
	allCustomFields := make([]IssueProperty, 0)
	for _, props := range properties {
		for _, v := range props {
			found := false
			for i, existing := range allCustomFields {
				if existing.PropertyName == v.PropertyName && existing.PropertyIssueType == v.PropertyIssueType {
					allCustomFields[i] = v
					found = true
					break
				}
			}
			if !found {
				allCustomFields = append(allCustomFields, v)
			}
		}
	}
	return allCustomFields
}

func (client *DBClient) GetIssuePropertyByID(id int64) (*IssueProperty, error) {
	var property IssueProperty
	if err := client.Where(IssueProperty{}).Where("id = ?", id).Limit(1).Find(&property).Error; err != nil {
		return nil, err
	}
	return &property, nil
}

func (client *DBClient) GetIssuePropertiesByTime(orgID int64) ([]IssueProperty, error) {
	var properties []IssueProperty
	if err := client.Where(IssueProperty{}).Where("org_id = ?", orgID).Order("updated_at").Find(&properties).Error; err != nil {
		return nil, err
	}
	return properties, nil
}

func (client *DBClient) GetIssuePropertiesByRelation(ID int64) (*IssueProperty, error) {
	var property IssueProperty
	if err := client.Table("dice_issue_property").Where("relation = ?", ID).First(&property).Error; err != nil {
		return nil, err
	}
	return &property, nil
}

// GetIssuePropertyByName 根据 name 获取 property 信息
func (client *DBClient) GetIssuePropertyByName(orgID int64, Name string, PropertyIssueType string, scopeType string, scopeID int64) (*IssueProperty, error) {
	var property IssueProperty
	if err := client.Table("dice_issue_property").Where("org_id = ?", orgID).
		Where("property_name = ?", Name).Where("property_issue_type = ?", PropertyIssueType).
		Where("scope_type = ?", scopeType).Where("scope_id = ?", scopeID).First(&property).Error; err != nil {
		return nil, err
	}
	return &property, nil
}
