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

package podDistribution

import (
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-pods/common/table"
)

type PodDistribution struct {
	CtxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Data   map[string]interface{} `json:"data"`
	State  table.State `json:"state"`
}

type Data struct {
	Value int    `json:"value"`
	Label string `json:"label"`
	Color string `json:"color"`
	Tip   string `json:"tip"`
}
