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
	"testing"
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
					UsageRate: Distribution{
						RenderType: "",
						Value:      "30",
						Status:     "",
						Tip:        "",
					},
				}, {
					UsageRate: Distribution{
						RenderType: "",
						Value:      "10",
						Status:     "",
						Tip:        "",
					},
				}, {
					UsageRate: Distribution{
						RenderType: "",
						Value:      "20",
						Status:     "",
						Tip:        "",
					},
				}},
				sortColumn: "UsageRate",
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

