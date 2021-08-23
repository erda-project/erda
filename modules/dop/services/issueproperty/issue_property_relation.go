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
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
)

func (ip *IssueProperty) CreatePropertyRelation(req *apistructs.IssuePropertyRelationCreateRequest) error {
	var propertyInstances []dao.IssuePropertyRelation
	for _, p := range req.Property {
		// 如果是单选或多选，每条枚举值建立联系
		if p.PropertyType.IsOptions() {
			// 必填项
			if p.Required == true && len(p.Values) == 0 {
				return apierrors.ErrCreateIssue.MissingParameter(fmt.Sprintf("必填字段\"%v\"未填写", p.PropertyID))
			}
			for _, v := range p.Values {
				propertyInstances = append(propertyInstances, dao.IssuePropertyRelation{
					OrgID:           req.OrgID,
					ProjectID:       req.ProjectID,
					IssueID:         req.IssueID,
					PropertyID:      p.PropertyID,
					PropertyValueID: v,
				})
			}
		} else {
			if p.Required == true && p.ArbitraryValue == "" {
				return apierrors.ErrCreateIssue.MissingParameter(fmt.Sprintf("必填字段\"%v\"未填写", p.PropertyID))
			}
			if p.ArbitraryValue == "" {
				continue
			}
			propertyInstances = append(propertyInstances, dao.IssuePropertyRelation{
				OrgID:          req.OrgID,
				ProjectID:      req.ProjectID,
				IssueID:        req.IssueID,
				PropertyID:     p.PropertyID,
				ArbitraryValue: p.GetArb(),
			})
		}
	}
	for _, p := range req.Property {
		if err := ip.db.DeletePropertyRelationsByPropertyID(req.IssueID, p.PropertyID); err != nil {
			return err
		}
	}
	if err := ip.db.CreatePropertyRelations(propertyInstances); err != nil {
		return err
	}
	return nil
}

func (ip *IssueProperty) UpdatePropertyRelation(req *apistructs.IssuePropertyRelationUpdateRequest) error {
	issue, err := ip.db.GetIssue(req.IssueID)
	if err != nil {
		return err
	}
	if req.PropertyType.IsOptions() == false {
		if req.Required == true && req.ArbitraryValue == "" {
			return apierrors.ErrUpdateIssue.MissingParameter("arbitraryValue")
		}
		_, err := ip.db.GetPropertyRelationByIssueID(req.IssueID, req.PropertyID)
		if err != nil {
			// 如果该字段原本无记录,新增
			if gorm.IsRecordNotFoundError(err) {
				return ip.db.CreatePropertyRelation(&dao.IssuePropertyRelation{
					OrgID:          req.OrgID,
					ProjectID:      req.ProjectID,
					IssueID:        req.IssueID,
					PropertyID:     req.PropertyID,
					ArbitraryValue: req.GetArb(),
				})
			}
			return err
		}
		// 如果本来有记录
		return ip.db.UpdatePropertyRelationArbitraryValue(req.IssueID, req.PropertyID, req.GetArb())
	}
	var propertyInstances []dao.IssuePropertyRelation
	for _, v := range req.Values {
		propertyInstances = append(propertyInstances, dao.IssuePropertyRelation{
			OrgID:           req.OrgID,
			ProjectID:       int64(issue.ProjectID),
			IssueID:         req.IssueID,
			PropertyID:      req.PropertyID,
			ArbitraryValue:  "",
			PropertyValueID: v,
		})
	}
	if err := ip.db.DeletePropertyRelationsByPropertyID(req.IssueID, req.PropertyID); err != nil {
		return err
	}
	if err := ip.db.CreatePropertyRelations(propertyInstances); err != nil {
		return err
	}
	return nil
}

// GetPropertyRelation 根据字段id获取字段
func (ip *IssueProperty) GetPropertyRelation(req *apistructs.IssuePropertyRelationGetRequest) (*apistructs.IssueAndPropertyAndValue, error) {
	var instances []apistructs.IssuePropertyInstance
	mp := make(map[int64]int)
	// 获取该事件类型配置的全部自定义字段
	properties, err := ip.GetProperties(apistructs.IssuePropertiesGetRequest{
		OrgID:             req.OrgID,
		PropertyIssueType: req.PropertyIssueType,
	})
	if err != nil {
		return nil, err
	}
	// 构建property到instances的映射，instances中存放自定义字段信息（不含值）
	for i, pro := range properties {
		instances = append(instances, apistructs.IssuePropertyInstance{IssuePropertyIndex: pro})
		mp[pro.PropertyID] = i
	}
	// 填充instances每个自定义字段的值
	relations, err := ip.db.GetPropertyRelationByID(req.IssueID)
	for _, v := range relations {
		if mp[v.PropertyID] >= len(instances) {
			return nil, apierrors.ErrGetIssue.InvalidState("找不到使用的自定义字段")
		}
		if instances[mp[v.PropertyID]].IssuePropertyIndex.PropertyType.IsOptions() == false {
			instances[mp[v.PropertyID]].ArbitraryValue = v.ArbitraryValue
			continue
		}
		instances[mp[v.PropertyID]].EnumeratedValues = append(
			instances[mp[v.PropertyID]].EnumeratedValues, apistructs.PropertyEnumerate{ID: v.PropertyValueID})
	}

	res, err := ip.ConvertRelations(req.IssueID, instances)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (ip *IssueProperty) DeletePropertyRelation(issueID int64) error {
	return ip.db.DeletePropertyRelationByIssueID(issueID)
}
