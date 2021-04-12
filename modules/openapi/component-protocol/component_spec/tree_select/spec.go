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

package tree_select

type INode struct {
	Key        string `json:"key"`
	Id         string `json:"id"`
	PId        string `json:"pId"`
	Title      string `json:"title"`
	IsLeaf     bool   `json:"isLeaf"`
	Value      string `json:"value"`
	Selectable bool   `json:"selectable"`
	Disabled   bool   `json:"disabled"`
}

type Data struct {
	TreeData []INode `json:"treeData"`
}

type Props struct {
	Visible     bool   `json:"visible"`
	Placeholder string `json:"placeholder"`
	Title       string `json:"title"`
}
