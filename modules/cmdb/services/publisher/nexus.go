package publisher

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
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
