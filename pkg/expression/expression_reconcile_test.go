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

package expression

import (
	"fmt"
	"testing"
)

func Test_expression(t *testing.T) {

	var testData = []struct {
		express string
		result  string
	}{
		{
			"${{ '${{' == '}}' }}", " '${{' == '}}' ",
		},
		{
			"${{ '}}' == '}}' }}", " '}}' == '}}' ",
		},
		{
			"${{ '1' == '}}' }}", " '1' == '}}' ",
		},
		{
			"${{ '${{' == '2' }}", " '${{' == '2' ",
		},
		{
			"${{ '1' == '2' }}", " '1' == '2' ",
		},
		{
			"${{ '2' == '2' }}", " '2' == '2' ",
		},
		{
			"${{ '' == '2' }}", " '' == '2' ",
		},
		{
			"${{${{ == '2' }}}}", "${{ == '2' }}",
		},
	}

	for _, condition := range testData {
		if ReplacePlaceholder(condition.express) != condition.result {
			fmt.Println(" error express ", condition.express)
			t.Fail()
		}
	}
}

func TestReconcile(t *testing.T) {
	type args struct {
		condition string
	}
	tests := []struct {
		name     string
		args     args
		wantSign SignType
	}{
		{
			name:     "invalid aa",
			args:     args{condition: "aa"},
			wantSign: TaskJumpOver,
		},
		{
			name:     "1 != 2",
			args:     args{condition: "1 != 2"},
			wantSign: TaskNotJumpOver,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSign := Reconcile(tt.args.condition); gotSign.Sign != tt.wantSign {
				t.Errorf("Reconcile() = %v, want %v", gotSign, tt.wantSign)
			}
		})
	}
}
