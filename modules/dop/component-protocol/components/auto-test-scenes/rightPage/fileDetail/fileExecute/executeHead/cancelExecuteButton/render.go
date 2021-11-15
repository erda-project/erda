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

package cancelExecuteButton

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type ComponentAction struct {
	base.DefaultProvider
	sdk *cptype.SDK
	bdl *bundle.Bundle
}

func init() {
	base.InitProviderWithCreator("auto-test-scenes", "cancelExecuteButton",
		func() servicehub.Provider { return &ComponentAction{} })
}

func (ca *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {

	ca.bdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)

	switch event.Operation {
	case "cancelExecute":

		var req apistructs.PipelineCancelRequest
		req.PipelineID = uint64(c.State["pipelineId"].(float64))
		req.UserID = ca.sdk.Identity.UserID
		err := ca.bdl.CancelPipeline(req)
		if err != nil {
			return err
		}

		c.State["reloadScenesInfo"] = true
		c.Props = map[string]interface{}{
			"text":    "取消执行",
			"visible": false,
		}
	case cptype.InitializeOperation, cptype.RenderingOperation:
		c.Type = "Button"
		visible := true
		if _, ok := c.State["visible"]; ok {
			visible = c.State["visible"].(bool)
		}
		if _, ok := c.State["pipelineId"]; ok && visible {
			pipelineId := uint64(c.State["pipelineId"].(float64))
			if pipelineId > 0 {
				rsp, err := ca.bdl.GetPipeline(pipelineId)
				if err != nil {
					return err
				}
				if !rsp.Status.IsReconcilerRunningStatus() {
					visible = false
				}
			} else {
				visible = false
			}
		}
		c.Props = map[string]interface{}{
			"text":    "取消执行",
			"visible": visible,
		}
		c.Operations = map[string]interface{}{
			"click": map[string]interface{}{
				"key":    "cancelExecute",
				"reload": true,
			},
		}
	}
	return nil
}
