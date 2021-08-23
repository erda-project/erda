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

package marathon

import (
	"github.com/erda-project/erda/apistructs"
)

type Group struct {
	Id     string  `json:"id"`
	Apps   []App   `json:"apps"`
	Groups []Group `json:"groups"`
}

type App struct {
	Id   string   `json:"id"`
	Cmd  string   `json:"cmd,omitempty"`
	Args []string `json:"args,omitempty"`
	User string   `json:"user,omitempty"`

	Instances int     `json:"instances"`
	Cpus      float64 `json:"cpus"`
	Mem       float64 `json:"mem"`
	Disk      float64 `json:"disk"`

	Container AppContainer `json:"container"`

	Dependencies []string `json:"dependencies"`

	Env map[string]string `json:"env,omitempty"`

	Executor              string       `json:"executor,omitempty"`
	AcceptedResourceRoles []string     `json:"acceptedResourceRoles,omitempty"`
	Constraints           []Constraint `json:"constraints,omitempty"`

	Uris    []string             `json:"uris,omitempty"`
	Fetch   []AppFetch           `json:"fetch,omitempty"`
	Secrets map[string]AppSecret `json:"secrets,omitempty"`

	// Since: 1.5
	Networks []AppNetwork `json:"networks,omitempty"`
	// Deprecated: >= 1.5
	IpAddress      *AppIpAddress       `json:"ipAddress,omitempty"`
	Ports          []int               `json:"ports"`
	RequirePorts   bool                `json:"requirePorts,omitempty"`
	PortDefinition []AppPortDefinition `json:"portDefinitions,omitempty"`

	UpgradeStrategy            *AppUpgradeStrategy `json:"upgradeStrategy,omitempty"`
	BackoffSeconds             int                 `json:"backoffSeconds"`
	BackoffFactor              float32             `json:"backoffFactor"`
	MaxLaunchDelaySeconds      int                 `json:"maxLaunchDelaySeconds"`
	TaskKillGracePeriodSeconds int                 `json:"taskKillGracePeriodSeconds,omitempty"`

	HealthChecks    []AppHealthCheck    `json:"healthChecks"`
	ReadinessChecks []AppReadinessCheck `json:"readinessChecks,omitempty"`
	Labels          map[string]string   `json:"labels"`

	Tty bool `json:"tty,omitempty"`

	AppTasks
	AppCounts
	AppDeployments
}

type Constraint []string

// To get every instance info in one app
type AppTasks struct {
	Tasks []Task `json:"tasks,omitempty"`
}

type AppCounts struct {
	TasksStaged    int `json:"tasksStaged"`
	TasksRunning   int `json:"tasksRunning"`
	TasksHealthy   int `json:"tasksHealthy"`
	TasksUnhealthy int `json:"tasksUnhealthy"`
}

type AppDeployments struct {
	Deployments []AppDeployment `json:"deployments"`
}

type AppDeployment struct {
	Id string `json:"id"`
}

type AppNetwork struct {
	Name string `json:"name,omitempty"`
	Mode string `json:"mode,omitempty"`
}

type AppContainer struct {
	Type   string             `json:"type,omitempty"`
	Docker AppContainerDocker `json:"docker,omitempty"`
	// Since: 1.5
	PortMappings []AppContainerPortMapping `json:"portMappings,omitempty"`
	Volumes      []AppContainerVolume      `json:"volumes,omitempty"`
}

type AppContainerDocker struct {
	ForcePullImage bool                          `json:"forcePullImage,omitempty"`
	Image          string                        `json:"image"`
	Parameters     []AppContainerDockerParameter `json:"parameters,omitempty"`
	Privileged     bool                          `json:"privileged,omitempty"`
	// Deprecated: >=1.5
	Network string `json:"network,omitempty"`
	// Deprecated: >= 1.5
	PortMappings []AppContainerPortMapping `json:"portMappings"`
}

type AppContainerDockerParameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type AppContainerPortMapping struct {
	Name          string            `json:"name,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Protocol      string            `json:"protocol,omitempty"`
	ContainerPort int               `json:"containerPort"`
	HostPort      int               `json:"hostPort,omitempty"`
	ServicePort   int               `json:"servicePort,omitempty"`
}

type AppContainerVolume struct {
	Mode          string `json:"mode,omitempty"`
	ContainerPath string `json:"containerPath,omitempty"`
	HostPath      string `json:"hostPath,omitempty"`
	// TODO: refactor it
	Persistent *apistructs.PersistentVolume `json:"persistent,omitempty"`
}

type AppHealthCheck struct {
	GracePeriodSeconds     int  `json:"gracePeriodSeconds"`
	IgnoreHttp1xx          bool `json:"ignoreHttp1xx,omitempty"`
	IntervalSeconds        int  `json:"intervalSeconds,omitempty"`
	MaxConsecutiveFailures int  `json:"maxConsecutiveFailures,omitempty"`
	TimeoutSeconds         int  `json:"timeoutSeconds,omitempty"`
	DelaySeconds           int  `json:"delaySeconds"`

	Protocol  string                 `json:"protocol,omitempty"`
	Path      string                 `json:"path,omitempty"`
	PortIndex int                    `json:"portIndex,omitempty"`
	Port      int                    `json:"port,omitempty"`
	Command   *AppHealthCheckCommand `json:"command,omitempty"`
}

type AppHealthCheckCommand struct {
	Value string `json:"value,omitempty"`
}

type AppReadinessCheck struct {
	Name                    string `json:"name,omitempty"`
	Protocol                string `json:"protocol,omitempty"`
	Path                    string `json:"path,omitempty"`
	PortName                string `json:"portName,omitempty"`
	IntervalSeconds         int    `json:"intervalSeconds,omitempty"`
	TimeoutSeconds          int    `json:"timeoutSeconds,omitempty"`
	HttpStatusCodesForReady []int  `json:"httpStatusCodesForReady,omitempty"`
	PreserveLastResponse    bool   `json:"preserveLastResponse,omitempty"`
}

type AppIpAddress struct {
	NetworkName string                `json:"networkName"`
	Discovery   AppIpAddressDiscovery `json:"discovery"`
	Groups      []string              `json:"groups"`
	Labels      map[string]string     `json:"labels"`
}

type AppIpAddressDiscovery struct {
	Ports []AppIpAddressDiscoveryPort `json:"ports"`
}

type AppIpAddressDiscoveryPort struct {
	Number   int    `json:"number,omitempty"`
	Name     int    `json:"name,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

type AppPortDefinition struct {
	Port     int               `json:"port"`
	Protocol string            `json:"protocol,omitempty"`
	Name     string            `json:"name,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
}

type AppUpgradeStrategy struct {
	MaximumOverCapacity   float32 `json:"maximumOverCapacity,omitempty"`
	MinimumHealthCapacity float32 `json:"minimumHealthCapacity,omitempty"`
}

type AppFetch struct {
	Uri        string `json:"uri"`
	Executable bool   `json:"executable,omitempty"`
	Extract    bool   `json:"extract,omitempty"`
	Cache      bool   `json:"cache,omitempty"`
	DestPath   string `json:"destPath,omitempty"`
}

type AppSecret struct {
	Source string `json:"source"`
}

type GroupPutResponse struct {
	DeploymentId string                       `json:"deploymentId,omitempty"`
	Version      string                       `json:"version,omitempty"`
	Message      string                       `json:"message,omitempty"`
	Details      []GroupPutResponseDetail     `json:"details,omitempty"`
	Deployments  []GroupPutResponseDeployment `json:"deployments,omitempty"`
}

type GroupPutResponseDetail struct {
	Path   string   `json:"path,omitempty"`
	Errors []string `json:"errors,omitempty"`
}

type GroupPutResponseDeployment struct {
	Id string `json:"id,omitempty"`
}

type Queue struct {
	Queue []QueueOffer `json:"queue,omitempty"`
}

type QueueOffer struct {
	Count int             `json:"count,omitempty"`
	Delay QueueOfferDelay `json:"delay,omitempty"`
	App   App             `json:"app,omitempty"`
	// Overview of offer processing
	ProcessedOffersSummary ProcessedOffersSummary `json:"processedOffersSummary,omitempty"`
}

type QueueOfferDelay struct {
	TimeLeftSeconds int  `json:"timeLeftSeconds,omitempty"`
	Overdue         bool `json:"overdue,omitempty"`
}

// ProcessedOffersSummary Briefly describe whether the offer is in compliance
type ProcessedOffersSummary struct {
	RejectSummaryLastOffers []RejectSummaryLastOffer `json:"rejectSummaryLastOffers,omitempty"`
}

// RejectSummaryLastOffer Describe the failure of the recent offer
type RejectSummaryLastOffer struct {
	Reason    string `json:"reason,omitempty"`
	Declined  int    `json:"declined,omitempty"`
	Processed int    `json:"processed,omitempty"`
}

type GetErrorResponse struct {
	Message string `json:"message,omitempty"`
}

type AppStatus string

const (
	AppStatusRunning   AppStatus = "Running"
	AppStatusDeploying AppStatus = "Deploying"
	AppStatusSuspended AppStatus = "Suspended"
	AppStatusWaiting   AppStatus = "Waiting"
	AppStatusDelayed   AppStatus = "Delayed"
	AppStatusHealthy   AppStatus = "Healthy"
)

// wrap correct response and error
type GroupHTTPResult struct {
	Group
	GetErrorResponse
}

// wrap correct response and error
type QueueHTTPResult struct {
	Queue
	GetErrorResponse
}

type InstanceIpAddr struct {
	InstanceIp string `json:"ipAddress,omitempty"`
}

type HealthCheckResult struct {
	Alive bool `json:"alive,omitempty"`
}

type Task struct {
	Id                  string              `json:"id,omitempty"`
	State               string              `json:"state,omitempty"`
	InstanceIpAddresses []InstanceIpAddr    `json:"ipAddresses,omitempty"`
	Host                string              `json:"host,omitempty"`
	HealthCheckResults  []HealthCheckResult `json:"healthCheckResults,omitempty"`
}

type ShortApp struct {
	Instances int `json:"instances"`
}

type AppGet struct {
	App ShortApp `json:"app"`
}

type Deployments []Deployment

type Deployment struct {
	ID           string   `json:"id"`
	AffectedApps []string `json:"affectedApps"`
}

func (g *GroupPutResponse) ToString() string {
	var details string
	for _, d := range g.Details {
		var errors string
		for _, e := range d.Errors {
			if len(errors) == 0 {
				errors = e
			} else {
				errors = errors + "," + e
			}
		}

		if len(details) == 0 {
			details = "path: " + d.Path + ", details: " + errors
		} else {
			details = details + ";path: " + d.Path + ", details: " + errors
		}
	}
	return g.Message + "," + details
}

func (c Constraint) Equal(other Constraint) bool {
	if len(c) != len(other) {
		return false
	}
	for i := range c {
		if c[i] != other[i] {
			return false
		}
	}
	return true
}
