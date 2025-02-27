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
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type dbServiceImpl struct {
	db *dbclient.DBClient
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
