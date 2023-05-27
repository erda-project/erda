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
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/pkg/i18n"
)

type DataForFulfill struct {
	FileNameWithExt    string
	Locale             *i18n.LocaleResource
	CustomFieldMap     map[pb.PropertyIssueTypeEnum_PropertyIssueType][]*pb.IssuePropertyIndex
	Issues             []*pb.Issue
	ProjectID          uint64
	OrgID              int64
	IsDownloadTemplate bool
	StageMap           map[query.IssueStage]string
	IterationMap       map[int64]string // key: iteration id
	StateMap           map[int64]string // key: state id
	UsernameMap        map[string]string
	InclusionMap       map[int64][]int64 // key: issue id
	ConnectionMap      map[int64][]int64 // key: issue id

	PropertyRelationMap map[int64][]dao.IssuePropertyRelation
	PropertyEnumMap     map[query.PropertyEnumPair]string
}

func formatTimeFromTimestamp(timestamp *timestamp.Timestamp) string {
	return timestamp.AsTime().In(time.Local).Format("2006-01-02 15:04:05")
}

func formatIssueCustomFields(issue *pb.Issue, propertyType pb.PropertyIssueTypeEnum_PropertyIssueType, data DataForFulfill) []ExcelCustomField {
	var results []ExcelCustomField
	for _, customField := range data.CustomFieldMap[propertyType] {
		results = append(results, formatOneCustomField(customField, issue, data))
	}
	return results
}

func formatOneCustomField(cf *pb.IssuePropertyIndex, issue *pb.Issue, data DataForFulfill) ExcelCustomField {
	return ExcelCustomField{
		Title: cf.PropertyName,
		Value: getCustomFieldValue(cf, issue, data),
	}
}

func getCustomFieldValue(customField *pb.IssuePropertyIndex, issue *pb.Issue, data DataForFulfill) string {
	return query.GetCustomPropertyColumnValue(customField, data.PropertyRelationMap[issue.Id], data.PropertyEnumMap, data.UsernameMap)
}
