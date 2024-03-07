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

package query

import (
	"fmt"
	"strconv"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/common"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/internal/apps/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/strutil"
)

func (p *provider) GetProperties(req *pb.GetIssuePropertyRequest) ([]*pb.IssuePropertyIndex, error) {
	properties, err := p.db.GetIssueProperties(*req)
	if err != nil {
		return nil, err
	}
	propertyIndexes := BatchConvert(properties)
	propertyMap := make(map[int64][]string) // key: PropertyID  value: index
	// 只有公用字段会被任务类型模版使用
	if req.PropertyIssueType == pb.PropertyIssueTypeEnum_COMMON.String() {
		allProperties, err := p.db.GetIssueProperties(pb.GetIssuePropertyRequest{
			OrgID:                req.OrgID,
			ScopeType:            req.ScopeType,
			ScopeID:              req.ScopeID,
			OnlyCurrentScopeType: req.OnlyCurrentScopeType,
		})
		if err != nil {
			return nil, err
		}
		for _, p := range allProperties {
			if p.PropertyIssueType != pb.PropertyIssueTypeEnum_COMMON.String() {
				propertyMap[p.Relation] = append(propertyMap[p.Relation], GetZhName(p.PropertyIssueType))
			}
		}
	}
	for i, v := range propertyIndexes {
		if req.PropertyIssueType == pb.PropertyIssueTypeEnum_COMMON.String() {
			propertyIndexes[i].RelatedIssue = strutil.DedupSlice(propertyMap[v.PropertyID])
		} else {
			propertyIndexes[i].RelatedIssue = append(propertyIndexes[i].RelatedIssue, GetZhName(v.PropertyIssueType.String()))
		}
		// 如果不是单选或者多选，不需要获取枚举值
		if common.IsOptions(v.PropertyType.String()) == false {
			continue
		}
		var (
			values []dao.IssuePropertyValue
			err    error
		)
		// 如果是COMMON类型获取自己的枚举值，如果不是则获取关联的COMMON类型的枚举值
		if v.PropertyIssueType == pb.PropertyIssueTypeEnum_COMMON {
			values, err = p.db.GetIssuePropertyValues(v.PropertyID)
		} else {
			values, err = p.db.GetIssuePropertyValues(v.Relation)
		}
		if err != nil {
			return nil, err
		}
		for index, val := range values {
			propertyIndexes[i].EnumeratedValues = append(propertyIndexes[i].EnumeratedValues, &pb.Enumerate{
				Index: int64(index),
				Id:    int64(val.ID),
				Name:  val.Name,
			})
		}
		if common.IsOptions(propertyIndexes[i].PropertyType.String()) && len(propertyIndexes[i].EnumeratedValues) > 0 && propertyIndexes[i].Required == true {
			propertyIndexes[i].Values = append(propertyIndexes[i].Values, propertyIndexes[i].EnumeratedValues[0].Id)
		}
	}
	return propertyIndexes, nil
}

func GetZhName(t string) string {
	switch t {
	case pb.IssueTypeEnum_REQUIREMENT.String():
		return "需求"
	case pb.IssueTypeEnum_TASK.String():
		return "任务"
	case pb.IssueTypeEnum_BUG.String():
		return "缺陷"
	case pb.IssueTypeEnum_EPIC.String():
		return "史诗"
	case pb.PropertyIssueTypeEnum_COMMON.String():
		return "公用"
	default:
		panic(fmt.Sprintf("invalid issue type: %s", string(t)))
	}
}

func Convert(is *dao.IssueProperty) *pb.IssuePropertyIndex {
	return &pb.IssuePropertyIndex{
		PropertyID:        int64(is.ID),
		ScopeType:         pb.ScopeTypeEnum_ScopeType(pb.ScopeTypeEnum_ScopeType_value[is.ScopeType]),
		ScopeID:           is.ScopeID,
		OrgID:             is.OrgID,
		Required:          is.Required,
		PropertyType:      pb.PropertyTypeEnum_PropertyType(pb.PropertyTypeEnum_PropertyType_value[is.PropertyType]),
		PropertyName:      is.PropertyName,
		DisplayName:       is.DisplayName,
		PropertyIssueType: pb.PropertyIssueTypeEnum_PropertyIssueType(pb.PropertyIssueTypeEnum_PropertyIssueType_value[is.PropertyIssueType]),
		Relation:          is.Relation,
		Index:             is.Index,
	}
}

