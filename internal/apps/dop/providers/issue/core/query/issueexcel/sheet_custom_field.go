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

package issueexcel

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/common"
	issuedao "github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/excel"
)

func (data DataForFulfill) genCustomFieldSheet() (excel.Rows, error) {
	var lines excel.Rows
	// title: custom field id, custom field name, custom field type, custom field value
	title := excel.Row{
		excel.NewTitleCell("custom field id"),
		excel.NewTitleCell("custom field name"),
		excel.NewTitleCell("custom field type"),
		excel.NewTitleCell("custom field detail (json)"),
	}
	lines = append(lines, title)
	// data
	for propertyType, properties := range data.CustomFieldMapByTypeName {
		for _, cf := range properties {
			cfInfo, err := json.Marshal(cf)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal custom field info, custom field id: %d, err: %v", cf.PropertyID, err)
			}
			lines = append(lines, excel.Row{
				excel.NewCell(strconv.FormatInt(cf.PropertyID, 10)),
				excel.NewCell(cf.PropertyName),
				excel.NewCell(propertyType.String()),
				excel.NewCell(string(cfInfo)),
			})
		}
	}

	return lines, nil
}

func (data DataForFulfill) decodeCustomFieldSheet(excelSheets [][][]string) ([]*pb.IssuePropertyIndex, error) {
	if data.IsOldExcelFormat() {
		return nil, nil
	}
	sheet := excelSheets[indexOfSheetCustomField]
	var customFields []*pb.IssuePropertyIndex
	for i, row := range sheet {
		if i == 0 {
			continue
		}
		if len(row) != 4 {
			return nil, fmt.Errorf("invalid custom field row, row: %v", row)
		}
		var property pb.IssuePropertyIndex
		if err := json.Unmarshal([]byte(row[3]), &property); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom field detail, row: %v, err: %v", row, err)
		}
		customFields = append(customFields, &property)
	}
	return customFields, nil
}

