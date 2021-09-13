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

package jaeger

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/providers/kafka"
	pb "github.com/erda-project/erda-proto-go/oap/collector/receiver/jaeger/pb"
	"github.com/erda-project/erda/modules/oap/collector/receivers/common"
)

type config struct {
	// some fields of config for this provider
	Enable bool `file:"enable" default:"true" desc:"enable jaeger receiver"`
	Kafka  struct {
		Producer kafka.ProducerConfig `file:"producer"  desc:"kafka Producer Config"`
	} `file:"kafka"`
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger

	jaegerService pb.JaegerServiceServer
	Register      transport.Register `autowired:"service-register" optional:"true"`
	Kafka         kafka.Interface    `autowired:"kafka@receiver-jaeger"`
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	if p.Cfg.Enable && p.Register != nil {
		writer, err := p.Kafka.NewProducer(&p.Cfg.Kafka.Producer)
		if err != nil {
			return err
		}
		p.jaegerService = &jaegerServiceImpl{Log: p.Log, writer: writer}
		pb.RegisterJaegerServiceImp(p.Register, p.jaegerService, transport.WithHTTPOptions(transhttp.WithDecoder(ThriftDecoder)),
			transport.WithInterceptors(common.Authentication, common.TagOverwrite))
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.oap.collector.receiver.jaeger.JaegerService" || ctx.Type() == pb.JaegerServiceServerType() || ctx.Type() == pb.JaegerServiceHandlerType():
		return p.jaegerService
	}
	return p
}

func init() {
	servicehub.Register("erda.oap.collector.receiver.jaeger", &servicehub.Spec{
		Services:    pb.ServiceNames(),
		Description: "here is description of erda.oap.collector.receiver.jaeger",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
