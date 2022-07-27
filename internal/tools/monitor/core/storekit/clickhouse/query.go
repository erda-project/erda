package clickhouse

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/doug-martin/goqu/v9"
)

type Query interface {
	QueryRaw(orgName string, expr *goqu.SelectDataset) (driver.Rows, error)
}
