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

package query

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
)

func TestDepthCopyQueryConditions(t *testing.T) {
	tests := []struct {
		name string
		want *pb.TraceQueryConditions
	}{
		{"case1", &TraceQueryConditions},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DepthCopyQueryConditions()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DepthCopyQueryConditions() = %v, want %v", got, tt.want)
			}
			// got point
			gotPoint := getMemoryPoint(got)
			gotPointOthers := getMemoryPoint(got.Others)
			gotPointLimit := getMemoryPoint(got.Limit)
			gotPointSort := getMemoryPoint(got.Sort)
			gotPointTraceStatus := getMemoryPoint(got.TraceStatus)

			// TraceQueryConditions point
			wantPoint := getMemoryPoint(tt.want)
			wantPointOthers := getMemoryPoint(tt.want.Others)
			wantPointLimit := getMemoryPoint(tt.want.Limit)
			wantPointSort := getMemoryPoint(tt.want.Sort)
			wantPointTraceStatus := getMemoryPoint(tt.want.TraceStatus)

			if gotPoint == wantPoint {
				t.Errorf("gotPointServiceName = %v, wantPointServiceName %v", gotPoint, wantPoint)
			}
			if gotPointOthers == wantPointOthers {
				t.Errorf("gotPointOthers = %v, wantPointOthers %v", gotPointOthers, wantPointOthers)
			}
			if gotPointLimit == wantPointLimit {
				t.Errorf("gotPointServiceName = %v, wantPointServiceName %v", gotPointLimit, wantPointLimit)
			}
			if gotPointSort == wantPointSort {
				t.Errorf("gotPointServiceName = %v, wantPointServiceName %v", gotPointSort, wantPointSort)
			}
			if gotPointTraceStatus == wantPointTraceStatus {
				t.Errorf("gotPointServiceName = %v, wantPointServiceName %v", gotPointTraceStatus, wantPointTraceStatus)
			}
		})
	}
}

func getMemoryPoint(need interface{}) string {
	return fmt.Sprintf("%p", need)
}

func Test_clone(t *testing.T) {
	type args struct {
		src *pb.TraceQueryConditions
	}
	tests := []struct {
		name    string
		args    args
		want    *pb.TraceQueryConditions
		wantErr bool
	}{
		{"case1", args{src: &TraceQueryConditions}, &TraceQueryConditions, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := clone(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("clone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("clone() got = %v, want %v", got, tt.want)
			}
			gotPoint := getMemoryPoint(got)
			wantPoint := getMemoryPoint(tt.want)
			if gotPoint == wantPoint {
				t.Errorf("gotPointServiceName = %v, wantPointServiceName %v", gotPoint, wantPoint)
			}
		})
	}
}
