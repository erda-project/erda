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

	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/conf"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	legacydao "github.com/erda-project/erda/internal/core/legacy/dao"
	"github.com/erda-project/erda/pkg/i18n"
)

type DataForFulfill struct {
	ExportOnly DataForFulfillExportOnly
	ImportOnly DataForFulfillImportOnly

	// common
	OrgID                 int64
	ProjectID             uint64
	Locale                *i18n.LocaleResource
	StageMap              map[query.IssueStage]string
	IterationMapByID      map[int64]*dao.Iteration           // key: iteration id
	IterationMapByName    map[string]*dao.Iteration          // key: iteration name
	StateMap              map[int64]string                   // key: state id
	StateMapByTypeAndName map[string]map[string]int64        // key: state name
	ProjectMemberMap      map[string]apistructs.Member       // key: user id
	OrgMemberMap          map[string]apistructs.Member       // key: user id
	LabelMapByName        map[string]apistructs.ProjectLabel // key: label name

	CustomFieldMap map[pb.PropertyIssueTypeEnum_PropertyIssueType][]*pb.IssuePropertyIndex
	//CustomFieldMapByName map[string]*pb.IssuePropertyIndex
	PropertyEnumMap map[query.PropertyEnumPair]string
}

type DataForFulfillExportOnly struct {
	FileNameWithExt          string
	Issues                   []*pb.Issue
	IsDownloadTemplate       bool
	IssuePropertyRelationMap map[int64][]dao.IssuePropertyRelation
	InclusionMap             map[int64][]int64 // key: issue id
	ConnectionMap            map[int64][]int64 // key: issue id
}
type DataForFulfillImportOnly struct {
	LabelDB  *legacydao.DBClient
	DB       *dao.DBClient
	Bdl      *bundle.Bundle
	Identity userpb.UserServiceServer
	Property pb.IssueCoreServiceServer

	BaseInfo *DataForFulfillImportOnlyBaseInfo

	IsOldExcelFormat bool

	CurrentProjectIssueMap map[uint64]bool
}

func (data DataForFulfill) ShouldUpdateWhenIDSame() bool {
	// only can do id-same-update when erda-platform is same && project-id is same
	if data.IsSameErdaPlatform() {
		if data.ImportOnly.BaseInfo.OriginalErdaProjectID == data.ProjectID {
			return true
		}
	}
	// all other cases, can not do id-same-update
	return false
}

func (data DataForFulfill) IsSameErdaPlatform() bool {
	if data.ImportOnly.BaseInfo == nil {
		panic("baseInfo is nil")
	}
	return data.ImportOnly.BaseInfo.OriginalErdaPlatform == conf.DiceClusterName()
}

func (data DataForFulfill) IsOldExcelFormat() bool {
	return data.ImportOnly.IsOldExcelFormat
}

// JudgeIfIsOldExcelFormat old Excel format have only one sheet
func (data *DataForFulfill) JudgeIfIsOldExcelFormat(excelSheets [][][]string) {
	isOld := len(excelSheets) == 1
	if isOld {
		data.ImportOnly.IsOldExcelFormat = true
		data.ImportOnly.BaseInfo = &DataForFulfillImportOnlyBaseInfo{
			OriginalErdaPlatform:  "",
			OriginalErdaProjectID: 0,
		}
	}
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
	return query.GetCustomPropertyColumnValue(customField, data.ExportOnly.IssuePropertyRelationMap[issue.Id], data.PropertyEnumMap, data.ProjectMemberMap)
}
