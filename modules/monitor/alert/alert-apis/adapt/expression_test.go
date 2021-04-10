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

package adapt

import (
	"reflect"
	"testing"
)

func Test_convertMillisecondToUnit(t *testing.T) {
	type args struct {
		t int64
	}
	tests := []struct {
		name      string
		args      args
		wantValue int64
		wantUnit  string
	}{
		{
			name: "test_convertMillisecondToUnit",
			args: args{
				t: 48,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotUnit := convertMillisecondToUnit(tt.args.t)
			if gotValue != tt.wantValue {
				t.Errorf("convertMillisecondToUnit() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
			if gotUnit != tt.wantUnit {
				t.Errorf("convertMillisecondToUnit() gotUnit = %v, want %v", gotUnit, tt.wantUnit)
			}
		})
	}
}

func Test_convertMillisecondByUnit(t *testing.T) {
	type args struct {
		value int64
		unit  string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "test_convertMillisecondByUnit",
			args: args{
				value: 12345,
				unit:  "seconds",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertMillisecondByUnit(tt.args.value, tt.args.unit); got != tt.want {
				t.Errorf("convertMillisecondByUnit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertString(t *testing.T) {
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{
			name: "test_convertString",
			args: args{
				[]byte{97, 98, 99, 100},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := convertString(tt.args.obj)
			if got != tt.want {
				t.Errorf("convertString() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("convertString() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_convertDataByType(t *testing.T) {
	type args struct {
		obj      interface{}
		dataType string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "test_convertDataByType",
			args: args{
				obj:      "erda",
				dataType: "string",
			},
		},
		{
			name: "test_convertDataByType",
			args: args{
				obj:      1234,
				dataType: "string",
			},
		},
		{
			name: "test_convertDataByType",
			args: args{
				obj:      "erda",
				dataType: "number",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertDataByType(tt.args.obj, tt.args.dataType)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertDataByType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertDataByType() got = %v, want %v", got, tt.want)
			}
		})
	}
}
