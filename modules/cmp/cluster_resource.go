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

package cmp

import (
	"context"
	"fmt"
	"time"

	jsi "github.com/json-iterator/go"
	types2 "github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/wrangler/pkg/data"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda-proto-go/cmp/dashboard/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/cache"
)

func (p *provider) GetClustersResources(ctx context.Context, cReq *pb.GetClustersResourcesRequest) (*pb.GetClusterResourcesResponse, error) {
	resp := &pb.GetClusterResourcesResponse{}
	for _, clusterName := range cReq.ClusterNames {
		detail := &pb.ClusterResourceDetail{ClusterName: clusterName}
		nodesList, err := p.ListSteveResource(ctx, &apistructs.SteveRequest{
			NoAuthentication: true,
			Type:             apistructs.K8SNode,
			ClusterName:      clusterName,
		})
		if err != nil {
			logrus.Errorf("failed to get cluster resource for cluster %s, %v", clusterName, err)
			detail.Err = err.Error()
			resp.List = append(resp.List, detail)
			continue
		}
		var nodes []data.Object
		for _, obj := range nodesList {
			nodes = append(nodes, obj.Data())
		}

		nodesAllocatedRes, err := GetNodesAllocatedRes(ctx, p, true, clusterName, "", "", nodes)
		if err != nil {
			logrus.Errorf("failed to get nodes allocated resource for cluster %s, %v", clusterName, err)
			detail.Err = err.Error()
			resp.List = append(resp.List, detail)
			continue
		}
		for _, node := range nodes {
			nodeName := node.String("metadata", "name")
			allocatableCPUQty, _ := resource.ParseQuantity(node.String("status", "allocatable", "cpu"))
			allocatableMemQty, _ := resource.ParseQuantity(node.String("status", "allocatable", "memory"))
			capacityCPUQty, _ := resource.ParseQuantity(node.String("status", "capacity", "cpu"))
			capacityMemQty, _ := resource.ParseQuantity(node.String("status", "capacity", "memory"))

			allocatableCPU := allocatableCPUQty.MilliValue()
			allocatableMem := allocatableMemQty.Value()
			capacityCPU := capacityCPUQty.MilliValue()
			capacityMem := capacityMemQty.Value()
			allocatedCPU := nodesAllocatedRes[nodeName].CPU
			allocatedMem := nodesAllocatedRes[nodeName].Mem

			labels := node.Map("metadata", "labels")
			var labelArr []string
			for k, v := range labels {
				labelArr = append(labelArr, fmt.Sprintf("%s=%s", k, v))
			}
			detail.Hosts = append(detail.Hosts, &pb.HostResourceDetail{
				Host:           nodeName,
				CpuAllocatable: uint64(allocatableCPU),
				CpuTotal:       uint64(capacityCPU),
				CpuRequest:     uint64(allocatedCPU),
				MemAllocatable: uint64(allocatableMem),
				MemTotal:       uint64(capacityMem),
				MemRequest:     uint64(allocatedMem),
				Labels:         labelArr,
			})
		}
		detail.Success = true
		resp.List = append(resp.List, detail)
	}
	resp.Total = uint32(len(resp.List))
	return resp, nil
}

func (p *provider) GetNamespacesResources(ctx context.Context, nReq *pb.GetNamespacesResourcesRequest) (*pb.GetNamespacesResourcesResponse, error) {
	resp := &pb.GetNamespacesResourcesResponse{}
	nss := make(map[string][]string)
	for _, req := range nReq.Namespaces {
		nss[req.ClusterName] = append(nss[req.ClusterName], req.Namespace)
	}

	for cluster, namespaces := range nss {
		item := &pb.ClusterResourceItem{ClusterName: cluster}
		nsAllocatableRes, err := GetNamespaceAllocatedRes(ctx, p, true, cluster, "", "", namespaces)
		if err != nil {
			logrus.Errorf("failed to get namespace allocated resource for cluster %s", cluster)
			item.Err = err.Error()
			resp.List = append(resp.List, item)
			continue
		}
		for _, namespace := range namespaces {
			item.List = append(item.List, &pb.NamespaceResourceDetail{
				Namespace:  namespace,
				CpuRequest: uint64(nsAllocatableRes[namespace].CPU),
				MemRequest: uint64(nsAllocatableRes[namespace].Mem),
			})
		}
		item.Success = true
		resp.List = append(resp.List, item)
	}
	resp.Total = uint32(len(resp.List))
	return resp, nil
}

type AllocatedRes struct {
	CPU    int64 `json:"cpu"`
	Mem    int64 `json:"mem"`
	PodNum int64 `json:"podNum"`
}

const (
	nodeCacheType = "nodeAllocatedRes"
	nsCacheType   = "nsAllocatedRes"
)

