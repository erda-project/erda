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
	"bytes"
	"testing"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
	"github.com/erda-project/erda/tools/cli/utils"
)

func TestRuntimeCommandShape(t *testing.T) {
	if RUNTIME.Name != "runtime" {
		t.Fatalf("runtime command name = %q, want runtime", RUNTIME.Name)
	}
	if RUNTIMELIST.ParentName != "RUNTIME" {
		t.Fatalf("runtime list parent = %q, want RUNTIME", RUNTIMELIST.ParentName)
	}
	if RUNTIMESTATUS.ParentName != "RUNTIME" {
		t.Fatalf("runtime status parent = %q, want RUNTIME", RUNTIMESTATUS.ParentName)
	}
	if RUNTIMELOGS.ParentName != "RUNTIME" {
		t.Fatalf("runtime logs parent = %q, want RUNTIME", RUNTIMELOGS.ParentName)
	}
	if RUNTIMEINSTANCE.ParentName != "RUNTIME" {
		t.Fatalf("runtime instance parent = %q, want RUNTIME", RUNTIMEINSTANCE.ParentName)
	}
	if RUNTIMEINSTANCELIST.ParentName != "RUNTIMEINSTANCE" {
		t.Fatalf("runtime instance list parent = %q, want RUNTIMEINSTANCE", RUNTIMEINSTANCELIST.ParentName)
	}
	if RUNTIMEINSTANCELOGS.ParentName != "RUNTIMEINSTANCE" {
		t.Fatalf("runtime instance logs parent = %q, want RUNTIMEINSTANCE", RUNTIMEINSTANCELOGS.ParentName)
	}
	if !bytes.Contains([]byte(RUNTIMESTATUS.Example), []byte("erda-cli runtime status")) {
		t.Fatalf("runtime status example = %q, want runtime status usage", RUNTIMESTATUS.Example)
	}
	if !bytes.Contains([]byte(RUNTIMEINSTANCELOGS.Example), []byte("erda-cli runtime instance logs")) {
		t.Fatalf("runtime instance logs example = %q, want runtime instance logs usage", RUNTIMEINSTANCELOGS.Example)
	}
}

func TestRuntimeCommandFlags(t *testing.T) {
	for _, name := range []string{"workspace", "runtime-id"} {
		if !hasCommandFlag(RUNTIMESTATUS.Flags, name) {
			t.Fatalf("runtime status missing --%s", name)
		}
		if !hasCommandFlag(RUNTIMELIST.Flags, name) {
			t.Fatalf("runtime list missing --%s", name)
		}
	}

	for _, name := range []string{"workspace", "runtime-id", "service", "instance", "tail", "stream", "watch"} {
		if !hasCommandFlag(RUNTIMELOGS.Flags, name) {
			t.Fatalf("runtime logs missing --%s", name)
		}
	}

	for _, name := range []string{"workspace", "runtime-id", "service"} {
		if !hasCommandFlag(RUNTIMEINSTANCELIST.Flags, name) {
			t.Fatalf("runtime instance list missing --%s", name)
		}
	}
	if !hasCommandFlag(RUNTIMEINSTANCELIST.Flags, "all") {
		t.Fatalf("runtime instance list missing --all")
	}

	for _, name := range []string{"workspace", "runtime-id", "service", "instance", "tail", "stream", "watch"} {
		if !hasCommandFlag(RUNTIMEINSTANCELOGS.Flags, name) {
			t.Fatalf("runtime instance logs missing --%s", name)
		}
	}
}

func TestRuntimeListUsesResolvedWorkspace(t *testing.T) {
	origGetWorkspaceBranch := getWorkspaceBranch
	origGetWorkspaceInfo := getWorkspaceInfo
	origGetOrgDetail := getOrgDetail
	origResolveWorkspaceApplication := resolveWorkspaceApplication
	origGetCurrentBranchWorkspace := getCurrentBranchWorkspace
	origListApplicationRuntimes := listApplicationRuntimes
	origRuntimeStdout := runtimeStdout
	t.Cleanup(func() {
		getWorkspaceBranch = origGetWorkspaceBranch
		getWorkspaceInfo = origGetWorkspaceInfo
		getOrgDetail = origGetOrgDetail
		resolveWorkspaceApplication = origResolveWorkspaceApplication
		getCurrentBranchWorkspace = origGetCurrentBranchWorkspace
		listApplicationRuntimes = origListApplicationRuntimes
		runtimeStdout = origRuntimeStdout
	})

	getWorkspaceBranch = func(string) (string, error) {
		return "feature/demo", nil
	}
	getWorkspaceInfo = func(string, string) (utils.GitterURLInfo, error) {
		return utils.GitterURLInfo{
			OrganizationURLInfo: utils.OrganizationURLInfo{Org: "erda"},
			Project:             "demo-project",
			Application:         "demo-app",
		}, nil
	}
	getOrgDetail = func(*command.Context, string) (apistructs.OrgDTO, error) {
		return apistructs.OrgDTO{ID: 1001}, nil
	}
	resolveWorkspaceApplication = func(*command.Context, uint64, string, string) (uint64, int64, error) {
		return 2001, 3001, nil
	}
	getCurrentBranchWorkspace = func(*command.Context, uint64, string) (string, error) {
		return "TEST", nil
	}
	listApplicationRuntimes = func(*command.Context, string, uint64) ([]apistructs.RuntimeSummaryDTO, error) {
		return []apistructs.RuntimeSummaryDTO{
			{RuntimeInspectDTO: apistructs.RuntimeInspectDTO{ID: 1001, Status: "Healthy", UpdatedAt: time.Date(2026, 4, 15, 9, 0, 0, 0, time.UTC), Extra: map[string]interface{}{"workspace": "DEV"}}},
			{RuntimeInspectDTO: apistructs.RuntimeInspectDTO{ID: 1002, Status: "Healthy", UpdatedAt: time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC), Extra: map[string]interface{}{"workspace": "TEST"}}},
		}, nil
	}

	var out bytes.Buffer
	runtimeStdout = &out

	if err := RuntimeList(&command.Context{}, "", 0); err != nil {
		t.Fatalf("RuntimeList() error = %v", err)
	}

	got := out.String()
	if !bytes.Contains([]byte(got), []byte("runtimes (appID=3001, workspace=TEST, total=1)")) {
		t.Fatalf("RuntimeList() output = %q, want filtered header", got)
	}
	if !bytes.Contains([]byte(got), []byte("1002")) {
		t.Fatalf("RuntimeList() output = %q, want TEST runtime row", got)
	}
	if bytes.Contains([]byte(got), []byte("1001")) {
		t.Fatalf("RuntimeList() output = %q, should exclude non-workspace runtime", got)
	}
}

