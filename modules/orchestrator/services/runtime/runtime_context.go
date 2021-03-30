package runtime

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
)

// DeployContext 部署上下文
type DeployContext struct {
	Runtime        *dbclient.Runtime
	App            *apistructs.ApplicationDTO
	LastDeployment *dbclient.Deployment
	// ReleaseId to deploy
	ReleaseID  string
	Operator   string
	DeployType string

	// Extras:
	// used for pipeline
	BuildID uint64
	// used for ability
	AddonActions map[string]interface{}
	// used for runtime-addon
	InstanceID string
	// used for
	Scale0 bool

	// 不由 orchestrator 来推进部署
	SkipPushByOrch bool
}
