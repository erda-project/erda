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

package dbclient

import (
	"reflect"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
)

func Test_makeOrderEvents(t *testing.T) {
	var nowTime = time.Now()
	type args struct {
		events    []*pb.PipelineEvent
		newEvents []*pb.PipelineEvent
	}
	tests := []struct {
		name string
		args args
		want orderedEvents
	}{
		{
			name: "empty events add one newEvents",
			args: args{
				events: nil,
				newEvents: []*pb.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(20 * time.Second)),
						Count:          1,
						Type:           "type",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: &pb.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(20 * time.Second)),
					Count:          1,
					Type:           "type",
				},
			},
		},
		{
			name: "empty events add two differ newEvents",
			args: args{
				events: nil,
				newEvents: []*pb.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(20 * time.Second)),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason1",
						Message: "message1",
						Source: &pb.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
						Count:          1,
						Type:           "type",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: &pb.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(20 * time.Second)),
					Count:          1,
					Type:           "type",
				},
				{
					Reason:  "reason1",
					Message: "message1",
					Source: &pb.PipelineEventSource{
						Component: "component1",
						Host:      "host1",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
					Count:          1,
					Type:           "type",
				},
			},
		},
		{
			name: "empty events add two same newEvents",
			args: args{
				events: nil,
				newEvents: []*pb.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(20 * time.Second)),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
						Count:          1,
						Type:           "type",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: &pb.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
					Count:          2,
					Type:           "type",
				},
			},
		},
		{
			name: "empty events add three no order lastTimestamp newEvents",
			args: args{
				events: nil,
				newEvents: []*pb.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(20 * time.Second)),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason1",
						Message: "message1",
						Source: &pb.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(10 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(30 * time.Second)),
						Count:          1,
						Type:           "type1",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: &pb.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(20 * time.Second)),
					Count:          1,
					Type:           "type",
				},
				{
					Reason:  "reason1",
					Message: "message1",
					Source: &pb.PipelineEventSource{
						Component: "component1",
						Host:      "host1",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(10 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(30 * time.Second)),
					Count:          1,
					Type:           "type1",
				},
				{
					Reason:  "reason",
					Message: "message",
					Source: &pb.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
					Count:          1,
					Type:           "type",
				},
			},
		},
		{
			name: "empty events add three order lastTimestamp newEvents",
			args: args{
				events: nil,
				newEvents: []*pb.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(20 * time.Second)),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason1",
						Message: "message1",
						Source: &pb.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(10 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(30 * time.Second)),
						Count:          1,
						Type:           "type1",
					},
					{
						Reason:  "reason2",
						Message: "message2",
						Source: &pb.PipelineEventSource{
							Component: "component2",
							Host:      "host2",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
						Count:          1,
						Type:           "type2",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: &pb.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(20 * time.Second)),
					Count:          1,
					Type:           "type",
				},
				{
					Reason:  "reason1",
					Message: "message1",
					Source: &pb.PipelineEventSource{
						Component: "component1",
						Host:      "host1",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(10 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(30 * time.Second)),
					Count:          1,
					Type:           "type1",
				},
				{
					Reason:  "reason2",
					Message: "message2",
					Source: &pb.PipelineEventSource{
						Component: "component2",
						Host:      "host2",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
					Count:          1,
					Type:           "type2",
				},
			},
		},
		{
			name: "empty events add three no order lastTimestamp newEvents，two events count++",
			args: args{
				events: nil,
				newEvents: []*pb.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(20 * time.Second)),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(30 * time.Second)),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason1",
						Message: "message1",
						Source: &pb.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(10 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
						Count:          1,
						Type:           "type1",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: &pb.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(30 * time.Second)),
					Count:          2,
					Type:           "type",
				},
				{
					Reason:  "reason1",
					Message: "message1",
					Source: &pb.PipelineEventSource{
						Component: "component1",
						Host:      "host1",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(10 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
					Count:          1,
					Type:           "type1",
				},
			},
		},
		{
			name: "empty events add three no order lastTimestamp newEvents，three events count++",
			args: args{
				events: nil,
				newEvents: []*pb.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(20 * time.Second)),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(30 * time.Second)),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(10 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
						Count:          1,
						Type:           "type",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: &pb.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
					Count:          3,
					Type:           "type",
				},
			},
		},
		{
			name: "events add in order",
			args: args{
				events: []*pb.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
						Count:          3,
						Type:           "type",
					},
				},
				newEvents: []*pb.PipelineEvent{
					{
						Reason:  "reason1",
						Message: "message1",
						Source: &pb.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(50 * time.Second)),
						Count:          1,
						Type:           "type",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: &pb.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
					Count:          3,
					Type:           "type",
				},
				{
					Reason:  "reason1",
					Message: "message1",
					Source: &pb.PipelineEventSource{
						Component: "component1",
						Host:      "host1",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
					Count:          1,
					Type:           "type",
				},
				{
					Reason:  "reason",
					Message: "message",
					Source: &pb.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(50 * time.Second)),
					Count:          1,
					Type:           "type",
				},
			},
		},
		{
			name: "events first add counts",
			args: args{
				events: []*pb.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
						Count:          3,
						Type:           "type",
					},
				},
				newEvents: []*pb.PipelineEvent{
					{
						Reason:  "reason1",
						Message: "message1",
						Source: &pb.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(60 * time.Second)),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(50 * time.Second)),
						Count:          1,
						Type:           "type",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: &pb.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(50 * time.Second)),
					Count:          4,
					Type:           "type",
				},
				{
					Reason:  "reason1",
					Message: "message1",
					Source: &pb.PipelineEventSource{
						Component: "component1",
						Host:      "host1",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(60 * time.Second)),
					Count:          1,
					Type:           "type",
				},
			},
		},
		{
			name: "events first add counts three add counts",
			args: args{
				events: []*pb.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
						Count:          3,
						Type:           "type",
					},
				},
				newEvents: []*pb.PipelineEvent{
					{
						Reason:  "reason1",
						Message: "message1",
						Source: &pb.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(60 * time.Second)),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason1",
						Message: "message1",
						Source: &pb.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(-30 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(80 * time.Second)),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason2",
						Message: "message2",
						Source: &pb.PipelineEventSource{
							Component: "component2",
							Host:      "host2",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(90 * time.Second)),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(50 * time.Second)),
						Count:          1,
						Type:           "type",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: &pb.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(50 * time.Second)),
					Count:          4,
					Type:           "type",
				},
				{
					Reason:  "reason1",
					Message: "message1",
					Source: &pb.PipelineEventSource{
						Component: "component1",
						Host:      "host1",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(-30 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(80 * time.Second)),
					Count:          2,
					Type:           "type",
				},
				{
					Reason:  "reason2",
					Message: "message2",
					Source: &pb.PipelineEventSource{
						Component: "component2",
						Host:      "host2",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(90 * time.Second)),
					Count:          1,
					Type:           "type",
				},
			},
		},

		{
			name: "sort events and add events",
			args: args{
				events: []*pb.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: &pb.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
						Count:          3,
						Type:           "type",
					},
					{
						Reason:  "reason1",
						Message: "message1",
						Source: &pb.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(30 * time.Second)),
						Count:          1,
						Type:           "type",
					},
				},
				newEvents: []*pb.PipelineEvent{
					{
						Reason:  "reason2",
						Message: "message2",
						Source: &pb.PipelineEventSource{
							Component: "component2",
							Host:      "host2",
						},
						FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
						LastTimestamp:  timestamppb.New(nowTime.Add(90 * time.Second)),
						Count:          1,
						Type:           "type",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason1",
					Message: "message1",
					Source: &pb.PipelineEventSource{
						Component: "component1",
						Host:      "host1",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(30 * time.Second)),
					Count:          1,
					Type:           "type",
				},
				{
					Reason:  "reason",
					Message: "message",
					Source: &pb.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(-20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(40 * time.Second)),
					Count:          3,
					Type:           "type",
				},
				{
					Reason:  "reason2",
					Message: "message2",
					Source: &pb.PipelineEventSource{
						Component: "component2",
						Host:      "host2",
					},
					FirstTimestamp: timestamppb.New(nowTime.Add(20 * time.Second)),
					LastTimestamp:  timestamppb.New(nowTime.Add(90 * time.Second)),
					Count:          1,
					Type:           "type",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeOrderEvents(tt.args.events, tt.args.newEvents); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeOrderEvents() = %v, want %v", got, tt.want)
			}
		})
	}
}