func TestRuntimeListFallsBackWhenNameMissing(t *testing.T) {
	runtimes := []apistructs.RuntimeSummaryDTO{
		{
			RuntimeInspectDTO: apistructs.RuntimeInspectDTO{
				ID:              1002,
				Name:            "",
				ApplicationName: "demo-app",
				Status:          "Healthy",
				ReleaseVersion:  "v1.0.0",
				UpdatedAt:       time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC),
				Extra:           map[string]interface{}{"workspace": "TEST"},
			},
		},
	}

	var out bytes.Buffer
	if err := writeRuntimeList(&out, 3001, "TEST", runtimes); err != nil {
		t.Fatalf("writeRuntimeList() error = %v", err)
	}

	got := out.String()
	if !bytes.Contains([]byte(got), []byte("demo-app/TEST")) {
		t.Fatalf("writeRuntimeList() output = %q, want fallback display name", got)
	}
}

func TestRuntimeStatusResolvesLatestRuntimeInWorkspace(t *testing.T) {
	origGetWorkspaceBranch := getWorkspaceBranch
	origGetWorkspaceInfo := getWorkspaceInfo
	origGetOrgDetail := getOrgDetail
	origResolveWorkspaceApplication := resolveWorkspaceApplication
	origGetCurrentBranchWorkspace := getCurrentBranchWorkspace
	origListApplicationRuntimes := listApplicationRuntimes
	origInspectRuntime := inspectRuntime
	origRuntimeStdout := runtimeStdout
	t.Cleanup(func() {
		getWorkspaceBranch = origGetWorkspaceBranch
		getWorkspaceInfo = origGetWorkspaceInfo
		getOrgDetail = origGetOrgDetail
		resolveWorkspaceApplication = origResolveWorkspaceApplication
		getCurrentBranchWorkspace = origGetCurrentBranchWorkspace
		listApplicationRuntimes = origListApplicationRuntimes
		inspectRuntime = origInspectRuntime
		runtimeStdout = origRuntimeStdout
	})

	getWorkspaceBranch = func(string) (string, error) {
		return "release/1.0", nil
	}
	getWorkspaceInfo = func(string, string) (utils.GitterURLInfo, error) {
		return utils.GitterURLInfo{
			OrganizationURLInfo: utils.OrganizationURLInfo{Org: "erda"},
			Project:             "demo-project",
			Application:         "demo-app",
		}, nil
	}
	getOrgDetail = func(*command.Context, string) (apistructs.OrgDTO, error) {
		return apistructs.OrgDTO{ID: 1001}, nil
	}
	resolveWorkspaceApplication = func(*command.Context, uint64, string, string) (uint64, int64, error) {
		return 2001, 3001, nil
	}
	getCurrentBranchWorkspace = func(*command.Context, uint64, string) (string, error) {
		return "PROD", nil
	}
	listApplicationRuntimes = func(*command.Context, string, uint64) ([]apistructs.RuntimeSummaryDTO, error) {
		return []apistructs.RuntimeSummaryDTO{
			{RuntimeInspectDTO: apistructs.RuntimeInspectDTO{ID: 2001, Status: "Healthy", UpdatedAt: time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC), Extra: map[string]interface{}{"workspace": "PROD"}}},
			{RuntimeInspectDTO: apistructs.RuntimeInspectDTO{ID: 2005, Status: "Healthy", UpdatedAt: time.Date(2026, 4, 15, 11, 0, 0, 0, time.UTC), Extra: map[string]interface{}{"workspace": "PROD"}}},
		}, nil
	}

	var gotRuntimeID uint64
	inspectRuntime = func(_ *command.Context, orgID uint64, runtimeID uint64, applicationID uint64, workspace string) (apistructs.RuntimeInspectDTO, error) {
		if orgID != 1001 {
			t.Fatalf("inspectRuntime() orgID = %d, want 1001", orgID)
		}
		gotRuntimeID = runtimeID
		return apistructs.RuntimeInspectDTO{
			ID:              runtimeID,
			Name:            "demo-prod",
			Status:          "Healthy",
			ApplicationName: "demo-app",
			ApplicationID:   applicationID,
			ReleaseVersion:  "v1.0.0",
			UpdatedAt:       time.Date(2026, 4, 15, 11, 0, 0, 0, time.UTC),
			Extra:           map[string]interface{}{"workspace": workspace},
			Services: map[string]*apistructs.RuntimeInspectServiceDTO{
				"web": {
					Status: "Healthy",
					Deployments: apistructs.RuntimeServiceDeploymentsDTO{
						Replicas: 2,
					},
				},
			},
		}, nil
	}

	var out bytes.Buffer
	runtimeStdout = &out

	if err := RuntimeStatus(&command.Context{}, "", 0); err != nil {
		t.Fatalf("RuntimeStatus() error = %v", err)
	}
	if gotRuntimeID != 2005 {
		t.Fatalf("RuntimeStatus() resolved runtimeID = %d, want 2005", gotRuntimeID)
	}

	got := out.String()
	if !bytes.Contains([]byte(got), []byte("Runtime")) {
		t.Fatalf("RuntimeStatus() output = %q, want runtime header", got)
	}
	if !bytes.Contains([]byte(got), []byte("demo-prod")) || !bytes.Contains([]byte(got), []byte("web")) {
		t.Fatalf("RuntimeStatus() output = %q, want runtime summary and service row", got)
	}
}

