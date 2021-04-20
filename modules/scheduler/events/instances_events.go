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
	"sync"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/modules/scheduler/events/eventtypes"

	"github.com/pkg/errors"
)

func convertInstanceStatus(originEventStatus string) string {
	switch originEventStatus {
	case KILLED:
		return INSTANCE_KILLED
	case RUNNING:
		return INSTANCE_RUNNING
	case FAILED:
		return INSTANCE_FAILED
	case FINISHED:
		return INSTANCE_FINISHED
	}
	return originEventStatus
}

func executorName2ClusterName(executorName string) string {
	if v, ok := conf.GetConfStore().ExecutorStore.Load(executorName); ok {
		config := v.(*conf.ExecutorConfig)
		return config.ClusterName
	}
	return ""
}

func handleInstanceStatusChangedEvents(e *eventtypes.StatusEvent, lstore *sync.Map) error {
	evm := GetEventManager()
	layer, err := getLayerInfoFromEvent(e.TaskId, e.Type)
	if err != nil {
		return err
	}

	ie := apistructs.InstanceStatusData{
		ClusterName:    executorName2ClusterName(e.Cluster),
		RuntimeName:    layer.RuntimeName,
		ServiceName:    layer.ServiceName,
		ID:             e.TaskId,
		IP:             e.IP,
		InstanceStatus: convertInstanceStatus(e.Status),
		Host:           e.Host,
		Message:        e.Message,
		Timestamp:      time.Now().UnixNano(),
	}

	v, ok := lstore.Load(e.TaskId)
	if !ok {
		lstore.Store(e.TaskId, ie)
	} else {
		instance, ok := v.(apistructs.InstanceStatusData)
		if !ok {
			return errors.Errorf("instance(id:%s) wrong format in lstore", e.TaskId)
		}

		//Ignore the running event after healthy or unhealthy
		if (instance.InstanceStatus == HEALTHY || instance.InstanceStatus == UNHEALTHY) && e.Status == RUNNING {
			ie.InstanceStatus = instance.InstanceStatus
			return nil
		}

		// There is no ip information in marathon's health check event
		if e.Status == HEALTHY || e.Status == UNHEALTHY {
			ie.IP = instance.IP
			ie.Host = instance.Host
		}

		// marathon bug: For an instance, after sending running and healthy, it will repeat running and healthy
		// Filter recurring events
		if convertInstanceStatus(e.Status) == instance.InstanceStatus {
			return nil
		}

		// If the instance is deleted, the corresponding record will be deleted from the local cache. Otherwise update the record.
		if ie.InstanceStatus == INSTANCE_KILLED || ie.InstanceStatus == INSTANCE_FAILED || ie.InstanceStatus == INSTANCE_FINISHED {
			lstore.Delete(e.TaskId)
			// marathon bug: The container ip passed in the killed event is host ip in most cases
			ie.IP = instance.IP
		} else {
			lstore.Store(e.TaskId, ie)
		}
	}

	return evm.notifier.Send(ie, WithDest(map[string]interface{}{"WEBHOOK": apistructs.EventHeader{
		Event:     INSTANCE_STATUS,
		Action:    "changed",
		OrgID:     "-1",
		ProjectID: "-1",
	}}))
}
