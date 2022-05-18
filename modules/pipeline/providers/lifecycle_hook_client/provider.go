package lifecycle_hook_client

import (
	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	pb "github.com/erda-project/erda-proto-go/core/pipeline/lifecycle_hook_client/pb"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
}

// +provider
type provider struct {
	Cfg      *config
	Log      logs.Logger
	Register transport.Register
	MySQL    mysqlxorm.Interface

	lifeCycleService *LifeCycleService
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.lifeCycleService = &LifeCycleService{
		logger:        p.Log,
		dbClient:      &dbclient.Client{Engine: p.MySQL.DB()},
		hookClientMap: map[string]*pb.LifeCycleClient{},
	}
	if p.Register != nil {
		pb.RegisterLifeCycleServiceImp(p.Register, p.lifeCycleService, apis.Options())
	}
	if err := p.lifeCycleService.loadLifecycleHookClient(); err != nil {
		return err
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.pipeline.lifecycle_hook_client.LifeCycleService" || ctx.Type() == pb.LifeCycleServiceServerType() || ctx.Type() == pb.LifeCycleServiceHandlerType():
		return p.lifeCycleService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.pipeline.lifecycle_hook_client", &servicehub.Spec{
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
