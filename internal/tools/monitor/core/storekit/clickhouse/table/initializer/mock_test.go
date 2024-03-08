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
	"time"

	ckdriver "github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/erda-project/erda-infra/providers/clickhouse"
	"github.com/erda-project/erda/internal/tools/monitor/core/settings/retention-strategy"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"
)

type MockRetention struct {
}

func (r MockRetention) Default() *retention.TTL {
	//TODO implement me
	panic("implement me")
}

func (r MockRetention) DefaultHotDataTTL() time.Duration {
	return time.Second
}

func (r MockRetention) GetTTL(key string) *retention.TTL {
	panic("implement me")
}

func (r MockRetention) DefaultTTL() time.Duration {
	return time.Second
}

func (r MockRetention) GetConfigKey(name string, tags map[string]string) string {
	panic("implement me")
}

func (r MockRetention) GetTTLByTags(name string, tags map[string]string) time.Duration {
	panic("implement me")
}

func (r MockRetention) Loading(ctx context.Context) {
	panic("implement me")
}

type MockLoader struct {
}

func (m MockLoader) ExistsWriteTable(tenant, key string) (ok bool, writeTableName string) {
	panic("implement me")
}

func (m MockLoader) GetSearchTable(tenant string) (string, *loader.TableMeta) {
	panic("implement me")
}

func (m MockLoader) ReloadTables() chan error {
	panic("implement me")
}

func (m MockLoader) WatchLoadEvent(listener func(map[string]*loader.TableMeta)) {
	panic("implement me")
}

func (m MockLoader) WaitAndGetTables(ctx context.Context) map[string]*loader.TableMeta {
	panic("implement me")
}

func (m MockLoader) Database() string {
	panic("implement me")
}

type MockClickhouse struct {
	checkExec func(sql string)
}

func (m MockClickhouse) NewWriter(opts *clickhouse.WriterOptions) *clickhouse.Writer {
	panic("implement me")
}

func (m MockClickhouse) Client() ckdriver.Conn {
	return MockCkDriver{
		checkExec: m.checkExec,
	}
}

type MockCkDriver struct {
	checkExec func(sql string)
}

func (m MockCkDriver) Contributors() []string {
	panic("implement me")
}

func (m MockCkDriver) ServerVersion() (*ckdriver.ServerVersion, error) {
	panic("implement me")
}

func (m MockCkDriver) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return nil
}

func (m MockCkDriver) Query(ctx context.Context, query string, args ...interface{}) (ckdriver.Rows, error) {
	panic("implement me")
}

func (m MockCkDriver) QueryRow(ctx context.Context, query string, args ...interface{}) ckdriver.Row {
	panic("implement me")
}

func (m MockCkDriver) PrepareBatch(ctx context.Context, query string, opts ...ckdriver.PrepareBatchOption) (ckdriver.Batch, error) {
	panic("implement me")
}

func (m MockCkDriver) Exec(ctx context.Context, query string, args ...interface{}) error {
	m.checkExec(query)
	return nil
}

func (m MockCkDriver) AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error {
	panic("implement me")
}

func (m MockCkDriver) Ping(ctx context.Context) error {
	panic("implement me")
}

func (m MockCkDriver) Stats() ckdriver.Stats {
	panic("implement me")
}

func (m MockCkDriver) Close() error {
	panic("implement me")
}