func TestRuntimeInstanceListUsesRuntimeServices(t *testing.T) {
	origGetWorkspaceInfo := getWorkspaceInfo
	origGetOrgDetail := getOrgDetail
	origResolveWorkspaceApplication := resolveWorkspaceApplication
	origInspectRuntime := inspectRuntime
	origListRuntimeServicePodsForInstances := listRuntimeServicePodsForInstances
	origListStoppedRuntimeServiceContainersForInstances := listStoppedRuntimeServiceContainersForInstances
	origRuntimeStdout := runtimeStdout
	t.Cleanup(func() {
		getWorkspaceInfo = origGetWorkspaceInfo
		getOrgDetail = origGetOrgDetail
		resolveWorkspaceApplication = origResolveWorkspaceApplication
		inspectRuntime = origInspectRuntime
		listRuntimeServicePodsForInstances = origListRuntimeServicePodsForInstances
		listStoppedRuntimeServiceContainersForInstances = origListStoppedRuntimeServiceContainersForInstances
		runtimeStdout = origRuntimeStdout
	})

	getWorkspaceInfo = func(string, string) (utils.GitterURLInfo, error) {
		return utils.GitterURLInfo{
			OrganizationURLInfo: utils.OrganizationURLInfo{Org: "erda"},
			Project:             "demo-project",
			Application:         "demo-app",
		}, nil
	}
	getOrgDetail = func(*command.Context, string) (apistructs.OrgDTO, error) {
		return apistructs.OrgDTO{ID: 1001}, nil
	}
	resolveWorkspaceApplication = func(*command.Context, uint64, string, string) (uint64, int64, error) {
		return 2001, 3001, nil
	}
	inspectRuntime = func(_ *command.Context, orgID uint64, runtimeID uint64, applicationID uint64, workspace string) (apistructs.RuntimeInspectDTO, error) {
		if orgID != 1001 {
			t.Fatalf("inspectRuntime() orgID = %d, want 1001", orgID)
		}
		return apistructs.RuntimeInspectDTO{
			ID: runtimeID,
			Services: map[string]*apistructs.RuntimeInspectServiceDTO{
				"web":    {},
				"worker": {},
			},
		}, nil
	}

	calledServices := make([]string, 0, 2)
	listRuntimeServicePodsForInstances = func(_ *command.Context, orgID uint64, runtimeID int64, service string) (apistructs.Pods, error) {
		if orgID != 1001 {
			t.Fatalf("listRuntimeServicePodsForInstances() orgID = %d, want 1001", orgID)
		}
		calledServices = append(calledServices, service)
		return apistructs.Pods{
			{
				Service:   service,
				IPAddress: "10.0.0.1",
				Phase:     "Running",
				StartedAt: "2026-04-15T12:00:00Z",
				PodName:   service + "-0",
			},
		}, nil
	}
	listStoppedRuntimeServiceContainersForInstances = func(_ *command.Context, orgID uint64, runtimeID int64, service string) (apistructs.Containers, error) {
		t.Fatalf("listStoppedRuntimeServiceContainersForInstances() should not be called without --all")
		return nil, nil
	}

	var out bytes.Buffer
	runtimeStdout = &out

	if err := RuntimeInstanceList(&command.Context{}, "TEST", 2001, "", false); err != nil {
		t.Fatalf("RuntimeInstanceList() error = %v", err)
	}

	if len(calledServices) != 2 {
		t.Fatalf("RuntimeInstanceList() services queried = %v, want 2 services", calledServices)
	}
	got := out.String()
	if !bytes.Contains([]byte(got), []byte("instances (runtimeID=2001, total=2)")) {
		t.Fatalf("RuntimeInstanceList() output = %q, want instances header", got)
	}
	if !bytes.Contains([]byte(got), []byte("web-0")) || !bytes.Contains([]byte(got), []byte("worker-0")) {
		t.Fatalf("RuntimeInstanceList() output = %q, want aggregated instance rows", got)
	}
}

