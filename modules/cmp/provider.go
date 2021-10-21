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

// Package cmp Core components of multi-cloud management platform
package cmp

import (
	"context"
	"embed"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	componentprotocol "github.com/erda-project/erda-infra/providers/component-protocol"
	"github.com/erda-project/erda-infra/providers/component-protocol/protocol"
	"github.com/erda-project/erda-infra/providers/i18n"
	alertpb "github.com/erda-project/erda-proto-go/cmp/alert/pb"
	pb2 "github.com/erda-project/erda-proto-go/cmp/dashboard/pb"
	monitor "github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	credentialpb "github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/cmp/metrics"
	"github.com/erda-project/erda/modules/cmp/steve"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

//go:embed component-protocol/scenarios
var scenarioFS embed.FS

type provider struct {
	Server     pb.MetricServiceServer              `autowired:"erda.core.monitor.metric.MetricService"`
	Credential credentialpb.AccessKeyServiceServer `autowired:"erda.core.services.authentication.credentials.accesskey.AccessKeyService" optional:"true"`

	Register        transport.Register `autowired:"service-register" optional:"true"`
	Metrics         *metrics.Metric
	Monitor         monitor.AlertServiceServer `autowired:"erda.core.monitor.alert.AlertService" optional:"true"`
	Protocol        componentprotocol.Interface
	Tran            i18n.Translator `translator:"component-protocol"`
	SteveAggregator *steve.Aggregator
}

type Provider interface {
	SteveServer
	metrics.Interface
}

// Run Run the provider
func (p *provider) Run(ctx context.Context) error {
	runtime.GOMAXPROCS(2)
	p.Metrics = metrics.New(p.Server, ctx)
	logrus.Info("cmp provider is running...")
	ctxNew := context.WithValue(ctx, "metrics", p.Metrics)
	return p.initialize(ctxNew)
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.Protocol.SetI18nTran(p.Tran)
	p.Protocol.WithContextValue(types.GlobalCtxKeyBundle, bundle.New(
		bundle.WithAllAvailableClients(),
		bundle.WithHTTPClient(
			httpclient.New(
				httpclient.WithTimeout(time.Second, time.Second*90),
				httpclient.WithEnableAutoRetry(false),
			)),
	))
	protocol.MustRegisterProtocolsFromFS(scenarioFS)
	pb2.RegisterClusterResourceImp(p.Register, p, apis.Options())
	alertpb.RegisterAlertServiceImp(p.Register, p, apis.Options())

	return nil
}

func init() {
	servicehub.Register("cmp", &servicehub.Spec{
		Services:    []string{"cmp"},
		Description: "Core components of multi-cloud management platform.",
		Creator:     func() servicehub.Provider { return &provider{} },
	})
}
