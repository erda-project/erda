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

package kafka_mirror

import (
	"fmt"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/kafka"
)

type define struct{}

func (d *define) Service() []string      { return []string{"kafka-mirror"} }
func (d *define) Dependencies() []string { return []string{"kafka", "kafka@output"} }
func (d *define) Summary() string        { return "read from a kafka and write to anohter kafka" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{}    { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	Input  kafka.ConsumerConfig `file:"input"`
	Output kafka.ProducerConfig `file:"output"`
}

type provider struct {
	C      *config
	L      logs.Logger
	kafka  kafka.Interface
	output writer.Writer
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.kafka = ctx.Service("kafka").(kafka.Interface)
	kafkaOutput := ctx.Service("kafka@output").(kafka.Interface)
	w, err := kafkaOutput.NewProducer(&p.C.Output)
	if err != nil {
		return fmt.Errorf("fail to create kafka producer: %s", err)
	}
	p.output = w
	return nil
}

// Start .
func (p *provider) Start() error {
	return p.kafka.NewConsumer(&p.C.Input, p.invoke)
}

func (p *provider) Close() error {
	p.L.Debug("not support close kafka mirror")
	return nil
}

func init() {
	servicehub.RegisterProvider("kafka-mirror", &define{})
}
