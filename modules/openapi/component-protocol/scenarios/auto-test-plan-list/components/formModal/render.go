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

package formModal

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	auto_test_plan_list "github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-plan-list"
)

type TestPlanManageFormModal struct{}

func RenderCreator() protocol.CompRender {
	return &TestPlanManageFormModal{}
}

type FormModalData struct {
	Name    string   `json:"name"`
	Owners  []string `json:"owners"`
	SpaceID uint64   `json:"spaceID"`
}

func (tpm *TestPlanManageFormModal) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	projectIDStr := fmt.Sprintf("%v", bdl.InParams["projectId"])
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		return err
	}

	// 新建测试计划
	switch event.Operation.String() {
	case "submit":
		if _, ok := c.State["formData"]; !ok {
			return errors.Errorf("formData is empty")
		}
		formDataByte, err := json.Marshal(c.State["formData"])
		if err != nil {
			return err
		}

		if _, ok := c.State["isUpdate"]; ok && c.State["isUpdate"].(bool) {
			defer delete(c.State, "isUpdate")
			// 更新完需要把这个数据置为false，以进入渲染创建表单的逻辑
			c.State["formModalVisible"] = false
			// 更新测试计划
			var req apistructs.TestPlanV2UpdateRequest
			if err := json.Unmarshal(formDataByte, &req); err != nil {
				return err
			}
			req.UserID = bdl.Identity.UserID
			req.TestPlanID = uint64(c.State["formModalTestPlanID"].(float64))
			if err := bdl.Bdl.UpdateTestPlanV2(req); err != nil {
				return err
			}
			c.State["visible"] = false
			c.State["formData"] = nil
		} else if _, ok := c.State["isCreate"]; ok && c.State["isCreate"].(bool) {
			defer delete(c.State, "isCreate")
			// 创建完需要把这个数据置为false
			c.State["addTest"] = false
			// 创建测试计划
			var req apistructs.TestPlanV2CreateRequest
			if err := json.Unmarshal(formDataByte, &req); err != nil {
				return err
			}
			req.UserID = bdl.Identity.UserID
			req.ProjectID = projectID

			if err := bdl.Bdl.CreateTestPlanV2(req); err != nil {
				return err
			}
			c.State["visible"] = false
			c.State["formData"] = nil
		}
	}

	// 更新动作呼出的表单页
	if _, ok := c.State["formModalVisible"]; ok && c.State["formModalVisible"].(bool) {
		defer delete(c.State, "formModalVisible")
		resp, err := bdl.Bdl.GetTestPlanV2(c.State["formModalTestPlanID"].(uint64))
		if err != nil {
			return err
		}
		space, err := bdl.Bdl.GetTestSpace(resp.Data.SpaceID)
		if err != nil {
			return err
		}
		testSpaces := []map[string]interface{}{{"name": space.Name, "value": resp.Data.SpaceID}}
		tsBytes, err := json.Marshal(testSpaces)
		if err != nil {
			return err
		}
		formData := map[string]interface{}{"name": resp.Data.Name, "spaceId": resp.Data.SpaceID, "owners": resp.Data.Owners}
		c.State["visible"] = true
		c.State["formData"] = formData
		c.State["isUpdate"] = true
		c.Props = auto_test_plan_list.GenUpdateFormModalProps(tsBytes)
		return nil
	}

	if _, ok := c.State["addTest"]; ok && c.State["addTest"].(bool) {
		defer delete(c.State, "addTest")
		// 创建动作呼出的表单页面
		c.State["visible"] = true
		c.State["formData"] = nil
		result, err := bdl.Bdl.ListTestSpace(int64(projectID), 500, 1)
		if err != nil {
			return err
		}
		testSpaces := make([]map[string]interface{}, 0, result.Total)
		for _, v := range result.List {
			testSpaces = append(testSpaces, map[string]interface{}{"name": v.Name, "value": v.ID})
		}
		tsBytes, err := json.Marshal(testSpaces)
		if err != nil {
			return err
		}
		c.State["isCreate"] = true
		c.Props = auto_test_plan_list.GenCreateFormModalProps(tsBytes)
		return nil
	}

	return nil
}
