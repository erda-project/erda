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

import (
	"github.com/erda-project/erda/apistructs"
)

type CommonTable struct {
	State      State                                            `json:"state"`
	Type       string                                           `json:"type"`
	Operations map[apistructs.OperationKey]apistructs.Operation `json:"operations"`
	Props      Props                                            `json:"props"`
	Data       Data                                             `json:"data"`
}

type Data struct {
	List []NotifyTableList `json:"list"`
}

type NotifyTableList struct {
	Id        int64   `json:"id"`
	Name      string  `json:"name"`
	Targets   Target  `json:"targets"`
	CreatedAt string  `json:"createdAt"`
	Operate   Operate `json:"operate"`
	Enable    bool    `json:"enable"`
}

type Target struct {
	Value      []apistructs.Value `json:"value"`
	RoleMap    map[string]string  `json:"roleMap"`
	RenderType string             `json:"renderType"`
}

type Operate struct {
	RenderType string                `json:"renderType"`
	Value      string                `json:"value"`
	Operations map[string]Operations `json:"operations"`
}

type Operations struct {
	Key     string `json:"key"`
	Text    string `json:"text"`
	Reload  bool   `json:"reload"`
	Confirm string `json:"confirm"`
	Meta    Meta   `json:"meta"`
}

type Meta struct {
	Id int64 `json:"id"`
}

type FormData struct {
	Id       int64    `json:"id"`
	Name     string   `json:"name"`
	Items    []string `json:"items"`
	Target   string   `json:"target"`
	Channels []string `json:"channels"`
}

type State struct {
	EditId    uint64 `json:"editId"`
	Operation string `json:"operation"`
	Visible   bool   `json:"visible"`
}

type Props struct {
	RowKey     string       `json:"rowKey"`
	Columns    []PropColumn `json:"columns"`
	Pagination bool         `json:"pagination"`
}

type PropColumn struct {
	Title     string `json:"title"`
	DataIndex string `json:"dataIndex"`
	Width     uint64 `json:"width"`
}
