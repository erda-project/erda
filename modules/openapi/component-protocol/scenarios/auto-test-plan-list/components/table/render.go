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

package table

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	auto_test_plan_list "github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-plan-list"
)

type TestPlanManageTable struct{}

func RenderCreator() protocol.CompRender {
	return &TestPlanManageTable{}
}

type TableItem struct {
	//Assignee    map[string]string `json:"assignee"`
	Id          uint64                 `json:"id"`
	Name        string                 `json:"name"`
	Owners      map[string]interface{} `json:"owners"`
	TestSpace   string                 `json:"testSpace"`
	Operate     Operate                `json:"operate"`
	PassRate    PassRate               `json:"passRate"`
	ExecuteTime string                 `json:"executeTime"`
}

type PassRate struct {
	RenderType string `json:"renderType"`
	Value      string `json:"value"`
}

type Operate struct {
	Operations map[string]interface{} `json:"operations"`
	RenderType string                 `json:"renderType"`
}

// OperationData 解析OperationData
type OperationData struct {
	Meta struct {
		ID         uint64 `json:"id"`
		IsArchived bool   `json:"isArchived"`
	} `json:"meta"`
}

type SortData struct {
	Field string `json:"field"`
	Order string `json:"order"`
}

const (
	OrderAscend  string = "ascend"
	OrderDescend string = "descend"
)

func (tpmt *TestPlanManageTable) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	projectIDStr := fmt.Sprintf("%v", bdl.InParams["projectId"])
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
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
		var operationData OperationData
		odBytes, err := json.Marshal(event.OperationData)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(odBytes, &operationData); err != nil {
			return err
		}
		c.State["formModalVisible"] = true
		c.State["formModalTestPlanID"] = operationData.Meta.ID
		return nil
	case "archive":
		var operationData OperationData
		odBytes, err := json.Marshal(event.OperationData)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(odBytes, &operationData); err != nil {
			return err
		}
		testplan, err := bdl.Bdl.GetTestPlanV2(operationData.Meta.ID)
		if err != nil {
			return err
		}
		if err := bdl.Bdl.UpdateTestPlanV2(apistructs.TestPlanV2UpdateRequest{
			Name:         testplan.Data.Name,
			Desc:         testplan.Data.Desc,
			SpaceID:      testplan.Data.SpaceID,
			Owners:       testplan.Data.Owners,
			IsArchived:   &operationData.Meta.IsArchived,
			TestPlanID:   testplan.Data.ID,
			IdentityInfo: apistructs.IdentityInfo{UserID: bdl.Identity.UserID},
		}); err != nil {
			return err
		}
		now := strconv.FormatInt(time.Now().Unix(), 10)
		project, err := bdl.Bdl.GetProject(projectID)
		if err != nil {
			return err
		}
		audit := apistructs.Audit{
			UserID:       cond.UserID,
			ScopeType:    apistructs.ProjectScope,
			ScopeID:      projectID,
			OrgID:        project.OrgID,
			ProjectID:    projectID,
			Result:       "success",
			StartTime:    now,
			EndTime:      now,
			TemplateName: apistructs.ArchiveTestplanTemplate,
			Context: map[string]interface{}{
				"projectName":  project.Name,
				"testPlanName": testplan.Data.Name,
			},
		}
		if !operationData.Meta.IsArchived {
			audit.TemplateName = apistructs.UnarchiveTestPlanTemplate
		}
		if err := bdl.Bdl.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: audit}); err != nil {
			return err
		}
	}

	// filter带过来的
	if _, ok := c.State["name"]; ok {
		cond.Name = c.State["name"].(string)
	}

	if v, ok := c.State["archive"]; ok && v != nil {
		var isArchive bool
		if s := v.(bool); s == true {
			isArchive = true
		}
		cond.IsArchived = &isArchive
	}
	// orderBy and ASC
	err = convertSortData(&cond, c)
	if err != nil {
		return err
	}

	r, err := bdl.Bdl.PagingTestPlansV2(cond)
	if err != nil {
		return err
	}
	// data
	l := []TableItem{}
	for _, data := range r.List {
		item := TableItem{
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
				Operations: map[string]interface{}{},
			},
			PassRate: PassRate{
				RenderType: "progress",
				Value:      fmt.Sprintf("%.f", data.PassRate),
			},
			ExecuteTime: convertExecuteTime(data),
		}
		if data.IsArchived == true {
			item.Operate.Operations["archive"] = map[string]interface{}{
				"key":    "archive",
				"text":   "取消归档",
				"reload": true,
				"meta":   map[string]interface{}{"id": data.ID, "isArchived": false},
			}
		} else {
			item.Operate.Operations["archive"] = map[string]interface{}{
				"key":    "archive",
				"text":   "归档",
				"reload": true,
				"meta":   map[string]interface{}{"id": data.ID, "isArchived": true},
			}
			item.Operate.Operations["edit"] = map[string]interface{}{
				"key":       "edit",
				"text":      "编辑",
				"reload":    true,
				"meta":      map[string]interface{}{"id": data.ID},
				"showIndex": 2,
			}
		}
		l = append(l, item)
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

func convertSortData(req *apistructs.TestPlanV2PagingRequest, c *apistructs.Component) error {
	if _, ok := c.State["sorterData"]; !ok {
		return nil
	}
	var sortData SortData
	sortDataByte, err := json.Marshal(c.State["sorterData"])
	if err != nil {
		return err
	}
	err = json.Unmarshal(sortDataByte, &sortData)
	if err != nil {
		return err
	}

	if sortData.Field == "passRate" {
		sortData.Field = "pass_rate"
	} else if sortData.Field == "executeTime" {
		sortData.Field = "execute_time"
	}

	req.OrderBy = sortData.Field
	if sortData.Order == OrderAscend {
		req.Asc = true
	} else if sortData.Order == OrderDescend {
		req.Asc = false
	}

	return nil
}

func convertExecuteTime(data *apistructs.TestPlanV2) string {
	if data.ExecuteTime == nil {
		return ""
	}
	var executeTime string
	executeTime = data.ExecuteTime.Format("2006-01-02 15:04:05")
	return executeTime
}

// TODO:
//  1.创建更新完测试计划，要返回更新后的数据 (组件数据传递)
//  2.分页
//  3.根据标题过滤
