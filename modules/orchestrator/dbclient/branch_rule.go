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

package dbclient

import (
	"github.com/erda-project/erda/apistructs"
)

type BranchRule struct {
	ID                uint64
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

// QueryBranchRules 查询分支列表
func (db *DBClient) QueryBranchRules(scopeType apistructs.ScopeType, scopeID uint64) ([]BranchRule, error) {
	var result []BranchRule
	err := db.Model(&BranchRule{}).Where("scope_type =? and scope_id=?", scopeType, scopeID).Find(&result).Error
	return result, err
}

func (db *DBClient) QueryBranchRulesByScope(scopeType apistructs.ScopeType) ([]BranchRule, error) {
	var result []BranchRule
	err := db.Model(&BranchRule{}).Where("scope_type =?", scopeType).Find(&result).Error
	return result, err
}

// GetBranchRulesCount 查询分支规则数量
func (db *DBClient) GetBranchRulesCount(scopeType apistructs.ScopeType) (int64, error) {
	var count int64
	err := db.Model(&BranchRule{}).Where("scope_type =?", scopeType).Count(&count).Error
	return count, err
}
