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

package mcp_server

import (
	"testing"

	"gotest.tools/assert"
)

func TestBuildConstraint(t *testing.T) {
	tests := []struct {
		target string
		want   string
	}{
		{
			target: "1",
			want:   ">=1.0.0 <2.0.0",
		},
		{
			target: "1.2",
			want:   ">=1.2.0 <1.3.0",
		},
		{
			target: "1.2.3",
			want:   "=1.2.3",
		},
	}

	for _, tt := range tests {
		got, err := buildConstraint(tt.target)
		if err != nil {
			t.Errorf("buildConstraint(%s) error: %v", tt.target, err)
			continue
		}
		assert.Equal(t, tt.want, got.String())
	}
}
