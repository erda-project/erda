// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package orgapis

var indexHostSummary = []string{"spot-machine_summary-full_cluster", "spot-machine_summary-full_cluster.*"}

const (
	orgPrefix = "org-"
	offline   = "offline"
	any       = "*"
	topHits   = "top_hits"

	nameContainerSummary = "container_summary"

	timestamp       = "timestamp"
	id              = "id"
	labels          = "labels"
	os              = "os"
	kernelVersion   = "kernel_version"
	image           = "image"
	terminusVersion = "terminus_version"
	isDeleted       = "is_deleted"

	cpuUsagePercent  = "cpu_usage_percent"
	memUsagePercent  = "mem_usage_percent"
	diskUsagePercent = "disk_usage_percent"
	cpuDispPercent   = "cpu_disp_percent"
	memDispPercent   = "mem_disp_percent"
	loadPercent      = "load_percent"

	cpuUsage   = "cpu_usage"
	cpuRequest = "cpu_request"
	cpuLimit   = "cpu_limit"
	cpuOrigin  = "cpu_origin"
	cpuTotal   = "cpu_total"
	memUsage   = "mem_usage"
	memRequest = "mem_request"
	memLimit   = "mem_limit"
	memOrigin  = "mem_origin"
	memTotal   = "mem_total"
	diskUsage  = "disk_usage"
	diskLimit  = "disk_limit"
	diskTotal  = "disk_total"

	load1  = "load1"
	load5  = "load5"
	load15 = "load15"
	tasks  = "tasks"

	cluster               = "cluster"
	clusterName           = "cluster_name"
	hostIP                = "host_ip"
	cpus                  = "cpus"
	mem                   = "mem"
	host                  = "host"
	containerID           = "container_id"
	instanceID            = "instance_id"
	instanceType          = "instance_type"
	orgID                 = "org_id"
	orgName               = "org_name"
	projectID             = "project_id"
	projectName           = "project_name"
	applicationID         = "application_id"
	applicationName       = "application_name"
	runtimeID             = "runtime_id"
	runtimeName           = "runtime_name"
	workspace             = "workspace"
	serviceID             = "service_id"
	serviceName           = "service_name"
	jobID                 = "job_id"
	addonID               = "addon_id"
	componentName         = "component_name"
	instanceTypeAll       = "all"
	instanceTypeJob       = "job"
	instanceTypeService   = "service"
	instanceTypeComponent = "component"
	instanceTypeAddon     = "addon"
	podName               = "pod_name"
	fields                = "fields"
	fieldsPrefix          = fields + "."
	fieldsLabels          = fieldsPrefix + labels
	tags                  = "tags"
	tagsPrefix            = tags + "."
	tagsClusterName       = tagsPrefix + clusterName
	tagsLabels            = tagsPrefix + labels
	tagsHostIP            = tagsPrefix + hostIP
	tagsCPUs              = tagsPrefix + cpus
	tagsMem               = tagsPrefix + mem
	tagsContainerID       = tagsPrefix + containerID
	tagsInstanceType      = tagsPrefix + instanceType
	tagsOrgName           = tagsPrefix + orgName
	tagsProjectID         = tagsPrefix + projectID
	tagsProjectName       = tagsPrefix + projectName
	tagsApplicationID     = tagsPrefix + applicationID
	tagsApplicationName   = tagsPrefix + applicationName
	tagsWorkspace         = tagsPrefix + workspace
	tagsRuntimeID         = tagsPrefix + runtimeID
	tagsRuntimeName       = tagsPrefix + runtimeName
	tagsServiceID         = tagsPrefix + serviceID
	tagsServiceName       = tagsPrefix + serviceName
	tagsAddonID           = tagsPrefix + addonID
	tagsJobID             = tagsPrefix + jobID
	tagsComponentName     = tagsPrefix + componentName
	tagsTerminusVersion   = tagsPrefix + terminusVersion
	tagsIsDeleted         = tagsPrefix + isDeleted
	tagsPodName           = tagsPrefix + podName
)

var (
	percents = []string{
		"0%-40%",
		"40%-70%",
		"70%-99%",
		">=99%",
	}
)
