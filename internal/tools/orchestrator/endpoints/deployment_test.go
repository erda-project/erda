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

package endpoints

import (
	"net/http/httptest"
	"testing"

	"github.com/erda-project/erda/pkg/http/httputil"
)

func TestDeploymentCancelOperator(t *testing.T) {
	tests := []struct {
		name         string
		headerUserID string
		bodyOperator string
		want         string
	}{
		{
			name:         "authenticated user overrides body operator",
			headerUserID: "verified-user",
			bodyOperator: "spoofed-user",
			want:         "verified-user",
		},
		{
			name:         "token caller keeps body operator",
			bodyOperator: "token-operator",
			want:         "token-operator",
		},
		{
			name: "missing operator remains empty",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/deployments/1/actions/cancel", nil)
			if tt.headerUserID != "" {
				req.Header.Set(httputil.UserHeader, tt.headerUserID)
			}

			if got := deploymentCancelOperator(req, tt.bodyOperator); got != tt.want {
				t.Fatalf("deploymentCancelOperator() = %q, want %q", got, tt.want)
			}
		})
	}
}
