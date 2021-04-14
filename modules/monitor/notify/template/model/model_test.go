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

package model

import "testing"

func TestValidateString(t *testing.T) {
	type args struct {
		field string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test_ValidateString",
			args: args{},
			want: "this field not be empty",
		},
		{
			name: "test_ValidateString",
			args: args{
				field: "ddfdf",
			},
			want: "",
		},
		{
			name: "test_ValidateString",
			args: args{
				field: "      ",
			},
			want: `this field can't just contain spaces`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateString(tt.args.field); got != tt.want {
				t.Errorf("ValidateString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateInt(t *testing.T) {
	type args struct {
		field int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test_ValidateInt",
			args: args{
				34,
			},
			want: "",
		},
		{
			name: "test_ValidateInt",
			args: args{
				0,
			},
			want: "this field must not be zero",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateInt(tt.args.field); got != tt.want {
				t.Errorf("ValidateInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckElements(t *testing.T) {
	type args struct {
		formats  []map[string]string
		title    []string
		template []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_CheckElements",
			args: args{
				formats: []map[string]string{
					{
						"template": "notify template",
					},
				},
				title:    []string{"template title"},
				template: []string{"this is notify template"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckElements(tt.args.formats, tt.args.title, tt.args.template); (err != nil) != tt.wantErr {
				t.Errorf("CheckElements() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
