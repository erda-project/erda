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

package edit_list

type Props struct {
	Visible bool   `json:"visible"`
	Temp    []Temp `json:"temp"`
}

type Temp struct {
	Title  string `json:"title"`
	Key    string `json:"key"`
	Width  int    `json:"width"`
	Flex   int    `json:"flex"`
	Render Render `json:"render"`
}

type Render struct {
	Type             string                 `json:"type,omitempty"` // 'input' | 'text' | 'select' | 'inputSelect'
	ValueConvertType string                 `json:"valueConvertType,omitempty"`
	Options          []PropChangeOption     `json:"options,omitempty"`
	Required         bool                   `json:"required,omitempty"`
	UniqueValue      bool                   `json:"uniqueValue,omitempty"`
	Operations       map[string]interface{} `json:"operations,omitempty"`
	Rules            []PropRenderRule       `json:"rules,omitempty"`
	Props            PropRenderProp         `json:"props,omitempty"`
	DefaultValue     interface{}            `json:"defaultValue,omitempty"`
}

type PropRenderProp struct {
	MaxLength   int64              `json:"maxLength,omitempty"`
	Placeholder string             `json:"placeholder,omitempty"`
	Options     []PropChangeOption `json:"options,omitempty"`
}

type PropRenderRule struct {
	Pattern string `json:"pattern,omitempty"`
	Msg     string `json:"msg,omitempty"`
}

type PropChangeOption struct {
	Label    string             `json:"label"`
	Value    string             `json:"value"`
	IsLeaf   bool               `json:"isLeaf"`
	Children []PropChangeOption `json:"children"`
}

func (pct *PropChangeOption) FindValue(v string) *PropChangeOption {
	if pct.Value == v {
		return pct
	}
	for i := range pct.Children {
		k := pct.Children[i].FindValue(v)
		if k != nil {
			return k
		}
	}
	return nil
}
