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

package wxj

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
)

const (
	InitializeOperation      cptype.OperationKey = "__Initialize__"
	RenderingOperation       cptype.OperationKey = "__Rendering__"
	AddParallelOperationKey  cptype.OperationKey = "addParallel"
	CopyParallelOperationKey cptype.OperationKey = "copyParallel"
	CopyToOperationKey       cptype.OperationKey = "copyTo"
	MoveItemOperationKey     cptype.OperationKey = "moveItem"
	MoveGroupOperationKey    cptype.OperationKey = "moveGroup"
	EditOperationKey         cptype.OperationKey = "edit"
	DeleteOperationKey       cptype.OperationKey = "delete"
	SplitOperationKey        cptype.OperationKey = "split"
	SwitchOperationKey       cptype.OperationKey = "switch"
	ClickOperationKey        cptype.OperationKey = "clickItem"
)

type OperationFunc []func(s *SceneStage) error

var OperationRender = map[cptype.OperationKey]OperationFunc{
	InitializeOperation:      []func(s *SceneStage) error{RenderList},
	RenderingOperation:       []func(s *SceneStage) error{RenderList},
	AddParallelOperationKey:  []func(s *SceneStage) error{RenderAddParallel, RenderList},
	CopyParallelOperationKey: []func(s *SceneStage) error{RenderCopyParallel, RenderList},
	CopyToOperationKey:       []func(s *SceneStage) error{RenderCopyTo, RenderList},
	MoveItemOperationKey:     []func(s *SceneStage) error{RenderItemMove, RenderList},
	MoveGroupOperationKey:    []func(s *SceneStage) error{RenderGroupMove, RenderList},
	EditOperationKey:         []func(s *SceneStage) error{RenderEdit, RenderList},
	DeleteOperationKey:       []func(s *SceneStage) error{RenderDelete, RenderList},
	SplitOperationKey:        []func(s *SceneStage) error{RenderSplit, RenderList},
	SwitchOperationKey:       []func(s *SceneStage) error{RenderSwitch, RenderList},
}

func RenderList(s *SceneStage) error {
	_, scenes, err := s.atTestPlan.ListAutotestScene(apistructs.AutotestSceneRequest{
		SceneID: s.State.SetID,
	})
	if err != nil {
		return err
	}

	s.Data.Type = "sort-item"
	s.Data.List = func() []StageData {
		stages := make([]StageData, 0, len(scenes))
		for _, v := range scenes {
			stages = append(stages, NewStageData(v))
		}
		return stages
	}()

	s.Operations[MoveItemOperationKey.String()] = OperationInfo{
		OperationBaseInfo: OperationBaseInfo{Key: MoveItemOperationKey.String(), Reload: true},
	}
	s.Operations[MoveGroupOperationKey.String()] = OperationInfo{
		OperationBaseInfo: OperationBaseInfo{Key: MoveGroupOperationKey.String(), Reload: true},
	}
	s.Operations[ClickOperationKey.String()] = OperationInfo{
		OperationBaseInfo: OperationBaseInfo{Key: ClickOperationKey.String(), Reload: true},
	}
	return nil
}

func RenderAddParallel(s *SceneStage) error {
	meta, err := GetOpsInfo(s.event.OperationData)
	if err != nil {
		return err
	}
	preScene, err := s.atTestPlan.GetAutotestScene(apistructs.AutotestSceneRequest{SceneID: meta.ID})
	if err != nil {
		return err
	}
	_, err = s.atTestPlan.CreateAutotestScene(apistructs.AutotestSceneRequest{
		Name:        "",
		Description: "",
		SetID:       0,
		SceneGroupID: func() uint64 {
			if preScene.GroupID == 0 {
				return preScene.ID
			}
			return preScene.GroupID
		}(),
	})
	return err
}

func RenderCopyParallel(s *SceneStage) error {
	meta, err := GetOpsInfo(s.event.OperationData)
	if err != nil {
		return err
	}
	_, err = s.atTestPlan.CopyAutotestScene(apistructs.AutotestSceneCopyRequest{
		PreID:        meta.ID,
		SceneID:      meta.ID,
		SetID:        s.State.SetID,
		IdentityInfo: apistructs.IdentityInfo{UserID: s.sdk.Identity.UserID},
	}, false, nil)
	return err
}

func RenderCopyTo(s *SceneStage) error {

	return nil
}

func RenderItemMove(s *SceneStage) error {
	dragGroupKey := uint64(s.State.DragParams.DragGroupKey)
	dropGroupKey := uint64(s.State.DragParams.DropGroupKey)

	req := apistructs.AutotestSceneMoveRequest{
		IdentityInfo: apistructs.IdentityInfo{UserID: s.sdk.Identity.UserID},
		FirstID:      uint64(s.State.DragParams.DragKey),
		LastID:       uint64(s.State.DragParams.DragKey),
		TargetID: func() uint64 {
			if s.State.DragParams.DropKey == -1 { // Move to the end and be independent group
				return 0
			}
			return uint64(s.State.DragParams.DropKey)
		}(),
		IsGroup: false,
	}

	// Find preID
	if s.State.DragParams.DropKey == -1 { // Move to the end and be independent group
		_, scenes, err := s.atTestPlan.ListAutotestScene(apistructs.AutotestSceneRequest{SetID: s.State.SetID})
		if err != nil {
			return err
		}
		req.PreID = scenes[len(scenes)-1].ID
	} else {
		switch s.State.DragParams.Position {
		case 0: // Inside target
			return nil
		case 1: // Behind the target
			scene, err := s.atTestPlan.GetAutotestScene(apistructs.AutotestSceneRequest{SceneID: uint64(s.State.DragParams.DropKey)})
			if err != nil {
				return err
			}
			req.PreID = uint64(s.State.DragParams.DropKey)
			// The order of the linked list has not changed in the same group
			if req.PreID == scene.PreID && dragGroupKey == dropGroupKey {
				return nil
			}

		case -1: // In front of the target
			scene, err := s.atTestPlan.GetAutotestScene(apistructs.AutotestSceneRequest{SceneID: uint64(s.State.DragParams.DropKey)})
			if err != nil {
				return err
			}
			req.PreID = scene.PreID
			// The order of the linked list has not changed in the same group
			if req.PreID == req.LastID && dragGroupKey == dropGroupKey {
				return nil
			}
		default:
			return errors.New("unknown position")
		}
	}
	return s.atTestPlan.MoveAutotestSceneV2(req)
}

