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
	"strings"

	cksdk "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
	tsql "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/es-tsql"
	"github.com/erda-project/erda/pkg/common/trace"
)

func (p *provider) Query(ctx context.Context, q tsql.Query) (*model.ResultSet, error) {
	newCtx, span := otel.Tracer("executive").Start(ctx, "metric.clickhouse")
	defer span.End()

	searchSource := q.SearchSource()
	expr, ok := searchSource.(*goqu.SelectDataset)
	if !ok || expr == nil {
		return nil, errors.New("invalid search source")
	}
	if len(q.Sources()) <= 0 {
		return nil, errors.New("no source")
	}
	table, _ := p.Loader.GetSearchTable(q.OrgName())

	if len(q.OrgName()) > 0 {
		expr = expr.Where(goqu.C("org_name").Eq(q.OrgName()))
	}
	if len(q.TerminusKey()) > 0 {
		expr = expr.Where(goqu.C("tenant_id").Eq(q.TerminusKey()))
	}

	span.SetAttributes(attribute.String("org_name", q.OrgName()))
	span.SetAttributes(attribute.String("tenant_id", q.TerminusKey()))
	span.SetAttributes(attribute.String("table", table))

	var metrics []string
	for _, s := range q.Sources() {
		metrics = append(metrics, s.Name)
	}

	span.SetAttributes(attribute.String("metrics", strings.Join(metrics, ",")))

	expr = expr.Where(goqu.C("metric_group").In(metrics))

	expr = expr.From(table)

	sql, _, err := expr.ToSQL()

	if err != nil {
		return nil, errors.Wrap(err, "failed to generate SQL")
	}

	// add tail liters
	var sb strings.Builder
	if subLiters := q.SubSearchSource(); subLiters != nil {
		if tail, ok := subLiters.(map[string]string); ok {
			////"\"%s\" ASC"
			//sb.WriteString()
			for _, v := range tail {
				sb.WriteString(v)
			}
		}
	}

	sql = strings.ReplaceAll(sql, "\"%s\" ASC", sb.String())

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("invalid expresion to sql: %s", sql))
	}

	result := &model.ResultSet{}

	span.SetAttributes(attribute.String("sql", sql))

	if q.Debug() {
		result.Details = sql
		fmt.Println(result.Details)
		return result, nil
	}

	rows, err := p.clickhouse.Client().Query(p.buildQueryContext(newCtx), sql)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to query: %s", sql))
	}
	if rows.Err() != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to query: %s", sql))
	}
	defer rows.Close()

	result.Data, err = q.ParseResult(newCtx, rows)

	if result.Data != nil && result.Data.Rows != nil {
		span.SetAttributes(attribute.Int("result_total", len(result.Data.Rows)))
	}
	span.SetAttributes(trace.BigStringAttribute("result", result))

	if err != nil {
		p.Log.Error("clickhouse metric query is error ", sql, err)
		return nil, err
	}
	return result, nil
}
func (p *provider) buildQueryContext(ctx context.Context) context.Context {
	settings := map[string]interface{}{}
	if p.Cfg.QueryTimeout > 0 {
		settings["max_execution_time"] = int(p.Cfg.QueryTimeout.Seconds()) + 5
	}
	if p.Cfg.QueryMaxThreads > 0 {
		settings["max_threads"] = p.Cfg.QueryMaxThreads
	}
	if p.Cfg.QueryMaxMemory > 0 {
		settings["max_memory_usage"] = p.Cfg.QueryMaxMemory
	}
	if len(settings) == 0 {
		return ctx
	}

	newCtx, span := otel.Tracer("clickhouse").Start(ctx, "sdk.clickhouse")
	defer span.End()

	ctx = cksdk.Context(newCtx, cksdk.WithSettings(settings), cksdk.WithSpan(span.SpanContext()))
	return ctx
}

func (p *provider) QueryRaw(orgName string, expr *goqu.SelectDataset) (driver.Rows, error) {
	table, _ := p.Loader.GetSearchTable(orgName)
	expr = expr.From(table)
	sql, _, err := expr.ToSQL()
	if err != nil {
		return nil, err
	}
	return p.clickhouse.Client().Query(context.Background(), sql)
}
