package exception

import (
	"fmt"

	"github.com/gocql/gocql"

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
	exceptionService *exceptionService
	cassandraSession *gocql.Session
}

func (p *provider) Init(ctx servicehub.Context) error {
	c := ctx.Service("cassandra").(cassandra.Interface)
	session, err := c.Session(&p.Cfg.Cassandra)
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
