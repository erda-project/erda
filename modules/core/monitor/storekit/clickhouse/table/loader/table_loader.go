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

package loader

import (
	"context"
	"encoding/json"
	"time"

	ck "github.com/ClickHouse/clickhouse-go/v2"
)

func (p *provider) storeTablesToCache() error {
	tables, ok := p.tables.Load().(map[string]*TableMeta)
	if !ok {
		return nil
	}
	expire := 3 * p.Cfg.ReloadInterval
	bytes, _ := json.Marshal(tables)
	err := p.Redis.Set(p.Cfg.CacheKeyPrefix+"-all", string(bytes), expire).Err()
	return err
}

func (p *provider) runClickhouseTablesLoader(ctx context.Context) error {
	p.suppressCacheLoader = true
	p.Log.Info("start clickhouse tables loader")
	defer p.Log.Info("exit clickhouse tables loader")
	timer := time.NewTimer(0)
	defer timer.Stop()
	var notifiers []chan error
	for {
		select {
		case <-ctx.Done():
			return nil
		case ch := <-p.reloadCh:
			if ch != nil {
				notifiers = append(notifiers, ch)
			}
		case <-timer.C:
		}

		p.loadLock.Lock()

		err := p.reloadTablesFromClickhouse(ctx)
		if err != nil {
			p.Log.Errorf("failed to reload indices from clickhouse: %s", err)
		}

	drain:
		for {
			select {
			case ch := <-p.reloadCh:
				if ch != nil {
					notifiers = append(notifiers, ch)
				}
			default:
				break drain
			}
		}

		for _, notifier := range notifiers {
			notifier <- err
			close(notifier)
		}
		notifiers = nil

		if p.needSyncTablesToCache {
			err = p.storeTablesToCache()
			if err != nil {
				p.Log.Errorf("failed to sync tables to cache: %s", err)
			}
		}

		timer.Reset(p.Cfg.ReloadInterval)
		p.loadLock.Unlock()
	}
}

func (p *provider) reloadTablesFromClickhouse(ctx context.Context) error {
	var tables []struct {
		Name string `ch:"name"`
	}
	err := p.Clickhouse.Client().
		Select(ctx, &tables, "select name from system.tables where database = @db", ck.Named("db", p.Cfg.Database))
	if err != nil {
		return err
	}
	tablesMeta := map[string]*TableMeta{}
	for _, table := range tables {
		tablesMeta[table.Name] = &TableMeta{}
	}
	ch := p.updateTables(tablesMeta)
	select {
	case <-ch:
	case <-ctx.Done():
	}
	return nil
}
