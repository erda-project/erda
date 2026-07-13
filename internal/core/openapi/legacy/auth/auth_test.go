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

package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/erda-project/erda/internal/core/openapi/legacy/api/spec"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/internal/core/user/common"
)

func TestDualAuthSelection(t *testing.T) {
	authSpec := &spec.Spec{CheckLogin: true, CheckToken: true}
	auth := Auth{UserAuth: cookieUserAuthFacade{}}

	tests := []struct {
		name          string
		spec          *spec.Spec
		configure     func(*http.Request)
		want          checkType
		wantErrDetail string
	}{
		{
			name: "session cookie",
			configure: func(req *http.Request) {
				req.AddCookie(&http.Cookie{Name: "erda_iam", Value: "test-session"})
			},
			want: LOGIN,
		},
		{
			name: "authorization token",
			configure: func(req *http.Request) {
				req.Header.Set(HeaderAuthorization, "Bearer test-token")
			},
			want: TOKEN,
		},
		{
			name: "basic authorization",
			spec: &spec.Spec{CheckBasicAuth: true},
			configure: func(req *http.Request) {
				req.Header.Set(HeaderAuthorization, "Basic dXNlcjpwYXNz")
			},
			want: BASICAUTH,
		},
		{
			name: "optional login",
			spec: &spec.Spec{TryCheckLogin: true},
			want: TRY_LOGIN,
		},
		{
			name: "no auth required",
			spec: &spec.Spec{},
			want: NONE,
		},
		{
			name:          "unauthenticated",
			want:          NONE,
			wantErrDetail: "lack of required auth header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/deployments/1/actions/cancel", nil)
			if tt.configure != nil {
				tt.configure(req)
			}

			testSpec := tt.spec
			if testSpec == nil {
				testSpec = authSpec
			}
			got, err := auth.whichCheck(req, testSpec)
			if got != tt.want {
				t.Fatalf("whichCheck() = %v, want %v", got, tt.want)
			}
			if tt.wantErrDetail == "" {
				if err != nil {
					t.Fatalf("whichCheck() unexpected error: %v", err)
				}
				return
			}
			if err == nil || err.Error() != tt.wantErrDetail {
				t.Fatalf("whichCheck() error = %v, want %q", err, tt.wantErrDetail)
			}
		})
	}
}

type cookieUserAuthFacade struct{}

func (cookieUserAuthFacade) NewState() domain.UserAuthState {
	return cookieUserAuthState{}
}

type cookieUserAuthState struct{}

func (cookieUserAuthState) GetOrgInfo(string, string) (uint64, error) {
	return 0, nil
}

func (cookieUserAuthState) IsLogin(req *http.Request) domain.UserAuthResult {
	if _, err := req.Cookie("erda_iam"); err == nil {
		return domain.UserAuthResult{Code: domain.AuthSuccess}
	}
	return domain.UserAuthResult{Code: domain.Unauthed}
}

func (cookieUserAuthState) GetInfo(*http.Request) (*common.UserInfo, domain.UserAuthResult) {
	return nil, domain.UserAuthResult{Code: domain.AuthSuccess}
}

func (cookieUserAuthState) GetScopeInfo(*http.Request) (common.UserScopeInfo, domain.UserAuthResult) {
	return common.UserScopeInfo{}, domain.UserAuthResult{Code: domain.AuthSuccess}
}

func (cookieUserAuthState) Login(string, url.Values) error {
	return nil
}

func (cookieUserAuthState) PwdLogin(string, string) error {
	return nil
}
