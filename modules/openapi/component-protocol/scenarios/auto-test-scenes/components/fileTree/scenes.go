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

package fileTree

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
)

func (i *ComponentFileTree) RenderFileTree(inParams InParams) error {
	req := apistructs.SceneSetRequest{
		SpaceID: inParams.SpaceId,
	}
	req.UserID = i.CtxBdl.Identity.UserID
	rsp, err := i.CtxBdl.Bdl.GetSceneSets(req)
	if err != nil {
		return err
	}

	isSelectScene := inParams.SceneID != ""
	id, _ := strconv.Atoi(inParams.SceneID)
	sceneId := uint64(id)
	id, _ = strconv.Atoi(inParams.SetID)
	setId := uint64(id)

	var res []SceneSet
	for _, s := range rsp {
		set := initSceneSet(&s)
		if s.ID == setId {
			if isSelectScene {
				children, find, err := i.findScene(s.ID, sceneId)
				if err != nil {
					return err
				}
				if find {
					set.Children = children
					i.State.SceneId__urlQuery = inParams.SceneID
					i.State.SetId__urlQuery = strconv.Itoa(int(s.ID))
					i.State.SceneSetKey = int(s.ID)
					i.State.ExpandedKeys = append(i.State.ExpandedKeys, "sceneset-"+strconv.Itoa(int(s.ID)))
					i.State.SelectedKeys = append(i.State.SelectedKeys, "scene-"+inParams.SceneID)
					i.State.SceneId = sceneId
				}
			} else {
				i.State.SelectedKeys = append(i.State.SelectedKeys, "sceneset-"+inParams.SetID)
				i.State.SceneSetKey = int(setId)
				i.State.PageNo = 1
				i.State.SceneId = 0
			}
		}
		res = append(res, set)
	}

	i.Data = res
	i.State.FormVisible = false
	return nil
}

func (i *ComponentFileTree) findScene(setId uint64, sceneId uint64) ([]Scene, bool, error) {
	req := apistructs.AutotestSceneRequest{
		SetID: setId,
	}
	req.UserID = i.CtxBdl.Identity.UserID
	find := false
	_, rsp, err := i.CtxBdl.Bdl.ListAutoTestScene(req)
	if err != nil {
		return nil, find, err
	}

	var res []Scene
	for _, s := range rsp {
		if sceneId == s.ID {
			find = true
		}
		scene := initScene(s, int(setId))
		res = append(res, scene)
	}

	return res, find, nil
}

func (i *ComponentFileTree) RenderSceneSets(inParams InParams) error {
	req := apistructs.SceneSetRequest{
		SpaceID: inParams.SpaceId,
	}
	req.UserID = i.CtxBdl.Identity.UserID
	rsp, err := i.CtxBdl.Bdl.GetSceneSets(req)
	if err != nil {
		return err
	}

	var res []SceneSet
	for _, s := range rsp {
		set := initSceneSet(&s)
		setId := "sceneset-" + strconv.FormatUint(s.ID, 10)
		if exist(setId, i.State.ExpandedKeys) || exist(setId, i.State.SelectedKeys) || i.State.SceneSetKey == int(s.ID) {
			//idx == 0 ||
			children, err := i.getScenes(s.ID)
			if err != nil {
				return err
			}
			set.Children = children
		}
		res = append(res, set)
	}

	// selectKey, expandedKeys := findSelectedKeysExpandedKeys(res, inParams.SelectedKeys)
	if i.State.SceneSetKey != 0 {
		i.State.ExpandedKeys = append(i.State.ExpandedKeys, "sceneset-"+strconv.Itoa(i.State.SceneSetKey))
	}

	i.Data = res
	// if len(i.State.SelectedKeys) == 0 && len(res) > 0 {
	// 	id := res[0].Key[9:]
	// 	setId, _ := strconv.Atoi(id)
	// 	i.State.SceneSetKey = setId
	// 	i.State.ExpandedKeys = append(i.State.ExpandedKeys, "sceneset-"+id)
	// 	if len(res[0].Children) > 0 {
	// 		id := res[0].Children[0].Key[6:]
	// 		sceneId, _ := strconv.Atoi(id)
	// 		i.State.SceneId__urlQuery = uint64(sceneId)
	// 		i.State.SceneId = uint64(sceneId)
	// 		i.State.SelectedKeys = []string{"scene-" + id}
	// 	} else {
	// 		i.State.PageNo = 1
	// 		i.State.SceneId = 0
	// 		i.State.SelectedKeys = []string{"sceneset-" + id}
	// 	}
	// }
	i.State.FormVisible = false
	return nil
}

