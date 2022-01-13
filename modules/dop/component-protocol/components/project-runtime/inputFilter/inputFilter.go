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
	"encoding/base64"
	"encoding/json"

	"github.com/sirupsen/logrus"

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
	State State `json:"state"`
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
	Conditions           []Condition `json:"conditions"`
	Base64UrlQueryParams string      `json:"inputFilter__urlQuery,omitempty"`
	Values               `json:"values"`
}

func (p *InputFilter) Init(ctx servicehub.Context) error {
	return p.DefaultProvider.Init(ctx)
}

func (p *InputFilter) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	sdk := cputil.SDK(ctx)
	p.sdk = sdk
	c.Operations = p.getOperation()
	err := p.getState(sdk, c)
	if err != nil {
		return err
	}
	switch event.Operation {
	case cptype.InitializeOperation:
		if p.State.Values["title"] == "" {
			if urlquery := sdk.InParams.String("inputFilter__urlQuery"); urlquery != "" {
				if err := p.flushOptsByFilter(urlquery); err != nil {
					logrus.Errorf("failed to parse input filter values ,err :%v", err)
					return err
				}
			}
		}
	case cptype.RenderingOperation, common.ProjectRuntimeFilter:
		(*gs)["nameFilter"] = p.State.Values
	}
	urlParam, err := p.generateUrlQueryParams()
	if err != nil {
		return err
	}
	p.State.Base64UrlQueryParams = urlParam
	err = common.Transfer(p.State, &c.State)
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
func (p *InputFilter) getState(sdk *cptype.SDK, c *cptype.Component) error {
	b, err := json.Marshal(c)
	if err != nil {
		logrus.Errorf("failed to parse input filter values ,err :%v", err)
	}
	if err = json.Unmarshal(b, p); err != nil {
		return err
	}
	p.State.Conditions = []Condition{
		{
			Key:         "title",
			Placeholder: sdk.I18n("search by runtime name"),
			Type:        "input",
		},
	}
	return nil
}

func (p *InputFilter) flushOptsByFilter(filterEntity string) error {
	b, err := base64.StdEncoding.DecodeString(filterEntity)
	if err != nil {
		return err
	}
	p.State.Values = Values{}
	return json.Unmarshal(b, &p.State.Values)
}

func (p *InputFilter) generateUrlQueryParams() (string, error) {
	fb, err := json.Marshal(p.State.Values)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(fb), nil
}

func init() {
	base.InitProviderWithCreator(common.ScenarioKey, "inputFilter", func() servicehub.Provider {
		return &InputFilter{}
	})
}
