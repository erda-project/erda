// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	akpb "github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb"
)

type config struct {
	Auth struct {
		Username string `file:"username"`
		Password string `file:"password"`
		Force    bool   `file:"force"`
	}
	Output         kafka.ProducerConfig `file:"output"`
	TaSamplingRate float64              `file:"ta_sampling_rate" default:"100"`

	SignAuth signAuthConfig `file:"sign_auth"`
}

type define struct{}

func (m *define) Services() []string     { return []string{"metrics-collector"} }
func (m *define) Dependencies() []string { return []string{"http-server", "kafka-producer"} }
func (m *define) Summary() string        { return "log and metrics collector" }
func (m *define) Description() string    { return m.Summary() }
func (m *define) Config() interface{}    { return &config{} }
func (m *define) Creator() servicehub.Creator {
	return func() servicehub.Provider { return &provider{} }
}

// provider .
type provider struct {
	Cfg              *config
	Logger           logs.Logger
	writer           writer.Writer
	Kafka            kafka.Interface
	AccessKeyService akpb.AccessKeyServiceServer `autowired:"erda.core.services.authentication.credentials.accesskey.AccessKeyService"`

	auth *Authenticator
}

func (p *provider) Init(ctx servicehub.Context) error {
	if err := p.initAuthenticator(context.TODO()); err != nil {
		return err
	}
	w, err := p.Kafka.NewProducer(&p.Cfg.Output)
	if err != nil {
		return fmt.Errorf("fail to create kafka producer: %s", err)
	}
	p.writer = w

	r := ctx.Service("http-server",
		// telemetry.HttpMetric(),
		interceptors.CORS(),
		interceptors.Recover(p.Logger),
	).(httpserver.Router)
	if err := p.intRoute(r); err != nil {
		return fmt.Errorf("fail to init route: %s", err)
	}
	return nil
}

func (p *provider) initAuthenticator(ctx context.Context) error {
	p.auth = &Authenticator{
		store:            make(map[string]*akpb.AccessKeysItem),
		logger:           p.Logger.Sub("Authenticator"),
		AccessKeyService: p.AccessKeyService,
	}
	if err := p.auth.syncAccessKey(ctx); err != nil {
		return err
	}
	go func() {
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
	servicehub.RegisterProvider("monitor-collector", &define{})
}
