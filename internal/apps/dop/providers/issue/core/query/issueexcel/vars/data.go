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
	"fmt"

	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/conf"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/internal/apps/dop/services/apierrors"
	legacydao "github.com/erda-project/erda/internal/core/legacy/dao"
	"github.com/erda-project/erda/pkg/excel"
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

	IsOldExcelFormat bool

	CurrentProjectIssueMap map[uint64]bool

	OrgMemberByDesensitizedKey map[string]string // key: phone / email / nick / name
	UserIDByNick               map[string]string // key: nick, value: userID

	Warnings []string // used to record warnings

	Sheets SheetsInfo
}

type (
	SheetsInfo struct {
		Must     MustSheetsInfo
		Optional OptionalSheetsInfo
	}
	MustSheetsInfo struct {
		IssueInfo []IssueSheetModel
	}
	OptionalSheetsInfo struct {
		BaseInfo        *DataForFulfillImportOnlyBaseInfo
		UserInfo        []apistructs.Member
		LabelInfo       []*pb.ProjectLabel
		CustomFieldInfo []*pb.IssuePropertyIndex
		IterationInfo   []*dao.Iteration
		StateInfo       *StateInfo
	}
	DataForFulfillImportOnlyBaseInfo struct {
		OriginalErdaPlatform  string // get from dop conf.DiceClusterName()
		OriginalErdaProjectID uint64
		AllProjectIssues      bool
	}
	StateInfo struct {
		States        []dao.IssueState
		StateJoinSQLs []dao.IssueStateJoinSQL
	}
)

func (data DataForFulfill) ShouldUpdateWhenIDSame() bool {
	// no base info sheet, trust as simple import, can do id-same-update if issue-id found in current project
	if data.ImportOnly.Sheets.Optional.BaseInfo == nil {
		return true
	}
	// only can do id-same-update when erda-platform is same && project-id is same
	if data.IsSameErdaPlatform() {
		if data.ImportOnly.Sheets.Optional.BaseInfo.OriginalErdaProjectID == data.ProjectID {
			return true
		}
	}
	// all other cases, can not do id-same-update
	return false
}

func (data DataForFulfill) IsSameErdaPlatform() bool {
	if data.ImportOnly.Sheets.Optional.BaseInfo == nil {
		panic("baseInfo is nil")
	}
	return data.ImportOnly.Sheets.Optional.BaseInfo.OriginalErdaPlatform == conf.DiceClusterName()
}

func (data DataForFulfill) IsOldExcelFormat() bool {
	return data.ImportOnly.IsOldExcelFormat
}

// JudgeIfIsOldExcelFormat old Excel format have only one sheet and excel[0][0][0] = "ID"
func (data *DataForFulfill) JudgeIfIsOldExcelFormat(df excel.DecodedFile) {
	// only one sheet
	if len(df.Sheets.L) != 1 {
		return
	}

	// [0][0] is ID
	if colLen := len(df.Sheets.L[0].UnmergedSlice); colLen == 0 {
		return
	}
	if rowLen := len(df.Sheets.L[0].UnmergedSlice[0]); rowLen == 0 {
		return
	}
	if df.Sheets.L[0].UnmergedSlice[0][0] != "ID" {
		return
	}

	// set value
	data.ImportOnly.IsOldExcelFormat = true
	data.ImportOnly.Sheets.Optional.BaseInfo = &DataForFulfillImportOnlyBaseInfo{
		OriginalErdaPlatform:  "",
		OriginalErdaProjectID: 0,
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
	if data.ExportOnly.IsDownloadTemplate {
		return false
	}
	return data.ExportOnly.AllProjectIssues
}
