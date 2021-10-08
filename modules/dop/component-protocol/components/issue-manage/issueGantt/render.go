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

package issueGantt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-manage/issueViewGroup"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/gantt"
)

func (g *Gantt) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	// import component data
	if err := g.Import(c); err != nil {
		logrus.Errorf("import component failed, err:%v", err)
		return err
	}
	g.sdk = cputil.SDK(ctx)
	g.bdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)

	// check visible
	visible := false
	if v, ok := c.State["issueViewGroupValue"]; ok {
		if viewType, ok := v.(string); ok {
			if viewType != issueViewGroup.ViewTypeGantt {
				visible = false
				c.Props = map[string]interface{}{}
				c.Props.(map[string]interface{})["visible"] = visible
				return nil
			}
		}
		g.Props.Visible = true
		visible = true
	}
	if !visible {
		c.Props = map[string]interface{}{}
		c.Props.(map[string]interface{})["visible"] = visible
		return nil
	}

	issueType, _ := g.sdk.InParams["fixedIssueType"].(string)
	g.State.IssueType = issueType

	// listen on operation
	switch event.Operation.String() {
	case apistructs.InitializeOperation.String(), apistructs.RenderingOperation.String():
		g.setDefaultState(event.Operation)
		g.setOperations()
		err := g.RenderOnFilter(ctx, c)
		if err != nil {
			logrus.Errorf("render on filter failed, err:%v", err)
			return err
		}
	case gantt.OpChangePageNo.String():
		err := g.RenderOnFilter(ctx, c)
		if err != nil {
			logrus.Errorf("render on filter failed, err:%v", err)
			return err
		}
	default:
		logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, scenario, event)
	}

	// export rendered component data
	if err := g.Export(c, gs); err != nil {
		logrus.Errorf("export component failed, err:%v", err)
		return err
	}
	return nil
}

func (g *Gantt) RenderOnFilter(ctx context.Context, c *cptype.Component) error {
	// get filter request

	req, err := g.getFilterReq(c)
	if err != nil {
		logrus.Errorf("get filter request failed, err:%v", err)
		return err
	}
	g.State.PageSize = 10
	// query issues
	req.PageSize = 10
	rsp, err := g.bdl.PageIssues(*req)
	if err != nil {
		logrus.Errorf("page issue failed, request: %v, err: %v", req, err)
		return err
	}
	// get issues edge time
	edgeMinTime, edgeMaxTime := getEdgeTime(rsp.Data.List)

	// generate gantt props
	g.genProps(ctx, edgeMinTime, edgeMaxTime)

	// generate gantt data
	err = g.genData(rsp.Data.List, edgeMinTime, edgeMaxTime)
	if err != nil {
		logrus.Errorf("generate data failed, request:%v, err:%v", req, err)
		return err
	}
	g.setStateTotal(rsp.Data.Total)
	return nil
}

func (g Gantt) getFilterReq(c *cptype.Component) (req *apistructs.IssuePagingRequest, err error) {
	projectid, err := strconv.ParseUint(g.sdk.InParams["projectId"].(string), 10, 64)
	orgid, err := strconv.ParseUint(g.sdk.Identity.OrgID, 10, 64)

	cond := apistructs.IssuePagingRequest{}
	filterCond, ok := c.State["filterConditions"]
	if ok {
		filterCondS, err := json.Marshal(filterCond)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(filterCondS, &cond); err != nil {
			return nil, err
		}
	} else {
		cond.OrgID = int64(orgid)
		cond.ProjectID = projectid
		cond.IssueListRequest.IdentityInfo.UserID = g.sdk.Identity.UserID
	}
	if g.State.PageSize == 0 {
		cond.PageSize = gantt.DefaultPageSize
		cond.PageNo = gantt.DefaultPageNo
		g.State.PageSize = cond.PageSize
		g.State.PageNo = cond.PageNo
	} else {
		cond.PageNo = g.State.PageNo
		cond.PageSize = g.State.PageSize
	}
	// TODO: test remove
	// cond.IDs = []int64{777, 663}
	// cond.PageSize = 1
	cond.OrderBy = "assignee"
	cond.Type = getIssueTypes(g.State.IssueType)
	return &cond, nil
}

func getIssueTypes(t string) []apistructs.IssueType {
	switch t {
	case "ALL":
		return []apistructs.IssueType{apistructs.IssueTypeTask, apistructs.IssueTypeRequirement, apistructs.IssueTypeBug}
	case apistructs.IssueTypeTask.String():
		return []apistructs.IssueType{apistructs.IssueTypeTask}
	case apistructs.IssueTypeRequirement.String():
		return []apistructs.IssueType{apistructs.IssueTypeRequirement}
	case apistructs.IssueTypeBug.String():
		return []apistructs.IssueType{apistructs.IssueTypeBug}
	default:
		logrus.Warnf("wrong issue type: %v, assign all", t)
		return []apistructs.IssueType{apistructs.IssueTypeTask, apistructs.IssueTypeRequirement, apistructs.IssueTypeBug}
	}
}

func (g *Gantt) Import(c *cptype.Component) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, g); err != nil {
		return err
	}
	return nil
}

func (g *Gantt) Export(c *cptype.Component, gs *cptype.GlobalStateData) error {
	// set component data
	b, err := json.Marshal(g)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, c); err != nil {
		return err
	}
	// set url query state
	err = g.SetStateToUrlQuery(c)
	if err != nil {
		return err
	}
	// set global state: uids
	(*gs)[protocol.GlobalInnerKeyUserIDs.String()] = g.Uids
	return nil
}

func getStateUrlQueryKey() string {
	return fmt.Sprintf("%s__urlQuery", CompName)
}

func (g Gantt) SetStateToUrlQuery(c *cptype.Component) error {
	g.State.Total = 0
	b, err := json.Marshal(g.State)
	if err != nil {
		return err
	}
	urlQueryStr := base64.StdEncoding.EncodeToString(b)
	c.State[getStateUrlQueryKey()] = urlQueryStr
	return nil
}
