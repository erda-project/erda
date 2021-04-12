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

package fileInfo

import (
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentFileInfo struct {
	ctxBdl protocol.ContextBundle

	CommonFileInfo
}

type CommonFileInfo struct {
	Version    string                                           `json:"version,omitempty"`
	Name       string                                           `json:"name,omitempty"`
	Type       string                                           `json:"type,omitempty"`
	Props      map[string]interface{}                           `json:"props,omitempty"`
	State      State                                            `json:"state,omitempty"`
	Operations map[apistructs.OperationKey]apistructs.Operation `json:"operations,omitempty"`
	Data       Data                                             `json:"data,omitempty"`
	InParams   InParams                                         `json:"inParams,omitempty"`
}

type InParams struct {
}

type Data struct {
	Name           string `json:"name"`
	Description    string `json:"description"`         // 描述
	CreatorID      string `json:"creatorID,omitempty"` // 创建者
	UpdaterID      string `json:"updaterID,omitempty"` // 更新者
	CreateATString string `json:"createAtString"`
	UpdateATString string `json:"updateAtString"`
}

type PropColumn struct {
	Label    string `json:"label"`
	ValueKey string `json:"valueKey"`
}

type State struct {
	TestPlanId uint64 `json:"testPlanId"`
	Visible    bool   `json:"visible"`
}
