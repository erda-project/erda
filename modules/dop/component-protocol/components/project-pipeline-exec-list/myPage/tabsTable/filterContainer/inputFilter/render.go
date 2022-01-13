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

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
)

func init() {
	base.InitProviderWithCreator("project-pipeline-exec-list", "inputFilter", func() servicehub.Provider {
		return &InputFilter{}
	})
}

func (i *InputFilter) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if err := i.InitFromProtocol(ctx, c, gs); err != nil {
		return err
	}
	i.SetType()
	i.SetName()
	i.SetProps()
	i.SetOperations()
	i.SetState(ctx, i.State.FrontendConditionValues.Name)
	i.gsHelper.SetPipelineNameFilter(i.State.FrontendConditionValues.Name)
	return i.SetToProtocolComponent(c)
}
