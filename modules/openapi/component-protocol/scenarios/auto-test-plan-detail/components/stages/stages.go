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
	"fmt"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
)

func RenderStage(step apistructs.TestPlanV2Step) (StageData, error) {
	groupID := int(step.GroupID)
	if groupID == 0 {
		groupID = int(step.ID)
	}
	pd := StageData{
		Title:      fmt.Sprintf("#%d 场景集: %s", step.ID, step.SceneSetName),
		ID:         step.ID,
		GroupID:    groupID,
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
	for _, v := range rsp.Data.Steps {
		stageData, err := RenderStage(*v)
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
		Meta: OpMetaInfo{},
	}
	i.Operations["moveItem"] = omi
	i.Operations["moveGroup"] = omg
	i.Operations["clickItem"] = ocl

	return nil
}

func (i *ComponentStageForm) RenderCreateStagesForm(opsData interface{}) error {
	meta, err := GetOpsInfo(opsData)
	if err != nil {
		return err
	}
	preStep, err := i.ctxBdl.Bdl.GetTestPlanV2Step(meta.ID)
	if err != nil {
		return err
	}
	groupID := preStep.GroupID
	if groupID == 0 {
		groupID = preStep.ID
	}

	_, err = i.ctxBdl.Bdl.CreateTestPlansV2Step(apistructs.TestPlanV2StepAddRequest{
		PreID:      meta.ID,
		TestPlanID: i.State.TestPlanId,
		GroupID:    groupID,
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

func (i *ComponentStageForm) RenderGroupMoveStagesForm() (err error) {
	var (
		req apistructs.TestPlanV2StepMoveRequest
	)
	//	dragGroupKey := uint64(i.State.DragParams.DragGroupKey)
	//	dropGroupKey := uint64(i.State.DragParams.DropGroupKey)

	req.UserID = i.ctxBdl.Identity.UserID
	req.TestPlanID = i.State.TestPlanId
	req.IsGroup = true

	testPlanDrag, err := i.ctxBdl.Bdl.GetTestPlanV2(i.State.TestPlanId)
	if err != nil {
		return err
	}
	if len(testPlanDrag.Data.Steps) <= 0 {
		return errors.New("the dragGroupKey is not exists")
	}
	req.StepID = testPlanDrag.Data.Steps[0].ID
	req.LastStepID = testPlanDrag.Data.Steps[len(testPlanDrag.Data.Steps)-1].ID

	switch i.State.DragParams.Position {
	case 0: // inside target
		return nil
	case -1: // in front of the target
		testPlanDrop, err := i.ctxBdl.Bdl.GetTestPlanV2(i.State.TestPlanId)
		if err != nil {
			return err
		}
		if len(testPlanDrop.Data.Steps) <= 0 {
			return errors.New("the dropGroupKey is not exists")
		}
		req.PreID = testPlanDrop.Data.Steps[0].PreID
	case 1: // behind the target
		testPlanDrop, err := i.ctxBdl.Bdl.GetTestPlanV2(i.State.TestPlanId)
		if err != nil {
			return err
		}
		if len(testPlanDrop.Data.Steps) <= 0 {
			return errors.New("the dropGroupKey is not exists")
		}
		req.PreID = testPlanDrop.Data.Steps[len(testPlanDrop.Data.Steps)-1].ID

	default:
		return errors.New("unknown position")
	}
	return i.ctxBdl.Bdl.MoveTestPlansV2Step(req)
}

func (i *ComponentStageForm) RenderItemMoveStagesForm() (err error) {
	var (
		step     *apistructs.TestPlanV2Step
		req      apistructs.TestPlanV2StepMoveRequest
		testPlan *apistructs.TestPlanV2GetResponse
	)
	req.UserID = i.ctxBdl.Identity.UserID
	req.TestPlanID = i.State.TestPlanId
	req.StepID = uint64(i.State.DragParams.DragKey)
	req.LastStepID = uint64(i.State.DragParams.DragKey)
	req.IsGroup = false
	if i.State.DragParams.DropKey == -1 {
		req.TargetStepID = 0
	} else {
		req.TargetStepID = uint64(i.State.DragParams.DropKey)
	}

	// find preID
	if i.State.DragParams.DropKey == -1 { // move to the end and be independent group
		testPlan, err = i.ctxBdl.Bdl.GetTestPlanV2(i.State.TestPlanId)
		if err != nil {
			return err
		}
		req.PreID = testPlan.Data.Steps[len(testPlan.Data.Steps)-1].ID
	} else {
		switch i.State.DragParams.Position {
		case 0: // inside target
			return nil
		case 1: // behind the target
			step, err = i.ctxBdl.Bdl.GetTestPlanV2Step(uint64(i.State.DragParams.DragKey))
			if err != nil {
				return
			}
			req.PreID = uint64(i.State.DragParams.DropKey)
		case -1: // in front of the target
			step, err = i.ctxBdl.Bdl.GetTestPlanV2Step(uint64(i.State.DragParams.DropKey))
			if err != nil {
				return
			}
			req.PreID = step.PreID
		default:
			return errors.New("unknown position")
		}
	}
	return i.ctxBdl.Bdl.MoveTestPlansV2Step(req)
}