// GetNamespaceAllocatedRes get nodes allocated resource from cache, and update cache in goroutine
func GetNamespaceAllocatedRes(ctx context.Context, server SteveServer, noAuthentication bool, clusterName, userID, orgID string, namespaces []string) (map[string]AllocatedRes, error) {
	var pods []types2.APIObject
	hasExpired := false
	nsAllocatedRes := make(map[string]AllocatedRes)
	for _, namespace := range namespaces {
		value, expired, err := cache.GetFreeCache().Get(cache.GenerateKey(clusterName, namespace, nsCacheType))
		if err != nil {
			return nil, err
		}
		if expired {
			hasExpired = true
		}
		if value != nil {
			var nar AllocatedRes
			if err = jsi.Unmarshal(value[0].Value().([]byte), &nar); err != nil {
				logrus.Errorf("failed to unmarshal for namespace %s in GetNodeAllocatedRes, %v", namespace, err)
				continue
			}
			nsAllocatedRes[namespace] = nar
			continue
		}
		if pods == nil {
			req := &apistructs.SteveRequest{
				NoAuthentication: noAuthentication,
				UserID:           userID,
				OrgID:            orgID,
				Type:             apistructs.K8SPod,
				ClusterName:      clusterName,
			}
			pods, err = server.ListSteveResource(ctx, req)
			if err != nil {
				return nil, err
			}
		}
		cpu, mem, podNum := CalculateNamespaceAllocatedRes(namespace, pods)
		nar := AllocatedRes{
			CPU:    cpu,
			Mem:    mem,
			PodNum: podNum,
		}
		value, err = cache.MarshalValue(nar)
		if err != nil {
			logrus.Errorf("failed to marshal value for namespace %s in GetNamespaceAllocatedRes, %v", namespace, err)
			continue
		}
		if err = cache.GetFreeCache().Set(cache.GenerateKey(clusterName, namespace, nsCacheType), value, time.Second.Nanoseconds()*30); err != nil {
			logrus.Errorf("failed to set cache for namespace %s in GetNamespaceAllocatedRes, %v", namespace, err)
			continue
		}

		if err = jsi.Unmarshal(value[0].Value().([]byte), &nar); err != nil {
			logrus.Errorf("failed to unmarshal for namespace %s in GetNamespaceAllocatedRes, %v", namespace, err)
			continue
		}
		nsAllocatedRes[namespace] = nar
	}
	if hasExpired {
		go func() {
			var err error
			if pods == nil {
				req := &apistructs.SteveRequest{
					NoAuthentication: noAuthentication,
					UserID:           userID,
					OrgID:            orgID,
					Type:             apistructs.K8SPod,
					ClusterName:      clusterName,
				}

				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				pods, err = server.ListSteveResource(ctx, req)
				if err != nil {
					logrus.Errorf("failed to list pods in GetNamespaceAllocatedRes goroutine, %v", err)
					return
				}
			}
			for _, namespace := range namespaces {
				cpu, mem, podNum := CalculateNamespaceAllocatedRes(namespace, pods)
				nar := AllocatedRes{
					CPU:    cpu,
					Mem:    mem,
					PodNum: podNum,
				}
				value, err := cache.MarshalValue(nar)
				if err != nil {
					logrus.Errorf("failed to marshal value for namespace %s in GetNamespaceAllocatedRes goroutine, %v", namespace, err)
					continue
				}
				if err = cache.GetFreeCache().Set(cache.GenerateKey(clusterName, namespace, nsCacheType), value, time.Second.Nanoseconds()*30); err != nil {
					logrus.Errorf("failed to set cache for namespace %s in GetNamespaceAllocatedRes goroutine, %v", namespace, err)
					continue
				}
			}
			logrus.Infof("update namespace allocated resource cache succeeded")
		}()
	}
	return nsAllocatedRes, nil
}

// CalculateNamespaceAllocatedRes calculate allocated cpu, memory and pods for target namespaces
func CalculateNamespaceAllocatedRes(name string, pods []types2.APIObject) (cpu, mem, podNum int64) {
	cpuQty := resource.NewQuantity(0, resource.DecimalSI)
	memQty := resource.NewQuantity(0, resource.BinarySI)
	for _, obj := range pods {
		pod := obj.Data()
		if pod.String("metadata", "namespace") != name || pod.String("status", "phase") == "Failed" ||
			pod.String("status", "phase") == "Succeeded" {
			continue
		}
		podNum++
		containers := pod.Slice("spec", "containers")
		for _, container := range containers {
			requestsCPU := resource.NewQuantity(0, resource.DecimalSI)
			requestsMem := resource.NewQuantity(0, resource.BinarySI)
			requests := container.String("resources", "requests", "cpu")
			if requests != "" {
				*requestsCPU, _ = resource.ParseQuantity(requests)
			}
			requests = container.String("resources", "requests", "memory")
			if requests != "" {
				*requestsMem, _ = resource.ParseQuantity(requests)
			}
			cpuQty.Add(*requestsCPU)
			memQty.Add(*requestsMem)
		}
	}
	return cpuQty.MilliValue(), memQty.Value(), podNum
}

