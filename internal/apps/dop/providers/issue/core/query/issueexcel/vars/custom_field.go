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

package vars

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
)

func FormatTimeFromTimestamp(timestamp *timestamppb.Timestamp) string {
	return timestamp.AsTime().In(time.Local).Format("2006-01-02 15:04:05")
}

func FormatIssueCustomFields(issue *pb.Issue, propertyType pb.PropertyIssueTypeEnum_PropertyIssueType, data *DataForFulfill) []ExcelCustomField {
	var results []ExcelCustomField
	for _, customField := range data.CustomFieldMapByTypeName[propertyType] {
		results = append(results, FormatOneCustomField(customField, issue, data))
	}
	return results
}

func FormatOneCustomField(cf *pb.IssuePropertyIndex, issue *pb.Issue, data *DataForFulfill) ExcelCustomField {
	return ExcelCustomField{
		Title: cf.PropertyName,
		Value: GetCustomFieldValue(cf, issue, data),
	}
}

func GetCustomFieldValue(customField *pb.IssuePropertyIndex, issue *pb.Issue, data *DataForFulfill) string {
	return query.GetCustomPropertyColumnValue(customField, data.ExportOnly.IssuePropertyRelationMap[issue.Id], data.ExportOnly.PropertyEnumMap, data.ProjectMemberByUserID)
}
