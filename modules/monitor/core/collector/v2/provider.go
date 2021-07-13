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
	"errors"
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/kafka"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/modules/monitor/core/collector/v2/outputs/console"
	kafka2 "github.com/erda-project/erda/modules/monitor/core/collector/v2/outputs/kafka"
)

type config struct {
	SignAuth       signAuthConfig `file:"sign_auth"`
	Output         outputConfig   `file:"output"`
	TaSamplingRate float64        `file:"ta_sampling_rate" default:"100"`
	Limiter        limitConfig    `file:"limiter"`
}

type outputConfig struct {
	Name string `file:"name" default:"kafka" env:"OUTPUT_NAME"`
	kafka.ProducerConfig
}

type limitConfig struct {
	RequestBodySize string `file:"request_body_size"`
}

type define struct{}

func (m *define) Services() []string     { return []string{} }
func (m *define) Dependencies() []string { return []string{"http-server", "kafka-producer"} }
func (m *define) Summary() string        { return "log and metrics collector" }
func (m *define) Description() string    { return m.Summary() }
func (m *define) Config() interface{}    { return &config{} }
func (m *define) Creator() servicehub.Creator {
	return func() servicehub.Provider { return &provider{} }
}

type provider struct {
	Cfg    *config
	Logger logs.Logger
	writer writer.Writer
	Kafka  kafka.Interface

	auth   *Authenticator
	output Output
}

func (p *provider) Init(ctx servicehub.Context) error {
	if err := p.initAuthenticator(context.TODO()); err != nil {
		return err
	}
	if err := p.initOutput(p.Cfg.Output); err != nil {
		return err
	}

	r := ctx.Service("http-server",
		// telemetry.HttpMetric(),
		interceptors.CORS(),
		interceptors.Recover(p.Logger),
	).(httpserver.Router)
	if err := p.intRouteV2(r); err != nil {
		return fmt.Errorf("initRouteV2 faile: %w", err)
	}
	return nil
}

func (p *provider) initOutput(cfg outputConfig) error {
	switch cfg.Name {
	case kafka2.Selector:
		pcfg := &kafka.ProducerConfig{
			Parallelism: p.Cfg.Output.Parallelism,
		}
		pcfg.Batch = p.Cfg.Output.Batch

		w, err := p.Kafka.NewProducer(pcfg)
		if err != nil {
			return fmt.Errorf("fail to create kafka producer: %s", err)
		}
		p.output = kafka2.New(w)
	case console.Selector:
		p.output = console.New()
	default:
		return errors.New("must specify a output")
	}
	return nil
}

func (p *provider) initAuthenticator(ctx context.Context) error {
	p.auth = &Authenticator{
		store:  make(map[string]*model.AccessKey),
		logger: p.Logger.Sub("Authenticator"),
	}
	if err := p.auth.syncAccessKey(ctx); err != nil {
		return err
	}
	go func() {
		p.Logger.Info("start syncAccessKey...")
		tick := time.NewTicker(p.Cfg.SignAuth.SyncInterval)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
			}
			if err := p.auth.syncAccessKey(ctx); err != nil {
				p.Logger.Errorf("auth.syncAccessKey failed. err: %s", err)
			}
		}
	}()
	return nil
}

func init() {
	servicehub.RegisterProvider("monitor-collector-v2", &define{})
}
