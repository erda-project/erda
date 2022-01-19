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
	"context"
	"sync"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/oap/collector/common/filter"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
)

var providerName = plugins.WithPrefixProcessor("aggregator")

type config struct {
	Filter       filter.Config   `file:"filter"`
	Rules        map[string]rule `file:"rules"` // map[<field_key>]<rule>
	EvalDuration time.Duration   `file:"eval_duration"`
}

type rule struct {
	Functions []string `file:"functions"`
	Alias     string   `file:"alias"`
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger

	consumer model.ObservableDataConsumerFunc
	appender *Appender
}

func (p *provider) ComponentID() model.ComponentID {
	return model.ComponentID(providerName)
}

func (p *provider) Process(data model.ObservableData) (model.ObservableData, error) {
	// TODO add to cache
	data.RangeFunc(func(item *model.DataItem) (bool, *model.DataItem) {
		for fieldKey, r := range p.Cfg.Rules {
			if err := p.appender.AddItem(item, fieldKey, r); err != nil {
				p.Log.Errorf("add to appender err: %s", err)
			}
		}
		return false, item
	})
	return data, nil
}

func (p *provider) StartProcessor(consumer model.ObservableDataConsumerFunc) {
	p.consumer = consumer
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	p.appender = NewAppender()
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.Cfg.EvalDuration)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
		cache := p.appender.createCacheSnapshot()
		var wg sync.WaitGroup
		wg.Add(len(cache))
		for ref, se := range cache {
			go func(index uint64, data Series) {
				defer wg.Done()
				od, err := data.Eval()
				if err != nil {
					p.Log.Errorf("ref<%d> evaluation error: %s", err)
				}
				p.consumer(od)
			}(ref, se)
		}
		wg.Wait()
	}
}

func init() {
	servicehub.Register("erda.oap.collector.processor.aggregator", &servicehub.Spec{
		Services: []string{
			providerName,
		},
		Description: "here is description of erda.oap.collector.processor.aggregator",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
