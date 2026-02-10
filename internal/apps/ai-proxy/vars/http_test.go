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

package vars

import "testing"

func TestTrimBearer(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "BearerPrefix",
			input: "Bearer token",
			want:  "token",
		},
		{
			name:  "bearerPrefix",
			input: "bearer token",
			want:  "token",
		},
		{
			name:  "BearerOnly",
			input: "Bearer ",
			want:  "",
		},
		{
			name:  "bearerOnly",
			input: "bearer ",
			want:  "",
		},
		{
			name:  "NoPrefix",
			input: "token",
			want:  "token",
		},
		{
			name:  "Empty",
			input: "",
			want:  "",
		},
		{
			name:  "LeadingSpace",
			input: " Bearer token",
			want:  " Bearer token",
		},
		{
			name:  "NotExactPrefix",
			input: "BearerBearer token",
			want:  "BearerBearer token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrimBearer(tt.input)
			if got != tt.want {
				t.Fatalf("TrimBearer(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