func TestRuntimeInstanceListAllIncludesStoppedInstances(t *testing.T) {
	origGetWorkspaceInfo := getWorkspaceInfo
	origGetOrgDetail := getOrgDetail
	origResolveWorkspaceApplication := resolveWorkspaceApplication
	origInspectRuntime := inspectRuntime
	origListRuntimeServicePodsForInstances := listRuntimeServicePodsForInstances
	origListStoppedRuntimeServiceContainersForInstances := listStoppedRuntimeServiceContainersForInstances
	origRuntimeStdout := runtimeStdout
	t.Cleanup(func() {
		getWorkspaceInfo = origGetWorkspaceInfo
		getOrgDetail = origGetOrgDetail
		resolveWorkspaceApplication = origResolveWorkspaceApplication
		inspectRuntime = origInspectRuntime
		listRuntimeServicePodsForInstances = origListRuntimeServicePodsForInstances
		listStoppedRuntimeServiceContainersForInstances = origListStoppedRuntimeServiceContainersForInstances
		runtimeStdout = origRuntimeStdout
	})

	getWorkspaceInfo = func(string, string) (utils.GitterURLInfo, error) {
		return utils.GitterURLInfo{
			OrganizationURLInfo: utils.OrganizationURLInfo{Org: "erda"},
			Project:             "demo-project",
			Application:         "demo-app",
		}, nil
	}
	getOrgDetail = func(*command.Context, string) (apistructs.OrgDTO, error) {
		return apistructs.OrgDTO{ID: 1001}, nil
	}
	resolveWorkspaceApplication = func(*command.Context, uint64, string, string) (uint64, int64, error) {
		return 2001, 3001, nil
	}
	inspectRuntime = func(_ *command.Context, orgID uint64, runtimeID uint64, applicationID uint64, workspace string) (apistructs.RuntimeInspectDTO, error) {
		return apistructs.RuntimeInspectDTO{
			ID: runtimeID,
			Services: map[string]*apistructs.RuntimeInspectServiceDTO{
				"web": {},
			},
		}, nil
	}
	listRuntimeServicePodsForInstances = func(_ *command.Context, orgID uint64, runtimeID int64, service string) (apistructs.Pods, error) {
		return apistructs.Pods{
			{
				Service:   service,
				IPAddress: "10.0.0.1",
				Phase:     "Running",
				StartedAt: "2026-04-15T12:00:00Z",
				UpdatedAt: "2026-04-16T11:30:38Z",
				PodName:   service + "-0",
			},
		}, nil
	}
	listStoppedRuntimeServiceContainersForInstances = func(_ *command.Context, orgID uint64, runtimeID int64, service string) (apistructs.Containers, error) {
		return apistructs.Containers{
			{
				Service:     service,
				Status:      "Stopped",
				IPAddress:   "10.0.0.2",
				StartedAt:   "2026-04-15T10:00:00Z",
				FinishedAt:  "2026-04-16T10:52:20Z",
				ContainerID: "stopped-container",
			},
		}, nil
	}

	var out bytes.Buffer
	runtimeStdout = &out

	if err := RuntimeInstanceList(&command.Context{}, "TEST", 2001, "", true); err != nil {
		t.Fatalf("RuntimeInstanceList() error = %v", err)
	}

	got := out.String()
	if !bytes.Contains([]byte(got), []byte("instances (runtimeID=2001, total=2)")) {
		t.Fatalf("RuntimeInstanceList() output = %q, want instances header", got)
	}
	if !bytes.Contains([]byte(got), []byte("web-0")) || !bytes.Contains([]byte(got), []byte("stopped-container")) {
		t.Fatalf("RuntimeInstanceList() output = %q, want running and stopped instances", got)
	}
	if !bytes.Contains([]byte(got), []byte("Stopped")) {
		t.Fatalf("RuntimeInstanceList() output = %q, want stopped status row", got)
	}
	if !bytes.Contains([]byte(got), []byte("FINISHEDAT")) {
		t.Fatalf("RuntimeInstanceList() output = %q, want finishedAt header", got)
	}
	if !bytes.Contains([]byte(got), []byte("2026-04-16T10:52:20Z")) {
		t.Fatalf("RuntimeInstanceList() output = %q, want finishedAt value", got)
	}
	if !bytes.Contains([]byte(got), []byte("web-0               Running")) || !bytes.Contains([]byte(got), []byte("-")) {
		t.Fatalf("RuntimeInstanceList() output = %q, want placeholder finishedAt for running instance", got)
	}
	if bytes.Index([]byte(got), []byte("web-0")) > bytes.Index([]byte(got), []byte("stopped-container")) {
		t.Fatalf("RuntimeInstanceList() output = %q, want reverse chronological ordering", got)
	}
}

