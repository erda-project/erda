package runtimeapis

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
)

type define struct{}

func (d *define) Service() []string      { return []string{"project-apis"} }
func (d *define) Dependencies() []string { return []string{"http-server", "metrics-query"} }
func (d *define) Summary() string        { return "project apis" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider { return &provider{} }
}

type provider struct {
	L       logs.Logger
	metricq metricq.Queryer
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.metricq = ctx.Service("metrics-query").(metricq.Queryer)
	routes := ctx.Service("http-server", interceptors.Recover(p.L)).(httpserver.Router)
	return p.intRoutes(routes)
}

func init() {
	servicehub.RegisterProvider("project-apis", &define{})
}
