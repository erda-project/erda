// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aggregator

import (
	"fmt"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/oap/collector/core/model/odata"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
)

var providerName = plugins.WithPrefixProcessor("aggregator")

type config struct {
	Keypass    map[string][]string `file:"keypass"`
	Keydrop    map[string][]string `file:"keydrop"`
	Keyinclude []string            `file:"keyinclude"`
	Keyexclude []string            `file:"keyexclude"`

	Rules []RuleConfig `file:"rules"`
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger

	cache  map[uint64]aggregate
	rulers []*ruler
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
