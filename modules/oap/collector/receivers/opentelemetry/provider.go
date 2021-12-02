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

package opentelemetry

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/providers/kafka"
	"github.com/erda-project/erda/modules/oap/collector/receivers/common"

	pb "github.com/erda-project/erda-proto-go/oap/collector/receiver/opentelemetry/pb"
)

var serviceName = "erda.oap.collector.receiver.opentelemetry.OpenTelemetryService"

type config struct {
	Kafka struct {
		Producer kafka.ProducerConfig `file:"producer"  desc:"kafka Producer Config"`
	} `file:"kafka"`
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger

	otlpService pb.OpenTelemetryServiceServer

	Register     transport.Register  `autowired:"service-register" optional:"true"`
	Kafka        kafka.Interface     `autowired:"kafka@receiver-opentelemetry"`
	Interceptors common.Interceptors `autowired:"erda.oap.collector.receiver.common.Interceptor"`
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	if p.Register != nil {
		writer, err := p.Kafka.NewProducer(&p.Cfg.Kafka.Producer)
		if err != nil {
			return err
		}
		p.otlpService = &otlpService{Log: p.Log, writer: writer}
		pb.RegisterOpenTelemetryServiceImp(p.Register, p.otlpService, transport.WithHTTPOptions(transhttp.WithDecoder(ProtoDecoder), transhttp.WithInterceptor(p.Interceptors.ExtractHttpHeaders)),
			transport.WithInterceptors(p.Interceptors.SpanTagOverwrite))
		//transport.WithInterceptors(p.Interceptors.Authentication, p.Interceptors.SpanTagOverwrite))
		//p.jaegerService = &jaegerServiceImpl{Log: p.Log, writer: writer}
		//pb.RegisterJaegerServiceImp(p.Register, p.jaegerService, transport.WithHTTPOptions(transhttp.WithDecoder(ThriftDecoder), transhttp.WithInterceptor(injectCtx)),
		//	transport.WithInterceptors(p.Interceptors.Authentication, p.Interceptors.SpanTagOverwrite))
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == serviceName:
		return p.otlpService
	}
	return p
}

func init() {
	servicehub.Register("erda.oap.collector.receiver.opentelemetry", &servicehub.Spec{
		Services:    pb.ServiceNames(),
		Description: "here is description of erda.oap.collector.receiver.opentelemetry",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
