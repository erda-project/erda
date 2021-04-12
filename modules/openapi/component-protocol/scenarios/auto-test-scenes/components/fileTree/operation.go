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

type SceneSetOperation struct {
	Key    string                `json:"key"`
	Text   string                `json:"text"`
	Reload bool                  `json:"reload"`
	Show   bool                  `json:"show"`
	Meta   SceneSetOperationMeta `json:"meta"`
}

type SceneSetOperationMeta struct {
	ParentKey int `json:"parentKey,omitempty"`
	Key       int `json:"key,omitempty"`
}

// type ClickSceneSetOperation struct {
// 	BaseOperation
// 	Meta SceneSetOperationMeta `json:"meta"`
// }

type DeleteOperation struct {
	Key     string                `json:"key"`
	Text    string                `json:"text"`
	Confirm string                `json:"confirm"`
	Reload  bool                  `json:"reload"`
	Show    bool                  `json:"show"`
	Meta    SceneSetOperationMeta `json:"meta"`
}

type DeleteOperationData struct {
	Key string `json:"key"`
}
