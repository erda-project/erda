package collector

import (
	"context"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
)

var providerName = plugins.WithPrefixExporter("collector")

type config struct {
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	return nil
}

// Run this is optional
func (p *provider) Run(ctx context.Context) error {
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services: []string{
			providerName,
		},
		Description: "here is description of erda.oap.collector.exporter.collector",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
