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

package exception

import (
	"fmt"

	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/cassandra"
	pb "github.com/erda-project/erda-proto-go/msp/apm/exception/pb"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
	Cassandra cassandra.SessionConfig `file:"cassandra"`
}

// +provider
type provider struct {
	Cfg              *config
	Log              logs.Logger
	Register         transport.Register
	Cassandra        cassandra.Interface `autowired:"cassandra"`
	exceptionService *exceptionService
	cassandraSession *cassandra.Session
}

func (p *provider) Init(ctx servicehub.Context) error {
	session, err := p.Cassandra.NewSession(&p.Cfg.Cassandra)
	if err != nil {
		return fmt.Errorf("fail to create cassandra session: %s", err)
	}
	p.cassandraSession = session
	p.exceptionService = &exceptionService{p}
	if p.Register != nil {
		pb.RegisterExceptionServiceImp(p.Register, p.exceptionService, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.msp.apm.exception.ExceptionService" || ctx.Type() == pb.ExceptionServiceServerType() || ctx.Type() == pb.ExceptionServiceHandlerType():
		return p.exceptionService
	}
	return p
}

func init() {
	servicehub.Register("erda.msp.apm.exception", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                pb.Types(),
		OptionalDependencies: []string{"service-register"},
		Description:          "",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
