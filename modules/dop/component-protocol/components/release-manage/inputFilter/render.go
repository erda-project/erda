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

package inputFilter

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
)

func init() {
	base.InitProviderWithCreator("release-manage", "inputFilter", func() servicehub.Provider {
		return &ComponentInputFilter{}
	})
}

func (c *ComponentInputFilter) Render(ctx context.Context, component *cptype.Component, _ cptype.Scenario,
	event cptype.ComponentEvent, _ *cptype.GlobalStateData) error {
	c.InitComponent(ctx)
	if err := c.GenComponentState(component); err != nil {
		return err
	}

	if event.Operation == cptype.InitializeOperation {
		if err := c.DecodeURLQuery(); err != nil {
			return errors.Errorf("failed to decode url query for input filter component, %v", err)
		}
	}
	c.SetComponentValue()
	if err := c.EncodeURLQuery(); err != nil {
		return errors.Errorf("failed to encode url query for input filter component, %v", err)
	}
	c.Transfer(component)
	return nil
}

func (c *ComponentInputFilter) InitComponent(ctx context.Context) {
	sdk := cputil.SDK(ctx)
	c.sdk = sdk
}

func (c *ComponentInputFilter) GenComponentState(component *cptype.Component) error {
	if component == nil || component.State == nil {
		return nil
	}

	data, err := json.Marshal(component.State)
	if err != nil {
		logrus.Errorf("failed to marshal for inputFilter state, %v", err)
		return err
	}
	var state State
	err = json.Unmarshal(data, &state)
	if err != nil {
		logrus.Errorf("failed to unmarshal for inputFilter state, %v", err)
		return err
	}
	c.State = state
	return nil
}

func (c *ComponentInputFilter) DecodeURLQuery() error {
	queryData, ok := c.sdk.InParams["inputFilter__urlQuery"].(string)
	if !ok {
		return nil
	}
	decode, err := base64.StdEncoding.DecodeString(queryData)
	if err != nil {
		return err
	}

	query := make(map[string]interface{})
	if err = json.Unmarshal(decode, &query); err != nil {
		return err
	}

	c.State.Values.Version, _ = query["version"].(string)
	return nil
}

func (c *ComponentInputFilter) EncodeURLQuery() error {
	urlQuery := make(map[string]interface{})
	urlQuery["version"] = c.State.Values.Version

	jsonData, err := json.Marshal(urlQuery)
	if err != nil {
		return err
	}

	encode := base64.StdEncoding.EncodeToString(jsonData)
	c.State.InputFilterURLQuery = encode
	return nil
}

func (c *ComponentInputFilter) SetComponentValue() {
	c.State.Conditions = []Condition{
		{
			EmptyText:   "all",
			Fixed:       true,
			Key:         "version",
			Label:       "version",
			Placeholder: c.sdk.I18n("searchByVersion"),
			Type:        "input",
		},
	}
	c.Operations = map[string]interface{}{
		"filter": Operation{
			Key:    "filter",
			Reload: true,
		},
	}
}

func (c *ComponentInputFilter) Transfer(component *cptype.Component) {
	component.State = map[string]interface{}{
		"conditions":            c.State.Conditions,
		"values":                c.State.Values,
		"inputFilter__urlQuery": c.State.InputFilterURLQuery,
	}
	component.Operations = c.Operations
}
