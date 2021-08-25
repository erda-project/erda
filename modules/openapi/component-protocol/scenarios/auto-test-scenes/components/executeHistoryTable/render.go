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

package executeHistoryTable

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ExecuteHistoryTable struct {
	CtxBdl     protocol.ContextBundle
	Type       string                 `json:"type"`
	State      State                  `json:"state"`
	Props      map[string]interface{} `json:"props"`
	Operations map[string]interface{} `json:"operations"`
	Data       map[string]interface{} `json:"data"`
}

type State struct {
	Total      int64  `json:"total"`
	PageSize   int64  `json:"pageSize"`
	PageNo     int64  `json:"pageNo"`
	SceneID    int64  `json:"sceneId"`
	PipelineID uint64 `json:"pipelineId"`
}

const (
	DefaultPageSize = 15
	DefaultPageNo   = 1
)

type Operation struct {
	Key      string                 `json:"key"`
	Reload   bool                   `json:"reload"`
	FillMeta string                 `json:"fillMeta"`
	Meta     map[string]interface{} `json:"meta"`
}

type props struct {
	Key            string                   `json:"key"`
	Label          string                   `json:"label"`
	Component      string                   `json:"component"`
	Required       bool                     `json:"required"`
	Rules          []map[string]interface{} `json:"rules"`
	ComponentProps map[string]interface{}   `json:"componentProps,omitempty"`
}

type inParams struct {
	ProjectID int64 `json:"projectId"`
}

type columns struct {
	Title     string `json:"title"`
	DataIndex string `json:"dataIndex"`
	Width     int    `json:"width,omitempty"`
}

type operationData struct {
	Meta meta `json:"meta"`
}

type rowData struct {
	PipelineID uint64 `json:"pipelineId"`
}

type meta struct {
	RowData rowData `json:"rowData"`
}

func (a *ExecuteHistoryTable) Import(c *apistructs.Component) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, a); err != nil {
		return err
	}
	return nil
}

func (a *ExecuteHistoryTable) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	// import component data
	if err := a.Import(c); err != nil {
		logrus.Errorf("import component failed, err:%v", err)
		return err
	}
	a.CtxBdl = ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

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

	defer func() {
		fail := a.marshal(c)
		if err == nil && fail != nil {
			err = fail
		}
		// export rendered component data
		c.Operations = getOperations()
		c.Props = getProps()
	}()

	// listen on operation
	switch event.Operation {
	case apistructs.ExecuteChangePageNoOperationKey, apistructs.RenderingOperation, apistructs.InitializeOperation:
		if err := a.handlerListOperation(a.CtxBdl, c, inParams, event); err != nil {
			return err
		}
	case apistructs.ExecuteClickRowNoOperationKey:
		if err := a.handlerClickRowOperation(a.CtxBdl, c, inParams, event); err != nil {
			return err
		}
	}

	return nil
}

func (a *ExecuteHistoryTable) marshal(c *apistructs.Component) error {
	stateValue, err := json.Marshal(a.State)
	if err != nil {
		return err
	}
	var state map[string]interface{}
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}

	propValue, err := json.Marshal(a.Props)
	if err != nil {
		return err
	}
	var props interface{}
	err = json.Unmarshal(propValue, &props)
	if err != nil {
		return err
	}
	c.Props = getProps()
	c.State = state
	c.Type = a.Type
	return nil
}

func getOperations() map[string]interface{} {
	return map[string]interface{}{
		"changePageNo": Operation{
			Key:    "changePageNo",
			Reload: true,
		},
		"clickRow": Operation{
			Key:      "clickRow",
			Reload:   true,
			FillMeta: "rowData",
			Meta:     map[string]interface{}{"rowData": ""},
		},
	}
}

func getProps() map[string]interface{} {
	return map[string]interface{}{
		"rowKey": "id",
		"columns": []columns{
			{
				Title:     "版本",
				DataIndex: "version",
				Width:     60,
			},
			{
				Title:     "ID",
				DataIndex: "pipelineId",
			},
			{
				Title:     "状态",
				DataIndex: "status",
			},
			{
				Title:     "执行人",
				DataIndex: "runUser",
			},
			{
				Title:     "触发时间",
				DataIndex: "triggerTime",
			},
		},
	}
}

func getStatus(req apistructs.PipelineStatus) map[string]interface{} {
	res := map[string]interface{}{"renderType": "textWithBadge", "value": req.ToDesc()}
	if req.IsFailedStatus() {
		res["status"] = "error"
	} else if req.IsSuccessStatus() {
		res["status"] = "success"
	} else if req.IsReconcilerRunningStatus() {
		res["status"] = "processing"
	} else {
		res["status"] = "default"
	}
	return res
}

func (e *ExecuteHistoryTable) setData(pipeline *apistructs.PipelinePageListData, num int64, event apistructs.OperationKey) error {
	lists := []map[string]interface{}{}
	if len(pipeline.Pipelines) > 0 && event != apistructs.ExecuteChangePageNoOperationKey {
		e.State.PipelineID = pipeline.Pipelines[0].ID
	} else if len(pipeline.Pipelines) == 0 {
		e.State.PipelineID = 0
	}
	for _, each := range pipeline.Pipelines {
		var timeLayoutStr = "2006-01-02 15:04:05" //go中的时间格式化必须是这个时间
		var runUser string
		if each.Extra.RunUser != nil {
			runUser = each.Extra.RunUser.Name
		}
		list := map[string]interface{}{
			//"id":          each.ID,
			"version":     "#" + strconv.FormatInt(num, 10),
			"pipelineId":  each.ID,
			"status":      getStatus(each.Status),
			"runUser":     runUser,
			"triggerTime": each.TimeCreated.Format(timeLayoutStr),
		}
		lists = append(lists, list)
		num--
	}
	e.Data["list"] = lists
	return nil
}

func (e *ExecuteHistoryTable) handlerListOperation(bdl protocol.ContextBundle, c *apistructs.Component, inParams inParams, event apistructs.ComponentEvent) error {
	if e.State.SceneID == 0 {
		c.Data = map[string]interface{}{}
		e.State = State{
			SceneID:    0,
			PipelineID: 0,
		}
		return nil
	}
	req := apistructs.PipelinePageListRequest{
		YmlNames: []string{strconv.FormatInt(e.State.SceneID, 10)},
		Sources:  []apistructs.PipelineSource{apistructs.PipelineSourceAutoTest},
	}
	if e.State.PageNo == 0 {
		e.State.PageNo = DefaultPageNo
		e.State.PageSize = DefaultPageSize
	}
	req.PageNum = int(e.State.PageNo)
	req.PageSize = int(e.State.PageSize)
	list, err := bdl.Bdl.PageListPipeline(req)
	if err != nil {
		return err
	}
	e.State.Total = list.Total
	err = e.setData(list, list.Total-(e.State.PageNo-1)*e.State.PageSize, event.Operation)
	if err != nil {
		return err
	}
	c.Data = e.Data
	return nil
}

func (e *ExecuteHistoryTable) handlerClickRowOperation(bdl protocol.ContextBundle, c *apistructs.Component, inParams inParams, event apistructs.ComponentEvent) error {
	res := operationData{}
	b, err := json.Marshal(event.OperationData)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &res); err != nil {
		return err
	}
	e.State.PipelineID = res.Meta.RowData.PipelineID
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ExecuteHistoryTable{
		Props:      map[string]interface{}{},
		Operations: map[string]interface{}{},
		Data:       map[string]interface{}{},
	}
}
