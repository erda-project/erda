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

package workCards

import (
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/cardlist"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/cardlist/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/admin/component-protocol/components/personal-workbench/common"
	"github.com/erda-project/erda/modules/admin/component-protocol/types"
	"github.com/erda-project/erda/modules/admin/services/workbench"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type WorkCards struct {
	base.DefaultProvider
	impl.DefaultCard
	filterReq apistructs.IssuePagingRequest
	State     State `json:"state"`
	bdl       *bundle.Bundle
	wb        *workbench.Workbench
}

func (wc *WorkCards) RegisterCardListStarOp(opData cardlist.OpCardListStar) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		data := cardlist.Data{}
		err := common.Transfer(opData.ClientData.DataRef, &data)
		if err != nil {
			return
		}
		tabName := wc.getTableName(sdk)
		for _, card := range data.Cards {
			if card.Star {
				continue
			}
			id, err := strconv.ParseUint(card.ID, 10, 64)
			if err != nil {
				logrus.Error(err)
				continue
			}
			req := apistructs.UnSubscribeReq{
				Type:   apistructs.SubscribeType(tabName),
				TypeID: id,
				UserID: sdk.Identity.UserID,
			}
			err = wc.bdl.DeleteSubscribe(sdk.Identity.UserID, sdk.Identity.OrgID, req)
			if err != nil {
				logrus.Errorf("star %v %v failed, id: %v, error: %v", req.Type, req.TypeID, err)
				return
			}
		}

	}
}

type State struct {
	TabName string `json:"tabName"`
}

func (wc *WorkCards) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		wc.wb = sdk.Ctx.Value(types.WorkbenchSvc).(*workbench.Workbench)
		wc.bdl = sdk.Ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
		wc.LoadList(sdk)
	}
}

func (wc *WorkCards) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return wc.RegisterInitializeOp()
}

