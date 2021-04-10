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
	"fmt"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentFileTree struct {
	CtxBdl   protocol.ContextBundle
	Type     string                 `json:"type"`
	Props    map[string]interface{} `json:"props"`
	State    State                  `json:"state"`
	Data     []Data                 `json:"data"`
	Disabled bool
}

func (a *ComponentFileTree) SetBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil {
		err := fmt.Errorf("invalie bundle")
		return err
	}
	a.CtxBdl = b
	return nil
}

type State struct {
	ExpandedKeys         []string             `json:"expandedKeys"`
	SelectedKeys         []string             `json:"selectedKeys"`
	NodeFormModalAddNode NodeFormModalAddNode `json:"nodeFormModalAddNode"`
}

type Data struct {
	Key           string                 `json:"key"`
	Title         string                 `json:"title"`
	Icon          string                 `json:"icon"`
	IsLeaf        bool                   `json:"isLeaf"`
	ClickToExpand bool                   `json:"clickToExpand"` // always true, 该参数表示前端支持点击目录名进行折叠或展开; false 的话只有点击前面的小三角才能折叠或展开
	Selectable    bool                   `json:"selectable"`
	Operations    map[string]interface{} `json:"operations"`
	Children      []Data                 `json:"children"`
}

type ComponentNodeFormModal struct {
	Type       string                 `json:"type"`
	Props      map[string]interface{} `json:"props"`
	State      map[string]interface{} `json:"state"`
	Operations map[string]interface{} `json:"operations"`
}

type InParams struct {
	AppId        string `json:"appId"`
	ProjectId    string `json:"projectId"`
	SelectedKeys string `json:"selectedKeys"`
}
