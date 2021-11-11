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
	"context"
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/pkg/strutil"
)

func init() {
	base.InitProviderWithCreator("auto-test-scenes", "folderDetailTable",
		func() servicehub.Provider { return &ComponentAction{} })
}

func (ca *ComponentAction) RenderState(c *cptype.Component) error {
	var state State
	b, err := json.Marshal(c.State)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &state); err != nil {
		return err
	}
	ca.State = state
	return nil
}

func (ca *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	ca.bdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	ca.sdk = cputil.SDK(ctx)
	if c.State == nil {
		c.State = map[string]interface{}{}
	}
	if err := ca.RenderState(c); err != nil {
		return err
	}

	// init
	c.State["visible"] = false
	c.State["sceneSetKey"] = 0
	c.State["actionType"] = ""
	if ca.State.IsClick {
		ca.State.IsClick = false
		ca.State.SceneId = 0
	}
	if ca.State.IsClickFolderTableRow {
		ca.State.IsClickFolderTableRow = false
		ca.State.ClickFolderTableRowID = 0
	}

	ca.State.PageSize = 50
	ca.State.AutotestSceneRequest = apistructs.AutotestSceneRequest{
		SetID:    ca.State.SetId,
		SceneID:  ca.State.SceneId,
		PageNo:   ca.State.PageNo,
		PageSize: ca.State.PageSize,
	}
	ca.State.AutotestSceneRequest.UserID = ca.sdk.Identity.UserID

	// set props
	ca.setProps(c)
	if ca.Props["visible"] == false {
		return nil
	}

	ca.Operations = make(map[string]interface{})
	var page = Operation{
		Key:    "changePageNo",
		Reload: true,
	}
	ca.Operations["changePageNo"] = page
	ca.Operations["clickRow"] = ClickRowOperation{
		Key:      "clickRow",
		Reload:   true,
		FillMeta: "rowData",
		Meta: ClickMeta{
			RowData: Data{},
		},
	}

	inParamsBytes, err := json.Marshal(cputil.SDK(ctx).InParams)
	if err != nil {
		return fmt.Errorf("failed to marshal inParams, inParams:%+v, err:%v", cputil.SDK(ctx).InParams, err)
	}

	var inParams InParams
	if err := json.Unmarshal(inParamsBytes, &inParams); err != nil {
		return err
	}

	switch apistructs.OperationKey(event.Operation) {
	case apistructs.InitializeOperation, apistructs.RenderingOperation:
		if err := ca.ListScene(); err != nil {
			return err
		}
	case apistructs.AutoTestFolderDetailDeleteOperationKey:
		if err := ca.DeleteScene(event.OperationData); err != nil {
			return err
		}
		if err := ca.ListScene(); err != nil {
			return err
		}
	case apistructs.AutoTestFolderDetailEditOperationKey:
		if err := ca.UpdateScene(event.OperationData, c); err != nil {
			return err
		}
		if err := ca.ListScene(); err != nil {
			return err
		}
	case apistructs.AutoTestFolderDetailPageOperationKey:
		if err := ca.ListScene(); err != nil {
			return err
		}
	case apistructs.AutoTestFolderDetailCopyOperationKey:
		if err := ca.CopyScene(event.OperationData, c, inParams); err != nil {
			return err
		}
		if err := ca.ListScene(); err != nil {
			return err
		}
	case apistructs.AutoTestFolderDetailClickOperationKey:
		if err := ca.ClickScene(event); err != nil {
			return err
		}
	}

	// set state
	c.Operations = ca.Operations
	setState(c, ca.State)
	if c.Data == nil {
		c.Data = make(map[string]interface{})
	}
	c.Data["list"] = ca.Data
	(*gs)[protocol.GlobalInnerKeyUserIDs.String()] = strutil.DedupSlice(ca.UserIDs, true)
	return nil
}

func setState(c *cptype.Component, state State) {
	// c.State["setId"] = state.AutotestSceneRequest.SetID
	// c.State["sceneId"] = state.AutotestSceneRequest.SceneID
	c.State["pageNo"] = state.PageNo
	c.State["pageSize"] = state.PageSize
	c.State["total"] = state.Total
	c.State["isClickFolderTableRow"] = state.IsClickFolderTableRow
	c.State["clickFolderTableRowID"] = state.ClickFolderTableRowID
}

func (ca *ComponentAction) setProps(c *cptype.Component) {
	props := make(map[string]interface{})
	if ca.State.SceneId == 0 && ca.State.SetId != 0 {
		props["visible"] = true
	} else {
		props["visible"] = false
	}

	props["columns"] = []map[string]interface{}{
		{"title": "用例名称", "dataIndex": "caseName"},
		{"title": "步骤数", "dataIndex": "stepCount"},
		{"title": "最新状态", "dataIndex": "latestStatus"},
		{"title": "创建人", "dataIndex": "creator"},
		{"title": "创建时间", "dataIndex": "createdAt"},
		{"title": "操作", "dataIndex": "operate", "width": 150},
	}
	props["rowKey"] = "id"
	ca.Props = props
	c.Props = props
}
