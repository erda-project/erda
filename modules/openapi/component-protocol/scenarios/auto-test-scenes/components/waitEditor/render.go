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

package waitEditor

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentAction struct{}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	v, err := json.Marshal(c.State["stepId"])
	if err != nil {
		return err
	}
	var stepID int
	err = json.Unmarshal(v, &stepID)
	if err != nil {
		return nil
	}
	if stepID <= 0 {
		return nil
	}
	switch event.Operation {
	case "submit":
		formDataJson, err := json.Marshal(c.State["formData"])
		if err != nil {
			return err
		}
		formData := map[string]interface{}{}
		err = json.Unmarshal(formDataJson, &formData)
		if err != nil {
			return err
		}

		var req apistructs.AutotestSceneRequest
		req.ID = uint64(stepID)
		req.Value = string(formDataJson)
		req.UserID = bdl.Identity.UserID
		_, err = bdl.Bdl.UpdateAutoTestSceneStep(req)
		if err != nil {
			return err
		}
		c.State["drawVisible"] = false
	case "cancel":
		c.State["drawVisible"] = false
	case apistructs.InitializeOperation, apistructs.RenderingOperation:
		req := apistructs.AutotestGetSceneStepReq{
			ID:     uint64(stepID),
			UserID: bdl.Identity.UserID,
		}
		step, err := bdl.Bdl.GetAutoTestSceneStep(req)
		if err != nil {
			return err
		}
		var waitTime int
		if step.Value == "" {
			waitTime = 0
		} else {
			var value apistructs.AutoTestRunWait
			if err := json.Unmarshal([]byte(step.Value), &value); err != nil {
				return err
			}
			waitTime = value.WaitTime
		}
		c.State["formData"] = map[string]interface{}{
			"waitTime": waitTime,
		}
		c.State["drawVisible"] = true
		c.Props = map[string]interface{}{
			"fields": []map[string]interface{}{
				{
					"label":          "等待时间(s)",
					"component":      "inputNumber",
					"required":       true,
					"key":            "waitTime",
					"componentProps": map[string]interface{}{"min": 1},
				},
			},
		}
		c.Operations = map[string]interface{}{
			"submit": map[string]interface{}{
				"key":    "submit",
				"reload": true,
			},
			"cancel": map[string]interface{}{
				"reload": true,
				"key":    "cancel",
			},
		}
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}
