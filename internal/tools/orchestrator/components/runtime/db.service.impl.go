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
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type dbServiceImpl struct {
	db *dbclient.DBClient
}

func (d *dbServiceImpl) FindRuntimeOrCreate(uniqueId spec.RuntimeUniqueId, operator string, source apistructs.RuntimeSource, clusterName string, clusterId uint64, gitRepoAbbrev string, projectID, orgID uint64, deploymentOrderId, releaseVersion, extraParams string) (*dbclient.Runtime, bool, error) {
	return d.db.FindRuntimeOrCreate(uniqueId, operator, source, clusterName, clusterId, gitRepoAbbrev, projectID, orgID, deploymentOrderId, releaseVersion, extraParams)
}

func (d *dbServiceImpl) FindPreDeploymentOrCreate(uniqueId spec.RuntimeUniqueId, dice *diceyml.DiceYaml) (*dbclient.PreDeployment, error) {
	return d.db.FindPreDeploymentOrCreate(uniqueId, dice)
}

func (d *dbServiceImpl) CreateOrUpdateRuntimeService(service *dbclient.RuntimeService, overrideStatus bool) error {
	return d.db.CreateOrUpdateRuntimeService(service, overrideStatus)
}

func (d *dbServiceImpl) CreateDeployment(deployment *dbclient.Deployment) error {
	return d.db.CreateDeployment(deployment)
}

func (d *dbServiceImpl) FindRuntimesByAppId(appId uint64) ([]dbclient.Runtime, error) {
	return d.db.FindRuntimesByAppId(appId)
}

func (d *dbServiceImpl) FindLastSuccessDeployment(runtimeId uint64) (*dbclient.Deployment, error) {
	return d.db.FindLastSuccessDeployment(runtimeId)
}

func (d *dbServiceImpl) FindRuntimesNewerThan(minId uint64, limit int) ([]dbclient.Runtime, error) {
	return d.db.FindRuntimesNewerThan(minId, limit)
}

func (d *dbServiceImpl) ListRuntimeByOrgCluster(clusterName string, orgID uint64) ([]dbclient.Runtime, error) {
	return d.db.ListRuntimeByOrgCluster(clusterName, orgID)
}

func (d *dbServiceImpl) ListRoutingInstanceByOrgCluster(clusterName string, orgID uint64) ([]dbclient.AddonInstanceRouting, error) {
	return d.db.ListRoutingInstanceByOrgCluster(clusterName, orgID)
}

func (d *dbServiceImpl) GetDeployment(id uint64) (*dbclient.Deployment, error) {
	return d.db.GetDeployment(id)
}

func (d *dbServiceImpl) FindRuntimesInApps(appIDs []uint64, env string) (map[uint64][]*dbclient.Runtime, []uint64, error) {
	return d.db.FindRuntimesInApps(appIDs, env)
}

func (d *dbServiceImpl) FindLastDeploymentIDsByRutimeIDs(runtimeIDs []uint64) ([]uint64, error) {
	return d.db.FindLastDeploymentIDsByRutimeIDs(runtimeIDs)
}

func (d *dbServiceImpl) FindDeploymentsByIDs(ids []uint64) (map[uint64]dbclient.Deployment, error) {
	return d.db.FindDeploymentsByIDs(ids)
}

func (d *dbServiceImpl) GetAppRuntimeNumberByWorkspace(projectId uint64, env string) (uint64, error) {
	return d.db.GetAppRuntimeNumberByWorkspace(projectId, env)
}
func (d *dbServiceImpl) FindTopDeployments(runtimeId uint64, limit int) ([]dbclient.Deployment, error) {
	return d.db.FindTopDeployments(runtimeId, limit)
}

func (d *dbServiceImpl) FindNotOutdatedOlderThan(runtimeId uint64, maxId uint64) ([]dbclient.Deployment, error) {
	return d.db.FindNotOutdatedOlderThan(runtimeId, maxId)
}

func (d *dbServiceImpl) UpdateDeployment(deployment *dbclient.Deployment) error {
	return d.db.UpdateDeployment(deployment)
}

