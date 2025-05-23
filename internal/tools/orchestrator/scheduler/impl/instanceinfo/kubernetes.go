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

package instanceinfo

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
)

// Pod status and message constants
const (
	PodStatusHealthy    = "Healthy"
	PodStatusUnHealthy  = "Unhealthy"
	PodStatusCreating   = "Creating"
	PodStatusTerminated = "Terminated"

	PodConditionFalse        = "False"
	PodDefaultMessage        = "Ok"
	PodPendingDefaultMessage = "Waiting for scheduling"

	// Reason constants for container and pod status
	ReasonUnschedulable         = "Unschedulable"
	ReasonErrImagePull          = "ErrImagePull"
	ReasonImagePullBackOff      = "ImagePullBackOff"
	ReasonRunContainerError     = "RunContainerError"
	ReasonCrashLoopBackOff      = "CrashLoopBackOff"
	ReasonNodeAffinity          = "didn't match Pod's node affinity"
	ReasonInsufficientCPU       = "Insufficient cpu"
	ReasonInsufficientMemory    = "Insufficient memory"
	ReasonNoSuchFileOrDirectory = "no such file or directory"
	ReasonBackOff               = "back-off"
	ReasonRestartingContainer   = "restarting failed container"
)

const (
	MsgImageNotFound                = "not found"
	MsgImagePullBackoff             = "Back-off pulling image"
	MsgImagePullFailedTrans         = "image not found or can not pull"
	MsgNodeAffinity                 = ReasonNodeAffinity
	MsgInsufficientCPU              = ReasonInsufficientCPU
	MsgInsufficientMemory           = ReasonInsufficientMemory
	MsgInsufficientCPUAndMemory     = "Insufficient memory and memory"
	MsgNoSuchFileOrDirectory        = ReasonNoSuchFileOrDirectory
	MsgNoSuchFileOrDirectoryTrans   = "some file or directory not found"
	MsgBackOff                      = ReasonBackOff
	MsgBackOffRestartContainer      = ReasonRestartingContainer
	MsgBackOffRestartContainerTrans = "container process exit"
)

type FilterContainerResult struct {
	ClusterName      string
	FilterContainers []corev1.Container
}

// GetInstancePodFromK8s retrieves the status of pods for a given service in a ServiceGroup.
func (i *InstanceInfoImpl) GetInstancePodFromK8s(sg *apistructs.ServiceGroup, serviceName string) ([]apistructs.Pod, error) {
	k8sPods, err := extractK8sPodsFromServiceGroup(sg, serviceName)
	if err != nil {
		return nil, err
	}
	pods := make([]apistructs.Pod, 0, len(k8sPods))
	for _, pod := range k8sPods {
		if !(pod.Status.Phase == corev1.PodPending || pod.Status.Phase == corev1.PodRunning) {
			logrus.WithFields(logrus.Fields{
				"namespace": pod.Namespace,
				"pod":       pod.Name,
			}).Warn("Pod is in terminated or unknown status, ignoring.")
			continue
		}
		podInfo, err := buildPodInfo(pod, serviceName)
		if err != nil {
			return nil, err
		}
		pods = append(pods, podInfo)
	}
	if len(pods) == 0 {
		return nil, errors.Errorf("No pods found for service %s, pod may be creating, please wait", serviceName)
	}
	return pods, nil
}

// extractK8sPodsFromServiceGroup parses the k8sPods from ServiceGroup.Extra.
// TODO: optimize pod raw in context
func extractK8sPodsFromServiceGroup(sg *apistructs.ServiceGroup, serviceName string) ([]corev1.Pod, error) {
	if sg == nil || sg.Extra == nil {
		return nil, errors.New("service group or extra is nil")
	}
	data, ok := sg.Extra[serviceName]
	if !ok {
		return nil, errors.Errorf("get service %s current pods failed: no pods found in sg.Extra for service", serviceName)
	}
	var k8sPods []corev1.Pod
	if err := json.Unmarshal([]byte(data), &k8sPods); err != nil {
		return nil, errors.Errorf("get service %s current pods failed: %v", serviceName, err)
	}
	return k8sPods, nil
}

// filterContainers
// Container not deployed by erda: containers not contains env: DICE_CLUSTER_NAME
func filterContainers(containers []corev1.Container) *FilterContainerResult {
	var (
		clusterName string
		filtered    []corev1.Container
	)

	for _, c := range containers {
		for _, env := range c.Env {
			if env.Name == apistructs.DICE_CLUSTER_NAME.String() {
				filtered = append(filtered, c)
				clusterName = env.Value
				break
			}
		}
	}

	return &FilterContainerResult{
		ClusterName:      clusterName,
		FilterContainers: filtered,
	}
}

