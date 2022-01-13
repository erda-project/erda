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

package page

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/recallsong/go-utils/container/slice"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/list"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/list/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/admin/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-runtime/common"
)

type List struct {
	base.DefaultProvider
	impl.DefaultList
	PageNo   uint64
	PageSize uint64
	Total    uint64
	Bdl      *bundle.Bundle
	Sdk      *cptype.SDK
}

func (p *List) RegisterItemStarOp(opData list.OpItemStar) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
	}
}

func (p *List) RegisterItemClickGotoOp(opData list.OpItemClickGoto) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
	}
}

func (p *List) RegisterItemClickOp(opData list.OpItemClick) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		// rerender list after any operation
		defer p.RegisterRenderingOp()
		req := apistructs.RuntimeScaleRecords{}
		idStr := opData.ClientData.DataRef.ID
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			logrus.Errorf("failed to parse id ,err: %v", err)
		}
		req.IDs = []uint64{id}

		orgId, err := strconv.ParseUint(p.Sdk.Identity.OrgID, 10, 64)
		if err != nil {
			logrus.Errorf("failed to parse org-id ,err: %v", err)
		}
		switch opData.ClientData.OperationRef.ID {
		case common.DeleteOp:
			var resp apistructs.BatchRuntimeDeleteResults
			resp, err = p.Bdl.BatchUpdateDelete(req, orgId, p.Sdk.Identity.UserID, "delete")
			if err != nil {
				logrus.Errorf("delete runtime err %v", err)
			}
			if len(resp.UnDeletedIds) != 0 {
				logrus.Errorf("failed to %s runtimes ,ids :%v", opData.ClientData.OperationRef.ID, resp.UnDeletedIds)
			}
		case common.ReStartOp:
			var resp apistructs.BatchRuntimeReDeployResults
			resp, err = p.Bdl.BatchUpdateReDeploy(req, orgId, p.Sdk.Identity.UserID, "reDeploy")
			if err != nil {
				logrus.Errorf("redeploy runtime err %v", err)
			}
			if len(resp.UnReDeployedIds) != 0 {
				logrus.Errorf("failed to %s runtimes ,ids :%v", opData.ClientData.OperationRef.ID, resp.UnReDeployedIds)
			}
		}
	}
}

type ClickClientData struct {
	SelectedRowID string
	DataRef       list.MoreOpItem
}

func (p *List) RegisterBatchOp(opData list.OpBatchRowsHandle) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		defer p.RegisterRenderingOp()
		req := apistructs.RuntimeScaleRecords{}
		for _, idStr := range opData.ClientData.SelectedRowIDs {
			id, err := strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				logrus.Errorf("failed to parse id ,err: %v", err)
			}
			req.IDs = append(req.IDs, id)
		}

		orgId, err := strconv.ParseUint(p.Sdk.Identity.OrgID, 10, 64)
		if err != nil {
			logrus.Errorf("failed to parse org-id ,err: %v", err)
		}
		switch opData.ClientData.SelectedOptionsID {
		case common.DeleteOp:
			var resp apistructs.BatchRuntimeDeleteResults
			resp, err = p.Bdl.BatchUpdateDelete(req, orgId, p.Sdk.Identity.UserID, "delete")
			if err != nil {
				logrus.Errorf("delete runtime err %v", err)
			}
			if len(resp.UnDeletedIds) != 0 {
				logrus.Errorf("failed to %s runtimes ,ids :%v", opData.ClientData.SelectedOptionsID, resp.UnDeletedIds)
			}
		case common.ReStartOp:
			var resp apistructs.BatchRuntimeReDeployResults
			resp, err = p.Bdl.BatchUpdateReDeploy(req, orgId, p.Sdk.Identity.UserID, "reDeploy")
			if err != nil {
				logrus.Errorf("redeploy runtime err %v", err)
			}
			if len(resp.UnReDeployedIds) != 0 {
				logrus.Errorf("failed to %s runtimes ,ids :%v", opData.ClientData.SelectedOptionsID, resp.UnReDeployedIds)
			}
		}

	}
}

