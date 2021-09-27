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

package type_conversion

import "testing"

func TestInterfaceToUint64(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "test_int",
			args: args{
				value: 1,
			},
			want:    uint64(1),
			wantErr: false,
		},
		{
			name: "test_string",
			args: args{
				value: "1",
			},
			want:    uint64(1),
			wantErr: false,
		},
		{
			name: "test_float64",
			args: args{
				value: float64(1),
			},
			want:    uint64(1),
			wantErr: false,
		},
		{
			name: "test_empty",
			args: args{
				value: nil,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "test_other",
			args: args{
				value: int32(1),
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InterfaceToUint64(tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("InterfaceToUint64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("InterfaceToUint64() got = %v, want %v", got, tt.want)
			}
		})
	}
}
