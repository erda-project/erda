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

import "github.com/erda-project/erda/apistructs"

type AddNodeOperation struct {
	Key      string                  `json:"key"`
	Text     string                  `json:"text"`
	Reload   bool                    `json:"reload"`
	Command  AddNodeOperationCommand `json:"command"`
	Disabled bool                    `json:"disabled"`
}

type AddNodeOperationCommand struct {
	Key    string                       `json:"key"`
	Target string                       `json:"target"`
	State  AddNodeOperationCommandState `json:"state"`
}

type AddNodeOperationCommandState struct {
	Visible  bool                                 `json:"visible"`
	FormData AddNodeOperationCommandStateFormData `json:"formData"`
	// NodeFormModal 传递的数据
	NodeFormModalAddNode NodeFormModalAddNode `json:"nodeFormModalAddNode"`
}

type ClickBranchNodeOperation struct {
	Key    string                       `json:"key"`
	Text   string                       `json:"text"`
	Reload bool                         `json:"reload"`
	Show   bool                         `json:"show"` // 是否在菜单里展示
	Meta   ClickBranchNodeOperationMeta `json:"meta"`
}

type ClickBranchNodeOperationMeta struct {
	ParentKey string `json:"parentKey,omitempty"`
}

type NodeFormModalAddNode struct {
	Results apistructs.UnifiedFileTreeNode `json:"nodeFormModalAddNode"`
	Branch  string                         `json:"branch"`
}

type AddNodeOperationCommandStateFormData struct {
	Branch    string `json:"branch"`
	Name      string `json:"name"`
	AddResult apistructs.UnifiedFileTreeNode
}

type DeleteOperation struct {
	Key         string              `json:"key"`
	Text        string              `json:"text"`
	Confirm     string              `json:"confirm"`
	Reload      bool                `json:"reload"`
	Disabled    bool                `json:"disabled"`
	DisabledTip string              `json:"disabledTip"`
	Meta        DeleteOperationData `json:"meta"`
}

type DeleteOperationData struct {
	Key string `json:"key"`
}

type AddDefaultOperations struct {
	Key      string                  `json:"key"`
	Text     string                  `json:"text"`
	Reload   bool                    `json:"reload"`
	Show     bool                    `json:"show"`
	Meta     AddDefaultOperationData `json:"meta"`
	Disabled bool                    `json:"disabled"`
}

type AddDefaultOperationData struct {
	Key string `json:"key"`
}
