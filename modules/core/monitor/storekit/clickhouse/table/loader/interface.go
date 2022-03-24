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
	"fmt"

	"github.com/erda-project/erda/modules/core/monitor/storekit/clickhouse/table"
)

type Interface interface {
	ExistsTable(tenant, key string) bool
	GetSearchTable(tenant string) string
	ReloadTables() chan error
	Tables() map[string]*TableMeta
}

func (p *provider) ExistsTable(tenant, key string) bool {
	// todo: make table prefix configurable
	tableName := table.NormalizeKey(fmt.Sprintf("logs_%s_%s", tenant, key))
	tables, ok := p.tables.Load().(map[string]*TableMeta)
	if !ok {
		return false
	}
	_, ok = tables[tableName]
	return ok
}

func (p *provider) GetSearchTable(tenant string) string {
	searchTableName := table.NormalizeKey(fmt.Sprintf("logs_%s_search", tenant))
	tables, ok := p.tables.Load().(map[string]*TableMeta)
	if !ok {
		return p.Cfg.DefaultSearchTable
	}
	_, ok = tables[searchTableName]
	if ok {
		return searchTableName
	}
	return p.Cfg.DefaultSearchTable
}

func (p *provider) ReloadTables() chan error {
	ch := make(chan error, 1)
	p.reloadCh <- ch
	return ch
}

func (p *provider) Tables() map[string]*TableMeta {
	tables, _ := p.tables.Load().(map[string]*TableMeta)
	return tables
}