func getAppTextMeta(sdk *cptype.SDK, app apistructs.AppWorkBenchItem) (metas []cardlist.TextMeta) {
	mrData := make(cptype.OpServerData)
	runtimeData := make(cptype.OpServerData)

	mrOps := common.Operation{
		JumpOut: false,
		Target:  "appOpenMr",
		Query:   map[string]interface{}{"projectId": app.ProjectID, "appId": app.ID},
		Params:  nil,
	}
	runtimeOps := common.Operation{
		JumpOut: false,
		Target:  "deploy",
		Query:   map[string]interface{}{"projectId": app.ProjectID, "appId": app.ID},
		Params:  nil,
	}
	err := common.Transfer(mrOps, &mrData)
	if err != nil {
		logrus.Error(err)
		return
	}
	err = common.Transfer(runtimeOps, &runtimeData)
	if err != nil {
		logrus.Error(err)
		return
	}
	metas = []cardlist.TextMeta{
		{
			MainText: float64(app.AppOpenMrNum),
			SubText:  "MR " + sdk.I18n("Count"),
			Operations: map[cptype.OperationKey]cptype.Operation{"clickGoto": {
				ServerData: &mrData,
			}},
		},
		{
			MainText: float64(app.AppRuntimeNum),
			SubText:  "Runtime " + sdk.I18n("Count"),
			Operations: map[cptype.OperationKey]cptype.Operation{"clickGoto": {
				ServerData: &runtimeData,
			}},
		},
	}
	return
}
func (wc *WorkCards) getProjTextMeta(sdk *cptype.SDK, project apistructs.WorkbenchProjOverviewItem) (metas []cardlist.TextMeta) {
	todayData := make(cptype.OpServerData)
	expireData := make(cptype.OpServerData)
	metas = make([]cardlist.TextMeta, 0)
	project.ProjectDTO.Type = types.ProjTypeDevops
	switch project.ProjectDTO.Type {
	case types.ProjTypeDevops:
		urls, err := wc.wb.GetIssueQueries(project.ProjectDTO.ID)
		if err != nil {
			return
		}
		todayOp := common.Operation{
			JumpOut: false,
			Target:  "projectAllIssue",
			Query:   map[string]interface{}{"issueFilter__urlQuery": urls.TodayExpireQuery},
			Params:  map[string]interface{}{"projectId": project.ProjectDTO.ID},
		}
		expireOp := common.Operation{
			JumpOut: false,
			Target:  "projectAllIssue",
			Query:   map[string]interface{}{"issueFilter__urlQuery": urls.ExpiredQuery},
			Params:  map[string]interface{}{"projectId": project.ProjectDTO.ID},
		}

		err = common.Transfer(expireOp, &expireData)
		if err != nil {
			logrus.Error(err)
			return
		}
		err = common.Transfer(todayOp, &todayData)
		if err != nil {
			logrus.Error(err)
			return
		}
		if project.IssueInfo == nil {
			project.IssueInfo = &apistructs.ProjectIssueInfo{}
		}

		metas = []cardlist.TextMeta{
			{
				MainText: float64(project.IssueInfo.ExpiredIssueNum),
				SubText:  sdk.I18n("expired"),
				Operations: map[cptype.OperationKey]cptype.Operation{
					"clickGoto": cptype.Operation{
						ServerData: &expireData,
					},
				},
			},
			{
				MainText: float64(project.IssueInfo.TotalIssueNum),
				SubText:  sdk.I18n("today expire"),
				Operations: map[cptype.OperationKey]cptype.Operation{
					"clickGoto": cptype.Operation{
						ServerData: &todayData,
					},
				},
			},
		}
		return
	case types.ProjTypeMSP:
		if project.IssueInfo == nil {
			project.IssueInfo = &apistructs.ProjectIssueInfo{}
		}
		metas = []cardlist.TextMeta{
			{
				MainText: float64(project.StatisticInfo.ServiceCount),
				SubText:  sdk.I18n("service count"),
				Operations: map[cptype.OperationKey]cptype.Operation{
					"clickGoto": cptype.Operation{
						ServerData: &expireData,
					},
				},
			},
			{
				MainText: float64(project.StatisticInfo.Last24HAlertCount),
				SubText:  sdk.I18n("last 24 hour alarm count"),
				Operations: map[cptype.OperationKey]cptype.Operation{
					"clickGoto": cptype.Operation{
						ServerData: &todayData,
					},
				},
			},
		}
		return
	default:
		return
	}
}
func (wc *WorkCards) getProjectCardOps(params workbench.UrlParams, project apistructs.WorkbenchProjOverviewItem) (ops map[cptype.OperationKey]cptype.Operation) {
	ops = make(map[cptype.OperationKey]cptype.Operation)
	ops["star"] = cptype.Operation{}
	serviceOp := common.Operation{
		JumpOut: false,
		Params:  map[string]interface{}{},
	}
	err := common.Transfer(params, &serviceOp.Params)
	if err != nil {
		logrus.Error("card operation error :%v", err)
		return
	}
	target := ""
	switch project.ProjectDTO.Type {
	case common.DevOpsProject:
		target = "project"
		serviceOp.Params["projectId"] = project.ProjectDTO.ID
	case common.MspProject:
		target = "mspServiceList"

	}
	serviceOp.Target = target

	serverData := make(cptype.OpServerData)

	err = common.Transfer(serviceOp, &serverData)
	if err != nil {
		logrus.Error(err)
		return
	}
	ops["clickGoto"] = cptype.Operation{ServerData: &serverData}
	return
}

