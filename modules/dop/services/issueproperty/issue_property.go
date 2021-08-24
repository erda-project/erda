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

package issueproperty

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateProperty 添加事项字段
func (ip *IssueProperty) CreateProperty(req *apistructs.IssuePropertyCreateRequest) (*apistructs.IssuePropertyIndex, error) {
	// 获取原有该任务类型事项数量，默认新增事项排序级为最低
	properties, err := ip.db.GetIssueProperties(apistructs.IssuePropertiesGetRequest{
		OrgID:             req.OrgID,
		PropertyIssueType: req.PropertyIssueType,
	})
	var startIndex int64 = 0
	for _, v := range properties {
		if v.Index >= startIndex {
			startIndex = v.Index + 1
		}
	}

	if err != nil {
		return nil, err
	}
	issueProperty := &dao.IssueProperty{
		ScopeID:           req.ScopeID,
		ScopeType:         req.ScopeType,
		OrgID:             req.OrgID,
		PropertyName:      req.PropertyName,
		DisplayName:       req.PropertyName, // 两个name相同
		PropertyType:      req.PropertyType,
		Required:          req.Required,
		PropertyIssueType: req.PropertyIssueType,
		Relation:          req.Relation,
		Index:             startIndex,
	}
	// 校验字段属性
	if issueProperty.PropertyType == "" {
		return nil, apierrors.ErrCreateIssueProperty.MissingParameter("PropertyType")
	}
	if issueProperty.PropertyIssueType != apistructs.PropertyIssueTypeCommon && issueProperty.Relation <= 0 {
		return nil, apierrors.ErrCreateIssueProperty.MissingParameter("relation")
	}
	if issueProperty.PropertyName == "" {
		return nil, apierrors.ErrCreateIssueProperty.MissingParameter("PropertyName")
	}
	if len(req.PropertyName) > 100 {
		return nil, apierrors.ErrCreateIssueProperty.InvalidParameter("PropertyName is longer than 100")
	}
	// 重名检测
	propertyName, err := ip.GetByName(req.OrgID, req.PropertyName, req.PropertyIssueType)
	if err != nil {
		return nil, apierrors.ErrCreateIssueProperty.InternalError(err)
	}
	if propertyName != nil {
		return nil, apierrors.ErrCreateIssueProperty.AlreadyExists()
	}
	// 校验字段枚举值属性
	var value []dao.IssuePropertyValue
	for _, val := range req.EnumeratedValues {
		if len(val.Name) > 100 {
			err = apierrors.ErrCreateIssueProperty.InvalidParameter("EnumeratedName is longer than 100")
			return nil, err
		}
		value = append(value, dao.IssuePropertyValue{
			PropertyID: int64(issueProperty.ID),
			Name:       val.Name,
		})
	}
	// 添加字段数据
	if err := ip.db.CreateIssueProperty(issueProperty); err != nil {
		return nil, err
	}
	propertyIndex := &apistructs.IssuePropertyIndex{
		PropertyID:        int64(issueProperty.ID),
		ScopeID:           issueProperty.ScopeID,
		ScopeType:         issueProperty.ScopeType,
		OrgID:             issueProperty.OrgID,
		PropertyName:      issueProperty.PropertyName,
		DisplayName:       issueProperty.DisplayName,
		PropertyType:      issueProperty.PropertyType,
		Required:          issueProperty.Required,
		PropertyIssueType: issueProperty.PropertyIssueType,
		Relation:          issueProperty.Relation,
		Index:             issueProperty.Index,
	}
	// 如果不是创建公有字段,或者公有字段的类型不是单选、多选 则不需要向数据库添加枚举值
	if issueProperty.PropertyIssueType != apistructs.PropertyIssueTypeCommon {
		return propertyIndex, nil
	} else if issueProperty.PropertyType.IsOptions() == false {
		return propertyIndex, nil
	}
	// 添加字段枚举值数据
	for i := range value {
		value[i].PropertyID = int64(issueProperty.ID)
	}
	if err = ip.db.CreateIssuePropertyValues(value); err != nil {
		return nil, err
	}
	for i, v := range value {
		propertyIndex.EnumeratedValues = append(propertyIndex.EnumeratedValues, apistructs.Enumerate{
			ID:    int64(v.ID),
			Name:  v.Name,
			Index: int64(i),
		})
	}
	return propertyIndex, nil
}

