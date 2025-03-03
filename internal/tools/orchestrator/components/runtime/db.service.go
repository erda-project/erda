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

package runtime

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

type DBService interface {
	GetRuntimeAllowNil(id uint64) (*dbclient.Runtime, error)
	FindRuntime(id spec.RuntimeUniqueId) (*dbclient.Runtime, error)
	FindLastDeployment(id uint64) (*dbclient.Deployment, error)
	FindDomainsByRuntimeId(id uint64) ([]dbclient.RuntimeDomain, error)
	GetRuntimeHPARulesByRuntimeId(runtimeID uint64) ([]dbclient.RuntimeHPA, error)
	GetRuntimeVPARulesByRuntimeId(runtimeID uint64) ([]dbclient.RuntimeVPA, error)
	GetRuntime(id uint64) (*dbclient.Runtime, error)
	UpdateRuntime(runtime *dbclient.Runtime) error
	FindPreDeployment(uniqueId spec.RuntimeUniqueId) (*dbclient.PreDeployment, error)
	FindRuntimesByIds(ids []uint64) ([]dbclient.Runtime, error)
	GetUnDeletableAttachMentsByRuntimeID(orgID, runtimeID uint64) (*[]dbclient.AddonAttachment, error)
	GetInstanceRouting(id string) (*dbclient.AddonInstanceRouting, error)
	UpdateAttachment(addonAttachment *dbclient.AddonAttachment) error
	UpdatePreDeployment(pre *dbclient.PreDeployment) error
	FindRuntimeOrCreate(uniqueId spec.RuntimeUniqueId, operator string, source apistructs.RuntimeSource,
		clusterName string, clusterId uint64, gitRepoAbbrev string, projectID, orgID uint64, deploymentOrderId,
		releaseVersion, extraParams string) (*dbclient.Runtime, bool, error)
	FindPreDeploymentOrCreate(uniqueId spec.RuntimeUniqueId, dice *diceyml.DiceYaml) (*dbclient.PreDeployment, error)
	CreateOrUpdateRuntimeService(service *dbclient.RuntimeService, overrideStatus bool) error
	CreateDeployment(deployment *dbclient.Deployment) error
	FindRuntimesByAppId(appId uint64) ([]dbclient.Runtime, error)
	FindLastSuccessDeployment(runtimeId uint64) (*dbclient.Deployment, error)
	FindRuntimesNewerThan(minId uint64, limit int) ([]dbclient.Runtime, error)
	ListRuntimeByOrgCluster(clusterName string, orgID uint64) ([]dbclient.Runtime, error)
	ListRoutingInstanceByOrgCluster(clusterName string, orgID uint64) ([]dbclient.AddonInstanceRouting, error)
	GetDeployment(id uint64) (*dbclient.Deployment, error)
	FindRuntimesInApps(appIDs []uint64, env string) (map[uint64][]*dbclient.Runtime, []uint64, error)
	FindLastDeploymentIDsByRutimeIDs(runtimeIDs []uint64) ([]uint64, error)
	FindDeploymentsByIDs(ids []uint64) (map[uint64]dbclient.Deployment, error)
	GetAppRuntimeNumberByWorkspace(projectId uint64, env string) (uint64, error)
	FindTopDeployments(runtimeId uint64, limit int) ([]dbclient.Deployment, error)
	FindNotOutdatedOlderThan(runtimeId uint64, maxId uint64) ([]dbclient.Deployment, error)
	UpdateDeployment(deployment *dbclient.Deployment) error
	GetRuntimeByProjectIDs(projectIDs []uint64) (*[]dbclient.Runtime, error)
	ListAddonInstancesByProjectIDs(projectIDs []uint64, exclude ...string) (*[]dbclient.AddonInstance, error)
	GetAddonNodesByInstanceIDs(instanceIDs []string) (*[]dbclient.AddonNode, error)
	GetRuntimeHPAByServices(id spec.RuntimeUniqueId, services []string) ([]dbclient.RuntimeHPA, error)
	GetRuntimeVPAByServices(id spec.RuntimeUniqueId, services []string) ([]dbclient.RuntimeVPA, error)
}
