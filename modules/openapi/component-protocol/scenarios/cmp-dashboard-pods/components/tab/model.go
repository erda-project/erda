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

package tab

const (
	CPU_TAB    = "cpu"
	CPU_TAB_ZH = "cpu分析"

	MEM_TAB    = "mem"
	MEM_TAB_ZH = "mem分析"
)

type SteveTab struct {
	Type       string                 `json:"type"`
	Props      Props                  `json:"props"`
	State      State                  `json:"state"`
	Operations map[string]interface{} `json:"operations"`
}

type Props struct {
	TabMenu []MenuPair `json:"tab_menu"`
}

type MenuPair struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type State struct {
	ActiveKey string `json:"active_key"`
}

type Meta struct {
	ActiveKey string `json:"active_key"`
}

type Operation struct {
	Key      string `json:"key"`
	Reload   bool   `json:"reload"`
	FillMeta string `json:"fillMeta"`
	Meta     Meta   `json:"meta"`
}
