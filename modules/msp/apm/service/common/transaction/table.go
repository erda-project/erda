package transaction

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/components/table"
	"github.com/erda-project/erda-infra/providers/i18n"
)

const (
	ColumnTransactionName table.ColumnKey = "transactionName"
	ColumnReqCount        table.ColumnKey = "reqCount"
	ColumnErrorCount      table.ColumnKey = "errorCount"
	ColumnSlowCount       table.ColumnKey = "slowCount"
	ColumnAvgDuration     table.ColumnKey = "avgDuration"
)

func InitTable(lang i18n.LanguageCodes, i18n i18n.Translator) table.Table {
	return table.Table{
		Columns: table.ColumnsInfo{
			Orders: []table.ColumnKey{ColumnTransactionName, ColumnReqCount, ColumnErrorCount, ColumnSlowCount, ColumnAvgDuration},
			ColumnsMap: map[table.ColumnKey]table.Column{
				ColumnTransactionName: {Title: i18n.Text(lang, string(ColumnTransactionName))},
				ColumnReqCount:        {Title: i18n.Text(lang, string(ColumnReqCount))},
				ColumnErrorCount:      {Title: i18n.Text(lang, string(ColumnErrorCount))},
				ColumnSlowCount:       {Title: i18n.Text(lang, string(ColumnSlowCount))},
				ColumnAvgDuration:     {Title: i18n.Text(lang, string(ColumnAvgDuration))},
			},
		},
	}
}