// buildPodInfo converts a Kubernetes pod to a business Pod object.
func buildPodInfo(pod corev1.Pod, serviceName string) (apistructs.Pod, error) {
	// containers filter
	fr := filterContainers(pod.Spec.Containers)
	pod.Spec.Containers = fr.FilterContainers
	clusterName := fr.ClusterName

	var (
		podHealthy         = PodStatusHealthy
		podRestartCount    = int32(0)
		containersResource []apistructs.PodContainer
		message            = getPodDefaultMessage(pod)
	)

	switch pod.Status.Phase {
	case corev1.PodPending:
		podHealthy = PodStatusCreating
		containersResource = buildContainersResourcePending(pod, &message)
	case corev1.PodRunning:
		var err error
		podHealthy, containersResource, podRestartCount, message, err = buildContainersResourceRunning(pod, serviceName)
		if err != nil {
			return apistructs.Pod{}, err
		}
	}

	if pod.Status.StartTime == nil {
		pod.Status.StartTime = &pod.CreationTimestamp
	}
	updateTime := getPodUpdateTime(pod)

	return apistructs.Pod{
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
	}, nil
}

// getPodDefaultMessage returns the default message for a pod.
func getPodDefaultMessage(pod corev1.Pod) string {
	if pod.Status.Message != "" {
		return pod.Status.Message
	}
	return PodDefaultMessage
}

// getPodUpdateTime returns the update time for a pod.
func getPodUpdateTime(pod corev1.Pod) *metav1.Time {
	for _, v := range pod.Status.Conditions {
		if v.Type == corev1.PodReady {
			return &v.LastTransitionTime
		}
	}
	if pod.Status.StartTime != nil {
		return pod.Status.StartTime
	}
	return &pod.CreationTimestamp
}

// buildContainersResourcePending builds container resources for a pod in Pending phase.
func buildContainersResourcePending(pod corev1.Pod, message *string) []apistructs.PodContainer {
	var containersResource []apistructs.PodContainer
	nameToIndex := make(map[string]int)
	for _, container := range pod.Spec.Containers {
		idx := len(containersResource)
		containersResource = append(containersResource, apistructs.PodContainer{
			ContainerName: container.Name,
			Image:         container.Image,
			Resource:      buildContainerResource(container),
		})
		nameToIndex[container.Name] = idx
	}
	*message = PodPendingDefaultMessage
	if len(pod.Status.ContainerStatuses) > 0 {
		for _, podContainerStatus := range pod.Status.ContainerStatuses {
			if !podContainerStatus.Ready {
				msg := getPodContainerMessage(podContainerStatus.State)
				if idx, ok := nameToIndex[podContainerStatus.Name]; ok {
					containersResource[idx].Message = msg
				} else if len(containersResource) > 0 {
					containersResource[0].Message = msg
				}
			}
		}
	} else {
		for _, podCondition := range pod.Status.Conditions {
			if string(podCondition.Status) == PodConditionFalse {
				msg := convertReasonMessageToHumanReadableMessage(podCondition.Reason, podCondition.Message)
				if len(containersResource) > 0 {
					containersResource[0].Message = msg
				}
				*message = msg
				break
			}
		}
	}
	return containersResource
}

// buildContainersResourceRunning builds container resources for a pod in Running phase.
func buildContainersResourceRunning(pod corev1.Pod, serviceName string) (string, []apistructs.PodContainer, int32, string, error) {
	podHealthy := PodStatusHealthy
	podRestartCount := int32(0)
	containersResource := make([]apistructs.PodContainer, 0, len(pod.Status.ContainerStatuses))
	message := PodDefaultMessage

	for _, podCondition := range pod.Status.Conditions {
		if podCondition.Status != corev1.ConditionTrue {
			podHealthy = PodStatusUnHealthy
			break
		}
	}

	for _, podContainerStatus := range pod.Status.ContainerStatuses {
		// Only process containers that exist in pod.Spec.Containers (already filtered)
		var matchedContainer *corev1.Container
		for _, container := range pod.Spec.Containers {
			if container.Name == podContainerStatus.Name {
				matchedContainer = &container
				break
			}
		}
		if matchedContainer == nil {
			continue
		}
		if podContainerStatus.RestartCount > podRestartCount {
			podRestartCount = podContainerStatus.RestartCount
		}
		containerID := extractContainerID(podContainerStatus.ContainerID)
		if containerID == "" {
			return podHealthy, nil, 0, message, errors.Errorf("Pod status does not have enough info for pod, pod for service %s is creating, please wait", serviceName)
		}
		if !podContainerStatus.Ready {
			podHealthy = PodStatusUnHealthy
			message = getPodContainerMessage(podContainerStatus.State)
		}
		containerResource := buildContainerResource(*matchedContainer)
		containersResource = append(containersResource, apistructs.PodContainer{
			ContainerID:   containerID,
			ContainerName: podContainerStatus.Name,
			Image:         podContainerStatus.Image,
			Resource:      containerResource,
			Message:       message,
		})
	}

	containersResource = sortContainers(containersResource, getMainContainerNameByServiceName(pod.Spec.Containers, serviceName))

	if pod.UID == "" || pod.Status.PodIP == "" || pod.Status.HostIP == "" ||
		pod.Status.Phase == "" || pod.Status.StartTime == nil || len(containersResource) == 0 {
		return podHealthy, nil, 0, message, errors.Errorf("Pod status does not have enough info for pod, pod for service %s is creating, please wait", serviceName)
	}
	return podHealthy, containersResource, podRestartCount, message, nil
}

