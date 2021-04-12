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

package notifyConfigTable

import (
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/table"
)

var (
	roleMap = map[string]string{
		"Owner": "应用所有者",
		"Lead":  "应用主管",
		"Ops":   "运维",
		"Dev":   "开发工程师",
		"QA":    "测试工程师",
	}
)

type Notify struct {
	CtxBdl protocol.ContextBundle
	table.CommonTable
}

type DeleteNotifyOperation struct {
	Meta DeleteOperationData `json:"meta"`
}

type DeleteOperationData struct {
	Id uint64 `json:"id"`
}

type EditOperation struct {
	Key    string `json:"key"`
	Text   string `json:"text"`
	Reload bool   `json:"reload"`
	Meta   Meta   `json:"meta"`
}

type Meta struct {
	Id uint64 `json:"id"`
}

type InParams struct {
	ScopeType string `json:"scopeType"`
	ScopeId   string `json:"scopeId"`
}

func (n *Notify) genProps() {
	props := table.Props{
		RowKey: "id",
		Columns: []table.PropColumn{
			{
				Title:     "通知名称",
				DataIndex: "name",
			},
			{
				Title:     "通知对象",
				DataIndex: "targets",
				Width:     200,
			},
			{
				Title:     "创建时间",
				DataIndex: "createdAt",
			},
			{
				Title:     "操作",
				DataIndex: "operate",
				Width:     150,
			},
		},
	}
	n.Props = props
}
