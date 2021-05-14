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

package exporter

import (
	"fmt"

	"github.com/recallsong/go-utils/encoding/md5x"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/kafka"
)

type define struct{}

func (d *define) Service() []string      { return []string{"logs-exporter-base"} }
func (d *define) Dependencies() []string { return []string{"kafka"} }
func (d *define) Summary() string        { return "logs exporter base" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{}    { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	Input        kafka.ConsumerConfig `file:"input"`
	Output       string               `file:"output" env:"MONITOR_LOG_OUTPUT"`
	OutputConfig string               `file:"output_config" env:"MONITOR_LOG_OUTPUT_CONFIG"`
	Filters      map[string]string    `file:"filters"`
}

type provider struct {
	C     *config
	L     logs.Logger
	kafka kafka.Interface
}

func (p *provider) Init(ctx servicehub.Context) error {
	if len(p.C.Input.Group) <= 0 {
		p.C.Input.Group = fmt.Sprintf("%s-%s", p.C.Output, md5x.SumString(p.C.OutputConfig).String16())
	}
	p.kafka = ctx.Service("kafka").(kafka.Interface)
	return nil
}

// NewConsumer
func (p *provider) NewConsumer(fn OutputFactory) error {
	return p.kafka.NewConsumerWitchCreator(&p.C.Input, func(i int) (kafka.ConsumerFunc, error) {
		output, err := fn(i)
		if err != nil {
			return nil, fmt.Errorf("fail to create output %s", err)
		}
		c := &consumer{
			filters: p.C.Filters,
			log:     p.L,
			output:  output,
		}
		return c.Invoke, nil
	})
}

func init() {
	servicehub.RegisterProvider("logs-exporter-base", &define{})
}
