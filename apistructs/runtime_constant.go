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

package apistructs

type RuntimeSource string

const (
	PIPELINE     RuntimeSource = "PIPELINE"
	ABILITY      RuntimeSource = "ABILITY"
	RUNTIMEADDON RuntimeSource = "RUNTIMEADDON"
	RELEASE      RuntimeSource = "RELEASE"
)

const (
	RuntimeStatusHealthy     = "Healthy"
	RuntimeStatusUnHealthy   = "UnHealthy"
	RuntimeStatusProgressing = "Progressing"
	RuntimeStatusInit        = "Init"
	RuntimeStatusUnknown     = "Unknown" // It should be not exist

	ServiceStatusHealthy   = "Healthy"   // 运行中，预期实例数与实际实例数相等，且都通过健康检查
	ServiceStatusUnHealthy = "UnHealthy" // 预期实例数与实际实例数不相等，或者至少一个副本的健康检查未收到或未通过
	ServiceStatusUnknown   = "Unknown"

	RuntimeEventTypeTotal = "total"
)

// Flow:
//
// WAITAPPROVE(optional) -> I/W -> DEPLOYING -> OK
//                           |          `---> FAILED
//                           |          `---> CANCELING -> CANCELED
//                           |                       `---> FAILED
//                           `-> CANCELED
type DeploymentStatus string

const (
	DeploymentStatusWaitApprove DeploymentStatus = "WAITAPPROVE"
	DeploymentStatusInit        DeploymentStatus = "INIT"
	DeploymentStatusWaiting     DeploymentStatus = "WAITING"
	DeploymentStatusDeploying   DeploymentStatus = "DEPLOYING"
	DeploymentStatusOK          DeploymentStatus = "OK"
	DeploymentStatusFailed      DeploymentStatus = "FAILED"
	DeploymentStatusCanceling   DeploymentStatus = "CANCELING"
	DeploymentStatusCanceled    DeploymentStatus = "CANCELED"
)

type DeploymentPhase string

const (
	DeploymentPhaseInit      DeploymentPhase = "INIT"
	DeploymentPhaseAddon     DeploymentPhase = "ADDON_REQUESTING"
	DeploymentPhaseScript    DeploymentPhase = "SCRIPT_APPLYING"
	DeploymentPhaseService   DeploymentPhase = "SERVICE_DEPLOYING"
	DeploymentPhaseRegister  DeploymentPhase = "DISCOVERY_REGISTER"
	DeploymentPhaseCompleted DeploymentPhase = "COMPLETED"
)
