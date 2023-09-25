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

package workContainer

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/admin/personal-workbench/component-protocol/components/personal-workbench/common"
	"github.com/erda-project/erda/internal/apps/admin/personal-workbench/component-protocol/components/personal-workbench/common/gshelper"
)

type WorkContainer struct {
	Options map[string]interface{} `json:"options"`
}

func (w *WorkContainer) SetToProtocolComponent(c *cptype.Component) error {
	b, err := json.Marshal(w)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return err
	}
	return nil
}

func (w *WorkContainer) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	gh := gshelper.NewGSHelper(gs)
	tp, _ := gh.GetWorkbenchItemType()
	switch tp {
	case apistructs.WorkbenchItemPerformanceMeasure:
		w.Options["visible"] = false
	default:
		w.Options["visible"] = true
	}
	return w.SetToProtocolComponent(c)
}

func init() {
	base.InitProviderWithCreator(common.ScenarioKey, "workContainer", func() servicehub.Provider {
		return &WorkContainer{Options: map[string]interface{}{}}
	})
}
