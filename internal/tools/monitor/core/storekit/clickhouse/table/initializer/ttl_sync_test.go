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
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda/internal/tools/monitor/core/settings/retention-strategy"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"
)

func Test_syncTTL(t *testing.T) {

	var (
		database   = "monitor"
		defaultTTL = &retention.TTL{
			HotData: time.Hour * 24 * 3,
			All:     time.Hour * 24 * 7,
		}
		ttl = &retention.TTL{
			HotData: time.Hour * 3,
			All:     time.Hour * 24,
		}
		tablePrefix = "logs"
		tableMetas  = map[string]*loader.TableMeta{
			"monitor.logs":              {TTLBaseField: "toDateTime(timestamp)", TTLDays: 1, TimeKey: "timestamp"},
			"monitor.logs_all":          {},
			"monitor.logs_erda_xxx":     {TTLBaseField: "toDateTime(timestamp)", TTLDays: 7, TimeKey: "timestamp"},
			"monitor.logs_erda_search":  {},
			"monitor.logs_erda_xxx_all": {},
		}
	)

	var (
		changingTTL = map[string]*retention.TTL{}
	)

	monkey.Patch((*MockRetention).Default, func(ret *MockRetention) *retention.TTL {
		return defaultTTL
	})
	defer monkey.Unpatch((*MockRetention).Default)

	monkey.Patch((*MockRetention).GetTTL, func(ret *MockRetention, key string) *retention.TTL {
		return ttl
	})
	defer monkey.Unpatch((*MockRetention).GetTTL)

	monkey.Patch((*MockLoader).WaitAndGetTables, func(l *MockLoader, ctx context.Context) map[string]*loader.TableMeta {
		return tableMetas
	})
	defer monkey.Unpatch((*MockLoader).WaitAndGetTables)

	monkey.Patch((*provider).AlterTableTTL, func(p *provider, tableName string, meta *loader.TableMeta, ttl *retention.TTL) {
		changingTTL[tableName] = ttl
	})
	defer monkey.Unpatch((*provider).AlterTableTTL)

	p := &provider{
		Cfg: &config{
			Database:        database,
			TablePrefix:     tablePrefix,
			TTLSyncInterval: time.Millisecond,
		},
		Log:       logrusx.New(),
		Retention: &MockRetention{},
		Loader:    &MockLoader{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()
	p.syncTTL(ctx)

	assert.EqualValues(t, 2, len(changingTTL))

	assert.EqualValues(t, 7, changingTTL["monitor.logs"].GetTTLByDays())

	assert.EqualValues(t, 1, changingTTL["monitor.logs_erda_xxx"].GetTTLByDays())

}

func Test_syncHotColdTTL(t *testing.T) {

	var (
		database   = "monitor"
		defaultTTL = &retention.TTL{
			HotData: time.Hour * 24 * 3,
			All:     time.Hour * 24 * 7,
		}
		ttl = &retention.TTL{
			HotData: time.Hour * 24 * 1,
			All:     time.Hour * 24 * 2,
		}
		tablePrefix = "logs"
		tableMetas  = map[string]*loader.TableMeta{
			"monitor.logs":               {TTLBaseField: "toDateTime(timestamp)", TTLDays: 1, TimeKey: "timestamp"},                // default table
			"monitor.logs_erda_xxx":      {TTLBaseField: "toDateTime(timestamp)", TTLDays: 7, TimeKey: "timestamp"},                // not default
			"monitor.logs_erda_xxx2":     {TTLBaseField: "toDateTime(timestamp)", TTLDays: 7, HotTTLDays: 2, TimeKey: "timestamp"}, // not default
			"monitor.logs_erda_xxx3":     {TTLBaseField: "toDateTime(timestamp)", TTLDays: 1, HotTTLDays: 1, TimeKey: "timestamp"}, // not default
			"monitor.logs_erda_xxx4":     {TTLBaseField: "toDateTime(timestamp)", TTLDays: 2, HotTTLDays: 2, TimeKey: "timestamp"}, // not default
			"monitor.logs_all":           {},                                                                                       //no ttl base field
			"monitor.logs_erda_search":   {},                                                                                       //no ttl base field
			"monitor.logs_erda_xxx_all":  {},                                                                                       //no ttl base field
			"monitor.logs_erda_xxx2_all": {},                                                                                       //no ttl base field
			"monitor.logs_erda_xxx3_all": {},                                                                                       //no ttl base field
			"monitor.logs_erda_xxx4_all": {},                                                                                       //no ttl base field
		}
	)

	var (
		changingTTL = map[string]*retention.TTL{}
	)

	monkey.Patch((*MockRetention).Default, func(ret *MockRetention) *retention.TTL {
		return defaultTTL
	})
	defer monkey.Unpatch((*MockRetention).Default)

	monkey.Patch((*MockRetention).GetTTL, func(ret *MockRetention, key string) *retention.TTL {
		return ttl
	})
	defer monkey.Unpatch((*MockRetention).GetTTL)

	monkey.Patch((*MockLoader).WaitAndGetTables, func(l *MockLoader, ctx context.Context) map[string]*loader.TableMeta {
		return tableMetas
	})
	defer monkey.Unpatch((*MockLoader).WaitAndGetTables)

	monkey.Patch((*provider).AlterTableTTL, func(p *provider, tableName string, meta *loader.TableMeta, ttl *retention.TTL) {
		changingTTL[tableName] = ttl
	})
	defer monkey.Unpatch((*provider).AlterTableTTL)

	p := &provider{
		Cfg: &config{
			Database:        database,
			TablePrefix:     tablePrefix,
			TTLSyncInterval: time.Millisecond,
			ColdHotEnable:   true,
		},
		Log:       logrusx.New(),
		Retention: &MockRetention{},
		Loader:    &MockLoader{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()
	p.syncTTL(ctx)

	assert.EqualValues(t, 5, len(changingTTL))

	assert.EqualValues(t, 3, changingTTL["monitor.logs"].GetHotTTLByDays())
	assert.EqualValues(t, 7, changingTTL["monitor.logs"].GetTTLByDays())

	assert.EqualValues(t, 1, changingTTL["monitor.logs_erda_xxx"].GetHotTTLByDays())
	assert.EqualValues(t, 2, changingTTL["monitor.logs_erda_xxx"].GetTTLByDays())

	assert.EqualValues(t, 1, changingTTL["monitor.logs_erda_xxx2"].GetHotTTLByDays())
	assert.EqualValues(t, 2, changingTTL["monitor.logs_erda_xxx2"].GetTTLByDays())

	assert.EqualValues(t, 1, changingTTL["monitor.logs_erda_xxx3"].GetHotTTLByDays())
	assert.EqualValues(t, 2, changingTTL["monitor.logs_erda_xxx3"].GetTTLByDays())

	assert.EqualValues(t, 1, changingTTL["monitor.logs_erda_xxx4"].GetHotTTLByDays())
	assert.EqualValues(t, 2, changingTTL["monitor.logs_erda_xxx4"].GetTTLByDays())
}
