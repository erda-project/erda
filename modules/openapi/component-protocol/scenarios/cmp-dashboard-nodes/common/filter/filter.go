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

package filter

var PropsInstance = Props{
	LabelWidth: 40,
	Fields: []Field{
		{
			Label: "env",
			Key:   "env",
			Type:  "select",
			Options: []Option{
				{Label: "开发环境", Value: "dev"},
				{Label: "测试环境", Value: "test"},
				{Label: "预发环境", Value: "staging"},
				{Label: "生产环境", Value: "prod"},
			},
		},
		{
			Key:   "service",
			Label: "服务",
			Type:  "select",
			Options: []Option{
				{Label: "有状态服务", Value: "stateful"},
				{Label: "无状态服务", Value: "stateless"},
			},
		},
		{
			Key:   "packjob",
			Label: "构建",
			Type:  "select",
			Options: []Option{
				{Label: "pack-job", Value: "packJob"},
			},
		},
		{
			Key:   "other",
			Label: "其他",
			Type:  "select",
			Options: []Option{
				{Label: "集群服务", Value: "cluster-service"},
				{Label: "自定义", Value: "custom"},
				{Label: "独占", Value: "mono"},
				{Label: "锁住节点", Value: "cordon"},
				{Label: "驱逐节点", Value: "drain"},
				{Label: "平台组件", Value: "platform"},
			},
		},

	},
}

type Filter struct {
	Type       string                     `json:"type"`
	Operations map[string]FilterOperation `json:"operations"`
	State      State                      `json:"state"`
	Props      Props                      `json:"props"`
}

type State struct {
	Values Values `json:"values"`
}

type Props struct {
	LabelWidth int     `json:"label_width"`
	Fields     []Field `json:"fields"`
}

type FilterOperation struct {
	Key    string `json:"key"`
	Reload bool   `json:"reload"`
}

type Values struct {
	keys map[string][]string
}

type Field struct {
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Options     []Option `json:"options"`
	Key         string   `json:"key"`
	Placeholder string   `json:"placeholder"`
}

type Option struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
