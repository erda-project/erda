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

package list

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/i18n"
)

type NopTranslator struct{}

func (t NopTranslator) Get(lang i18n.LanguageCodes, key, def string) string { return key }

func (t NopTranslator) Text(lang i18n.LanguageCodes, key string) string { return key }

func (t NopTranslator) Sprintf(lang i18n.LanguageCodes, key string, args ...interface{}) string {
	return fmt.Sprintf(key, args...)
}

func TestList_GetExtraContent(t *testing.T) {
	type args struct {
		res *ResData
	}
	tests := []struct {
		name string
		args args
		want ExtraContent
	}{
		// TODO: Add test cases.
		{
			name: "1",
			args: args{res: &ResData{
				CpuUsed:     1,
				CpuTotal:    2,
				MemoryUsed:  3,
				MemoryTotal: 4,
				DiskUsed:    9,
				DiskTotal:   10,
			}},
			want: ExtraContent{
				Type:   "PieChart",
				RowNum: 3,
				ExtraData: []ExtraData{
					{
						Name:  "CPU Rate",
						Value: 50,
						Total: 100,
						Color: "green",
						Info: []ExtraDataItem{
							{
								Main: "50.000%",
								Sub:  "Rate",
							}, {
								Main: "1.000core",
								Sub:  "Distribution",
							}, {
								Main: "2.000core",
								Sub:  "CPU" + "Quota",
							},
						},
					},
					{
						Name:  "Memory Rate",
						Value: 75,
						Total: 100,
						Color: "green",
						Info: []ExtraDataItem{
							{
								Main: "75.000%",
								Sub:  "Rate",
							}, {
								Main: "3.000",
								Sub:  "Distribution",
							}, {
								Main: "4.000",
								Sub:  "Memory" + "Quota",
							},
						},
					},
					{
						Name:  "Disk Rate",
						Value: 90,
						Total: 100,
						Color: "green",
						Info: []ExtraDataItem{
							{
								Main: "90.000%",
								Sub:  "Rate",
							}, {
								Main: "9.000",
								Sub:  "Distribution",
							}, {
								Main: "10.000",
								Sub:  "Disk" + "Quota",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &List{
				SDK: &cptype.SDK{Tran: NopTranslator{}},
			}
			if got := l.GetExtraContent(tt.args.res); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetExtraContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestList_GetExtraInfos(t *testing.T) {
	type args struct {
		clusterInfo *ClusterInfoDetail
	}
	tests := []struct {
		name string
		args args
		want []ExtraInfos
	}{
		// TODO: Add test cases.
		{
			name: "1",
			args: args{clusterInfo: &ClusterInfoDetail{}},
			want: []ExtraInfos{
				ExtraInfos{
					Icon:    "management",
					Text:    "-",
					Tooltip: "manage type",
				},
				ExtraInfos{
					Icon:    "create-time",
					Text:    "-",
					Tooltip: "create time",
				},
				ExtraInfos{
					Icon:    "machine",
					Text:    "0",
					Tooltip: "machine count",
				},
				ExtraInfos{
					Icon:    "type",
					Text:    "-",
					Tooltip: "cluster type",
				},
				ExtraInfos{
					Icon:    "version",
					Text:    "-",
					Tooltip: "cluster version",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &List{
				SDK: &cptype.SDK{Tran: NopTranslator{}},
			}
			if got := l.GetExtraInfos(tt.args.clusterInfo); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetExtraInfos() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestList_SetComponentValue(t *testing.T) {
	type args struct {
		c *cptype.Component
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "1",
			args:    args{c: &cptype.Component{}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &List{}
			if err := l.SetComponentValue(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("SetComponentValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
