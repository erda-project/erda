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

package errors

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParseValidateError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "test nil",
			args: args{
				err: nil,
			},
			want: nil,
		},
		{
			name: "test missing parameter",
			args: args{
				err: fmt.Errorf(`invalid field %s: value '%v' must not be an empty string`, "field1.field2", ""),
			},
			want: NewMissingParameterError("field1.field2"),
		},
		{
			name: "test missing parameter",
			args: args{
				err: fmt.Errorf(`invalid field %s: value '%v' must be greater than '0'`, "field1", "-1"),
			},
			want: NewInvalidParameterError("field1", fmt.Sprintf(`value '%v' must be greater than '0'`, "-1")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := ParseValidateError(tt.args.err); !reflect.DeepEqual((interface{})(err), (interface{})(tt.want)) {
				t.Errorf("ParseValidateError() got %v, want %v", err, tt.want)
			}
		})
	}
}
