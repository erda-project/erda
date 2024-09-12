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
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/pkg/user"
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

	PodStatusHealthy    string = "Healthy"
	PodStatusUnHealthy  string = "Unhealthy"
	PodStatusCreating   string = "Creating"
	PodStatusTerminated string = "Terminated"

	PodConditionFalse                     string = "False"
	PodDefaultMessage                     string = "Ok"
	PodPendingDefaultMessage              string = "Waiting for scheduling"
	PodPendingReasonUnSchedulable         string = "Unschedulable"
	PodUnHealthyReasonErrImagePull        string = "ErrImagePull"
	PodUnHealthyReasonUnSchedulable       string = "Failed to Schedule Pod"
	ContainerStateReasonImagePullBackOff  string = "ImagePullBackOff"
	PodUnHealthyReasonImagePullBackOff    string = "Failed to pull image"
	ContainerStateReasonRunContainerError string = "RunContainerError"
	PodUnHealthyReasonRunContainerError   string = "Failed to start container"
	ContainerStateReasonCrashLoopBackOff  string = "CrashLoopBackOff"
	PodUnHealthyReasonCrashLoopBackOff    string = "Failed to run command"

	messageContainsImageNotFound                string = "not found"
	messageContainsImagePullBackoff             string = "Back-off pulling image"
	messageContainsImagePullFailedTrans         string = "image not found or can not pull"
	messageContainsNodeAffinity                 string = "didn't match Pod's node affinity"
	messageContainsInsufficientCPU              string = "Insufficient cpu"
	messageContainsInsufficientMemory           string = "Insufficient memory"
	messageContainsInsufficientCPUAndMemory     string = "Insufficient memory and memory"
	messageContainsNoSuchFileOrDirectory        string = "no such file or directory"
	messageContainsNoSuchFileOrDirectoryTrans   string = "some file or directory not found"
	messageContainsBackOff                      string = "back-off"
	messageContainsBackOffRestartConatiner      string = "restarting failed container"
	messageContainsBackOffRestartConatinerTrans string = "container process exit"
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

	currPods, err := e.getPodStatusFromK8s(runtimeID, serviceName)
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
		startat := ""
		updateat := ""
		podHealthy := PodStatusUnHealthy
		switch v.Phase {
		case string(corev1.PodRunning):
			podHealthy = PodStatusHealthy
		case string(corev1.PodPending):
			podHealthy = PodStatusCreating
		case string(corev1.PodFailed), string(corev1.PodUnknown):
			podHealthy = PodStatusUnHealthy
		case string(corev1.PodSucceeded):
			podHealthy = PodStatusTerminated
		}
		if v.StartedAt != nil {
			startat = v.StartedAt.Format(time.RFC3339Nano)
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
				updateat = cInstance.StartedAt
				break
			}
		}

		pod := apistructs.Pod{
			Uid:           v.Uid,
			IPAddress:     v.PodIP,
			Host:          v.HostIP,
			Phase:         podHealthy,
			Message:       v.Message,
			StartedAt:     startat,
			UpdatedAt:     updateat,
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

// getPodStatusFromK8s TODO: move all logic to the implementation.
func (e *Endpoints) getPodStatusFromK8s(runtimeID, serviceName string) ([]apistructs.Pod, error) {
	currPods := make([]apistructs.Pod, 0)
	runtimeId, _ := strconv.ParseUint(runtimeID, 10, 64)
	sg, err := e.runtime.GetRuntimeServiceCurrentPods(runtimeId, serviceName)
	if err != nil {
		return currPods, errors.Errorf("get runtimeId %d service %s current pods GetRuntimeServiceCurrentPods failed:%v", runtimeId, serviceName, err)
	}

	if _, ok := sg.Extra[serviceName]; !ok {
		return currPods, errors.Errorf("get runtimeId %d service %s current pods failed: no pods found in sg.Extra for service", runtimeId, serviceName)
	}

	var k8sPods []corev1.Pod
	err = json.Unmarshal([]byte(sg.Extra[serviceName]), &k8sPods)
	if err != nil {
		return currPods, errors.Errorf("get runtimeId %d service %s current pods failed: %v", runtimeId, serviceName, err)
	}

	for _, pod := range k8sPods {
		if pod.Status.Phase != corev1.PodPending && pod.Status.Phase != corev1.PodRunning {
			logrus.Warnf("Pod %s/%s had Status in terminated or unknown, ignoring it.", pod.Namespace, pod.Name)
			continue
		}

		// TODO change `DICE_CLUSTER_NAME` to  `core.erda.cloud/cluster-name`
		clusterName := ""
		for _, container := range pod.Spec.Containers {
			if clusterName != "" {
				break
			}
			for _, env := range container.Env {
				if env.Name == apistructs.DICE_CLUSTER_NAME.String() {
					clusterName = env.Value
					break
				}
			}
		}
		message := PodDefaultMessage
		if pod.Status.Message != "" {
			message = pod.Status.Message
		}

		podRestartCount := int32(0)
		containersResource := make([]apistructs.PodContainer, 0)
		podHealthy := PodStatusHealthy
		switch pod.Status.Phase {
		case corev1.PodPending:
			podHealthy = PodStatusCreating
			containerResource := apistructs.ContainerResource{}
			nameToIndex := make(map[string]int)
			for idx, container := range pod.Spec.Containers {
				requestmem, _ := container.Resources.Requests.Memory().AsInt64()
				limitmem, _ := container.Resources.Limits.Memory().AsInt64()
				containerResource = apistructs.ContainerResource{
					MemRequest: int(requestmem / 1024 / 1024),
					MemLimit:   int(limitmem / 1024 / 1024),
					CpuRequest: (container.Resources.Requests.Cpu().AsApproximateFloat64() * 1000) / 1000,
					CpuLimit:   (container.Resources.Limits.Cpu().AsApproximateFloat64() * 1000) / 1000,
				}
				containersResource = append(containersResource, apistructs.PodContainer{
					ContainerName: container.Name,
					Image:         container.Image,
					Resource:      containerResource,
				})
				nameToIndex[container.Name] = idx
			}
			message = PodPendingDefaultMessage
			if len(pod.Status.ContainerStatuses) > 0 {
				for _, podContainerStatus := range pod.Status.ContainerStatuses {
					if podContainerStatus.Ready != true {
						message = getPodContainerMessage(podContainerStatus.State)
						if _, ok := nameToIndex[podContainerStatus.Name]; ok {
							containersResource[nameToIndex[podContainerStatus.Name]].Message = message
						} else {
							containersResource[0].Message = message
						}
					}
				}
			} else {
				for _, podCondition := range pod.Status.Conditions {
					if string(podCondition.Status) == PodConditionFalse {
						message = convertReasonMessageToHumanReadableMessage(podCondition.Reason, podCondition.Message)
						containersResource[0].Message = message
						break
					}
				}
			}

		default:
			for _, podCondition := range pod.Status.Conditions {
				if podCondition.Status != corev1.ConditionTrue {
					podHealthy = PodStatusUnHealthy
					break
				}
			}

			for _, podContainerStatus := range pod.Status.ContainerStatuses {
				if podContainerStatus.RestartCount > podRestartCount {
					podRestartCount = podContainerStatus.RestartCount
				}
				containerID := ""
				if len(strings.Split(podContainerStatus.ContainerID, "://")) <= 1 {
					return currPods, errors.Errorf("Pod status containerStatuses no containerID, neew pod for service %s is creating, please wait", serviceName)
				} else {
					containerID = strings.Split(podContainerStatus.ContainerID, "://")[1]
				}

				if podContainerStatus.Ready != true {
					podHealthy = PodStatusUnHealthy
					message = getPodContainerMessage(podContainerStatus.State)
				}

				containerResource := apistructs.ContainerResource{}
				for _, container := range pod.Spec.Containers {
					if container.Name == podContainerStatus.Name {
						requestmem, _ := container.Resources.Requests.Memory().AsInt64()
						limitmem, _ := container.Resources.Limits.Memory().AsInt64()
						containerResource = apistructs.ContainerResource{
							MemRequest: int(requestmem / 1024 / 1024),
							MemLimit:   int(limitmem / 1024 / 1024),
							CpuRequest: (container.Resources.Requests.Cpu().AsApproximateFloat64() * 1000) / 1000,
							CpuLimit:   (container.Resources.Limits.Cpu().AsApproximateFloat64() * 1000) / 1000,
						}
						break
					}
				}

				containersResource = append(containersResource, apistructs.PodContainer{
					ContainerID:   containerID,
					ContainerName: podContainerStatus.Name,
					Image:         podContainerStatus.Image,
					Resource:      containerResource,
					Message:       message,
				})
			}

			containersResource = sortContainers(containersResource,
				getMainContainerNameByServiceName(pod.Spec.Containers, serviceName),
			)

			if pod.UID == "" || pod.Status.PodIP == "" || pod.Status.HostIP == "" ||
				pod.Status.Phase == "" || pod.Status.StartTime == nil || len(containersResource) == 0 {
				return currPods, errors.Errorf("Pod status have no enough info for pod, pod for service %s is creating, please wait", serviceName)
			}
		}

		if pod.Status.StartTime == nil {
			pod.Status.StartTime = &pod.CreationTimestamp
		}

		var updateTime *metav1.Time
		for _, v := range pod.Status.Conditions {
			if v.Type == corev1.PodReady {
				lt := v.LastTransitionTime
				updateTime = &lt
			}
		}
		if updateTime == nil {
			updateTime = pod.Status.StartTime
		}

		currPods = append(currPods, apistructs.Pod{
			Uid:           string(pod.UID),
			IPAddress:     pod.Status.PodIP,
			Host:          pod.Status.HostIP,
			Phase:         podHealthy,
			Message:       containersResource[0].Message,
			StartedAt:     pod.Status.StartTime.Format(time.RFC3339Nano),
			UpdatedAt:     updateTime.Format(time.RFC3339Nano),
			Service:       serviceName,
			ClusterName:   clusterName,
			PodName:       pod.Name,
			K8sNamespace:  pod.Namespace,
			RestartCount:  podRestartCount,
			PodContainers: containersResource,
		})
	}

	if len(currPods) == 0 {
		return currPods, errors.Errorf("No pods get for service %s, pod may be is creating, please wait", serviceName)
	}
	return currPods, nil
}

func getPodContainerMessage(containerState corev1.ContainerState) string {
	message := ""
	switch {
	case containerState.Waiting != nil:
		message = convertReasonMessageToHumanReadableMessage(containerState.Waiting.Reason, containerState.Waiting.Message)
	case containerState.Terminated != nil:
		message = convertReasonMessageToHumanReadableMessage(containerState.Terminated.Reason, containerState.Terminated.Message)
	}
	return message
}

func convertReasonMessageToHumanReadableMessage(containerReason, containerMessage string) string {

	extractMessage := ""
	switch {
	case strings.Contains(containerMessage, messageContainsImageNotFound), strings.Contains(containerMessage, messageContainsImagePullBackoff):
		extractMessage = messageContainsImagePullFailedTrans
	case strings.Contains(containerMessage, messageContainsNodeAffinity):
		extractMessage = messageContainsNodeAffinity
	case strings.Contains(containerMessage, messageContainsInsufficientCPU), strings.Contains(containerMessage, messageContainsInsufficientMemory):
		if strings.Contains(containerMessage, messageContainsInsufficientCPU) {
			extractMessage = messageContainsInsufficientCPU
		}
		if strings.Contains(containerMessage, messageContainsInsufficientMemory) {
			extractMessage = messageContainsInsufficientMemory
		}
		if strings.Contains(containerMessage, messageContainsInsufficientCPU) && strings.Contains(containerMessage, messageContainsInsufficientMemory) {
			extractMessage = messageContainsInsufficientCPUAndMemory
		}
	case strings.Contains(containerMessage, messageContainsNoSuchFileOrDirectory):
		extractMessage = messageContainsNoSuchFileOrDirectoryTrans
	case strings.Contains(containerMessage, messageContainsBackOff) || strings.Contains(containerMessage, messageContainsBackOffRestartConatiner):
		extractMessage = messageContainsBackOffRestartConatinerTrans
	default:
		extractMessage = containerMessage
	}

	switch containerReason {
	case ContainerStateReasonImagePullBackOff, PodUnHealthyReasonErrImagePull:
		return fmt.Sprintf("%s: %s", PodUnHealthyReasonImagePullBackOff, extractMessage)
	case PodPendingReasonUnSchedulable:
		return fmt.Sprintf("%s: %s", PodUnHealthyReasonUnSchedulable, extractMessage)
	case ContainerStateReasonRunContainerError:
		return fmt.Sprintf("%s: %s", PodUnHealthyReasonRunContainerError, extractMessage)
	case ContainerStateReasonCrashLoopBackOff:
		return fmt.Sprintf("%s: %s", PodUnHealthyReasonCrashLoopBackOff, extractMessage)
	default:
		return fmt.Sprintf("%s: %s", containerReason, extractMessage)
	}
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

// sortContainers
// TODO: move all logic to the implementation.
func sortContainers(containers []apistructs.PodContainer, mainServiceName string) []apistructs.PodContainer {
	var sortedContainers []apistructs.PodContainer

	if len(containers) < 2 {
		return containers
	}

	for _, container := range containers {
		if container.ContainerName == mainServiceName {
			sortedContainers = append([]apistructs.PodContainer{container}, sortedContainers...)
		} else {
			sortedContainers = append(sortedContainers, container)
		}
	}

	return sortedContainers
}

// getMainContainerNameByServiceName
// TODO: move all logic to the implementation.
func getMainContainerNameByServiceName(containers []corev1.Container, serviceName string) string {
	if serviceName == "" || len(containers) == 1 {
		return containers[0].Name
	}

	for _, c := range containers {
		for _, e := range c.Env {
			if e.Name == apistructs.EnvDiceServiceName && e.Value == serviceName {
				return c.Name
			}
		}
	}

	return containers[0].Name
}
