package core

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/oap/collector/core/config"
)

// +provider
type provider struct {
	Cfg *config.Config
	Log logs.Logger

	c *collector
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	c, err := newCollector(ctx, p.Cfg, p.Log.Sub("collector"))
	if err != nil {
		return fmt.Errorf("new collector err: %w", err)
	}
	p.c = c
	return nil
}

// Run this is optional
func (p *provider) Run(ctx context.Context) error {
	p.c.start(ctx)
	return nil
}

func init() {
	servicehub.Register("erda.oap.collector.core", &servicehub.Spec{
		Services:    []string{},
		Description: "core logic for schedule",
		ConfigFunc: func() interface{} {
			return &config.Config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
		// receivers + processors + exporters
		Dependencies: []string{
			// receivers
			"erda.oap.collector.receiver.dummy",
			// processors
			// exporters
			"erda.oap.collector.exporter.stdout",
		},
	})
}
