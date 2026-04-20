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

package common

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
)

func GetCurrentBranchWorkspace(ctx *command.Context, appID uint64, branch string) (string, error) {
	params := buildBranchWorkspaceQueryParams(appID)
	var resp apistructs.PipelineAppAllValidBranchWorkspaceResponse
	httpResp, err := ctx.Get().Path("/api/cicds/actions/app-all-valid-branch-workspaces").
		Param("appID", params["appID"]).
		Do().JSON(&resp)
	if err != nil {
		return "", err
	}
	if !httpResp.IsOK() {
		return "", errors.Errorf("status fail, status code: %d, err: %+v", httpResp.StatusCode(), resp.Error)
	}
	if !resp.Success {
		return "", errors.Errorf("status fail: %+v", resp.Error)
	}
	for _, item := range resp.Data {
		if item.Name == branch {
			return item.Workspace, nil
		}
	}
	return "", fmt.Errorf("no workspace mapping found for branch %s", branch)
}

func buildBranchWorkspaceQueryParams(appID uint64) map[string]string {
	return map[string]string{
		"appID": strconv.FormatUint(appID, 10),
	}
}

func ListApplicationRuntimes(ctx *command.Context, orgID string, appID uint64) ([]apistructs.RuntimeSummaryDTO, error) {
	var resp apistructs.RuntimeListResponse
	httpResp, err := ctx.Get().Path("/api/runtimes").
		Header("org", orgID).
		Param("applicationId", strconv.FormatUint(appID, 10)).
		Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !httpResp.IsOK() {
		return nil, errors.Errorf("status fail, status code: %d, err: %+v", httpResp.StatusCode(), resp.Error)
	}
	if !resp.Success {
		return nil, errors.Errorf("status fail: %+v", resp.Error)
	}
	return resp.Data, nil
}

func InspectRuntime(ctx *command.Context, orgID uint64, runtimeID uint64, applicationID uint64, workspace string) (apistructs.RuntimeInspectDTO, error) {
	var resp apistructs.RuntimeInspectResponse
	req := ctx.Get().Path(fmt.Sprintf("/api/runtimes/%d", runtimeID)).
		Header("org", strconv.FormatUint(orgID, 10)).
		Header("Org-ID", strconv.FormatUint(orgID, 10))
	if applicationID > 0 {
		req = req.Param("applicationId", strconv.FormatUint(applicationID, 10))
	}
	if workspace != "" {
		req = req.Param("workspace", workspace)
	}
	httpResp, err := req.Do().JSON(&resp)
	if err != nil {
		return apistructs.RuntimeInspectDTO{}, err
	}
	if !httpResp.IsOK() {
		return apistructs.RuntimeInspectDTO{}, errors.Errorf("status fail, status code: %d, err: %+v", httpResp.StatusCode(), resp.Error)
	}
	if !resp.Success {
		return apistructs.RuntimeInspectDTO{}, errors.Errorf("status fail: %+v", resp.Error)
	}
	return resp.Data, nil
}

func ListRuntimeServicePods(ctx *command.Context, orgID uint64, runtimeID int64, service string) (apistructs.Pods, error) {
	var resp apistructs.PodListResponse
	httpResp, err := ctx.Get().Path("/api/instances/actions/get-service-pods").
		Header("org", strconv.FormatUint(orgID, 10)).
		Header("Org-ID", strconv.FormatUint(orgID, 10)).
		Param("runtimeID", strconv.FormatInt(runtimeID, 10)).
		Param("serviceName", service).
		Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !httpResp.IsOK() {
		return nil, errors.Errorf("status fail, status code: %d, err: %+v", httpResp.StatusCode(), resp.Error)
	}
	if !resp.Success {
		return nil, errors.Errorf("status fail: %+v", resp.Error)
	}
	return resp.Data, nil
}

