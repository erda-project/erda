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

package clickhouse

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
	tsql "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/es-tsql"
)

func (p *provider) Query(ctx context.Context, q tsql.Query) (*model.ResultSet, error) {
	searchSource := q.SearchSource()
	expr, ok := searchSource.(*goqu.SelectDataset)
	if !ok {
		return nil, errors.New("invalid search source")
	}
	if len(q.Sources()) <= 0 {
		return nil, errors.New("no source")
	}
	_, cluster := q.Sources()[0].Name, q.Sources()[0].Database

	table, _ := p.Loader.GetSearchTable(cluster)

	expr.From(table)

	sql, _, err := expr.ToSQL()

	if subLiters := q.SubSearchSource(); subLiters != nil {
		if tail, ok := subLiters.(map[string]string); ok {
			for _, v := range tail {
				sql += " " + v
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("invalid expresion to sql %s", err)
	}

	result := &model.ResultSet{}

	if q.Debug() {
		result.Details = sql
		fmt.Println(result.Details)
		return result, nil
	}

	rows, err := p.clickhouse.Client().Query(p.buildQueryContext(ctx), sql)
	if err != nil {
		return nil, err
	}

	result.Data, err = q.ParseResult(rows)
	return result, nil
}
func (p *provider) buildQueryContext(ctx context.Context) context.Context {
	return ctx
}