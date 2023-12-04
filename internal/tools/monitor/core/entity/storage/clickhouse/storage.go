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

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/providers/clickhouse"
	"github.com/erda-project/erda/internal/tools/monitor/core/entity"
	"github.com/erda-project/erda/internal/tools/monitor/core/entity/storage"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit"
	tablepkg "github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table"
)

func (p *provider) NewWriter(ctx context.Context) (storekit.BatchWriter, error) {
	return p.clickhouse.NewWriter(&clickhouse.WriterOptions{
		Encoder: func(data interface{}) (item *clickhouse.WriteItem, err error) {
			entityData := data.(*entity.Entity)
			item = &clickhouse.WriteItem{
				Data: entityData,
			}
			var (
				wait  <-chan error
				table string
			)

			if p.Retention == nil {
				return nil, errors.Errorf("provider storage-rentention-strategy@entity is required")
			}

			key := p.Retention.GetConfigKey(entityData.Key, entityData.Values)
			ttl := p.Retention.GetTTL(key)
			if len(key) > 0 {
				wait, table = p.Creator.Ensure(ctx, entityData.Values["org_name"], key, tablepkg.FormatTTLToDays(ttl))
			} else {
				wait, table = p.Creator.Ensure(ctx, entityData.Values["org_name"], "", tablepkg.FormatTTLToDays(ttl))
			}

			if wait != nil {
				select {
				case <-wait:
				case <-ctx.Done():
					return nil, storekit.ErrExitConsume
				}
			}
			item.Table = table
			return item, nil
		},
	}), nil
}

func (p *provider) SetEntity(ctx context.Context, data *entity.Entity) error {
	return nil
}

func (p *provider) RemoveEntity(ctx context.Context, typ, key string) (bool, error) {
	return true, nil
}

func (p *provider) GetEntity(ctx context.Context, typ, key string) (*entity.Entity, error) {
	table, _ := p.Loader.GetSearchTable("")
	sql := goqu.From(table).Where(
		goqu.L("type", typ),
		goqu.L("key", key),
	)
	exp, _, err := sql.ToSQL()
	if err != nil {
		return nil, err
	}

	var (
		res    []*entity.Entity
		cancel context.CancelFunc
	)
	ctx, cancel = context.WithTimeout(ctx, p.Cfg.QueryTimeout)
	defer cancel()
	if err = p.clickhouse.Client().Select(ctx, &res, exp); err != nil {
		return nil, err
	}
	if len(res) > 0 {
		return res[0], nil
	}
	return nil, nil
}

func (p *provider) ListEntities(ctx context.Context, opts *storage.ListOptions) ([]*entity.Entity, int64, error) {
	table, _ := p.Loader.GetSearchTable("")
	sql := goqu.From(table).Select(
		goqu.L("min(timestamp)").As("_timestamp"),
		goqu.L("max(update_timestamp)").As("_update_timestamp"),
		goqu.L("any(id)").As("id"),
		goqu.L("any(type)").As("_type"),
		goqu.L("key"),
		goqu.L("any(values)").As("_values"),
		goqu.L("any(labels)").As("_labels"),
	).GroupBy("key")

	var limit uint = 0
	if opts != nil {
		if len(opts.Type) > 0 {
			sql = sql.Where(goqu.L("type").Eq(opts.Type))
		}
		for k, v := range opts.Labels {
			sql = sql.Where(goqu.L("labels[?]", k).Eq(v))
		}
		if opts.CreateTimeUnixNanoMin > 0 || opts.CreateTimeUnixNanoMax > 0 {
			sql = sql.Where(goqu.C("timestamp").Gte(goqu.L("fromUnixTimestamp64Nano(cast(?,'Int64'))", opts.CreateTimeUnixNanoMin)),
				goqu.C("timestamp").Lt(goqu.L("fromUnixTimestamp64Nano(cast(?,'Int64'))", opts.CreateTimeUnixNanoMax)))
		}
		if opts.UpdateTimeUnixNanoMin > 0 || opts.UpdateTimeUnixNanoMax > 0 {
			sql = sql.Where(goqu.C("update_timestamp").Gte(goqu.L("fromUnixTimestamp64Nano(cast(?,'Int64'))", opts.UpdateTimeUnixNanoMin)),
				goqu.C("update_timestamp").Lt(goqu.L("fromUnixTimestamp64Nano(cast(?,'Int64'))", opts.UpdateTimeUnixNanoMax)))
		}
		limit = uint(opts.Limit)
	}

	if limit <= 0 {
		limit = 100
	}

	selectSQL := sql.Limit(limit)
	expr, _, err := selectSQL.ToSQL()
	if err != nil {
		p.Log.Errorf("failed to convert sql, %v", err)
		return nil, 0, err
	}
	if opts.Debug {
		p.Log.Infof("Query entity clickhouse SQL:\n%s", expr)
	}

	var (
		entities []entity.GroupedEntity
		counts   []struct {
			Count uint64 `ch:"count"`
		}
		cancel context.CancelFunc
	)
	ctx, cancel = context.WithTimeout(ctx, p.Cfg.QueryTimeout)
	defer cancel()
	if err := p.clickhouse.Client().Select(ctx, &entities, expr); err != nil {
		p.Log.Errorf("failed to query clickhouse, %v", err)
		return nil, 0, err
	}

	countSQL := sql.Select(goqu.L("count(*)").As("count"))
	expr, _, err = countSQL.ToSQL()
	if err != nil {
		p.Log.Errorf("failed to convert to count sql, %v", err)
		return nil, 0, err
	}
	if opts.Debug {
		p.Log.Infof("Query count of entity clickhouse SQL:\n%s", expr)
	}
	if err := p.clickhouse.Client().Select(ctx, &counts, expr); err != nil {
		p.Log.Errorf("failed to query count, %v", err)
		return nil, 0, err
	}

	var res []*entity.Entity
	for _, e := range entities {
		res = append(res, &entity.Entity{
			Timestamp:          e.Timestamp,
			UpdateTimestamp:    e.UpdateTimestamp,
			ID:                 e.ID,
			Type:               e.Type,
			Key:                e.Key,
			Values:             e.Values,
			Labels:             e.Labels,
			CreateTimeUnixNano: e.Timestamp.UnixNano(),
			UpdateTimeUnixNano: e.UpdateTimestamp.UnixNano(),
		})
	}
	var count uint64 = 0
	if len(counts) > 0 {
		count = counts[0].Count
	}
	return res, int64(count), nil
}
