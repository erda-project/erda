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
	"strings"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/strutil"
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

func (data DataForFulfill) decodeStateSheet(df excel.DecodedFile) error {
	if data.IsOldExcelFormat() {
		return nil
	}
	s, ok := df.Sheets.M[nameOfSheetState]
	if !ok {
		return nil
	}
	sheet := s.UnmergedSlice
	// check sheet
	if len(sheet) <= 1 {
		return fmt.Errorf("invalid state sheet, title or data not found")
	}
	var state []dao.IssueState
	var stateRelations []dao.IssueStateJoinSQL
	// only one row
	row := sheet[1]
	if err := json.Unmarshal([]byte(row[0]), &state); err != nil {
		return fmt.Errorf("failed to unmarshal state info, err: %v", err)
	}
	if err := json.Unmarshal([]byte(row[1]), &stateRelations); err != nil {
		return fmt.Errorf("failed to unmarshal state relation info, err: %v", err)
	}
	data.ImportOnly.Sheets.Optional.StateInfo = &StateInfo{
		States:        state,
		StateJoinSQLs: stateRelations,
	}
	return nil
}

func (data *DataForFulfill) syncState(originalProjectStatesInfo *StateInfo) error {
	if originalProjectStatesInfo == nil {
		return nil
	}
	ctx := apis.WithUserIDContext(context.Background(), apistructs.SystemUserID)
	// compare original & current project states
	// update data.StateMapByID

	// 分事项类型，每个类型比较状态的差异，只创建，不删除，因为新项目可能有新的状态
	originalStateByTypeAndName := make(map[string]map[string]dao.IssueState)
	for _, state := range originalProjectStatesInfo.States {
		if _, ok := originalStateByTypeAndName[state.IssueType]; !ok {
			originalStateByTypeAndName[state.IssueType] = make(map[string]dao.IssueState)
		}
		originalStateByTypeAndName[state.IssueType][state.Name] = state
	}
	var originalStateNeedCreate []dao.IssueState
	originalStateNeedCreateMap := make(map[string]dao.IssueState) // key: issueType + stateName
	// 遍历原项目状态，找到需要新增的状态
	for issueType, stateMap := range originalStateByTypeAndName {
		for stateName, state := range stateMap {
			if _, ok := data.StateMapByTypeAndName[issueType][stateName]; !ok { // 不存在，需要新增
				originalStateNeedCreate = append(originalStateNeedCreate, state)
				originalStateNeedCreateMap[issueType+stateName] = state
			}
		}
	}
	// 遍历 issue 里的状态，找到只在 issue sheet 里声明的新状态，包括新老格式
	for _, issueSheetModel := range data.ImportOnly.Sheets.Must.IssueInfo {
		// 如果 state 已经在当前项目存在，跳过
		if _, ok := data.StateMapByTypeAndName[issueSheetModel.Common.IssueType.String()][issueSheetModel.Common.State]; ok {
			continue
		}
		// 如果 state 已经在 originalStateNeedCreateMap 中存在，跳过
		if _, ok := originalStateNeedCreateMap[issueSheetModel.Common.IssueType.String()+issueSheetModel.Common.State]; ok {
			continue
		}
		// 确认是一个新的状态，则需要新增
		newState := dao.IssueState{
			IssueType: issueSheetModel.Common.IssueType.String(),
			Name:      issueSheetModel.Common.State,
			Belong:    tryToGuessNewStateBelong(issueSheetModel.Common.State, issueSheetModel).String(),
		}
		originalStateNeedCreate = append(originalStateNeedCreate, newState)
		originalStateNeedCreateMap[issueSheetModel.Common.IssueType.String()+issueSheetModel.Common.State] = newState
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
	stateMapByID, stateMapByTypeAndName, err := RefreshDataState(data.ProjectID, data.ImportOnly.DB)
	if err != nil {
		return fmt.Errorf("failed to refresh data state, err: %v", err)
	}
	data.StateMap = stateMapByID
	data.StateMapByTypeAndName = stateMapByTypeAndName

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
	sort.SliceStable(relations, func(i, j int) bool {
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

// tryToGuessNewStateBelong
// 规则：
// 1. 如果 model 的完成时间不为空，则属于已完成，否则继续判断
// 2. 如果包含这些关键字，则属于已完成：已
// 3. 如果包含这些关键字，则属于未开始：未、待
// 4. 如果包含这些关键字，则属于进行中：中、正在
// 5. 其他情况，根据关键字进行匹配
// - 已完成：完成、关闭
// - 未开始：新建
// - 进行中：
// 6. 默认属于处理中
func tryToGuessNewStateBelong(name string, model IssueSheetModel) pb.IssueStateBelongEnum_StateBelong {
	if model.Common.FinishAt != nil && !model.Common.FinishAt.IsZero() {
		return getDoneStateBelongByIssueType(model.Common.IssueType)
	}
	if strutil.HasPrefixes(name, "已") {
		return getDoneStateBelongByIssueType(model.Common.IssueType)
	}
	if strutil.HasPrefixes(name, "未", "待") {
		return pb.IssueStateBelongEnum_OPEN
	}
	if strutil.HasPrefixes(name, "正在") || strutil.HasSuffixes(name, "中") {
		return pb.IssueStateBelongEnum_WORKING
	}
	if strutil.HasPrefixes(name, "完成", "关闭") {
		return getDoneStateBelongByIssueType(model.Common.IssueType)
	}
	if strings.Contains(name, "新建") {
		return pb.IssueStateBelongEnum_OPEN
	}
	return pb.IssueStateBelongEnum_WORKING
}

func getDoneStateBelongByIssueType(issueType pb.IssueTypeEnum_Type) pb.IssueStateBelongEnum_StateBelong {
	switch issueType {
	case pb.IssueTypeEnum_BUG:
		return pb.IssueStateBelongEnum_CLOSED
	default:
		return pb.IssueStateBelongEnum_DONE
	}
}
