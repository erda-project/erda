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

package dao

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/model"
)

// CreateBranchRule 创建分支规则
func (client *DBClient) CreateBranchRule(branchRule *model.BranchRule) error {
	return client.Create(branchRule).Error
}

// UpdateBranchRule 更新分支规则
func (client *DBClient) UpdateBranchRule(branchRule *model.BranchRule) error {
	return client.Save(branchRule).Error
}

// GetBranchRule 获取分支规则
func (client *DBClient) GetBranchRule(id int64) (model.BranchRule, error) {
	var branchRule model.BranchRule
	err := client.Where("id = ?", id).Find(&branchRule).Error
	return branchRule, err
}

// DeleteBranchRule 删除分支规则
func (client *DBClient) DeleteBranchRule(id int64) error {
	return client.Where("id = ?", id).Delete(&model.BranchRule{}).Error
}

// DeleteBranchRuleByScope 批量删除分支规则
func (client *DBClient) DeleteBranchRuleByScope(scopeType apistructs.ScopeType, scopeID int64) error {
	return client.Where("scope_type =? and scope_id=?", scopeType, scopeID).Delete(&model.BranchRule{}).Error
}

// QueryBranchRules 查询分支列表
func (client *DBClient) QueryBranchRules(scopeType apistructs.ScopeType, scopeID int64) ([]model.BranchRule, error) {
	var result []model.BranchRule
	err := client.Model(&model.BranchRule{}).Where("scope_type =? and scope_id=?", scopeType, scopeID).Find(&result).Error
	return result, err
}

func (client *DBClient) QueryBranchRulesByScope(scopeType apistructs.ScopeType) ([]model.BranchRule, error) {
	var result []model.BranchRule
	err := client.Model(&model.BranchRule{}).Where("scope_type =?", scopeType).Find(&result).Error
	return result, err
}

// GetBranchRulesCount 查询分支规则数量
func (client *DBClient) GetBranchRulesCount(scopeType apistructs.ScopeType) (int64, error) {
	var count int64
	err := client.Model(&model.BranchRule{}).Where("scope_type =?", scopeType).Count(&count).Error
	return count, err
}
