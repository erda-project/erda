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

package dbclient

import (
	"reflect"
	"testing"
)

func Test_transferMustMatchLabelsToMap(t *testing.T) {
	type args struct {
		ss []string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string][]string
		wantErr bool
	}{
		{
			name: "three labels",
			args: args{
				ss: []string{"a=b", "a=c", "b=d"},
			},
			want: map[string][]string{
				"a": {"b", "c"},
				"b": {"d"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transferMustMatchLabelsToMap(tt.args.ss)
			if (err != nil) != tt.wantErr {
				t.Errorf("transferMustMatchLabelsToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("transferMustMatchLabelsToMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}