func (p *List) BeforeHandleOp(sdk *cptype.SDK) {
	p.Sdk = sdk
	p.Bdl = p.Sdk.Ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	p.PageNo = 1
	p.PageSize = 10
}

func (p *List) RegisterChangePage(opData list.OpChangePage) (opFunc cptype.OperationFunc) {
	logrus.Infof("change page client data: %+v", opData)
	if opData.ClientData.PageNo > 0 {
		p.PageNo = opData.ClientData.PageNo
	}
	if opData.ClientData.PageSize > 0 {
		p.PageSize = opData.ClientData.PageSize
	}
	return p.RegisterRenderingOp()
}

func (p *List) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		p.Sdk = sdk
		p.StdDataPtr = p.getData()
	}
}

func (p *List) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		p.Sdk = sdk
		p.StdDataPtr = p.getData()
	}
}

func (p *List) Init(ctx servicehub.Context) error {
	return p.DefaultProvider.Init(ctx)
}

func (p *List) getData() *list.Data {
	data := &list.Data{}
	data.PageNo = p.PageNo
	data.PageSize = p.PageSize
	var runtimes []bundle.GetApplicationRuntimesDataEle
	var runtimeIdToAppNameMap map[uint64]string
	if gsRuntimes, ok := (*p.Sdk.GlobalState)["runtimes"]; !ok {
		logrus.Infof("not found runtimes")
		getEnv, ok := p.Sdk.InParams["env"].(string)
		if !ok {
			logrus.Errorf("env is empty")
			return data
		}
		projectId, err := strconv.ParseUint(p.Sdk.InParams["projectId"].(string), 10, 64)
		if err != nil {
			logrus.Errorf("parse oid failed,%v", err)
			return data
		}
		oid, err := strconv.ParseUint(p.Sdk.Identity.OrgID, 10, 64)
		if err != nil {
			logrus.Errorf("parse oid failed,%v", err)
			return data
		}
		apps, err := p.Bdl.GetAppsByProject(projectId, oid, p.Sdk.Identity.UserID)
		if err != nil {
			logrus.Errorf("get my app failed,%v", err)
			return data
		}
		appIds := make([]uint64, 0)
		appIdToName := make(map[uint64]string)
		for i := 0; i < len(apps.List); i++ {
			appIds = append(appIds, apps.List[i].ID)
			appIdToName[apps.List[i].ID] = apps.List[i].Name
		}
		runtimesByApp, err := p.Bdl.ListRuntimesGroupByApps(oid, p.Sdk.Identity.UserID, appIds)
		if err != nil {
			logrus.Errorf("get my app failed,%v", err)
			return data
		}
		runtimeIdToAppNameMap = make(map[uint64]string)
		for _, v := range runtimesByApp {
			for _, appRuntime := range v {
				if getEnv == appRuntime.Extra.Workspace {
					runtimes = append(runtimes, *appRuntime)
					runtimeIdToAppNameMap[appRuntime.ID] = appIdToName[appRuntime.ApplicationID]
				}
			}
		}
	} else {
		err := common.Transfer(gsRuntimes, &runtimes)
		if err != nil {
			logrus.Errorf("failed to transfer runtimes, gsruntimes %v", err)
			return data
		}
		runtimeIdToAppNameMap = make(map[uint64]string)
		if gsMap, ok := (*p.Sdk.GlobalState)["runtimeIdToAppName"]; ok {
			err = common.Transfer(gsMap, &runtimeIdToAppNameMap)
			if err != nil {
				logrus.Errorf("failed to transfer runtimeMap, runtimeMap %v,err %v", (*p.Sdk.GlobalState)["runtimeIdToAppName"], err)
				return data
			}
		} else {
			logrus.Errorf("failed to transfer runtimeMap, runtimeMap %#v", (*p.Sdk.GlobalState)["runtimeIdToAppName"])
			return data
		}

	}

	oid, err := strconv.ParseUint(p.Sdk.Identity.OrgID, 10, 64)
	if err != nil {
		logrus.Errorf("failed to get oid ,%v", err)
		return data
	}

	userReq := apistructs.UserListRequest{}
	for _, runtime := range runtimes {
		userReq.UserIDs = append(userReq.UserIDs, runtime.Creator)
	}
	users, err := p.Bdl.ListUsers(userReq)
	if err != nil {
		logrus.Errorf("failed to load users,err:%v", err)
	}
	uidToName := make(map[string]string)
	for _, user := range users.Users {
		if user.Nick == "" {
			uidToName[user.ID] = user.Name
		} else {
			uidToName[user.ID] = user.Nick
		}
	}
	ids := make([]string, 0)

	deployId := p.Sdk.InParams["deployId"]
	runtimeMap := make(map[string]bundle.GetApplicationRuntimesDataEle)
	for _, appRuntime := range runtimes {
		//healthyMap := make(map[string]int)
		deployIdStr := strconv.FormatUint(appRuntime.LastOperatorId, 10)
		if deployId != nil && deployId != "" {
			if deployId != deployIdStr {
				continue
			}
		}
		services, err := p.Bdl.GetRuntimeServices(appRuntime.ID, oid, p.Sdk.Identity.UserID)
		if err != nil {
			logrus.Errorf("failed to get runtime %s of detail %v", appRuntime.Name, err)
			continue
		}
		healthyCnt := 0
		for _, s := range services.Services {
			if s.Status == "Healthy" {
				healthyCnt++
			}
		}
		//healthyMap[appRuntime.Name] = healthyCnt
		var healthStr = ""
		if len(services.Services) != 0 {
			healthStr = fmt.Sprintf("%d/%d", healthyCnt, len(services.Services))
		}
		idStr := strconv.FormatUint(appRuntime.ID, 10)
		appIdStr := strconv.FormatUint(appRuntime.ApplicationID, 10)
		nameStr := appRuntime.Name
		if runtimeIdToAppNameMap[appRuntime.ID] != nameStr {
			nameStr = runtimeIdToAppNameMap[appRuntime.ID] + "/" + nameStr
		}

		data.List = append(data.List, list.Item{
			ID:             idStr,
			Title:          nameStr,
			MainState:      getMainState(appRuntime.Status),
			TitleState:     getTitleState(p.Sdk, appRuntime.RawDeploymentStatus, deployIdStr, appIdStr, appRuntime.DeleteStatus),
			KvInfos:        getKvInfos(p.Sdk, runtimeIdToAppNameMap[appRuntime.ID], uidToName[appRuntime.Creator], appRuntime.DeploymentOrderName, appRuntime.ReleaseVersion, healthStr, appRuntime),
			Selectable:     true,
			Operations:     getOperations(appRuntime.ProjectID, appRuntime.ApplicationID, appRuntime.ID),
			MoreOperations: getMoreOperations(p.Sdk, fmt.Sprintf("%d", appRuntime.ID)),
		})
		ids = append(ids, idStr)
		runtimeMap[idStr] = appRuntime
	}

	data.Operations = p.getBatchOperation(p.Sdk, ids)
	// filter by name and advanced condition
	var advancedFilter map[string][]string
	var nameFilter map[string]string
	if advancedFilterMap, ok := (*p.Sdk.GlobalState)["advanceFilter"]; ok {
		err := common.Transfer(advancedFilterMap, &advancedFilter)
		if err != nil {
			logrus.Errorf("parse advanced filter failed err :%v", err)
		}
		for k, v := range advancedFilter {
			if len(v) == 0 {
				delete(advancedFilter, k)
			}
		}
	}
	if nameFilterMap, ok := (*p.Sdk.GlobalState)["nameFilter"]; ok {
		err := common.Transfer(nameFilterMap, &nameFilter)
		if err != nil {
			logrus.Errorf("parse input filter failed err :%v", err)
		}
	}
	var filterName string
	filterName, ok := nameFilter[common.FilterInputCondition]
	if !ok {
		filterName = ""
	}
	logrus.Infof("inputFilter: %v", filterName)
	logrus.Infof("advanceFilter: %#v", advancedFilter)
	var needFilter = data.List
	data.List = make([]list.Item, 0)
	for i := 0; i < len(needFilter); i++ {
		runtime := runtimeMap[needFilter[i].ID]
		if filterName != "" {
			if !strings.Contains(needFilter[i].Title, filterName) {
				continue
			}
		}
		if p.doFilter(advancedFilter, runtime, runtime.LastOperateTime.UnixNano()/1e6, runtimeIdToAppNameMap[runtime.ID], needFilter[i].KvInfos[1].Value) {
			data.List = append(data.List, needFilter[i])
		}
	}
	//logrus.Infof("list after filter: %#v", data.List)

	slice.Sort(data.List, func(i, j int) bool {
		return data.List[i].Title < data.List[j].Title
	})
	data.Total = uint64(len(data.List))
	start := (p.PageNo - 1) * p.PageSize
	if start >= data.Total {
		start = 0
	}
	//logrus.Infof("list after sort: %#v", data.List)

	end := uint64(math.Min(float64((p.PageNo)*p.PageSize), float64(data.Total)))
	data.List = data.List[start:end]

	return data
}

