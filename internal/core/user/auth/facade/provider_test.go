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

package facade

import (
	"net/http"
	"testing"

	"github.com/erda-project/erda/internal/core/user/auth/domain"
)

func TestProvider_NewState(t *testing.T) {
	p := &provider{
		Cfg: &Config{CookieName: "sid"},
		// UserOAuthSvc, IdentitySvc, Bundle left nil for this test
	}
	state := p.NewState()
	if state == nil {
		t.Fatal("NewState() should not return nil")
	}
	// State implements UserAuthState; without cookie IsLogin returns Unauthed
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	result := state.IsLogin(req)
	if result.Code != domain.Unauthed {
		t.Errorf("IsLogin without cookie expected Unauthed, got code %d detail %q", result.Code, result.Detail)
	}
}
