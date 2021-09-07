package memTable

import (
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-pods/common/table"
)

type MemTable struct {
	table.Table
	Data []table.RowItem
}
