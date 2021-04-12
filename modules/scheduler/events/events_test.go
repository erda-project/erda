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

package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeServiceStatus(t *testing.T) {
	ev := RuntimeEvent{
		RuntimeName: "whatever",
		ServiceStatuses: []ServiceStatus{
			{
				ServiceName: "x1",
				Replica:     0,
			},
			{
				ServiceName: "x2",
				Replica:     1,
				InstanceStatuses: []InstanceStatus{
					{
						ID:             "XX12",
						Ip:             "10.11.12.13",
						InstanceStatus: HEALTHY,
					},
				},
			},
			{
				ServiceName: "x3",
				Replica:     2,
				InstanceStatuses: []InstanceStatus{
					{
						ID:             "XX23",
						Ip:             "11.12.13.14",
						InstanceStatus: RUNNING,
					},
					{
						ID:             "XX24",
						Ip:             "12.13.14.15",
						InstanceStatus: HEALTHY,
					},
				},
			},
			{
				ServiceName: "x4",
				Replica:     0,
				InstanceStatuses: []InstanceStatus{
					{
						ID:             "y1",
						Ip:             "110.12.0.14",
						InstanceStatus: INSTANCE_RUNNING,
					},
					{
						ID:             "y2",
						Ip:             "110.13.0.15",
						InstanceStatus: INSTANCE_RUNNING,
					},
				},
			},
		},
	}

	computeServiceStatus(&ev)
	assert.Equal(t, "Healthy", ev.ServiceStatuses[0].ServiceStatus)
	assert.Equal(t, "Healthy", ev.ServiceStatuses[1].ServiceStatus)
	assert.Equal(t, "UnHealthy", ev.ServiceStatuses[2].ServiceStatus)
	assert.Equal(t, "UnHealthy", ev.ServiceStatuses[3].ServiceStatus)
}
