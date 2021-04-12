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

package panel

type Props struct {
	Visible        bool                   `json:"visible"`
	Fields         []Field                `json:"fields"`
	Column         int                    `json:"column"`
	Colon          bool                   `json:"colon"`
	ColumnNum      int                    `json:"columnNum"`
	IsMultiColumn  bool                   `json:"isMultiColumn"`
	Layout         PropsLayout            `json:"layout"` // 'vertical' | 'horizontal'
	Data           map[string]interface{} `json:"data"`
	Type           PropsType              `json:"type"` // 'Z' | 'N'
	NumOfRowsLimit int                    `json:"numOfRowsLimit"`
}

type Field struct {
	Label      string      `json:"label"`
	ValueKey   interface{} `json:"valueKey"`
	RenderType string      `json:"renderType"` // ellipsis
}

type PropsLayout string

const (
	PropsLayoutVertical   PropsLayout = "vertical"
	PropsLayoutHorizontal PropsLayout = "horizontal"
)

type PropsType string

const (
	PropsTypeZ PropsType = "Z"
	PropsTypeN PropsType = "N"
)

type Data struct {
	Data map[string]interface{} `json:"data"`
}
