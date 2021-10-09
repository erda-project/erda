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

package table

import (
	"context"
	"reflect"
	"testing"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/cmp"
	"github.com/erda-project/erda/modules/cmp/metrics"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func TestSortByNode(t *testing.T) {
	type args struct {
		data       []RowItem
		sortColumn string
		asc        bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				data: []RowItem{{
					Node: Node{
						Renders: []interface{}{
							[]interface{}{NodeLink{
								RenderType: "linkText",
								Value:      "123",
							},
							},
						},
					},
				}, {
					Node: Node{
						Renders: []interface{}{[]interface{}{
							NodeLink{
								RenderType: "linkText",
								Value:      "321",
							}, Labels{
								RenderType: "tagsRow",
							}},
						},
					},
				}, {
					ID: "1",
					Node: Node{
						Renders: []interface{}{[]interface{}{
							NodeLink{
								RenderType: "linkText",
								Value:      "123123",
							},
						},
						},
					},
				}},
				sortColumn: "Node",
				asc:        false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortByNode(tt.args.data, tt.args.sortColumn, tt.args.asc)
		})
	}
}

func TestSortByDistribution(t *testing.T) {
	type args struct {
		data       []RowItem
		sortColumn string
		asc        bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "testDistribution",
			args: args{
				data: []RowItem{{
					Distribution: Distribution{
						RenderType: "",
						Value:      "30",
						Status:     "",
						Tip:        "",
					},
				}, {
					Distribution: Distribution{
						RenderType: "",
						Value:      "20",
						Status:     "",
						Tip:        "",
					},
				}, {
					Distribution: Distribution{
						RenderType: "",
						Value:      "10",
						Status:     "",
						Tip:        "",
					},
				}},
				sortColumn: "Distribution",
				asc:        false,
			},
		},
		{
			name: "testUsage",
			args: args{
				data: []RowItem{{
					Usage: Distribution{
						RenderType: "",
						Value:      "28",
						Status:     "",
						Tip:        "",
					},
				}, {
					Usage: Distribution{
						RenderType: "",
						Value:      "30",
						Status:     "",
						Tip:        "",
					},
				}, {
					Usage: Distribution{
						RenderType: "",
						Value:      "10",
						Status:     "",
						Tip:        "",
					},
				}},
				sortColumn: "Usage",
				asc:        false,
			},
		},
		{
			name: "testUsageRate",
			args: args{
				data: []RowItem{{
					UnusedRate: Distribution{
						RenderType: "",
						Value:      "30",
						Status:     "",
						Tip:        "",
					},
				}, {
					UnusedRate: Distribution{
						RenderType: "",
						Value:      "10",
						Status:     "",
						Tip:        "",
					},
				}, {
					UnusedRate: Distribution{
						RenderType: "",
						Value:      "20",
						Status:     "",
						Tip:        "",
					},
				}},
				sortColumn: "UnusedRate",
				asc:        false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortByDistribution(tt.args.data, tt.args.sortColumn, tt.args.asc)
		})
	}
}

func TestSortByString(t *testing.T) {
	type args struct {
		data       []RowItem
		sortColumn string
		asc        bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "testRole",
			args: args{
				data: []RowItem{{
					Role: "worker",
				}, {
					Role: "lb",
				}, {
					Role: "master",
				}},
				sortColumn: "Role",
				asc:        false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortByString(tt.args.data, tt.args.sortColumn, tt.args.asc)
		})
	}
}

func TestSortByStatus(t *testing.T) {
	type args struct {
		data       []RowItem
		sortColumn string
		asc        bool
	}
	tests := []struct {
		name string
		args args
	}{{
		name: "testStatus",
		args: args{
			data: []RowItem{{
				Status: SteveStatus{
					Value: "Ready",
				},
			}, {
				Status: SteveStatus{
					Value: "Ready",
				},
			}, {
				Status: SteveStatus{
					Value: "error",
				},
			}},
			sortColumn: "Status",
			asc:        false,
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortByStatus(tt.args.data, tt.args.sortColumn, tt.args.asc)
		})
	}
}

func TestTable_GetScaleValue(t1 *testing.T) {
	type args struct {
		a            float64
		b            float64
		resourceType TableType
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				a:            1024,
				b:            102400000,
				resourceType: "text",
			},
			want: "1K/102400K",
		},
		{
			name: "test1",
			args: args{
				a:            1024,
				b:            1024,
				resourceType: "text",
			},
			want: "1K/1K",
		},
		{
			name: "test2",
			args: args{
				a:            2047,
				b:            1024,
				resourceType: Memory,
			},
			want: "2.0K/1.0K",
		},
		{
			name: "test3",
			args: args{
				a:            1024,
				b:            1024,
				resourceType: Cpu,
			},
			want: "1.024/1.024",
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Table{}
			if got := t.GetScaleValue(tt.args.a, tt.args.b, tt.args.resourceType); got != tt.want {
				t1.Errorf("GetScaleValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTable_GetUnusedRate(t1 *testing.T) {
	type fields struct {
	}
	type args struct {
		metricsData  *metrics.MetricsData
		resourceType TableType
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   DistributionValue
	}{
		// TODO: Add test cases.
		{
			name:   "text",
			fields: fields{},
			args: args{
				metricsData: &metrics.MetricsData{Used: 1, Request: 1.2,
					Total: 10},
				resourceType: Cpu,
			},
			want: DistributionValue{"0.200/1.200", "16.7"},
		},
		{
			name:   "text",
			fields: fields{},
			args: args{
				metricsData: &metrics.MetricsData{Used: 1, Request: 1.2,
					Total: 10},
				resourceType: Memory,
			},
			want: DistributionValue{"0.2/1.2", "16.7"},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Table{}
			if got := t.GetUnusedRate(tt.args.metricsData, tt.args.resourceType); !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("GetUnusedRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTable_GetDistributionValue(t1 *testing.T) {
	type fields struct {
		TableComponent  GetRowItem
		DefaultProvider base.DefaultProvider
		SDK             *cptype.SDK
		Ctx             context.Context
		Server          cmp.SteveServer
		Type            string
		Props           map[string]interface{}
		Operations      map[string]interface{}
		State           State
	}
	type args struct {
		metricsData  *metrics.MetricsData
		resourceType TableType
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   DistributionValue
	}{
		{
			name: "1",
			args: args{
				metricsData:  &metrics.MetricsData{1, 2, 3},
				resourceType: Pod,
			},
			want: DistributionValue{
				Text:    "2/3",
				Percent: "66.7",
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Table{}
			if got := t.GetDistributionValue(tt.args.metricsData, tt.args.resourceType); !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("GetDistributionValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTable_GetUsageValue(t1 *testing.T) {
	type fields struct {
		TableComponent  GetRowItem
		DefaultProvider base.DefaultProvider
		SDK             *cptype.SDK
		Ctx             context.Context
		Server          cmp.SteveServer
		Type            string
		Props           map[string]interface{}
		Operations      map[string]interface{}
		State           State
	}
	type args struct {
		metricsData  *metrics.MetricsData
		resourceType TableType
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   DistributionValue
	}{
		{
			name: "1",
			args: args{
				metricsData:  &metrics.MetricsData{1, 2, 3},
				resourceType: Pod,
			},
			want: DistributionValue{
				Text:    "1/3",
				Percent: "33.3",
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Table{}
			if got := t.GetUsageValue(tt.args.metricsData, tt.args.resourceType); !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("GetUsageValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
