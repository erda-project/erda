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

package apistructs

import "time"

type RuntimeInspectDTO struct {
	ID uint64 `json:"id"`
	// runtime名称
	Name                  string        `json:"name"`
	ServiceGroupName      string        `json:"serviceGroupName"`
	ServiceGroupNamespace string        `json:"serviceGroupNamespace"`
	Source                RuntimeSource `json:"source"`
	// 状态
	Status       string                               `json:"status"`
	DeployStatus DeploymentStatus                     `json:"deployStatus"`
	DeleteStatus string                               `json:"deleteStatus"`
	ReleaseID    string                               `json:"releaseId"`
	ClusterID    uint64                               `json:"clusterId"`
	ClusterName  string                               `json:"clusterName"`
	ClusterType  string                               `json:"clusterType"`
	Resources    RuntimeServiceResourceDTO            `json:"resources"`
	Extra        map[string]interface{}               `json:"extra"` // TODO: move fields out of extra
	ProjectID    uint64                               `json:"projectID"`
	Services     map[string]*RuntimeInspectServiceDTO `json:"services"`
	// 模块发布错误信息
	ModuleErrMsg map[string]map[string]string `json:"lastMessage"`
	TimeCreated  time.Time                    `json:"timeCreated"` // Deprecated: use CreatedAt instead
	CreatedAt    time.Time                    `json:"createdAt"`
	UpdatedAt    time.Time                    `json:"updatedAt"`
	Errors       []ErrorResponse              `json:"errors"`
}

type RuntimeInspectServiceDTO struct {
	Status      string                       `json:"status"`
	Deployments RuntimeServiceDeploymentsDTO `json:"deployments"`
	Resources   RuntimeServiceResourceDTO    `json:"resources"`
	Envs        map[string]string            `json:"envs"`
	Addrs       []string                     `json:"addrs"` // TODO: better name?
	Expose      []string                     `json:"expose"`
	Errors      []ErrorResponse              `json:"errors"`
}

type RuntimeSummaryDTO struct {
	RuntimeInspectDTO
	LastOperator       string    `json:"lastOperator"`
	LastOperatorName   string    `json:"lastOperatorName"`   // Deprecated
	LastOperatorAvatar string    `json:"lastOperatorAvatar"` // Deprecated
	LastOperateTime    time.Time `json:"lastOperateTime"`
}

type RuntimeDTO struct {
	ID              uint64          `json:"id"`
	Name            string          `json:"name"`
	GitBranch       string          `json:"gitBranch"` // Deprecated: use name instead
	Workspace       string          `json:"workspace"`
	ClusterName     string          `json:"clusterName"`
	ClusterId       uint64          `json:"clusterId"` // Deprecated: use ClusterName instead
	Status          string          `json:"status"`
	ApplicationID   uint64          `json:"applicationId"`
	ApplicationName string          `json:"applicationName"`
	ProjectID       uint64          `json:"projectId"`
	ProjectName     string          `json:"projectName"`
	OrgID           uint64          `json:"orgId"`
	Errors          []ErrorResponse `json:"errors"`
}

// TODO: currently same as RuntimeInspectServiceDTO, we should combine these two
type RuntimeServiceDTO struct {
	ID          uint64                       `json:"id"`
	RuntimeID   uint64                       `json:"runtimeId"`
	ServiceName string                       `json:"serviceName"`
	Status      string                       `json:"status"`
	Deployments RuntimeServiceDeploymentsDTO `json:"deployments"`
	Resources   RuntimeServiceResourceDTO    `json:"resources"`
	Envs        map[string]string            `json:"envs"`
	Expose      []string                     `json:"expose"`
	Errors      []ErrorResponse              `json:"errors"`
}

type RuntimeServiceDeploymentsDTO struct {
	Replicas int `json:"replicas"`
}

type RuntimeServiceResourceDTO struct {
	CPU  float64 `json:"cpu"`
	Mem  int     `json:"mem"`
	Disk int     `json:"disk"`
}

// TODO: same as spec.RuntimeInstance, need to combine two
type RuntimeInstanceDTO struct {
	ID          uint64    `json:"id"`
	InstanceID  string    `json:"instanceId"`
	RuntimeID   uint64    `json:"runtimeId"`
	ServiceName string    `json:"serviceName"`
	IP          string    `json:"ip"`
	Status      string    `json:"status"`
	Stage       string    `json:"stage"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type DeploymentStatusDTO struct {
	// 发布Id
	DeploymentID uint64 `json:"deploymentId"`
	// 状态
	Status DeploymentStatus `json:"status"`
	// 发布过程
	Phase DeploymentPhase `json:"phase"`
	// 失败原因
	FailCause string `json:"failCause"`
	// 模块错误信息
	ModuleErrMsg map[string]string           `json:"lastMessage"`
	Runtime      *DeploymentStatusRuntimeDTO `json:"runtime"`
}

// Deprecated: use RuntimeInspect api to get ServiceGroup Info
type DeploymentStatusRuntimeDTO struct {
	Services  map[string]*DeploymentStatusRuntimeServiceDTO `json:"services"`
	Endpoints map[string]*DeploymentStatusRuntimeServiceDTO `json:"endpoints"`
}

// Deprecated: use RuntimeInspect api to get ServiceGroup Info
type DeploymentStatusRuntimeServiceDTO struct {
	PublicHosts []string `json:"publicHosts"`
	Host        string   `json:"host"`
	Ports       []int    `json:"ports"`
}

type PreDiceDTO struct {
	Name     string                               `json:"name,omitempty"`
	Envs     map[string]string                    `json:"envs,omitempty"`
	Services map[string]*RuntimeInspectServiceDTO `json:"services,omitempty"`
}

type DeploymentCreateResponseDTO struct {
	DeploymentID  uint64 `json:"deploymentId"`
	ApplicationID uint64 `json:"applicationId"`
	RuntimeID     uint64 `json:"runtimeId"`
}

type DeploymentCreateResponsePipelineDTO struct {
	PipelineID uint64 `json:"pipelineId"`
}
