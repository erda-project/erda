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

package deployment_order

import (
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/services/deployment"
	"github.com/erda-project/erda/modules/orchestrator/services/runtime"
)

const (
	appOrderPrefix        = "a_"
	projectOrderPrefix    = "p_"
	orderNameTmpl         = "%s_%d"
	release               = "RELEASE"
	gitBranchLabel        = "gitBranch"
	orderStatusWaitDeploy = "WAITDEPLOY"
)

type DeploymentOrder struct {
	db     *dbclient.DBClient
	bdl    *bundle.Bundle
	rt     *runtime.Runtime
	deploy *deployment.Deployment
}

type Option func(*DeploymentOrder)

func New(options ...Option) *DeploymentOrder {
	r := &DeploymentOrder{}
	for _, op := range options {
		op(r)
	}
	return r
}

// WithDBClient with database client
func WithDBClient(db *dbclient.DBClient) Option {
	return func(r *DeploymentOrder) {
		r.db = db
	}
}

// WithBundle with bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(d *DeploymentOrder) {
		d.bdl = bdl
	}
}

// WithRuntime with runtime service
func WithRuntime(rt *runtime.Runtime) Option {
	return func(d *DeploymentOrder) {
		d.rt = rt
	}
}

// WithDeployment with deployment service
func WithDeployment(deploy *deployment.Deployment) Option {
	return func(d *DeploymentOrder) {
		d.deploy = deploy
	}
}
