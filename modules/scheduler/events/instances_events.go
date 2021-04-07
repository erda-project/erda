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

		// 忽略 healthy 或者 unhealthy 后的 running 事件
		if (instance.InstanceStatus == HEALTHY || instance.InstanceStatus == UNHEALTHY) && e.Status == RUNNING {
			ie.InstanceStatus = instance.InstanceStatus
			return nil
		}

		// marathon 的健康检查事件没有 ip 信息
		if e.Status == HEALTHY || e.Status == UNHEALTHY {
			ie.IP = instance.IP
			ie.Host = instance.Host
		}

		// marathon bug: 对于一个实例发完 running 和 healthy 后会重复发 running 和 healthy
		// 过滤重复事件
		if convertInstanceStatus(e.Status) == instance.InstanceStatus {
			return nil
		}

		// 实例如果被删除，则从本地缓存中删除对应记录。否则更新记录。
		if ie.InstanceStatus == INSTANCE_KILLED || ie.InstanceStatus == INSTANCE_FAILED || ie.InstanceStatus == INSTANCE_FINISHED {
			lstore.Delete(e.TaskId)
			// marathon bug: killed 事件中传过来的容器 ip 绝大多数情况下是 host ip
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
