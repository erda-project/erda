// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