func (d *dbServiceImpl) FindPreDeployment(uniqueId spec.RuntimeUniqueId) (*dbclient.PreDeployment, error) {
	var preBuild dbclient.PreDeployment
	if err := d.db.Table("ps_v2_pre_builds").
		Where("project_id = ? AND env = ? AND git_branch = ?", uniqueId.ApplicationId, uniqueId.Workspace, uniqueId.Name).
		Take(&preBuild).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find PreDeployment, uniqueId: %v", uniqueId)
	}
	return &preBuild, nil
}

func (d *dbServiceImpl) FindRuntimesByIds(ids []uint64) ([]dbclient.Runtime, error) {
	var runtimes []dbclient.Runtime
	if len(ids) == 0 {
		return runtimes, nil
	}
	if err := d.db.
		Where("id in (?)", ids).
		Find(&runtimes).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find runtimes by Ids: %v", ids)
	}
	return runtimes, nil
}

func (d *dbServiceImpl) GetUnDeletableAttachMentsByRuntimeID(orgID, runtimeID uint64) (*[]dbclient.AddonAttachment, error) {
	var attachments []dbclient.AddonAttachment
	if err := d.db.
		Where("org_id = ?", orgID).
		Where("app_id = ?", runtimeID).
		Where("is_deleted != ?", apistructs.AddonDeleted).
		Find(&attachments).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get addon attachments info, runtimeID : %d",
			runtimeID)
	}
	return &attachments, nil
}

func (d *dbServiceImpl) GetInstanceRouting(id string) (*dbclient.AddonInstanceRouting, error) {
	var instanceRouting dbclient.AddonInstanceRouting
	if err := d.db.Where("id = ?", id).Where("is_deleted = ?", apistructs.AddonNotDeleted).
		Find(&instanceRouting).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &instanceRouting, nil
}

func (d *dbServiceImpl) UpdateAttachment(addonAttachment *dbclient.AddonAttachment) error {
	return d.db.Save(addonAttachment).Error
}

func (d *dbServiceImpl) UpdatePreDeployment(pre *dbclient.PreDeployment) error {
	if err := d.db.Table("ps_v2_pre_builds").Save(pre).Error; err != nil {
		return errors.Wrapf(err, "failed to update PreDeployment, pre: %v", pre)
	}
	return nil
}

func (d *dbServiceImpl) UpdateRuntime(runtime *dbclient.Runtime) error {
	return d.db.UpdateRuntime(runtime)
}

func (d *dbServiceImpl) GetRuntime(id uint64) (*dbclient.Runtime, error) {
	return d.db.GetRuntime(id)
}

func (d *dbServiceImpl) FindDomainsByRuntimeId(id uint64) ([]dbclient.RuntimeDomain, error) {
	return d.db.FindDomainsByRuntimeId(id)
}

func (d *dbServiceImpl) FindLastDeployment(id uint64) (*dbclient.Deployment, error) {
	return d.db.FindLastDeployment(id)
}

func (d *dbServiceImpl) FindRuntime(id spec.RuntimeUniqueId) (*dbclient.Runtime, error) {
	return d.db.FindRuntime(id)
}

func (d *dbServiceImpl) GetRuntimeAllowNil(id uint64) (*dbclient.Runtime, error) {
	return d.db.GetRuntimeAllowNil(id)
}

func (d *dbServiceImpl) GetRuntimeHPARulesByRuntimeId(runtimeID uint64) ([]dbclient.RuntimeHPA, error) {
	return d.db.GetRuntimeHPARulesByRuntimeId(runtimeID)
}

func (d *dbServiceImpl) GetRuntimeVPARulesByRuntimeId(runtimeID uint64) ([]dbclient.RuntimeVPA, error) {
	return d.db.GetRuntimeVPARulesByRuntimeId(runtimeID)
}

func newDBService(db *dbclient.DBClient) DBService {
	return &dbServiceImpl{db: db}
}

func NewDBService(orm *gorm.DB) DBService {
	return newDBService(&dbclient.DBClient{
		DBEngine: &dbengine.DBEngine{
			DB: orm,
		},
	})
}
