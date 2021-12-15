package core

import (
	"context"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
)

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	p.Log.Info("message: ", p.Cfg.Message)
	return nil
}

// Run this is optional
func (p *provider) Run(ctx context.Context) error {
	tick := time.NewTicker(3 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			p.Log.Info("do something...")
		case <-ctx.Done():
			return nil
		}
	}
}

func init() {
	servicehub.Register("erda.oap.collector.core", &servicehub.Spec{
		Services: []string{
			"erda.oap.collector.core-service",
		},
		Description: "here is description of erda.oap.collector.core",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
