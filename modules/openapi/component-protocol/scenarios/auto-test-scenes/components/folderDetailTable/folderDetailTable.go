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

package folderDetailTable

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
)

func GetOpsInfo(opsData interface{}) (*Meta, error) {
	if opsData == nil {
		err := fmt.Errorf("empty operation data")
		return nil, err
	}
	var op DataOperation
	cont, err := json.Marshal(opsData)
	if err != nil {
		logrus.Errorf("marshal inParams failed, content:%v, err:%v", opsData, err)
		return nil, err
	}
	err = json.Unmarshal(cont, &op)
	if err != nil {
		logrus.Errorf("unmarshal move out request failed, content:%v, err:%v", cont, err)
		return nil, err
	}
	meta := op.Meta
	return &meta, nil
}

func (ca *ComponentAction) ListScene() error {
	return ca.doListScene(ca.State.AutotestSceneRequest)
}

func (ca *ComponentAction) doListScene(req apistructs.AutotestSceneRequest) error {
	total, list, err := ca.ctxBdl.Bdl.ListAutoTestScene(req)
	if err != nil {
		return err
	}
	data := []Data{}
	ca.UserIDs = nil
	for _, v := range list {
		createStr := v.CreateAt.Format("2006-01-02")
		ops := make(map[string]interface{})
		ops["delete"] = DataOperation{
			Key:     apistructs.AutoTestFolderDetailDeleteOperationKey.String(),
			Text:    "删除",
			Reload:  true,
			Confirm: "是否确认删除",
			Meta:    Meta{ID: v.ID},
		}
		if v.RefSetID <= 0 {
			ops["edit"] = DataOperation{
				Key:    apistructs.AutoTestFolderDetailEditOperationKey.String(),
				Text:   "编辑",
				Reload: true,
				Meta:   Meta{ID: v.ID},
			}
		}
		ops["copy"] = DataOperation{
			Key:    apistructs.AutoTestFolderDetailCopyOperationKey.String(),
			Text:   "复制",
			Reload: true,
			Meta:   Meta{ID: v.ID},
		}
		dop := DataOperate{
			RenderType: "tableOperation",
			Operations: ops,
		}
		dt := Data{
			ID:        v.ID,
			CaseName:  v.Name,
			StepCount: strconv.Itoa(int(v.StepCount)),
			LatestStatus: LatestStatus{
				RenderType: "textWithBadge",
				Value:      v.Status.Value(),
				Status:     v.Status,
			},
			Creator: Creator{
				RenderType: "userAvatar",
				Value:      v.CreatorID,
			},
			CreatedAt: createStr,
			Operate:   dop,
		}
		ca.UserIDs = append(ca.UserIDs, v.CreatorID)
		data = append(data, dt)
	}
	ca.Data = data
	ca.State.Total = total
	return nil
}

func (ca *ComponentAction) DeleteScene(ops interface{}) error {
	mt, err := GetOpsInfo(ops)
	if err != nil {
		return err
	}
	ca.State.AutotestSceneRequest.SceneID = mt.ID
	if err := ca.ctxBdl.Bdl.DeleteAutoTestScene(ca.State.AutotestSceneRequest); err != nil {
		return err
	}
	return nil
}

func (ca *ComponentAction) UpdateScene(ops interface{}, c *apistructs.Component) error {
	mt, err := GetOpsInfo(ops)
	if err != nil {
		return err
	}

	// 打开fileFormModel
	c.State["visible"] = true
	c.State["sceneId"] = mt.ID
	c.State["actionType"] = "UpdateScene"

	c.State["isClick"] = true
	return nil
}

func (ca *ComponentAction) CopyScene(ops interface{}, c *apistructs.Component, inParams InParams) error {
	mt, err := GetOpsInfo(ops)
	if err != nil {
		return err
	}

	setId := uint64(mt.ID)
	req := apistructs.AutotestSceneCopyRequest{
		SpaceID: inParams.SpaceId,
		PreID:   setId,
		SceneID: setId,
		SetID:   ca.State.SetId,
	}
	req.UserID = ca.ctxBdl.Identity.UserID

	_, err = ca.ctxBdl.Bdl.CopyAutoTestScene(req)
	if err != nil {
		return err
	}
	return nil
}

func getOperation(operationData *ClickRowOperation, event apistructs.ComponentEvent) error {
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

func (ca *ComponentAction) ClickScene(event apistructs.ComponentEvent) error {
	var operationData ClickRowOperation
	if err := getOperation(&operationData, event); err != nil {
		return err
	}
	ca.State.IsClickFolderTableRow = true
	ca.State.ClickFolderTableRowID = operationData.Meta.RowData.ID

	var req apistructs.AutotestSceneRequest
	req.SceneID = operationData.Meta.RowData.ID
	req.UserID = ca.ctxBdl.Identity.UserID
	scene, err := ca.ctxBdl.Bdl.GetAutoTestScene(req)
	if err != nil {
		return err
	}
	if scene.RefSetID != 0 {
		var req apistructs.AutotestSceneRequest
		req.SetID = scene.RefSetID
		req.UserID = ca.ctxBdl.Identity.UserID
		ca.Props["visible"] = true
		return ca.doListScene(req)
	} else {
		ca.Props["visible"] = false
	}
	return nil
}