func TestRuntimeLogsFiltersByServiceAndInstance(t *testing.T) {
	origGetWorkspaceInfo := getWorkspaceInfo
	origGetOrgDetail := getOrgDetail
	origResolveWorkspaceApplication := resolveWorkspaceApplication
	origListRuntimeServicePodsForLogs := listRuntimeServicePodsForLogs
	origListStoppedRuntimeServiceContainersForLogs := listStoppedRuntimeServiceContainersForLogs
	origGetRuntimePodLogs := getRuntimePodLogs
	origGetRuntimeStoppedContainerLogs := getRuntimeStoppedContainerLogs
	origRuntimeStdout := runtimeStdout
	t.Cleanup(func() {
		getWorkspaceInfo = origGetWorkspaceInfo
		getOrgDetail = origGetOrgDetail
		resolveWorkspaceApplication = origResolveWorkspaceApplication
		listRuntimeServicePodsForLogs = origListRuntimeServicePodsForLogs
		listStoppedRuntimeServiceContainersForLogs = origListStoppedRuntimeServiceContainersForLogs
		getRuntimePodLogs = origGetRuntimePodLogs
		getRuntimeStoppedContainerLogs = origGetRuntimeStoppedContainerLogs
		runtimeStdout = origRuntimeStdout
	})

	getWorkspaceInfo = func(string, string) (utils.GitterURLInfo, error) {
		return utils.GitterURLInfo{
			OrganizationURLInfo: utils.OrganizationURLInfo{Org: "erda"},
			Project:             "demo-project",
			Application:         "demo-app",
		}, nil
	}
	getOrgDetail = func(*command.Context, string) (apistructs.OrgDTO, error) {
		return apistructs.OrgDTO{ID: 1001}, nil
	}
	resolveWorkspaceApplication = func(*command.Context, uint64, string, string) (uint64, int64, error) {
		return 2001, 3001, nil
	}
	listRuntimeServicePodsForLogs = func(_ *command.Context, orgID uint64, runtimeID int64, service string) (apistructs.Pods, error) {
		return apistructs.Pods{
			{
				Service:      service,
				ClusterName:  "erda-cloud",
				PodName:      "web-0",
				K8sNamespace: "project-387-prod",
				PodContainers: []apistructs.PodContainer{
					{ContainerName: service, ContainerID: "matched"},
				},
			},
			{
				Service:      service,
				ClusterName:  "erda-cloud",
				PodName:      "web-1",
				K8sNamespace: "project-387-prod",
				PodContainers: []apistructs.PodContainer{
					{ContainerName: service, ContainerID: "other"},
				},
			},
		}, nil
	}
	getRuntimePodLogs = func(_ *command.Context, orgName string, applicationID uint64, pod apistructs.Pod, containerName string, containerID string, stream string, tail int) (*apistructs.DashboardSpotLogData, error) {
		return &apistructs.DashboardSpotLogData{
			Lines: []apistructs.DashboardSpotLogLine{
				{TimeStamp: "2026-04-15T12:00:00Z", Content: "line-" + containerID},
			},
		}, nil
	}
	listStoppedRuntimeServiceContainersForLogs = func(_ *command.Context, orgID uint64, runtimeID int64, service string) (apistructs.Containers, error) {
		t.Fatal("stopped containers should not be queried when running pod matched")
		return nil, nil
	}
	getRuntimeStoppedContainerLogs = func(_ *command.Context, orgName string, applicationID uint64, container apistructs.Container, stream string, tail int) (*apistructs.DashboardSpotLogData, error) {
		t.Fatal("stopped logs should not be queried when running pod matched")
		return nil, nil
	}

	var out bytes.Buffer
	runtimeStdout = &out

	if err := RuntimeLogs(&command.Context{}, "TEST", 2001, "web", "web-0", 200, "stdout", false); err != nil {
		t.Fatalf("RuntimeLogs() error = %v", err)
	}

	got := out.String()
	if !bytes.Contains([]byte(got), []byte("line-matched")) {
		t.Fatalf("RuntimeLogs() output = %q, want matched container logs", got)
	}
	if bytes.Contains([]byte(got), []byte("line-other")) {
		t.Fatalf("RuntimeLogs() output = %q, should exclude unmatched instance logs", got)
	}
}

