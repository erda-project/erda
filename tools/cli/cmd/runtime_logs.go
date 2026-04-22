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
	"sort"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var RUNTIMELOGS = command.Command{
	ParentName: "RUNTIME",
	Name:       "logs",
	ShortHelp:  "View runtime logs",
	Example: ` $ erda-cli runtime logs
  $ erda-cli runtime logs --service web
  $ erda-cli runtime logs --service web --instance pod-0 --watch`,
	Flags: []command.Flag{
		command.StringFlag{Short: "", Name: "workspace", Doc: "workspace to query", DefaultValue: ""},
		command.Uint64Flag{Short: "i", Name: "runtime-id", Doc: "show a specific runtime id", DefaultValue: 0},
		command.StringFlag{Short: "s", Name: "service", Doc: "filter by service name", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "instance", Doc: "filter by instance name", DefaultValue: ""},
		command.IntFlag{Short: "", Name: "tail", Doc: "number of recent log lines to fetch first", DefaultValue: 200},
		command.StringFlag{Short: "", Name: "stream", Doc: "log stream: stdout, stderr, or all", DefaultValue: ""},
		command.BoolFlag{Short: "w", Name: "watch", Doc: "watch for new log lines", DefaultValue: false},
	},
	Run: RuntimeLogs,
}

type runtimeLogEntry struct {
	service  string
	instance string
	line     apistructs.DashboardSpotLogLine
}

var (
	listRuntimeServicePodsForLogs              = common.ListRuntimeServicePods
	listStoppedRuntimeServiceContainersForLogs = common.ListRuntimeServiceStoppedContainers
	getRuntimePodLogs                          = common.GetRuntimePodLogs
	getRuntimeStoppedContainerLogs             = common.GetRuntimeStoppedContainerLogs
	runtimeLogsSleep                           = time.Sleep
	runtimeLogsShouldStop                      = func() bool { return false }
)

func RuntimeLogs(ctx *command.Context, workspace string, runtimeID uint64, service string, instance string, tail int, stream string, watch bool) error {
	resolved, err := resolveRuntimeContext(ctx, workspace, runtimeID, true)
	if err != nil {
		return err
	}

	normalizedStream, err := normalizeLogStream(stream)
	if err != nil {
		return err
	}

	if watch {
		return watchRuntimeLogs(ctx, resolved, service, instance, tail, normalizedStream)
	}

	entries, sourceCount, err := fetchRuntimeLogEntries(ctx, resolved, service, instance, tail, normalizedStream)
	if err != nil {
		return err
	}
	if sourceCount == 0 {
		return fmt.Errorf("no runtime log source matched for runtime %d", resolved.runtimeID)
	}
	return writeRuntimeLogEntries(runtimeStdout, entries, sourceCount > 1)
}

func watchRuntimeLogs(ctx *command.Context, resolved *resolvedRuntimeContext, service string, instance string, tail int, stream string) error {
	seen := make(map[string]struct{})
	for {
		entries, sourceCount, err := fetchRuntimeLogEntries(ctx, resolved, service, instance, tail, stream)
		if err != nil {
			return err
		}
		if sourceCount == 0 {
			return fmt.Errorf("no runtime log source matched for runtime %d", resolved.runtimeID)
		}

		unseen := filterUnseenRuntimeLogEntries(entries, seen)
		if err := writeRuntimeLogEntries(runtimeStdout, unseen, sourceCount > 1); err != nil {
			return err
		}
		if runtimeLogsShouldStop() {
			return nil
		}
		runtimeLogsSleep(2 * time.Second)
		if runtimeLogsShouldStop() {
			return nil
		}
	}
}

