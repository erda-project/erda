package cpuChart

import (
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard/common"
)

type CpuChart struct {
	CtxBdl     protocol.ContextBundle
	State common.State `json:"state"`
	Data []common.ChartDataItem `json:"data"`
}


