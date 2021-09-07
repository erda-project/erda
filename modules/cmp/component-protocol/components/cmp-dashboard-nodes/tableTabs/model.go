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

package tableTabs

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

const (
	CPU_TAB    = "cpu-analysis"
	CPU_TAB_ZH = "cpu分析"

	MEM_TAB    = "mem-analysis"
	MEM_TAB_ZH = "mem分析"

	POD_TAB    = "pod-analysis"
	POD_TAB_ZH = "pod分析"
)

type TableTabs struct {
	base.DefaultProvider
	SDK        *cptype.SDK
	Type       string                 `json:"type"`
	Props      Props                  `json:"props"`
	State      State                  `json:"state"`
	Operations map[string]interface{} `json:"operations"`
}

type Props struct {
	TabMenu []MenuPair `json:"tabMenu"`
}

type MenuPair struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type State struct {
	ActiveKey string `json:"activeKey"`
}

type Meta struct {
	ActiveKey string `json:"activeKey"`
}

type Operation struct {
	Key      string `json:"key"`
	Reload   bool   `json:"reload"`
	FillMeta string `json:"fillMeta"`
	Meta     Meta   `json:"meta"`
}