func (p *List) doFilter(conds map[string][]string, appRuntime bundle.GetApplicationRuntimesDataEle, deployAt int64, appName, deploymentOrderName string) bool {
	if conds == nil || len(conds) == 0 {
		return true
	}
	for k, v := range conds {
		switch k {
		case common.FilterApp:
			for _, value := range v {
				if appName == value {
					return true
				}
			}
		//case common.FilterRuntimeStatus:
		//	for _, value := range v {
		//		if value == appRuntime.RawStatus {
		//			return true
		//		}
		//	}
		case common.FilterDeployStatus:
			for _, value := range v {
				if value == appRuntime.RawDeploymentStatus {
					return true
				}
			}
		case common.FilterDeployOrderName:
			for _, value := range v {
				if deploymentOrderName == value {
					return true
				}
			}
		case common.FilterDeployTime:
			startTime, err := strconv.ParseInt(v[0], 10, 64)
			if err != nil {
				logrus.Errorf("parse filter time range failed ,err :%v", err)
			}
			endTime, err := strconv.ParseInt(v[1], 10, 64)
			if err != nil {
				logrus.Errorf("parse filter time range failed ,err :%v", err)
			}
			if startTime <= deployAt && endTime >= deployAt {
				return true
			}
		}
	}
	return false
}

