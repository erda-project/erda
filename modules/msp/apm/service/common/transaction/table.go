package transaction

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/components/table"
	"github.com/erda-project/erda-infra/providers/i18n"
)

const (
	columnTransactionName table.ColumnKey = "transactionName"
	columnReqCount        table.ColumnKey = "reqCount"
	columnErrorCount      table.ColumnKey = "errorCount"
	columnSlowCount       table.ColumnKey = "slowCount"
	columnAvgDuration     table.ColumnKey = "avgDuration"
)

func InitTable(lang i18n.LanguageCodes, i18n i18n.Translator) table.Table {
	return table.Table{
		Columns: table.ColumnsInfo{
			Orders: []table.ColumnKey{columnTransactionName, columnReqCount, columnErrorCount, columnSlowCount, columnAvgDuration},
			ColumnsMap: map[table.ColumnKey]table.Column{
				columnTransactionName: {Title: i18n.Text(lang, string(columnTransactionName))},
				columnReqCount:        {Title: i18n.Text(lang, string(columnReqCount))},
				columnErrorCount:      {Title: i18n.Text(lang, string(columnErrorCount))},
				columnSlowCount:       {Title: i18n.Text(lang, string(columnSlowCount))},
				columnAvgDuration:     {Title: i18n.Text(lang, string(columnAvgDuration))},
			},
		},
	}
}
