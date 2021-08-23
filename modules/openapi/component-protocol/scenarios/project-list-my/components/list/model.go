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
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentList struct {
	ctxBdl protocol.ContextBundle

	CommonList
}

type CommonList struct {
	Version    string                 `json:"version,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Type       string                 `json:"type,omitempty"`
	Props      Props                  `json:"props,omitempty"`
	State      State                  `json:"state,omitempty"`
	Operations map[string]interface{} `json:"operations,omitempty"`
	Data       Data                   `json:"data,omitempty"`
}

type Props struct {
	PageSizeOptions []string `json:"pageSizeOptions"`
	Visible         bool     `json:"visible"`
}

type Data struct {
	List []ProItem `json:"list"`
}

type State struct {
	PageNo        uint64                 `json:"pageNo"`
	PageSize      uint64                 `json:"pageSize"`
	Total         uint64                 `json:"total"`
	Query         map[string]interface{} `json:"query"` // 搜索
	IsFirstFilter bool                   `json:"isFirstFilter"`
	IsEmpty       bool                   `json:"isEmpty"` // 数据是否为空
}

type ProItem struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	PrefixImg   string                 `json:"prefixImg"`
	ExtraInfos  []ExtraInfos           `json:"extraInfos"`
	Operations  map[string]interface{} `json:"operations"`
	ProjectId   uint64                 `json:"projectId"`
}

type ExtraInfos struct {
	Icon    string `json:"icon,omitempty"`
	Text    string `json:"text,omitempty"`
	Tooltip string `json:"tooltip,omitempty"`
	Type    string `json:"type,omitempty"`
}

type Command struct {
	Key    string `json:"key"`
	Target string `json:"target"`
}

type PageSizeNo struct {
	PageSize uint64 `json:"pageSize"`
	PageNo   uint64 `json:"pageNo"`
}

type Meta struct {
	ID          uint64     `json:"id,omitempty"`
	PageSize    PageSizeNo `json:"pageSize,omitempty"`
	PageNo      PageSizeNo `json:"pageNo,omitempty"`
	ProjectId   uint64     `json:"projectId,omitempty"`
	ProjectName string     `json:"projectName,omitempty"`
}

type Operation struct {
	Key      string  `json:"key,omitempty"`
	Reload   bool    `json:"reload"`
	FillMeta string  `json:"fillMeta,omitempty"`
	Text     string  `json:"text,omitempty"`
	Show     bool    `json:"show,omitempty"`
	Command  Command `json:"command,omitempty"`
	Confirm  string  `json:"confirm,omitempty"`
	Meta     Meta    `json:"meta,omitempty"`
}
