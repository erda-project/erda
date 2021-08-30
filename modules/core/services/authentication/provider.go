package authentication

import (
	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	pb "github.com/erda-project/erda-proto-go/core/services/authentication/pb"
)

type config struct {
}

// +provider
type provider struct {
	Cfg                   *config
	Log                   logs.Logger
	Register              transport.Register
	authenticationService *authenticationService
}

func (p *provider) Init(ctx servicehub.Context) error {
	// TODO initialize something ...

	p.authenticationService = &authenticationService{p}
	if p.Register != nil {
		pb.RegisterAuthenticationServiceImp(p.Register, p.authenticationService)
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.services.authentication.AuthenticationService" || ctx.Type() == pb.AuthenticationServiceServerType() || ctx.Type() == pb.AuthenticationServiceHandlerType():
		return p.authenticationService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.services.authentication", &servicehub.Spec{
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
