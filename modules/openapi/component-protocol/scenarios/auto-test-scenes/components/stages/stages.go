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

package stages

import (
	"encoding/json"
	"strconv"

	"github.com/erda-project/erda/apistructs"
)

type MU struct {
	URL    string `json:"url"`
	Method string `json:"method"`
}

type value struct {
	ApiSpec MU `json:"apiSpec"`
}

func RenderStage(groupID uint64, step apistructs.AutoTestSceneStep) (StageData, error) {
	title := "#" + strconv.Itoa(int(step.ID)) + " "
	if step.Type == apistructs.StepTypeWait {
		if step.Value == "" {
			title = title + "空 等待"
		} else {
			var value apistructs.AutoTestRunWait
			if err := json.Unmarshal([]byte(step.Value), &value); err != nil {
				return StageData{}, err
			}
			title = title + "等待 " + strconv.Itoa(value.WaitTime) + " 秒"
		}
	} else if step.Type == apistructs.StepTypeAPI {
		if step.Value == "" {
			title = title + "空 接口"
		} else {
			var value value
			if err := json.Unmarshal([]byte(step.Value), &value); err != nil {
				return StageData{}, err
			}
			title = title + "接口: " + step.Name + " " + value.ApiSpec.Method + ":" + value.ApiSpec.URL
		}
	} else if step.Type == apistructs.StepTypeConfigSheet {
		//if step.Value == "" {
		//	title = "空 配置单"
		//} else {
		//var value apistructs.AutoTestRunConfigSheet
		//if err := json.Unmarshal([]byte(step.Value), &value); err != nil {
		//	return StageData{}, err
		//}
		//title = "配置单: " + value.ConfigSheetName
		title = title + "配置单: " + step.Name
		//}
	} else if step.Type == apistructs.StepTypeScene {
		title = title + "嵌套场景: " + step.Name
	} else if step.Type == apistructs.StepTypeCustomScript {
		title = title + "自定义任务: " + step.Name
	}
	pd := StageData{
		Title:      title,
		ID:         step.ID,
		GroupID:    int64(groupID),
		Operations: make(map[string]interface{}),
	}
	if step.Type == apistructs.StepTypeAPI {
		o := CreateOperation{}
		o.Key = apistructs.AutoTestSceneStepCopyOperationKey.String()
		o.Icon = "fz1"
		o.HoverTip = "复制接口"
		o.Disabled = false
		o.Reload = true
		o.HoverShow = true
		o.Meta.ID = step.ID
		pd.Operations["copy"] = o

		o2 := CreateOperation{}
		o2.Key = apistructs.AutoTestSceneStepCreateOperationKey.String()
		o2.Icon = "add"
		o2.HoverTip = "添加并行接口"
		o2.Disabled = false
		o2.Reload = true
		o2.HoverShow = true
		o2.Meta.ID = step.ID
		pd.Operations["add"] = o2
	}

	os := OperationInfo{
		OperationBaseInfo: OperationBaseInfo{
			Icon:     "split",
			Key:      apistructs.AutoTestSceneStepSplitOperationKey.String(),
			HoverTip: "改为串行",
			Disabled: false,
			Reload:   true,
		},
	}
	os.Meta.ID = step.ID
	if groupID == step.ID && len(step.Children) == 0 {

		os.Disabled = true
	}
	pd.Operations["split"] = os

	oc := OperationInfo{}
	oc.Key = apistructs.AutoTestSceneStepDeleteOperationKey.String()
	oc.Icon = "shanchu"
	oc.Disabled = false
	oc.Reload = true
	oc.Confirm = "是否确认删除"
	oc.Meta = OpMetaInfo{AutotestSceneRequest: apistructs.AutotestSceneRequest{
		AutoTestSceneParams: apistructs.AutoTestSceneParams{
			ID: pd.ID,
		},
	}}
	pd.Operations["delete"] = oc

	return pd, nil
}

func (i *ComponentStageForm) RenderListStageForm() error {
	rsp, err := i.ctxBdl.Bdl.ListAutoTestSceneStep(i.State.AutotestSceneRequest)

	if err != nil {
		return err
	}
	var list []StageData
	for _, v := range rsp {
		stageData, err := RenderStage(v.ID, v)
		if err != nil {
			return err
		}
		list = append(list, stageData)
		for _, s := range v.Children {
			stageData, err := RenderStage(v.ID, s)
			if err != nil {
				return err
			}
			list = append(list, stageData)
		}
	}

	i.Data.List = list
	i.Data.Type = "sort-item"

	i.Operations = make(map[string]interface{})
	omi := OperationInfo{
		OperationBaseInfo: OperationBaseInfo{
			Key:    apistructs.AutoTestSceneStepMoveItemOperationKey.String(),
			Reload: true,
		},
	}
	omg := OperationInfo{
		OperationBaseInfo: OperationBaseInfo{
			Key:    apistructs.AutoTestSceneStepMoveGroupOperationKey.String(),
			Reload: true,
		},
	}
	ocl := OperationInfo{
		OperationBaseInfo: OperationBaseInfo{
			Key:      "clickItem",
			Reload:   true,
			FillMeta: "data",
		},
	}
	i.Operations["moveItem"] = omi
	i.Operations["moveGroup"] = omg
	i.Operations["clickItem"] = ocl

	return nil
}

