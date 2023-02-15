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

package apigateway

import (
	"reflect"
	"testing"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers"
)

func Test_provider_CheckIfHasCustomConfig(t *testing.T) {
	type fields struct {
		DefaultDeployHandler *handlers.DefaultDeployHandler
		Cfg                  *config
		Log                  logs.Logger
		DB                   *gorm.DB
		PipelineSvc          pb.PipelineServiceServer
	}
	type args struct {
		clusterConfig map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]string
		want1  bool
	}{
		{
			name: "Test_01",
			fields: fields{
				DefaultDeployHandler: nil,
				Cfg: &config{
					MainClusterInfo: struct {
						Name       string `file:"name"`
						RootDomain string `file:"root_domain"`
						Protocol   string `file:"protocol"`
						HttpPort   string `file:"http_port"`
						HttpsPort  string `file:"https_port"`
					}(struct {
						Name       string
						RootDomain string
						Protocol   string
						HttpPort   string
						HttpsPort  string
					}{Name: "xxx", RootDomain: "mse-daily.terminus.io", Protocol: "https", HttpPort: "80", HttpsPort: "443"})},
				Log:         nil,
				DB:          nil,
				PipelineSvc: nil,
			},
			args: args{
				clusterConfig: map[string]string{
					handlers.GatewayProviderVendorKey: "MSE",
					handlers.GatewayEndpoint:          "mse-daily.terminus.io",
				},
			},
			want: map[string]string{
				"HEPA_GATEWAY_HOST": "https://hepa.mse-daily.terminus.io",
				"HEPA_GATEWAY_PORT": "443",
				"GATEWAY_ENDPOINT":  "mse-daily.terminus.io",
			},
			want1: true,
		},
		{
			name: "Test_02",
			fields: fields{
				DefaultDeployHandler: nil,
				Cfg: &config{
					MainClusterInfo: struct {
						Name       string `file:"name"`
						RootDomain string `file:"root_domain"`
						Protocol   string `file:"protocol"`
						HttpPort   string `file:"http_port"`
						HttpsPort  string `file:"https_port"`
					}(struct {
						Name       string
						RootDomain string
						Protocol   string
						HttpPort   string
						HttpsPort  string
					}{Name: "xxx", RootDomain: "mse-daily.terminus.io", Protocol: "http", HttpPort: "80", HttpsPort: "443"})},
				Log:         nil,
				DB:          nil,
				PipelineSvc: nil,
			},
			args: args{
				clusterConfig: map[string]string{
					handlers.GatewayProviderVendorKey: "MSE",
					handlers.GatewayEndpoint:          "mse-daily.terminus.io",
				},
			},
			want: map[string]string{
				"HEPA_GATEWAY_HOST": "http://hepa.mse-daily.terminus.io",
				"HEPA_GATEWAY_PORT": "80",
				"GATEWAY_ENDPOINT":  "mse-daily.terminus.io",
			},
			want1: true,
		},
		{
			name: "Test_03",
			fields: fields{
				DefaultDeployHandler: nil,
				Cfg: &config{
					MainClusterInfo: struct {
						Name       string `file:"name"`
						RootDomain string `file:"root_domain"`
						Protocol   string `file:"protocol"`
						HttpPort   string `file:"http_port"`
						HttpsPort  string `file:"https_port"`
					}(struct {
						Name       string
						RootDomain string
						Protocol   string
						HttpPort   string
						HttpsPort  string
					}{Name: "xxx", RootDomain: "mse-daily.terminus.io", Protocol: "http", HttpPort: "80", HttpsPort: "443"})},
				Log:         nil,
				DB:          nil,
				PipelineSvc: nil,
			},
			args: args{
				clusterConfig: map[string]string{
					handlers.GatewayProviderVendorKey: "MSE",
					"DICE_ROOT_DOMAIN":                "mse-daily.terminus.io",
				},
			},
			want: map[string]string{
				"HEPA_GATEWAY_HOST": "http://hepa.mse-daily.terminus.io",
				"HEPA_GATEWAY_PORT": "80",
				"GATEWAY_ENDPOINT":  "mse-daily.terminus.io",
			},
			want1: true,
		},
		{
			name: "Test_04",
			fields: fields{
				DefaultDeployHandler: nil,
				Cfg: &config{
					MainClusterInfo: struct {
						Name       string `file:"name"`
						RootDomain string `file:"root_domain"`
						Protocol   string `file:"protocol"`
						HttpPort   string `file:"http_port"`
						HttpsPort  string `file:"https_port"`
					}(struct {
						Name       string
						RootDomain string
						Protocol   string
						HttpPort   string
						HttpsPort  string
					}{Name: "xxx", RootDomain: "mse-daily.terminus.io", Protocol: "http", HttpPort: "80", HttpsPort: "443"})},
				Log:         nil,
				DB:          nil,
				PipelineSvc: nil,
			},
			args: args{
				clusterConfig: map[string]string{
					"DICE_ROOT_DOMAIN": "mse-daily.terminus.io",
				},
			},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				DefaultDeployHandler: tt.fields.DefaultDeployHandler,
				Cfg:                  tt.fields.Cfg,
				Log:                  tt.fields.Log,
				DB:                   tt.fields.DB,
				PipelineSvc:          tt.fields.PipelineSvc,
			}
			got, got1 := p.CheckIfHasCustomConfig(tt.args.clusterConfig)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckIfHasCustomConfig() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("CheckIfHasCustomConfig() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