// buildContainerResource builds a business ContainerResource from a k8s container
func buildContainerResource(container corev1.Container) apistructs.ContainerResource {
	requester, _ := container.Resources.Requests.Memory().AsInt64()
	limit, _ := container.Resources.Limits.Memory().AsInt64()
	return apistructs.ContainerResource{
		MemRequest: int(requester / 1024 / 1024),
		MemLimit:   int(limit / 1024 / 1024),
		CpuRequest: (container.Resources.Requests.Cpu().AsApproximateFloat64() * 1000) / 1000,
		CpuLimit:   (container.Resources.Limits.Cpu().AsApproximateFloat64() * 1000) / 1000,
	}
}

// extractContainerID extracts the container ID from a k8s containerID string.
func extractContainerID(containerID string) string {
	parts := strings.Split(containerID, "://")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// getPodContainerMessage returns a human-readable message for a container state.
func getPodContainerMessage(containerState corev1.ContainerState) string {
	switch {
	case containerState.Waiting != nil:
		return convertReasonMessageToHumanReadableMessage(containerState.Waiting.Reason, containerState.Waiting.Message)
	case containerState.Terminated != nil:
		return convertReasonMessageToHumanReadableMessage(containerState.Terminated.Reason, containerState.Terminated.Message)
	default:
		return ""
	}
}

// convertReasonMessageToHumanReadableMessage converts k8s reason/message to a human-readable message.
func convertReasonMessageToHumanReadableMessage(containerReason, containerMessage string) string {
	extractMessage := ""
	switch {
	case strings.Contains(containerMessage, MsgImageNotFound), strings.Contains(containerMessage, MsgImagePullBackoff):
		extractMessage = MsgImagePullFailedTrans
	case strings.Contains(containerMessage, MsgNodeAffinity):
		extractMessage = MsgNodeAffinity
	case strings.Contains(containerMessage, MsgInsufficientCPU), strings.Contains(containerMessage, MsgInsufficientMemory):
		if strings.Contains(containerMessage, MsgInsufficientCPU) && strings.Contains(containerMessage, MsgInsufficientMemory) {
			extractMessage = MsgInsufficientCPUAndMemory
		} else if strings.Contains(containerMessage, MsgInsufficientCPU) {
			extractMessage = MsgInsufficientCPU
		} else {
			extractMessage = MsgInsufficientMemory
		}
	case strings.Contains(containerMessage, MsgNoSuchFileOrDirectory):
		extractMessage = MsgNoSuchFileOrDirectoryTrans
	case strings.Contains(containerMessage, MsgBackOff) || strings.Contains(containerMessage, MsgBackOffRestartContainer):
		extractMessage = MsgBackOffRestartContainerTrans
	default:
		extractMessage = containerMessage
	}

	switch containerReason {
	case ReasonImagePullBackOff, ReasonErrImagePull:
		return fmt.Sprintf("%s: %s", "Failed to pull image", extractMessage)
	case ReasonUnschedulable:
		return fmt.Sprintf("%s: %s", "Failed to Schedule Pod", extractMessage)
	case ReasonRunContainerError:
		return fmt.Sprintf("%s: %s", "Failed to start container", extractMessage)
	case ReasonCrashLoopBackOff:
		return fmt.Sprintf("%s: %s", "Failed to run command", extractMessage)
	default:
		return fmt.Sprintf("%s: %s", containerReason, extractMessage)
	}
}

// sortContainers puts the main container first in the list.
func sortContainers(containers []apistructs.PodContainer, mainServiceName string) []apistructs.PodContainer {
	if len(containers) < 2 {
		return containers
	}
	var sorted []apistructs.PodContainer
	for _, c := range containers {
		if c.ContainerName == mainServiceName {
			sorted = append([]apistructs.PodContainer{c}, sorted...)
		} else {
			sorted = append(sorted, c)
		}
	}
	return sorted
}

// getMainContainerNameByServiceName returns the main container name for a service.
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
