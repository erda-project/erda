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
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/nexus"
)

func (o *Org) EnsureNexusOrgGroupRepos(org *apistructs.OrgDTO) error {
	// group repos

	// TODO nexus 3.24
	//// maven
	//if err := o.ensureNexusOrgMavenGroupRepos(org); err != nil {
	//	return err
	//}

	// npm
	if err := o.ensureNexusNpmGroupOrgRepos(org); err != nil {
		return err
	}

	// docker
	if err := o.ensureNexusDockerGroupOrgRepos(org); err != nil {
		return err
	}

	return nil
}

func (o *Org) GetOrgLevelNexus(orgID uint64, req *apistructs.OrgNexusGetRequest) (*apistructs.OrgNexusGetResponseData, error) {
	if orgID == 0 {
		return nil, apierrors.ErrGetOrgNexus.MissingParameter("orgID")
	}

	orgRepos, err := o.nexusSvc.ListRepositories(apistructs.NexusRepositoryListRequest{
		OrgID:   &orgID,
		Formats: req.Formats,
		Types:   req.Types,
	})
	if err != nil {
		return nil, apierrors.ErrGetOrgNexus.InternalError(err)
	}

	data := apistructs.OrgNexusGetResponseData{
		OrgGroupRepos:         make(map[nexus.RepositoryFormat]*apistructs.NexusRepository),
		OrgSnapshotRepos:      make(map[nexus.RepositoryFormat]*apistructs.NexusRepository),
		PublisherReleaseRepos: make(map[nexus.RepositoryFormat]*apistructs.NexusRepository),
		ThirdPartyProxyRepos:  make(map[nexus.RepositoryFormat][]*apistructs.NexusRepository),
	}

	for _, repo := range orgRepos {
		switch repo.Type {
		case nexus.RepositoryTypeGroup:
			data.OrgGroupRepos[repo.Format] = repo
		case nexus.RepositoryTypeHosted:
			if repo.PublisherID == nil {
				data.OrgSnapshotRepos[repo.Format] = repo
			} else {
				data.PublisherReleaseRepos[repo.Format] = repo
			}
		case nexus.RepositoryTypeProxy:
			data.ThirdPartyProxyRepos[repo.Format] = append(data.ThirdPartyProxyRepos[repo.Format], repo)
		}
	}

	return &data, nil
}

func (o *Org) ShowOrgNexusPassword(req *apistructs.OrgNexusShowPasswordRequest) (map[uint64]string, error) {
	users, err := o.nexusSvc.ListUsers(apistructs.NexusUserListRequest{
		UserIDs:        req.NexusUserIDs,
		PublisherID:    nil,
		OrgID:          &req.OrgID,
		RepoID:         nil,
		DecodePassword: true,
	})
	if err != nil {
		return nil, apierrors.ErrShowOrgNexusPassword.InternalError(err)
	}

	userPasswordMap := make(map[uint64]string, len(users))
	for _, user := range users {
		userPasswordMap[user.ID] = user.Password
	}

	var notFoundUserIDs []uint64
	for _, reqUserID := range req.NexusUserIDs {
		if _, ok := userPasswordMap[reqUserID]; !ok {
			notFoundUserIDs = append(notFoundUserIDs, reqUserID)
		}
	}
	if len(notFoundUserIDs) > 0 {
		return nil, apierrors.ErrShowOrgNexusPassword.InvalidParameter(
			fmt.Sprintf("couldn't found corresponding nexus users, invalid nexus user ids: %v", notFoundUserIDs))
	}
	return userPasswordMap, nil
}
