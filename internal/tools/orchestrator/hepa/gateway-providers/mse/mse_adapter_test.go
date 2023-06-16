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

package mse

import (
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	mseclient "github.com/alibabacloud-go/mse-20190531/v3/client"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	gateway_providers "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers"
	. "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	mseCommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/dto"
)

func TestNewMseAdapter(t *testing.T) {
	type args struct {
		az string
	}
	tests := []struct {
		name    string
		args    args
		want    gateway_providers.GatewayAdapter
		wantErr bool
	}{
		{
			name: "Test_01",
			args: args{
				az: "test",
			},
			want: &MseAdapterImpl{
				ProviderName:    mseCommon.MseProviderName,
				Bdl:             &bundle.Bundle{},
				AccessKeyID:     "aliyunaccesskeyid",
				AccessKeySecret: "aliyunaccesskeysecret",
				GatewayUniqueID: "aliyunmsegatewayid",
				GatewayEndpoint: "mse.cn-hangzhou.aliyuncs.com",
				ClusterName:     "test",
			},
		},
		{
			name: "Test_02",
			args: args{
				az: "test",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test_03",
			args: args{
				az: "test",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test_04",
			args: args{
				az: "test",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &MseAdapterImpl{
				Bdl:             &bundle.Bundle{},
				GatewayEndpoint: "mse.cn-hangzhou.aliyuncs.com",
			}

			monkey.PatchInstanceMethod(reflect.TypeOf(adapter), "GetMSEPluginsByAPI", func(_ *MseAdapterImpl, name *string, category *int32, enableOnly *bool) ([]*mseclient.GetPluginsResponseBodyData, error) {
				pluginName := "key-auth"
				var id int64 = 3

				return []*mseclient.GetPluginsResponseBodyData{
					{
						Id:   &id,
						Name: &pluginName,
					},
				}, nil
			})

			monkey.PatchInstanceMethod(reflect.TypeOf(adapter.Bdl), "QueryClusterInfo", func(_ *bundle.Bundle, _ string) (apistructs.ClusterInfoData, error) {
				if tt.name == "Test_02" {
					return apistructs.ClusterInfoData{"ALIYUN_ACCESS_KEY_SECRET": "aliyunaccesskeysecret", "ALIYUN_MSE_GATEWAY_ID": "aliyunmsegatewayid"}, nil
				}
				if tt.name == "Test_03" {
					return apistructs.ClusterInfoData{"ALIYUN_ACCESS_KEY_ID": "aliyunaccesskeyid", "ALIYUN_MSE_GATEWAY_ID": "aliyunmsegatewayid"}, nil
				}
				if tt.name == "Test_04" {
					return apistructs.ClusterInfoData{"ALIYUN_ACCESS_KEY_ID": "aliyunaccesskeyid", "ALIYUN_ACCESS_KEY_SECRET": "aliyunaccesskeysecret"}, nil
				}

				return apistructs.ClusterInfoData{"ALIYUN_ACCESS_KEY_ID": "aliyunaccesskeyid", "ALIYUN_ACCESS_KEY_SECRET": "aliyunaccesskeysecret", "ALIYUN_MSE_GATEWAY_ID": "aliyunmsegatewayid"}, nil
			})
			defer monkey.UnpatchAll()

			got, err := NewMseAdapter(tt.args.az)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMseAdapter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.name == "Test_01" {
				gotResult, ok := got.(*MseAdapterImpl)
				if !ok {
					t.Errorf("NewMseAdapter() got = %+v, want %+v", got, tt.want)
				}
				wantResult, _ := tt.want.(*MseAdapterImpl)

				gotResult.Bdl = wantResult.Bdl

				if !reflect.DeepEqual(gotResult, wantResult) {
					t.Errorf("NewMseAdapter() got = %+v, want %+v", got, tt.want)
				}
			}
		})
	}
}

func TestMseAdapterImpl_GatewayProviderExist(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Test_01",
			fields: fields{
				ProviderName: "MSE",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}

			monkey.PatchInstanceMethod(reflect.TypeOf(impl), "GetMSEGatewayByAPI", func(_ *MseAdapterImpl) (*mseclient.GetGatewayResponseBodyData, error) {
				return &mseclient.GetGatewayResponseBodyData{}, nil
			})
			defer monkey.UnpatchAll()

			if got := impl.GatewayProviderExist(); got != tt.want {
				t.Errorf("GatewayProviderExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMseAdapterImpl_GetVersion(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name:    "Test_01",
			fields:  fields{ProviderName: "MSE"},
			want:    mseCommon.MseVersion,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			got, err := impl.GetVersion()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetVersion() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMseAdapterImpl_CheckPluginEnabled(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		pluginName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "Test_01",
			fields:  fields{ProviderName: "MSE"},
			args:    args{pluginName: "cors"},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			got, err := impl.CheckPluginEnabled(tt.args.pluginName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPluginEnabled() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckPluginEnabled() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMseAdapterImpl_DeleteConsumer(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Test_01",
			fields:  fields{ProviderName: "MSE"},
			args:    args{id: "1"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			if err := impl.DeleteConsumer(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteConsumer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMseAdapterImpl_CreateOrUpdateRoute(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		req *RouteReqDto
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *RouteRespDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{req: &RouteReqDto{
				Protocols:     []string{"http", "https"},
				Methods:       []string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"},
				Hosts:         []string{"test.io"},
				Paths:         []string{"/test"},
				StripPath:     nil,
				PreserveHost:  nil,
				Service:       &Service{Id: "1"},
				RegexPriority: 0,
				RouteId:       "1",
				PathHandling:  nil,
				Tags:          nil,
			}},
			want: &RouteRespDto{
				Id:        "1",
				Protocols: []string{"http", "https"},
				Methods:   []string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"},
				Hosts:     []string{"test.io"},
				Paths:     []string{"/test"},
				Service:   Service{Id: "1"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			got, err := impl.CreateOrUpdateRoute(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateOrUpdateRoute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateOrUpdateRoute() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMseAdapterImpl_DeleteRoute(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		routeId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Test_01",
			fields:  fields{ProviderName: "MSE"},
			args:    args{routeId: "1"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			if err := impl.DeleteRoute(tt.args.routeId); (err != nil) != tt.wantErr {
				t.Errorf("DeleteRoute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMseAdapterImpl_UpdateRoute(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		req *RouteReqDto
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *RouteRespDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{req: &RouteReqDto{
				Protocols:     []string{"http", "https"},
				Methods:       []string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"},
				Hosts:         []string{"test.io"},
				Paths:         []string{"/test"},
				StripPath:     nil,
				PreserveHost:  nil,
				Service:       &Service{Id: "1"},
				RegexPriority: 0,
				RouteId:       "1",
				PathHandling:  nil,
				Tags:          nil,
			}},
			want:    &RouteRespDto{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			got, err := impl.UpdateRoute(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateRoute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateRoute() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMseAdapterImpl_CreateCredential(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		req *CredentialReqDto
	}
	tTime := time.Now().Unix()
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CredentialDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{
				req: &CredentialReqDto{
					ConsumerId: "1",
					PluginName: "cors",
					Config: &CredentialDto{
						ConsumerId:   "1",
						CreatedAt:    0,
						Id:           "1",
						Key:          "xx",
						RedirectUrl:  "/test",
						RedirectUrls: []string{"/test"},
						Name:         "xxx",
						ClientId:     "1",
						ClientSecret: "xxyyzz",
						Secret:       "xxxyyy",
						Username:     "aaa",
					},
				},
			},
			want: &CredentialDto{
				ConsumerId:   "1",
				CreatedAt:    tTime,
				Id:           "1",
				Key:          "xx",
				RedirectUrl:  "/test",
				RedirectUrls: []string{"/test"},
				Name:         "cors",
				ClientId:     "1",
				ClientSecret: "xxyyzz",
				Secret:       "xxxyyy",
				Username:     "aaa",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			got, err := impl.CreateCredential(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCredential() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateCredential() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMseAdapterImpl_CreateOrUpdatePluginById(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		req *PluginReqDto
	}
	reqConfig := make(map[string]interface{})
	consumers := make([]dto.Consumers, 0)
	consumers = append(consumers, dto.Consumers{
		Name:       "abc",
		Credential: "xxxxyyyy",
	})
	reqConfig["whitelist"] = consumers
	tTime := time.Now().Unix()
	enable := true
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *PluginRespDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: mseCommon.MseProviderName},
			args: args{
				req: &PluginReqDto{
					Name:       "xxx",
					ServiceId:  "1",
					RouteId:    "1",
					ConsumerId: "1",
					Route:      &KongObj{Id: "1"},
					Service:    &KongObj{Id: "1"},
					Consumer:   &KongObj{Id: "1"},
					Config:     nil,
					Enabled:    &enable,
					Id:         "1",
					PluginId:   "1",
				},
			},
			want: &PluginRespDto{
				Id:         "1",
				ServiceId:  "1",
				RouteId:    "1",
				ConsumerId: "1",
				Route:      &KongObj{Id: "1"},
				Service:    &KongObj{Id: "1"},
				Consumer:   &KongObj{Id: "1"},
				Name:       "xxx",
				Config:     nil,
				Enabled:    true,
				CreatedAt:  tTime,
				PolicyId:   "1",
			},
			wantErr: false,
		},
		{
			name:   "Test_02",
			fields: fields{ProviderName: mseCommon.MseProviderName},
			args: args{
				req: &PluginReqDto{
					Name:       mseCommon.MsePluginKeyAuth,
					ServiceId:  "1",
					RouteId:    "1",
					ConsumerId: "1",
					Route:      &KongObj{Id: "1"},
					Service:    &KongObj{Id: "1"},
					Consumer:   &KongObj{Id: "1"},
					Config:     reqConfig,
					Enabled:    &enable,
					Id:         "1",
					PluginId:   "1",
				},
			},
			want: &PluginRespDto{
				Id:         "1",
				ServiceId:  "1",
				RouteId:    "1",
				ConsumerId: "1",
				Route:      &KongObj{Id: "1"},
				Service:    &KongObj{Id: "1"},
				Consumer:   &KongObj{Id: "1"},
				Name:       mseCommon.MsePluginKeyAuth,
				Config:     reqConfig,
				Enabled:    true,
				CreatedAt:  tTime,
				PolicyId:   "1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}

			var keyAuthID int64 = 1

			mseCommon.MapClusterNameToMSEPluginNameToPluginID[impl.ClusterName] = make(map[string]*int64)
			mseCommon.MapClusterNameToMSEPluginNameToPluginID[impl.ClusterName][mseCommon.MsePluginKeyAuth] = &keyAuthID

			monkey.PatchInstanceMethod(reflect.TypeOf(impl), "GetMSEPluginConfigByIDByAPI", func(impl *MseAdapterImpl, pluginId *int64) (*mseclient.GetPluginConfigResponseBodyData, error) {
				cfList := make([]*mseclient.GetPluginConfigResponseBodyDataGatewayConfigList, 0)
				config := "consumers: \n# 注意！该凭证仅做示例使用，请勿用于具体业务，造成安全风险\n- credential: 5d1a401b3aef4ee5bc39a99654672f92\n  name: bbb\n- credential: 5d1a401b3aef4ee5bc39a99654672f91\n  name: aaa\nkeys:\n- apikey\n- x-api-key\n_rules_:\n# 规则一：按路由名称匹配生效\n- _match_route_:\n  - project-5846-test-dice-test-5846-api-754bbce0af034774ac2b8f74c7e070a6-0149aa-058a61f1-ccaae83a5fe7846d18d4ac80940a8fdc3\n  allow:\n  - aaa\n- _match_route_:\n  - project-5846-test-dice-test-5846-api-66f56a64312143cfbbbd04eca82eca56-bff529-51ec627a-ccaae83a5fe7846d18d4ac80940a8fdc3\n  allow:\n  - bbb\n  \n"
				var configLevel int32 = 0
				var id int64 = 154
				enabled := true
				cfList = append(cfList, &mseclient.GetPluginConfigResponseBodyDataGatewayConfigList{
					Config:      &config,
					ConfigLevel: &configLevel,
					Id:          &id,
					Enable:      &enabled,
				})
				return &mseclient.GetPluginConfigResponseBodyData{
					GatewayConfigList: cfList,
				}, nil
			})

			monkey.PatchInstanceMethod(reflect.TypeOf(impl), "UpdateMSEPluginConfigByIDByAPI", func(impl *MseAdapterImpl, pluginId *int64, configId *int64, config *string, configLevel *int32, enable *bool) (*mseclient.UpdatePluginConfigResponseBody, error) {
				return &mseclient.UpdatePluginConfigResponseBody{}, nil
			})

			monkey.PatchInstanceMethod(reflect.TypeOf(impl), "GetMSEGatewayRouteNameByZoneName", func(impl *MseAdapterImpl, zoneName string, domainName *string) (string, error) {
				return "xxxxyyyy", nil
			})

			defer monkey.UnpatchAll()

			got, err := impl.CreateOrUpdatePluginById(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateOrUpdatePluginById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				if !reflect.DeepEqual(*got, *tt.want) {
					t.Errorf("CreateOrUpdatePluginById() got = %v, want %v", *got, *tt.want)
				}
			}
		})
	}
}

func TestMseAdapterImpl_CreateOrUpdateService(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		req *ServiceReqDto
	}
	retry := 5
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ServiceRespDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{
				req: &ServiceReqDto{
					Name:           "xxx",
					Url:            "/test",
					Protocol:       "http",
					Host:           "test.io",
					Port:           8080,
					Path:           "/test",
					Retries:        &retry,
					ConnectTimeout: 1000,
					WriteTimeout:   500,
					ReadTimeout:    500,
					ServiceId:      "1",
				},
			},
			want: &ServiceRespDto{
				Id:       "1",
				Name:     "xxx",
				Protocol: "http",
				Host:     "test.io",
				Port:     8080,
				Path:     "/test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			got, err := impl.CreateOrUpdateService(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateOrUpdateService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateOrUpdateService() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMseAdapterImpl_DeleteService(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		serviceId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Test_01",
			fields:  fields{ProviderName: "MSE"},
			args:    args{serviceId: "1"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			if err := impl.DeleteService(tt.args.serviceId); (err != nil) != tt.wantErr {
				t.Errorf("DeleteService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMseAdapterImpl_DeletePluginIfExist(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		req *PluginReqDto
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{
				req: &PluginReqDto{
					Name:       "xxx",
					ServiceId:  "1",
					RouteId:    "1",
					ConsumerId: "1",
					Route:      &KongObj{Id: "1"},
					Service:    &KongObj{Id: "1"},
					Consumer:   &KongObj{Id: "1"},
					Config:     nil,
					Id:         "1",
					PluginId:   "1",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			if err := impl.DeletePluginIfExist(tt.args.req); (err != nil) != tt.wantErr {
				t.Errorf("DeletePluginIfExist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMseAdapterImpl_CreateOrUpdatePlugin(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		req *PluginReqDto
	}
	enable := true
	tTime := time.Now().Unix()
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *PluginRespDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{req: &PluginReqDto{
				Name:       "xxx",
				ServiceId:  "1",
				RouteId:    "1",
				ConsumerId: "1",
				Route:      &KongObj{Id: "1"},
				Service:    &KongObj{Id: "1"},
				Consumer:   &KongObj{Id: "1"},
				Enabled:    &enable,
				Config:     nil,
				Id:         "1",
				CreatedAt:  tTime,
				PluginId:   "1",
			}},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			got, err := impl.CreateOrUpdatePlugin(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateOrUpdatePlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateOrUpdatePlugin() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMseAdapterImpl_AddPlugin(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		req *PluginReqDto
	}
	tTime := time.Now().Unix()
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *PluginRespDto
		wantErr bool
	}{
		{
			name: "Test_01",
			fields: fields{
				ProviderName: "MSE",
			},
			args: args{
				req: &PluginReqDto{
					Name:       "xxx",
					ServiceId:  "1",
					RouteId:    "1",
					ConsumerId: "1",
					Route:      nil,
					Service:    nil,
					Consumer:   nil,
					Enabled:    nil,
					Config:     nil,
					Id:         "001",
					CreatedAt:  0,
					PluginId:   "x0001",
				},
			},
			want: &PluginRespDto{
				Id:         "001",
				ServiceId:  "1",
				RouteId:    "1",
				ConsumerId: "1",
				Route:      nil,
				Service:    nil,
				Consumer:   nil,
				Name:       "xxx",
				Config:     nil,
				Enabled:    true,
				CreatedAt:  tTime,
				PolicyId:   "x0001",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			got, err := impl.AddPlugin(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddPlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddPlugin() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMseAdapterImpl_PutPlugin(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		req *PluginReqDto
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *PluginRespDto
		wantErr bool
	}{
		{
			name:    "Test_01",
			fields:  fields{ProviderName: "MSE"},
			args:    args{},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			got, err := impl.PutPlugin(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("PutPlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PutPlugin() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMseAdapterImpl_UpdatePlugin(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		req *PluginReqDto
	}
	enable := true
	tTime := time.Now().Unix()
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *PluginRespDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{req: &PluginReqDto{
				Name:       "xxx",
				ServiceId:  "1",
				RouteId:    "1",
				ConsumerId: "1",
				Route:      &KongObj{Id: "1"},
				Service:    &KongObj{Id: "1"},
				Consumer:   &KongObj{Id: "1"},
				Enabled:    &enable,
				Config:     nil,
				Id:         "1",
				CreatedAt:  tTime,
				PluginId:   "1",
			}},
			want: &PluginRespDto{
				Id:         "1",
				ServiceId:  "1",
				RouteId:    "1",
				ConsumerId: "1",
				Route:      &KongObj{Id: "1"},
				Service:    &KongObj{Id: "1"},
				Consumer:   &KongObj{Id: "1"},
				Name:       "xxx",
				Config:     nil,
				Enabled:    true,
				CreatedAt:  tTime,
				PolicyId:   "1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			got, err := impl.UpdatePlugin(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdatePlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdatePlugin() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMseAdapterImpl_RemovePlugin(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Test_01",
			fields:  fields{ProviderName: "MSE"},
			args:    args{id: "1"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			if err := impl.RemovePlugin(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("RemovePlugin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMseAdapterImpl_DeleteCredential(t *testing.T) {
	type fields struct {
		Bdl             *bundle.Bundle
		ProviderName    string
		AccessKeyID     string
		AccessKeySecret string
		GatewayUniqueID string
		GatewayEndpoint string
	}
	type args struct {
		consumerId    string
		pluginName    string
		credentialStr string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test_01",
			fields: fields{
				Bdl:             nil,
				ProviderName:    "",
				AccessKeyID:     "",
				AccessKeySecret: "",
				GatewayUniqueID: "",
				GatewayEndpoint: "",
			},
			args: args{
				consumerId:    "",
				pluginName:    mseCommon.MsePluginKeyAuth,
				credentialStr: "xxxxyyyyzzzz",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				Bdl:             tt.fields.Bdl,
				ProviderName:    tt.fields.ProviderName,
				AccessKeyID:     tt.fields.AccessKeyID,
				AccessKeySecret: tt.fields.AccessKeySecret,
				GatewayUniqueID: tt.fields.GatewayUniqueID,
				GatewayEndpoint: tt.fields.GatewayEndpoint,
			}

			monkey.PatchInstanceMethod(reflect.TypeOf(impl), "GetMSEPluginConfigByIDByAPI", func(impl *MseAdapterImpl, pluginId *int64) (*mseclient.GetPluginConfigResponseBodyData, error) {
				cfList := make([]*mseclient.GetPluginConfigResponseBodyDataGatewayConfigList, 0)
				config := "consumers: \n# 注意！该凭证仅做示例使用，请勿用于具体业务，造成安全风险\n- credential: 5d1a401b3aef4ee5bc39a99654672f92\n  name: bbb\n- credential: 5d1a401b3aef4ee5bc39a99654672f91\n  name: aaa\nkeys:\n- apikey\n- x-api-key\n_rules_:\n# 规则一：按路由名称匹配生效\n- _match_route_:\n  - project-5846-test-dice-test-5846-api-754bbce0af034774ac2b8f74c7e070a6-0149aa-058a61f1-ccaae83a5fe7846d18d4ac80940a8fdc3\n  allow:\n  - aaa\n- _match_route_:\n  - project-5846-test-dice-test-5846-api-66f56a64312143cfbbbd04eca82eca56-bff529-51ec627a-ccaae83a5fe7846d18d4ac80940a8fdc3\n  allow:\n  - bbb\n  \n"
				var configLevel int32 = 0
				cfList = append(cfList, &mseclient.GetPluginConfigResponseBodyDataGatewayConfigList{
					Config:      &config,
					ConfigLevel: &configLevel,
				})
				return &mseclient.GetPluginConfigResponseBodyData{
					GatewayConfigList: cfList,
				}, nil
			})

			monkey.PatchInstanceMethod(reflect.TypeOf(impl), "UpdateMSEPluginConfigByIDByAPI", func(impl *MseAdapterImpl, pluginId *int64, configId *int64, config *string, configLevel *int32, enable *bool) (*mseclient.UpdatePluginConfigResponseBody, error) {
				return &mseclient.UpdatePluginConfigResponseBody{}, nil
			})

			defer monkey.UnpatchAll()

			if err := impl.DeleteCredential(tt.args.consumerId, tt.args.pluginName, tt.args.credentialStr); (err != nil) != tt.wantErr {
				t.Errorf("DeleteCredential() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMseAdapterImpl_GetCredentialList(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		consumerId string
		pluginName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CredentialListDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{
				consumerId: "1",
				pluginName: "cors",
			},
			want:    &CredentialListDto{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			got, err := impl.GetCredentialList(tt.args.consumerId, tt.args.pluginName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCredentialList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCredentialList() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMseAdapterImpl_CreateAclGroup(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		consumerId string
		customId   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{
				consumerId: "1",
				customId:   "1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			if err := impl.CreateAclGroup(tt.args.consumerId, tt.args.customId); (err != nil) != tt.wantErr {
				t.Errorf("CreateAclGroup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMseAdapterImpl_CreateUpstream(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		req *UpstreamDto
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *UpstreamDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{req: &UpstreamDto{
				Id:           "1",
				Name:         "xx",
				Healthchecks: HealthchecksDto{},
			}},
			want: &UpstreamDto{
				Id:           "1",
				Name:         "xx",
				Healthchecks: HealthchecksDto{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			got, err := impl.CreateUpstream(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUpstream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateUpstream() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMseAdapterImpl_GetUpstreamStatus(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		upstreamId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *UpstreamStatusRespDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args:   args{upstreamId: "1"},
			want: &UpstreamStatusRespDto{
				Data: []TargetDto{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			got, err := impl.GetUpstreamStatus(tt.args.upstreamId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUpstreamStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUpstreamStatus() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMseAdapterImpl_DeleteUpstreamTarget(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		upstreamId string
		targetId   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{
				upstreamId: "",
				targetId:   "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			if err := impl.DeleteUpstreamTarget(tt.args.upstreamId, tt.args.targetId); (err != nil) != tt.wantErr {
				t.Errorf("DeleteUpstreamTarget() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMseAdapterImpl_TouchRouteOAuthMethod(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Test_01",
			fields:  fields{ProviderName: "MSE"},
			args:    args{id: "1"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			if err := impl.TouchRouteOAuthMethod(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("TouchRouteOAuthMethod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMseAdapterImpl_GetRoutes(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []RouteRespDto
		wantErr bool
	}{
		{
			name:    "Test_01",
			fields:  fields{ProviderName: "MSE"},
			want:    []RouteRespDto{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			got, err := impl.GetRoutes()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRoutes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRoutes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMseAdapterImpl_GetRoutesWithTag(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		tag string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []RouteRespDto
		wantErr bool
	}{
		{
			name:    "Test_01",
			fields:  fields{},
			args:    args{tag: "xxx"},
			want:    []RouteRespDto{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			got, err := impl.GetRoutesWithTag(tt.args.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRoutesWithTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRoutesWithTag() got = %v, want %v", got, tt.want)
			}
		})
	}
}