// DeleteProperty 删除事项字段
func (ip *IssueProperty) DeleteProperty(orgID int64, propertyIssueType apistructs.PropertyIssueType, PropertyID int64, index int64) error {
	// 如果该字段被关联，不能删除
	property, err := ip.GetByRelation(PropertyID)
	if err != nil {
		return apierrors.ErrDeleteIssueProperty.InternalError(err)
	}
	if property != nil {
		return apierrors.ErrDeleteIssueProperty.InvalidParameter("该字段已被其他字段引用")
	}

	// 删除用了该字段的全部事件实例
	if err := ip.db.DeletePropertyRelationsByPropertyID(-1, PropertyID); err != nil {
		return err
	}
	// 删除事项字段的枚举值
	if err := ip.db.DeleteIssuePropertyValuesByPropertyID(PropertyID); err != nil {
		return err
	}
	// 删除事项字段
	if err := ip.db.DeleteIssueProperty(orgID, propertyIssueType, PropertyID, index); err != nil {
		return err
	}
	return nil
}

// GetProperty 根据字段id获取字段
func (ip *IssueProperty) GetPropertyByID(PropertyID int64) (*apistructs.IssuePropertyIndex, error) {
	property, err := ip.db.GetIssuePropertyByID(PropertyID)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, apierrors.ErrGetIssueProperty.NotFound()
		}
		return nil, apierrors.ErrGetIssueProperty.InternalError(err)
	}
	response := ip.Convert(property)
	return response, err
}

// GetProperties 获取企业下的全部字段
func (ip *IssueProperty) GetProperties(req apistructs.IssuePropertiesGetRequest) ([]apistructs.IssuePropertyIndex, error) {
	properties, err := ip.db.GetIssueProperties(req)
	if err != nil {
		return nil, err
	}
	propertyIndexes := ip.BatchConvert(properties)
	propertyMap := make(map[int64][]string) // key: PropertyID  value: index
	// 只有公用字段会被任务类型模版使用
	if req.PropertyIssueType == apistructs.PropertyIssueTypeCommon {
		allProperties, err := ip.db.GetIssueProperties(apistructs.IssuePropertiesGetRequest{
			OrgID: req.OrgID,
		})
		if err != nil {
			return nil, err
		}
		for _, p := range allProperties {
			if p.PropertyIssueType != apistructs.PropertyIssueTypeCommon {
				propertyMap[p.Relation] = append(propertyMap[p.Relation], p.PropertyIssueType.GetZhName())
			}
		}
	}
	for i, v := range propertyIndexes {
		if req.PropertyIssueType == apistructs.PropertyIssueTypeCommon {
			propertyIndexes[i].RelatedIssue = strutil.DedupSlice(propertyMap[v.PropertyID])
		} else {
			propertyIndexes[i].RelatedIssue = append(propertyIndexes[i].RelatedIssue, v.PropertyIssueType.GetZhName())
		}
		// 如果不是单选或者多选，不需要获取枚举值
		if v.PropertyType.IsOptions() == false {
			continue
		}
		var (
			values []dao.IssuePropertyValue
			err    error
		)
		// 如果是COMMON类型获取自己的枚举值，如果不是则获取关联的COMMON类型的枚举值
		if v.PropertyIssueType == apistructs.PropertyIssueTypeCommon {
			values, err = ip.db.GetIssuePropertyValues(v.PropertyID)
		} else {
			values, err = ip.db.GetIssuePropertyValues(v.Relation)
		}
		if err != nil {
			return nil, err
		}
		for index, val := range values {
			propertyIndexes[i].EnumeratedValues = append(propertyIndexes[i].EnumeratedValues, apistructs.Enumerate{
				Index: int64(index),
				ID:    int64(val.ID),
				Name:  val.Name,
			})
		}
		if propertyIndexes[i].PropertyType.IsOptions() && len(propertyIndexes[i].EnumeratedValues) > 0 && propertyIndexes[i].Required == true {
			propertyIndexes[i].Values = append(propertyIndexes[i].Values, propertyIndexes[i].EnumeratedValues[0].ID)
		}
	}
	return propertyIndexes, err
}

