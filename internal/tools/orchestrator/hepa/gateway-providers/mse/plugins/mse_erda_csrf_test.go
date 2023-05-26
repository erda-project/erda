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

func Test_getErdaCSRFSourceConfig(t *testing.T) {
	type args struct {
		config map[string]interface{}
	}

	cf := make(map[string]interface{})
	cf[common.MseErdaCSRFConfigUserCookie] = []string{common.MseErdaCSRFDefaultUserCookie}
	cf[common.MseErdaCSRFConfigExcludedMethod] = []string{
		http.MethodGet,
	}
	cf[common.MseErdaCSRFConfigTokenCookie] = common.MseErdaCSRFDefaultTokenName

	cf[common.MseErdaCSRFConfigTokenDomain] = common.MseErdaCSRFDefaultTokenDomain
	cf[common.MseErdaCSRFConfigCookieSecure] = common.MseErdaCSRFDefaultCookieSecure
	cf[common.MseErdaCSRFConfigValidTTL] = common.MseErdaCSRFDefaultValidTTL

	cf[common.MseErdaCSRFConfigErrStatus] = common.MseErdaCSRFDefaultErrStatus
	cf[common.MseErdaCSRFConfigRefreshTTL] = common.MseErdaCSRFDefaultRefreshTTL
	cf[common.MseErdaCSRFConfigErrMsg] = common.MseErdaCSRFDefaultErrMsg
	cf[common.MseErdaCSRFConfigSecret] = common.MseErdaCSRFDefaultJWTSecret

	cf[common.MseErdaCSRFRouteSwitch] = false

	tests := []struct {
		name               string
		args               args
		wantExcludedMethod []string
		wantUserCookie     string
		wantTokenName      string
		wantTokenDomain    string
		wantErrMsg         string
		wantJSecret        string
		wantValidTTL       int64
		wantRefreshTTL     int64
		wantErrStatus      int64
		wantCookieSecure   bool
		wantDisable        bool
		wantErr            bool
	}{
		{
			name: "Test_01",
			args: args{
				config: cf,
			},
			wantExcludedMethod: []string{
				http.MethodGet,
			},
			wantUserCookie:   common.MseErdaCSRFDefaultUserCookie,
			wantTokenName:    common.MseErdaCSRFDefaultTokenName,
			wantTokenDomain:  common.MseErdaCSRFDefaultTokenDomain,
			wantErrMsg:       common.MseErdaCSRFDefaultErrMsg,
			wantJSecret:      common.MseErdaCSRFDefaultJWTSecret,
			wantValidTTL:     common.MseErdaCSRFDefaultValidTTL,
			wantRefreshTTL:   common.MseErdaCSRFDefaultRefreshTTL,
			wantErrStatus:    common.MseErdaCSRFDefaultErrStatus,
			wantCookieSecure: common.MseErdaCSRFDefaultCookieSecure,
			wantDisable:      true,
			wantErr:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExcludedMethod, gotUserCookie, gotTokenName, gotTokenDomain, gotErrMsg, gotJSecret, gotValidTTL, gotRefreshTTL, gotErrStatus, gotCookieSecure, gotDisable, err := getErdaCSRFSourceConfig(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("getErdaCSRFSourceConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotExcludedMethod, tt.wantExcludedMethod) {
				t.Errorf("getErdaCSRFSourceConfig() gotExcludedMethod = %v, want %v", gotExcludedMethod, tt.wantExcludedMethod)
			}
			if gotUserCookie != tt.wantUserCookie {
				t.Errorf("getErdaCSRFSourceConfig() gotUserCookie = %v, want %v", gotUserCookie, tt.wantUserCookie)
			}
			if gotTokenName != tt.wantTokenName {
				t.Errorf("getErdaCSRFSourceConfig() gotTokenName = %v, want %v", gotTokenName, tt.wantTokenName)
			}
			if gotTokenDomain != tt.wantTokenDomain {
				t.Errorf("getErdaCSRFSourceConfig() gotTokenDomain = %v, want %v", gotTokenDomain, tt.wantTokenDomain)
			}
			if gotErrMsg != tt.wantErrMsg {
				t.Errorf("getErdaCSRFSourceConfig() gotErrMsg = %v, want %v", gotErrMsg, tt.wantErrMsg)
			}
			if gotJSecret != tt.wantJSecret {
				t.Errorf("getErdaCSRFSourceConfig() gotJSecret = %v, want %v", gotJSecret, tt.wantJSecret)
			}
			if gotValidTTL != tt.wantValidTTL {
				t.Errorf("getErdaCSRFSourceConfig() gotValidTTL = %v, want %v", gotValidTTL, tt.wantValidTTL)
			}
			if gotRefreshTTL != tt.wantRefreshTTL {
				t.Errorf("getErdaCSRFSourceConfig() gotRefreshTTL = %v, want %v", gotRefreshTTL, tt.wantRefreshTTL)
			}
			if gotErrStatus != tt.wantErrStatus {
				t.Errorf("getErdaCSRFSourceConfig() gotErrStatus = %v, want %v", gotErrStatus, tt.wantErrStatus)
			}
			if gotCookieSecure != tt.wantCookieSecure {
				t.Errorf("getErdaCSRFSourceConfig() gotCookieSecure = %v, want %v", gotCookieSecure, tt.wantCookieSecure)
			}
			if gotDisable != tt.wantDisable {
				t.Errorf("getErdaCSRFSourceConfig() gotDisable = %v, want %v", gotDisable, tt.wantDisable)
			}
		})
	}
}