func ListRuntimeServiceStoppedContainers(ctx *command.Context, orgID uint64, runtimeID int64, service string) (apistructs.Containers, error) {
	var resp apistructs.ContainerListResponse
	httpResp, err := ctx.Get().Path("/api/instances/actions/get-service").
		Header("org", strconv.FormatUint(orgID, 10)).
		Header("Org-ID", strconv.FormatUint(orgID, 10)).
		Param("runtimeID", strconv.FormatInt(runtimeID, 10)).
		Param("serviceName", service).
		Param("status", "stopped").
		Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !httpResp.IsOK() {
		return nil, errors.Errorf("status fail, status code: %d, err: %+v", httpResp.StatusCode(), resp.Error)
	}
	if !resp.Success {
		return nil, errors.Errorf("status fail: %+v", resp.Error)
	}
	return resp.Data, nil
}

func ListRuntimeServiceContainers(ctx *command.Context, runtimeID int64, service string) (apistructs.Containers, error) {
	var resp apistructs.ContainerListResponse
	httpResp, err := ctx.Get().Path("/api/instances/actions/get-service").
		Param("runtimeID", strconv.FormatInt(runtimeID, 10)).
		Param("serviceName", service).
		Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !httpResp.IsOK() {
		return nil, errors.Errorf("status fail, status code: %d, err: %+v", httpResp.StatusCode(), resp.Error)
	}
	if !resp.Success {
		return nil, errors.Errorf("status fail: %+v", resp.Error)
	}
	return resp.Data, nil
}

func GetRuntimePodLogs(ctx *command.Context, orgName string, applicationID uint64, pod apistructs.Pod, containerName string, containerID string, stream string, tail int) (*apistructs.DashboardSpotLogData, error) {
	return getRuntimeContainerLogs(ctx, runtimeContainerLogRequest{
		orgName:       orgName,
		applicationID: applicationID,
		clusterName:   pod.ClusterName,
		podName:       pod.PodName,
		podNamespace:  pod.K8sNamespace,
		containerName: containerName,
		containerID:   containerID,
		live:          true,
		stream:        stream,
		tail:          tail,
	})
}

func GetRuntimeStoppedContainerLogs(ctx *command.Context, orgName string, applicationID uint64, container apistructs.Container, stream string, tail int) (*apistructs.DashboardSpotLogData, error) {
	return getRuntimeContainerLogs(ctx, runtimeContainerLogRequest{
		orgName:       orgName,
		applicationID: applicationID,
		clusterName:   container.ClusterName,
		podName:       container.PodName,
		podNamespace:  container.PodNamespace,
		containerName: container.Service,
		containerID:   container.ContainerID,
		live:          false,
		stream:        stream,
		tail:          tail,
	})
}

type runtimeContainerLogRequest struct {
	orgName       string
	applicationID uint64
	clusterName   string
	podName       string
	podNamespace  string
	containerName string
	containerID   string
	live          bool
	stream        string
	tail          int
}

func getRuntimeContainerLogs(ctx *command.Context, opts runtimeContainerLogRequest) (*apistructs.DashboardSpotLogData, error) {
	count := int64(-200)
	if opts.tail > 0 {
		count = int64(-opts.tail)
	}

	var resp apistructs.DashboardSpotLogResponse
	req := ctx.Get().Path("/api/runtime/logs").
		Header("org", opts.orgName).
		Param("applicationId", strconv.FormatUint(opts.applicationID, 10)).
		Param("id", opts.containerID).
		Param("source", string(apistructs.DashboardSpotLogSourceContainer)).
		Param("stream", opts.stream).
		Param("count", strconv.FormatInt(count, 10)).
		Param("start", "0").
		Param("end", strconv.FormatInt(time.Now().UnixNano(), 10)).
		Param("clusterName", opts.clusterName).
		Param("podName", opts.podName).
		Param("podNamespace", opts.podNamespace).
		Param("containerName", opts.containerName).
		Param("live", strconv.FormatBool(opts.live))
	httpResp, err := req.Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !httpResp.IsOK() {
		return nil, errors.Errorf("status fail, status code: %d, err: %+v", httpResp.StatusCode(), resp.Error)
	}
	if !resp.Success {
		return nil, errors.Errorf("status fail: %+v", resp.Error)
	}
	return &resp.Data, nil
}
