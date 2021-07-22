package cpuTable

import (
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard/common"
)

type MemInfoTable struct {
	common.Table
	Data [] common.RowItem
}
