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

package initializer

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"

	"github.com/erda-project/erda/internal/tools/monitor/core/settings/retention-strategy"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"
)

func (p *provider) syncTTL(ctx context.Context) {
	p.Log.Infof("run sync ttl with interval: %v", p.Cfg.TTLSyncInterval)

	ticker := time.NewTicker(p.Cfg.TTLSyncInterval)
	defer ticker.Stop()
	for {
		p.Log.Infof("start check and sync ttls...")
		tables := p.Loader.WaitAndGetTables(ctx)
		for t, meta := range tables {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if meta.TTLDays == 0 || len(meta.TimeKey) == 0 {
				continue
			}

			var ttl *retention.TTL
			if t == fmt.Sprintf("%s.%s", p.Cfg.Database, p.Cfg.TablePrefix) {
				// default table
				ttl = p.Retention.Default()
				p.Log.Debugf("check default %s table ttl, hot: %v->%v, cold: %v->%v", t, meta.HotTTLDays, ttl.GetHotTTLByDays(), meta.TTLDays, ttl.GetTTLByDays())
				if !p.needTTLUpdate(ttl, meta) {
					continue
				}

				p.AlterTableTTL(t, meta, ttl)
				continue

			} else {
				// tenant table
				database, tenant, key, ok := p.extractTenantAndKey(t, meta, tables)

				if !ok {
					continue
				}

				ttl := p.Retention.GetTTL(key)
				p.Log.Debugf("check org %s table ttl, hot: %v->%v, cold: %v->%v", t, meta.HotTTLDays, ttl.GetHotTTLByDays(), meta.TTLDays, ttl.GetTTLByDays())

				if !p.needTTLUpdate(ttl, meta) {
					continue
				}
				writeTable := fmt.Sprintf("%s.%s_%s_%s", database, p.Cfg.TablePrefix, tenant, key)
				p.AlterTableTTL(writeTable, meta, ttl)
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (p *provider) needTTLUpdate(ttl *retention.TTL, meta *loader.TableMeta) bool {
	if p.Cfg.ColdHotEnable && ttl.GetTTLByDays() > 0 && ttl.All > ttl.HotData {
		return meta.TTLDays != ttl.GetTTLByDays() || meta.HotTTLDays != ttl.GetHotTTLByDays()
	} else {
		return meta.TTLDays != ttl.GetTTLByDays()
	}
}

func (p *provider) AlterTableTTL(tableName string, meta *loader.TableMeta, ttl *retention.TTL) {
	sql := "ALTER TABLE <table> ON CLUSTER '{cluster}' MODIFY TTL <time_key> + INTERVAL <ddl_days> DAY;"
	ttlHotDays, ttlDays := ttl.GetHotTTLByDays(), ttl.GetTTLByDays()

	if p.Cfg.ColdHotEnable && ttlHotDays > 0 && ttlDays > ttlHotDays {
		sql = "ALTER TABLE <table> ON CLUSTER '{cluster}' MODIFY TTL <time_key> + toIntervalDay(<hot_ddl_days>) TO VOLUME 'slow', <time_key> + toIntervalDay(<ddl_days>);"
	}
	if len(meta.TimeKey) <= 0 {
		p.Log.Warnf("failed exec ttl, not time key!!", tableName)
		return
	}
	replacer := strings.NewReplacer(
		"<time_key>", meta.TimeKey,
		"<ddl_days>", strconv.FormatInt(ttlDays, 10),
		"<table>", tableName,
		"<hot_ddl_days>", strconv.FormatInt(ttlHotDays, 10))
	sql = replacer.Replace(sql)
	err := p.Clickhouse.Client().Exec(clickhouse.Context(context.Background(), clickhouse.WithSettings(map[string]interface{}{
		"materialize_ttl_after_modify": 0,
	})), sql)
	if err != nil {
		p.Log.Warnf("failed to change ttl of table[%s] to %v day, sql: %s, error: %s", tableName, meta, sql, err)
	} else {
		p.Log.Infof("finish change ttl of table[%s] to %v day", tableName, meta)
	}
}
