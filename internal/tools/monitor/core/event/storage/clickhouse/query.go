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

	"github.com/doug-martin/goqu/v9"

	"github.com/erda-project/erda/internal/tools/monitor/core/event"
	"github.com/erda-project/erda/internal/tools/monitor/core/event/storage"
)

func (p *provider) QueryPaged(ctx context.Context, sel *storage.Selector, pageNo, pageSize int) ([]*event.Event, error) {
	table, _ := p.Loader.GetSearchTable("")
	sql := goqu.From(table).Where(
		goqu.C("timestamp").Gte(goqu.L("fromUnixTimestamp64Nano(cast(?,'Int64'))", sel.Start)),
		goqu.C("timestamp").Lt(goqu.L("fromUnixTimestamp64Nano(cast(?,'Int64'))", sel.End)),
	)

	for _, filter := range sel.Filters {
		val, ok := filter.Value.(string)
		if !ok {
			continue
		}
		if len(val) == 0 {
			continue
		}

		switch filter.Op {
		case storage.EQ:
			splits := strings.Split(filter.Key, ".")
			if len(splits) > 1 && splits[0] == "tags" {
				filter.Key = fmt.Sprintf("%s['%s']", splits[0], splits[1])
			}
			sql = sql.Where(goqu.L(filter.Key).Eq(val))
		case storage.REGEXP:
			sql = sql.Where(goqu.L("match(?,?)", filter.Key, val))
		}
	}

	sql = sql.Offset(uint((pageNo - 1) * pageSize)).Limit(uint(pageSize))
	expr, _, err := sql.ToSQL()
	if err != nil {
		return nil, err
	}
	if sel.Debug {
		p.Log.Infof("Query event clickhouse SQL:\n%s", expr)
	}

	var (
		events []event.Event
		cancel context.CancelFunc
	)
	ctx, cancel = context.WithTimeout(ctx, p.Cfg.QueryTimeout)
	defer cancel()
	if err = p.clickhouse.Client().Select(ctx, &events, expr); err != nil {
		return nil, err
	}

	var res []*event.Event
	for i := range events {
		events[i].Timestamp = events[i].Time.UnixNano()
		res = append(res, &events[i])
	}
	return res, nil
}
