package cpuTable

import (
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard/common"
)

type CpuInfoTable struct {
	common.Table
	Data []common.RowItem
}

