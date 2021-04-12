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

package folderDetailTable

import (
	"github.com/erda-project/erda/apistructs"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentAction struct {
	ctxBdl protocol.ContextBundle

	State      State                  `json:"state"`
	Data       []Data                 `json:"data"`
	Props      map[string]interface{} `json:"props"`
	Operations map[string]interface{} `json:"operations"`

	UserIDs []string `json:"-"`
}

type State struct {
	AutotestSceneRequest  apistructs.AutotestSceneRequest `json:"autotestSceneRequest"`
	SceneId               uint64                          `json:"sceneId"`
	SetId                 uint64                          `json:"setId"`
	Total                 uint64                          `json:"total"`
	PageNo                uint64                          `json:"pageNo"`
	PageSize              uint64                          `json:"pageSize"`
	IsClick               bool                            `json:"isClick"`               // 点击目录树
	IsClickFolderTableRow bool                            `json:"isClickFolderTableRow"` // 点击场景列表的一行
	ClickFolderTableRowID uint64                          `json:"clickFolderTableRowID"` // 点击行的ID
}

type LatestStatus struct {
	RenderType string                 `json:"renderType"`
	Value      string                 `json:"value"`
	Status     apistructs.SceneStatus `json:"status"`
}

type Creator struct {
	RenderType string `json:"renderType"`
	Value      string `json:"value"`
}

type Data struct {
	ID           uint64       `json:"id"`
	CaseName     string       `json:"caseName"`
	StepCount    string       `json:"stepCount"`
	LatestStatus LatestStatus `json:"latestStatus"`
	Creator      Creator      `json:"creator"`
	CreatedAt    string       `json:"createdAt"`
	Operate      DataOperate  `json:"operate"`
}

type Meta struct {
	ID uint64 `json:"id"`
}

type DataOperation struct {
	Key     string `json:"key"`
	Text    string `json:"text"`
	Reload  bool   `json:"reload"`
	Confirm string `json:"confirm"`
	Meta    Meta   `json:"meta"`
}

type DataOperate struct {
	RenderType string                 `json:"renderType"`
	Operations map[string]interface{} `json:"operations"`
}

type Operation struct {
	Key    string `json:"key"`
	Reload bool   `json:"reload"`
}
type InParams struct {
	SpaceId uint64 `json:"spaceId"`
	SceneID string `json:"sceneId__urlQuery"`
}

type ClickRowOperation struct {
	Key      string    `json:"key"`
	Reload   bool      `json:"reload"`
	FillMeta string    `json:"fillMeta"`
	Meta     ClickMeta `json:"meta"`
}

type ClickMeta struct {
	RowData Data `json:"rowData"`
}
