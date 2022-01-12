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
	"reflect"
	"testing"
	"time"

	"github.com/erda-project/erda-infra/providers/component-protocol/components/list"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-runtime/common"
)

type NopTranslator struct{}

func (t NopTranslator) Get(lang i18n.LanguageCodes, key, def string) string { return key }

func (t NopTranslator) Text(lang i18n.LanguageCodes, key string) string { return key }

func (t NopTranslator) Sprintf(lang i18n.LanguageCodes, key string, args ...interface{}) string {
	return fmt.Sprintf(key, args...)
}

var defaultSDK = &cptype.SDK{

	GlobalState: &cptype.GlobalStateData{},
	Tran:        &NopTranslator{},
}

func TestList_doFilter(t *testing.T) {
	type args struct {
		conditions          map[string][]string
		appRuntime          *bundle.GetApplicationRuntimesDataEle
		deployAt            int64
		appName             string
		deploymentOrderName string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "1",
			args: args{
				conditions: map[string][]string{common.FilterApp: {"1"}},
			},
			want: false,
		},
		{
			name: "2",
			args: args{
				conditions: map[string][]string{
					common.FilterDeployStatus: {""},
				}, appRuntime: &bundle.GetApplicationRuntimesDataEle{
					RawDeploymentStatus: "1",
				}},
			want: false,
		},
		{
			name: "3",
			args: args{conditions: map[string][]string{common.FilterDeployOrderName: {""}}},
			want: true,
		},
		{
			name: "4",
			args: args{conditions: map[string][]string{common.FilterDeployTime: {"0", "0"}}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &List{}
			if got := p.doFilter(tt.args.conditions, tt.args.appRuntime, tt.args.deployAt, tt.args.appName, tt.args.deploymentOrderName); got != tt.want {
				t.Errorf("doFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getTitleState(t *testing.T) {
	type args struct {
		sdk          *cptype.SDK
		deployStatus string
		deploymentId string
		appId        string
	}
	tests := []struct {
		name string
		args args
		want []list.StateInfo
	}{
		{
			name: "1",
			args: args{
				sdk:          defaultSDK,
				deployStatus: string(apistructs.DeploymentStatusInit),
			},
		},
		{
			name: "1",
			args: args{
				sdk:          defaultSDK,
				deployStatus: string(apistructs.DeploymentStatusOK),
			},
		},
		{
			name: "1",
			args: args{
				sdk:          defaultSDK,
				deployStatus: string(apistructs.DeploymentStatusFailed),
			},
		},
		{
			name: "1",
			args: args{
				sdk:          defaultSDK,
				deployStatus: string(apistructs.DeploymentStatusCanceling),
			},
		},
		{
			name: "1",
			args: args{
				sdk:          defaultSDK,
				deployStatus: string(apistructs.DeploymentStatusCanceled),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getTitleState(tt.args.sdk, tt.args.deployStatus, tt.args.deploymentId, tt.args.appId)
		})
	}
}

func Test_getMainState(t *testing.T) {
	type args struct {
		runtimeStatus string
	}
	tests := []struct {
		name string
		args args
		want *list.StateInfo
	}{
		{
			name: "1",
			args: args{runtimeStatus: apistructs.RuntimeStatusHealthy},
		},
		{
			name: "2",
			args: args{runtimeStatus: apistructs.RuntimeStatusUnHealthy},
		},
		{
			name: "3",
			args: args{runtimeStatus: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getMainState(tt.args.runtimeStatus)
		})
	}
}

func Test_getOperations(t *testing.T) {
	type args struct {
		projectId uint64
		appId     uint64
		runtimeId uint64
	}
	tests := []struct {
		name string
		args args
		want map[cptype.OperationKey]cptype.Operation
	}{
		{
			name: "1",
			args: args{
				projectId: 0,
				appId:     0,
				runtimeId: 0,
			},
			want: map[cptype.OperationKey]cptype.Operation{
				"clickGoto": {
					ServerData: &cptype.OpServerData{
						"target": "runtimeDetailRoot",
						"params": map[string]string{
							"projectId": "0",
							"appId":     "0",
							"runtimeId": "0",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getOperations(tt.args.projectId, tt.args.appId, tt.args.runtimeId); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getOperations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestList_getBatchOperation(t *testing.T) {
	type args struct {
		sdk *cptype.SDK
		ids []string
	}
	tests := []struct {
		name string
		args args
		want map[cptype.OperationKey]cptype.Operation
	}{
		{
			name: "1",
			args: args{
				sdk: defaultSDK,
				ids: []string{"1"},
			},
			want: map[cptype.OperationKey]cptype.Operation{
				"changePage": {},
				"batchRowsHandle": {
					ServerData: &cptype.OpServerData{
						"options": []list.OpBatchRowsHandleOptionServerData{
							{
								AllowedRowIDs: []string{"1"}, Icon: "chongxinqidong", ID: common.ReStartOp, Text: "重启", // allowedRowIDs = null 或不传这个key，表示所有都可选，allowedRowIDs=[]表示当前没有可选择，此处应该可以不传
							},
							{
								AllowedRowIDs: []string{"1"}, Icon: "remove", ID: common.DeleteOp, Text: "删除",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := List{}
			if got := p.getBatchOperation(tt.args.sdk, tt.args.ids); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getBatchOperation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getKvInfos(t *testing.T) {
	type args struct {
		sdk             *cptype.SDK
		appName         string
		creatorName     string
		deployOrderName string
		deployVersion   string
		healthStr       string
		runtime         *bundle.GetApplicationRuntimesDataEle
	}
	tests := []struct {
		name string
		args args
		want []list.KvInfo
	}{
		// TODO: Add test cases.
		{
			name: "1",
			args: args{
				sdk:             defaultSDK,
				appName:         "",
				creatorName:     "",
				deployOrderName: "",
				deployVersion:   "",
				healthStr:       "",
				runtime:         &bundle.GetApplicationRuntimesDataEle{LastOperateTime: time.Now()},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getKvInfos(tt.args.sdk, tt.args.appName, tt.args.creatorName, tt.args.deployOrderName, tt.args.deployVersion, tt.args.healthStr, tt.args.runtime)
		})
	}
}
