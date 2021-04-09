// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package browser

import (
	"fmt"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/kafka"
	"github.com/erda-project/erda/modules/monitor/utils"
	"github.com/labstack/gommon/log"
)

type define struct{}

func (d *define) Services() []string     { return []string{"browser-analytics"} }
func (d *define) Dependencies() []string { return []string{"kafka"} }
func (d *define) Summary() string        { return "browser-analytics" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{}    { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	Ipdb   string               `file:"ipdb"`
	Output kafka.ProducerConfig `file:"output"`
	Input  kafka.ConsumerConfig `file:"input"`
}

type provider struct {
	Cfg    *config
	Log    logs.Logger
	ipdb   *utils.Locator
	output writer.Writer
	kafka  kafka.Interface
}

func (p *provider) Init(ctx servicehub.Context) error {
	ipdb, err := utils.NewLocator(p.Cfg.Ipdb)
	if err != nil {
		return fmt.Errorf("fail to load ipdb: %s", err)
	}
	p.ipdb = ipdb
	log.Infof("load ipdb from %s", p.Cfg.Ipdb)

	p.kafka = ctx.Service("kafka").(kafka.Interface)
	w, err := p.kafka.NewProducer(&p.Cfg.Output)
	if err != nil {
		return fmt.Errorf("fail to create kafka producer: %s", err)
	}
	p.output = w
	return nil
}

// Start .
func (p *provider) Start() error {
	return p.kafka.NewConsumer(&p.Cfg.Input, p.invoke)
}

// Close .
func (p *provider) Close() error {
	p.Log.Debug("not support close kafka consumer")
	return nil
}

func init() {
	servicehub.RegisterProvider("browser-analytics", &define{})
}
