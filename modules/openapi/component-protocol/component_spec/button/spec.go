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

package button

type Props struct {
	Text       string                 `json:"text"`
	Type       interface{}            `json:"type"`
	Ghost      bool                   `json:"ghost"`
	Menu       []MenuItem             `json:"menu"`
	PrefixIcon string                 `json:"prefixIcon"`
	Style      map[string]interface{} `json:"style"`
	SuffixIcon string                 `json:"suffixIcon"`
	Tooltip    string                 `json:"tooltip"`
	Visible    bool                   `json:"visible"`
}

type MenuItem struct {
	Key         string                 `json:"key"`
	Text        string                 `json:"text"`
	Disabled    bool                   `json:"disabled"`
	DisabledTip string                 `json:"disabledTip"`
	PrefixIcon  string                 `json:"prefixIcon"`
	Confirm     string                 `json:"confirm"`
	Operations  map[string]interface{} `json:"operations"`
}