func fetchRuntimeLogEntries(ctx *command.Context, resolved *resolvedRuntimeContext, service string, instance string, tail int, normalizedStream string) ([]runtimeLogEntry, int, error) {
	services := []string{service}
	if service == "" {
		runtime, err := inspectRuntime(ctx, resolved.orgID, resolved.runtimeID, resolved.applicationID, resolved.workspace)
		if err != nil {
			return nil, 0, err
		}
		services = services[:0]
		for name := range runtime.Services {
			services = append(services, name)
		}
		sort.Strings(services)
	}

	var entries []runtimeLogEntry
	sourceCount := 0
	for _, serviceName := range services {
		matchedRunningSource := false
		pods, err := listRuntimeServicePodsForLogs(ctx, resolved.orgID, int64(resolved.runtimeID), serviceName)
		if err != nil {
			return nil, 0, err
		}
		for _, pod := range pods {
			if instance != "" && !matchRuntimePodInstance(pod, instance) {
				continue
			}
			containerName, containerID := runtimeLogTargetContainer(pod, serviceName)
			if containerID == "" {
				continue
			}
			matchedRunningSource = true
			sourceCount++
			logData, err := getRuntimePodLogs(ctx, resolved.orgName, resolved.applicationID, pod, containerName, containerID, normalizedStream, tail)
			if err != nil {
				return nil, 0, err
			}
			for _, line := range logData.Lines {
				entries = append(entries, runtimeLogEntry{
					service:  pod.Service,
					instance: pod.PodName,
					line:     line,
				})
			}
		}
		if instance == "" || matchedRunningSource {
			continue
		}
		containers, err := listStoppedRuntimeServiceContainersForLogs(ctx, resolved.orgID, int64(resolved.runtimeID), serviceName)
		if err != nil {
			return nil, 0, err
		}
		for _, container := range containers {
			if !matchRuntimeContainerInstance(container, instance) || container.ContainerID == "" {
				continue
			}
			sourceCount++
			logData, err := getRuntimeStoppedContainerLogs(ctx, resolved.orgName, resolved.applicationID, container, normalizedStream, tail)
			if err != nil {
				return nil, 0, err
			}
			for _, line := range logData.Lines {
				entries = append(entries, runtimeLogEntry{
					service:  container.Service,
					instance: runtimeInstanceName(container),
					line:     line,
				})
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].line.TimeStamp == entries[j].line.TimeStamp {
			if entries[i].service == entries[j].service {
				return entries[i].instance < entries[j].instance
			}
			return entries[i].service < entries[j].service
		}
		return entries[i].line.TimeStamp < entries[j].line.TimeStamp
	})
	return entries, sourceCount, nil
}

func writeRuntimeLogEntries(w interface{ Write([]byte) (int, error) }, entries []runtimeLogEntry, multipleSources bool) error {
	for _, entry := range entries {
		if multipleSources {
			if _, err := fmt.Fprintf(w, "[%s/%s] %s\n", entry.service, entry.instance, entry.line.Content); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintln(w, entry.line.Content); err != nil {
			return err
		}
	}
	return nil
}

func matchRuntimePodInstance(pod apistructs.Pod, instance string) bool {
	if pod.PodName == instance || pod.Uid == instance {
		return true
	}
	for _, c := range pod.PodContainers {
		if c.ContainerID == instance || c.ContainerName == instance {
			return true
		}
	}
	return false
}

func matchRuntimeContainerInstance(container apistructs.Container, instance string) bool {
	return container.PodName == instance || container.ID == instance || container.ContainerID == instance
}

func runtimeLogTargetContainer(pod apistructs.Pod, serviceName string) (string, string) {
	for _, c := range pod.PodContainers {
		if c.ContainerName == serviceName && c.ContainerID != "" {
			return c.ContainerName, c.ContainerID
		}
	}
	for _, c := range pod.PodContainers {
		if c.ContainerID != "" {
			return c.ContainerName, c.ContainerID
		}
	}
	return "", ""
}

func filterUnseenRuntimeLogEntries(entries []runtimeLogEntry, seen map[string]struct{}) []runtimeLogEntry {
	filtered := make([]runtimeLogEntry, 0, len(entries))
	for _, entry := range entries {
		key := fmt.Sprintf("%s|%s|%s|%d|%s", entry.service, entry.instance, entry.line.TimeStamp, entry.line.Offset, entry.line.Content)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		filtered = append(filtered, entry)
	}
	return filtered
}
