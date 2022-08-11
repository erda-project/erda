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

package addButton

import (
	"context"

	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/protocol"
)

type TestPlanManageAddButton struct {
}

func (tpm *TestPlanManageAddButton) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if event.Operation.String() == "addTest" {
		if c.State == nil {
			c.State = make(map[string]interface{}, 0)
		}
		c.State["addTest"] = true
		c.State["isUpdate"] = false
	}

	return nil
}

func init() {
	cpregister.RegisterLegacyComponent("auto-test-plan-list", "addButton", func() protocol.CompRender { return &TestPlanManageAddButton{} })
}
