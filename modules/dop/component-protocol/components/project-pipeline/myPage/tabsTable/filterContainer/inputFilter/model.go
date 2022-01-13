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

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-pipeline/common/gshelper"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

type (
	InputFilter struct {
		Type       string                 `json:"type"`
		Name       string                 `json:"name"`
		Props      map[string]interface{} `json:"props"`
		Operations map[string]interface{} `json:"operations"`
		State      State                  `json:"state"`
		gsHelper   *gshelper.GSHelper
	}
	State struct {
		Base64UrlQueryParams    string                 `json:"inputFilter__urlQuery"`
		FrontendConditionProps  []filter.PropCondition `json:"conditions"`
		FrontendConditionValues FrontendConditions     `json:"values"`
	}
	Condition struct {
		Key         string `json:"key"`
		Placeholder string `json:"placeholder"`
		Type        string `json:"type"`
	}
	FrontendConditions struct {
		Name string `json:"name"`
	}
)

func (i *InputFilter) SetType() {
	i.Type = "ContractiveFilter"
}

func (i *InputFilter) SetName() {
	i.Name = "inputFilter"
}

func (i *InputFilter) SetProps() {
	i.Props = map[string]interface{}{
		"delay": 1000,
	}
}

func (i *InputFilter) SetOperations() {
	i.Operations = map[string]interface{}{
		OperationKeyFilter.String(): map[string]interface{}{
			"key":    OperationKeyFilter.String(),
			"reload": true,
		},
	}
}

func (i *InputFilter) SetState(ctx context.Context, name string) {
	i.State = State{
		FrontendConditionProps: []filter.PropCondition{
			{
				Key:         "name",
				Placeholder: cputil.I18n(ctx, "searchByPipelineName"),
				Type:        filter.PropConditionTypeInput,
			},
		},
		FrontendConditionValues: FrontendConditions{
			Name: name,
		},
	}
}

func (i *InputFilter) SetToProtocolComponent(c *cptype.Component) error {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &c)
}

func (i *InputFilter) InitFromProtocol(ctx context.Context, c *cptype.Component, gs *cptype.GlobalStateData) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, i); err != nil {
		return err
	}
	i.gsHelper = gshelper.NewGSHelper(gs)
	return nil
}

const OperationKeyFilter filter.OperationKey = "filter"

func (i *InputFilter) generateUrlQueryParams() (string, error) {
	fb, err := json.Marshal(i.State.FrontendConditionValues)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(fb), nil
}

func (i *InputFilter) flushOptsByFilter(filterEntity string) error {
	b, err := base64.StdEncoding.DecodeString(filterEntity)
	if err != nil {
		return err
	}
	i.State.FrontendConditionValues = FrontendConditions{}
	return json.Unmarshal(b, &i.State.FrontendConditionValues)
}
