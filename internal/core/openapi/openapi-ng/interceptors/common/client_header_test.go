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

package common

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInterceptorRedactsSensitiveHeadersAfterHandling(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/deployments/1/actions/cancel", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Cookie", "erda_iam=test-session")
	req.Header.Set("Proxy-Authorization", "Basic test-proxy-token")
	req.Header.Set("OPENAPI-CSRF-TOKEN", "test-csrf-token")
	req.Header.Set("Content-Type", "application/json")

	handler := (&provider{}).Interceptor(func(rw http.ResponseWriter, req *http.Request) {
		if got := req.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("handler Authorization = %q, want original value", got)
		}
		if got := req.Header.Get("Cookie"); got != "erda_iam=test-session" {
			t.Fatalf("handler Cookie = %q, want original value", got)
		}
		if got := req.Header.Get("Proxy-Authorization"); got != "Basic test-proxy-token" {
			t.Fatalf("handler Proxy-Authorization = %q, want original value", got)
		}
		if got := req.Header.Get("OPENAPI-CSRF-TOKEN"); got != "test-csrf-token" {
			t.Fatalf("handler OPENAPI-CSRF-TOKEN = %q, want original value", got)
		}
	})

	handler(httptest.NewRecorder(), req)

	for _, header := range []string{"Authorization", "Cookie", "Proxy-Authorization", "OPENAPI-CSRF-TOKEN"} {
		if got := req.Header.Get(header); got != redactedHeaderValue {
			t.Errorf("logged %s = %q, want %q", header, got, redactedHeaderValue)
		}
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want preserved value", got)
	}
}
