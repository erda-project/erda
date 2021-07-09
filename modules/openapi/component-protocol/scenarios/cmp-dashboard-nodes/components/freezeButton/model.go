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

package freezeButton

import protocol "github.com/erda-project/erda/modules/openapi/component-protocol"

type FreezeButton struct {
	ctxBdl     protocol.ContextBundle
	Type       string
	Props      Props
	Operations map[string]Operation `json:"operations,omitempty"`
}
type Props struct {
	Text    string
	Type    string
	Tooltip string
}
type Meta struct {
	NodeName    string `json:"node_name"`
	ClusterName string `json:"cluster_name"`
}
type Operation struct {
	Key           string      `json:"key"`
	Reload        bool        `json:"reload"`
	FillMeta      string      `json:"fillMeta"`
	Meta          interface{} `json:"meta"`
	ClickableKeys interface{} `json:"clickableKeys"`
	Command       Command     `json:"command,omitempty"`
}
type Command struct {
	Key    string `json:"key"`
	Target string `json:"target"`
}
