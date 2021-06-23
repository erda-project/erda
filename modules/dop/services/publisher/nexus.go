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

package publisher

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/model"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/nexus"
)

func (p *Publisher) ensureNexusHostedRepo(publisher *model.Publisher) error {
	// npm
	if err := p.ensureNexusNpmHostedRepo(publisher); err != nil {
		return err
	}
	// maven
	if err := p.ensureNexusMavenHostedRepo(publisher); err != nil {
		return err
	}

	return nil
}

// ensurePublisherDeploymentUser 幂等保证 nexus repo deployment user 存在
func (p *Publisher) ensurePublisherDeploymentUser(req apistructs.NexusDeploymentUserEnsureRequest) (*apistructs.NexusUser, error) {
	// 查询 nexus repo db 记录
	dbRepo, err := p.db.GetNexusRepository(req.RepoID)
	if err != nil {
		return nil, apierrors.ErrGetNexusRepoRecord.InternalError(err)
	}
	nexusUser, err := p.nexusSvc.EnsureUser(apistructs.NexusUserEnsureRequest{
		ClusterName:            dbRepo.ClusterName,
		RepoID:                 &req.RepoID,
		OrgID:                  dbRepo.OrgID,
		UserName:               nexus.MakeDeploymentUserName(dbRepo.Name),
		Password:               req.Password,
		RepoPrivileges:         map[uint64][]nexus.PrivilegeAction{req.RepoID: nexus.RepoDeploymentPrivileges},
		SyncConfigToPipelineCM: req.SyncConfigToPipelineCM,
		NexusServer:            req.NexusServer,
		ForceUpdatePassword:    true,
	})
	if err != nil {
		return nil, err
	}
	return nexusUser, nil
}
