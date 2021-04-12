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

package fileTree

import (
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentFileTree struct {
	CtxBdl protocol.ContextBundle

	CompName   string                 `json:"name"`
	Data       []SceneSet             `json:"data"`
	Operations map[string]interface{} `json:"operations"`
	State      State                  `json:"state"`
	Props      map[string]interface{} `json:"props"`
}

type SceneSet struct {
	Key           string                 `json:"key"`
	Title         string                 `json:"title"`
	Icon          string                 `json:"icon"`
	IsColorIcon   bool                   `json:"isColorIcon"`
	IsLeaf        bool                   `json:"isLeaf"`
	ClickToExpand bool                   `json:"clickToExpand"`
	Selectable    bool                   `json:"selectable"`
	Operations    map[string]interface{} `json:"operations"`
	Children      []Scene                `json:"children"`
	Type          string                 `json:"type"`
}

type Scene struct {
	Key        string                 `json:"key"`
	Title      string                 `json:"title"`
	Icon       string                 `json:"icon"`
	IsLeaf     bool                   `json:"isLeaf"`
	Operations map[string]interface{} `json:"operations"`
	Type       string                 `json:"type"`
}

type OperationBase struct {
	Key    string `json:"key"`
	Text   string `json:"text"`
	Reload bool   `json:"reload"`
}

type AddSceneOperation struct {
	OperationBase
}

type InParams struct {
	SpaceId      uint64 `json:"spaceId"`
	SelectedKeys string `json:"selectedKeys"`
	SceneID      string `json:"sceneId__urlQuery"`
	SetID        string `json:"setId__urlQuery"`
	ProjectId    uint64 `json:"projectId"`
}

type DragParams struct {
	DragKey  string `json:"dragKey"`
	DropKey  string `json:"dropKey"`
	Position int64  `json:"position"`
	DragType string `json:"dragType"`
	DropType string `json:"dropType"`
}

type State struct {
	ExpandedKeys            []string   `json:"expandedKeys"`
	SelectedKeys            []string   `json:"selectedKeys"`
	FormVisible             bool       `json:"formVisible"`
	SceneSetKey             int        `json:"sceneSetKey"`
	ActionType              string     `json:"actionType"`
	SetId__urlQuery         string     `json:"setId__urlQuery"`
	SceneId__urlQuery       string     `json:"sceneId__urlQuery"`
	SceneId                 uint64     `json:"sceneId"`
	DragParams              DragParams `json:"dragParams"`
	PageNo                  int        `json:"pageNo"`
	IsClickScene            bool       `json:"isClickScene"`
	IsClickFolderTable      bool       `json:"isClickFolderTable"`
	ClickFolderTableSceneID uint64     `json:"clickFolderTableSceneID"`
}

type Operation struct {
	Key    string `json:"key"`
	Reload bool   `json:"reload"`
}
