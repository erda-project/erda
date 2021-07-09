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

	POD_TAB    = "pod"
	POD_TAB_ZH = "pod分析"
)

type SteveTab struct {
	Type       string     `json:"type"`
	Props      Props      `json:"props,omitempty"`
	State      PropsState `json:"state,omitempty"`
	Operations map[string]interface{}
}
type Props struct {
	TabMenu []MenuPair `json:"tab_menu,omitempty"`
}
type MenuPair struct {
	key  string
	name string
}
type PropsState struct {
	ActiveKey string `json:"active_key"`
}
