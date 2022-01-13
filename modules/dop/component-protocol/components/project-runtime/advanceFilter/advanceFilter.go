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
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/filter"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/filter/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-runtime/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
)

// todo placeholder
var placeHolders = map[string]string{
	"deploymentStatus":    "select status",
	"runtimeStatus":       "select status",
	"app":                 "select app",
	"deployTime":          "select deployTime",
	"deploymentOrderName": "select deploymentOrderName",
}

type AdvanceFilter struct {
	base.DefaultProvider
	bdl *bundle.Bundle
	impl.DefaultFilter
	Values cptype.ExtraMap
}

type Option struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
type Condition struct {
	Key         string   `json:"key"`
	Label       string   `json:"label"`
	Placeholder string   `json:"placeholder"`
	Type        string   `json:"type"`
	Options     []Option `json:"options"`
}

func (af *AdvanceFilter) RegisterFilterOp(opData filter.OpFilter) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		af.Values = make(cptype.ExtraMap)
		err := common.Transfer(opData.ClientData.Values, &af.Values)
		if err != nil {
			return
		}
		(*sdk.GlobalState)["advanceFilter"] = af.Values
		urlParam, err := af.generateUrlQueryParams(af.Values)
		if err != nil {
			return
		}
		(*af.StdStatePtr)["inputFilter__urlQuery"] = urlParam
		af.StdDataPtr = af.getData(sdk)
	}
}

func (af *AdvanceFilter) generateUrlQueryParams(Values cptype.ExtraMap) (string, error) {
	fb, err := json.Marshal(Values)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(fb), nil
}

func (af *AdvanceFilter) RegisterFilterItemSaveOp(opData filter.OpFilterItemSave) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {

	}
}

func (af *AdvanceFilter) RegisterFilterItemDeleteOp(opData filter.OpFilterItemDelete) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {

	}
}
func (af *AdvanceFilter) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return af.RegisterInitializeOp()
}

func (af *AdvanceFilter) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		err := common.Transfer(sdk.Comp.State, af.StdStatePtr)
		if err != nil {
			return
		}
		if urlquery := sdk.InParams.String("inputFilter__urlQuery"); urlquery != "" {
			if err = af.flushOptsByFilter(urlquery); err != nil {
				logrus.Errorf("failed to transfer values in component advance filter")
				return
			}
		}
		(*sdk.GlobalState)["advanceFilter"] = af.Values
		af.StdDataPtr = af.getData(sdk)
	}
}

func (af *AdvanceFilter) flushOptsByFilter(filterEntity string) error {
	b, err := base64.StdEncoding.DecodeString(filterEntity)
	if err != nil {
		return err
	}
	v := cptype.ExtraMap{}
	err = json.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	af.Values = v
	return nil
}
func (af *AdvanceFilter) BeforeHandleOp(sdk *cptype.SDK) {
	af.bdl = sdk.Ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
}

func (af *AdvanceFilter) getData(sdk *cptype.SDK) *filter.Data {
	data := &filter.Data{}
	getEnv, ok := sdk.InParams["env"].(string)
	if !ok {
		logrus.Errorf("env is empty")
		return data
	}
	projectId, err := strconv.ParseUint(sdk.InParams["projectId"].(string), 10, 64)
	if err != nil {
		logrus.Errorf("parse oid failed,%v", err)
		return data
	}
	oid, err := strconv.ParseUint(sdk.Identity.OrgID, 10, 64)
	if err != nil {
		logrus.Errorf("parse oid failed,%v", err)
		return data
	}
	apps, err := af.bdl.GetAppsByProject(projectId, oid, sdk.Identity.UserID)
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
	runtimesByApp, err := af.bdl.ListRuntimesGroupByApps(oid, sdk.Identity.UserID, appIds)
	if err != nil {
		logrus.Errorf("get my app failed,%v", err)
		return data
	}
	// status condition
	appNameMap := make(map[string]bool)
	deploymentStatusMap := make(map[string]bool)
	//runtimeStatusMap := make(map[string]bool)
	deploymentOrderNameMap := make(map[string]bool)
	runtimeIdToAppNameMap := make(map[uint64]string)
	selectRuntimes := make([]bundle.GetApplicationRuntimesDataEle, 0)

	for _, v := range runtimesByApp {
		for _, appRuntime := range v {
			if getEnv == appRuntime.Extra.Workspace {
				if appRuntime.DeploymentOrderName != "" {
					deploymentOrderNameMap[appRuntime.DeploymentOrderName] = true
				}
				//runtimeStatusMap[appRuntime.RawStatus] = true
				deploymentStatusMap[appRuntime.RawDeploymentStatus] = true
				selectRuntimes = append(selectRuntimes, *appRuntime)
				if appIdToName[appRuntime.ApplicationID] != "" {
					runtimeIdToAppNameMap[appRuntime.ID] = appIdToName[appRuntime.ApplicationID]
					appNameMap[appIdToName[appRuntime.ApplicationID]] = true
				}
			}
		}
	}
	// set runtimes in global state
	(*sdk.GlobalState)["runtimes"] = selectRuntimes
	// runtimeNameToAppName
	(*sdk.GlobalState)["runtimeIdToAppName"] = runtimeIdToAppNameMap

	// filter values

	var conds []Condition
	conds = append(conds, getSelectCondition(sdk, deploymentStatusMap, "deploymentStatus"))
	//conds = append(conds, getSelectCondition(sdk, runtimeStatusMap, "runtimeStatus"))
	conds = append(conds, getSelectCondition(sdk, appNameMap, "app"))
	conds = append(conds, getSelectCondition(sdk, deploymentOrderNameMap, "deploymentOrderName"))
	conds = append(conds, getRangeCondition(sdk, "deployTime"))
	err = common.Transfer(conds, &data.Conditions)
	if err != nil {
		return nil
	}
	data.Operations = af.getOperation()
	return data
}

func (af *AdvanceFilter) getOperation() map[cptype.OperationKey]cptype.Operation {
	return map[cptype.OperationKey]cptype.Operation{
		"filter": {},
	}
}
func getSelectCondition(sdk *cptype.SDK, keys map[string]bool, key string) Condition {

	c := Condition{
		Key:         key,
		Label:       sdk.I18n(key),
		Placeholder: sdk.I18n(placeHolders[key]),
		Type:        "select",
	}
	for k := range keys {
		c.Options = append(c.Options, Option{
			Label: sdk.I18n(k),
			Value: k,
		})
	}
	return c
}

func getRangeCondition(sdk *cptype.SDK, key string) Condition {
	c := Condition{
		Key:         key,
		Label:       sdk.I18n(key),
		Placeholder: sdk.I18n(placeHolders[key]),
		Type:        "dateRange",
	}
	return c
}

func (af *AdvanceFilter) Init(ctx servicehub.Context) error {
	return af.DefaultProvider.Init(ctx)
}

func init() {
	base.InitProviderWithCreator(common.ScenarioKey, "advanceFilter", func() servicehub.Provider {
		return &AdvanceFilter{}
	})
}
