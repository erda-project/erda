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

package issuestate

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
)

// IssueState issue state service 对象
type IssueState struct {
	bdl *bundle.Bundle
	db  IssueStater
}

// Option 定义 IssueState 对象配置选项
type Option func(*IssueState)

// New 新建 issue state 对象
func New(options ...Option) *IssueState {
	is := &IssueState{}
	for _, op := range options {
		op(is)
	}
	return is
}

// WithDBClient 配置 db client
func WithDBClient(db IssueStater) Option {
	return func(is *IssueState) {
		is.db = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(is *IssueState) {
		is.bdl = bdl
	}
}

// CreateIssueState 创建事件状态请求
func (is *IssueState) CreateIssueState(req *apistructs.IssueStateCreateRequest) (*dao.IssueState, error) {
	states, err := is.db.GetIssuesStatesByProjectID(req.ProjectID, req.IssueType)
	var maxIndex int64 = -1
	for _, v := range states {
		if v.Index > maxIndex {
			maxIndex = v.Index
		}
	}
	if err != nil {
		return nil, err
	}
	createState := &dao.IssueState{
		ProjectID: req.ProjectID,
		IssueType: req.IssueType,
		Name:      req.StateName,
		Belong:    req.StateBelong,
		Index:     maxIndex + 1,
		Role:      "Ops,Dev,QA,Owner,Lead",
	}
	if err = is.db.CreateIssuesState(createState); err != nil {
		return nil, err
	}
	return createState, nil
}

// DeleteIssueState 删除事件状态请求
func (is *IssueState) DeleteIssueState(stateID int64) error {
	// 如果有事件是该状态则不可删除
	_, err := is.db.GetIssueByState(stateID)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return err
		}
	} else {
		return apierrors.ErrDeleteIssueState.InvalidState("有事件处于该状态,不可删除")
	}
	// 删除该状态的关联
	if err := is.db.DeleteIssuesStateRelationByStartID(stateID); err != nil {
		return err
	}
	// 删除状态
	if err := is.db.DeleteIssuesState(stateID); err != nil {
		return err
	}
	return nil
}

// GetIssueStates 获取状态列表请求
func (is *IssueState) GetIssueStates(req *apistructs.IssueStatesGetRequest) ([]apistructs.IssueTypeState, error) {
	var states []apistructs.IssueTypeState
	st, err := is.db.GetIssuesStatesByProjectID(req.ProjectID, "")
	if err != nil {
		return nil, err
	}
	states = append(states, apistructs.IssueTypeState{
		IssueType: apistructs.IssueTypeTask,
	})
	states = append(states, apistructs.IssueTypeState{
		IssueType: apistructs.IssueTypeRequirement,
	})
	states = append(states, apistructs.IssueTypeState{
		IssueType: apistructs.IssueTypeBug,
	})
	states = append(states, apistructs.IssueTypeState{
		IssueType: apistructs.IssueTypeEpic,
	})
	for _, v := range st {
		if v.IssueType == apistructs.IssueTypeTask {
			states[0].State = append(states[0].State, v.Name)
		} else if v.IssueType == apistructs.IssueTypeRequirement {
			states[1].State = append(states[1].State, v.Name)
		} else if v.IssueType == apistructs.IssueTypeBug {
			states[2].State = append(states[2].State, v.Name)
		} else if v.IssueType == apistructs.IssueTypeEpic {
			states[3].State = append(states[3].State, v.Name)
		}
	}
	return states, nil
}

func (is *IssueState) GetIssueStatesMap(req *apistructs.IssueStatesGetRequest) (map[apistructs.IssueType][]apistructs.IssueStatus, error) {
	stateMap := map[apistructs.IssueType][]apistructs.IssueStatus{
		apistructs.IssueTypeRequirement: make([]apistructs.IssueStatus, 0),
		apistructs.IssueTypeTask:        make([]apistructs.IssueStatus, 0),
		apistructs.IssueTypeBug:         make([]apistructs.IssueStatus, 0),
	}
	st, err := is.db.GetIssuesStatesByProjectID(req.ProjectID, "")
	if err != nil {
		return nil, err
	}
	for _, v := range st {
		if _, ok := stateMap[v.IssueType]; ok {
			stateMap[v.IssueType] = append(stateMap[v.IssueType], apistructs.IssueStatus{
				ProjectID:   v.ProjectID,
				IssueType:   v.IssueType,
				StateID:     int64(v.ID),
				StateName:   v.Name,
				StateBelong: v.Belong,
				Index:       v.Index,
			})
		}
	}
	return stateMap, nil
}

func (is *IssueState) GetIssueStateIDs(req *apistructs.IssueStatesGetRequest) ([]int64, error) {
	st, err := is.db.GetIssuesStates(req)
	if err != nil {
		return nil, err
	}
	res := make([]int64, 0)
	for _, v := range st {
		res = append(res, int64(v.ID))
	}
	return res, nil
}

func (is *IssueState) GetIssuesStatesByID(id int64) (*apistructs.IssueStatus, error) {
	state, err := is.db.GetIssueStateByID(id)
	if err != nil {
		return nil, err
	}
	status := &apistructs.IssueStatus{
		ProjectID:   state.ProjectID,
		IssueType:   state.IssueType,
		StateID:     int64(state.ID),
		StateName:   state.Name,
		StateBelong: state.Belong,
		Index:       state.Index,
	}
	return status, nil
}

