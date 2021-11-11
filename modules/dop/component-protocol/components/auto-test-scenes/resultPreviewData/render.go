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

package resultPreviewData

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type ResultPreviewData struct {
	base.DefaultProvider
	sdk        *cptype.SDK
	bdl        *bundle.Bundle
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

func init() {
	base.InitProviderWithCreator("auto-test-scenes", "resultPreviewData",
		func() servicehub.Provider { return &ResultPreviewData{} })
}

func (a *ResultPreviewData) Import(c *cptype.Component) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, a); err != nil {
		return err
	}
	return nil
}

func (a *ResultPreviewData) Export(c *cptype.Component, gs *cptype.GlobalStateData) error {
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

func (a *ResultPreviewData) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	// import component data
	if err := a.Import(c); err != nil {
		logrus.Errorf("import component failed, err:%v", err)
		return err
	}

	a.sdk = cputil.SDK(ctx)
	a.bdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	if cputil.SDK(ctx).InParams == nil {
		return fmt.Errorf("params is empty")
	}
	inParamsBytes, err := json.Marshal(cputil.SDK(ctx).InParams)
	if err != nil {
		return fmt.Errorf("failed to marshal inParams, inParams:%+v, err:%v", cputil.SDK(ctx).InParams, err)
	}
	var inParams inParams
	if err := json.Unmarshal(inParamsBytes, &inParams); err != nil {
		return err
	}
	// listen on operation
	switch event.Operation {
	case cptype.OperationKey(apistructs.ExecuteChangePageNoOperationKey):
		if err := a.handlerListOperation(c, inParams, event); err != nil {
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

func (e *ResultPreviewData) handlerListOperation(c *cptype.Component, inParams inParams, event cptype.ComponentEvent) error {

	if e.State.PageNo == 0 {
		e.State.PageNo = DefaultPageNo
		e.State.PageSize = DefaultPageSize
	}

	list, err := e.bdl.GetPipeline(e.State.PipelineID)
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
