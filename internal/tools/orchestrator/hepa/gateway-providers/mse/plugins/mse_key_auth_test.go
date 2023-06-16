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

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/dto"
)

func Test_mergeKeyAuthConfig(t *testing.T) {
	type args struct {
		CurrentKeyAutoConfig dto.MsePluginConfig
		updateKeyAutoConfig  dto.MsePluginConfig
	}
	tests := []struct {
		name    string
		args    args
		want    dto.MsePluginConfig
		wantErr bool
	}{
		{
			name: "Test_01",
			args: args{
				CurrentKeyAutoConfig: dto.MsePluginConfig{
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
				updateKeyAutoConfig: dto.MsePluginConfig{
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
			want: dto.MsePluginConfig{
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
						Name:       MseDefaultConsumerName,
						Credential: MseDefaultConsumerCredential,
					},
				},
				Keys:     []string{"appKey", "x-app-key"},
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
						MatchRoute: []string{MseDefaultRouteName},
						Allow:      []string{MseDefaultConsumerName},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Test_02",
			args: args{
				CurrentKeyAutoConfig: dto.MsePluginConfig{
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
				updateKeyAutoConfig: dto.MsePluginConfig{
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
			want: dto.MsePluginConfig{
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
						Name:       MseDefaultConsumerName,
						Credential: MseDefaultConsumerCredential,
					},
				},
				Keys:     []string{"appKey", "x-app-key"},
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
						MatchRoute: []string{MseDefaultRouteName},
						Allow:      []string{MseDefaultConsumerName},
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
