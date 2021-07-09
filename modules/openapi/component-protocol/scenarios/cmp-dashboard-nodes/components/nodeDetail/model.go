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

package nodeDetail

import (
	v1 "k8s.io/api/core/v1"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common"
)

type NodeDetail struct {
	CtxBdl     protocol.ContextBundle
	RenderType string            `json:"render_type"`
	NodeStatus NodeStatus        `json:"node_status"`
	NodeInfo   v1.NodeSystemInfo `json:"node_info"`
	State      common.State      `json:"state"`
}
type NodeStatus []common.SteveStatusEnum
