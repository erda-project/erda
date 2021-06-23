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

package org

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/conf"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	"github.com/erda-project/erda/pkg/nexus"
)

// ensureNexusOrgMavenGroupRepo
// 1. maven hosted org snapshot repo
// 2. maven hosted publisher release repo
// 3. maven proxy official publisher repo
// 4. maven proxy thirdparty repos
// 5. one maven group org repo
func (o *Org) ensureNexusOrgMavenGroupRepos(org *apistructs.OrgDTO) error {
	var mavenMemberRepos []*apistructs.NexusRepository

	// 1. maven hosted org snapshot repo
	mavenHostedOrgSnapshotRepo, err := o.ensureNexusMavenHostedOrgSnapshotRepo(org)
	if err != nil {
		return err
	}
	mavenMemberRepos = append(mavenMemberRepos, mavenHostedOrgSnapshotRepo)

	// 2. maven hosted publisher release repo
	publisherID := o.GetPublisherID(int64(org.ID))
	var mavenHostedPublisherReleaseRepo *apistructs.NexusRepository
	if publisherID > 0 {
		dbRepos, err := o.db.ListNexusRepositories(apistructs.NexusRepositoryListRequest{
			PublisherID: &[]uint64{uint64(publisherID)}[0],
			OrgID:       &[]uint64{uint64(org.ID)}[0],
			Formats:     []nexus.RepositoryFormat{nexus.RepositoryFormatMaven},
			Types:       []nexus.RepositoryType{nexus.RepositoryTypeHosted},
		})
		if err != nil {
			return err
		}
		if len(dbRepos) > 0 {
			mavenHostedPublisherReleaseRepo = o.nexusSvc.ConvertRepo(dbRepos[0])
		}
	}
	if mavenHostedPublisherReleaseRepo != nil {
		mavenMemberRepos = append(mavenMemberRepos, mavenHostedPublisherReleaseRepo)
	}

	// 3. maven proxy official publisher repo
	// TODO

	// 4. maven proxy thirdpary repos
	thirdPartyDbRepos, err := o.db.ListNexusRepositories(apistructs.NexusRepositoryListRequest{
		OrgID:   &[]uint64{uint64(org.ID)}[0],
		Formats: []nexus.RepositoryFormat{nexus.RepositoryFormatMaven},
		Types:   []nexus.RepositoryType{nexus.RepositoryTypeProxy},
	})
	if err != nil {
		return err
	}
	mavenMemberRepos = append(mavenMemberRepos, o.nexusSvc.ConvertRepos(thirdPartyDbRepos)...)

	// 5. one maven group org repo
	_, err = o.ensureNexusMavenGroupOrgRepo(org, mavenMemberRepos)
	return err
}

