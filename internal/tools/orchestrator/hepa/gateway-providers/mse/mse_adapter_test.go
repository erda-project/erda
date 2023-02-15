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

	gateway_providers "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers"
	. "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/kong/dto"
)

func TestNewMseAdapter(t *testing.T) {
	tests := []struct {
		name string
		want gateway_providers.GatewayAdapter
	}{
		{
			name: "Test_01",
			want: &MseAdapterImpl{
				ProviderName: Mse_Provider_Name,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMseAdapter(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMseAdapter() = %v, want %v", got, tt.want)
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
		// TODO: Add test cases.
		{
			name:    "Test_01",
			fields:  fields{ProviderName: "MSE"},
			want:    Mse_Version,
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
		req *KongRouteReqDto
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *KongRouteRespDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{req: &KongRouteReqDto{
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
			want: &KongRouteRespDto{
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
		req *KongRouteReqDto
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *KongRouteRespDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{req: &KongRouteReqDto{
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
			want:    &KongRouteRespDto{},
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
		req *KongCredentialReqDto
	}
	tTime := time.Now().Unix()
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *KongCredentialDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{
				req: &KongCredentialReqDto{
					ConsumerId: "1",
					PluginName: "cors",
					Config: &KongCredentialDto{
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
			want: &KongCredentialDto{
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
		req *KongPluginReqDto
	}
	tTime := time.Now().Unix()
	enable := true
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *KongPluginRespDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{
				req: &KongPluginReqDto{
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
			want: &KongPluginRespDto{
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
			got, err := impl.CreateOrUpdatePluginById(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateOrUpdatePluginById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateOrUpdatePluginById() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMseAdapterImpl_CreateOrUpdateService(t *testing.T) {
	type fields struct {
		ProviderName string
	}
	type args struct {
		req *KongServiceReqDto
	}
	retry := 5
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *KongServiceRespDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{
				req: &KongServiceReqDto{
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
			want: &KongServiceRespDto{
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
		req *KongPluginReqDto
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
				req: &KongPluginReqDto{
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
		req *KongPluginReqDto
	}
	enable := true
	tTime := time.Now().Unix()
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *KongPluginRespDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{req: &KongPluginReqDto{
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
		req *KongPluginReqDto
	}
	tTime := time.Now().Unix()
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *KongPluginRespDto
		wantErr bool
	}{
		{
			name: "Test_01",
			fields: fields{
				ProviderName: "MSE",
			},
			args: args{
				req: &KongPluginReqDto{
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
			want: &KongPluginRespDto{
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
		req *KongPluginReqDto
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *KongPluginRespDto
		wantErr bool
	}{
		// TODO: Add test cases.
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
		req *KongPluginReqDto
	}
	enable := true
	tTime := time.Now().Unix()
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *KongPluginRespDto
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{req: &KongPluginReqDto{
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
			want: &KongPluginRespDto{
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
		ProviderName string
	}
	type args struct {
		consumerId   string
		pluginName   string
		credentialId string
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
				consumerId:   "1",
				pluginName:   "cors",
				credentialId: "1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &MseAdapterImpl{
				ProviderName: tt.fields.ProviderName,
			}
			if err := impl.DeleteCredential(tt.args.consumerId, tt.args.pluginName, tt.args.credentialId); (err != nil) != tt.wantErr {
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
		want    *KongCredentialListDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{
				consumerId: "1",
				pluginName: "cors",
			},
			want:    &KongCredentialListDto{},
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
		req *KongUpstreamDto
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *KongUpstreamDto
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args: args{req: &KongUpstreamDto{
				Id:           "1",
				Name:         "xx",
				Healthchecks: HealthchecksDto{},
			}},
			want: &KongUpstreamDto{
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
		want    *KongUpstreamStatusRespDto
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:   "Test_01",
			fields: fields{ProviderName: "MSE"},
			args:   args{upstreamId: "1"},
			want: &KongUpstreamStatusRespDto{
				Data: []KongTargetDto{},
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
		want    []KongRouteRespDto
		wantErr bool
	}{
		{
			name:    "Test_01",
			fields:  fields{ProviderName: "MSE"},
			want:    []KongRouteRespDto{},
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
		want    []KongRouteRespDto
		wantErr bool
	}{
		{
			name:    "Test_01",
			fields:  fields{},
			args:    args{tag: "xxx"},
			want:    []KongRouteRespDto{},
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
