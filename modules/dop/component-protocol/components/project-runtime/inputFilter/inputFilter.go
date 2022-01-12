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

package page

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-runtime/common"
)

type InputFilter struct {
	base.DefaultProvider
	Type  string
	sdk   *cptype.SDK
	State State
}
type Values map[string]interface{}

type Condition struct {
	EmptyText   string `json:"emptyText"`
	Key         string `json:"key"`
	Label       string `json:"label"`
	HaveFilter  bool   `json:"haveFilter"`
	Type        string `json:"type"`
	Placeholder string `json:"placeholder"`
}
type State struct {
	Conditions []Condition `json:"conditions"`
	Values     `json:"values"`
}

func (p *InputFilter) Init(ctx servicehub.Context) error {
	return p.DefaultProvider.Init(ctx)
}

func (p *InputFilter) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	sdk := cputil.SDK(ctx)
	p.sdk = sdk
	c.Operations = p.getOperation()
	switch event.Operation {
	case cptype.InitializeOperation:
		p.State = p.getState(sdk)
	case cptype.RenderingOperation, common.ProjectRuntimeFilter:
		err := common.Transfer(c.State, &p.State)
		if err != nil {
			return err
		}
		(*gs)["nameFilter"] = p.State.Values
	}
	err := common.Transfer(p.State, &c.State)
	if err != nil {
		return err
	}
	return nil
}
func (p *InputFilter) getOperation() map[string]interface{} {
	return map[string]interface{}{
		"filter": map[string]interface{}{"key": "filter", "reload": true},
	}
}
func (p *InputFilter) getState(sdk *cptype.SDK) State {
	s := State{
		Conditions: []Condition{
			{
				Key:         "title",
				Placeholder: sdk.I18n("search by runtime name"),
				Type:        "input",
			},
		},
		Values: Values{},
	}
	return s
}

func init() {
	base.InitProviderWithCreator(common.ScenarioKey, "inputFilter", func() servicehub.Provider {
		return &InputFilter{}
	})
}
