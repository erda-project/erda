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

package scatterPlot

import (
	"reflect"
	"testing"
	"time"

	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

func TestScatterData(t *testing.T) {
	type args struct {
		issues []dao.IssueItem
	}
	t1 := time.Now()
	t2 := t1.AddDate(0, 0, 1)
	t3 := t1.AddDate(0, 0, -1)
	tests := []struct {
		name string
		args args
		want [][]float32
	}{
		{
			name: "test",
			args: args{
				issues: []dao.IssueItem{
					{
						StartTime: &t1,
					},
					{
						FinishTime: &t1,
					},
					{
						BaseModel: dbengine.BaseModel{
							CreatedAt: t3,
						},
						StartTime:  &t1,
						FinishTime: &t2,
					},
				},
			},
			want: [][]float32{
				{
					24, 24,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScatterData(tt.args.issues)
			if len(got.Option.Series) == 0 {
				return
			}
			expect := got.Option.Series[0].Data
			if !reflect.DeepEqual(expect, tt.want) {
				t.Errorf("ScatterData() = %v, want %v", expect, tt.want)
			}
		})
	}
}

func Test_milliToHour(t *testing.T) {
	type args struct {
		m int64
	}
	tests := []struct {
		name string
		args args
		want float32
	}{
		{
			name: "test",
			args: args{
				m: 36000000,
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := milliToHour(tt.args.m); got != tt.want {
				t.Errorf("milliToHour() = %v, want %v", got, tt.want)
			}
		})
	}
}
