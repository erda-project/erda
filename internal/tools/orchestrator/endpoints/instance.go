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

package endpoints

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/pkg/user"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/impl/instanceinfo"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httputil"
)

// InstanceRequestType 获取实例输入的请求参数
type InstanceRequestType string

const (
	INSTANCE_REQUEST_TYPE_PROJECT     InstanceRequestType = "project"
	INSTANCE_REQUEST_TYPE_APPLICATION InstanceRequestType = "application"
	INSTANCE_REQUEST_TYPE_SERVICE     InstanceRequestType = "service"
	INSTANCE_REQUEST_TYPE_RUNTIME     InstanceRequestType = "runtime"
)

// ListServiceInstance 获取runtime 下服务实例列表
func (e *Endpoints) ListServiceInstance(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	_, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrListInstance.NotLogin().ToResp(), nil
	}
	if err != nil {
		return apierrors.ErrListInstance.NotLogin().ToResp(), nil
	}

	// runtimeID
	runtimeID := r.URL.Query().Get("runtime")
	if runtimeID == "" {
		runtimeID = r.URL.Query().Get("runtimeID")
		if runtimeID == "" {
			return apierrors.ErrListInstance.MissingParameter("runtimeID").ToResp(), nil
		}
	}

	// serviceName
	serviceName := r.URL.Query().Get("service")
	if serviceName == "" {
		serviceName = r.URL.Query().Get("serviceName")
		if serviceName == "" {
			return apierrors.ErrListInstance.MissingParameter("serviceName").ToResp(), nil
		}
	}

	// status
	status := r.URL.Query().Get("status")

	// instanceip
	ip := r.URL.Query().Get("ip")

	req := apistructs.InstanceInfoRequest{
		RuntimeID:   runtimeID,
		ServiceName: serviceName,
	}
	if ip != "" {
		req.InstanceIP = ip
	}
	switch status {
	case "", "running":
		req.Phases = []string{apistructs.InstanceStatusHealthy, apistructs.InstanceStatusUnHealthy, apistructs.InstanceStatusRunning}
	case "stopped":
		req.Phases = []string{apistructs.InstanceStatusDead}
	default:
		req.Phases = []string{status}
	}

	instances, err := e.getContainers(req)
	if err != nil {
		return apierrors.ErrListInstance.InternalError(err).ToResp(), nil
	}

	// 按时间降序排列
	sort.Sort(sort.Reverse(&instances))
	return httpserver.OkResp(instances)
}

