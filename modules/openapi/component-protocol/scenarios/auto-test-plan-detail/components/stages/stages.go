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
	"github.com/erda-project/erda/apistructs"
)

func RenderStage(index int, step apistructs.TestPlanV2Step) (StageData, error) {
	pd := StageData{
		Title:      "场景集: " + step.SceneSetName,
		ID:         step.ID,
		GroupID:    index,
		Operations: make(map[string]interface{}),
	}
	o := CreateOperation{}
	o.Key = apistructs.AutoTestSceneStepCreateOperationKey.String()
	o.Icon = "add"
	o.HoverTip = "添加步骤"
	o.Disabled = false
	o.Reload = true
	o.HoverShow = true
	o.Meta.ID = step.ID
	pd.Operations["add"] = o

	oc := OperationInfo{}
	oc.Key = apistructs.AutoTestSceneStepDeleteOperationKey.String()
	oc.Icon = "shanchu"
	oc.Disabled = false
	oc.Reload = true
	oc.Meta.ID = step.ID
	pd.Operations["delete"] = oc

	return pd, nil
}

func (i *ComponentStageForm) RenderListStageForm() error {
	rsp, err := i.ctxBdl.Bdl.GetTestPlanV2(i.State.TestPlanId)
	if err != nil {
		return err
	}
	var list []StageData
	for i, v := range rsp.Data.Steps {
		stageData, err := RenderStage(i, *v)
		if err != nil {
			return err
		}
		list = append(list, stageData)
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
	ocl := OperationInfo{
		OperationBaseInfo: OperationBaseInfo{
			Key:      "clickItem",
			Reload:   true,
			FillMeta: "data",
		},
		Meta: OpMetaInfo{},
	}
	i.Operations["moveItem"] = omi
	i.Operations["clickItem"] = ocl

	return nil
}

func (i *ComponentStageForm) RenderCreateStagesForm(opsData interface{}) error {
	meta, err := GetOpsInfo(opsData)
	if err != nil {
		return err
	}

	_, err = i.ctxBdl.Bdl.CreateTestPlansV2Step(apistructs.TestPlanV2StepAddRequest{
		PreID:      meta.ID,
		TestPlanID: i.State.TestPlanId,
		IdentityInfo: apistructs.IdentityInfo{
			UserID: i.ctxBdl.Identity.UserID,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (i *ComponentStageForm) RenderDeleteStagesForm(opsData interface{}) error {
	meta, err := GetOpsInfo(opsData)
	if err != nil {
		return err
	}

	err = i.ctxBdl.Bdl.DeleteTestPlansV2Step(apistructs.TestPlanV2StepDeleteRequest{
		StepID:     meta.ID,
		TestPlanID: i.State.TestPlanId,
		IdentityInfo: apistructs.IdentityInfo{
			UserID: i.ctxBdl.Identity.UserID,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (i *ComponentStageForm) RenderMoveStagesForm() (err error) {
	var (
		step *apistructs.TestPlanV2Step
		req  apistructs.TestPlanV2StepUpdateRequest
	)
	req.UserID = i.ctxBdl.Identity.UserID
	req.TestPlanID = i.State.TestPlanId
	req.StepID = i.State.DragParams.DragKey
	if i.State.DragParams.Position == -1 {
		step, err = i.ctxBdl.Bdl.GetTestPlanV2Step(i.State.DragParams.DropKey)
		if err != nil {
			return
		}
		if step.PreID == i.State.DragParams.DragKey {
			return
		}
		req.PreID = step.PreID
	} else {
		step, err = i.ctxBdl.Bdl.GetTestPlanV2Step(i.State.DragParams.DragKey)
		if err != nil {
			return
		}
		if step.PreID == i.State.DragParams.DropKey {
			return
		}
		req.PreID = i.State.DragParams.DropKey
	}
	return i.ctxBdl.Bdl.MoveTestPlansV2Step(req)
}
