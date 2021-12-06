package service

import (
	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	pb "github.com/erda-project/erda-proto-go/msp/apm/service/pb"
	"github.com/erda-project/erda/pkg/common/apis"
	perm "github.com/erda-project/erda/pkg/common/permission"
)

type config struct {
}

// +provider
type provider struct {
	Cfg               *config
	Log               logs.Logger
	Register          transport.Register
	apmServiceService *apmServiceService
	Metric            metricpb.MetricServiceServer `autowired:"erda.core.monitor.metric.MetricService"`
	Perm              perm.Interface               `autowired:"permission"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.apmServiceService = &apmServiceService{p}
	if p.Register != nil {
		pb.RegisterApmServiceServiceImp(p.Register, p.apmServiceService, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.msp.apm.service.ApmServiceService" || ctx.Type() == pb.ApmServiceServiceServerType() || ctx.Type() == pb.ApmServiceServiceHandlerType():
		return p.apmServiceService
	}
	return p
}

func init() {
	servicehub.Register("erda.msp.apm.service", &servicehub.Spec{
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
