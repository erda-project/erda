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

package meta

import (
	"context"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"
)

func (p *provider) runClickhouseMetaLoader(ctx context.Context) error {
	p.suppressCacheLoader = true
	p.Log.Info("start clickhouse meta loader")
	defer p.Log.Info("exit clickhouse meta loader")
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

		err := p.reloadMetaFromClickhouse(ctx)
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
			err = p.setCache()
			if err != nil {
				p.Log.Errorf("failed to sync tables to cache: %s", err)
			}
		}

		timer.Reset(p.Cfg.ReloadInterval)
		if p.Cfg.Once {
			p.Log.Info("once fetch is finish")
			return nil
		}
		p.loadLock.Unlock()
	}
}

var now = func() time.Time {
	return time.Now()
}

func (p *provider) reloadMetaFromClickhouse(ctx context.Context) error {

	/*
		CREATE TABLE IF NOT EXISTS <database>.metrics_meta ON CLUSTER '{cluster}'
		(
		    `org_name`            LowCardinality(String),
		    `tenant_id`           LowCardinality(String),
		    `metric_group`        LowCardinality(String),
		    `timestamp`           DateTime64(9,'Asia/Shanghai') CODEC (DoubleDelta),
		    `number_field_keys`   Array(LowCardinality(String)),
		    `string_field_keys`   Array(LowCardinality(String)),
		    `tag_keys`            Array(LowCardinality(String)),
		    INDEX idx_timestamp TYPE minmax GRANULARITY 2
		)
		ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{cluster}-{shard}/{database}/metrics_meta', '{replica}')
		ORDER BY (org_name, tenant_id, metric_group, number_field_keys, string_field_keys, tag_keys);
		TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;
	*/

	/*
	  When time series data is just written, a large number of timelines will be generated
	  In order to reduce the number of scanned rows and improve execution efficiency, try to use the data after clickhouse merge to participate in the calculation
	  The data gap will be calculated for a period of time
	*/
	end := now().UnixNano() - int64(p.Cfg.IgnoreGap)
	start := end + int64(p.Cfg.MetaStartTime)

	expr := goqu.From(fmt.Sprintf("%s.%s", p.Cfg.Database, p.Cfg.MetaTable))

	expr = expr.Select(goqu.C("metric_group"))
	expr = expr.SelectAppend(goqu.C("org_name"))
	expr = expr.SelectAppend(goqu.C("tenant_id"))
	expr = expr.SelectAppend(goqu.L("groupUniqArray(arrayJoin(if(empty(string_field_keys),[null],string_field_keys)))").As("sk"))
	expr = expr.SelectAppend(goqu.L("groupUniqArray(arrayJoin(if(empty(number_field_keys),[null],number_field_keys)))").As("nk"))
	expr = expr.SelectAppend(goqu.L("groupUniqArray(arrayJoin(if(empty(tag_keys),[null],tag_keys)))").As("tk"))

	expr = expr.Where(
		goqu.C("timestamp").Gte(goqu.L("fromUnixTimestamp64Nano(cast(?,'Int64'))", start)),
		goqu.C("timestamp").Lt(goqu.L("fromUnixTimestamp64Nano(cast(?,'Int64'))", end)),
	)
	expr = expr.GroupBy("metric_group", "org_name", "tenant_id")

	sql, _, err := expr.ToSQL()
	if err != nil {
		return errors.Wrap(err, "clickhouse meta loader, build sql is fail")
	}

	rows, err := p.Clickhouse.Client().Query(context.Background(), sql)
	if err != nil {
		return errors.Wrap(err, "failed to query metric meta")
	}

	var metas []MetricMeta

	for rows.Next() {
		cm := MetricMeta{}
		err := rows.ScanStruct(&cm)
		if err != nil {
			return errors.Wrap(err, "failed to scan metric meta")
		}
		metas = append(metas, cm)
	}

	ch := p.updateMetrics(metas)
	select {
	case <-ch:
	case <-ctx.Done():
	}
	return nil
}

func (p *provider) runCacheMetaLoader(ctx context.Context) error {
	p.Log.Info("start cache meta loader")
	defer p.Log.Info("exit cache meta loader")
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
		if p.suppressCacheLoader {
			p.loadLock.Unlock()
			for _, notifier := range notifiers {
				p.reloadCh <- notifier
			}
			return nil
		}

		err := p.reloadMetaFromCache(ctx)
		if err != nil {
			p.Log.Errorf("failed to reload meta from cache: %s", err)
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
		timer.Reset(p.Cfg.ReloadInterval)
		p.loadLock.Unlock()
	}
}

func (p *provider) reloadMetaFromCache(ctx context.Context) error {
	meta, err := p.getCache()
	if err != nil {
		return errors.Wrap(err, "get meta from cache is fail")
	}

	ch := p.updateMetrics(meta)
	select {
	case <-ch:
	case <-ctx.Done():
	}
	return nil
}
