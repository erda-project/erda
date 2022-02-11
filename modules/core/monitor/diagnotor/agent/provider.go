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

package diagnotor

import (
	"os"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/core/monitor/diagnotor/pb"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
	GatherInterval          time.Duration `file:"gather_interval" default:"5s"`
	TargetContainerCpuLimit int64         `file:"target_container_cpu_limit" env:"TARGET_CONTAINER_CPU_LIMIT"`
	TargetContainerMemLimit int64         `file:"target_container_mem_limit" env:"TARGET_CONTAINER_MEM_LIMIT"`
}

// +provider
type provider struct {
	Cfg      *config
	Log      logs.Logger
	Register transport.Register `autowired:"service-register" optional:"true"`

	exit                  func() error
	diagnotorAgentService *diagnotorAgentService
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.diagnotorAgentService = &diagnotorAgentService{
		p:          p,
		pid:        os.Getpid(),
		lastStatus: &pb.HostProcessStatus{},
	}
	if p.Register != nil {
		pb.RegisterDiagnotorAgentServiceImp(p.Register, p.diagnotorAgentService, apis.Options())
	}
	ctx.AddTask(p.diagnotorAgentService.runGatherProcStat)
	p.exit = ctx.Hub().Close
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.monitor.diagnotor.DiagnotorAgentService" || ctx.Type() == pb.DiagnotorAgentServiceServerType() || ctx.Type() == pb.DiagnotorAgentServiceHandlerType():
		return p.diagnotorAgentService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.monitor.diagnotor", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                pb.Types(),
		OptionalDependencies: []string{"service-register"},
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
