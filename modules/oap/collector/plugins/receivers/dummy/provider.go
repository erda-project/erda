package dummy

import (
	"strings"
	"sync"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	mpb "github.com/erda-project/erda-proto-go/oap/metrics/pb"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
	"google.golang.org/protobuf/types/known/structpb"
)

var providerName = plugins.WithPrefixReceiver("dummy")

type config struct {
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger

	label         string
	consumerFuncs []model.MetricReceiverConsumeFunc
	mu            sync.RWMutex
}

func (p *provider) ComponentID() model.ComponentID {
	return model.ComponentID(strings.Join([]string{providerName, p.label}, "@"))
}

func (p *provider) RegisterConsumeFunc(consumer model.MetricReceiverConsumeFunc) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.consumerFuncs = append(p.consumerFuncs, consumer)
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	p.label = ctx.Label()
	p.consumerFuncs = make([]model.MetricReceiverConsumeFunc, 0)
	ticker := time.NewTicker(3 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
			}

			p.mu.RLock()
			for _, fn := range p.consumerFuncs {
				fn(model.Metrics{Metrics: []*mpb.Metric{
					{
						Name:         "mock-metric",
						TimeUnixNano: uint64(time.Now().UnixNano()),
						Attributes: map[string]string{
							"cluster_name": "dev",
						},
						DataPoints: map[string]*structpb.Value{
							"x": structpb.NewNumberValue(0.1),
						},
					},
				}})
			}
			p.mu.RUnlock()
		}
	}()
	return nil
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services:    []string{},
		Description: "dummy receiver for debug&test",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
