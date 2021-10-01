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

package formatter

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	jsi "github.com/json-iterator/go"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/erda-project/erda/modules/cmp/cache"
	"github.com/erda-project/erda/modules/cmp/queue"
)

var queryQueue *queue.QueryQueue

func init() {
	queueSize := 10
	if size, err := strconv.Atoi(os.Getenv("NODE_QUEUE_SIZE")); err == nil && size > queueSize {
		queueSize = size
	}
	queryQueue = queue.NewQueryQueue(queueSize)
}

type NodeFormatter struct {
	ctx       context.Context
	podClient corev1.PodInterface
	podsCache *cache.Cache
}

type res struct {
	CPU       int64
	CPUStr    string
	Memory    int64
	MemoryStr string
	Pods      int64
	PodsStr   string
}

type cacheKey struct {
	nodeName string
}

func (c *cacheKey) getKey() string {
	return fmt.Sprintf("nodeAllocatedResCache-%s", c.nodeName)
}

func NewNodeFormatter(ctx context.Context, k8sInterface kubernetes.Interface) *NodeFormatter {
	return &NodeFormatter{
		ctx:       ctx,
		podClient: k8sInterface.CoreV1().Pods(""),
		podsCache: cache.FreeCache,
	}
}

func (n *NodeFormatter) Formatter(request *types.APIRequest, resource *types.RawResource) {
	allocatableRes := parseRes(resource, "allocatable")
	capacityRes := parseRes(resource, "capacity")
	unallocatableRes := map[string]interface{}{
		"CPU":    capacityRes.CPU - allocatableRes.CPU,
		"Memory": capacityRes.Memory - allocatableRes.Memory,
		"Pods":   capacityRes.Pods - capacityRes.Pods,
	}
	parsedRes := map[string]interface{}{
		"unallocatable": unallocatableRes,
		"capacity": map[string]interface{}{
			"CPU":       capacityRes.CPU,
			"Memory":    capacityRes.Memory,
			"Pods":      capacityRes.Pods,
			"CPUStr":    capacityRes.CPUStr,
			"MemoryStr": capacityRes.MemoryStr,
		},
	}

	nodeName := resource.ID
	key := &cacheKey{nodeName}
	data := resource.APIObject.Data()
	value, expired, err := n.podsCache.Get(key.getKey())
	if value == nil || err != nil {
		allocatedRes, err := n.getNodeAllocatedRes(request.Context(), nodeName)
		if err != nil {
			logrus.Errorf("failed to get allocated resource for node %s, %v", nodeName, err)
			return
		}
		val, _ := cache.MarshalValue(allocatedRes)
		err = n.podsCache.Set(key.getKey(), val, time.Minute.Nanoseconds())
		if err != nil {
			logrus.Errorf("failed to update cache, key:%s", key.getKey())
		}

		parsedRes["allocated"] = allocatedRes
		data.SetNested(parsedRes, "extra", "parsedResource")
		return
	}

	if expired {
		logrus.Infof("pods data expired, need update, key:%s", key.getKey())
		if !cache.ExpireFreshQueue.IsFull() {
			task := &queue.Task{
				Key: key.getKey(),
				Do: func() {
					allocatedRes, err := n.getNodeAllocatedRes(n.ctx, nodeName)
					if err != nil {
						logrus.Errorf("failed to get allocated resource for node %s, %v", nodeName, err)
						return
					}
					val, _ := cache.MarshalValue(allocatedRes)
					err = n.podsCache.Set(key.getKey(), val, 5*time.Minute.Nanoseconds())
					if err != nil {
						logrus.Errorf("failed to update cache, key:%s", key.getKey())
					}
				},
			}
			cache.ExpireFreshQueue.Enqueue(task)
		} else {
			logrus.Warnf("queue size is full, task is ignored, key:%s", key.getKey())
		}
	}
	allocatedRes := map[string]interface{}{}
	if err = jsi.Unmarshal(value[0].Value().([]byte), &allocatedRes); err != nil {
		logrus.Errorf("failed to unmarshal allocatedResource, %v", err)
	}
	parsedRes["allocated"] = allocatedRes
	data.SetNested(parsedRes, "extra", "parsedResource")
}

func (n *NodeFormatter) getNodeAllocatedRes(ctx context.Context, nodeName string) (map[string]interface{}, error) {
	fieldSelector := fmt.Sprintf("spec.nodeName=%s,status.phase!=Failed,status.phase!=Succeeded", nodeName)
	clusterName := n.ctx.Value("clusterName").(string)
	logrus.Infof("[DEBUG] start list pods")
	queryQueue.Acquire(clusterName, 1)
	pods, err := n.podClient.List(ctx, v1.ListOptions{
		FieldSelector: fieldSelector,
	})
	queryQueue.Release(clusterName, 1)
	logrus.Infof("[DEBUG] end list pods")
	if err != nil {
		return nil, err
	}

	cpu := resource.NewQuantity(0, resource.DecimalSI)
	mem := resource.NewQuantity(0, resource.BinarySI)
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			requestedCPU := container.Resources.Requests.Cpu()
			if requestedCPU != nil {
				cpu.Add(*requestedCPU)
			}
			requestedMem := container.Resources.Requests.Memory()
			if requestedMem != nil {
				mem.Add(*requestedMem)
			}
		}
	}
	return map[string]interface{}{
		"CPU":       cpu.MilliValue(),
		"CPUStr":    cpu.String(),
		"Memory":    mem.Value(),
		"MemoryStr": mem.String(),
		"Pods":      int64(len(pods.Items)),
	}, nil
}

func parseRes(raw *types.RawResource, resType string) *res {
	cpu := raw.APIObject.Data().String("status", resType, "cpu")
	mem := raw.APIObject.Data().String("status", resType, "memory")
	pods := raw.APIObject.Data().String("status", resType, "pods")

	parsedCPU, _ := resource.ParseQuantity(cpu)
	parsedMem, _ := resource.ParseQuantity(mem)
	parsedPods, _ := resource.ParseQuantity(pods)

	return &res{
		CPU:       parsedCPU.MilliValue(),
		Memory:    parsedMem.Value(),
		Pods:      parsedPods.Value(),
		CPUStr:    parsedCPU.String(),
		MemoryStr: parsedMem.String(),
	}
}
