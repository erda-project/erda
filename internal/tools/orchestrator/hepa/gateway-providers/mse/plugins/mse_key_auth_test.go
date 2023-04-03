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

package plugins

import (
	"reflect"
	"sort"
	"testing"

	mseclient "github.com/alibabacloud-go/mse-20190531/v3/client"

	. "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	mseCommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/dto"
)

func Test_mergeKeyAuthConfig(t *testing.T) {
	type args struct {
		CurrentKeyAutoConfig dto.MsePluginKeyAuthConfig
		updateKeyAutoConfig  dto.MsePluginKeyAuthConfig
	}
	tests := []struct {
		name    string
		args    args
		want    dto.MsePluginKeyAuthConfig
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Test_01",
			args: args{
				CurrentKeyAutoConfig: dto.MsePluginKeyAuthConfig{
					Consumers: []dto.Consumers{
						{
							Name:       "aaa",
							Credential: "5d1a401b3aef4ee5bc39a99654672f91",
						},
						{
							Name:       "bbb",
							Credential: "5d1a401b3aef4ee5bc39a99654672f92",
						},
					},
					Keys:     []string{"apikey", "x-api-key"},
					InQuery:  true,
					InHeader: true,
					Rules: []dto.Rules{
						{
							MatchRoute: []string{"project-5846-test-dice-test-5846-api-754bbce0af034774ac2b8f74c7e070a6-0149aa-058a61f1-ccaae83a5fe7846d18d4ac80940a8fdc3"},
							Allow:      []string{"aaa"},
						},
						{
							MatchRoute: []string{"project-5846-test-dice-test-5846-api-66f56a64312143cfbbbd04eca82eca56-bff529-51ec627a-ccaae83a5fe7846d18d4ac80940a8fdc3"},
							Allow:      []string{"bbb"},
						},
					},
				},
				updateKeyAutoConfig: dto.MsePluginKeyAuthConfig{
					Consumers: []dto.Consumers{
						{
							Name:       "633.5846.TEST.erda-jicheng:abc",
							Credential: "8d22576204d74e869179dd7d19503570",
						},
					},
					Keys:     []string{"apikey", "x-api-key"},
					InQuery:  true,
					InHeader: true,
					Rules: []dto.Rules{
						{
							MatchRoute: []string{"project-5846-test-dice-test-5846-api-3170552e65d444e984a82e33ec44727f-93ac19-96b79390-ccaae83a5fe7846d18d4ac80940a8fdc3"},
							Allow:      []string{"633.5846.TEST.erda-jicheng:abc"},
						},
					},
				},
			},
			want: dto.MsePluginKeyAuthConfig{
				Consumers: []dto.Consumers{
					{
						Name:       "aaa",
						Credential: "5d1a401b3aef4ee5bc39a99654672f91",
					},
					{
						Name:       "bbb",
						Credential: "5d1a401b3aef4ee5bc39a99654672f92",
					},
					{
						Name:       "633.5846.TEST.erda-jicheng:abc",
						Credential: "8d22576204d74e869179dd7d19503570",
					},
					{
						Name:       DEFAULT_MSE_CONSUMER_NAME,
						Credential: DEFAULT_MSE_CONSUMER_CREDENTIAL,
					},
				},
				Keys:     []string{"appkey", "x-app-key"},
				InQuery:  true,
				InHeader: true,
				Rules: []dto.Rules{
					{
						MatchRoute: []string{"project-5846-test-dice-test-5846-api-754bbce0af034774ac2b8f74c7e070a6-0149aa-058a61f1-ccaae83a5fe7846d18d4ac80940a8fdc3"},
						Allow:      []string{"aaa"},
					},
					{
						MatchRoute: []string{"project-5846-test-dice-test-5846-api-66f56a64312143cfbbbd04eca82eca56-bff529-51ec627a-ccaae83a5fe7846d18d4ac80940a8fdc3"},
						Allow:      []string{"bbb"},
					},
					{
						MatchRoute: []string{"project-5846-test-dice-test-5846-api-3170552e65d444e984a82e33ec44727f-93ac19-96b79390-ccaae83a5fe7846d18d4ac80940a8fdc3"},
						Allow:      []string{"633.5846.TEST.erda-jicheng:abc"},
					},
					{
						MatchRoute: []string{DEFAULT_MSE_ROUTE_NAME},
						Allow:      []string{DEFAULT_MSE_CONSUMER_NAME},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Test_02",
			args: args{
				CurrentKeyAutoConfig: dto.MsePluginKeyAuthConfig{
					Consumers: []dto.Consumers{
						{
							Name:       "aaa",
							Credential: "5d1a401b3aef4ee5bc39a99654672f91",
						},
						{
							Name:       "bbb",
							Credential: "5d1a401b3aef4ee5bc39a99654672f92",
						},
					},
					Keys:     []string{"apikey", "x-api-key"},
					InQuery:  true,
					InHeader: true,
					Rules: []dto.Rules{
						{
							MatchRoute: []string{"project-5846-test-dice-test-5846-api-754bbce0af034774ac2b8f74c7e070a6-0149aa-058a61f1-ccaae83a5fe7846d18d4ac80940a8fdc3"},
							Allow:      []string{"aaa", "bbb"},
						},
						{
							MatchRoute: []string{"project-5846-test-dice-test-5846-api-66f56a64312143cfbbbd04eca82eca56-bff529-51ec627a-ccaae83a5fe7846d18d4ac80940a8fdc3"},
							Allow:      []string{"bbb"},
						},
					},
				},
				updateKeyAutoConfig: dto.MsePluginKeyAuthConfig{
					Consumers: []dto.Consumers{
						{
							Name:       "aaa",
							Credential: "5d1a401b3aef4ee5bc39a99654672f91",
						},
					},
					Keys:     []string{"apikey", "x-api-key"},
					InQuery:  true,
					InHeader: true,
					Rules: []dto.Rules{
						{
							MatchRoute: []string{"project-5846-test-dice-test-5846-api-754bbce0af034774ac2b8f74c7e070a6-0149aa-058a61f1-ccaae83a5fe7846d18d4ac80940a8fdc3"},
							Allow:      []string{"aaa"},
						},
					},
				},
			},
			want: dto.MsePluginKeyAuthConfig{
				Consumers: []dto.Consumers{
					{
						Name:       "aaa",
						Credential: "5d1a401b3aef4ee5bc39a99654672f91",
					},
					{
						Name:       "bbb",
						Credential: "5d1a401b3aef4ee5bc39a99654672f92",
					},
					{
						Name:       DEFAULT_MSE_CONSUMER_NAME,
						Credential: DEFAULT_MSE_CONSUMER_CREDENTIAL,
					},
				},
				Keys:     []string{"appkey", "x-app-key"},
				InQuery:  true,
				InHeader: true,
				Rules: []dto.Rules{
					{
						MatchRoute: []string{"project-5846-test-dice-test-5846-api-754bbce0af034774ac2b8f74c7e070a6-0149aa-058a61f1-ccaae83a5fe7846d18d4ac80940a8fdc3"},
						Allow:      []string{"aaa"},
					},
					{
						MatchRoute: []string{"project-5846-test-dice-test-5846-api-66f56a64312143cfbbbd04eca82eca56-bff529-51ec627a-ccaae83a5fe7846d18d4ac80940a8fdc3"},
						Allow:      []string{"bbb"},
					},
					{
						MatchRoute: []string{DEFAULT_MSE_ROUTE_NAME},
						Allow:      []string{DEFAULT_MSE_CONSUMER_NAME},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mergeKeyAuthConfig(tt.args.CurrentKeyAutoConfig, tt.args.updateKeyAutoConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("mergeKeyAuthConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Sort(dto.SortConsumers{got.Consumers, func(p, q *dto.Consumers) bool {
				return p.Credential < q.Credential // Credential 递增排序
			}})
			sort.Sort(dto.SortConsumers{tt.want.Consumers, func(p, q *dto.Consumers) bool {
				return p.Credential < q.Credential // Credential 递增排序
			}})

			sort.Sort(dto.SortRules{got.Rules, func(p, q *dto.Rules) bool {
				return p.MatchRoute[0] < q.MatchRoute[0] // Credential 递增排序
			}})
			sort.Sort(dto.SortRules{tt.want.Rules, func(p, q *dto.Rules) bool {
				return p.MatchRoute[0] < q.MatchRoute[0] // Credential 递增排序
			}})

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeKeyAuthConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateKeyAuthConfig(t *testing.T) {
	type args struct {
		req      *PluginReqDto
		confList map[string][]mseclient.GetPluginConfigResponseBodyDataGatewayConfigList
	}

	var configLevel int32 = 0
	var id int64 = 1
	var gatewayId int64 = 2686
	enable := true
	config := "consumers: \n# 注意！该凭证仅做示例使用，请勿用于具体业务，造成安全风险\n- credential: 5d1a401b3aef4ee5bc39a99654672f92\n  name: bbb\n- credential: 5d1a401b3aef4ee5bc39a99654672f91\n  name: aaa\nkeys:\n- apikey\n- x-api-key\n_rules_:\n# 规则一：按路由名称匹配生效\n- _match_route_:\n  - project-5846-test-dice-test-5846-api-754bbce0af034774ac2b8f74c7e070a6-0149aa-058a61f1-ccaae83a5fe7846d18d4ac80940a8fdc3\n  allow:\n  - aaa\n- _match_route_:\n  - project-5846-test-dice-test-5846-api-66f56a64312143cfbbbd04eca82eca56-bff529-51ec627a-ccaae83a5fe7846d18d4ac80940a8fdc3\n  allow:\n  - bbb\n  \n"
	cL := make(map[string][]mseclient.GetPluginConfigResponseBodyDataGatewayConfigList)
	ls := make([]mseclient.GetPluginConfigResponseBodyDataGatewayConfigList, 0)
	ls = append(ls, mseclient.GetPluginConfigResponseBodyDataGatewayConfigList{
		Config:          &config,
		ConfigLevel:     &configLevel,
		Enable:          &enable,
		GatewayId:       &gatewayId,
		GatewayUniqueId: nil,
		GmtCreate:       nil,
		GmtModified:     nil,
		Id:              &id,
		PluginId:        nil,
	})
	cL[MsePluginConfigLevelGlobal] = ls

	reqConfig := make(map[string]interface{})
	consumers := make([]dto.Consumers, 0)
	consumers = append(consumers, dto.Consumers{
		Name:       "abc",
		Credential: "xxxxyyyy",
	})
	reqConfig["whitelist"] = consumers
	tests := []struct {
		name    string
		args    args
		want    string
		want1   int64
		wantErr bool
	}{
		{
			name: "Test_01",
			args: args{
				req: &PluginReqDto{
					Name:         mseCommon.MsePluginKeyAuth,
					Config:       reqConfig,
					MSERouteName: "xyz",
				},
				confList: cL,
			},
			want:    "",
			want1:   1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got1, err := CreateKeyAuthConfig(tt.args.req, tt.args.confList)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateKeyAuthConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got1 != tt.want1 {
				t.Errorf("CreateKeyAuthConfig() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