func exist(target string, arr []string) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}

func initSceneSet(s *apistructs.SceneSet) SceneSet {
	id := int(s.ID)
	set := SceneSet{
		Key:           "sceneset-" + strconv.FormatUint(s.ID, 10),
		Title:         s.Name,
		Icon:          "folder",
		IsColorIcon:   true,
		IsLeaf:        false,
		ClickToExpand: true,
		Selectable:    true,
		Type:          "sceneset",
	}

	var expand = SceneSetOperation{
		Key: "ExpandSceneSet",
		// Text:   "展开",
		Reload: true,
		Show:   false,
		Meta: SceneSetOperationMeta{
			ParentKey: id,
		},
	}

	var click = SceneSetOperation{
		Key: "ClickSceneSet",
		// Text:   "展开",
		Reload: true,
		Show:   false,
		Meta: SceneSetOperationMeta{
			ParentKey: id,
		},
	}

	var addScene = SceneSetOperation{
		Key:    "AddScene",
		Text:   "添加场景",
		Reload: true,
		Show:   true,
		Meta: SceneSetOperationMeta{
			ParentKey: id,
		},
	}

	var refSceneSet = SceneSetOperation{
		Key:    "RefSceneSet",
		Text:   "引用场景集",
		Reload: true,
		Show:   true,
		Meta: SceneSetOperationMeta{
			ParentKey: id,
		},
	}

	var edit = SceneSetOperation{
		Key:    "UpdateSceneSet",
		Text:   "编辑场景集",
		Reload: true,
		Show:   true,
		Meta: SceneSetOperationMeta{
			ParentKey: id,
		},
	}

	var delete = DeleteOperation{
		Key:     "DeleteSceneSet",
		Text:    "删除",
		Confirm: "是否确认删除",
		Reload:  true,
		Show:    true,
		Meta: SceneSetOperationMeta{
			ParentKey: id,
		},
	}

	set.Operations = map[string]interface{}{}
	set.Operations["expand"] = expand
	set.Operations["click"] = click
	set.Operations["addScene"] = addScene
	set.Operations["editScene"] = edit
	set.Operations["delete"] = delete
	set.Operations["refSceneSet"] = refSceneSet
	return set
}

func initScene(scene apistructs.AutoTestScene, setId int) Scene {
	var s Scene
	sceneId := strconv.FormatUint(scene.ID, 10)
	s.Title = "#" + sceneId + " " + scene.Name
	if scene.RefSetID == 0 {
		s.Icon = "dm"
	} else {
		s.Icon = "folder"
	}
	s.IsLeaf = true
	id := int(scene.ID)
	s.Key = "scene-" + sceneId
	s.Type = "scene"

	var click = SceneSetOperation{
		Key:    "ClickScene",
		Reload: true,
		Show:   false,
		Meta: SceneSetOperationMeta{
			ParentKey: setId,
			Key:       id,
		},
	}

	var edit = SceneSetOperation{
		Key:    "UpdateScene",
		Text:   "编辑场景",
		Reload: true,
		Show:   true,
		Meta: SceneSetOperationMeta{
			ParentKey: setId,
			Key:       id,
		},
	}

	var copy = SceneSetOperation{
		Key:    "CopyScene",
		Text:   "复制场景",
		Reload: true,
		Show:   true,
		Meta: SceneSetOperationMeta{
			ParentKey: setId,
			Key:       id,
		},
	}

	s.Operations = map[string]interface{}{}
	s.Operations["click"] = click
	if scene.RefSetID <= 0 {
		s.Operations["editScene"] = edit
	}
	s.Operations["copyScene"] = copy

	var deleteOperation = DeleteOperation{
		Key:     "DeleteScene",
		Text:    "删除",
		Confirm: "是否确认删除",
		Reload:  true,
		Show:    true,
		Meta: SceneSetOperationMeta{
			ParentKey: setId,
			Key:       id,
		},
	}
	s.Operations["delete"] = deleteOperation
	return s
}

