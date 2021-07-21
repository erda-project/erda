package nodeDetail

import (
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard/common"
	v1 "k8s.io/api/core/v1"
)

type NodeDetail struct {
	CtxBdl     protocol.ContextBundle
	RenderType string            `json:"render_type"`
	NodeStatus NodeStatus        `json:"node_status"`
	NodeInfo   v1.NodeSystemInfo `json:"node_info"`
	State      common.State      `json:"state"`
}
type NodeStatus []common.SteveStatusEnum

