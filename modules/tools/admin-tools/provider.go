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

package admin_tools

import (
	"fmt"

	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/cassandra"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/kafka"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pkg/bundle-ex/cmdb"
	"github.com/gocql/gocql"
	"github.com/olivere/elastic"
)

type define struct{}

func (d *define) Service() []string { return []string{"admin-tools"} }
func (d *define) Dependencies() []string {
	return []string{"http-server@admin", "elasticsearch", "cassandra", "kafka"}
}
func (d *define) Summary() string     { return "admin tools" }
func (d *define) Description() string { return d.Summary() }
func (d *define) Config() interface{} { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	Cassandra cassandra.SessionConfig `file:"cassandra"`
}

type provider struct {
	C         *config
	L         logs.Logger
	bundle    *bundle.Bundle
	cmdb      *cmdb.Cmdb
	metricq   metricq.Queryer
	es        *elastic.Client
	cassandra *gocql.Session
	kafka     struct {
		kafka.Interface
		producer writer.Writer
	}
	ctx servicehub.Context
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.ctx = ctx
	// kafka
	p.kafka.Interface = ctx.Service("kafka").(kafka.Interface)
	pc := &kafka.ProducerConfig{
		Parallelism: 0,
		Shared:      true,
	}
	pc.Batch.Size = 0
	producer, err := p.kafka.NewProducer(pc)
	if err != nil {
		return err
	}
	p.kafka.producer = producer

	// elasticsearch
	es := ctx.Service("elasticsearch").(elasticsearch.Interface)
	p.es = es.Client()

	// cassandra
	cassandra := ctx.Service("cassandra").(cassandra.Interface)
	session, err := cassandra.Session(&p.C.Cassandra)
	if err != nil {
		return fmt.Errorf("fail to create cassandra session: %s", err)
	}
	p.cassandra = session

	// http
	routes := ctx.Service("http-server@admin", interceptors.Recover(p.L), interceptors.CORS()).(httpserver.Router)
	return p.intRoutes(routes)
}

// Start .
func (p *provider) Start() error {
	return nil
}

// Close .
func (p *provider) Close() error {
	return p.kafka.producer.Close()
}

func init() {
	servicehub.RegisterProvider("admin-tools", &define{})
}
