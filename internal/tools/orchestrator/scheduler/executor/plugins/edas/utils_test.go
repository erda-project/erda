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

package edas

import (
	"reflect"
	"testing"
)

func TestAppendCommonHeaders(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected map[string]string
	}{
		{
			name:    "empty",
			headers: map[string]string{"hello": "world"},
			expected: map[string]string{
				"hello":         "world",
				"Cache-Control": "no-cache",
				"Pragma":        "no-cache",
				"Connection":    "keep-alive",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AppendCommonHeaders(tt.headers)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("appendCommonHeaders(), got: %v, want: %v", got, tt.expected)
			}
		})
	}
}

func TestEDASAppInfo(t *testing.T) {
	type args struct {
		sgType      string
		sgID        string
		serviceName string
	}

	tests := []struct {
		name   string
		args   args
		expect string
	}{
		{
			name: "compose app info",
			args: args{
				sgType:      "service",
				sgID:        "1",
				serviceName: "app-demo",
			},
			expect: "service-1-app-demo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, got := composeEDASAppInfo(tt.args.sgType, tt.args.sgID, tt.args.serviceName); got != tt.expect {
				t.Fatalf("composeEDASAppInfo, expecte: %v, got: %v", tt.expect, got)
			}
		})
	}
}
