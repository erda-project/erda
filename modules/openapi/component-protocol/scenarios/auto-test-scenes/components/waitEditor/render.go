// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
		var waitTimeSec int
		if step.Value == "" {
			waitTimeSec = 0
		} else {
			var value apistructs.AutoTestRunWait
			if err := json.Unmarshal([]byte(step.Value), &value); err != nil {
				return err
			}
			if value.WaitTime > 0 {
				value.WaitTimeSec = value.WaitTime
			}
			waitTimeSec = value.WaitTimeSec
		}
		c.State["formData"] = map[string]interface{}{
			"waitTimeSec": waitTimeSec,
		}
		c.State["drawVisible"] = true
		c.Props = map[string]interface{}{
			"fields": []map[string]interface{}{
				{
					"label":          "等待时间(s)",
					"component":      "inputNumber",
					"required":       true,
					"key":            "waitTimeSec",
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
