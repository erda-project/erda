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
