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

package apiEditor

type State struct {
	// 小试一把
	AttemptTest            AttemptTestAll `json:"attemptTest"`
	ConfigManageNamespaces string         `json:"configManageNamespaces"`

	// 步骤信息
	StepId    uint64 `json:"stepId"`
	SceneId   uint64 `json:"sceneId"`
	IsFirstIn bool   `json:"isFirstIn"` // 是否从步骤中点击进入
}

type AttemptTestAll struct {
	Visible bool                   `json:"visible"`
	Status  string                 `json:"status"`
	Data    map[string]interface{} `json:"data"`
}

type Menu struct {
	Text       string                 `json:"text"`
	Key        string                 `json:"key"`
	Operations map[string]interface{} `json:"operations"`
}

type ClickOperation struct {
	Key    string      `json:"key"`
	Reload bool        `json:"reload"`
	Meta   interface{} `json:"meta"`
}

type Meta struct {
	Env       string `json:"env"`
	ScenesID  uint64 `json:"scenesID"`
	ConfigEnv string `json:"configEnv"`
}

type OpMetaInfo struct {
	Meta Meta `json:"meta"`
}
