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

package permission

import (
	"context"
	"testing"
)

func TestFieldValue(t *testing.T) {
	type args struct {
		field string
		req   interface{}
	}

	tests := []struct {
		name      string
		args      args
		want      string
		wantError bool
	}{
		{
			name: "struct",
			args: args{
				req: &struct {
					Field1 string
				}{Field1: "fieldval1"},
				field: "Field1",
			},
			want: "fieldval1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getter := FieldValue(tt.args.field)
			val, err := getter(context.Background(), tt.args.req)
			if err != nil {
				if !tt.wantError {
					t.Errorf("FieldValue(%q)(req) got error: %s", tt.args.field, err)
				}
			} else if err == nil {
				if tt.wantError {
					t.Errorf("FieldValue(%q)(req) successfully, want error", tt.args.field)
				} else if val != tt.want {
					t.Errorf("FieldValue(%q)(req) = %v, want %v", tt.args.field, val, tt.want)
				}
			}
		})
	}
}

type testInterface interface {
	TestFunc_1()
	函数_1()
}

func Test_getMethodName(t *testing.T) {
	type args struct {
		method interface{}
	}
	tests := []struct {
		name      string
		args      args
		want      string
		wantError bool
	}{
		{
			name: "test empty",
			args: args{},
			want: "",
		},
		{
			name: "test string",
			args: args{
				method: "hello",
			},
			want: "hello",
		},
		{
			name: "interafce method",
			args: args{
				method: testInterface.TestFunc_1,
			},
			want: "TestFunc_1",
		},
		{
			name: "interafce Chinese method name",
			args: args{
				method: testInterface.函数_1,
			},
			want: "函数_1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				err := recover()
				if err == nil && tt.wantError {
					t.Errorf("getMethodName() successfully, want error")
				} else if err != nil && !tt.wantError {
					t.Errorf("getMethodName() got error: %s", err)
				}
			}()
			if got := getMethodName(tt.args.method); got != tt.want {
				t.Errorf("getMethodName() = %v, want %v", got, tt.want)
			}
		})
	}
}
