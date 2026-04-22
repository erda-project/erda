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

package cmd

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var RUNTIMEINSTANCE = command.Command{
	ParentName: "RUNTIME",
	Name:       "instance",
	ShortHelp:  "Runtime instance operations",
	Example:    `erda-cli runtime instance list`,
}

var (
	listRuntimeServicePodsForInstances              = common.ListRuntimeServicePods
	listStoppedRuntimeServiceContainersForInstances = common.ListRuntimeServiceStoppedContainers
)

func RuntimeInstanceList(ctx *command.Context, workspace string, runtimeID uint64, service string, all bool) error {
	resolved, err := resolveRuntimeContext(ctx, workspace, runtimeID, true)
	if err != nil {
		return err
	}

	services := []string{service}
	if service == "" {
		runtime, err := inspectRuntime(ctx, resolved.orgID, resolved.runtimeID, resolved.applicationID, resolved.workspace)
		if err != nil {
			return err
		}
		services = services[:0]
		for name := range runtime.Services {
			services = append(services, name)
		}
		sort.Strings(services)
	}

	var containers apistructs.Containers
	for _, serviceName := range services {
		servicePods, err := listRuntimeServicePodsForInstances(ctx, resolved.orgID, int64(resolved.runtimeID), serviceName)
		if err != nil {
			return err
		}
		containers = append(containers, runtimePodsToContainers(servicePods)...)
		if !all {
			continue
		}
		serviceContainers, err := listStoppedRuntimeServiceContainersForInstances(ctx, resolved.orgID, int64(resolved.runtimeID), serviceName)
		if err != nil {
			return err
		}
		containers = append(containers, serviceContainers...)
	}

	return writeRuntimeInstanceList(runtimeStdout, resolved.runtimeID, containers)
}

func RuntimeInstanceLogs(ctx *command.Context, instanceArg string, workspace string, runtimeID uint64, service string, instance string, tail int, stream string, watch bool) error {
	instance = resolveRuntimeInstanceLogTarget(instanceArg, instance)
	if instance == "" {
		return fmt.Errorf("instance is required, use `erda-cli runtime instance logs <instance>` or `--instance`")
	}
	return RuntimeLogs(ctx, workspace, runtimeID, service, instance, tail, stream, watch)
}

func writeRuntimeInstanceList(w io.Writer, runtimeID uint64, containers apistructs.Containers) error {
	fmt.Fprintf(w, "instances (runtimeID=%d, total=%d)\n", runtimeID, len(containers))

	sort.Slice(containers, func(i, j int) bool {
		iTime := runtimeInstanceSortTime(containers[i])
		jTime := runtimeInstanceSortTime(containers[j])
		if !iTime.Equal(jTime) {
			return iTime.After(jTime)
		}
		if containers[i].Service != containers[j].Service {
			return containers[i].Service < containers[j].Service
		}
		return runtimeInstanceName(containers[i]) < runtimeInstanceName(containers[j])
	})

	rows := make([][]string, 0, len(containers))
	for _, container := range containers {
		rows = append(rows, []string{
			container.Service,
			runtimeInstanceName(container),
			container.Status,
			container.IPAddress,
			container.StartedAt,
			runtimeInstanceFinishedAt(container),
		})
	}

	return table.NewTable(table.WithWriter(w)).Header([]string{
		"service", "instance", "status", "ip", "startedAt", "finishedAt",
	}).Data(rows).Flush()
}

func runtimeInstanceName(container apistructs.Container) string {
	if container.PodName != "" {
		return container.PodName
	}
	if container.ContainerID != "" {
		return container.ContainerID
	}
	if container.ID != "" {
		return container.ID
	}
	return "-"
}

func runtimePodsToContainers(pods apistructs.Pods) apistructs.Containers {
	containers := make(apistructs.Containers, 0, len(pods))
	for _, pod := range pods {
		containers = append(containers, apistructs.Container{
			K8sInstanceMetaInfo: apistructs.K8sInstanceMetaInfo{
				PodName:      pod.PodName,
				PodNamespace: pod.K8sNamespace,
			},
			IPAddress:   pod.IPAddress,
			Host:        pod.Host,
			Status:      pod.Phase,
			Message:     pod.Message,
			StartedAt:   pod.StartedAt,
			UpdatedAt:   pod.UpdatedAt,
			Service:     pod.Service,
			ClusterName: pod.ClusterName,
		})
	}
	return containers
}

func runtimeInstanceSortTime(container apistructs.Container) time.Time {
	if ts, ok := parseRuntimeInstanceTime(container.StartedAt); ok {
		return ts
	}
	return time.Time{}
}

func runtimeInstanceFinishedAt(container apistructs.Container) string {
	if strings.TrimSpace(container.FinishedAt) == "" {
		return "-"
	}
	return container.FinishedAt
}

func resolveRuntimeInstanceLogTarget(instanceArg, instanceFlag string) string {
	if instanceArg != "" {
		return instanceArg
	}
	return instanceFlag
}

func parseRuntimeInstanceTime(value string) (time.Time, bool) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, false
	}
	ts, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, false
	}
	return ts, true
}