func (o *Org) ensureNexusMavenHostedOrgSnapshotRepo(org *apistructs.OrgDTO) (*apistructs.NexusRepository, error) {
	nexusServer := nexus.Server{
		Addr:     conf.CentralNexusAddr(),
		Username: conf.CentralNexusUsername(),
		Password: conf.CentralNexusPassword(),
	}
	// ensure repo
	mavenRepoName := nexus.MakeOrgRepoName(nexus.RepositoryFormatMaven, nexus.RepositoryTypeHosted, uint64(org.ID), "snapshot")
	repo, err := o.nexusSvc.EnsureRepository(apistructs.NexusRepositoryEnsureRequest{
		OrgID:       &[]uint64{uint64(org.ID)}[0],
		PublisherID: nil,
		ClusterName: conf.DiceClusterName(),
		NexusServer: nexusServer,
		NexusCreateRequest: nexus.MavenHostedRepositoryCreateRequest{
			HostedRepositoryCreateRequest: nexus.HostedRepositoryCreateRequest{
				BaseRepositoryCreateRequest: nexus.BaseRepositoryCreateRequest{
					Name:   mavenRepoName,
					Online: true,
					Storage: nexus.HostedRepositoryStorageConfig{
						BlobStoreName:               mavenRepoName,
						StrictContentTypeValidation: true,
						WritePolicy:                 nexus.RepositoryStorageWritePolicyAllowRedploy,
					},
				},
			},
			Maven: nexus.RepositoryMavenConfig{
				VersionPolicy: nexus.RepositoryMavenVersionPolicySnapshot,
				LayoutPolicy:  nexus.RepositoryMavenLayoutPolicyPermissive,
			},
		},
		SyncConfigToPipelineCM: apistructs.NexusSyncConfigToPipelineCM{
			SyncOrg: &apistructs.NexusSyncConfigToPipelineCMItem{
				ConfigPrefix: "org.maven.snapshot.",
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// ensure maven hosted org snapshot deployment user
	user, err := o.ensureNexusHostedOrgDeploymentUser(org, repo, apistructs.NexusSyncConfigToPipelineCM{
		SyncOrg: &apistructs.NexusSyncConfigToPipelineCMItem{
			ConfigPrefix: "org.maven.snapshot.deployment.",
		},
	})
	if err != nil {
		return nil, err
	}
	repo.User = user

	return repo, nil
}

func (o *Org) ensureNexusMavenGroupOrgRepo(org *apistructs.OrgDTO, mavenMemberRepos []*apistructs.NexusRepository) (*apistructs.NexusRepository, error) {
	mavenGroupOrgRepoName := nexus.MakeOrgRepoName(nexus.RepositoryFormatMaven, nexus.RepositoryTypeGroup, uint64(org.ID))
	repo, err := o.nexusSvc.EnsureRepository(apistructs.NexusRepositoryEnsureRequest{
		OrgID:       &[]uint64{uint64(org.ID)}[0],
		PublisherID: nil,
		ClusterName: conf.DiceClusterName(),
		NexusServer: nexus.Server{
			Addr:     conf.CentralNexusAddr(),
			Username: conf.CentralNexusUsername(),
			Password: conf.CentralNexusPassword(),
		},
		NexusCreateRequest: nexus.MavenGroupRepositoryCreateRequest{
			GroupRepositoryCreateRequest: nexus.GroupRepositoryCreateRequest{
				BaseRepositoryCreateRequest: nexus.BaseRepositoryCreateRequest{
					Name:   mavenGroupOrgRepoName,
					Online: true,
					Storage: nexus.HostedRepositoryStorageConfig{
						BlobStoreName:               mavenGroupOrgRepoName,
						StrictContentTypeValidation: true,
						WritePolicy:                 nexus.RepositoryStorageWritePolicyReadOnly,
					},
					Cleanup: nil,
				},
				Group: nexus.RepositoryGroupConfig{
					MemberNames: func() []string {
						var memberNames []string
						for _, repo := range mavenMemberRepos {
							memberNames = append(memberNames, repo.Name)
						}
						return memberNames
					}(),
				},
			},
		},
		SyncConfigToPipelineCM: apistructs.NexusSyncConfigToPipelineCM{
			SyncOrg: &apistructs.NexusSyncConfigToPipelineCMItem{
				ConfigPrefix: "org.maven.",
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// ensure maven group org readonly user
	_, err = o.ensureNexusMavenGroupOrgReadonlyUser(org, repo, apistructs.NexusSyncConfigToPipelineCM{
		SyncOrg: &apistructs.NexusSyncConfigToPipelineCMItem{
			ConfigPrefix: "org.maven.readonly.",
		},
	})
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (o *Org) ensureNexusMavenGroupOrgReadonlyUser(
	org *apistructs.OrgDTO,
	groupRepo *apistructs.NexusRepository,
	syncCM apistructs.NexusSyncConfigToPipelineCM,
) (*apistructs.NexusUser, error) {
	if groupRepo.OrgID == nil || *groupRepo.OrgID != uint64(org.ID) {
		return nil, fmt.Errorf("group repo's org id %d mismatch org id %d", groupRepo.OrgID, org.ID)
	}
	userName := nexus.MakeReadonlyUserName(groupRepo.Name)
	return o.nexusSvc.EnsureUser(apistructs.NexusUserEnsureRequest{
		ClusterName:            groupRepo.ClusterName,
		RepoID:                 &groupRepo.ID,
		OrgID:                  groupRepo.OrgID,
		UserName:               userName,
		Password:               uuid.UUID(),
		RepoPrivileges:         map[uint64][]nexus.PrivilegeAction{groupRepo.ID: nexus.RepoReadOnlyPrivileges},
		SyncConfigToPipelineCM: syncCM,
		NexusServer: nexus.Server{
			Addr:     conf.CentralNexusAddr(),
			Username: conf.CentralNexusUsername(),
			Password: conf.CentralNexusPassword(),
		},
		ForceUpdatePassword: true,
	})
}

func (o *Org) ensureNexusHostedOrgDeploymentUser(
	org *apistructs.OrgDTO,
	repo *apistructs.NexusRepository,
	syncCM apistructs.NexusSyncConfigToPipelineCM,
) (*apistructs.NexusUser, error) {

	if repo.OrgID == nil || *repo.OrgID != uint64(org.ID) {
		return nil, fmt.Errorf("org repo's org id %d mismatch org id %d", repo.OrgID, org.ID)
	}
	userName := nexus.MakeDeploymentUserName(repo.Name)
	return o.nexusSvc.EnsureUser(apistructs.NexusUserEnsureRequest{
		ClusterName:            repo.ClusterName,
		RepoID:                 &repo.ID,
		OrgID:                  repo.OrgID,
		UserName:               userName,
		Password:               uuid.UUID(),
		RepoPrivileges:         map[uint64][]nexus.PrivilegeAction{repo.ID: nexus.RepoDeploymentPrivileges},
		SyncConfigToPipelineCM: syncCM,
		NexusServer: nexus.Server{
			Addr:     conf.CentralNexusAddr(),
			Username: conf.CentralNexusUsername(),
			Password: conf.CentralNexusPassword(),
		},
		ForceUpdatePassword: true,
	})
}
