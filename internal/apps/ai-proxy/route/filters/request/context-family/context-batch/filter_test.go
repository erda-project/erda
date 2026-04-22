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

package context

import "testing"

func TestParseCreateBatchRequest(t *testing.T) {
	testCases := []struct {
		name    string
		raw     map[string]any
		wantErr bool
	}{
		{
			name: "valid request",
			raw: map[string]any{
				"input_file_id":     "file-abc",
				"endpoint":          "/v1/chat/completions",
				"completion_window": "24h",
			},
			wantErr: false,
		},
		{
			name: "missing input_file_id",
			raw: map[string]any{
				"endpoint":          "/v1/chat/completions",
				"completion_window": "24h",
			},
			wantErr: true,
		},
		{
			name: "missing endpoint",
			raw: map[string]any{
				"input_file_id":     "file-abc",
				"completion_window": "24h",
			},
			wantErr: true,
		},
		{
			name: "missing completion_window",
			raw: map[string]any{
				"input_file_id": "file-abc",
				"endpoint":      "/v1/chat/completions",
			},
			wantErr: true,
		},
		{
			name: "endpoint must start with slash",
			raw: map[string]any{
				"input_file_id":     "file-abc",
				"endpoint":          "v1/chat/completions",
				"completion_window": "24h",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseCreateBatchRequest(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if got == nil {
				t.Fatalf("expected non-nil result")
			}
		})
	}
}
