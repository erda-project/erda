package collector

import (
	"context"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
)

var providerName = plugins.WithPrefixReceiver("collector")

type config struct {
	// some fields of config for this provider
	Message string `file:"message" flag:"msg" default:"hi" desc:"message to print"`
}

// +provider
type provider struct {
	Cfg    *config
	Log    logs.Logger
	Router httpserver.Router `autowired:"http-router"`

	consumer model.ObservableDataConsumerFunc
}

func (p *provider) ComponentID() model.ComponentID {
	return model.ComponentID(providerName)
}

func (p *provider) RegisterConsumer(consumer model.ObservableDataConsumerFunc) {
	p.consumer = consumer
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {

	return nil
}

func (p *provider) routes() {
	// old
	// p.Router.POST("/collect/logs/:source", p.collectLogs)
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
	servicehub.Register(providerName, &servicehub.Spec{
		Description: "here is description of erda.oap.collector.receiver.collector",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
