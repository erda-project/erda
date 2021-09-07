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

import "github.com/erda-project/erda/modules/openapi/component-protocol/components/base"

const (
	CPU_TAB    = "cpu"
	CPU_TAB_ZH = "cpu分析"

	MEM_TAB    = "mem"
	MEM_TAB_ZH = "mem分析"

	POD_TAB    = "pod"
	POD_TAB_ZH = "pod分析"
)

type TableTabs struct {
	base.DefaultProvider
	Type       string     `json:"type"`
	Props      Props      `json:"props"`
	Operations Operations `json:"operations"`
	State      State      `json:"state"`
}

type Props struct {
	TabMenu []TabMenu `json:"tabMenu"`
}

type Operations struct {
	OnChange OnChange `json:"onChange"`
}

type State struct {
	ActiveKey string `json:"activeKey"`
}

type TabMenu struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type OnChange struct {
	Key    string `json:"key"`
	Reload bool   `json:"reload"`
}
