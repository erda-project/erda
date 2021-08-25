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

package storage

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bluele/gcache"
	"github.com/gocql/gocql"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/cassandra"
	"github.com/erda-project/erda-infra/providers/kafka"
)

func Test_provider_Init(t *testing.T) {
	mp := mockProvider()
	err := mp.Init(&mockContext{
		l: logrusx.New(),
	})
	assert.Nil(t, err)
}

type mockContext struct {
	l logs.Logger
}

func (m *mockContext) AddTask(task func(context.Context) error, options ...servicehub.TaskOption) {
	panic("implement me")
}

func (m *mockContext) Hub() *servicehub.Hub {
	return &servicehub.Hub{}
}

func (m *mockContext) Config() interface{} {
	return &config{
		Input: kafka.ConsumerConfig{},
		Output: struct {
			LogSchema struct {
				OrgRefreshInterval time.Duration `file:"org_refresh_interval" default:"2m" env:"LOG_SCHEMA_ORG_REFRESH_INTERVAL"`
			} `file:"log_schema"`
			Cassandra struct {
				cassandra.WriterConfig  `file:"writer_config"`
				cassandra.SessionConfig `file:"session_config"`
				GCGraceSeconds          int           `file:"gc_grace_seconds" default:"86400"`
				DefaultTTL              time.Duration `file:"default_ttl" default:"168h"`
				TTLReloadInterval       time.Duration `file:"ttl_reload_interval" default:"3m"`
				CacheStoreInterval      time.Duration `file:"cache_store_interval" default:"3m"`
			} `file:"cassandra"`
			IDKeys []string `file:"id_keys"`
		}{},
	}
}

func (m *mockContext) Logger() logs.Logger {
	return m.l
}

func (m *mockContext) Service(name string, options ...interface{}) interface{} {
	switch name {
	case "cassandra":
		return &mockCassandraInf{}
	default:
		return nil
	}
}

type mockMysql struct {
	db *gorm.DB
}

func newMockMysql() *mockMysql {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	// construct
	rows := sqlmock.NewRows([]string{"org_name", "names", "filters", "config", "key"}).
		AddRow("erda", "", "", "", "")
	mock.ExpectQuery("^SELECT (.*)").WillReturnRows(rows)

	gdb, err := gorm.Open("mysql", sqlDB)
	if err != nil {
		panic(err)
	}
	return &mockMysql{
		db: gdb,
	}
}

func (m *mockMysql) construct() {

}

func (m *mockMysql) DB() *gorm.DB {
	return m.db
}

type mockCassandraInf struct {
}

func (m *mockCassandraInf) CreateKeyspaces(ksc ...*cassandra.KeyspaceConfig) error {
	return nil
}

func (m *mockCassandraInf) Session(cfg *cassandra.SessionConfig) (*gocql.Session, error) {
	return nil, nil
}

func (m *mockCassandraInf) NewBatchWriter(session *gocql.Session, c *cassandra.WriterConfig, builderCreator func() cassandra.StatementBuilder) writer.Writer {
	return &mockWriter{}
}

func mockProvider() *provider {
	p := &provider{}
	p.Cfg = &config{
		Input: kafka.ConsumerConfig{},
		Output: struct {
			LogSchema struct {
				OrgRefreshInterval time.Duration `file:"org_refresh_interval" default:"2m" env:"LOG_SCHEMA_ORG_REFRESH_INTERVAL"`
			} `file:"log_schema"`
			Cassandra struct {
				cassandra.WriterConfig  `file:"writer_config"`
				cassandra.SessionConfig `file:"session_config"`
				GCGraceSeconds          int           `file:"gc_grace_seconds" default:"86400"`
				DefaultTTL              time.Duration `file:"default_ttl" default:"168h"`
				TTLReloadInterval       time.Duration `file:"ttl_reload_interval" default:"3m"`
				CacheStoreInterval      time.Duration `file:"cache_store_interval" default:"3m"`
			} `file:"cassandra"`
			IDKeys []string `file:"id_keys"`
		}{},
	}
	p.cache = gcache.New(128).LRU().Build()
	p.ttl = mockMysqlStore()
	p.Mysql = newMockMysql()
	p.Log = logrusx.New()
	return p
}