func (wc *WorkCards) getAppCardOps(app apistructs.AppWorkBenchItem) (ops map[cptype.OperationKey]cptype.Operation) {
	ops = make(map[cptype.OperationKey]cptype.Operation)
	ops["star"] = cptype.Operation{}
	serviceOp := common.Operation{
		JumpOut: false,
		Target:  "mspServiceList",
		Params:  map[string]interface{}{"projectId": app.ProjectID, "appId": app.ID},
	}
	serverData := make(cptype.OpServerData)

	err := common.Transfer(serviceOp, &serverData)
	if err != nil {
		logrus.Error(err)
		return
	}
	ops["clickGoto"] = cptype.Operation{ServerData: &serverData}
	return
}
func (wc *WorkCards) getAppIconOps(sdk *cptype.SDK, app apistructs.AppWorkBenchItem) (iops []cardlist.IconOperations) {
	iops = make([]cardlist.IconOperations, 0)
	gotoData := cardlist.OpCardListGotoData{
		JumpOut: false,
		Params: cptype.ExtraMap{
			"projectId": app.ProjectID,
			"appId":     app.ID,
		},
	}
	pipelineServerData := make(cptype.OpServerData)
	apiDesignServerData := make(cptype.OpServerData)
	deployData := make(cptype.OpServerData)
	repositoryServerData := make(cptype.OpServerData)
	common.Transfer(gotoData, &pipelineServerData)
	common.Transfer(gotoData, &apiDesignServerData)
	common.Transfer(gotoData, &deployData)
	common.Transfer(gotoData, &repositoryServerData)
	pipelineServerData["target"] = "repo"
	apiDesignServerData["target"] = "pipelineRoot"
	deployData["target"] = "appApiDesign"
	repositoryServerData["target"] = "deploy"
	iops = []cardlist.IconOperations{
		{
			Icon: "daimacangku",
			Tip:  sdk.I18n("code repository"),
			Operations: map[cptype.OperationKey]cptype.Operation{
				"clickGoto": {ServerData: &repositoryServerData},
			}},
		{
			Icon: "liushuixian",
			Tip:  sdk.I18n("pipeline"),
			Operations: map[cptype.OperationKey]cptype.Operation{
				"clickGoto": {ServerData: &pipelineServerData},
			}},
		{
			Icon: "Apisheji",
			Tip:  sdk.I18n("api design"),
			Operations: map[cptype.OperationKey]cptype.Operation{
				"clickGoto": {ServerData: &apiDesignServerData},
			}},
		{
			Icon: "bushuzhongxin",
			Tip:  sdk.I18n("deploy center"),
			Operations: map[cptype.OperationKey]cptype.Operation{
				"clickGoto": {ServerData: &deployData},
			}},
	}
	return
}
func (wc *WorkCards) getProjIconOps(sdk *cptype.SDK, project apistructs.WorkbenchProjOverviewItem, params workbench.UrlParams) []cardlist.IconOperations {

	gotoData := cardlist.OpCardListGotoData{
		JumpOut: false,
		Params:  make(cptype.ExtraMap),
	}
	err := common.Transfer(params, &gotoData.Params)
	if err != nil {
		logrus.Error(err)
		return nil
	}
	gotoData.Params["projectId"] = project.ProjectDTO.ID
	switch project.ProjectDTO.Type {
	case types.ProjTypeMSP:
		serviceListServerData := make(cptype.OpServerData)
		monitorServerData := make(cptype.OpServerData)
		traceServerData := make(cptype.OpServerData)
		logAnalysisServerData := make(cptype.OpServerData)
		common.Transfer(gotoData, &serviceListServerData)
		common.Transfer(gotoData, &monitorServerData)
		common.Transfer(gotoData, &traceServerData)
		common.Transfer(gotoData, &logAnalysisServerData)
		serviceListServerData["target"] = "mspServiceList"
		monitorServerData["target"] = "mspMonitorServiceAnalyze"
		traceServerData["target"] = "microTrace"
		logAnalysisServerData["target"] = "mspLogAnalyze"
		return []cardlist.IconOperations{
			{
				Icon: "fuwuliebiao",
				Tip:  sdk.I18n("service list"),
				Operations: map[cptype.OperationKey]cptype.Operation{
					"clickGoto": {
						ServerData: &serviceListServerData,
					},
				},
			},
			{
				Icon: "fuwujiankong",
				Tip:  sdk.I18n("service list"),
				Operations: map[cptype.OperationKey]cptype.Operation{
					"clickGoto": {
						ServerData: &monitorServerData,
					},
				},
			},
			{
				Icon: "lianluzhuizong",
				Tip:  sdk.I18n("service list"),
				Operations: map[cptype.OperationKey]cptype.Operation{
					"clickGoto": {
						ServerData: &traceServerData,
					},
				},
			},
			{
				Icon: "rizhifenxi",
				Tip:  sdk.I18n("service list"),
				Operations: map[cptype.OperationKey]cptype.Operation{
					"clickGoto": {
						ServerData: &logAnalysisServerData,
					},
				},
			},
		}
	case types.ProjTypeDevops:
		projectManageServerData := make(cptype.OpServerData)
		appDevelopServerData := make(cptype.OpServerData)
		testManageServerData := make(cptype.OpServerData)
		serviceMonitorServerData := make(cptype.OpServerData)
		projectSettingServerData := make(cptype.OpServerData)

		common.Transfer(gotoData, &projectManageServerData)
		common.Transfer(gotoData, &appDevelopServerData)
		common.Transfer(gotoData, &testManageServerData)
		common.Transfer(gotoData, &serviceMonitorServerData)
		common.Transfer(gotoData, &projectSettingServerData)
		projectManageServerData["target"] = "projectAllIssue"
		appDevelopServerData["target"] = "projectApps"
		testManageServerData["target"] = "projectTestDashboard"
		serviceMonitorServerData["target"] = "mspServiceList"
		projectSettingServerData["target"] = "projectSetting"
		return []cardlist.IconOperations{
			{
				Icon: "xiangmuguanli",
				Tip:  sdk.I18n("project manage"),
				Operations: map[cptype.OperationKey]cptype.Operation{
					"clickGoto": {ServerData: &projectManageServerData},
				},
			},
			{
				Icon: "yingyongkaifa",
				Tip:  sdk.I18n("app develop"),
				Operations: map[cptype.OperationKey]cptype.Operation{
					"clickGoto": {ServerData: &appDevelopServerData},
				},
			},
			{
				Icon: "ceshiguanli",
				Tip:  sdk.I18n("test manage"),
				Operations: map[cptype.OperationKey]cptype.Operation{
					"clickGoto": {ServerData: &testManageServerData},
				},
			},
			{
				Icon: "xiangmuguanli",
				Tip:  sdk.I18n("service monitor"),
				Operations: map[cptype.OperationKey]cptype.Operation{
					"clickGoto": {ServerData: &serviceMonitorServerData},
				},
			},
			{
				Icon: "xiangmushezhi",
				Tip:  sdk.I18n("project setting"),
				Operations: map[cptype.OperationKey]cptype.Operation{
					"clickGoto": {ServerData: &projectSettingServerData},
				},
			},
		}
	default:
		return []cardlist.IconOperations{}
	}
}

