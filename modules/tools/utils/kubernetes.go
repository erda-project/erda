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

package utils

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/recallsong/go-utils/lang/size"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	// Resource .
	Resource struct {
		Name       string
		MemRequest int64
		MemLimit   int64
		CPURequest int64
		CPULimit   int64
	}
	// NodeInfo .
	NodeInfo struct {
		HostIP string
		Resource
		MemAllocatable int64
		MemCapacity    int64
		CPUAllocatable int64
		CPUCapacity    int64
		HasEnv         Containers
		NotHasEnv      Containers
	}
	// Containers .
	Containers []*Resource
)

func (cs Containers) String() string {
	var names []string
	for _, item := range cs {
		names = append(names, item.Name)
	}
	return strings.Join(names, ",")
}

// Start .
func (p *provider) showK8sNodeResources() error {
	clientset := p.k8s
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	hosts := map[string]*NodeInfo{}
	for _, node := range nodes.Items {
		ip := node.Name
		for _, addr := range node.Status.Addresses {
			if addr.Type == v1.NodeInternalIP {
				ip = addr.Address
				break
			}
		}
		hosts[node.Name] = &NodeInfo{
			HostIP:         ip,
			MemAllocatable: convertQuantity(node.Status.Allocatable.Memory().String(), 1),
			MemCapacity:    convertQuantity(node.Status.Capacity.Memory().String(), 1),
			CPUAllocatable: convertQuantity(node.Status.Allocatable.Cpu().String(), 1000),
			CPUCapacity:    convertQuantity(node.Status.Capacity.Cpu().String(), 1000),
		}
	}
	pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	var notExistEnv, existEnv int
	for _, pod := range pods.Items {
		spec := pod.Spec
		node, ok := hosts[spec.NodeName]
		if !ok {
			fmt.Printf("[WARN] not found node %s for pod %s in namespace %s\n", spec.NodeName, pod.Name, pod.Namespace)
			continue
		}
		var memReq, memLimit, cpuReq, cpuLimit int64
		for _, c := range spec.Containers {
			res := c.Resources
			memReq = convertQuantity(res.Requests.Memory().String(), 1)
			memLimit = convertQuantity(res.Limits.Memory().String(), 1)
			cpuReq = convertQuantity(res.Requests.Cpu().String(), 1000)
			cpuLimit = convertQuantity(res.Limits.Cpu().String(), 1000)
			node.MemRequest += memReq
			node.MemLimit += memLimit
			node.CPURequest += cpuReq
			node.CPULimit += cpuLimit
			var find bool
			for _, item := range c.Env {
				if strings.Contains(item.Name, "DICE_MEM_") {
					find = true
					break
				}
			}
			if !find && (memReq > 0 || cpuReq > 0) {
				fmt.Println("N", spec.NodeName, pod.Namespace+"/"+pod.Name,
					"mem_req", size.FormatBytes(memReq),
					"mem_limit", size.FormatBytes(memLimit),
					"cpu_req", cpuReq,
					"cpu_limit", cpuLimit,
				)
				notExistEnv++
				node.NotHasEnv = append(node.NotHasEnv, &Resource{
					Name:       fmt.Sprintf("%s/%s", pod.Name, pod.Namespace),
					MemRequest: memReq,
					MemLimit:   memLimit,
					CPURequest: cpuReq,
					CPULimit:   cpuLimit,
				})
			} else {
				existEnv++
				node.HasEnv = append(node.HasEnv, &Resource{
					Name:       fmt.Sprintf("%s/%s", pod.Name, pod.Namespace),
					MemRequest: memReq,
					MemLimit:   memLimit,
					CPURequest: cpuReq,
					CPULimit:   cpuLimit,
				})
			}
		}
		_ = node
	}
	fmt.Println("exist DICE_MEM_REQUEST OR DICE_CPU_REQUEST containers:", existEnv)
	fmt.Println("not exist DICE_MEM_REQUEST OR DICE_CPU_REQUEST containers:", notExistEnv)
	fmt.Println()
	fmt.Println("ip", "mem_request", "mem_limit", "mem_allocatable", "mem_capacity", "cpu_request", "cpu_limit", "cpu_allocatable", "cpu_capacity")
	for _, host := range hosts {
		var res Resource
		for _, item := range host.NotHasEnv {
			res.MemRequest += item.MemRequest
			res.MemLimit += item.MemLimit
			res.CPURequest += item.CPURequest
			res.CPULimit += item.CPULimit
		}
		fmt.Printf("%s, "+
			"mem_req:%s (miss %s), mem_limit:%s (miss %s), mem_alloc:%s, mem_cap:%s, "+
			"cpu_req:%dm (miss %dm), cpu_limit:%dm (miss %dm), cpu_alloc:%dm, cpu_cap:%dm\n",
			host.HostIP,
			size.FormatBytes(host.MemRequest), size.FormatBytes(res.MemRequest),
			size.FormatBytes(host.MemLimit), size.FormatBytes(res.MemLimit),
			size.FormatBytes(host.MemAllocatable), size.FormatBytes(res.MemRequest),
			host.CPURequest, res.CPURequest,
			host.CPULimit, res.CPULimit,
			host.CPUAllocatable, host.CPUCapacity,
		)
	}
	return nil
}

func convertQuantity(s string, m float64) int64 {
	q, err := resource.ParseQuantity(s)
	if err != nil {
		fmt.Printf("E! Failed to parse quantity - %v\n", err)
		return 0
	}
	f, err := strconv.ParseFloat(fmt.Sprint(q.AsDec()), 64)
	if err != nil {
		fmt.Printf("E! Failed to parse float - %v", err)
		return 0
	}
	if m < 1 {
		m = 1
	}
	return int64(f * m)
}
