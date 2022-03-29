package aggregator

import (
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/core/model/odata"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
)

var providerName = plugins.WithPrefixProcessor("aggregator")

type config struct {
	Keypass    map[string][]string `file:"keypass"`
	Keyinclude []string            `file:"keyinclude"`

	PushInterval time.Duration `file:"push_interval" default:"60s"`
	Rules        []RuleConfig  `file:"rules"`
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger

	consumer model.ObservableDataConsumerFunc
	cache    map[uint64]aggregate
	rulers   []*ruler
}
type aggregate struct {
	data map[string]interface{}
}

func (p *provider) ComponentConfig() interface{} {
	return p.Cfg
}

func (p *provider) Process(in odata.ObservableData) (odata.ObservableData, error) {
	if in.SourceType() == odata.MetricType {
		return p.add(in), nil
	}
	return in, nil
}

func (p *provider) StartProcessor(consumer model.ObservableDataConsumerFunc) {
	p.consumer = consumer
}

func (p *provider) add(in odata.ObservableData) odata.ObservableData {
	id := in.HashID()
	_, ok := p.cache[id]
	if !ok {
		agg := aggregate{
			data: in.Pairs(),
		}
		for _, rule := range p.rulers {
			agg.data = rule.Fn(nil, agg.data)
		}
		p.cache[id] = agg
		return in
	}

	pre := p.cache[id]
	for _, rule := range p.rulers {
		pre.data = rule.Fn(pre.data, in.Pairs())
	}
	p.cache[id] = pre

	return &odata.Metric{
		Meta: odata.NewMetadata(),
		Data: pre.data,
	}
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	p.cache = make(map[uint64]aggregate)
	p.rulers = make([]*ruler, len(p.Cfg.Rules))
	for idx, r := range p.Cfg.Rules {
		rr, err := newRuler(r)
		if err != nil {
			return fmt.Errorf("newRuler err: %w", err)
		}
		p.rulers[idx] = rr
	}
	return nil
}

// func (p *provider) Run(ctx context.Context) error {
// 	ticker := time.NewTicker(p.Cfg.PushInterval)
// 	defer ticker.Stop()
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			p.push()
// 			return nil
// 		case <-ticker.C:
// 			p.push()
// 		}
// 	}
// }

// func (p *provider) push() {
// 	for _, agg := range p.cache {
// 		p.consumer(&odata.Metric{
// 			Meta: odata.NewMetadata(),
// 			Data: agg.data,
// 		})
// 	}
// }

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services: []string{
			providerName,
		},
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
