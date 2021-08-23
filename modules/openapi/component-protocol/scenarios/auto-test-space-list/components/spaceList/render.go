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

package spaceList

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	spec "github.com/erda-project/erda/modules/openapi/component-protocol/component_spec/table"
)

type ComponentSpaceList struct {
	CtxBdl protocol.ContextBundle
	State  state                  `json:"state"`
	Props  spec.Props             `json:"props"`
	Data   map[string]interface{} `json:"data"`
}

type state struct {
	Total    int64 `json:"total"`
	PageSize int64 `json:"pageSize"`
	PageNo   int64 `json:"pageNo"`
}

type props struct {
	RowKey  string    `json:"rowKey,omitempty"`
	Columns []columns `json:"columns,omitempty"`
}

type spaceList struct {
	ID      uint64                 `json:"id"`
	Name    string                 `json:"name"`
	Desc    string                 `json:"desc"`
	Operate dataTask               `json:"operate"`
	Status  map[string]interface{} `json:"status"`
}

// apistructs.AutoTestSpaceStatus
type columns struct {
	Title     string `json:"title"`
	DataIndex string `json:"dataIndex"`
	Width     int    `json:"width,omitempty"`
}

type inParams struct {
	ProjectID int64 `json:"projectId"`
}

type RenderType string

var (
	RenderTable RenderType = "tableOperation"
)

type dataTask struct {
	RenderType RenderType             `json:"renderType"`
	Value      string                 `json:"value"`
	Operations map[string]interface{} `json:"operations"`
}

func (a *ComponentSpaceList) SetBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil {
		err := fmt.Errorf("invalid bundle")
		return err
	}
	a.CtxBdl = b
	return nil
}

func (a *ComponentSpaceList) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	err := a.SetBundle(bdl)
	if err != nil {
		return err
	}

	if a.CtxBdl.InParams == nil {
		return fmt.Errorf("params is empty")
	}

	inParamsBytes, err := json.Marshal(a.CtxBdl.InParams)
	if err != nil {
		return fmt.Errorf("failed to marshal inParams, inParams:%+v, err:%v", a.CtxBdl.InParams, err)
	}

	var inParams inParams
	if err := json.Unmarshal(inParamsBytes, &inParams); err != nil {
		return err
	}
	switch event.Operation {
	case apistructs.AutoTestSpaceChangePageNoOperationKey, apistructs.AutoTestSpaceChangePageSizeOperationKey, apistructs.InitializeOperation, apistructs.RenderingOperation:
		if err := a.handlerListOperation(bdl, c, inParams, event); err != nil {
			return err
		}
	case apistructs.AutoTestSpaceDeleteOperationKey:
		if err := a.handlerDeleteOperation(bdl, c, inParams, event); err != nil {
			return err
		}
	case apistructs.AutoTestSpaceCopyOperationKey:
		if err := a.handlerCopyOperation(bdl, c, inParams, event); err != nil {
			return err
		}
	case apistructs.AutoTestSpaceRetryOperationKey:
		if err := a.handlerRetryOperation(bdl, c, inParams, event); err != nil {
			return err
		}
	case apistructs.AutoTestSpaceExportOperationKey:
		if err := a.handlerExportOperation(bdl, c, inParams, event); err != nil {
			return err
		}
	}
	c.Operations = getOperations()
	a.Props = getProps()
	return nil
}

func getStatus(req apistructs.AutoTestSpaceStatus) map[string]interface{} {
	res := map[string]interface{}{"renderType": "textWithBadge"}
	if req == apistructs.TestSpaceFailed {
		res["status"] = "error"
		res["value"] = "失败"
	}
	if req == apistructs.TestSpaceCopying {
		res["status"] = "processing"
		res["value"] = "复制中"
	}
	return res
}

func (a *ComponentSpaceList) setData(spaces apistructs.AutoTestSpaceList) error {
	lists := []spaceList{}
	for _, each := range spaces.List {
		list := spaceList{
			ID:   each.ID,
			Name: each.Name,
			Desc: each.Description,
			Operate: dataTask{
				RenderType: RenderTable,
				Value:      "",
				Operations: map[string]interface{}{},
			},
			Status: getStatus(each.Status),
		}
		if each.Status == apistructs.TestSpaceFailed {
			edit.Command = setCommand(list)
			edit.Disabled = true
			list.Operate.Operations["a-edit"] = edit
			delete.Meta = setMeta(list)
			delete.Disabled = false
			list.Operate.Operations["delete"] = delete
			retry.Meta = setMeta(list)
			export.Meta = setMeta(list)
			export.Disabled = true
			list.Operate.Operations["export"] = export
			list.Operate.Operations["retry"] = retry
		} else {
			edit.Command = setCommand(list)
			copy.Meta = setMeta(list)
			delete.Meta = setMeta(list)
			delete.Disabled = true
			copy.Disabled = true
			edit.Disabled = true
			if each.Status == apistructs.TestSpaceOpen {
				edit.Disabled = false
				copy.Disabled = false
			}
			list.Operate.Operations["a-edit"] = edit
			list.Operate.Operations["copy"] = copy
			export.Meta = setMeta(list)
			list.Operate.Operations["export"] = export
			list.Operate.Operations["delete"] = delete
		}
		lists = append(lists, list)
	}
	a.Data["list"] = lists

	return nil
}

func getOperations() map[string]interface{} {
	return map[string]interface{}{
		"changePageNo": ListSpaceOperation{
			Key:    "changePageNo",
			Reload: true,
		},
		"changePageSize": ListSpaceOperation{
			Key:    "changePageSize",
			Reload: true,
		},
		"clickRow": ClickRowOperation{
			Key:    "clickRow",
			Reload: false,
			Command: ClickRowOperationCommand{
				Key:     "goto",
				State:   map[string]interface{}{},
				Target:  "project_test_spaceDetail_scenes",
				JumpOut: false,
			},
		},
	}
}

func getProps() spec.Props {
	return spec.Props{
		PageSizeOptions: []string{"10", "20", "50", "100"},
		RowKey:          "id",
		Columns: []spec.Column{
			{
				Title:     "空间名",
				DataIndex: "name",
			},
			{
				Title:     "描述",
				DataIndex: "desc",
			},
			{
				Title:     "状态",
				DataIndex: "status",
			},
			{
				Title:     "操作",
				DataIndex: "operate",
				Width:     180,
			},
		},
		Visible: true,
	}
}

func RenderCreator() protocol.CompRender {
	return &ComponentSpaceList{
		CtxBdl: protocol.ContextBundle{},
		State:  state{},
		Data:   map[string]interface{}{},
	}
}