func (e *Endpoints) getContainers(req apistructs.InstanceInfoRequest) (apistructs.Containers, error) {
	instanceList, err := e.instanceinfoImpl.GetInstanceInfo(req)
	if err != nil {
		return nil, err
	}
	instances := make(apistructs.Containers, 0, len(instanceList))
	for _, v := range instanceList {
		instance := apistructs.Container{
			K8sInstanceMetaInfo: parseInstanceMeta(v.Meta),
			ID:                  v.TaskID,
			ContainerID:         v.ContainerID,
			IPAddress:           v.ContainerIP,
			Host:                v.HostIP,
			Image:               v.Image,
			CPU:                 v.CpuRequest,
			Memory:              int64(v.MemRequest),
			Status:              v.Phase,
			ExitCode:            v.ExitCode,
			Message:             v.Message,
			StartedAt:           v.StartedAt.Format(time.RFC3339Nano),
			Service:             v.ServiceName,
			ClusterName:         v.Cluster,
		}
		if v.FinishedAt != nil {
			instance.FinishedAt = v.FinishedAt.Format(time.RFC3339Nano)
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

func (e *Endpoints) ListServicePod(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	_, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrListInstance.NotLogin().ToResp(), nil
	}
	orgID, err := getOrgID(r)
	if err != nil {
		return apierrors.ErrListInstance.NotLogin().ToResp(), nil
	}
	// runtimeID
	runtimeID := r.URL.Query().Get("runtime")
	if runtimeID == "" {
		runtimeID = r.URL.Query().Get("runtimeID")
		if runtimeID == "" {
			return apierrors.ErrListInstance.MissingParameter("runtimeID").ToResp(), nil
		}
	}

	// serviceName
	serviceName := r.URL.Query().Get("service")
	if serviceName == "" {
		serviceName = r.URL.Query().Get("serviceName")
		if serviceName == "" {
			return apierrors.ErrListInstance.MissingParameter("serviceName").ToResp(), nil
		}
	}

	runtimeId, err := strconv.ParseUint(runtimeID, 10, 64)
	if err != nil {
		return nil, errors.Errorf("failed to parse runtime id %s: %v", runtimeID, err)
	}
	sg, err := e.runtime.GetRuntimeServiceCurrentPods(runtimeId, serviceName)
	if err != nil {
		return nil, errors.Errorf("failed to get runtime %d service %s, %v", runtimeId, serviceName, err)
	}

	currPods, err := e.instanceinfoImpl.GetInstancePodFromK8s(sg, serviceName)
	if err != nil {
		logrus.Warnf("get pod status from kubernetes failed, runtime: %s, service: %s, err: %v",
			runtimeID, serviceName, err)
	}

	if len(currPods) > 0 {
		cPods := make(apistructs.Pods, 0, len(currPods))
		for idx := range currPods {
			cPods = append(cPods, currPods[idx])
		}
		// 按时间降序排列
		sort.Sort(sort.Reverse(&cPods))
		return httpserver.OkResp(cPods)
	}

	req := apistructs.PodInfoRequest{
		OrgID:       strconv.FormatUint(orgID, 10),
		RuntimeID:   runtimeID,
		ServiceName: serviceName,
	}

	containersReq := apistructs.InstanceInfoRequest{
		RuntimeID:   runtimeID,
		ServiceName: serviceName,
		Phases:      []string{apistructs.InstanceStatusHealthy, apistructs.InstanceStatusUnHealthy, apistructs.InstanceStatusRunning},
	}

	instances, err := e.getContainers(containersReq)
	if err != nil {
		return apierrors.ErrListInstance.InternalError(err).ToResp(), nil
	}

	podList, err := e.instanceinfoImpl.GetPodInfo(req)
	if err != nil {
		return apierrors.ErrListInstance.InternalError(err).ToResp(), nil
	}
	pods := make(apistructs.Pods, 0, len(podList))
	for _, v := range podList {
		var (
			startedAt,
			updatedAt string
			podHealthy = instanceinfo.PodStatusUnHealthy
		)

		switch v.Phase {
		case string(corev1.PodRunning):
			podHealthy = instanceinfo.PodStatusHealthy
		case string(corev1.PodPending):
			podHealthy = instanceinfo.PodStatusCreating
		case string(corev1.PodFailed), string(corev1.PodUnknown):
			podHealthy = instanceinfo.PodStatusUnHealthy
		case string(corev1.PodSucceeded):
			podHealthy = instanceinfo.PodStatusTerminated
		}
		if v.StartedAt != nil {
			startedAt = v.StartedAt.Format(time.RFC3339Nano)
		}
		containersResource := make([]apistructs.PodContainer, 0)
		for _, cInstance := range instances {
			if cInstance.PodName == v.PodName && cInstance.PodNamespace == v.K8sNamespace {
				containersResource = append(containersResource, apistructs.PodContainer{
					ContainerID:   cInstance.ContainerID,
					ContainerName: cInstance.ContainerName,
					Image:         cInstance.Image,
					Resource: apistructs.ContainerResource{
						MemRequest: v.MemRequest,
						MemLimit:   v.MemLimit,
						CpuRequest: v.CpuRequest,
						CpuLimit:   v.CpuLimit,
					},
				})
				updatedAt = cInstance.StartedAt
				break
			}
		}

		pod := apistructs.Pod{
			Uid:           v.Uid,
			IPAddress:     v.PodIP,
			Host:          v.HostIP,
			Phase:         podHealthy,
			Message:       v.Message,
			StartedAt:     startedAt,
			UpdatedAt:     updatedAt,
			Service:       v.ServiceName,
			ClusterName:   v.Cluster,
			PodName:       v.PodName,
			K8sNamespace:  v.K8sNamespace,
			PodContainers: containersResource,
		}
		pods = append(pods, pod)
	}
	// 按时间降序排列
	sort.Sort(sort.Reverse(&pods))
	return httpserver.OkResp(pods)
}

func buildInstanceUsageParams(instances apistructs.InstanceInfoDataList) (
	project_ids []string, app_ids []string, runtime_ids []string, service_names []string) {
	project_ids = []string{}
	app_ids = []string{}
	runtime_ids = []string{}
	service_names = []string{}
	for _, ins := range instances {
		project_ids = append(project_ids, ins.ProjectID)
		app_ids = append(app_ids, ins.ApplicationID)
		runtime_ids = append(runtime_ids, ins.RuntimeID)
		service_names = append(service_names, ins.ServiceName)
	}
	return
}

func (e *Endpoints) InstancesUsage(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	requestType := r.URL.Query().Get("type")
	if requestType == "" {
		return apierrors.ErrListInstance.MissingParameter("type").ToResp(), nil
	}
	var project_ids, app_ids, runtime_ids, service_names []string
	ip_s := r.URL.Query().Get("ip")
	if ip_s != "" {
		cluster := vars["cluster"]
		req := apistructs.InstanceInfoRequest{
			InstanceIP: ip_s,
			Phases:     []string{"unhealthy", "healthy", "running"}, // exclude 'dead'
			Workspace:  r.URL.Query().Get("environment"),
		}
		if cluster != "" {
			req.Cluster = cluster
		}
		instanceList, err := e.instanceinfoImpl.GetInstanceInfo(req)
		if err != nil {
			return apierrors.ErrListInstance.InternalError(err).ToResp(), nil
		}
		project_ids, app_ids, runtime_ids, service_names = buildInstanceUsageParams(instanceList)
	}

	switch InstanceRequestType(requestType) {
	case INSTANCE_REQUEST_TYPE_PROJECT:
		return e.ListProjectUsage(ctx, r, vars, project_ids)
	case INSTANCE_REQUEST_TYPE_APPLICATION:
		return e.ListAppUsage(ctx, r, vars, app_ids)
	case INSTANCE_REQUEST_TYPE_RUNTIME:
		return e.ListRuntimeUsage(ctx, r, vars, runtime_ids)
	case INSTANCE_REQUEST_TYPE_SERVICE:
		return e.ListServiceUsage(ctx, r, vars, service_names)
	default:
		return apierrors.ErrListInstance.InvalidParameter("type").ToResp(), nil
	}
}

// ListProjectUsage 获取 project 资源分配列表
func (e *Endpoints) ListProjectUsage(ctx context.Context, r *http.Request, vars map[string]string, limited_project_ids []string) (httpserver.Responser, error) {
	orgID := r.Header.Get(httputil.OrgHeader)
	if orgID == "" {
		return apierrors.ErrListInstance.InvalidParameter("org id header").ToResp(), nil
	}

	cluster := vars["cluster"]
	if cluster == "" {
		return apierrors.ErrListInstance.MissingParameter("cluster").ToResp(), nil
	}

	environment := r.URL.Query().Get("environment")

	result, err := e.instance.ListProjectUsageByCluster(orgID, cluster, environment, limited_project_ids)
	if err != nil {
		return apierrors.ErrListInstance.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(result)
}

// ListAppUsage 获取应用维度资源使用统计
func (e *Endpoints) ListAppUsage(ctx context.Context, r *http.Request, vars map[string]string, limited_app_ids []string) (httpserver.Responser, error) {
	orgID := r.Header.Get(httputil.OrgHeader)
	if orgID == "" {
		return apierrors.ErrListInstance.InvalidParameter("org id header").ToResp(), nil
	}

	projectID := r.URL.Query().Get("project")
	if projectID == "" {
		return apierrors.ErrListInstance.MissingParameter("project").ToResp(), nil
	}

	environment := r.URL.Query().Get("environment")

	result, err := e.instance.ListAppUsageByProject(orgID, projectID, environment, limited_app_ids)
	if err != nil {
		return apierrors.ErrListInstance.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(result)
}

// ListRuntimeUsage 获取 runtime 维度资源使用统计
func (e *Endpoints) ListRuntimeUsage(ctx context.Context, r *http.Request, vars map[string]string, limited_runtime_ids []string) (httpserver.Responser, error) {
	orgID := r.Header.Get(httputil.OrgHeader)
	if orgID == "" {
		return apierrors.ErrListInstance.InvalidParameter("org id header").ToResp(), nil
	}

	appID := r.URL.Query().Get("application")
	if appID == "" {
		return apierrors.ErrListInstance.MissingParameter("application").ToResp(), nil
	}

	environment := r.URL.Query().Get("environment")

	result, err := e.instance.ListRuntimeUsageByApp(orgID, appID, environment, limited_runtime_ids)
	if err != nil {
		return apierrors.ErrListInstance.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(result)
}

// ListServiceUsage 获取 service 维度资源使用统计
func (e *Endpoints) ListServiceUsage(ctx context.Context, r *http.Request, vars map[string]string, limited_service_name []string) (httpserver.Responser, error) {
	orgID := r.Header.Get(httputil.OrgHeader)
	if orgID == "" {
		return apierrors.ErrListInstance.InvalidParameter("org id header").ToResp(), nil
	}

	runtimeID := r.URL.Query().Get("runtime")
	if runtimeID == "" {
		return apierrors.ErrListInstance.MissingParameter("runtime").ToResp(), nil
	}

	environment := r.URL.Query().Get("environment")

	result, err := e.instance.ListServiceUsageByRuntime(orgID, runtimeID, environment, limited_service_name)
	if err != nil {
		return apierrors.ErrListInstance.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(result)
}

func (e *Endpoints) isInstanceRunning(status string) bool {
	switch status {
	case apistructs.InstanceStatusRunning, apistructs.InstanceStatusUnHealthy, apistructs.InstanceStatusHealthy:
		return true
	default:
		return false
	}
}

func parseInstanceMeta(meta string) apistructs.K8sInstanceMetaInfo {
	info := apistructs.K8sInstanceMetaInfo{}

	kvs := strings.Split(meta, ",")
	if len(kvs) == 0 {
		return info
	}

	for _, kv := range kvs {
		rs := strings.Split(kv, "=")
		if len(rs) != 2 {
			continue
		}
		k := rs[0]
		v := rs[1]

		switch k {
		case apistructs.K8sNamespace:
			info.PodNamespace = v
		case apistructs.K8sPodName:
			info.PodName = v
		case apistructs.K8sContainerName:
			info.ContainerName = v
		case apistructs.K8sPodUid:
			info.PodUid = v
		}
	}

	return info
}
