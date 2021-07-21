package memChart

import (
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard/common"
)

type MemChart struct {
	CtxBdl protocol.ContextBundle
	State  common.State    `json:"state"`
	Data   []common.ChartDataItem `json:"data"`
}

