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

package queue

import (
	"context"
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/core/pipeline/queue/pb"
	"github.com/erda-project/erda/bundle"
)

type config struct {
	DefaultQueueConcurrency int64   `env:"DEFAULT_QUEUE_CONCURRENCY" default:"20"`
	DefaultQueueCPU         float64 `env:"DEFAULT_QUEUE_CPU" default:"30"`
	DefaultQueueMemoryMB    float64 `env:"DEFAULT_QUEUE_MEMORY_MB" default:"30720"` // default: 30GB
}

type provider struct {
	bdl      *bundle.Bundle
	Cfg      *config
	Log      logs.Logger
	Register transport.Register

	QueueManager pb.QueueServiceServer
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bdl = bundle.New(bundle.WithErdaServer())
	return nil
}

func (s *provider) Run(ctx context.Context) error {
	return nil
}

func init() {
	interfaceType := reflect.TypeOf((*Interface)(nil)).Elem()
	servicehub.Register("queue", &servicehub.Spec{
		Services:     []string{"queue"},
		Types:        []reflect.Type{interfaceType},
		Dependencies: nil,
		Description:  "dop queue",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
