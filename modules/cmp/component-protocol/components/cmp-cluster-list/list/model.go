// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package list

import (
	"context"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type List struct {
	base.DefaultProvider
	SDK   *cptype.SDK
	Bdl   *bundle.Bundle
	Ctx   context.Context
	Data  map[string][]DataItem `json:"data"`
	State State                 `json:"state"`
}

type State struct {
	PageNo bool `json:"pageNo"`
}

type DataItem struct {
	ID            int                  `json:"id"`
	Title         string               `json:"title"`
	Description   string               `json:"description"`
	PrefixImg     string               `json:"prefixImg"`
	BackgroundImg string               `json:"backgroundImg"`
	ExtraInfos    []ExtraInfos         `json:"extraInfos"`
	Status        ItemStatus           `json:"status"`
	ExtraContent  ExtraContent         `json:"extraContent"`
	Operations    map[string]Operation `json:"operations"`
}

type ItemStatus struct {
	Text   string `json:"text"`
	Status string `json:"status"`
}

type ExtraContent struct {
	Type      string      `json:"type"`
	RowNum    int         `json:"rowNum"`
	ExtraData []ExtraData `json:"data"`
}

type ExtraData struct {
	Name  string          `json:"name"`
	Value float64         `json:"value"`
	Total int             `json:"total"`
	Color string          `json:"color"`
	Info  []ExtraDataItem `json:"info"`
}

type ExtraDataItem struct {
	Main string `json:"main"`
	Sub  string `json:"sub"`
}
type ExtraInfos struct {
	Icon    string `json:"icon"`
	Text    string `json:"text"`
	Tooltip string `json:"tooltip"`
}

type Operation struct {
	Key    string      `json:"key"`
	Reload bool        `json:"reload"`
	Show   bool        `json:"show"`
	Meta   interface{} `json:"meta"`
	Text   string      `json:"text"`
}

type Command struct {
	Key     string       `json:"key,omitempty"`
	Command CommandState `json:"state,omitempty"`
	Target  string       `json:"target,omitempty"`
	JumpOut bool         `json:"jumpOut,omitempty"`
}

type CommandState struct {
	Params   Params                 `json:"params,omitempty"`
	Visible  bool                   `json:"visible,omitempty"`
	FormData FormData               `json:"formData,omitempty"`
	Query    map[string]interface{} `json:"query,omitempty"`
}
type Params struct {
	NodeId string `json:"nodeId,omitempty"`
	NodeIP string `json:"nodeIP,omitempty"`
}

type FormData struct {
	RecordId string `json:"recordId,omitempty"`
}

type ResData struct {
	CpuUsed     float64
	CpuTotal    float64
	MemoryUsed  float64
	MemoryTotal float64
	DiskUsed    float64
	DiskTotal   float64
}

type ClusterInfoDetail struct {
	Version     string
	NodeCnt     int
	ClusterType string
	Management  string
	CreateTime  string
	Status      string
	RawStatus   string // "pending","online","offline" ,"initializing","initialize error","unknown"
}
