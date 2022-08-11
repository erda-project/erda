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
	_ "embed"
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
)

var (
	//go:embed state-zh_CN.json
	zhStateConfig   []byte
	zhStateInitData []apistructs.StateDefinitionCustomizeData
	//go:embed state-en_US.json
	enStateConfig   []byte
	enStateInitData []apistructs.StateDefinitionCustomizeData

	stateInitRole = "Ops,Dev,QA,Owner,Lead"
)

func init() {
	if err := json.Unmarshal(zhStateConfig, &zhStateInitData); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(enStateConfig, &enStateInitData); err != nil {
		panic(err)
	}
}

// IssueState issue state service 对象
type IssueState struct {
	bdl *bundle.Bundle
	db  *dao.DBClient
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
func WithDBClient(db *dao.DBClient) Option {
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

func (is *IssueState) GetIssuesStatesNameByID(id []int64) ([]apistructs.IssueStatus, error) {
	state, err := is.db.GetIssueStateByIDs(id)
	if err != nil {
		return nil, err
	}
	var status []apistructs.IssueStatus
	for _, v := range state {
		status = append(status, apistructs.IssueStatus{
			ProjectID:   v.ProjectID,
			IssueType:   apistructs.IssueType(v.IssueType),
			StateID:     int64(v.ID),
			StateName:   v.Name,
			StateBelong: apistructs.IssueStateBelong(v.Belong),
			Index:       v.Index,
		})
	}

	return status, nil
}

func (is *IssueState) getStateInitData(locale string) []apistructs.StateDefinitionCustomizeData {
	switch locale {
	case "zh-CN":
		return zhStateInitData
	case "en-US":
		return enStateInitData
	default:
		return zhStateInitData
	}
}

func (is *IssueState) InitProjectState(projectID int64, locale string) error {
	initData := is.getStateInitData(locale)
	for _, stateData := range initData {
		var (
			states    []dao.IssueState
			relations []dao.IssueStateRelation
		)
		for _, initState := range stateData.States {
			state := dao.IssueState{
				ProjectID: uint64(projectID),
				IssueType: string(stateData.IssueType),
				Name:      initState.Name,
				Belong:    string(initState.IssueStateBelong),
				Index:     initState.Index,
				Role:      stateInitRole,
			}
			if err := is.db.CreateIssuesState(&state); err != nil {
				return err
			}
			states = append(states, state)
		}
		for _, customRelation := range stateData.Relations {
			relation := dao.IssueStateRelation{
				ProjectID:    projectID,
				IssueType:    string(stateData.IssueType),
				StartStateID: int64(states[customRelation.From].ID),
				EndStateID:   int64(states[customRelation.To].ID),
			}
			relations = append(relations, relation)
		}
		if err := is.db.UpdateIssueStateRelations(projectID, string(stateData.IssueType), relations); err != nil {
			return err
		}
	}
	return nil
}