// UpdateProperty 修改公用事项字段
func (ip *IssueProperty) UpdateProperty(req *apistructs.IssuePropertyUpdateRequest) (*apistructs.IssuePropertyIndex, error) {
	oldIssueProperty, err := ip.db.GetIssuePropertyByID(req.PropertyID)
	if err != nil {
		return nil, err
	}
	// 限定可修改的字段
	issueProperty := &dao.IssueProperty{
		BaseModel: dbengine.BaseModel{
			ID: oldIssueProperty.ID,
		},
		ScopeID:           oldIssueProperty.ScopeID,
		ScopeType:         oldIssueProperty.ScopeType,
		OrgID:             oldIssueProperty.OrgID,
		PropertyName:      req.PropertyName,
		DisplayName:       req.DisplayName,
		PropertyType:      req.PropertyType,
		Required:          req.Required,
		PropertyIssueType: oldIssueProperty.PropertyIssueType,
		Index:             oldIssueProperty.Index,
		Relation:          oldIssueProperty.Relation,
	}
	// 字段类型转换的校验
	if oldIssueProperty.PropertyType.IsCanChange(issueProperty.PropertyType) != true {
		return nil, apierrors.ErrUpdateIssueProperty.InvalidParameter("非法的PropertyType改变")
	}
	// 重名检测
	propertyName, err := ip.GetByName(req.OrgID, req.PropertyName, req.PropertyIssueType)
	if err != nil {
		return nil, apierrors.ErrUpdateIssueProperty.InternalError(err)
	}
	if propertyName != nil && propertyName.ID != issueProperty.ID {
		return nil, apierrors.ErrUpdateIssueProperty.AlreadyExists()
	}
	// 如果字段已经被使用，不能修改
	propertyRelated, err := ip.GetByRelation(req.PropertyID)
	if err != nil {
		return nil, apierrors.ErrUpdateIssueProperty.InternalError(err)
	}
	if propertyRelated != nil {
		return nil, apierrors.ErrUpdateIssueProperty.InvalidParameter("被引用的字段禁止修改")
	}
	// 更新字段
	if err := ip.db.UpdateIssueProperty(issueProperty); err != nil {
		return nil, err
	}
	// 删除该字段原有枚举值
	if err := ip.db.DeleteIssuePropertyValuesByPropertyID(req.PropertyID); err != nil {
		return nil, err
	}
	var values []dao.IssuePropertyValue
	for _, v := range req.EnumeratedValues {
		values = append(values, dao.IssuePropertyValue{
			PropertyID: req.PropertyID,
			Name:       v.Name,
		})
	}
	// 如果是单选或者多选,添加该字段新枚举值
	if issueProperty.PropertyType.IsOptions() {
		if err := ip.db.CreateIssuePropertyValues(values); err != nil {
			return nil, err
		}
	}
	response := ip.Convert(issueProperty)
	for i, v := range values {
		response.EnumeratedValues = append(response.EnumeratedValues, apistructs.Enumerate{
			ID:    int64(v.ID),
			Name:  v.Name,
			Index: int64(i),
		})
	}
	return response, nil
}

