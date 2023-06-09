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
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/excel"
)

func (data DataForFulfill) genStateSheet() (excel.Rows, error) {
	var lines excel.Rows

	// title: state (JSON), state_relation (JSON)
	title := excel.Row{
		excel.NewTitleCell("state (json)"),
		excel.NewTitleCell("state_relation (json)"),
	}
	lines = append(lines, title)

	// data
	stateBytes, err := json.Marshal(data.ExportOnly.States)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state info, err: %v", err)
	}
	stateRelationBytes, err := json.Marshal(data.ExportOnly.StateRelations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state relation info, err: %v", err)
	}
	lines = append(lines, excel.Row{
		excel.NewCell(string(stateBytes)),
		excel.NewCell(string(stateRelationBytes)),
	})

	return lines, nil
}

func (data DataForFulfill) decodeStateSheet(excelSheets [][][]string) ([]dao.IssueState, []dao.IssueStateJoinSQL, error) {
	if data.IsOldExcelFormat() {
		return nil, nil, nil
	}
	sheet := excelSheets[indexOfSheetState]
	// check sheet
	if len(sheet) <= 1 {
		return nil, nil, fmt.Errorf("invalid state sheet, title or data not found")
	}
	var state []dao.IssueState
	var stateRelations []dao.IssueStateJoinSQL
	// only one row
	row := sheet[1]
	if err := json.Unmarshal([]byte(row[0]), &state); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal state info, err: %v", err)
	}
	if err := json.Unmarshal([]byte(row[1]), &stateRelations); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal state relation info, err: %v", err)
	}
	return state, stateRelations, nil
}

func (data *DataForFulfill) syncState(originalProjectStates []dao.IssueState, originalProjectStateRelations []dao.IssueStateJoinSQL) error {
	ctx := apis.WithInternalClientContext(context.Background(), "issue-import")
	// compare original & current project states
	// update data.StateMapByID

	// 分事项类型，每个类型比较状态的差异，只创建，不删除，因为新项目可能有新的状态
	originalStateByTypeAndName := make(map[string]map[string]dao.IssueState)
	for _, state := range originalProjectStates {
		if _, ok := originalStateByTypeAndName[state.IssueType]; !ok {
			originalStateByTypeAndName[state.IssueType] = make(map[string]dao.IssueState)
		}
		originalStateByTypeAndName[state.IssueType][state.Name] = state
	}
	var originalStateNeedCreate []dao.IssueState
	// 遍历原项目状态，找到需要新增的状态
	for issueType, stateMap := range originalStateByTypeAndName {
		for stateName, state := range stateMap {
			if _, ok := data.StateMapByTypeAndName[issueType][stateName]; !ok { // 不存在，需要新增
				originalStateNeedCreate = append(originalStateNeedCreate, state)
			}
		}
	}
	// 新增状态
	for i, stateNeedCreate := range originalStateNeedCreate {
		stateNeedCreate.ID = 0
		stateNeedCreate.ProjectID = data.ProjectID
		stateNeedCreate.Index = 0 // auto when create
		resp, err := data.ImportOnly.IssueCore.CreateIssueState(ctx, &pb.CreateIssueStateRequest{
			ProjectID:   data.ProjectID,
			IssueType:   stateNeedCreate.IssueType,
			StateName:   stateNeedCreate.Name,
			StateBelong: stateNeedCreate.Belong,
		})
		if err != nil {
			return fmt.Errorf("failed to create state: %s, err: %v", stateNeedCreate.Name, err)
		}
		originalStateNeedCreate[i].ID = resp.Data
	}
	// 尝试增加关联关系
	// 过于复杂，不如用户导入后再在界面上调整，所以这里只把新增的状态，移动到对应 belong 状态的最后一个
	// 根据类型遍历当前状态流转
	for issueType := range pb.IssueTypeEnum_Type_value {
		resp, err := data.ImportOnly.IssueCore.GetIssueStateRelation(ctx, &pb.GetIssueStateRelationRequest{
			ProjectID: data.ProjectID,
			IssueType: issueType,
		})
		if err != nil {
			return fmt.Errorf("failed to get state relation, type: %s, err: %v", issueType, err)
		}
		// 将 belong 排序不正确的状态(即新增)，调整正确
		currentRelations := resp.Data
		sortRelationsIntoBelongs(issueType, currentRelations)
		_, err = data.ImportOnly.IssueCore.UpdateIssueStateRelation(ctx, &pb.UpdateIssueStateRelationRequest{
			ProjectID: int64(data.ProjectID),
			Data:      currentRelations,
		})
		if err != nil {
			return fmt.Errorf("failed to update state relation, type: %s, err: %v", issueType, err)
		}
	}
	// 更新 data states 用于 issue 的状态 ID 映射
	stateMapByID, stateMapByTYpeAndName, err := RefreshDataState(data.ProjectID, data.ImportOnly.DB)
	if err != nil {
		return fmt.Errorf("failed to refresh data state, err: %v", err)
	}
	data.StateMap = stateMapByID
	data.StateMapByTypeAndName = stateMapByTYpeAndName

	return nil
}

// sortRelationsIntoBelongs
// 将 belong 按照以下顺序排序
func sortRelationsIntoBelongs(issueType string, relations []*pb.IssueStateRelation) {
	belongOrders := make([]string, 0, len(pb.IssueStateBelongEnum_StateBelong_name))
	for i := 0; i < len(pb.IssueStateBelongEnum_StateBelong_name); i++ {
		belongOrders = append(belongOrders, pb.IssueStateBelongEnum_StateBelong_name[int32(i)])
	}
	// 按照 belong 排序
	sort.Slice(relations, func(i, j int) bool {
		belongI := pb.IssueStateBelongEnum_StateBelong_value[relations[i].StateBelong]
		belongJ := pb.IssueStateBelongEnum_StateBelong_value[relations[j].StateBelong]
		return belongI < belongJ
	})
}

func RefreshDataState(projectID uint64, db *dao.DBClient) (map[int64]string, map[string]map[string]int64, error) {
	// state map
	stateMapByID := make(map[int64]string)
	stateMapByTypeAndName := make(map[string]map[string]int64) // outerkey: issueType, innerkey: stateName
	states, err := db.GetIssuesStatesByProjectID(projectID, "")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get states, err: %v", err)
	}
	for _, v := range states {
		stateMapByID[int64(v.ID)] = v.Name
		if _, ok := stateMapByTypeAndName[v.IssueType]; !ok {
			stateMapByTypeAndName[v.IssueType] = make(map[string]int64)
		}
		stateMapByTypeAndName[v.IssueType][v.Name] = int64(v.ID)
	}
	return stateMapByID, stateMapByTypeAndName, nil
}