func Test_mergeErdaCSRFConfig(t *testing.T) {
	type args struct {
		currentErdaCSRFConfig dto.MsePluginConfig
		updateErdaCSRFConfig  dto.MsePluginConfig
		updateForDisable      bool
	}

	cRules := make([]dto.Rules, 0)
	uRules := make([]dto.Rules, 0)

	cRules = append(cRules, dto.Rules{
		MatchRoute:     []string{MseDefaultRouteName},
		UserCookie:     common.MseErdaCSRFDefaultUserCookie,
		ExcludedMethod: []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace},
		TokenName:      common.MseErdaCSRFDefaultTokenName,
		TokenDomain:    common.MseErdaCSRFDefaultTokenDomain,
		CookieSecure:   common.MseErdaCSRFDefaultCookieSecure,
		ValidTTL:       common.MseErdaCSRFDefaultValidTTL,
		RefreshTTL:     common.MseErdaCSRFDefaultRefreshTTL,
		ErrStatus:      common.MseErdaCSRFDefaultErrStatus,
		ErrMsg:         common.MseErdaCSRFDefaultErrMsg,
		JWTSecret:      common.MseErdaCSRFDefaultJWTSecret,
	})
	cRules = append(cRules, dto.Rules{
		MatchRoute:     []string{"test-route"},
		UserCookie:     common.MseErdaCSRFDefaultUserCookie,
		ExcludedMethod: []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace},
		TokenName:      common.MseErdaCSRFDefaultTokenName,
		TokenDomain:    common.MseErdaCSRFDefaultTokenDomain,
		CookieSecure:   common.MseErdaCSRFDefaultCookieSecure,
		ValidTTL:       common.MseErdaCSRFDefaultValidTTL,
		RefreshTTL:     common.MseErdaCSRFDefaultRefreshTTL,
		ErrStatus:      common.MseErdaCSRFDefaultErrStatus,
		ErrMsg:         common.MseErdaCSRFDefaultErrMsg,
		JWTSecret:      common.MseErdaCSRFDefaultJWTSecret,
	})

	uRules = append(uRules, dto.Rules{
		MatchRoute:     []string{"test-route"},
		UserCookie:     common.MseErdaCSRFDefaultUserCookie,
		ExcludedMethod: []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace},
		TokenName:      common.MseErdaCSRFDefaultTokenName,
		TokenDomain:    common.MseErdaCSRFDefaultTokenDomain,
		CookieSecure:   common.MseErdaCSRFDefaultCookieSecure,
		ValidTTL:       common.MseErdaCSRFDefaultValidTTL,
		RefreshTTL:     common.MseErdaCSRFDefaultRefreshTTL,
		ErrStatus:      404,
		ErrMsg:         common.MseErdaCSRFDefaultErrMsg,
		JWTSecret:      common.MseErdaCSRFDefaultJWTSecret,
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
				currentErdaCSRFConfig: dto.MsePluginConfig{
					Rules: cRules,
				},
				updateErdaCSRFConfig: dto.MsePluginConfig{
					Rules: uRules,
				},
				updateForDisable: true,
			},
			want: dto.MsePluginConfig{
				Rules: []dto.Rules{{
					MatchRoute:     []string{MseDefaultRouteName},
					UserCookie:     common.MseErdaCSRFDefaultUserCookie,
					ExcludedMethod: []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace},
					TokenName:      common.MseErdaCSRFDefaultTokenName,
					TokenDomain:    common.MseErdaCSRFDefaultTokenDomain,
					CookieSecure:   common.MseErdaCSRFDefaultCookieSecure,
					ValidTTL:       common.MseErdaCSRFDefaultValidTTL,
					RefreshTTL:     common.MseErdaCSRFDefaultRefreshTTL,
					ErrStatus:      common.MseErdaCSRFDefaultErrStatus,
					ErrMsg:         common.MseErdaCSRFDefaultErrMsg,
					JWTSecret:      common.MseErdaCSRFDefaultJWTSecret,
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mergeErdaCSRFConfig(tt.args.currentErdaCSRFConfig, tt.args.updateErdaCSRFConfig, tt.args.updateForDisable)
			if (err != nil) != tt.wantErr {
				t.Errorf("mergeErdaCSRFConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeErdaCSRFConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
