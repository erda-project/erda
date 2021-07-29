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

package dbclient

import (
	"reflect"
	"testing"
	"time"

	"github.com/erda-project/erda/apistructs"
)

func Test_makeOrderEvents(t *testing.T) {
	var nowTime = time.Now()
	type args struct {
		events    []*apistructs.PipelineEvent
		newEvents []*apistructs.PipelineEvent
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
				newEvents: []*apistructs.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(-20 * time.Second),
						LastTimestamp:  nowTime.Add(20 * time.Second),
						Count:          1,
						Type:           "type",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: apistructs.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: nowTime.Add(-20 * time.Second),
					LastTimestamp:  nowTime.Add(20 * time.Second),
					Count:          1,
					Type:           "type",
				},
			},
		},
		{
			name: "empty events add two differ newEvents",
			args: args{
				events: nil,
				newEvents: []*apistructs.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(-20 * time.Second),
						LastTimestamp:  nowTime.Add(20 * time.Second),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason1",
						Message: "message1",
						Source: apistructs.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: nowTime.Add(20 * time.Second),
						LastTimestamp:  nowTime.Add(40 * time.Second),
						Count:          1,
						Type:           "type",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: apistructs.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: nowTime.Add(-20 * time.Second),
					LastTimestamp:  nowTime.Add(20 * time.Second),
					Count:          1,
					Type:           "type",
				},
				{
					Reason:  "reason1",
					Message: "message1",
					Source: apistructs.PipelineEventSource{
						Component: "component1",
						Host:      "host1",
					},
					FirstTimestamp: nowTime.Add(20 * time.Second),
					LastTimestamp:  nowTime.Add(40 * time.Second),
					Count:          1,
					Type:           "type",
				},
			},
		},
		{
			name: "empty events add two same newEvents",
			args: args{
				events: nil,
				newEvents: []*apistructs.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(-20 * time.Second),
						LastTimestamp:  nowTime.Add(20 * time.Second),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(20 * time.Second),
						LastTimestamp:  nowTime.Add(40 * time.Second),
						Count:          1,
						Type:           "type",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: apistructs.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: nowTime.Add(-20 * time.Second),
					LastTimestamp:  nowTime.Add(40 * time.Second),
					Count:          2,
					Type:           "type",
				},
			},
		},
		{
			name: "empty events add three no order lastTimestamp newEvents",
			args: args{
				events: nil,
				newEvents: []*apistructs.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(-20 * time.Second),
						LastTimestamp:  nowTime.Add(20 * time.Second),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(20 * time.Second),
						LastTimestamp:  nowTime.Add(40 * time.Second),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason1",
						Message: "message1",
						Source: apistructs.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: nowTime.Add(10 * time.Second),
						LastTimestamp:  nowTime.Add(30 * time.Second),
						Count:          1,
						Type:           "type1",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: apistructs.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: nowTime.Add(-20 * time.Second),
					LastTimestamp:  nowTime.Add(20 * time.Second),
					Count:          1,
					Type:           "type",
				},
				{
					Reason:  "reason1",
					Message: "message1",
					Source: apistructs.PipelineEventSource{
						Component: "component1",
						Host:      "host1",
					},
					FirstTimestamp: nowTime.Add(10 * time.Second),
					LastTimestamp:  nowTime.Add(30 * time.Second),
					Count:          1,
					Type:           "type1",
				},
				{
					Reason:  "reason",
					Message: "message",
					Source: apistructs.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: nowTime.Add(20 * time.Second),
					LastTimestamp:  nowTime.Add(40 * time.Second),
					Count:          1,
					Type:           "type",
				},
			},
		},
		{
			name: "empty events add three order lastTimestamp newEvents",
			args: args{
				events: nil,
				newEvents: []*apistructs.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(-20 * time.Second),
						LastTimestamp:  nowTime.Add(20 * time.Second),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason1",
						Message: "message1",
						Source: apistructs.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: nowTime.Add(10 * time.Second),
						LastTimestamp:  nowTime.Add(30 * time.Second),
						Count:          1,
						Type:           "type1",
					},
					{
						Reason:  "reason2",
						Message: "message2",
						Source: apistructs.PipelineEventSource{
							Component: "component2",
							Host:      "host2",
						},
						FirstTimestamp: nowTime.Add(20 * time.Second),
						LastTimestamp:  nowTime.Add(40 * time.Second),
						Count:          1,
						Type:           "type2",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: apistructs.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: nowTime.Add(-20 * time.Second),
					LastTimestamp:  nowTime.Add(20 * time.Second),
					Count:          1,
					Type:           "type",
				},
				{
					Reason:  "reason1",
					Message: "message1",
					Source: apistructs.PipelineEventSource{
						Component: "component1",
						Host:      "host1",
					},
					FirstTimestamp: nowTime.Add(10 * time.Second),
					LastTimestamp:  nowTime.Add(30 * time.Second),
					Count:          1,
					Type:           "type1",
				},
				{
					Reason:  "reason2",
					Message: "message2",
					Source: apistructs.PipelineEventSource{
						Component: "component2",
						Host:      "host2",
					},
					FirstTimestamp: nowTime.Add(20 * time.Second),
					LastTimestamp:  nowTime.Add(40 * time.Second),
					Count:          1,
					Type:           "type2",
				},
			},
		},
		{
			name: "empty events add three no order lastTimestamp newEvents，two events count++",
			args: args{
				events: nil,
				newEvents: []*apistructs.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(-20 * time.Second),
						LastTimestamp:  nowTime.Add(20 * time.Second),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(20 * time.Second),
						LastTimestamp:  nowTime.Add(30 * time.Second),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason1",
						Message: "message1",
						Source: apistructs.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: nowTime.Add(10 * time.Second),
						LastTimestamp:  nowTime.Add(40 * time.Second),
						Count:          1,
						Type:           "type1",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: apistructs.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: nowTime.Add(-20 * time.Second),
					LastTimestamp:  nowTime.Add(30 * time.Second),
					Count:          2,
					Type:           "type",
				},
				{
					Reason:  "reason1",
					Message: "message1",
					Source: apistructs.PipelineEventSource{
						Component: "component1",
						Host:      "host1",
					},
					FirstTimestamp: nowTime.Add(10 * time.Second),
					LastTimestamp:  nowTime.Add(40 * time.Second),
					Count:          1,
					Type:           "type1",
				},
			},
		},
		{
			name: "empty events add three no order lastTimestamp newEvents，three events count++",
			args: args{
				events: nil,
				newEvents: []*apistructs.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(-20 * time.Second),
						LastTimestamp:  nowTime.Add(20 * time.Second),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(20 * time.Second),
						LastTimestamp:  nowTime.Add(30 * time.Second),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(10 * time.Second),
						LastTimestamp:  nowTime.Add(40 * time.Second),
						Count:          1,
						Type:           "type",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: apistructs.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: nowTime.Add(-20 * time.Second),
					LastTimestamp:  nowTime.Add(40 * time.Second),
					Count:          3,
					Type:           "type",
				},
			},
		},
		{
			name: "events add in order",
			args: args{
				events: []*apistructs.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(-20 * time.Second),
						LastTimestamp:  nowTime.Add(40 * time.Second),
						Count:          3,
						Type:           "type",
					},
				},
				newEvents: []*apistructs.PipelineEvent{
					{
						Reason:  "reason1",
						Message: "message1",
						Source: apistructs.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: nowTime.Add(20 * time.Second),
						LastTimestamp:  nowTime.Add(40 * time.Second),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(-20 * time.Second),
						LastTimestamp:  nowTime.Add(50 * time.Second),
						Count:          1,
						Type:           "type",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: apistructs.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: nowTime.Add(-20 * time.Second),
					LastTimestamp:  nowTime.Add(40 * time.Second),
					Count:          3,
					Type:           "type",
				},
				{
					Reason:  "reason1",
					Message: "message1",
					Source: apistructs.PipelineEventSource{
						Component: "component1",
						Host:      "host1",
					},
					FirstTimestamp: nowTime.Add(20 * time.Second),
					LastTimestamp:  nowTime.Add(40 * time.Second),
					Count:          1,
					Type:           "type",
				},
				{
					Reason:  "reason",
					Message: "message",
					Source: apistructs.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: nowTime.Add(-20 * time.Second),
					LastTimestamp:  nowTime.Add(50 * time.Second),
					Count:          1,
					Type:           "type",
				},
			},
		},
		{
			name: "events first add counts",
			args: args{
				events: []*apistructs.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(-20 * time.Second),
						LastTimestamp:  nowTime.Add(40 * time.Second),
						Count:          3,
						Type:           "type",
					},
				},
				newEvents: []*apistructs.PipelineEvent{
					{
						Reason:  "reason1",
						Message: "message1",
						Source: apistructs.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: nowTime.Add(20 * time.Second),
						LastTimestamp:  nowTime.Add(60 * time.Second),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(-20 * time.Second),
						LastTimestamp:  nowTime.Add(50 * time.Second),
						Count:          1,
						Type:           "type",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: apistructs.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: nowTime.Add(-20 * time.Second),
					LastTimestamp:  nowTime.Add(50 * time.Second),
					Count:          4,
					Type:           "type",
				},
				{
					Reason:  "reason1",
					Message: "message1",
					Source: apistructs.PipelineEventSource{
						Component: "component1",
						Host:      "host1",
					},
					FirstTimestamp: nowTime.Add(20 * time.Second),
					LastTimestamp:  nowTime.Add(60 * time.Second),
					Count:          1,
					Type:           "type",
				},
			},
		},
		{
			name: "events first add counts three add counts",
			args: args{
				events: []*apistructs.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(-20 * time.Second),
						LastTimestamp:  nowTime.Add(40 * time.Second),
						Count:          3,
						Type:           "type",
					},
				},
				newEvents: []*apistructs.PipelineEvent{
					{
						Reason:  "reason1",
						Message: "message1",
						Source: apistructs.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: nowTime.Add(20 * time.Second),
						LastTimestamp:  nowTime.Add(60 * time.Second),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason1",
						Message: "message1",
						Source: apistructs.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: nowTime.Add(-30 * time.Second),
						LastTimestamp:  nowTime.Add(80 * time.Second),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason2",
						Message: "message2",
						Source: apistructs.PipelineEventSource{
							Component: "component2",
							Host:      "host2",
						},
						FirstTimestamp: nowTime.Add(20 * time.Second),
						LastTimestamp:  nowTime.Add(90 * time.Second),
						Count:          1,
						Type:           "type",
					},
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(-20 * time.Second),
						LastTimestamp:  nowTime.Add(50 * time.Second),
						Count:          1,
						Type:           "type",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason",
					Message: "message",
					Source: apistructs.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: nowTime.Add(-20 * time.Second),
					LastTimestamp:  nowTime.Add(50 * time.Second),
					Count:          4,
					Type:           "type",
				},
				{
					Reason:  "reason1",
					Message: "message1",
					Source: apistructs.PipelineEventSource{
						Component: "component1",
						Host:      "host1",
					},
					FirstTimestamp: nowTime.Add(-30 * time.Second),
					LastTimestamp:  nowTime.Add(80 * time.Second),
					Count:          2,
					Type:           "type",
				},
				{
					Reason:  "reason2",
					Message: "message2",
					Source: apistructs.PipelineEventSource{
						Component: "component2",
						Host:      "host2",
					},
					FirstTimestamp: nowTime.Add(20 * time.Second),
					LastTimestamp:  nowTime.Add(90 * time.Second),
					Count:          1,
					Type:           "type",
				},
			},
		},

		{
			name: "sort events and add events",
			args: args{
				events: []*apistructs.PipelineEvent{
					{
						Reason:  "reason",
						Message: "message",
						Source: apistructs.PipelineEventSource{
							Component: "component",
							Host:      "host",
						},
						FirstTimestamp: nowTime.Add(-20 * time.Second),
						LastTimestamp:  nowTime.Add(40 * time.Second),
						Count:          3,
						Type:           "type",
					},
					{
						Reason:  "reason1",
						Message: "message1",
						Source: apistructs.PipelineEventSource{
							Component: "component1",
							Host:      "host1",
						},
						FirstTimestamp: nowTime.Add(20 * time.Second),
						LastTimestamp:  nowTime.Add(30 * time.Second),
						Count:          1,
						Type:           "type",
					},
				},
				newEvents: []*apistructs.PipelineEvent{
					{
						Reason:  "reason2",
						Message: "message2",
						Source: apistructs.PipelineEventSource{
							Component: "component2",
							Host:      "host2",
						},
						FirstTimestamp: nowTime.Add(20 * time.Second),
						LastTimestamp:  nowTime.Add(90 * time.Second),
						Count:          1,
						Type:           "type",
					},
				},
			},
			want: orderedEvents{
				{
					Reason:  "reason1",
					Message: "message1",
					Source: apistructs.PipelineEventSource{
						Component: "component1",
						Host:      "host1",
					},
					FirstTimestamp: nowTime.Add(20 * time.Second),
					LastTimestamp:  nowTime.Add(30 * time.Second),
					Count:          1,
					Type:           "type",
				},
				{
					Reason:  "reason",
					Message: "message",
					Source: apistructs.PipelineEventSource{
						Component: "component",
						Host:      "host",
					},
					FirstTimestamp: nowTime.Add(-20 * time.Second),
					LastTimestamp:  nowTime.Add(40 * time.Second),
					Count:          3,
					Type:           "type",
				},
				{
					Reason:  "reason2",
					Message: "message2",
					Source: apistructs.PipelineEventSource{
						Component: "component2",
						Host:      "host2",
					},
					FirstTimestamp: nowTime.Add(20 * time.Second),
					LastTimestamp:  nowTime.Add(90 * time.Second),
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
