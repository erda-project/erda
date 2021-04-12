// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package resultPreviewData

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ResultPreviewData struct {
	CtxBdl     protocol.ContextBundle
	Type       string                 `json:"type"`
	Props      map[string]interface{} `json:"props"`
	Operations map[string]interface{} `json:"operations"`
	State      State                  `json:"state"`
	Data       map[string]interface{} `json:"data"`
}

type State struct {
	Total      int64  `json:"total"`
	PageSize   int64  `json:"pageSize"`
	PageNo     int64  `json:"pageNo"`
	PipelineID uint64 `json:"pipelineId"`
}

const (
	DefaultPageSize = 10
	DefaultPageNo   = 1
)

type Operation struct {
	Key    string `json:"key"`
	Reload bool   `json:"reload"`
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

type dataOperation struct {
	Key         string                 `json:"key"`
	Reload      bool                   `json:"reload"`
	Text        string                 `json:"text"`
	Disabled    bool                   `json:"disabled"`
	DisabledTip string                 `json:"disabledTip,omitempty"`
	Confirm     string                 `json:"confirm,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
	Command     map[string]interface{} `json:"command,omitempty"`
}

func (a *ResultPreviewData) Import(c *apistructs.Component) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, a); err != nil {
		return err
	}
	return nil
}

func (a *ResultPreviewData) Export(c *apistructs.Component, gs *apistructs.GlobalStateData) error {
	// set component data
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, c); err != nil {
		return err
	}
	return nil
}

func (a *ResultPreviewData) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
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
	// listen on operation
	switch event.Operation {
	case apistructs.ExecuteChangePageNoOperationKey:
		if err := a.handlerListOperation(a.CtxBdl, c, inParams, event); err != nil {
			return err
		}
	}
	// export rendered component data
	c.Operations = getOperations()
	c.Props = getProps()
	c.Type = "InfoPreview"
	return nil
}

func getOperations() map[string]interface{} {
	return map[string]interface{}{
		"changePageNo": Operation{
			Key:    "changePageNo",
			Reload: true,
		},
	}
}

func getProps() map[string]interface{} {
	return map[string]interface{}{
		"render": []interface{}{
			map[string]interface{}{
				"type":      "Title",
				"dataIndex": "title",
			},
			map[string]interface{}{
				"type":      "Desc",
				"dataIndex": "desc",
			},
			map[string]interface{}{
				"type": "BlockTitle",
				"props": map[string]interface{}{
					"title": "请求信息",
				},
			},
			map[string]interface{}{
				"type":      "API",
				"dataIndex": "apiInfo",
			},
			map[string]interface{}{
				"type":      "Table",
				"dataIndex": "header",
				"props": map[string]interface{}{
					"title": "请求头",
					"columns": []columns{
						{
							Title:     "名称",
							DataIndex: "name",
						},
						{
							Title:     "描述",
							DataIndex: "desc",
						},
						{
							Title:     "必需",
							DataIndex: "required",
						},
						{
							Title:     "默认值",
							DataIndex: "defaultValue",
						},
					},
				},
			},
			map[string]interface{}{
				"type":      "Table",
				"dataIndex": "urlParams",
				"props": map[string]interface{}{
					"title": "URL参数",
					"columns": []columns{
						{
							Title:     "名称",
							DataIndex: "name",
						},
						{
							Title:     "类型",
							DataIndex: "type",
						},
						{
							Title:     "描述",
							DataIndex: "desc",
						},
						{
							Title:     "必需",
							DataIndex: "required",
						},
						{
							Title:     "默认值",
							DataIndex: "defaultValue",
						},
					},
				},
			},
			map[string]interface{}{
				"type":      "Table",
				"dataIndex": "body",
				"props": map[string]interface{}{
					"title": "请求体",
					"columns": []columns{
						{
							Title:     "名称",
							DataIndex: "name",
						},
						{
							Title:     "类型",
							DataIndex: "type",
						},
						{
							Title:     "描述",
							DataIndex: "desc",
						},
						{
							Title:     "必需",
							DataIndex: "required",
						},
						{
							Title:     "默认值",
							DataIndex: "defaultValue",
						},
					},
				},
			},
			map[string]interface{}{
				"type":      "FileEditor",
				"dataIndex": "example",
				"props": map[string]interface{}{
					"title": "展示样例",
				},
			},
			map[string]interface{}{
				"type": "BlockTitle",
				"props": map[string]interface{}{
					"title": "响应信息",
				},
			},
			map[string]interface{}{
				"type": "Title",
				"props": map[string]interface{}{
					"title": "响应头: 无",
					"level": 2,
				},
			},
		},
	}
}

func (a *ResultPreviewData) setData(pipeline *apistructs.PipelineDetailDTO) error {
	lists := []map[string]interface{}{}
	a.Data["list"] = lists
	return nil
}

func (e *ResultPreviewData) handlerListOperation(bdl protocol.ContextBundle, c *apistructs.Component, inParams inParams, event apistructs.ComponentEvent) error {

	if e.State.PageNo == 0 {
		e.State.PageNo = DefaultPageNo
		e.State.PageSize = DefaultPageSize
	}

	list, err := bdl.Bdl.GetPipeline(e.State.PipelineID)
	if err != nil {
		return err
	}
	err = e.setData(list)
	if err != nil {
		return err
	}
	c.Data = e.Data
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ResultPreviewData{}
}