func TestRuntimeLogsFallsBackToStoppedInstance(t *testing.T) {
	origGetWorkspaceInfo := getWorkspaceInfo
	origGetOrgDetail := getOrgDetail
	origResolveWorkspaceApplication := resolveWorkspaceApplication
	origListRuntimeServicePodsForLogs := listRuntimeServicePodsForLogs
	origListStoppedRuntimeServiceContainersForLogs := listStoppedRuntimeServiceContainersForLogs
	origGetRuntimePodLogs := getRuntimePodLogs
	origGetRuntimeStoppedContainerLogs := getRuntimeStoppedContainerLogs
	origRuntimeStdout := runtimeStdout
	t.Cleanup(func() {
		getWorkspaceInfo = origGetWorkspaceInfo
		getOrgDetail = origGetOrgDetail
		resolveWorkspaceApplication = origResolveWorkspaceApplication
		listRuntimeServicePodsForLogs = origListRuntimeServicePodsForLogs
		listStoppedRuntimeServiceContainersForLogs = origListStoppedRuntimeServiceContainersForLogs
		getRuntimePodLogs = origGetRuntimePodLogs
		getRuntimeStoppedContainerLogs = origGetRuntimeStoppedContainerLogs
		runtimeStdout = origRuntimeStdout
	})

	getWorkspaceInfo = func(string, string) (utils.GitterURLInfo, error) {
		return utils.GitterURLInfo{
			OrganizationURLInfo: utils.OrganizationURLInfo{Org: "erda"},
			Project:             "demo-project",
			Application:         "demo-app",
		}, nil
	}
	getOrgDetail = func(*command.Context, string) (apistructs.OrgDTO, error) {
		return apistructs.OrgDTO{ID: 1001}, nil
	}
	resolveWorkspaceApplication = func(*command.Context, uint64, string, string) (uint64, int64, error) {
		return 2001, 3001, nil
	}
	listRuntimeServicePodsForLogs = func(_ *command.Context, orgID uint64, runtimeID int64, service string) (apistructs.Pods, error) {
		return apistructs.Pods{}, nil
	}
	getRuntimePodLogs = func(_ *command.Context, orgName string, applicationID uint64, pod apistructs.Pod, containerName string, containerID string, stream string, tail int) (*apistructs.DashboardSpotLogData, error) {
		t.Fatal("running pod logs should not be queried when no pod matched")
		return nil, nil
	}
	listStoppedRuntimeServiceContainersForLogs = func(_ *command.Context, orgID uint64, runtimeID int64, service string) (apistructs.Containers, error) {
		return apistructs.Containers{
			{
				Service:     service,
				ClusterName: "erda-cloud",
				ContainerID: "dead-container-id",
				K8sInstanceMetaInfo: apistructs.K8sInstanceMetaInfo{
					PodName:      "web-dead-0",
					PodNamespace: "project-387-prod",
				},
			},
		}, nil
	}
	getRuntimeStoppedContainerLogs = func(_ *command.Context, orgName string, applicationID uint64, container apistructs.Container, stream string, tail int) (*apistructs.DashboardSpotLogData, error) {
		if container.ContainerID != "dead-container-id" {
			t.Fatalf("containerID = %q, want dead-container-id", container.ContainerID)
		}
		return &apistructs.DashboardSpotLogData{
			Lines: []apistructs.DashboardSpotLogLine{
				{TimeStamp: "2026-04-20T15:34:33Z", Content: "dead log"},
			},
		}, nil
	}

	var out bytes.Buffer
	runtimeStdout = &out

	if err := RuntimeLogs(&command.Context{}, "TEST", 2001, "web", "web-dead-0", 200, "", false); err != nil {
		t.Fatalf("RuntimeLogs() error = %v", err)
	}

	if got := out.String(); !bytes.Contains([]byte(got), []byte("dead log")) {
		t.Fatalf("RuntimeLogs() output = %q, want stopped instance log", got)
	}
}

func TestRuntimeInstanceLogsRequiresInstance(t *testing.T) {
	err := RuntimeInstanceLogs(&command.Context{}, "", "TEST", 2001, "web", "", 200, "stdout", false)
	if err == nil {
		t.Fatal("RuntimeInstanceLogs() error = nil, want missing instance error")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("logs <instance>")) {
		t.Fatalf("RuntimeInstanceLogs() error = %q, want positional instance hint", err.Error())
	}
}

func TestResolveRuntimeInstanceLogTargetPrefersPositionalArg(t *testing.T) {
	got := resolveRuntimeInstanceLogTarget("pod-from-arg", "pod-from-flag")
	if got != "pod-from-arg" {
		t.Fatalf("resolveRuntimeInstanceLogTarget() = %q, want positional arg", got)
	}
}

func TestResolveRuntimeInstanceLogTargetFallsBackToFlag(t *testing.T) {
	got := resolveRuntimeInstanceLogTarget("", "pod-from-flag")
	if got != "pod-from-flag" {
		t.Fatalf("resolveRuntimeInstanceLogTarget() = %q, want flag value", got)
	}
}

