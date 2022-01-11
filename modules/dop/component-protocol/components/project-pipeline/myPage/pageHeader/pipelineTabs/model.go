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

package pipelineTabs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
)

type stateValue string

const (
	mimeState    stateValue = "mime"
	primaryState stateValue = "primary"
	allState     stateValue = "all"
)

var defaultState = mimeState

func (s stateValue) String() string {
	return string(s)
}

type (
	Tab struct {
		Type  string `json:"type"`
		Data  Data   `json:"data"`
		State State  `json:"state"`
	}
	Data struct {
		Options []Option `json:"options"`
	}
	Option struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}
	State struct {
		Value string
	}
)

type Num struct {
	MinePipelineNum    uint64 `json:"minePipelineNum"`
	PrimaryPipelineNum uint64 `json:"primaryPipelineNum"`
	AllPipelineNum     uint64 `json:"allPipelineNum"`
}

func (t *Tab) SetType() {
	t.Type = "RadioTabs"
}

func (t *Tab) SetData(ctx context.Context, num Num) {
	t.Data = Data{Options: []Option{
		{
			Label: cputil.I18n(ctx, "minePipeline") + fmt.Sprintf("(%d)", num.MinePipelineNum),
			Value: mimeState.String(),
		},
		{
			Label: cputil.I18n(ctx, "primaryPipeline") + fmt.Sprintf("(%d)", num.PrimaryPipelineNum),
			Value: primaryState.String(),
		},
		{
			Label: cputil.I18n(ctx, "allPipeline") + fmt.Sprintf("(%d)", num.AllPipelineNum),
			Value: allState.String(),
		},
	}}
}

func (t *Tab) SetState(s string) {
	t.State = State{Value: s}
}

func (t *Tab) SetToProtocolComponent(c *cptype.Component) error {
	b, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &c)
}

func (t *Tab) InitFromProtocol(ctx context.Context, c *cptype.Component) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, t)
}
