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
	"fmt"
	"time"

	"github.com/gocql/gocql"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/cassandra"
	"github.com/erda-project/erda-infra/providers/kafka"
)

type config struct {
	Input  kafka.ConsumerConfig `file:"input"`
	Output struct {
		Cassandra struct {
			cassandra.WriterConfig  `file:"writer_config"`
			cassandra.SessionConfig `file:"session_config"`
			GCGraceSeconds          int           `file:"gc_grace_seconds" default:"86400"`
			TTL                     time.Duration `file:"ttl" default:"168h"`
		} `file:"cassandra"`
		Kafka kafka.ProducerConfig `file:"kafka"`
	} `file:"output"`
}

type provider struct {
	Cfg    *config
	Log    logs.Logger
	ttlSec int
	kafka  kafka.Interface
	output struct {
		cassandra writer.Writer
		kafka     writer.Writer
	}
	cassandraSession *gocql.Session
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.ttlSec = int(p.Cfg.Output.Cassandra.TTL.Seconds())
	cassandra := ctx.Service("cassandra").(cassandra.Interface)
	session, err := cassandra.NewSession(&p.Cfg.Output.Cassandra.SessionConfig)
	p.cassandraSession = session.Session()
	if err != nil {
		return fmt.Errorf("fail to create cassandra session: %s", err)
	}
	err = p.initCassandra(session.Session())
	if err != nil {
		return err
	}
	p.output.cassandra = cassandra.NewBatchWriter(session, &p.Cfg.Output.Cassandra.WriterConfig, p.createTraceStatement)

	p.kafka = ctx.Service("kafka").(kafka.Interface)
	w, err := p.kafka.NewProducer(&p.Cfg.Output.Kafka)
	if err != nil {
		return fmt.Errorf("fail to create kafka producer: %s", err)
	}
	p.output.kafka = w
	return nil
}

// Start .
func (p *provider) Start() error {
	return p.kafka.NewConsumer(&p.Cfg.Input, p.invoke)
}

func (p *provider) Close() error {
	p.Log.Debug("not support close kafka consumer")
	return nil
}

func init() {
	servicehub.Register("trace-storage", &servicehub.Spec{
		Services:     []string{"trace-storage"},
		Dependencies: []string{"kafka", "cassandra"},
		Description:  "trace storage",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
