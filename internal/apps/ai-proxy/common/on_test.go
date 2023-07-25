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

package common_test

import (
	"net/http"
	"testing"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common"
)

func TestOn_On(t *testing.T) {
	// creating sample headers
	header := make(http.Header)
	header.Set("TestKey", "TestValue")

	// defining the test cases
	tests := []struct {
		name    string
		on      common.On
		header  http.Header
		want    bool
		wantErr bool
	}{
		{
			name:    "Test Case 1: Key Exists",
			on:      common.On{Key: "TestKey", Operator: "exist", Value: ""},
			header:  header,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Test Case 2: Key Value Match",
			on:      common.On{Key: "TestKey", Operator: "=", Value: "TestValue"},
			header:  header,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Test Case 3: Key Value Mismatch",
			on:      common.On{Key: "TestKey", Operator: "=", Value: "WrongValue"},
			header:  header,
			want:    false,
			wantErr: false,
		},
		{
			name:    "Test Case 4: Operator Error",
			on:      common.On{Key: "TestKey", Operator: "+", Value: "TestValue"},
			header:  header,
			want:    false,
			wantErr: true,
		},
	}

	// running the tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.on.On(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("On() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("On() got = %v, want %v", got, tt.want)
			}
		})
	}
}
