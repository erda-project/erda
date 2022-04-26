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

package actionmgr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getActionTypeVersion(t *testing.T) {
	tests := []struct {
		name        string
		wantType    string
		wantVersion string
	}{
		{
			name:        "git",
			wantType:    "git",
			wantVersion: "",
		},
		{
			name:        "git@1.0",
			wantType:    "git",
			wantVersion: "1.0",
		},
		{
			name:        "git@1.0@1.0",
			wantType:    "git",
			wantVersion: "1.0@1.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getActionTypeVersion(tt.name)
			assert.Equalf(t, tt.wantType, got, "getActionTypeVersion(%v)", tt.name)
			assert.Equalf(t, tt.wantVersion, got1, "getActionTypeVersion(%v)", tt.name)
		})
	}
}
