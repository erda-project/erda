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

package file_tree

type Spec struct {
	Type  string `json:"type"`
	Props Props  `json:"props"`
}

type Props struct {
	Searchable bool `json:"searchable"`
	Draggable  bool `json:"draggable"`
}

type Data struct {
	TreeData []INode `json:"treeData"`
}

type INode struct {
	Key           string                 `json:"key"`
	Title         string                 `json:"title"`
	Icon          string                 `json:"icon"`
	IsColorIcon   bool                   `json:"isColorIcon"`
	Children      []INode                `json:"children"`
	Selectable    bool                   `json:"selectable"`
	ClickToExpand bool                   `json:"clickToExpand"`
	IsLeaf        bool                   `json:"isLeaf"`
	Operations    map[string]interface{} `json:"operations"`
	Type          string                 `json:"type"`
}

type Field struct {
	Label    string      `json:"label"`
	ValueKey interface{} `json:"valueKey"`
}
