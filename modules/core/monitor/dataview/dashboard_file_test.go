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

package dataview

import (
	"strings"
	"testing"
)

func Test_dashboardFileName(t *testing.T) {
	type args struct {
		scope   string
		scopeId string
		viewIds []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"case1", args{
				scope:   "org",
				scopeId: "1",
				viewIds: []string{"1"},
			}, "b3JnLTEtMj",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dashboardFilename(tt.args.scope, tt.args.scopeId); !strings.HasPrefix(got, tt.want) {
				t.Errorf("dashboardFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompileToDest(t *testing.T) {
	type args struct {
		scope   string
		scopeId string
		data    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"case1", args{scope: "micro_service", scopeId: "test", data: `filter__metric_scope_id\":\"xxxx\", this is a test`}, `filter__metric_scope_id\":\"test\", this is a test`},
		{"case2", args{scope: "micro_service", scopeId: "test", data: `filter__metric_scope_id":"xxxx", this is a test`}, `filter__metric_scope_id":"test", this is a test`},
		{"case3", args{scope: "micro_service", scopeId: "test", data: `filter_terminus_key\":\"xxxx\", this is a test`}, `filter_terminus_key\":\"test\", this is a test`},
		{"case4", args{scope: "micro_service", scopeId: "test", data: `filter_terminus_key":"xxxx", this is a test`}, `filter_terminus_key":"test", this is a test`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompileToDest(tt.args.scope, tt.args.scopeId, tt.args.data); got != tt.want {
				t.Errorf("CompileToDest() = %v, want %v", got, tt.want)
			}
		})
	}
}
