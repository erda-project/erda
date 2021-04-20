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

package stages

import (
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentStageForm struct {
	ctxBdl protocol.ContextBundle

	CommonStageForm
}

type CommonStageForm struct {
	Version    string                 `json:"version,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Type       string                 `json:"type,omitempty"`
	State      State                  `json:"state,omitempty"`
	Operations map[string]interface{} `json:"operations,omitempty"`
	Data       Data                   `json:"data,omitempty"`
	InParams   InParams               `json:"inParams,omitempty"`
	Props      map[string]interface{} `json:"props,omitempty"`
}

type Data struct {
	List []StageData `json:"value"`
	Type string      `json:"type"`
}

type StageData struct {
	Title      string                 `json:"title"`
	ID         uint64                 `json:"id"`
	GroupID    int                    `json:"groupId"`
	Operations map[string]interface{} `json:"operations"`
}

type InParams struct {
	SceneID    string `json:"sceneId__urlQuery"`
	SceneSetID uint64 `json:"sceneSetId__urlQuery"`
}

type DragParams struct {
	DragKey  uint64 `json:"dragKey"`
	DropKey  uint64 `json:"dropKey"`
	Position int64  `json:"position"`
}

type State struct {
	Visible    bool       `json:"visible"`
	DragParams DragParams `json:"dragParams"`

	TestPlanId          uint64 `json:"testPlanId"`
	StepId              uint64 `json:"stepId"`
	ShowScenesSetDrawer bool   `json:"showScenesSetDrawer"`
}

type OperationBaseInfo struct {
	FillMeta  string `json:"fillMeta"`
	Key       string `json:"key"`
	Icon      string `json:"icon"`
	HoverTip  string `json:"hoverTip"`
	HoverShow bool   `json:"hoverShow"`
	// 操作展示名称
	Text string `json:"text"`
	// 确认提示
	Confirm string `json:"confirm,omitempty"`
	// 前端操作是否需要触发后端
	Reload      bool   `json:"reload"`
	Disabled    bool   `json:"disabled"`
	DisabledTip string `json:"disabledTip"`
}

type OpMetaData struct {
	Type   apistructs.StepAPIType   `json:"type"`   // 类型
	Method apistructs.StepAPIMethod `json:"method"` // method
	Value  string                   `json:"value"`  // 值
	Name   string                   `json:"name"`   // 名称
	ID     uint64                   `json:"id"`
}

type OpMetaInfo struct {
	ID   uint64                 `json:"id"`
	Data map[string]interface{} `json:"data"`
}

type OperationInfo struct {
	OperationBaseInfo
	Meta OpMetaInfo `json:"meta"`
}

type CreateOperation OperationInfo
type OnChangeOperation OperationInfo