func BatchConvert(properties []dao.IssueProperty) []*pb.IssuePropertyIndex {
	var response []*pb.IssuePropertyIndex
	for _, is := range properties {
		response = append(response, Convert(&is))
	}
	return response
}

func (pr *provider) CreatePropertyRelation(req *pb.CreateIssuePropertyInstanceRequest) error {
	var propertyInstances []dao.IssuePropertyRelation
	for _, p := range req.Property {
		// 如果是单选或多选，每条枚举值建立联系
		if common.IsOptions(p.PropertyType.String()) {
			// 必填项
			if p.Required == true && len(p.Values) == 0 {
				return apierrors.ErrCreateIssueProperty.MissingParameter(fmt.Sprintf("必填字段\"%v\"未填写", p.PropertyName))
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
			arbValue := GetArb(p)
			if p.Required == true && arbValue == "" {
				return apierrors.ErrCreateIssueProperty.MissingParameter(fmt.Sprintf("必填字段\"%v\"未填写", p.PropertyName))
			}
			if arbValue == "" {
				continue
			}
			propertyInstances = append(propertyInstances, dao.IssuePropertyRelation{
				OrgID:          req.OrgID,
				ProjectID:      req.ProjectID,
				IssueID:        req.IssueID,
				PropertyID:     p.PropertyID,
				ArbitraryValue: arbValue,
			})
		}
	}
	for _, p := range req.Property {
		if err := pr.db.DeletePropertyRelationsByPropertyID(req.IssueID, p.PropertyID); err != nil {
			return err
		}
	}
	if err := pr.db.CreatePropertyRelations(propertyInstances); err != nil {
		return err
	}
	return nil
}

func GetArb(i *pb.IssuePropertyInstance) string {
	if s := i.ArbitraryValue.GetNumberValue(); s != 0 {
		return strconv.Itoa(int(s))
	}
	return i.ArbitraryValue.GetStringValue()
}

func (p *provider) BatchGetProperties(orgID int64, issuesTypes []string) ([]*pb.IssuePropertyIndex, error) {
	var (
		properties []*pb.IssuePropertyIndex
		err        error
	)
	if len(issuesTypes) > 0 {
		var wg sync.WaitGroup
		errChan := make(chan error, len(issuesTypes))
		propertiesChan := make(chan []*pb.IssuePropertyIndex, len(issuesTypes))
		for _, issueType := range issuesTypes {
			wg.Add(1)
			go func(issueType string) {
				defer wg.Done()
				properties, err := p.GetProperties(&pb.GetIssuePropertyRequest{
					OrgID:             orgID,
					PropertyIssueType: issueType,
				})
				if err != nil {
					errChan <- err
					return
				}
				propertiesChan <- properties
			}(issueType)
		}
		wg.Wait()
		close(errChan)
		close(propertiesChan)
		for err := range errChan {
			if err != nil {
				return nil, apierrors.ErrGetIssueProperty.InternalError(err)
			}
		}
		for p := range propertiesChan {
			properties = append(properties, p...)
		}
	} else {
		// get by all issue types
		req := &pb.GetIssuePropertyRequest{OrgID: orgID}
		properties, err = p.GetProperties(req)
		if err != nil {
			return nil, err
		}
	}
	return properties, nil
}

// BatchGetIssuePropertyInstances .
// return:
//
//	@map[int64][]*dao.IssuePropertyRelation: key is issueID
//	@err: error
func (p *provider) BatchGetIssuePropertyInstances(orgID int64, issueType string, issueIDs []uint64) (map[uint64]*pb.IssueAndPropertyAndValue, error) {
	issueInstancesMap := make(map[uint64]*pb.IssueAndPropertyAndValue)
	// 获取该事件类型配置的全部自定义字段
	if len(issueType) == 0 {
		return nil, apierrors.ErrGetIssueProperty.MissingParameter("issueType")
	}
	properties, err := p.GetProperties(&pb.GetIssuePropertyRequest{
		OrgID:             orgID,
		PropertyIssueType: issueType,
	})
	if err != nil {
		return nil, err
	}

	// get all issue property relations
	relations, err := p.db.ListPropertyRelationsByIssueIDs(issueIDs)
	if err != nil {
		return nil, apierrors.ErrGetIssuePropertyInstance.InternalError(err)
	}
	issueRelationsMap := make(map[int64][]*dao.IssuePropertyRelation)
	for _, relation := range relations {
		issueRelationsMap[relation.IssueID] = append(issueRelationsMap[relation.IssueID], relation)
	}

	for issueID, issueRelations := range issueRelationsMap {
		instances, err := p.getIssuePropertyInstance(issueID, properties, issueRelations)
		if err != nil {
			return nil, err
		}
		issueInstancesMap[uint64(issueID)] = instances
	}

	return issueInstancesMap, nil
}

func (p *provider) getIssuePropertyInstance(issueID int64, properties []*pb.IssuePropertyIndex, relations []*dao.IssuePropertyRelation) (*pb.IssueAndPropertyAndValue, error) {
	var instances []*pb.IssuePropertyInstance
	propertyInstanceMap := make(map[int64]*pb.IssuePropertyInstance, len(properties))
	// 构建property到instances的映射，instances中存放自定义字段信息（不含值）
	for _, pro := range properties {
		instance := &pb.IssuePropertyInstance{
			PropertyID:        pro.PropertyID,
			ScopeID:           pro.ScopeID,
			ScopeType:         pro.ScopeType,
			OrgID:             pro.OrgID,
			PropertyName:      pro.PropertyName,
			DisplayName:       pro.DisplayName,
			PropertyType:      pro.PropertyType,
			Required:          pro.Required,
			PropertyIssueType: pro.PropertyIssueType,
			Relation:          pro.Relation,
			Index:             pro.Index,
			EnumeratedValues:  pro.EnumeratedValues,
			Values:            pro.Values,
			RelatedIssue:      pro.RelatedIssue,
		}
		instances = append(instances, instance)
		propertyInstanceMap[pro.PropertyID] = instance
	}
	// 填充instances每个自定义字段的值
	for _, v := range relations {
		instance, ok := propertyInstanceMap[v.PropertyID]
		if !ok {
			return nil, apierrors.ErrGetIssuePropertyInstance.InvalidState(
				fmt.Sprintf("找不到使用的自定义字段, issueID: %d, propertyID: %d", issueID, v.PropertyID))
		}
		if !common.IsOptions(instance.PropertyType.String()) {
			instance.ArbitraryValue = structpb.NewStringValue(v.ArbitraryValue)
			continue
		}
		instance.PropertyEnumeratedValues = append(
			instance.PropertyEnumeratedValues, &pb.PropertyEnumerate{Id: v.PropertyValueID})
	}

	return convertRelations(issueID, instances)
}

func (p *provider) GetIssuePropertyInstance(req *pb.GetIssuePropertyInstanceRequest) (*pb.IssueAndPropertyAndValue, error) {
	// 获取该事件类型配置的全部自定义字段
	properties, err := p.GetProperties(&pb.GetIssuePropertyRequest{
		OrgID:             req.OrgID,
		PropertyIssueType: req.PropertyIssueType,
		ScopeType:         req.ScopeType,
		ScopeID:           req.ScopeID,
	})
	if err != nil {
		return nil, err
	}
	relations, err := p.db.GetPropertyRelationByID(req.IssueID)
	if err != nil {
		return nil, err
	}
	return p.getIssuePropertyInstance(req.IssueID, properties, relations)
}

func convertRelations(issueID int64, relations []*pb.IssuePropertyInstance) (*pb.IssueAndPropertyAndValue, error) {
	res := &pb.IssueAndPropertyAndValue{
		IssueID: issueID,
	}
	for i, v := range relations {
		var arbitraryValue interface{}
		// 判断出参应该是数字还是字符串
		if v.PropertyType == pb.PropertyTypeEnum_Number && v.ArbitraryValue != nil {
			if val := v.ArbitraryValue.GetNumberValue(); val > 0 {
				arbitraryValue = val
			} else {
				arbitraryValue = v.ArbitraryValue.GetStringValue()
			}
		} else {
			arbitraryValue = v.ArbitraryValue
		}

		var arbi *structpb.Value
		if v.ArbitraryValue != nil {
			a, ok := arbitraryValue.(*structpb.Value)
			if ok {
				arbi = a
			} else {
				arb, err := structpb.NewValue(arbitraryValue)
				if err != nil {
					return nil, err
				}
				arbi = arb
			}
		}

		res.Property = append(res.Property, &pb.IssuePropertyExtraProperty{
			PropertyID:       v.PropertyID,
			PropertyType:     v.PropertyType,
			PropertyName:     v.PropertyName,
			Required:         v.Required,
			DisplayName:      v.DisplayName,
			ArbitraryValue:   arbi,
			EnumeratedValues: v.EnumeratedValues,
		})
		for _, val := range v.PropertyEnumeratedValues {
			res.Property[i].Values = append(res.Property[i].Values, val.Id)
		}
	}
	return res, nil
}