// UpdatePropertiesIndex 批量修改字段排序级
func (ip *IssueProperty) UpdatePropertiesIndex(req *apistructs.IssuePropertyIndexUpdateRequest) ([]apistructs.IssuePropertyIndex, error) {
	var propertiesIndex []dao.IssueProperty
	for _, issueProperty := range req.Data {
		propertiesIndex = append(propertiesIndex, dao.IssueProperty{
			BaseModel: dbengine.BaseModel{
				ID: uint64(issueProperty.PropertyID),
			},
			ScopeID:           issueProperty.ScopeID,
			ScopeType:         issueProperty.ScopeType,
			OrgID:             issueProperty.OrgID,
			PropertyName:      issueProperty.PropertyName,
			PropertyType:      issueProperty.PropertyType,
			Required:          issueProperty.Required,
			DisplayName:       issueProperty.DisplayName,
			PropertyIssueType: issueProperty.PropertyIssueType,
			Relation:          issueProperty.Relation,
			Index:             issueProperty.Index,
		})
	}
	if err := ip.db.UpdateIssuePropertiesIndex(propertiesIndex); err != nil {
		return nil, err
	}
	response := ip.BatchConvert(propertiesIndex)
	return response, nil
}

func (ip *IssueProperty) GetPropertyUpdateAt(orgID int64) (*apistructs.IssuePropertyUpdateTimes, error) {
	properties, err := ip.db.GetIssuePropertiesByTime(orgID)
	if err != nil {
		return nil, err
	}
	var updateAt apistructs.IssuePropertyUpdateTimes
	for _, v := range properties {
		switch v.PropertyIssueType {
		case apistructs.PropertyIssueTypeTask:
			updateAt.Task = v.UpdatedAt.Format("2006-01-02 15:04:05")
		case apistructs.PropertyIssueTypeBug:
			updateAt.Bug = v.UpdatedAt.Format("2006-01-02 15:04:05")
		case apistructs.PropertyIssueTypeEpic:
			updateAt.Epic = v.UpdatedAt.Format("2006-01-02 15:04:05")
		case apistructs.PropertyIssueTypeRequirement:
			updateAt.Requirement = v.UpdatedAt.Format("2006-01-02 15:04:05")
		}
	}
	return &updateAt, nil
}

// GetByName 根据 name 获取 property 详情
func (ip *IssueProperty) GetByName(orgID int64, name string, propertyIssueType apistructs.PropertyIssueType) (*dao.IssueProperty, error) {
	property, err := ip.db.GetIssuePropertyByName(orgID, name, propertyIssueType)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, nil
	}
	return property, nil
}

func (ip *IssueProperty) GetByRelation(id int64) (*dao.IssueProperty, error) {
	property, err := ip.db.GetIssuePropertiesByRelation(id)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, nil
	}
	return property, nil
}

func (ip *IssueProperty) GetByInstance(id int64) (*dao.IssuePropertyRelation, error) {
	relation, err := ip.db.GetPropertyRelationByPropertyID(id)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, nil
	}
	return relation, nil
}

func (ip *IssueProperty) GetBatchProperties(orgID int64, issuesType []apistructs.IssueType) ([]apistructs.IssuePropertyIndex, error) {
	var (
		properties []apistructs.IssuePropertyIndex
		err        error
	)
	if len(issuesType) != 1 {
		return nil, nil
	}
	switch issuesType[0] {
	case apistructs.IssueTypeTask:
		properties, err = ip.GetProperties(apistructs.IssuePropertiesGetRequest{OrgID: orgID, PropertyIssueType: apistructs.PropertyIssueTypeTask})
	case apistructs.IssueTypeBug:
		properties, err = ip.GetProperties(apistructs.IssuePropertiesGetRequest{OrgID: orgID, PropertyIssueType: apistructs.PropertyIssueTypeBug})
	case apistructs.IssueTypeRequirement:
		properties, err = ip.GetProperties(apistructs.IssuePropertiesGetRequest{OrgID: orgID, PropertyIssueType: apistructs.PropertyIssueTypeRequirement})
	case apistructs.IssueTypeEpic:
		properties, err = ip.GetProperties(apistructs.IssuePropertiesGetRequest{OrgID: orgID, PropertyIssueType: apistructs.PropertyIssueTypeEpic})
	}
	if err != nil {
		return nil, err
	}
	return properties, nil
}