func getTitleState(sdk *cptype.SDK, kind string) []cardlist.TitleState {
	switch kind {
	case apistructs.WorkbenchItemApp.String():
		return []cardlist.TitleState{{Text: sdk.I18n("default"), Status: "success"}}
	case apistructs.WorkbenchItemProj.String():
		return []cardlist.TitleState{{Text: sdk.I18n("default"), Status: "success"}}
	default:
		return []cardlist.TitleState{}
	}
}

func (wc *WorkCards) getTableName(sdk *cptype.SDK) string {
	tabStr := ""
	if tab, ok := (*sdk.GlobalState)[common.WorkTabKey]; !ok {
		tabStr = wc.State.TabName
	} else {
		tabStr = tab.(string)
	}
	return tabStr
}

func (wc *WorkCards) LoadList(sdk *cptype.SDK) {
	tabStr := wc.getTableName(sdk)
	data := cardlist.Data{}
	apiIdentity := apistructs.Identity{}
	apiIdentity.OrgID = sdk.Identity.OrgID
	apiIdentity.UserID = sdk.Identity.UserID
	switch tabStr {
	case apistructs.WorkbenchItemApp.String():
		data.Title = sdk.I18n("star application")
		apps, err := wc.wb.ListSubAppWbData(apiIdentity, 0)
		if err != nil {
			return
		}
		data.TitleSummary = fmt.Sprintf("%d", len(apps.List))
		for _, app := range apps.List {
			data.Cards = append(data.Cards, cardlist.Card{
				ID:             fmt.Sprintf("%d", app.ID),
				ImgURL:         app.Logo,
				Title:          app.Name,
				TitleState:     getTitleState(sdk, apistructs.WorkbenchItemApp.String()),
				Star:           true,
				IconOperations: wc.getAppIconOps(sdk, app),
				TextMeta:       getAppTextMeta(sdk, app),
				Operations:     wc.getAppCardOps(app),
			})
		}
	case apistructs.WorkbenchItemProj.String():
		data.Title = sdk.I18n("star project")
		projects, err := wc.wb.ListSubProjWbData(apiIdentity)
		if err != nil {
			return
		}
		data.TitleSummary = fmt.Sprintf("%d", len(projects.List))
		ids := make([]uint64, 0)
		for _, project := range projects.List {
			ids = append(ids, project.ProjectDTO.ID)
		}
		params, err := wc.wb.GetUrlCommonParams(sdk.Identity.UserID, sdk.Identity.OrgID, ids)
		if err != nil {
			logrus.Errorf("card list fail to get url params ,err :%v", err)
		}
		for i, project := range projects.List {
			data.Cards = append(data.Cards, cardlist.Card{
				ID:             fmt.Sprintf("%d", project.ProjectDTO.ID),
				ImgURL:         project.ProjectDTO.Logo,
				Title:          project.ProjectDTO.Name,
				TitleState:     getTitleState(sdk, apistructs.WorkbenchItemApp.String()),
				Star:           true,
				TextMeta:       wc.getProjTextMeta(sdk, project),
				IconOperations: wc.getProjIconOps(sdk, project, params[i]),
				Operations:     wc.getProjectCardOps(params[i], project),
			})
		}
	}
	wc.StdDataPtr = &data
}

func init() {
	base.InitProviderWithCreator(common.ScenarioKey, "workCards", func() servicehub.Provider {
		return &WorkCards{}
	})
}
