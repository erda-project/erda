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
