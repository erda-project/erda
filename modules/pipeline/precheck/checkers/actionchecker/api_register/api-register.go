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

package api_register

import (
	"context"
	"fmt"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/precheck/prechecktype"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/httpclientutil"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

type apiRegister struct{}

func New() *apiRegister {
	return &apiRegister{}
}

type ApiCheck struct {
	OrgId       string      `json:"orgId,omitempty"`
	ProjectId   string      `json:"projectId,omitempty"`
	Workspace   string      `json:"workspace,omitempty"`
	ClusterName string      `json:"clusterName,omitempty"`
	AppId       string      `json:"appId,omitempty"`
	RuntimeName string      `json:"runtimeName,omitempty"`
	ServiceName string      `json:"serviceName,omitempty"`
	Swagger     interface{} `json:"swagger,omitempty"`
}

type HttpResponse struct {
	Success bool   `json:"success,omitempty"`
	Err     ErrMsg `json:"err,omitempty"`
}

type ErrMsg struct {
	Code string `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

func (b *apiRegister) ActionType() pipelineyml.ActionType {
	return "api-register"
}

func (b *apiRegister) Check(ctx context.Context, data interface{}, itemsForCheck prechecktype.ItemsForCheck) (abort bool, messages []string) {
	// data type: pipelineyml.Action
	// actualAction, ok := data.(pipelineyml.Action)
	// if !ok {
	// 	abort = false
	// 	return
	// }

	// check dice.yml
	diceymlContent, _ := itemsForCheck.Files["dice.yml"]
	if diceymlContent == "" {
		abort = false
		return
	}

	havaApiGateway := CheckApiGateWay([]byte(diceymlContent), itemsForCheck.Labels[apistructs.LabelDiceWorkspace])
	if !havaApiGateway {
		abort = true
		messages = append(messages, fmt.Sprintf("not found addon api-gateway in your dice.yml"))
		return
	}
	abort = false

	// swagger.json compatibility check
	// if actualAction.Params != nil {
	// 	swagger, ok := actualAction.Params["swagger_path"]
	// 	if ok {
	// 		sjson, err := ioutil.ReadFile(swagger.(string))
	// 		if err != nil {
	// 			abort = false
	// 			//messages = append(messages, fmt.Sprintf("invalid param 'swagger_json', value: %s", swagger))
	// 			return
	// 		}
	// 		var swaggerJson interface{}
	// 		err = json.Unmarshal(sjson, &swaggerJson)
	// 		if err != nil {
	// 			abort = true
	// 			messages = append(messages, err.Error())
	// 		}
	// 		apiCheck := ApiCheck{
	// 			OrgId:       itemsForCheck.Labels[apistructs.LabelOrgID],
	// 			ProjectId:   itemsForCheck.Labels[apistructs.LabelProjectID],
	// 			Workspace:   itemsForCheck.Labels[apistructs.LabelDiceWorkspace],
	// 			ServiceName: actualAction.Params["service_name"].(string),
	// 			ClusterName: itemsForCheck.ClusterName,
	// 			AppId:       itemsForCheck.Labels[apistructs.LabelAppID],
	// 			RuntimeName: itemsForCheck.Labels[apistructs.LabelBranch],
	// 			Swagger:     sjson,
	// 		}
	// 		err = checkCompatibility(apiCheck)
	// 		if err != nil {
	// 			abort = true
	// 			messages = append(messages, fmt.Sprintf("check api compatibility failed, err:%s", err.Error()))
	// 			return
	// 		}
	// 	}
	// }

	return
}

func checkPlanExist(addons diceyml.AddOns) bool {
	for _, addon := range addons {
		if addon == nil {
			continue
		}
		if strings.Contains(addon.Plan, "api-gateway") {
			return true
		}
	}
	return false
}

func CheckApiGateWay(diceYml []byte, env string) bool {
	var diceymlWorkspace string
	switch env {
	case "PROD":
		diceymlWorkspace = "production"
	case "STAGING":
		diceymlWorkspace = "staging"
	case "TEST":
		diceymlWorkspace = "test"
	case "DEV":
		diceymlWorkspace = "development"
	}
	d, err := diceyml.New(diceYml, false)
	if err != nil {
		return false
	}
	if checkPlanExist(d.Obj().AddOns) {
		return true
	}

	envObject, ok := d.Obj().Environments[diceymlWorkspace]
	if ok && envObject != nil {
		if checkPlanExist(envObject.AddOns) {
			return true
		}
	}

	return false
}

func checkCompatibility(apiCheck ApiCheck) error {
	req := httpclient.New().Post(discover.Hepa(), httpclient.RetryOption{}).
		Path("/api/gateway/check-compatibility").JSONBody(apiCheck)
	if err := httpclientutil.DoJson(req, nil); err != nil {
		return err
	}
	return nil
}
