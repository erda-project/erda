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
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/oap/collector/core/model/odata"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
)

var providerName = plugins.WithPrefixProcessor("modifier")

type config struct {
	Rules []modifierCfg `file:"rules"`

	Namepass []string `file:"namepass"`
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger
}

func (p *provider) ComponentConfig() interface{} {
	return p.Cfg
}

func (p *provider) Process(in odata.ObservableData) (odata.ObservableData, error) {
	in.HandleAttributes(func(attr map[string]string) map[string]string {
		return p.modify(attr)
	})

	return in, nil
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
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
