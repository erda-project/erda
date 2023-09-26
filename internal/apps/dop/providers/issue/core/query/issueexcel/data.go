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
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"

	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/conf"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/internal/apps/dop/services/apierrors"
	legacydao "github.com/erda-project/erda/internal/core/legacy/dao"
	"github.com/erda-project/erda/pkg/i18n"
	"github.com/erda-project/erda/pkg/strutil"
)

type DataForFulfill struct {
	ExportOnly DataForFulfillExportOnly
	ImportOnly DataForFulfillImportOnly

	// common
	Bdl                   *bundle.Bundle
	UserID                string
	OrgID                 int64
	ProjectID             uint64
	Locale                *i18n.LocaleResource
	StageMap              map[query.IssueStage]string
	IterationMapByID      map[int64]*dao.Iteration           // key: iteration id
	IterationMapByName    map[string]*dao.Iteration          // key: iteration name
	StateMap              map[int64]string                   // key: state id
	StateMapByTypeAndName map[string]map[string]int64        // key: state name
	ProjectMemberByUserID map[string]apistructs.Member       // key: user id
	OrgMemberByUserID     map[string]apistructs.Member       // key: user id
	LabelMapByName        map[string]apistructs.ProjectLabel // key: label name

	// CustomFieldMapByTypeName outerKey: property type, innerKey: property name (可以直接使用 2 层 map 判断，每个类型保证不为空)
	// 在同一类型下，property name 唯一
	// 正常顺序是，先创建 common，然后被具体类型引用
	CustomFieldMapByTypeName map[pb.PropertyIssueTypeEnum_PropertyIssueType]map[string]*pb.IssuePropertyIndex

	AlreadyHaveProjectOwner bool
}

type DataForFulfillExportOnly struct {
	AllProjectIssues         bool // 全量导出
	FileNameWithExt          string
	Issues                   []*pb.Issue
	IsDownloadTemplate       bool
	IssuePropertyRelationMap map[int64][]dao.IssuePropertyRelation
	InclusionMap             map[int64][]int64 // key: issue id
	ConnectionMap            map[int64][]int64 // key: issue id
	States                   []dao.IssueState
	StateRelations           []dao.IssueStateJoinSQL
	PropertyEnumMap          map[query.PropertyEnumPair]string
}
type DataForFulfillImportOnly struct {
	LabelDB   *legacydao.DBClient
	DB        *dao.DBClient
	Identity  userpb.UserServiceServer
	IssueCore pb.IssueCoreServiceServer

	BaseInfo *DataForFulfillImportOnlyBaseInfo

	IsOldExcelFormat bool

	CurrentProjectIssueMap map[uint64]bool

	UserIDByNick map[string]string // key: nick, value: userID

	Warnings []string // used to record warnings
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

func (data *DataForFulfill) CheckPermission() error {
	// 如果是全量导出，则只有项目管理员和项目经理有权限
	if !data.IsFullExport() {
		return nil
	}
	roleResp, err := data.Bdl.ScopeRoleAccess(data.UserID, &apistructs.ScopeRoleAccessRequest{
		Scope: apistructs.Scope{
			Type: apistructs.ProjectScope,
			ID:   strutil.String(data.ProjectID),
		},
	})
	if err != nil {
		return apierrors.ErrExportExcelIssue.InternalError(err)
	}
	if !roleResp.Access {
		return apierrors.ErrExportExcelIssue.AccessDenied()
	}
	var canExportAll bool
	for _, role := range roleResp.Roles {
		if role == bundle.RoleProjectOwner || role == bundle.RoleProjectPM {
			canExportAll = true
			break
		}
	}
	if canExportAll {
		return nil
	}
	return apierrors.ErrExportExcelIssue.AccessDenied(fmt.Errorf("only project Owner or PM can export all issues"))
}

func (data *DataForFulfill) IsFullExport() bool {
	return data.ExportOnly.AllProjectIssues
}

func formatTimeFromTimestamp(timestamp *timestamp.Timestamp) string {
	return timestamp.AsTime().In(time.Local).Format("2006-01-02 15:04:05")
}

func formatIssueCustomFields(issue *pb.Issue, propertyType pb.PropertyIssueTypeEnum_PropertyIssueType, data DataForFulfill) []ExcelCustomField {
	var results []ExcelCustomField
	for _, customField := range data.CustomFieldMapByTypeName[propertyType] {
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
	return query.GetCustomPropertyColumnValue(customField, data.ExportOnly.IssuePropertyRelationMap[issue.Id], data.ExportOnly.PropertyEnumMap, data.ProjectMemberByUserID)
}
