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
	for propertyType, properties := range data.CustomFieldMap {
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

func (data *DataForFulfill) createCustomFieldIfNotExistsForImport(originalCustomFields []*pb.IssuePropertyIndex) error {
	ctx := apis.WithInternalClientContext(context.Background(), "issue-import")
	for _, originalCf := range originalCustomFields {
		currentProjectProperties := data.CustomFieldMap[originalCf.PropertyIssueType]
		var found bool
		for _, currentProjectProperty := range currentProjectProperties {
			if originalCf.PropertyName == currentProjectProperty.PropertyName {
				found = true
				break
			}
		}
		if found {
			continue
		}
		// not exists, do create
		// not exists
		resp, err := data.ImportOnly.Property.CreateIssueProperty(ctx, &pb.CreateIssuePropertyRequest{
			ScopeType:         originalCf.ScopeType,
			ScopeID:           int64(data.ProjectID),
			OrgID:             data.OrgID,
			PropertyName:      originalCf.PropertyName,
			DisplayName:       originalCf.DisplayName,
			PropertyType:      originalCf.PropertyType,
			Required:          originalCf.Required,
			PropertyIssueType: originalCf.PropertyIssueType,
			EnumeratedValues:  originalCf.EnumeratedValues,
			Relation:          originalCf.Relation,
			IdentityInfo:      nil,
		})
		if err != nil {
			return fmt.Errorf("failed to create custom field, custom field name: %s, err: %v", originalCf.PropertyName, err)
		}
		// update data map
		newCf := resp.Data
		data.CustomFieldMap[newCf.PropertyIssueType] = append(data.CustomFieldMap[newCf.PropertyIssueType], newCf)
		//if common.IsOptions(newCf.PropertyType.String()) {
		//	for _, enumValue := range newCf.EnumeratedValues {
		//		data.PropertyEnumMap[query.PropertyEnumPair{PropertyID: newCf.PropertyID, ValueID: enumValue.Id}] = enumValue.Name
		//	}
		//}
	}
	return nil
}

func (data DataForFulfill) createIssueCustomFieldRelation(issues []*issuedao.Issue, issueModelMapByIssueID map[uint64]*IssueSheetModel) error {
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
			properties := data.CustomFieldMap[cfType]
			var found bool
			for _, property := range properties {
				property := property
				if property.PropertyName == cf.Title {
					found = true
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
								return fmt.Errorf("failed to find enum value by name, enum value name: %s", cf.Value)
							}
						}
					} else {
						instance.ArbitraryValue = structpb.NewStringValue(cf.Value)
					}
					relationRequest.Property = append(relationRequest.Property, instance)
					break
				}
			}
			if !found {
				return fmt.Errorf("failed to find custom field by name, custom field name: %s", cf.Title)
			}
		}
		_, err := data.ImportOnly.Property.CreateIssuePropertyInstance(ctx, relationRequest)
		if err != nil {
			return fmt.Errorf("failed to create issue custom field relation, issue id: %d, err: %v", issue.ID, err)
		}
	}
	return nil
}