func (i *ComponentStageForm) RenderUpdateStagesForm() error {
	_, err := i.ctxBdl.Bdl.UpdateAutoTestSceneStep(i.State.AutotestSceneRequest)
	if err != nil {
		return err
	}
	return nil
}

func (i *ComponentStageForm) RenderCopyStagesForm(opsData interface{}) error {
	req, err := GetOpsInfo(opsData)
	if err != nil {
		return err
	}
	i.State.AutotestSceneRequest.ID = req.ID

	id, err := i.ctxBdl.Bdl.CreateAutoTestSceneStep(i.State.AutotestSceneRequest)
	if err != nil {
		return err
	}
	i.State.StepId = id
	return nil
}

func (i *ComponentStageForm) RenderCreateStagesForm(opsData interface{}) error {
	mate, err := GetOpsInfo(opsData)
	if err != nil {
		return err
	}

	stepReq := apistructs.AutotestGetSceneStepReq{
		ID: mate.ID,
	}
	oldStep, err := i.ctxBdl.Bdl.GetAutoTestSceneStep(stepReq)
	if err != nil {
		return err
	}
	req := apistructs.AutotestSceneRequest{
		SceneID:  oldStep.SceneID,
		Type:     oldStep.Type,
		Target:   int64(oldStep.ID),
		PreType:  apistructs.PreTypeParallel,
		GroupID:  1,
		Position: 1,
		IdentityInfo: apistructs.IdentityInfo{
			UserID: i.ctxBdl.Identity.UserID,
		},
	}
	id, err := i.ctxBdl.Bdl.CreateAutoTestSceneStep(req)
	if err != nil {
		return err
	}
	i.State.StepId = id
	return nil
}

func (i *ComponentStageForm) RenderDeleteStagesForm(opsData interface{}) error {
	req, err := GetOpsInfo(opsData)
	if err != nil {
		return err
	}
	i.State.AutotestSceneRequest.ID = req.ID

	err = i.ctxBdl.Bdl.DeleteAutoTestSceneStep(i.State.AutotestSceneRequest)
	if err != nil {
		return err
	}
	return nil
}

func (i *ComponentStageForm) RenderMoveStagesForm() error {
	req := apistructs.AutotestSceneRequest{
		SceneID:  i.State.AutotestSceneRequest.SceneID,
		Position: i.State.DragParams.Position,
		GroupID:  1,
		IdentityInfo: apistructs.IdentityInfo{
			UserID: i.State.AutotestSceneRequest.UserID,
		},
	}
	if i.State.DragParams.DragKey == 0 {
		// 拖动整组
		req.ID = uint64(i.State.DragParams.DragGroupKey)
		req.Target = i.State.DragParams.DropGroupKey
		req.IsGroup = true
	} else {
		// 把单个拖到最后单独成组
		if i.State.DragParams.DropGroupKey == -1 && i.State.DragParams.DropKey == -1 {
			req.ID = uint64(i.State.DragParams.DragKey)
			req.GroupID = -1
			req.Target = -1
		} else {
			// 拖动单个步骤与target合并
			req.ID = uint64(i.State.DragParams.DragKey)
			req.Target = i.State.DragParams.DropKey
		}

	}
	_, err := i.ctxBdl.Bdl.MoveAutoTestSceneStep(req)
	if err != nil {
		return err
	}
	return nil
}

func (i *ComponentStageForm) RenderSplitStagesForm(opsData interface{}) error {
	meta, err := GetOpsInfo(opsData)
	if err != nil {
		return err
	}
	req := apistructs.AutotestSceneRequest{
		SceneID:  i.State.AutotestSceneRequest.SceneID,
		Position: i.State.DragParams.Position,
		GroupID:  -1,
		IdentityInfo: apistructs.IdentityInfo{
			UserID: i.State.AutotestSceneRequest.UserID,
		},
	}
	req.ID = meta.ID
	_, err = i.ctxBdl.Bdl.MoveAutoTestSceneStep(req)
	if err != nil {
		return err
	}
	return nil
}
