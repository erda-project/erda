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

package action

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

// testSceneRun testSceneRun component protocol
func testPlanRun(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, globalStateData *apistructs.GlobalStateData) (err error) {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	projectId, err := strconv.Atoi(bdl.InParams["projectId"].(string))

	if err != nil {
		return err
	}
	var field []apistructs.FormPropItem
	props, ok := c.Props.(map[string]interface{})
	if !ok {
		return err
	}
	for key, val := range props {
		if key == "fields" {
			field = val.([]apistructs.FormPropItem)
			break
		}
	}

	// get testplan
	testPlanRequest := apistructs.TestPlanV2PagingRequest{
		ProjectID: uint64(projectId),
	}
	testPlanRequest.UserID = bdl.Identity.UserID
	plans, err := bdl.Bdl.PagingTestPlansV2(testPlanRequest)
	if err != nil {
		return err
	}
	testPlans := make([]map[string]interface{}, 0, plans.Total)
	for _, v := range plans.List {
		testPlans = append(testPlans, map[string]interface{}{"name": fmt.Sprintf("%s-%d", v.Name, v.ID), "value": v.ID})
	}
	// get globalConfigRequest
	globalConfigRequest := apistructs.AutoTestGlobalConfigListRequest{
		ScopeID: strconv.Itoa(projectId),
		Scope:   "project-autotest-testcase",
	}
	globalConfigRequest.UserID = bdl.Identity.UserID

	globalConfigs, err := bdl.Bdl.ListAutoTestGlobalConfig(globalConfigRequest)
	if err != nil {
		return err
	}
	cms := make([]map[string]interface{}, 0, len(globalConfigs))
	for _, v := range globalConfigs {
		cms = append(cms, map[string]interface{}{"name": v.DisplayName, "value": v.Ns})
	}

	var newField []apistructs.FormPropItem
	newField = fillTestPlanFields(field, testPlans, cms)
	newProps := map[string]interface{}{
		"fields": newField,
	}
	c.Props = newProps
	return nil
}

func fillTestPlanFields(field []apistructs.FormPropItem, testPlans []map[string]interface{}, cms []map[string]interface{}) []apistructs.FormPropItem {

	// Add task parameters
	taskParams := apistructs.FormPropItem{
		Component: "formGroup",
		ComponentProps: map[string]interface{}{
			"title": "任务参数",
		},
		Group: "params",
		Key:   "params",
	}
	testPlanField := apistructs.FormPropItem{
		Label:     "测试计划",
		Component: "select",
		Required:  true,
		Key:       "params.test_plan",
		ComponentProps: map[string]interface{}{
			"options": testPlans,
		},
		Group: "params",
	}
	globalConfigField := apistructs.FormPropItem{
		Label:     "参数配置",
		Component: "select",
		Required:  true,
		Key:       "params.cms",
		ComponentProps: map[string]interface{}{
			"options": cms,
		},
		Group: "params",
	}
	waitingResultField := apistructs.FormPropItem{
		Label:        "等待执行结果",
		Component:    "input",
		Required:     false,
		Key:          "params.waiting_result",
		Group:        "params",
		DefaultValue: false,
	}
	var newField []apistructs.FormPropItem
	for _, val := range field {
		newField = append(newField, val)
		if strings.EqualFold(val.Label, "执行条件") {
			newField = append(newField, taskParams)
			newField = append(newField, testPlanField)
			newField = append(newField, globalConfigField)
			newField = append(newField, waitingResultField)

		}
	}
	return newField
}