func TestRuntimeLogsWatchPrintsOnlyUnseenLines(t *testing.T) {
	origGetWorkspaceInfo := getWorkspaceInfo
	origGetOrgDetail := getOrgDetail
	origResolveWorkspaceApplication := resolveWorkspaceApplication
	origListRuntimeServicePodsForLogs := listRuntimeServicePodsForLogs
	origGetRuntimePodLogs := getRuntimePodLogs
	origGetRuntimePodLogsWithOptions := getRuntimePodLogsWithOptions
	origRuntimeStdout := runtimeStdout
	origRuntimeLogsSleep := runtimeLogsSleep
	origRuntimeLogsShouldStop := runtimeLogsShouldStop
	t.Cleanup(func() {
		getWorkspaceInfo = origGetWorkspaceInfo
		getOrgDetail = origGetOrgDetail
		resolveWorkspaceApplication = origResolveWorkspaceApplication
		listRuntimeServicePodsForLogs = origListRuntimeServicePodsForLogs
		getRuntimePodLogs = origGetRuntimePodLogs
		getRuntimePodLogsWithOptions = origGetRuntimePodLogsWithOptions
		runtimeStdout = origRuntimeStdout
		runtimeLogsSleep = origRuntimeLogsSleep
		runtimeLogsShouldStop = origRuntimeLogsShouldStop
	})

	getWorkspaceInfo = func(string, string) (utils.GitterURLInfo, error) {
		return utils.GitterURLInfo{
			OrganizationURLInfo: utils.OrganizationURLInfo{Org: "erda"},
			Project:             "demo-project",
			Application:         "demo-app",
		}, nil
	}
	getOrgDetail = func(*command.Context, string) (apistructs.OrgDTO, error) {
		return apistructs.OrgDTO{ID: 1001}, nil
	}
	resolveWorkspaceApplication = func(*command.Context, uint64, string, string) (uint64, int64, error) {
		return 2001, 3001, nil
	}
	listRuntimeServicePodsForLogs = func(_ *command.Context, orgID uint64, runtimeID int64, service string) (apistructs.Pods, error) {
		return apistructs.Pods{
			{
				Service:      service,
				ClusterName:  "erda-cloud",
				PodName:      "web-0",
				K8sNamespace: "project-387-prod",
				PodContainers: []apistructs.PodContainer{
					{ContainerName: service, ContainerID: "matched"},
				},
			},
		}, nil
	}

	calls := 0
	getRuntimePodLogs = func(_ *command.Context, orgName string, applicationID uint64, pod apistructs.Pod, containerName string, containerID string, stream string, tail int) (*apistructs.DashboardSpotLogData, error) {
		calls++
		switch calls {
		case 1:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{TimeStamp: "2026-04-15T12:00:00Z", Content: "first"},
				},
			}, nil
		default:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{TimeStamp: "2026-04-15T12:00:00Z", Content: "first"},
					{TimeStamp: "2026-04-15T12:00:01Z", Content: "second"},
				},
			}, nil
		}
	}
	getRuntimePodLogsWithOptions = func(_ *command.Context, orgName string, applicationID uint64, pod apistructs.Pod, containerName string, containerID string, opts common.RuntimeLogOptions) (*apistructs.DashboardSpotLogData, error) {
		calls++
		return &apistructs.DashboardSpotLogData{
			Lines: []apistructs.DashboardSpotLogLine{
				{TimeStamp: "2026-04-15T12:00:00Z", Content: "first"},
				{TimeStamp: "2026-04-15T12:00:01Z", Content: "second"},
			},
		}, nil
	}

	var out bytes.Buffer
	runtimeStdout = &out
	runtimeLogsSleep = func(time.Duration) {}
	runtimeLogsShouldStop = func() bool {
		return calls >= 2
	}

	if err := RuntimeLogs(&command.Context{}, "TEST", 2001, "web", "web-0", 200, "stdout", true); err != nil {
		t.Fatalf("RuntimeLogs() watch error = %v", err)
	}

	got := out.String()
	if bytes.Count([]byte(got), []byte("first")) != 1 {
		t.Fatalf("RuntimeLogs() watch output = %q, want first line printed once", got)
	}
	if !bytes.Contains([]byte(got), []byte("second")) {
		t.Fatalf("RuntimeLogs() watch output = %q, want second line printed", got)
	}
}