func getOperation(operationData *SceneSetOperation, event apistructs.ComponentEvent) error {
	if event.OperationData == nil {
		return nil
	}
	b, err := json.Marshal(event.OperationData)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &operationData); err != nil {
		return err
	}
	return nil
}

func (i *ComponentFileTree) RenderDeleteSceneSet(event apistructs.ComponentEvent, inParams InParams) error {
	var operationData SceneSetOperation
	if err := getOperation(&operationData, event); err != nil {
		return err
	}
	setId := operationData.Meta.ParentKey
	req := apistructs.AutotestSceneRequest{
		SetID: uint64(setId),
	}
	req.UserID = i.CtxBdl.Identity.UserID

	// _, rsp, err := i.CtxBdl.Bdl.ListAutoTestScene(req)
	// if rsp != nil {
	// 	return fmt.Errorf("Cannot delete sceneset, it is not empty")
	// }

	request := apistructs.SceneSetRequest{
		SetID:     uint64(setId),
		ProjectId: inParams.ProjectId,
	}
	request.UserID = i.CtxBdl.Identity.UserID

	if err := i.CtxBdl.Bdl.DeleteSceneSet(request); err != nil {
		return err
	}
	i.resetKeys()
	return nil
}

func (i *ComponentFileTree) resetKeys() {
	i.State.SceneSetKey = 0
	l := len(i.State.SelectedKeys)
	if l > 0 {
		i.State.SelectedKeys = i.State.SelectedKeys[:l-1]
	}
}

func (i *ComponentFileTree) RenderDeleteScene(event apistructs.ComponentEvent) error {
	var operationData SceneSetOperation
	if err := getOperation(&operationData, event); err != nil {
		return err
	}

	id := operationData.Meta.Key
	request := apistructs.AutotestSceneRequest{
		SceneID: uint64(id),
	}
	request.UserID = i.CtxBdl.Identity.UserID

	if err := i.CtxBdl.Bdl.DeleteAutoTestScene(request); err != nil {
		return err
	}

	i.State.SceneSetKey = operationData.Meta.ParentKey
	i.State.SceneId = 0
	return nil
}

func (i *ComponentFileTree) RenderExpandSceneSet(event apistructs.ComponentEvent) error {
	var operationData SceneSetOperation
	if err := getOperation(&operationData, event); err != nil {
		return err
	}

	setId := strconv.Itoa(operationData.Meta.ParentKey)
	key := "sceneset-" + setId
	for idx, set := range i.Data {
		if set.Key == key {
			id, _ := strconv.Atoi(setId)
			scenes, err := i.getScenes(uint64(id))
			if err != nil {
				return fmt.Errorf("failed to list scenes, err: %v", err)
			}
			i.Data[idx].Children = scenes
			i.State.ExpandedKeys = append(i.State.ExpandedKeys, set.Key)
		}
	}
	i.State.SelectedKeys = []string{key}
	i.State.PageNo = 1
	i.State.SceneSetKey = operationData.Meta.ParentKey
	i.State.SceneId = 0
	i.State.SceneId__urlQuery = ""
	i.State.SetId__urlQuery = setId
	return nil
}

func (i *ComponentFileTree) RenderClickScene(event apistructs.ComponentEvent) error {
	var operationData SceneSetOperation
	if err := getOperation(&operationData, event); err != nil {
		return err
	}

	id := operationData.Meta.Key

	var req apistructs.AutotestSceneRequest
	req.SceneID = uint64(id)
	req.UserID = i.CtxBdl.Identity.UserID
	scene, err := i.CtxBdl.Bdl.GetAutoTestScene(req)
	if err != nil {
		return err
	}
	if scene.RefSetID != 0 {
		i.SetSceneSetClick(int(scene.RefSetID))
	} else {
		i.State.SceneId__urlQuery = strconv.Itoa(id)
		i.State.SetId__urlQuery = strconv.Itoa(operationData.Meta.ParentKey)
		i.State.SceneSetKey = operationData.Meta.ParentKey
		i.State.SceneId = uint64(id)
	}
	return nil
}

