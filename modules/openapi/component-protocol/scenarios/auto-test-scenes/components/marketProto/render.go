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

package marketProto

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-scenes/components/apiEditor"
)

type MarketProto struct{}

func RenderCreator() protocol.CompRender {
	return &MarketProto{}
}

// OperationData 解析OperationData
type OperationData struct {
	Meta struct {
		KeyWord   string `json:"keyWord"`
		APISpecID uint64 `json:"selectApiSpecId"`
	} `json:"meta"`
}

func (mp *MarketProto) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	// 塞市场接口的数据
	tmpStepID, ok := c.State["stepId"]
	if !ok {
		return errors.New("stepId is empty")
	}
	// TODO: 类型要改
	stepID, err := strconv.ParseUint(fmt.Sprintf("%v", tmpStepID), 10, 64)
	if err != nil {
		return err
	}
	if stepID == 0 {
		return nil
	}

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	userID := bdl.Identity.UserID
	orgIDStr := bdl.Identity.OrgID
	if orgIDStr == "" {
		return errors.New("orgID is empty")
	}
	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		return err
	}
	// var orgID uint64 = 1

	var opreationData OperationData
	odBytes, err := json.Marshal(event.OperationData)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(odBytes, &opreationData); err != nil {
		return err
	}

	if c.Data == nil {
		c.Data = make(map[string]interface{}, 0)
	}

	c.State["apiSpecId"] = nil

	switch event.Operation.String() {
	case "searchAPISpec":
		if opreationData.Meta.KeyWord == "" {
			return nil
		}
		apis, err := bdl.Bdl.SearchAPIOperations(orgID, userID, opreationData.Meta.KeyWord)
		if err != nil {
			return err
		}
		c.Data["list"] = apis
		return nil
	case "changeAPISpec":
		if opreationData.Meta.APISpecID == 0 {
			// 切换成自定义的api
			c.Data["list"] = []apistructs.APIOperationSummary{}
			if _, ok := c.State["value"]; ok {
				delete(c.State, "value")
			}
			c.State["apiSpecId"] = apiEditor.EmptySpecID
			return nil
		}
		// 切换到api集市另一个api
		apiSpec, err := bdl.Bdl.GetAPIOperation(orgID, userID, opreationData.Meta.APISpecID)
		if err != nil {
			return err
		}
		c.Data["list"] = []apistructs.APIOperationSummary{
			{
				ID:          opreationData.Meta.APISpecID,
				OperationID: apiSpec.OperationID,
				Path:        apiSpec.Path,
				Version:     apiSpec.Version,
				Method:      apiSpec.Method,
			},
		}
		c.State["apiSpecId"] = opreationData.Meta.APISpecID
		c.State["value"] = opreationData.Meta.APISpecID
		return nil
	}

	// 渲染
	step, err := bdl.Bdl.GetAutoTestSceneStep(apistructs.AutotestGetSceneStepReq{
		ID:     stepID,
		UserID: userID,
	})
	if err != nil {
		return err
	}
	var apiInfo apistructs.APIInfo
	if step.Value == "" {
		step.Value = "{}"
	}
	if err := json.Unmarshal([]byte(step.Value), &apiInfo); err != nil {
		return err
	}
	if step.APISpecID != 0 {
		apiSpec, err := bdl.Bdl.GetAPIOperation(orgID, userID, step.APISpecID)
		if err != nil {
			return err
		}
		c.Data["list"] = []apistructs.APIOperationSummary{
			{
				ID:          step.APISpecID,
				OperationID: apiSpec.OperationID,
				Path:        apiSpec.Path,
				Version:     apiSpec.Version,
				Method:      apiSpec.Method,
			},
		}
		c.State["value"] = step.APISpecID
	} else {
		if _, ok := c.State["value"]; ok {
			delete(c.State, "value")
		}
	}

	return nil
}