// createCustomFieldIfNotExistsForImport
// 这里不考虑从 issues 里获取当前项目不存在的自定义字段并进行创建，因为:
// - 自定义字段是企业级别的，影响太大
// - 无法根据字段值判断值类型 (text, selection or others)
// - 无法判断是否必填
// - 无法判断和哪个类型关联
// - 即使强行创建为 text 类型，由于要被具体事项类型关联才可以使用，所以万一判断错了，想调整类型也不行，解绑会删除所有 issue 关联的值
func (data *DataForFulfill) createCustomFieldIfNotExistsForImport(originalCustomFields []*pb.IssuePropertyIndex) error {
	ctx := apis.WithInternalClientContext(context.Background(), "issue-import")

	originalCustomFieldsNeedCreate := make([]*pb.IssuePropertyIndex, 0, len(originalCustomFields))
	originalCommonCustomFieldsNeedCreate := make([]*pb.IssuePropertyIndex, 0, len(originalCustomFields))

	// 处理原有的自定义字段
	for _, originalCf := range originalCustomFields {
		originalCf := originalCf
		// 根据类型和名称，确认在当前企业是否已存在
		_, foundInCurrentOrg := data.CustomFieldMapByTypeName[originalCf.PropertyIssueType][originalCf.PropertyName]
		// 已存在，无需调整；可能存在 select 的枚举值不同等问题，不考虑，以当前企业为准
		if foundInCurrentOrg {
			continue
		}
		if originalCf.PropertyIssueType == pb.PropertyIssueTypeEnum_COMMON {
			originalCommonCustomFieldsNeedCreate = append(originalCommonCustomFieldsNeedCreate, originalCf)
		} else {
			originalCustomFieldsNeedCreate = append(originalCustomFieldsNeedCreate, originalCf)
		}
	}

	// 不考虑从 issues 里获取当前项目不存在的自定义字段并进行创建

	// do create
	// 需要先创建 common，再进行具体类型的创建，关联 common id
	for _, createCommon := range originalCommonCustomFieldsNeedCreate {
		polishPropertyValueEnumeratesForCreate(createCommon.EnumeratedValues)
		resp, err := data.ImportOnly.IssueCore.CreateIssueProperty(ctx, &pb.CreateIssuePropertyRequest{
			ScopeType:         createCommon.ScopeType,
			ScopeID:           data.OrgID, // 企业级别
			OrgID:             data.OrgID,
			PropertyName:      createCommon.PropertyName,
			DisplayName:       createCommon.DisplayName,
			PropertyType:      createCommon.PropertyType,
			Required:          createCommon.Required,
			PropertyIssueType: createCommon.PropertyIssueType,
			EnumeratedValues:  createCommon.EnumeratedValues,
			Relation:          0, // 0 for common
			IdentityInfo:      nil,
		})
		if err != nil {
			return fmt.Errorf("failed to create custom field, type: %s, name: %s, err: %v",
				pb.PropertyIssueTypeEnum_COMMON.String(), createCommon.PropertyName, err)
		}
		data.CustomFieldMapByTypeName[pb.PropertyIssueTypeEnum_COMMON][createCommon.PropertyName] = resp.Data
	}
	// 创建具体类型的自定义字段，都是基于 common 的引用
	for _, createCf := range originalCustomFieldsNeedCreate {
		// get common cf
		commonCf, ok := data.CustomFieldMapByTypeName[pb.PropertyIssueTypeEnum_COMMON][createCf.PropertyName]
		if !ok {
			return fmt.Errorf("failed to find corresponding common custom field, type: %s, name: %s",
				pb.PropertyIssueTypeEnum_COMMON.String(), createCf.PropertyName)
		}
		resp, err := data.ImportOnly.IssueCore.CreateIssueProperty(ctx, &pb.CreateIssuePropertyRequest{
			ScopeType:         createCf.ScopeType,
			ScopeID:           data.OrgID, // 企业级别
			OrgID:             data.OrgID,
			PropertyName:      commonCf.PropertyName,
			DisplayName:       commonCf.DisplayName,
			PropertyType:      commonCf.PropertyType,
			Required:          commonCf.Required,
			PropertyIssueType: createCf.PropertyIssueType,
			EnumeratedValues:  commonCf.EnumeratedValues, // ref
			Relation:          commonCf.PropertyID,       // ref
			IdentityInfo:      nil,
		})
		if err != nil {
			return fmt.Errorf("failed to create normal custom field, type: %s, name: %s, err: %v",
				createCf.PropertyIssueType.String(), createCf.PropertyName, err)
		}
		// resp.Data doesn't have EnumeratedValues
		resp.Data.EnumeratedValues = commonCf.EnumeratedValues
		data.CustomFieldMapByTypeName[createCf.PropertyIssueType][createCf.PropertyName] = resp.Data
	}

	// 有两个原因导致要重新获取自定义字段
	// 1. common 自定义字段的 enumeratesValues 是批量创建的，没有返回 value id
	// 2. 具体类型的自定义字段，resp.Data 里没有 enumeratesValues
	refreshed, err := RefreshDataCustomFields(data.OrgID, data.ImportOnly.IssueCore)
	if err != nil {
		return fmt.Errorf("failed to refresh custom fields, err: %v", err)
	}
	data.CustomFieldMapByTypeName = refreshed

	return nil
}

func polishPropertyValueEnumeratesForCreate(enumerates []*pb.Enumerate) {
	for i := range enumerates {
		enumerates[i].Id = 0
	}
}