func (i *ComponentFileTree) RenderClickSceneSet(event apistructs.ComponentEvent) error {
	var operationData SceneSetOperation
	if err := getOperation(&operationData, event); err != nil {
		return err
	}
	id := operationData.Meta.ParentKey

	i.State.PageNo = 1
	i.State.SceneSetKey = id
	i.State.SceneId = 0
	i.State.SetId__urlQuery = strconv.Itoa(id)
	i.State.SceneId__urlQuery = ""
	return nil
}

func (i *ComponentFileTree) RenderDragHelper(inParams InParams) error {
	l1, l2 := len(i.State.DragParams.DragType), len(i.State.DragParams.DropType)
	dragType, dropType := i.State.DragParams.DragType, i.State.DragParams.DropType
	from, _ := strconv.Atoi(i.State.DragParams.DragKey[l1+1:])
	to, _ := strconv.Atoi(i.State.DragParams.DropKey[l2+1:])
	if dragType == "sceneset" {
		if dropType == "sceneset" {
			if err := i.DragSceneSet(from, to, inParams.ProjectId); err != nil {
				return err
			}
		}
	} else {
		if dropType == "scene" {
			if err := i.DragScene(from, to); err != nil {
				return err
			}
		} else {
			if err := i.DragSceneToSceneSet(from, to); err != nil {
				return err
			}
		}
	}
	return nil
}

func (i *ComponentFileTree) DragSceneSet(from, to int, projectId uint64) error {
	req := apistructs.SceneSetRequest{
		SetID:     uint64(from),
		DropKey:   int64(to),
		Position:  i.State.DragParams.Position,
		ProjectId: projectId,
	}
	req.UserID = i.CtxBdl.Identity.UserID
	if err := i.CtxBdl.Bdl.DragSceneSet(req); err != nil {
		return err
	}
	return nil
}

func (i *ComponentFileTree) DragScene(from, to int) error {
	id := uint64(from)
	req := apistructs.AutotestSceneRequest{
		AutoTestSceneParams: apistructs.AutoTestSceneParams{
			ID: id,
		},
		Target:   int64(to),
		Position: i.State.DragParams.Position,
	}
	req.UserID = i.CtxBdl.Identity.UserID
	if _, err := i.CtxBdl.Bdl.MoveAutoTestScene(req); err != nil {
		return err
	}

	i.State.SelectedKeys = []string{"scene-" + strconv.Itoa(from)}
	i.State.SceneId__urlQuery = strconv.FormatUint(id, 10)
	i.State.SceneId = id
	return nil
}

func (i *ComponentFileTree) DragSceneToSceneSet(from, to int) error {
	id := uint64(from)
	req := apistructs.AutotestSceneRequest{
		AutoTestSceneParams: apistructs.AutoTestSceneParams{
			ID: id,
		},
		GroupID:  int64(to),
		Position: i.State.DragParams.Position,
	}
	req.UserID = i.CtxBdl.Identity.UserID
	if _, err := i.CtxBdl.Bdl.MoveAutoTestScene(req); err != nil {
		return err
	}
	i.State.SelectedKeys = []string{"scene-" + strconv.Itoa(from)}
	i.State.SceneId__urlQuery = strconv.FormatUint(id, 10)
	i.State.SceneId = id
	return nil
}

func (i *ComponentFileTree) getScenes(setId uint64) ([]Scene, error) {
	req := apistructs.AutotestSceneRequest{
		SetID: setId,
	}
	req.UserID = i.CtxBdl.Identity.UserID
	_, rsp, err := i.CtxBdl.Bdl.ListAutoTestScene(req)
	if err != nil {
		return nil, err
	}

	var res []Scene
	for _, s := range rsp {
		scene := initScene(s, int(setId))
		res = append(res, scene)
	}

	return res, nil
}

func (i *ComponentFileTree) RenderUpdateSceneSet(event apistructs.ComponentEvent) error {
	var operationData SceneSetOperation
	if err := getOperation(&operationData, event); err != nil {
		return err
	}

	i.State.FormVisible = true
	i.State.ActionType = "UpdateSceneSet"
	i.State.SceneSetKey = operationData.Meta.ParentKey
	i.State.SceneId = 0
	i.State.SelectedKeys = []string{"sceneset-" + strconv.Itoa(i.State.SceneSetKey)}
	return nil
}

