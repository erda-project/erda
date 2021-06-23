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
	"github.com/erda-project/erda/modules/dop/conf"
	"github.com/erda-project/erda/modules/dop/model"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	"github.com/erda-project/erda/pkg/nexus"
)

// ensureNexusNpmHostedRepo
// 1. 保证 nexus npm-hosted-repo 存在
// 2. 保证 nexus npm-hosted-repo-deployment user 存在
func (p *Publisher) ensureNexusNpmHostedRepo(publisher *model.Publisher) error {
	nexusServer := nexus.Server{
		Addr:     conf.CentralNexusAddr(),
		Username: conf.CentralNexusUsername(),
		Password: conf.CentralNexusPassword(),
	}
	repoName := nexus.MakePublisherRepoName(nexus.RepositoryFormatNpm, nexus.RepositoryTypeHosted, uint64(publisher.ID))
	// ensure repo
	repo, err := p.nexusSvc.EnsureRepository(apistructs.NexusRepositoryEnsureRequest{
		OrgID:       &[]uint64{uint64(publisher.OrgID)}[0],
		PublisherID: &[]uint64{uint64(publisher.ID)}[0],
		ClusterName: conf.DiceClusterName(),
		NexusServer: nexusServer,
		NexusCreateRequest: nexus.NpmHostedRepositoryCreateRequest{
			HostedRepositoryCreateRequest: nexus.HostedRepositoryCreateRequest{
				BaseRepositoryCreateRequest: nexus.BaseRepositoryCreateRequest{
					Name:   repoName,
					Online: true,
					Storage: nexus.HostedRepositoryStorageConfig{
						BlobStoreName:               repoName,
						StrictContentTypeValidation: true,
						WritePolicy:                 nexus.RepositoryStorageWritePolicyDisableRedeploy,
					},
					Cleanup: nil,
				},
			},
		},
		SyncConfigToPipelineCM: apistructs.NexusSyncConfigToPipelineCM{
			SyncPublisher: &apistructs.NexusSyncConfigToPipelineCMItem{
				ConfigPrefix: "publisher.npm.",
			},
		},
	})
	if err != nil {
		return err
	}

	// ensure deployment user
	_, err = p.ensurePublisherDeploymentUser(apistructs.NexusDeploymentUserEnsureRequest{
		RepoID:      repo.ID,
		Password:    uuid.UUID(),
		NexusServer: nexusServer,
		SyncConfigToPipelineCM: apistructs.NexusSyncConfigToPipelineCM{
			SyncPublisher: &apistructs.NexusSyncConfigToPipelineCMItem{
				ConfigPrefix: "publisher.npm.",
			},
		},
	})
	return err
}
