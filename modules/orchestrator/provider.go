package orchestrator

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
)

const serviceOrchestrator = "orchestrator"

type provider struct{}

func init() { servicehub.RegisterProvider(serviceOrchestrator, &provider{}) }

func (p *provider) Service() []string                 { return []string{serviceOrchestrator} }
func (p *provider) Dependencies() []string            { return []string{} }
func (p *provider) Init(ctx servicehub.Context) error { return nil }
func (p *provider) Run(ctx context.Context) error     { return Initialize() }
func (p *provider) Creator() servicehub.Creator {
	return func() servicehub.Provider { return &provider{} }
}
