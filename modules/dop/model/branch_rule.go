// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package model

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type BranchRule struct {
	dbengine.BaseModel

	ScopeType         apistructs.ScopeType
	ScopeID           int64
	Rule              string
	IsProtect         bool
	IsTriggerPipeline bool
	NeedApproval      bool
	Desc              string //规则说明
	Workspace         string `json:"workspace"`
	ArtifactWorkspace string `json:"artifactWorkspace"`
}

// TableName 设置模型对应数据库表名称
func (BranchRule) TableName() string {
	return "dice_branch_rules"
}

func (rule *BranchRule) ToApiData() *apistructs.BranchRule {
	return &apistructs.BranchRule{
		ID:                int64(rule.ID),
		Rule:              rule.Rule,
		ScopeID:           rule.ScopeID,
		ScopeType:         rule.ScopeType,
		IsProtect:         rule.IsProtect,
		NeedApproval:      rule.NeedApproval,
		IsTriggerPipeline: rule.IsTriggerPipeline,
		Desc:              rule.Desc,
		Workspace:         rule.Workspace,
		ArtifactWorkspace: rule.ArtifactWorkspace,
	}
}
