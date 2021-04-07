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

package events

import (
	"github.com/erda-project/erda/apistructs"
)

type EventName string

const (
	// create
	RuntimeCreated EventName = "RuntimeCreated"
	// delete
	RuntimeDeleting     EventName = "RuntimeDeleting"
	RuntimeDeleted      EventName = "RuntimeDeleted"
	RuntimeDeleteFailed EventName = "RuntimeDeleteFailed"
	// runtime status
	RuntimeStatusChanged EventName = "RuntimeStatusChanged"
	// service
	RuntimeServiceStatusChanged    EventName = "RuntimeServiceStatusChanged"
	RuntimeServiceInstancesChanged EventName = "RuntimeServiceInstancesChanged"
	// deploy
	RuntimeDeployStart         EventName = "RuntimeDeployStart"
	RuntimeDeployStatusChanged EventName = "RuntimeDeployStatusChanged"
	RuntimeDeployFailed        EventName = "RuntimeDeployFailed"
	RuntimeDeployCanceling     EventName = "RuntimeDeployCanceling"
	RuntimeDeployCanceled      EventName = "RuntimeDeployCanceled"
	RuntimeDeployCancelFailed  EventName = "RuntimeDeployCancelFailed"
	RuntimeDeployOk            EventName = "RuntimeDeployOk"
)

type ActionName string

const (
	R_ADD           ActionName = "R_ADD"
	R_DEL           ActionName = "R_DEL"
	R_DEPLOY_START  ActionName = "R_DEPLOY_START"
	R_DEPLOY_FAIL   ActionName = "R_DEPLOY_FAIL"
	R_DEPLOY_CANCEL ActionName = "R_DEPLOY_CANCEL"
	R_DEPLOY_OK     ActionName = "R_DEPLOY_OK"
)

type RuntimeEvent struct {
	EventName EventName              `json:"eventName"`
	Operator  string                 `json:"operator"`
	Runtime   *apistructs.RuntimeDTO `json:"runtime,omitempty"`
	// only used for RuntimeService* events
	Service *apistructs.RuntimeServiceDTO `json:"service,omitempty"`
	// only used for RuntimeServiceInstancesChanged
	Instances []*apistructs.RuntimeInstanceDTO `json:"instance,omitempty"`
	// only used for RuntimeDeploy* events
	Deployment *apistructs.Deployment `json:"deployment,omitempty"`
}