func (i *ComponentFileTree) RenderRefSceneSet(event apistructs.ComponentEvent) error {
	var operationData SceneSetOperation
	if err := getOperation(&operationData, event); err != nil {
		return err
	}

	i.State.FormVisible = true
	i.State.ActionType = "addRefSceneSet"
	i.State.SceneSetKey = operationData.Meta.ParentKey
	i.State.SceneId = 0
	setId := "sceneset-" + strconv.Itoa(operationData.Meta.ParentKey)
	i.State.ExpandedKeys = append(i.State.ExpandedKeys, setId)
	i.State.SelectedKeys = []string{setId}
	return nil
}

func (i *ComponentFileTree) RenderAddScene(event apistructs.ComponentEvent) error {
	var operationData SceneSetOperation
	if err := getOperation(&operationData, event); err != nil {
		return err
	}

	i.State.FormVisible = true
	i.State.ActionType = "AddScene"
	i.State.SceneSetKey = operationData.Meta.ParentKey
	i.State.SceneId = 0
	setId := "sceneset-" + strconv.Itoa(operationData.Meta.ParentKey)
	i.State.ExpandedKeys = append(i.State.ExpandedKeys, setId)
	i.State.SelectedKeys = []string{setId}
	return nil
}

func (i *ComponentFileTree) RenderUpdateScene(event apistructs.ComponentEvent) error {
	var operationData SceneSetOperation
	if err := getOperation(&operationData, event); err != nil {
		return err
	}

	i.State.SelectedKeys = []string{"scene-" + strconv.Itoa(operationData.Meta.Key)}
	i.State.FormVisible = true
	i.State.ActionType = "UpdateScene"
	i.State.SceneSetKey = operationData.Meta.ParentKey
	i.State.SceneId = uint64(operationData.Meta.Key)
	return nil
}

func (i *ComponentFileTree) RenderCopyScene(inParams InParams, event apistructs.ComponentEvent) error {
	var operationData SceneSetOperation
	if err := getOperation(&operationData, event); err != nil {
		return err
	}

	sceneId := uint64(operationData.Meta.Key)
	setId := operationData.Meta.ParentKey
	req := apistructs.AutotestSceneCopyRequest{
		SpaceID: inParams.SpaceId,
		PreID:   sceneId,
		SceneID: sceneId,
		SetID:   uint64(setId),
	}
	req.UserID = i.CtxBdl.Identity.UserID

	id, err := i.CtxBdl.Bdl.CopyAutoTestScene(req)
	if err != nil {
		return err
	}

	var autotestSceneRequest apistructs.AutotestSceneRequest
	autotestSceneRequest.UserID = i.CtxBdl.Identity.UserID
	autotestSceneRequest.ID = id
	autotestSceneRequest.SceneID = id
	scene, err := i.CtxBdl.Bdl.GetAutoTestScene(autotestSceneRequest)
	if err != nil {
		return err
	}

	i.State.SelectedKeys = []string{"scene-" + strconv.Itoa(int(id))}
	if scene.RefSetID != 0 {
		i.SetSceneSetClick(int(scene.RefSetID))
	} else {
		i.State.SceneId__urlQuery = strconv.Itoa(int(id))
		i.State.SetId__urlQuery = strconv.Itoa(setId)
		i.State.SceneSetKey = setId
		i.State.SceneId = id
	}
	return nil
}

func findSelectedKeysExpandedKeys(fileTreeData []SceneSet, selectedKeys string) ([]string, []string) {
	// 返回查询到的 key
	for _, v := range fileTreeData {
		if v.Key == selectedKeys {
			return []string{selectedKeys}, []string{v.Key}
		}
		for _, child := range v.Children {
			if child.Key == selectedKeys {
				return []string{selectedKeys}, []string{v.Key}
			}
		}
	}

	// 没有找到就返回第一个 key
	for _, v := range fileTreeData {
		for _, child := range v.Children {
			return []string{child.Key}, []string{v.Key}
		}
		return []string{v.Key}, []string{}
	}
	return nil, nil
}

func (i *ComponentFileTree) SetSceneSetClick(setID int) {
	i.State.PageNo = 1
	i.State.SceneSetKey = setID
	i.State.SceneId = 0
	i.State.SetId__urlQuery = strconv.Itoa(setID)
	i.State.SelectedKeys = []string{"sceneset-" + i.State.SetId__urlQuery}
	i.State.SceneId__urlQuery = ""
}