func RenderGroupMove(s *SceneStage) error {
	dragGroupKey := uint64(s.State.DragParams.DragGroupKey)
	dropGroupKey := uint64(s.State.DragParams.DropGroupKey)

	req := apistructs.AutotestSceneMoveRequest{
		IdentityInfo: apistructs.IdentityInfo{UserID: s.sdk.Identity.UserID},
		IsGroup:      true,
	}

	_, sceneDragGroup, err := s.atTestPlan.ListAutotestScene(apistructs.AutotestSceneRequest{SetID: s.State.SetID, SceneGroupID: dragGroupKey})
	if err != nil {
		return err
	}
	if len(sceneDragGroup) <= 0 {
		return errors.New("the dragGroupKey is not exists")
	}
	firstSceneDrag, lastSceneDrag := findFirstLastSceneInGroup(sceneDragGroup)
	req.FirstID = firstSceneDrag.ID
	req.LastID = lastSceneDrag.ID

	switch s.State.DragParams.Position {
	case 0: // Inside target
		return nil
	case -1: // In front of the target
		_, sceneDropGroup, err := s.atTestPlan.ListAutotestScene(apistructs.AutotestSceneRequest{SetID: s.State.SetID, SceneGroupID: dropGroupKey})
		if err != nil {
			return err
		}
		if len(sceneDropGroup) <= 0 {
			return errors.New("the dropGroupKey is not exists")
		}
		firstSceneDrop, _ := findFirstLastSceneInGroup(sceneDropGroup)
		req.PreID = firstSceneDrop.PreID
		req.TargetID = firstSceneDrop.ID
		// The order of the linked list has not changed
		if req.PreID == req.LastID {
			return nil
		}
	case 1: // Behind the target
		_, sceneDropGroup, err := s.atTestPlan.ListAutotestScene(apistructs.AutotestSceneRequest{SetID: s.State.SetID, SceneGroupID: dropGroupKey})
		if err != nil {
			return err
		}
		if len(sceneDropGroup) <= 0 {
			return errors.New("the dropGroupKey is not exists")
		}
		_, lastSceneDrop := findFirstLastSceneInGroup(sceneDropGroup)
		req.PreID = lastSceneDrop.ID
		req.TargetID = lastSceneDrop.ID
		// The order of the linked list has not changed
		if req.PreID == firstSceneDrag.PreID {
			return nil
		}
	default:
		return errors.New("unknown position")
	}
	return s.atTestPlan.MoveAutotestSceneV2(req)
}

func RenderEdit(s *SceneStage) error {
	meta, err := GetOpsInfo(s.event.OperationData)
	if err != nil {
		return err
	}
	_, err = s.atTestPlan.UpdateAutotestScene(apistructs.AutotestSceneSceneUpdateRequest{
		SceneID:     meta.ID,
		Name:        "",
		Description: "",
	})
	return err
}

func RenderDelete(s *SceneStage) error {
	meta, err := GetOpsInfo(s.event.OperationData)
	if err != nil {
		return err
	}
	return s.atTestPlan.DeleteAutotestScene(meta.ID, apistructs.IdentityInfo{UserID: s.sdk.Identity.UserID})
}

func RenderSplit(s *SceneStage) error {
	meta, err := GetOpsInfo(s.event.OperationData)
	if err != nil {
		return err
	}

	req := apistructs.AutotestSceneMoveRequest{
		IdentityInfo: apistructs.IdentityInfo{UserID: s.sdk.Identity.UserID},
		FirstID:      meta.ID,
		LastID:       meta.ID,
		TargetID:     0,
		IsGroup:      false,
	}
	_, sceneGroup, err := s.atTestPlan.ListAutotestScene(apistructs.AutotestSceneRequest{SetID: s.State.SetID, SceneGroupID: uint64(meta.Data["groupID"].(float64))})
	if err != nil {
		return err
	}
	if len(sceneGroup) <= 0 {
		return errors.New("the group is not exists")
	}

	if len(sceneGroup) == 1 {
		return nil
	}
	return s.atTestPlan.MoveAutotestSceneV2(req)
}

func RenderSwitch(s *SceneStage) error {
	return nil
}

func findFirstLastSceneInGroup(scenes []apistructs.AutoTestScene) (firstScene, lastScene apistructs.AutoTestScene) {
	sceneIDMap := make(map[uint64]apistructs.AutoTestScene, len(scenes))
	preIDMap := make(map[uint64]apistructs.AutoTestScene, len(scenes))
	for _, v := range scenes {
		sceneIDMap[v.ID] = v
		preIDMap[v.PreID] = v
	}
	for k := range preIDMap {
		if _, ok := sceneIDMap[k]; !ok {
			firstScene = preIDMap[k]
			break
		}
	}
	for k := range sceneIDMap {
		if _, ok := preIDMap[k]; !ok {
			lastScene = sceneIDMap[k]
			break
		}
	}
	return
}
