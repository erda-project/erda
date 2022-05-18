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

package source

import (
	"testing"

	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
)

func Test_chSpanCovertToSpan(t *testing.T) {
	type args struct {
		span *pb.Span
		cs   spanSeries
	}
	tests := []struct {
		name string
		args args
	}{
		{"case1", args{span: &pb.Span{}, cs: spanSeries{
			SpanId:        "test",
			TraceId:       "test",
			OperationName: "test",
			StartTime:     2,
			EndTime:       2,
			ParentSpanId:  "test_p",
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chSpanCovertToSpan(tt.args.span, tt.args.cs)
			if tt.args.span.Id != tt.args.cs.SpanId {
				t.Errorf("SpanId id not equal. span.Id: %s, cs.SpanId: %s", tt.args.span.Id, tt.args.cs.SpanId)
			}
			if tt.args.span.TraceId != tt.args.cs.TraceId {
				t.Errorf("TraceId not equal. span.TraceId: %s, cs.TraceId: %s", tt.args.span.TraceId, tt.args.cs.TraceId)
			}
			if tt.args.span.OperationName != tt.args.cs.OperationName {
				t.Errorf("OperationName not equal. span.OperationName: %s, cs.OperationName: %s", tt.args.span.OperationName, tt.args.cs.OperationName)
			}
			if tt.args.span.StartTime != tt.args.cs.StartTime {
				t.Errorf("StartTime not equal. StartTime: %v, cs.StartTime: %v", tt.args.span.StartTime, tt.args.cs.StartTime)
			}
			if tt.args.span.EndTime != tt.args.cs.EndTime {
				t.Errorf("EndTime not equal. span.EndTime: %v, cs.EndTime: %v", tt.args.span.EndTime, tt.args.cs.EndTime)
			}
			if tt.args.span.ParentSpanId != tt.args.cs.ParentSpanId {
				t.Errorf("ParentSpanId not equal. span.ParentSpanId: %s, cs.ParentSpanId: %s", tt.args.span.ParentSpanId, tt.args.cs.ParentSpanId)
			}
		})
	}
}
