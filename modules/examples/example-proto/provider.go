package example

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/examples/pb"
)

type config struct {
}

// +provider
type provider struct {
	Cfg            *config
	Log            logs.Logger
	Register       transport.Register
	greeterService *greeterService
	userService    *userService
}

func (p *provider) Init(ctx servicehub.Context) error {
	// TODO initialize something ...

	if p.Register != nil {
		p.greeterService = &greeterService{p}
		pb.RegisterGreeterServiceImp(p.Register, p.greeterService)

		p.userService = &userService{p}
		pb.RegisterUserServiceImp(p.Register, p.userService)
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.example.GreeterService" || ctx.Type() == pb.GreeterServiceServerType() || ctx.Type() == pb.GreeterServiceHandlerType():
		return p.greeterService
	case ctx.Service() == "erda.example.UserService" || ctx.Type() == pb.UserServiceServerType() || ctx.Type() == pb.UserServiceHandlerType():
		return p.userService
	}
	return p
}

func init() {
	servicehub.Register("erda.example", &servicehub.Spec{
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
