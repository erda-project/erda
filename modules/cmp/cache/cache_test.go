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

package cache

import (
	"reflect"
	"testing"
)

func TestCache_DecrementSize(t *testing.T) {
	cache := New(100)
	type args struct {
		size int64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "DecrementTest",
			args: args{
				100,
			},
		},
		{
			name: "DecrementTest",

			args: args{
				100,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := cache.DecrementSize(tt.args.size)
			if err != nil {

			}
		})
	}
}

func TestCache_Get(t *testing.T) {
	cache := New(100)
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    Values
		wantErr bool
	}{
		{"Get_Test",

			args{"metrics1"},
			Values{IntValue{0, 1}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.WriteMulti(map[string]bool{"metrics1": false}, map[string]Values{"metrics1": {IntValue{
				unixnano: 0,
				value:    1,
			}}})

			got, err := cache.Get("metrics1")

			if err != nil {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, IntValue{1, 2})
			}

			got, err = cache.Get("metrics2")

			if err == nil {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, IntValue{1, 2})
			}
		})
	}
}

func TestCache_IncreaseSize(t *testing.T) {
	cache := New(100)
	type args struct {
		size int64
	}
	tests := []struct {
		name string
		args args
	}{
		{"IncreaseSize_Test",
			args{
				10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache.IncreaseSize(tt.args.size)
		})
	}
}

func TestCache_Remove(t *testing.T) {
	cache := New(100)
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "RemoveTest",
			args:    args{key: "metrics1"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache.WriteMulti(map[string]bool{"metrics1": false}, map[string]Values{"metrics1": {IntValue{
				unixnano: 0,
				value:    0,
			}}})
			if err := cache.Remove(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCache_WriteMulti(t *testing.T) {
	cache := New(100)

	type args struct {
		serializable map[string]bool
		pairs        map[string]Values
	}
	tests := []struct {
		name string

		args    args
		wantErr bool
	}{
		{
			name: "RemoveTest",

			args: args{
				serializable: map[string]bool{
					"metrics1": false,
					"metrics2": true},
				pairs: map[string]Values{
					"metrics1": {IntValue{
						unixnano: 0,
						value:    0,
					}},
					"metrics2": {
						IntValue{
							unixnano: 1,
							value:    2,
						},
						IntValue{
							unixnano: 0,
							value:    1,
						}},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if err = cache.WriteMulti(tt.args.serializable, tt.args.pairs); (err != nil) != tt.wantErr {
				t.Errorf("WriteMulti() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := cache.WriteMulti(map[string]bool{"metrics1": false}, map[string]Values{"metrics1": {StringValue{
				unixnano: 0,
				value:    "1231231",
			}}}); (err != nil) != tt.wantErr {
				t.Errorf("WriteMulti() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := cache.Get("metrics1")
			if !reflect.DeepEqual(got, Values{StringValue{
				unixnano: 0,
				value:    "1231231",
			}}) {
				t.Errorf("WriteMulti() want = %v ,got = %v", Values{StringValue{
					unixnano: 0,
					value:    "1231231",
				}}, got)
				return
			}

			err = cache.WriteMulti(map[string]bool{"metrics2": true}, map[string]Values{"metrics2": {
				IntValue{
					unixnano: 1,
					value:    2,
				}, IntValue{
					unixnano: 2,
					value:    2,
				}, IntValue{
					unixnano: 3,
					value:    2,
				}, IntValue{
					unixnano: 4,
					value:    2,
				}}})
			err = cache.WriteMulti(map[string]bool{"metrics2": true}, map[string]Values{"metrics2": {
				IntValue{
					unixnano: 3,
					value:    2,
				}, IntValue{
					unixnano: 5,
					value:    2,
				}}})
			if err != nil {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCache_Write(t *testing.T) {
	cache := New(1024 * 100)

	type args struct {
		serializable map[string]bool
		pairs        map[string]Values
	}
	tests := []struct {
		name string

		args    args
		wantErr bool
	}{
		{
			name: "WriteTest",
			args: args{
				serializable: map[string]bool{
					"metricsInt":        false,
					"metricsStr":        false,
					"metricsFloat":      false,
					"metricsUint":       false,
					"metricsBool":       false,
					"metricsIntSeria":   true,
					"metricsStrSeria":   true,
					"metricsFloatSeria": true,
					"metricsUintSeria":  true,
					"metricsBoolSeria":  true,
				},
				pairs: map[string]Values{
					"metricsInt": {
						IntValue{
							unixnano: 0,
							value:    0,
						},
						IntValue{
							unixnano: 0,
							value:    10,
						},
					},

					"metricsStr": {
						StringValue{
							unixnano: 1,
							value:    "123123131",
						},
						StringValue{
							unixnano: 1,
							value:    "3213123",
						},
						StringValue{
							unixnano: 1,
							value:    "4121231",
						},
					},
					"metricsFloat": {
						FloatValue{
							unixnano: 1,
							value:    3.1415,
						},
						FloatValue{
							unixnano: 0,
							value:    3.32,
						},
					},
					"metricsUint": {
						UnsignedValue{
							unixnano: 1,
							value:    ^uint64(0),
						},
						UnsignedValue{
							unixnano: 2,
							value:    ^uint64(0) >> 1,
						},
					},
					"metricsBool": {
						BoolValue{
							unixnano: 0,
							value:    true,
						},
						BoolValue{
							unixnano: 2,
							value:    true,
						},
					},
					"metricsIntSeria": {
						IntValue{
							unixnano: 0,
							value:    0,
						}, IntValue{
							unixnano: 100,
							value:    2,
						},
					},
					"metricsStrSeria": {
						StringValue{
							unixnano: 1,
							value:    "123123131",
						},
					},
					"metricsFloatSeria": {
						FloatValue{
							unixnano: 1,
							value:    3.1415,
						},
						FloatValue{
							unixnano: 2,
							value:    3.1414,
						},
					},
					"metricsUintSeria": {
						UnsignedValue{
							unixnano: 1,
							value:    ^uint64(0),
						},
						UnsignedValue{
							unixnano: 1,
							value:    0,
						},
					},
					"metricsBoolSeria": {
						BoolValue{
							unixnano: 0,
							value:    true,
						},
						BoolValue{
							unixnano: 2,
							value:    true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "UpdateTest",
			args: args{
				serializable: map[string]bool{
					"metricsInt":   false,
					"metricsStr":   false,
					"metricsFloat": false,
					"metricsUint":  false,
					"metricsIntSeria":  true,
				},
				pairs: map[string]Values{
					"metricsInt": {IntValue{
						unixnano: 1,
						value:    1,
					}},
					"metricsStr": {
						StringValue{
							unixnano: 2,
							value:    "31",
						},
					},
					"metricsFloat": {
						FloatValue{
							unixnano: 3,
							value:    3.52414124124,
						},
					},
					"metricsUint": {
						UnsignedValue{
							unixnano: 4,
							value:    ^uint64(0) >> 1,
						},
					},
					"metricsIntSeria": {
						IntValue{
							unixnano: 1,
							value:    200,
						},
						IntValue{
							unixnano: 10,
							value:    200,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "WriteBigDataTest",
			args: args{
				serializable: map[string]bool{
					"metricsStr": false,
				},
				pairs: map[string]Values{

					"metricsStr": {
						StringValue{
							unixnano: 2,
							value:    string(make([]byte, 1024*1024)),
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cache.WriteMulti(tt.args.serializable, tt.args.pairs); (err != nil) != tt.wantErr {
				t.Errorf("WriteMulti() error = %v, wantErr %v", err, tt.wantErr)
			}
		})

	}
}