func (is *IssueState) GetIssueStatesBelong(req *apistructs.IssueStateRelationGetRequest) ([]apistructs.IssueStateState, error) {
	var states []apistructs.IssueStateState
	st, err := is.db.GetIssuesStatesByProjectID(req.ProjectID, req.IssueType)
	if err != nil {
		return nil, err
	}
	BelongMap := make(map[apistructs.IssueStateBelong][]apistructs.IssueStateName)
	for _, s := range st {
		BelongMap[s.Belong] = append(BelongMap[s.Belong], apistructs.IssueStateName{
			Name: s.Name,
			ID:   int64(s.ID),
		})
	}
	stateIndex := req.IssueType.GetStateBelongIndex()
	for _, state := range stateIndex {
		for key, value := range BelongMap {
			if key != state {
				continue
			}
			states = append(states, apistructs.IssueStateState{
				StateBelong: key,
				States:      value,
			})
		}
	}
	return states, nil
}

func (is *IssueState) GetIssuesStatesNameByID(id []int64) ([]apistructs.IssueStatus, error) {
	state, err := is.db.GetIssueStateByIDs(id)
	if err != nil {
		return nil, err
	}
	var status []apistructs.IssueStatus
	for _, v := range state {
		status = append(status, apistructs.IssueStatus{
			ProjectID:   v.ProjectID,
			IssueType:   v.IssueType,
			StateID:     int64(v.ID),
			StateName:   v.Name,
			StateBelong: v.Belong,
			Index:       v.Index,
		})
	}

	return status, nil
}

func (is *IssueState) InitProjectState(projectID int64) error {
	var (
		states    []dao.IssueState
		relations []dao.IssueStateRelation
	)
	relation := []int64{
		0, 1, 0, 2, 0, 3, 1, 2, 1, 3, 2, 3, 1, 0, 2, 0, 3, 0, 2, 1, 3, 1, 3, 2,
		4, 5, 5, 6,
		7, 8, 8, 9,
		10, 14, 15, 14, 12, 15, 13, 15, 14, 15, 10, 12, 15, 12, 10, 13, 15, 13, 12, 16, 13, 16, 14, 16, 10, 11, 11, 12, 11, 13, 11, 14, 11, 15, 15, 11,
		17, 20, 21, 20, 18, 21, 19, 21, 20, 21, 17, 18, 21, 18, 17, 19, 21, 19, 18, 22, 19, 22, 20, 22,
	}
	name := []string{
		"待处理", "进行中", "测试中", "已完成",
		"待处理", "进行中", "已完成",
		"待处理", "进行中", "已完成",
		"待处理", "进行中", "无需修复", "重复提交", "已解决", "重新打开", "已关闭",
		"待处理", "无需修复", "重复提交", "已解决", "重新打开", "已关闭",
	}
	belong := []apistructs.IssueStateBelong{
		"OPEN", "WORKING", "WORKING", "DONE",
		"OPEN", "WORKING", "DONE",
		"OPEN", "WORKING", "DONE",
		"OPEN", "WORKING", "WONTFIX", "WONTFIX", "RESOLVED", "REOPEN", "CLOSED",
		"OPEN", "WONTFIX", "WONTFIX", "RESOLVED", "REOPEN", "CLOSED",
	}
	index := []int64{
		0, 1, 2, 3,
		0, 1, 2,
		0, 1, 2,
		0, 1, 2, 3, 4, 5, 6,
		0, 1, 2, 3, 4, 5,
	}
	// state
	for i := 0; i < 23; i++ {
		states = append(states, dao.IssueState{
			ProjectID: uint64(projectID),
			Name:      name[i],
			Belong:    belong[i],
			Index:     index[i],
			Role:      "Ops,Dev,QA,Owner,Lead",
		})
		if i < 4 {
			states[i].IssueType = apistructs.IssueTypeRequirement
		} else if i < 7 {
			states[i].IssueType = apistructs.IssueTypeTask
		} else if i < 10 {
			states[i].IssueType = apistructs.IssueTypeEpic
		} else if i < 17 {
			states[i].IssueType = apistructs.IssueTypeBug
		} else if i < 23 {
			states[i].IssueType = apistructs.IssueTypeTicket
		}
		//if err := is.db.CreateIssuesState(&states[i]); err != nil {
		//	return err
		//}
	}
	// state relation
	for i := 0; i < 46; i++ {
		relations = append(relations, dao.IssueStateRelation{
			ProjectID:    projectID,
			StartStateID: int64(states[relation[i*2]].ID),
			EndStateID:   int64(states[relation[i*2+1]].ID),
		})
		if i < 12 {
			relations[i].IssueType = apistructs.IssueTypeRequirement
		} else if i < 14 {
			relations[i].IssueType = apistructs.IssueTypeTask
		} else if i < 16 {
			relations[i].IssueType = apistructs.IssueTypeEpic
		} else if i < 34 {
			relations[i].IssueType = apistructs.IssueTypeBug
		} else if i < 46 {
			relations[i].IssueType = apistructs.IssueTypeTicket
		}
	}
	return is.db.UpdateIssueStateRelations(projectID, apistructs.IssueTypeTask, relations)
}

type IssueStater interface {
	UpdateIssueStateRelations(projectID int64, issueType apistructs.IssueType, StateRelations []dao.IssueStateRelation) error
	CreateIssuesState(state *dao.IssueState) error
	GetIssuesStatesByProjectID(projectID uint64, issueType apistructs.IssueType) ([]dao.IssueState, error)
	GetIssueStateByIDs(ID []int64) ([]dao.IssueState, error)
	GetIssueStateByID(ID int64) (*dao.IssueState, error)
	GetIssuesStates(req *apistructs.IssueStatesGetRequest) ([]dao.IssueState, error)
	GetIssueByState(state int64) (*dao.Issue, error)
	DeleteIssuesStateRelationByStartID(id int64) error
	DeleteIssuesState(id int64) error
	GetIssuesStateRelations(projectID uint64, issueType apistructs.IssueType) ([]dao.IssueStateJoinSQL, error)
	UpdateIssueState(state *dao.IssueState) error
}
