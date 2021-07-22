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

package addLabelModel

type AddLabelModel struct {
	Type      string                 `json:"type"`
	Props     map[string][]Fields `json:"props"`
	State     State                  `json:"state"`
	Operation map[string]interface{} `json:"operation"`
}

type State struct {
	Visible  bool        `json:"visible"`
	FormData interface{} `json:"form_data"`
}

type Fields struct {
	Key            string
	Component      string
	Label          string
	Required       bool
	ComponentProps map[string][]Options
	Rules          []map[string]string
}
type Options struct {
	Name string `json:"name"`
	Value string `json:"value"`
}

type Rule struct {
	Msg     string `json:"msg"`
	Pattern string `json:"pattern"`
}
