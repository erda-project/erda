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

package ws

import (
	"github.com/erda-project/erda/apistructs"
)

const (
	R_DEPLOY_STATUS_UPDATE           = "R_DEPLOY_STATUS_UPDATE"
	R_RUNTIME_STATUS_CHANGED         = "R_RUNTIME_STATUS_CHANGED"
	R_RUNTIME_SERVICE_STATUS_CHANGED = "R_RUNTIME_SERVICE_STATUS_CHANGED"
	R_RUNTIME_DELETING               = "R_RUNTIME_DELETING"
	R_RUNTIME_DELETED                = "R_RUNTIME_DELETED"
)

type DeployStatusUpdatePayload struct {
	DeploymentId uint64                      `json:"deploymentId"`
	RuntimeId    uint64                      `json:"runtimeId"`
	Status       apistructs.DeploymentStatus `json:"status"`
	Phase        apistructs.DeploymentPhase  `json:"phase"`
	Step         apistructs.DeploymentPhase  `json:"step"` // Deprecated
	Extra        map[string]interface{}      `json:"extra"`
}

type RuntimeStatusChangedPayload struct {
	RuntimeId uint64                     `json:"runtimeId"`
	Status    string                     `json:"status"`
	Errors    []apistructs.ErrorResponse `json:"errors"`
}

type RuntimeServiceStatusChangedPayload struct {
	RuntimeId   uint64                     `json:"runtimeId"`
	ServiceName string                     `json:"serviceName"`
	Status      string                     `json:"status"`
	Errors      []apistructs.ErrorResponse `json:"errors"`
}

type RuntimeDeletingPayload struct {
	RuntimeId uint64 `json:"runtimeId"`
}

type RuntimeDeletedPayload struct {
	RuntimeId uint64 `json:"runtimeId"`
}
