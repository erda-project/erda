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
	"context"
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/kafka"
	"github.com/erda-project/erda/modules/core-services/model"
)

type config struct {
	Auth struct {
		Username string `file:"username"`
		Password string `file:"password"`
		Force    bool   `file:"force"`
	}
	SignAuth       signAuthConfig       `file:"sign_auth"`
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

	auth *Authenticator
}

func (c *collector) Init(ctx servicehub.Context) error {
	w, err := c.Kafka.NewProducer(&c.Cfg.Output)
	if err != nil {
		return fmt.Errorf("fail to create kafka producer: %s", err)
	}
	c.writer = w

	if err := c.initAuthenticator(context.TODO()); err != nil {
		return err
	}

	r := ctx.Service("http-server",
		// telemetry.HttpMetric(),
		interceptors.CORS(),
		interceptors.Recover(c.Logger),
	).(httpserver.Router)
	if err := c.intRoute(r); err != nil {
		return fmt.Errorf("fail to init route: %s", err)
	}
	if err := c.intRouteV2(r); err != nil {
		return fmt.Errorf("initRouteV2 faile: %w", err)
	}
	return nil
}

func (c *collector) initAuthenticator(ctx context.Context) error {
	c.auth = &Authenticator{
		store:  make(map[string]*model.AccessKey),
		logger: c.Logger.Sub("Authenticator"),
	}
	if err := c.auth.syncAccessKey(ctx); err != nil {
		return err
	}
	go func() {
		c.Logger.Info("start syncAccessKey...")
		tick := time.NewTicker(c.Cfg.SignAuth.SyncInterval)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
			}
			if err := c.auth.syncAccessKey(ctx); err != nil {
				c.Logger.Errorf("auth.syncAccessKey failed. err: %s", err)
			}
		}
	}()
	return nil
}

func init() {
	servicehub.RegisterProvider("monitor-collector", &define{})
}