// GetNodesAllocatedRes get nodes allocated resource from cache, and update cache in goroutine
func GetNodesAllocatedRes(ctx context.Context, server SteveServer, noAuthentication bool, clusterName, userID, orgID string, nodes []data.Object) (map[string]AllocatedRes, error) {
	var pods []types2.APIObject
	hasExpired := false
	nodesAllocatedRes := make(map[string]AllocatedRes)
	for _, node := range nodes {
		nodeName := node.String("metadata", "name")
		value, expired, err := cache.GetFreeCache().Get(cache.GenerateKey(clusterName, nodeName, nodeCacheType))
		if err != nil {
			return nil, err
		}
		if expired {
			hasExpired = true
		}
		if value != nil {
			var nar AllocatedRes
			if err = jsi.Unmarshal(value[0].Value().([]byte), &nar); err != nil {
				logrus.Errorf("failed to unmarshal for node %s in GetNodeAllocatedRes, %v", nodeName, err)
				continue
			}
			nodesAllocatedRes[nodeName] = nar
			continue
		}
		if pods == nil {
			req := &apistructs.SteveRequest{
				NoAuthentication: noAuthentication,
				UserID:           userID,
				OrgID:            orgID,
				Type:             apistructs.K8SPod,
				ClusterName:      clusterName,
			}
			pods, err = server.ListSteveResource(ctx, req)
			if err != nil {
				return nil, err
			}
		}
		cpu, mem, podNum := CalculateNodeAllocatedRes(nodeName, pods)
		nar := AllocatedRes{
			CPU:    cpu,
			Mem:    mem,
			PodNum: podNum,
		}
		value, err = cache.MarshalValue(nar)
		if err != nil {
			logrus.Errorf("failed to marshal value for node %s in GetNodeAllocatedRes, %v", nodeName, err)
			continue
		}
		if err = cache.GetFreeCache().Set(cache.GenerateKey(clusterName, nodeName, nodeCacheType), value, time.Second.Nanoseconds()*30); err != nil {
			logrus.Errorf("failed to set cache for node %s in GetNodeAllocatedRes, %v", nodeName, err)
			continue
		}

		if err = jsi.Unmarshal(value[0].Value().([]byte), &nar); err != nil {
			logrus.Errorf("failed to unmarshal for node %s in GetNodeAllocatedRes, %v", nodeName, err)
			continue
		}
		nodesAllocatedRes[nodeName] = nar
	}
	if hasExpired {
		go func() {
			var err error
			if pods == nil {
				req := &apistructs.SteveRequest{
					NoAuthentication: noAuthentication,
					UserID:           userID,
					OrgID:            orgID,
					Type:             apistructs.K8SPod,
					ClusterName:      clusterName,
				}

				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				pods, err = server.ListSteveResource(ctx, req)
				if err != nil {
					logrus.Errorf("failed to list pods in GetNodeAllocatedRes goroutine, %v", err)
					return
				}
			}
			for _, node := range nodes {
				nodeName := node.String("metadata", "name")
				cpu, mem, podNum := CalculateNodeAllocatedRes(nodeName, pods)
				nar := AllocatedRes{
					CPU:    cpu,
					Mem:    mem,
					PodNum: podNum,
				}
				value, err := cache.MarshalValue(nar)
				if err != nil {
					logrus.Errorf("failed to marshal value for node %s in GetNodeAllocatedRes goroutine, %v", nodeName, err)
					continue
				}
				if err = cache.GetFreeCache().Set(cache.GenerateKey(clusterName, nodeName, nodeCacheType), value, time.Second.Nanoseconds()*30); err != nil {
					logrus.Errorf("failed to set cache for node %s in GetNodeAllocatedRes goroutine, %v", nodeName, err)
					continue
				}
			}
			logrus.Infof("update node allocated resource cache succeeded")
		}()
	}
	return nodesAllocatedRes, nil
}

// CalculateNodeAllocatedRes calculate allocated cpu, memory and pods for target node
func CalculateNodeAllocatedRes(nodeName string, pods []types2.APIObject) (cpu, mem, podNum int64) {
	cpuQty := resource.NewQuantity(0, resource.DecimalSI)
	memQty := resource.NewQuantity(0, resource.BinarySI)
	for _, obj := range pods {
		pod := obj.Data()
		if pod.String("spec", "nodeName") != nodeName || pod.String("status", "phase") == "Failed" ||
			pod.String("status", "phase") == "Succeeded" {
			continue
		}
		podNum++
		containers := pod.Slice("spec", "containers")
		for _, container := range containers {
			requestsCPU := resource.NewQuantity(0, resource.DecimalSI)
			requestsMem := resource.NewQuantity(0, resource.BinarySI)
			requests := container.String("resources", "requests", "cpu")
			if requests != "" {
				*requestsCPU, _ = resource.ParseQuantity(requests)
			}
			requests = container.String("resources", "requests", "memory")
			if requests != "" {
				*requestsMem, _ = resource.ParseQuantity(requests)
			}
			cpuQty.Add(*requestsCPU)
			memQty.Add(*requestsMem)
		}
	}
	return cpuQty.MilliValue(), memQty.Value(), podNum
}
