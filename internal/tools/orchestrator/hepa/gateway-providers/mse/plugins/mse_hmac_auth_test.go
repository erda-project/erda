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

func Test_mergeHmacAuthConfig(t *testing.T) {
	type args struct {
		CurrentHmacAuthConfig dto.MsePluginConfig
		updateHmacAuthConfig  dto.MsePluginConfig
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
				CurrentHmacAuthConfig: dto.MsePluginConfig{
					Consumers: []dto.Consumers{
						{
							Name:   "aaa",
							Key:    "5d1a401b3aef4ee5bc39a99654672f91",
							Secret: "902450a7405942f5b205aaa1aa828caa",
						},
						{
							Name:   "bbb",
							Key:    "5d1a401b3aef4ee5bc39a99654672f92",
							Secret: "902450a7405942f5b205aaa1aa828cbb",
						},
					},
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
				updateHmacAuthConfig: dto.MsePluginConfig{
					Consumers: []dto.Consumers{
						{
							Name:   "633.5846.TEST.erda-jicheng:abc",
							Key:    "8d22576204d74e869179dd7d19503570",
							Secret: "902450a7405942f5b205aaa1aa828c0b",
						},
					},
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
						Name:   "aaa",
						Key:    "5d1a401b3aef4ee5bc39a99654672f91",
						Secret: "902450a7405942f5b205aaa1aa828caa",
					},
					{
						Name:   "bbb",
						Key:    "5d1a401b3aef4ee5bc39a99654672f92",
						Secret: "902450a7405942f5b205aaa1aa828cbb",
					},
					{
						Name:   "633.5846.TEST.erda-jicheng:abc",
						Key:    "8d22576204d74e869179dd7d19503570",
						Secret: "902450a7405942f5b205aaa1aa828c0b",
					},
					{
						Name:   MseDefaultConsumerName,
						Key:    MseDefaultConsumerKey,
						Secret: MseDefaultConsumerSecret,
					},
				},
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
				CurrentHmacAuthConfig: dto.MsePluginConfig{
					Consumers: []dto.Consumers{
						{
							Name:   "aaa",
							Key:    "5d1a401b3aef4ee5bc39a99654672f91",
							Secret: "902450a7405942f5b205aaa1aa828caa",
						},
						{
							Name:   "bbb",
							Key:    "5d1a401b3aef4ee5bc39a99654672f92",
							Secret: "902450a7405942f5b205aaa1aa828cbb",
						},
					},
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
				updateHmacAuthConfig: dto.MsePluginConfig{
					Consumers: []dto.Consumers{
						{
							Name:   "aaa",
							Key:    "5d1a401b3aef4ee5bc39a99654672f91",
							Secret: "902450a7405942f5b205aaa1aa828caa",
						},
					},
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
						Name:   "aaa",
						Key:    "5d1a401b3aef4ee5bc39a99654672f91",
						Secret: "902450a7405942f5b205aaa1aa828caa",
					},
					{
						Name:   "bbb",
						Key:    "5d1a401b3aef4ee5bc39a99654672f92",
						Secret: "902450a7405942f5b205aaa1aa828cbb",
					},
					{
						Name:   MseDefaultConsumerName,
						Key:    MseDefaultConsumerKey,
						Secret: MseDefaultConsumerSecret,
					},
				},
				Rules: []dto.Rules{
					{
						MatchRoute: []string{MseDefaultRouteName},
						Allow:      []string{MseDefaultConsumerName},
					},
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mergeHmacAuthConfig(tt.args.CurrentHmacAuthConfig, tt.args.updateHmacAuthConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("mergeHmacAuthConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Sort(dto.SortConsumers{got.Consumers, func(p, q *dto.Consumers) bool {
				return p.Key < q.Key // Credential 递增排序
			}})
			sort.Sort(dto.SortConsumers{tt.want.Consumers, func(p, q *dto.Consumers) bool {
				return p.Key < q.Key // Credential 递增排序
			}})

			sort.Sort(dto.SortRules{got.Rules, func(p, q *dto.Rules) bool {
				return p.MatchRoute[0] < q.MatchRoute[0] // Credential 递增排序
			}})
			sort.Sort(dto.SortRules{tt.want.Rules, func(p, q *dto.Rules) bool {
				return p.MatchRoute[0] < q.MatchRoute[0] // Credential 递增排序
			}})

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeHmacAuthConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
