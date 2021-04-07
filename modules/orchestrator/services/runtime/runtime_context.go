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