func getMainState(runtimeStatus string) *list.StateInfo {
	var (
		statusStr list.ItemCommStatus
	)
	switch runtimeStatus {
	case apistructs.RuntimeStatusHealthy:
		statusStr = common.FrontedStatusSuccess
	case apistructs.RuntimeStatusUnHealthy:
		statusStr = common.FrontedStatusError
	default:
		statusStr = common.FrontedStatusDefault
	}

	return &list.StateInfo{
		Status: statusStr,
	}
}

func getTitleState(sdk *cptype.SDK, deployStatus, deploymentId, appId, dStatus string) []list.StateInfo {
	if dStatus == "" {
		var deployStr list.ItemCommStatus
		switch deployStatus {
		case string(apistructs.DeploymentStatusInit):
			deployStr = common.FrontedStatusProcessing
		case string(apistructs.DeploymentStatusOK):
			deployStr = common.FrontedStatusSuccess
		case string(apistructs.DeploymentStatusFailed):
			deployStr = common.FrontedStatusError
		case string(apistructs.DeploymentStatusCanceling):
			deployStr = common.FrontedStatusWarning
		case string(apistructs.DeploymentStatusCanceled):
			deployStr = common.FrontedStatusDefault
		}
		return []list.StateInfo{
			{
				Status:     deployStr,
				Text:       sdk.I18n(deployStatus),
				SuffixIcon: "right",
				Operations: map[cptype.OperationKey]cptype.Operation{
					"click": {
						SkipRender: true,
						ServerData: &cptype.OpServerData{
							"logId": deploymentId,
							"appId": appId,
						},
					},
				},
			},
		}
	} else {
		return []list.StateInfo{
			{
				Status:     common.FrontedStatusProcessing,
				Text:       sdk.I18n(dStatus),
				Operations: nil,
			},
		}
	}
}

