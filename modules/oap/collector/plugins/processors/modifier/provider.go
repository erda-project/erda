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

package modifier

import (
	"fmt"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/oap/collector/core/model/odata"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
	"github.com/erda-project/erda/modules/oap/collector/plugins/processors/modifier/operator"
)

var providerName = plugins.WithPrefixProcessor("modifier")

type config struct {
	Rules []operator.ModifierCfg `file:"rules"`

	Keypass map[string][]string `file:"keypass"`
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger

	operators []operator.Operator
}

func (p *provider) ComponentConfig() interface{} {
	return p.Cfg
}

func (p *provider) Process(in odata.ObservableData) (odata.ObservableData, error) {
	for _, op := range p.operators {
		in.HandleKeyValuePair(func(pairs map[string]interface{}) map[string]interface{} {
			return op.Operate(pairs)
		})
	}
	return in, nil
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	ops := make([]operator.Operator, len(p.Cfg.Rules))
	for idx, cfg := range p.Cfg.Rules {
		creator, ok := operator.Creators[cfg.Action]
		if !ok {
			return fmt.Errorf("unsupported action: %q", cfg.Action)
		}
		ops[idx] = creator(cfg)
	}
	p.operators = ops
	return nil
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services: []string{
			providerName,
		},

		Description: "here is description of erda.oap.collector.processor.modifier",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
