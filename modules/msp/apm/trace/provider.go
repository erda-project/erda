package trace

import (
	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	pb "github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq"
	"github.com/erda-project/erda/modules/monitor/trace/query"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
}

// +provider
type provider struct {
	Cfg          *config
	Log          logs.Logger
	Register     transport.Register `autowired:"service-register" optional:"true"`
	traceService *traceService
	Metricq      metricq.Queryer    `autowired:"metrics-query"`
	Spanq        query.SpanQueryAPI `autowired:"trace-query"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.traceService = &traceService{
		p:       p,
		metricq: p.Metricq,
		spanq:   p.Spanq,
	}
	if p.Register != nil {
		pb.RegisterTraceServiceImp(p.Register, p.traceService, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.msp.apm.trace.TraceService" || ctx.Type() == pb.TraceServiceServerType() || ctx.Type() == pb.TraceServiceHandlerType():
		return p.traceService
	}
	return p
}

func init() {
	servicehub.Register("erda.msp.apm.trace", &servicehub.Spec{
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