func TestRuntimeLogsWatchStopsCleanlyWhenRequested(t *testing.T) {
	origGetWorkspaceInfo := getWorkspaceInfo
	origGetOrgDetail := getOrgDetail
	origResolveWorkspaceApplication := resolveWorkspaceApplication
	origListRuntimeServicePodsForLogs := listRuntimeServicePodsForLogs
	origGetRuntimePodLogs := getRuntimePodLogs
	origGetRuntimePodLogsWithOptions := getRuntimePodLogsWithOptions
	origRuntimeStdout := runtimeStdout
	origRuntimeLogsSleep := runtimeLogsSleep
	origRuntimeLogsShouldStop := runtimeLogsShouldStop
	t.Cleanup(func() {
		getWorkspaceInfo = origGetWorkspaceInfo
		getOrgDetail = origGetOrgDetail
		resolveWorkspaceApplication = origResolveWorkspaceApplication
		listRuntimeServicePodsForLogs = origListRuntimeServicePodsForLogs
		getRuntimePodLogs = origGetRuntimePodLogs
		getRuntimePodLogsWithOptions = origGetRuntimePodLogsWithOptions
		runtimeStdout = origRuntimeStdout
		runtimeLogsSleep = origRuntimeLogsSleep
		runtimeLogsShouldStop = origRuntimeLogsShouldStop
	})

	getWorkspaceInfo = func(string, string) (utils.GitterURLInfo, error) {
		return utils.GitterURLInfo{
			OrganizationURLInfo: utils.OrganizationURLInfo{Org: "erda"},
			Project:             "demo-project",
			Application:         "demo-app",
		}, nil
	}
	getOrgDetail = func(*command.Context, string) (apistructs.OrgDTO, error) {
		return apistructs.OrgDTO{ID: 1001}, nil
	}
	resolveWorkspaceApplication = func(*command.Context, uint64, string, string) (uint64, int64, error) {
		return 2001, 3001, nil
	}
	listRuntimeServicePodsForLogs = func(_ *command.Context, orgID uint64, runtimeID int64, service string) (apistructs.Pods, error) {
		return apistructs.Pods{
			{
				Service:      service,
				ClusterName:  "erda-cloud",
				PodName:      "web-0",
				K8sNamespace: "project-387-prod",
				PodContainers: []apistructs.PodContainer{
					{ContainerName: service, ContainerID: "matched"},
				},
			},
		}, nil
	}
	getRuntimePodLogs = func(_ *command.Context, orgName string, applicationID uint64, pod apistructs.Pod, containerName string, containerID string, stream string, tail int) (*apistructs.DashboardSpotLogData, error) {
		return &apistructs.DashboardSpotLogData{Lines: nil}, nil
	}
	getRuntimePodLogsWithOptions = func(_ *command.Context, orgName string, applicationID uint64, pod apistructs.Pod, containerName string, containerID string, opts common.RuntimeLogOptions) (*apistructs.DashboardSpotLogData, error) {
		return &apistructs.DashboardSpotLogData{Lines: nil}, nil
	}

	var out bytes.Buffer
	runtimeStdout = &out
	stop := false
	runtimeLogsSleep = func(time.Duration) {
		stop = true
	}
	runtimeLogsShouldStop = func() bool {
		return stop
	}

	if err := RuntimeLogs(&command.Context{}, "TEST", 2001, "web", "web-0", 200, "stdout", true); err != nil {
		t.Fatalf("RuntimeLogs() watch error = %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("RuntimeLogs() watch output = %q, want no output", out.String())
	}
}

func TestRuntimeLogsWatchPagesBeyondTailWindow(t *testing.T) {
	origGetWorkspaceInfo := getWorkspaceInfo
	origGetOrgDetail := getOrgDetail
	origResolveWorkspaceApplication := resolveWorkspaceApplication
	origListRuntimeServicePodsForLogs := listRuntimeServicePodsForLogs
	origGetRuntimePodLogs := getRuntimePodLogs
	origGetRuntimePodLogsWithOptions := getRuntimePodLogsWithOptions
	origRuntimeStdout := runtimeStdout
	origRuntimeLogsSleep := runtimeLogsSleep
	origRuntimeLogsShouldStop := runtimeLogsShouldStop
	t.Cleanup(func() {
		getWorkspaceInfo = origGetWorkspaceInfo
		getOrgDetail = origGetOrgDetail
		resolveWorkspaceApplication = origResolveWorkspaceApplication
		listRuntimeServicePodsForLogs = origListRuntimeServicePodsForLogs
		getRuntimePodLogs = origGetRuntimePodLogs
		getRuntimePodLogsWithOptions = origGetRuntimePodLogsWithOptions
		runtimeStdout = origRuntimeStdout
		runtimeLogsSleep = origRuntimeLogsSleep
		runtimeLogsShouldStop = origRuntimeLogsShouldStop
	})

	getWorkspaceInfo = func(string, string) (utils.GitterURLInfo, error) {
		return utils.GitterURLInfo{
			OrganizationURLInfo: utils.OrganizationURLInfo{Org: "erda"},
			Project:             "demo-project",
			Application:         "demo-app",
		}, nil
	}
	getOrgDetail = func(*command.Context, string) (apistructs.OrgDTO, error) {
		return apistructs.OrgDTO{ID: 1001}, nil
	}
	resolveWorkspaceApplication = func(*command.Context, uint64, string, string) (uint64, int64, error) {
		return 2001, 3001, nil
	}
	listRuntimeServicePodsForLogs = func(_ *command.Context, orgID uint64, runtimeID int64, service string) (apistructs.Pods, error) {
		return apistructs.Pods{
			{
				Service:      service,
				ClusterName:  "erda-cloud",
				PodName:      "web-0",
				K8sNamespace: "project-387-prod",
				PodContainers: []apistructs.PodContainer{
					{ContainerName: service, ContainerID: "matched"},
				},
			},
		}, nil
	}

	initialCalls := 0
	getRuntimePodLogs = func(_ *command.Context, orgName string, applicationID uint64, pod apistructs.Pod, containerName string, containerID string, stream string, tail int) (*apistructs.DashboardSpotLogData, error) {
		initialCalls++
		return &apistructs.DashboardSpotLogData{
			Lines: []apistructs.DashboardSpotLogLine{
				{TimeStamp: "1000", Content: "seed"},
			},
		}, nil
	}

	var incrementalCalls []common.RuntimeLogOptions
	getRuntimePodLogsWithOptions = func(_ *command.Context, orgName string, applicationID uint64, pod apistructs.Pod, containerName string, containerID string, opts common.RuntimeLogOptions) (*apistructs.DashboardSpotLogData, error) {
		incrementalCalls = append(incrementalCalls, opts)
		switch len(incrementalCalls) {
		case 1:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{TimeStamp: "1001", Content: "line-1"},
					{TimeStamp: "1002", Content: "line-2"},
				},
			}, nil
		case 2:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{TimeStamp: "1003", Content: "line-3"},
				},
			}, nil
		default:
			return &apistructs.DashboardSpotLogData{Lines: nil}, nil
		}
	}

	var out bytes.Buffer
	runtimeStdout = &out
	runtimeLogsSleep = func(time.Duration) {}
	runtimeLogsShouldStop = func() bool {
		return len(incrementalCalls) >= 2
	}

	if err := RuntimeLogs(&command.Context{}, "TEST", 2001, "web", "web-0", 2, "stdout", true); err != nil {
		t.Fatalf("RuntimeLogs() watch error = %v", err)
	}

	if initialCalls != 1 {
		t.Fatalf("initial tail calls = %d, want 1", initialCalls)
	}
	if len(incrementalCalls) != 2 {
		t.Fatalf("incremental calls = %d, want 2", len(incrementalCalls))
	}
	if incrementalCalls[0].Start != 999 || incrementalCalls[0].Count != 2 {
		t.Fatalf("first incremental call = %#v, want start=999 count=2", incrementalCalls[0])
	}
	if incrementalCalls[1].Start != 1001 || incrementalCalls[1].Count != 2 {
		t.Fatalf("second incremental call = %#v, want start=1001 count=2", incrementalCalls[1])
	}
	got := out.String()
	for _, want := range []string{"seed", "line-1", "line-2", "line-3"} {
		if !bytes.Contains([]byte(got), []byte(want)) {
			t.Fatalf("RuntimeLogs() watch output = %q, want %q", got, want)
		}
	}
}