func getOperations(projectId, appId, runtimeId uint64) map[cptype.OperationKey]cptype.Operation {
	projectIdStr := fmt.Sprintf("%d", projectId)
	appIdStr := fmt.Sprintf("%d", appId)
	runtimeIdStr := fmt.Sprintf("%d", runtimeId)
	return map[cptype.OperationKey]cptype.Operation{
		"clickGoto": {
			ServerData: &cptype.OpServerData{
				"target": "runtimeDetailRoot",
				"params": map[string]string{
					"projectId": projectIdStr,
					"appId":     appIdStr,
					"runtimeId": runtimeIdStr,
				},
			},
		},
	}
}

// getBatchOperation different runtime need different operation.
func (p List) getBatchOperation(sdk *cptype.SDK, ids []string) map[cptype.OperationKey]cptype.Operation {
	return map[cptype.OperationKey]cptype.Operation{
		"changePage": {},
		"batchRowsHandle": {
			ServerData: &cptype.OpServerData{
				"options": []list.OpBatchRowsHandleOptionServerData{
					{
						AllowedRowIDs: ids, Icon: "chongxinqidong", ID: common.ReStartOp, Text: sdk.I18n("restart"), // allowedRowIDs = null 或不传这个key，表示所有都可选，allowedRowIDs=[]表示当前没有可选择，此处应该可以不传
					},
					{
						AllowedRowIDs: ids, Icon: "remove", ID: common.DeleteOp, Text: sdk.I18n("delete"),
					},
				},
			},
		},
	}
}
func getMoreOperations(sdk *cptype.SDK, id string) []list.MoreOpItem {
	return []list.MoreOpItem{
		{
			ID:   common.DeleteOp,
			Icon: "remove",
			Text: sdk.I18n("delete"),
			Operations: map[cptype.OperationKey]cptype.Operation{
				"click": {
					Confirm:    sdk.I18n("delete confirm") + "?",
					ClientData: &cptype.OpClientData{},
				},
			},
		},
		{
			ID:   common.ReStartOp,
			Icon: "chongxinqidong",
			Text: sdk.I18n("restart"),
			Operations: map[cptype.OperationKey]cptype.Operation{
				"click": {
					ClientData: &cptype.OpClientData{},
				},
			},
		},
		//{
		//	ID:   id,
		//	Icon: "shuaxin",
		//	// todo
		//	Text: sdk.I18n("delete"),
		//	Operations: map[cptype.OperationKey]cptype.Operation{
		//		"click": {
		//			// todo
		//			Confirm:    sdk.I18n("delete confirm"),
		//			ClientData: &cptype.OpClientData{},
		//		},
		//	},
		//},
	}
}

func getKvInfos(sdk *cptype.SDK, appName, creatorName, deployOrderName, deployVersion, healthStr string, runtime bundle.GetApplicationRuntimesDataEle) []list.KvInfo {
	days := time.Now().Sub(runtime.CreatedAt).Hours() / float64(24)
	kvs := []list.KvInfo{
		{
			Key:   sdk.I18n("appName"),
			Value: appName,
		},
	}
	if deployOrderName != "" {
		kvs = append(kvs, list.KvInfo{
			Key:   sdk.I18n("deploymentOrderName"),
			Value: deployOrderName,
			Tip:   deployVersion,
		})
	}
	if healthStr != "" {
		kvs = append(kvs, list.KvInfo{
			Key:   sdk.I18n("service"),
			Value: healthStr,
		})
	}
	kvs = append(kvs, []list.KvInfo{
		{
			Key:   sdk.I18n("deployer"),
			Value: creatorName,
		},
		{
			Key:   sdk.I18n("running duration"),
			Value: fmt.Sprintf("%d", int64(days)) + sdk.I18n("day"),
		},
		{
			Key:   sdk.I18n("deployAt"),
			Value: runtime.LastOperateTime.Format("2006-01-02 15:04:05"),
		},
	}...)
	return kvs
}

func init() {
	base.InitProviderWithCreator(common.ScenarioKey, "list", func() servicehub.Provider {
		return &List{}
	})
}
