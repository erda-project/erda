// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
