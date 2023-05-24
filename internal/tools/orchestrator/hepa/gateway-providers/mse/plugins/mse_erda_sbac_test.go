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
	"net/http"
	"reflect"
	"testing"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/dto"
)

func Test_getErdaSBACSourceConfig(t *testing.T) {
	type args struct {
		config map[string]interface{}
	}

	cf := make(map[string]interface{})
	cf[common.MseErdaSBACConfigAccessControlAPI] = common.MseErdaSBACAccessControlAPI
	cf[common.MseErdaSBACConfigMatchPatterns] = []string{"^/"}
	cf[common.MseErdaSBACConfigHttpMethods] = map[string]bool{http.MethodGet: true}
	cf[common.MseErdaSBACConfigWithHeaders] = []string{"x_a"}
	cf[common.MseErdaSBACRouteSwitch] = false

	tests := []struct {
		name                 string
		args                 args
		wantAccessControlAPI string
		wantMatchPatterns    []string
		wantHttpMethods      []string
		wantWithHeaders      []string
		wantWithCookie       bool
		wantDisable          bool
		wantErr              bool
	}{
		{
			name: "Test_01",
			args: args{
				config: cf,
			},
			wantAccessControlAPI: common.MseErdaSBACAccessControlAPI,
			wantMatchPatterns:    []string{"^/"},
			wantHttpMethods:      []string{http.MethodGet},
			wantWithHeaders:      []string{"x_a"},
			wantWithCookie:       false,
			wantDisable:          true,
			wantErr:              false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAccessControlAPI, gotMatchPatterns, gotHttpMethods, gotWithHeaders, gotWithCookie, gotDisable, err := getErdaSBACSourceConfig(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("getErdaSBACSourceConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotAccessControlAPI != tt.wantAccessControlAPI {
				t.Errorf("getErdaSBACSourceConfig() gotAccessControlAPI = %v, want %v", gotAccessControlAPI, tt.wantAccessControlAPI)
			}
			if !reflect.DeepEqual(gotMatchPatterns, tt.wantMatchPatterns) {
				t.Errorf("getErdaSBACSourceConfig() gotMatchPatterns = %v, want %v", gotMatchPatterns, tt.wantMatchPatterns)
			}
			if !reflect.DeepEqual(gotHttpMethods, tt.wantHttpMethods) {
				t.Errorf("getErdaSBACSourceConfig() gotHttpMethods = %v, want %v", gotHttpMethods, tt.wantHttpMethods)
			}
			if !reflect.DeepEqual(gotWithHeaders, tt.wantWithHeaders) {
				t.Errorf("getErdaSBACSourceConfig() gotWithHeaders = %v, want %v", gotWithHeaders, tt.wantWithHeaders)
			}
			if gotWithCookie != tt.wantWithCookie {
				t.Errorf("getErdaSBACSourceConfig() gotWithCookie = %v, want %v", gotWithCookie, tt.wantWithCookie)
			}
			if gotDisable != tt.wantDisable {
				t.Errorf("getErdaSBACSourceConfig() gotDisable = %v, want %v", gotDisable, tt.wantDisable)
			}
		})
	}
}

func Test_mergeErdaSBACConfig(t *testing.T) {
	type args struct {
		currentErdaSBACConfig dto.MsePluginConfig
		updateErdaSBACConfig  dto.MsePluginConfig
		updateForDisable      bool
	}

	cRules := make([]dto.Rules, 0)
	uRules := make([]dto.Rules, 0)

	cRules = append(cRules, dto.Rules{
		MatchRoute:       []string{MseDefaultRouteName},
		AccessControlAPI: common.MseErdaSBACAccessControlAPI,
		HttpMethods:      []string{http.MethodGet},
		MatchPatterns:    []string{"^/"},
		WithHeaders:      []string{"x_a"},
		WithCookie:       false,
	})
	cRules = append(cRules, dto.Rules{
		MatchRoute:       []string{"test-route"},
		AccessControlAPI: common.MseErdaSBACAccessControlAPI,
		HttpMethods:      []string{http.MethodGet},
		MatchPatterns:    []string{"^/"},
		WithHeaders:      []string{"x_a"},
		WithCookie:       false,
	})

	uRules = append(uRules, dto.Rules{
		MatchRoute:       []string{"test-route"},
		AccessControlAPI: common.MseErdaSBACAccessControlAPI,
		HttpMethods:      []string{http.MethodGet, http.MethodPost},
		MatchPatterns:    []string{"^/"},
		WithHeaders:      []string{"x_a"},
		WithCookie:       false,
	})

	tests := []struct {
		name    string
		args    args
		want    dto.MsePluginConfig
		wantErr bool
	}{
		{
			name: "Test_01",
			args: args{
				currentErdaSBACConfig: dto.MsePluginConfig{
					Rules: cRules,
				},
				updateErdaSBACConfig: dto.MsePluginConfig{
					Rules: uRules,
				},
				updateForDisable: true,
			},
			want: dto.MsePluginConfig{
				Rules: []dto.Rules{{
					MatchRoute:       []string{MseDefaultRouteName},
					AccessControlAPI: common.MseErdaSBACAccessControlAPI,
					HttpMethods:      []string{http.MethodGet},
					MatchPatterns:    []string{"^/"},
					WithHeaders:      []string{"x_a"},
					WithCookie:       false,
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mergeErdaSBACConfig(tt.args.currentErdaSBACConfig, tt.args.updateErdaSBACConfig, tt.args.updateForDisable)
			if (err != nil) != tt.wantErr {
				t.Errorf("mergeErdaSBACConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeErdaSBACConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