func (data *DataForFulfill) createIssueCustomFieldRelation(issues []*issuedao.Issue, issueModelMapByIssueID map[uint64]*IssueSheetModel) error {
	ctx := apis.WithInternalClientContext(context.Background(), "issue-import")
	for _, issue := range issues {
		model, ok := issueModelMapByIssueID[issue.ID]
		if !ok {
			return fmt.Errorf("failed to find issue model by issue id, issue id: %d", issue.ID)
		}
		relationRequest := &pb.CreateIssuePropertyInstanceRequest{
			OrgID:        data.OrgID,
			ProjectID:    int64(data.ProjectID),
			IssueID:      int64(issue.ID),
			Property:     nil,
			IdentityInfo: nil,
		}
		var cfsNeedHandled []ExcelCustomField
		var cfType pb.PropertyIssueTypeEnum_PropertyIssueType
		switch issue.Type {
		case pb.IssueTypeEnum_REQUIREMENT.String():
			cfsNeedHandled = model.RequirementOnly.CustomFields
			cfType = pb.PropertyIssueTypeEnum_REQUIREMENT
		case pb.IssueTypeEnum_TASK.String():
			cfsNeedHandled = model.TaskOnly.CustomFields
			cfType = pb.PropertyIssueTypeEnum_TASK
		case pb.IssueTypeEnum_BUG.String():
			cfsNeedHandled = model.BugOnly.CustomFields
			cfType = pb.PropertyIssueTypeEnum_BUG
		default:
			return fmt.Errorf("invalid issue type, issue type: %s", issue.Type)
		}
		for _, cf := range cfsNeedHandled {
			if cf.Value == "" { // 兼容导出时就没有值的情况，比如后创建的自定义字段，之前的 issue 该字段没有值
				continue
			}
			property, ok := data.CustomFieldMapByTypeName[cfType][cf.Title]
			if !ok { // just ignore unknown custom field
				warnMsg := fmt.Sprintf("failed to find custom field, new issue id: %d, type: %s, name: %s", issue.ID, cfType.String(), cf.Title)
				data.ImportOnly.Warnings = append(data.ImportOnly.Warnings, warnMsg)
				continue
			}
			instance := &pb.IssuePropertyInstance{
				PropertyID:               property.PropertyID,
				ScopeID:                  property.ScopeID,
				ScopeType:                property.ScopeType,
				OrgID:                    property.OrgID,
				PropertyName:             property.PropertyName,
				DisplayName:              property.DisplayName,
				PropertyType:             property.PropertyType,
				Required:                 property.Required,
				PropertyIssueType:        property.PropertyIssueType,
				Relation:                 property.Relation,
				Index:                    property.Index,
				EnumeratedValues:         property.EnumeratedValues,
				Values:                   property.Values,
				RelatedIssue:             property.RelatedIssue, // related issue type
				ArbitraryValue:           nil,
				PropertyEnumeratedValues: nil,
			}
			if common.IsOptions(property.PropertyType.String()) {
				valuesInSheet := parseStringSliceByComma(cf.Value)
				for _, valueInSheet := range valuesInSheet {
					var foundEnumValue bool
					for _, enumValue := range property.EnumeratedValues {
						if enumValue.Name == valueInSheet {
							foundEnumValue = true
							instance.Values = append(instance.Values, enumValue.Id)
							break
						}
					}
					if !foundEnumValue {
						return fmt.Errorf("failed to find enum value by name, issue type: %s, property type: %s, property name: %s, enum value name: %s",
							property.PropertyIssueType, property.PropertyType, property.PropertyName, cf.Value)
					}
				}
			} else if property.PropertyType == pb.PropertyTypeEnum_Person { // 需要填写 userID
				userID, ok := data.ImportOnly.UserIDByNick[cf.Value]
				if !ok { // just log
					warnMsg := fmt.Sprintf("failed to find user id by nick in custom field, field name: %s, new issue id: %d, type: %s, name: %s, nick: %s",
						property.PropertyName, issue.ID, cfType.String(), cf.Title, cf.Value)
					data.ImportOnly.Warnings = append(data.ImportOnly.Warnings, warnMsg)
				}
				instance.ArbitraryValue = structpb.NewStringValue(userID)
			} else {
				instance.ArbitraryValue = structpb.NewStringValue(cf.Value)
			}
			relationRequest.Property = append(relationRequest.Property, instance)
		}
		_, err := data.ImportOnly.IssueCore.CreateIssuePropertyInstance(ctx, relationRequest)
		if err != nil {
			return fmt.Errorf("failed to create issue custom field relation, issue id: %d, err: %v", issue.ID, err)
		}
	}
	return nil
}

func RefreshDataCustomFields(orgID int64, i pb.IssueCoreServiceServer) (map[pb.PropertyIssueTypeEnum_PropertyIssueType]map[string]*pb.IssuePropertyIndex, error) {
	ctx := apis.WithInternalClientContext(context.Background(), "issue-import")
	resp, err := i.GetIssueProperty(ctx, &pb.GetIssuePropertyRequest{OrgID: orgID})
	if err != nil {
		return nil, fmt.Errorf("failed to batch get properties, err: %v", err)
	}
	customFieldMapByTypeName := make(map[pb.PropertyIssueTypeEnum_PropertyIssueType]map[string]*pb.IssuePropertyIndex)
	for i := range pb.PropertyIssueTypeEnum_PropertyIssueType_name { // ensure all types are included
		customFieldMapByTypeName[pb.PropertyIssueTypeEnum_PropertyIssueType(i)] = make(map[string]*pb.IssuePropertyIndex)
	}
	for _, v := range resp.Data {
		v := v
		customFieldMapByTypeName[v.PropertyIssueType][v.PropertyName] = v
	}
	return customFieldMapByTypeName, nil
}
