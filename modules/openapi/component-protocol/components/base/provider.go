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

package base

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/protocol"
)

var compCreatorMap = map[string]protocol.RenderCreator{}

type DefaultProvider struct{}

// Render is empty implement.
func (p *DefaultProvider) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	return nil
}

// Init .
func (p *DefaultProvider) Init(ctx servicehub.Context) error {
	scenario, compName := MustGetScenarioAndCompNameFromProviderKey(ctx.Key())
	protocol.MustRegisterComponent(&protocol.CompRenderSpec{
		Scenario: scenario,
		CompName: compName,
		RenderC: func() protocol.CompRender {
			if c, ok := compCreatorMap[ctx.Key()]; ok {
				return c()
			}
			return &DefaultProvider{}
		},
	})
	return nil
}

// Provide .
func (p *DefaultProvider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	return p
}

// InitProvider register component as provider.
func InitProvider(scenario, compName string) {
	InitProviderWithCreator(scenario, compName, nil)
}

// InitProviderWithCreator register component as provider with custom providerCreator.
func InitProviderWithCreator(scenario, compName string, creator servicehub.Creator) {
	if creator == nil {
		creator = func() servicehub.Provider { return &DefaultProvider{} }
	}
	servicehub.Register(MakeComponentProviderName(scenario, compName), &servicehub.Spec{Creator: creator})
	compCreatorMap[MakeComponentProviderName(scenario, compName)] = func() protocol.CompRender {
		if r, ok := creator().(protocol.CompRender); ok {
			return r
		}
		return &DefaultProvider{}
	}
}
