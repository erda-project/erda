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
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
)

var stateJoinSQL = "LEFT OUTER JOIN dice_issue_state_relations AS re on dice_issue_state.id=re.start_state_id"

type IssueStateJoinSQL struct {
	ID         int64                       `gorm:"primary_key"`
	ProjectID  uint64                      `gorm:"column:project_id"`
	IssueType  apistructs.IssueType        `gorm:"column:issue_type"`
	Name       string                      `gorm:"column:name"`
	Belong     apistructs.IssueStateBelong `gorm:"column:belong"`
	Index      int64                       `gorm:"column:index"`
	EndStateID int64                       `gorm:"column:end_state_id"`
}
type IssueState struct {
	BaseModel
	ProjectID uint64                      `gorm:"column:project_id"`
	IssueType apistructs.IssueType        `gorm:"column:issue_type"`
	Name      string                      `gorm:"column:name"`
	Belong    apistructs.IssueStateBelong `gorm:"column:belong"`
	Index     int64                       `gorm:"column:index"`
	Role      string                      `gorm:"column:role"`
}

func (IssueState) TableName() string {
	return "dice_issue_state"
}

type IssueStateRelation struct {
	BaseModel
	ProjectID    int64                `gorm:"column:project_id"`
	IssueType    apistructs.IssueType `gorm:"column:issue_type"`
	StartStateID int64                `gorm:"column:start_state_id"`
	EndStateID   int64                `gorm:"column:end_state_id"`
}

func (IssueStateRelation) TableName() string {
	return "dice_issue_state_relations"
}

func (client *DBClient) CreateIssuesState(state *IssueState) error {
	return client.Create(state).Error
}

func (client *DBClient) DeleteIssuesState(id int64) error {
	return client.Table("dice_issue_state").Where("id = ?", id).Delete(IssueState{}).Error
}

func (client *DBClient) DeleteIssuesStateRelationByStartID(id int64) error {
	return client.Table("dice_issue_state_relations").Where("start_state_id = ?", id).Delete(IssueStateRelation{}).Error
}

func (client *DBClient) DeleteIssuesStateByProjectID(projectID int64) error {
	err := client.Table("dice_issue_state").Where("project_id = ?", projectID).Delete(IssueState{}).Error
	if err != nil {
		return err
	}
	err = client.Table("dice_issue_state_relations").Where("project_id = ?", projectID).Delete(IssueStateRelation{}).Error
	return err
}

func (client *DBClient) UpdateIssuesStateIndex(id int64, index int64) error {
	return client.Table("dice_issue_state").Where("id = ?", id).Update("index", index).Error
}

func (client *DBClient) UpdateIssueStateRelations(projectID int64, issueType apistructs.IssueType, StateRelations []IssueStateRelation) error {
	if err := client.Table("dice_issue_state_relations").Where("project_id = ?", projectID).
		Where("issue_type = ?", issueType).Delete(IssueStateRelation{}).Error; err != nil {
		return err
	}
	return client.BulkInsert(StateRelations)
}

func (client *DBClient) UpdateIssueState(state *IssueState) error {
	return client.Save(state).Error
}

func (client *DBClient) UpdateIssuesStateBelong(id int64, belong apistructs.IssueStateBelong) error {
	return client.Table("dice_issue_state_relations").Where("id = ?", id).Update("state_belong", belong).Error
}

func (client *DBClient) GetIssuesStateRelations(projectID uint64, issueType apistructs.IssueType) ([]IssueStateJoinSQL, error) {
	var issueStateJoinSql []IssueStateJoinSQL
	db := client.Table("dice_issue_state").Select("*").Joins(stateJoinSQL).Where("dice_issue_state.project_id = ?", projectID)
	if issueType != "" {
		db = db.Where("dice_issue_state.issue_type = ?", issueType)
	}
	if err := db.Order("dice_issue_state.issue_type,dice_issue_state.index").
		Select("dice_issue_state.id,dice_issue_state.project_id,dice_issue_state.issue_type,dice_issue_state.name,dice_issue_state.belong,dice_issue_state.index,re.end_state_id").
		Scan(&issueStateJoinSql).Error; err != nil {
		return nil, err
	}
	return issueStateJoinSql, nil
}

func (client *DBClient) GetIssuesStatesByProjectID(projectID uint64, issueType apistructs.IssueType) ([]IssueState, error) {
	var states []IssueState
	db := client.Table("dice_issue_state").Where("project_id = ?", projectID)
	if issueType != "" {
		db = db.Where("issue_type = ?", issueType)
	}
	if err := db.Order("index").Find(&states).Error; err != nil {
		return nil, err
	}
	return states, nil
}

// get all state by projectID list
func (client *DBClient) GetIssuesStatesByProjectIDList(projectIDList []uint64) ([]IssueState, error) {
	var states []IssueState
	db := client.Table("dice_issue_state").Where("project_id in (?)", projectIDList)
	if err := db.Order("index").Find(&states).Error; err != nil {
		return nil, err
	}
	return states, nil
}

func (client *DBClient) GetIssueStateByID(ID int64) (*IssueState, error) {
	var state IssueState
	db := client.Table("dice_issue_state").Where("id = ?", ID)
	if err := db.Find(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

func (client *DBClient) GetIssueStateByIDs(ID []int64) ([]IssueState, error) {
	var state []IssueState
	db := client.Table("dice_issue_state").Where("id in (?)", ID)
	if err := db.Find(&state).Error; err != nil {
		return nil, err
	}
	return state, nil
}

func (client *DBClient) GetIssueStatePermission(role string, st int64, ed int64) (*IssueStateRelation, error) {
	// 是否有权限转移至ed状态
	var state IssueState
	db := client.Table("dice_issue_state").Where("id = ?", ed).Where("role LIKE ?", "%"+role+"%")
	if err := db.Find(&state).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
	}
	// 有权限，继续搜索是否能从st状态到达ed状态
	var permission IssueStateRelation
	db = client.Where("start_state_id = ?", st).
		Where("end_state_id = ?", ed)
	if err := db.Limit(1).Find(&permission).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

// GetClosedBugState 获取一个项目下已关闭的bug状态id
func (client *DBClient) GetClosedBugState(projectID int64) ([]IssueState, error) {
	var state []IssueState
	db := client.Table("dice_issue_state").Where("project_id = ?", projectID).Where("issue_type = ?", "BUG").
		Where("belong = ?", "CLOSED")
	if err := db.Find(&state).Error; err != nil {
		return nil, err
	}
	return state, nil
}
