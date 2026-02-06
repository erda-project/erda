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

package oauth

import "testing"

func TestReferValidate(t *testing.T) {
	p := provider{
		Cfg: &config{
			PlatformRootDomain: "erda.cloud",
		},
	}

	tests := []struct {
		name       string
		refer      string
		allowList  []string
		wantResult bool
	}{
		{
			name:       "Matching referer from uc",
			refer:      "https://uc.erda.cloud/login",
			wantResult: true,
		},
		{
			name:       "Matching referer with allowList",
			refer:      "https://sub.erdax.cloud/login",
			allowList:  []string{"sub.erdax.cloud"},
			wantResult: true,
		},
		{
			name:       "Matching platform domain",
			refer:      "https://erda.cloud/login",
			wantResult: true,
		},
		{
			name:       "Matching all sub domains",
			refer:      "https://sub.erda.cloud/login",
			allowList:  []string{"*.erda.cloud"},
			wantResult: true,
		},
		{
			name:       "Matching all sub domains list",
			refer:      "https://b.test.com/login",
			allowList:  []string{"*.erda.cloud", "*.test.com"},
			wantResult: true,
		},
		{
			name:       "Missmatch case",
			refer:      "https://btest.com/login",
			allowList:  []string{"*.erda.cloud", "*.test.com"},
			wantResult: false,
		},
		{
			name:       "Non-matching referer with allowList",
			refer:      "https://test.erdax.cloud/login",
			wantResult: false,
		},
		{
			name:       "Empty referer",
			refer:      "",
			wantResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p.Cfg.AllowedReferrers = tt.allowList
			p.referMatcher = p.buildReferMatcher()
			got := p.referMatcher.Match(tt.refer)
			if got != tt.wantResult {
				t.Errorf("referValidate() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}
