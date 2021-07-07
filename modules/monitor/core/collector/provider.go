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

package collector

import (
	"fmt"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/kafka"
)

type config struct {
	Auth struct {
		Username string `file:"username"`
		Password string `file:"password"`
		Force    bool   `file:"force"`
	}
	Output         kafka.ProducerConfig `file:"output"`
	TaSamplingRate float64              `file:"ta_sampling_rate" default:"100"`
}

type define struct{}

func (m *define) Services() []string     { return []string{"metrics-collector"} }
func (m *define) Dependencies() []string { return []string{"http-server", "kafka-producer"} }
func (m *define) Summary() string        { return "log and metrics collector" }
func (m *define) Description() string    { return m.Summary() }
func (m *define) Config() interface{}    { return &config{} }
func (m *define) Creator() servicehub.Creator {
	return func() servicehub.Provider { return &collector{} }
}

// collector .
type collector struct {
	Cfg    *config
	Logger logs.Logger
	writer writer.Writer
	Kafka  kafka.Interface
}

func (c *collector) Init(ctx servicehub.Context) error {
	w, err := c.Kafka.NewProducer(&c.Cfg.Output)
	if err != nil {
		return fmt.Errorf("fail to create kafka producer: %s", err)
	}
	c.writer = w

	r := ctx.Service("http-server",
		// telemetry.HttpMetric(),
		interceptors.CORS(),
		interceptors.Recover(c.Logger),
	).(httpserver.Router)
	if err := c.intRoute(r); err != nil {
		return fmt.Errorf("fail to init route: %s", err)
	}
	return nil
}

func init() {
	servicehub.RegisterProvider("monitor-collector", &define{})
}
