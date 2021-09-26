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

package table

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/pkg/type_conversion"
	auto_test_plan_list "github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-plan-list"
)

type TestPlanManageTable struct{}

func RenderCreator() protocol.CompRender {
	return &TestPlanManageTable{}
}

type TableItem struct {
	//Assignee    map[string]string `json:"assignee"`
	Id        uint64                 `json:"id"`
	Name      string                 `json:"name"`
	Owners    map[string]interface{} `json:"owners"`
	TestSpace string                 `json:"testSpace"`
	Operate   Operate                `json:"operate"`
}

type Operate struct {
	Operations map[string]interface{} `json:"operations"`
	RenderType string                 `json:"renderType"`
}

// OperationData 解析OperationData
type OperationData struct {
	Meta struct {
		ID uint64 `json:"id"`
	} `json:"meta"`
}

func (tpmt *TestPlanManageTable) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	projectID, err := type_conversion.InterfaceToUint64(bdl.InParams["projectId"])
	if err != nil {
		return err
	}

	cond := apistructs.TestPlanV2PagingRequest{ProjectID: projectID, PageNo: 1, PageSize: auto_test_plan_list.DefaultTablePageSize}
	cond.UserID = bdl.Identity.UserID

	switch event.Operation.String() {
	case "changePageNo":
		if _, ok := c.State["pageNo"]; ok {
			cond.PageNo = uint64(c.State["pageNo"].(float64))
		}
	case "edit":
		var opreationData OperationData
		odBytes, err := json.Marshal(event.OperationData)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(odBytes, &opreationData); err != nil {
			return err
		}
		c.State["formModalVisible"] = true
		c.State["formModalTestPlanID"] = opreationData.Meta.ID
		return nil
	}

	// filter带过来的
	if _, ok := c.State["name"]; ok {
		cond.Name = c.State["name"].(string)
	}

	r, err := bdl.Bdl.PagingTestPlansV2(cond)
	if err != nil {
		return err
	}
	// data
	l := []TableItem{}
	for _, data := range r.List {
		l = append(l, TableItem{
			Id:   data.ID,
			Name: data.Name,
			Owners: map[string]interface{}{
				"renderType": "userAvatar",
				"value":      data.Owners,
				"showIcon":   false,
			},
			TestSpace: data.SpaceName,
			Operate: Operate{
				RenderType: "tableOperation",
				Operations: map[string]interface{}{
					"edit": map[string]interface{}{
						"key":    "edit",
						"text":   "编辑",
						"reload": true,
						"meta":   map[string]interface{}{"id": data.ID},
					},
				},
			},
		})
	}
	c.Data = map[string]interface{}{}
	c.Data["list"] = l
	// state
	if c.State == nil {
		c.State = make(map[string]interface{}, 0)
	}
	c.State["total"] = r.Total
	c.State["pageNo"] = cond.PageNo
	c.State["pageSize"] = cond.PageSize
	c.State["formModalVisible"] = false
	c.State["formModalTestPlanID"] = 0
	(*gs)[protocol.GlobalInnerKeyUserIDs.String()] = r.UserIDs
	return nil
}

// TODO:
//  1.创建更新完测试计划，要返回更新后的数据 (组件数据传递)
//  2.分页
//  3.根据标题过滤
