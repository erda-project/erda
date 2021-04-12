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

package table

type Props struct {
	PageSizeOptions []string               `json:"pageSizeOptions"`
	Columns         []Column               `json:"columns"`
	RowKey          string                 `json:"rowKey"`
	ClassName       string                 `json:"className"`
	Title           string                 `json:"title"`
	Visible         bool                   `json:"visible"`
	RowSelection    map[string]interface{} `json:"rowSelection"`
}

type Column struct {
	Title           string      `json:"title"`
	DataIndex       string      `json:"dataIndex"`
	TitleRenderType string      `json:"titleRenderType"`
	Width           int         `json:"width,omitempty"`
	TitleTip        []string    `json:"titleTip"`
	Data            interface{} `json:"data"`
}

type Data struct {
	List []RowData `json:"list"`
}

type RowData map[string]interface{}
