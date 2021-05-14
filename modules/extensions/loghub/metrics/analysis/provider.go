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

package analysis

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/kafka"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda/modules/extensions/loghub/metrics/rules/db"
)

type define struct{}

func (d *define) Service() []string      { return []string{"logs-metrics-analysis"} }
func (d *define) Dependencies() []string { return []string{"kafka", "mysql"} }
func (d *define) Summary() string        { return "parse logs to metrics" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{}    { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	Filters    map[string]string `file:"filters"`
	Processors struct {
		Scope          string        `file:"scope"`
		ScopeID        string        `file:"scope_id"`
		ScopeIDKey     string        `file:"scope_id_key"`
		ReloadInterval time.Duration `file:"reload_interval" default:"3m"`
	} `file:"processors"`
	Input  kafka.ConsumerConfig `file:"input"`
	Output struct {
		Type      string               `file:"type"`
		Kafka     kafka.ProducerConfig `file:"kafka"`
		Collector struct {
			URL      string `file:"url"`
			UserName string `file:"username"`
			Password string `file:"password"`
		} `file:"collector"`
	} `file:"output"`
}

type provider struct {
	C          *config
	L          logs.Logger
	mysql      *gorm.DB
	kafka      kafka.Interface
	output     writer.Writer
	processors atomic.Value
	db         *db.DB
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.mysql = ctx.Service("mysql").(mysql.Interface).DB()
	p.db = db.New(p.mysql)
	p.kafka = ctx.Service("kafka").(kafka.Interface)
	w, err := p.kafka.NewProducer(&p.C.Output.Kafka)
	if err != nil {
		return fmt.Errorf("fail to create kafka producer: %s", err)
	}
	p.output = w
	return nil
}

// Start .
func (p *provider) Start() error {
	err := p.kafka.NewConsumer(&p.C.Input, p.invoke)
	if err != nil {
		return err
	}
	go func() {
		for {
			err := p.loadProcessors()
			if err != nil {
				p.L.Errorf("fail to load processors: %s", err)
			}
			time.Sleep(p.C.Processors.ReloadInterval)
		}
	}()
	return nil
}

func (p *provider) Close() error { return nil }

func init() {
	servicehub.RegisterProvider("logs-metrics-analysis", &define{})
}
