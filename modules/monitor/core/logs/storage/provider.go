// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/bluele/gcache"
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/cassandra"
	"github.com/erda-project/erda-infra/providers/kafka"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda/modules/monitor/core/logs/storage/schema"
	"github.com/jinzhu/gorm"
)

type define struct{}

func (d *define) Services() []string { return []string{"logs-store"} }
func (d *define) Dependencies() []string {
	return []string{"kafka", "cassandra", "mysql", "cassandra-manager"}
}
func (d *define) Summary() string     { return "logs store" }
func (d *define) Description() string { return d.Summary() }
func (d *define) Config() interface{} { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	Input  kafka.ConsumerConfig `file:"input"`
	Output struct {
		LogSchema struct {
			OrgRefreshInterval time.Duration `file:"org_refresh_interval" default:"2m" env:"LOG_SCHEMA_ORG_REFRESH_INTERVAL"`
		}
		Cassandra struct {
			cassandra.WriterConfig  `file:"writer_config"`
			cassandra.SessionConfig `file:"session_config"`
			GCGraceSeconds          int           `file:"gc_grace_seconds" default:"86400"`
			DefaultTTL              time.Duration `file:"default_ttl" default:"168h"`
			TTLReloadInterval       time.Duration `file:"ttl_reload_interval" default:"3m"`
			CacheStoreInterval      time.Duration `file:"cache_store_interval" default:"3m"`
		} `file:"cassandra"`
		IDKeys []string `file:"id_keys"`
	} `file:"output"`
}

type provider struct {
	Cfg        *config
	Log        logs.Logger
	mysql      *gorm.DB
	kafka      kafka.Interface
	output     writer.Writer
	ttl        ttlStore
	schema     schema.LogSchema
	cache      gcache.Cache
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.mysql = ctx.Service("mysql").(mysql.Interface).DB()
	p.kafka = ctx.Service("kafka").(kafka.Interface)

	cassandra := ctx.Service("cassandra").(cassandra.Cassandra)
	session, err := cassandra.Session(&p.C.Output.Cassandra.SessionConfig)
	if err != nil {
		return fmt.Errorf("fail to create cassandra session, err=%s", err)
	}

	p.output = cassandra.NewBatchWriter(session, &p.C.Output.Cassandra.WriterConfig, p.createLogStatementBuilder)

	p.ttl = &mysqlStore{
		ttlValue:      make(map[string]int),
		defaultTTLSec: int(p.C.Output.Cassandra.DefaultTTL.Seconds()),
		mysql:         p.mysql,
		L:             p.L.Sub("ttlStore"),
	}

	p.schema, err = schema.NewCassandraSchema(cassandra, p.Log.Sub("log schema"))
	if err != nil {
		return err
	}

	p.cache = gcache.New(128).LRU().Build()

	return nil
}

// Start .
func (p *provider) Start() error {
	p.ctx, p.cancelFunc = context.WithCancel(context.Background())

	go p.schema.RunDaemon(p.ctx, p.Cfg.Output.LogSchema.OrgRefreshInterval)
	go p.ttl.Run(p.ctx, p.Cfg.Output.Cassandra.TTLReloadInterval)
	go p.startStoreMetaCache(p.ctx)
	return p.kafka.NewConsumer(&p.Cfg.Input, p.invoke)

}

func (p *provider) Close() error {
	p.cancelFunc()
	return nil
}

func init() {
	servicehub.RegisterProvider("logs-store", &define{})
}
