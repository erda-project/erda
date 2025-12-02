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

package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name    string
		payload map[string]interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "accessToken with bearer prefix",
			payload: map[string]interface{}{"accessToken": "Bearer abc"},
			want:    "abc",
		},
		{
			name:    "accessToken plain",
			payload: map[string]interface{}{"accessToken": "abc"},
			want:    "abc",
		},
		{
			name:    "data fallback bearer",
			payload: map[string]interface{}{"data": "Bearer def"},
			want:    "def",
		},
		{
			name:    "missing token",
			payload: map[string]interface{}{},
			wantErr: true,
		},
		{
			name:    "blank token",
			payload: map[string]interface{}{"accessToken": "  "},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &NacosClient{}
			got, err := c.extractToken(tt.payload)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("token mismatch: got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEnsureTokenUsesCache(t *testing.T) {
	c := &NacosClient{bearerToken: "cached"}
	got, err := c.ensureToken()
	if err != nil {
		t.Fatalf("ensureToken returned error: %v", err)
	}
	if got != "cached" {
		t.Fatalf("ensureToken returned %q, want %q", got, "cached")
	}
}

func TestLoginTokenVariants(t *testing.T) {
	tests := []struct {
		name              string
		loginPayload      map[string]interface{}
		expectedToken     string
		expectLoginErr    bool
		expectNamespaceID string
	}{
		{
			name:              "accessToken plain",
			loginPayload:      map[string]interface{}{"accessToken": "abc123"},
			expectedToken:     "abc123",
			expectNamespaceID: "ns-1",
		},
		{
			name:              "accessToken bearer prefix",
			loginPayload:      map[string]interface{}{"accessToken": "Bearer def456"},
			expectedToken:     "def456",
			expectNamespaceID: "ns-1",
		},
		{
			name:              "data fallback bearer",
			loginPayload:      map[string]interface{}{"data": "Bearer xyz789"},
			expectedToken:     "xyz789",
			expectNamespaceID: "ns-1",
		},
		{
			name:           "missing token error",
			loginPayload:   map[string]interface{}{},
			expectLoginErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotAuthHeader string
			var gotQuery url.Values
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case loginPath:
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(tt.loginPayload)
				case namespacesPath:
					gotAuthHeader = r.Header.Get("Authorization")
					gotQuery = r.URL.Query()
					resp := map[string]interface{}{
						"data": []map[string]string{
							{"namespaceShowName": "test-ns", "namespace": tt.expectNamespaceID},
						},
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(resp)
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()

			client := NewNacosClient("", server.URL, "user", "pass")

			token, err := client.Login()
			if tt.expectLoginErr {
				if err == nil {
					t.Fatalf("expected login error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("login failed: %v", err)
			}
			if token != tt.expectedToken {
				t.Fatalf("unexpected token: got %q, want %q", token, tt.expectedToken)
			}

			_, err = client.GetNamespaceId("test-ns")
			if tt.expectNamespaceID != "" && err != nil {
				t.Fatalf("GetNamespaceId failed: %v", err)
			}
			if tt.expectNamespaceID != "" && gotAuthHeader != "Bearer "+tt.expectedToken {
				t.Fatalf("Authorization header mismatch: got %q, want %q", gotAuthHeader, "Bearer "+tt.expectedToken)
			}
			// Params should not be set for GetNamespaceId.
			if gotQuery != nil && len(gotQuery) != 0 {
				t.Fatalf("unexpected query params on namespace GET: %v", gotQuery.Encode())
			}
		})
	}
}
