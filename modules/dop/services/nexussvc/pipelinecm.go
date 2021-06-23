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

package nexussvc

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/nexus"
)

func (svc *NexusSvc) SyncRepoConfigToPipelineCM(syncConfig apistructs.NexusSyncConfigToPipelineCM, repo *apistructs.NexusRepository) error {
	if syncConfig.SyncPublisher != nil && repo.PublisherID != nil && *repo.PublisherID > 0 {
		if err := svc.bdl.CreateOrUpdatePipelineCmsNsConfigs(nexus.MakePublisherPipelineCmNs(*repo.PublisherID),
			apistructs.PipelineCmsUpdateConfigsRequest{
				PipelineSource: apistructs.PipelineSourceDice,
				KVs:            generateRepoPipelineCmConfigs(repo, syncConfig.SyncPublisher.ConfigPrefix),
			},
		); err != nil {
			return err
		}
	}
	if syncConfig.SyncOrg != nil && repo.OrgID != nil && *repo.OrgID > 0 {
		if err := svc.bdl.CreateOrUpdatePipelineCmsNsConfigs(nexus.MakeOrgPipelineCmsNs(*repo.OrgID),
			apistructs.PipelineCmsUpdateConfigsRequest{
				PipelineSource: apistructs.PipelineSourceDice,
				KVs:            generateRepoPipelineCmConfigs(repo, syncConfig.SyncOrg.ConfigPrefix),
			},
		); err != nil {
			return err
		}
	}
	if syncConfig.SyncPlatform != nil {
		if err := svc.bdl.CreateOrUpdatePipelineCmsNsConfigs(nexus.MakePlatformPipelineCmsNs(),
			apistructs.PipelineCmsUpdateConfigsRequest{
				PipelineSource: apistructs.PipelineSourceDice,
				KVs:            generateRepoPipelineCmConfigs(repo, syncConfig.SyncPlatform.ConfigPrefix),
			},
		); err != nil {
			return err
		}
	}

	return nil
}

func (svc *NexusSvc) SyncUserConfigToPipelineCM(syncConfig apistructs.NexusSyncConfigToPipelineCM, user *apistructs.NexusUser, repoFormat nexus.RepositoryFormat) error {
	if syncConfig.SyncPublisher != nil && user.PublisherID != nil && *user.PublisherID > 0 {
		if err := svc.bdl.CreateOrUpdatePipelineCmsNsConfigs(nexus.MakePublisherPipelineCmNs(*user.PublisherID),
			apistructs.PipelineCmsUpdateConfigsRequest{
				PipelineSource: apistructs.PipelineSourceDice,
				KVs:            generateUserPipelineCmConfigs(user.Name, user.Password, syncConfig.SyncPublisher.ConfigPrefix),
			},
		); err != nil {
			return err
		}
	}
	if syncConfig.SyncOrg != nil && user.OrgID != nil && *user.OrgID > 0 {
		if err := svc.bdl.CreateOrUpdatePipelineCmsNsConfigs(nexus.MakeOrgPipelineCmsNs(*user.OrgID),
			apistructs.PipelineCmsUpdateConfigsRequest{
				PipelineSource: apistructs.PipelineSourceDice,
				KVs:            generateUserPipelineCmConfigs(user.Name, user.Password, syncConfig.SyncOrg.ConfigPrefix),
			},
		); err != nil {
			return err
		}
	}
	if syncConfig.SyncPlatform != nil {
		if err := svc.bdl.CreateOrUpdatePipelineCmsNsConfigs(nexus.MakePlatformPipelineCmsNs(),
			apistructs.PipelineCmsUpdateConfigsRequest{
				PipelineSource: apistructs.PipelineSourceDice,
				KVs:            generateUserPipelineCmConfigs(user.Name, user.Password, syncConfig.SyncPlatform.ConfigPrefix),
			},
		); err != nil {
			return err
		}
	}

	return nil
}

func generateRepoPipelineCmConfigs(repo *apistructs.NexusRepository, keyPrefix string) map[string]apistructs.PipelineCmsConfigValue {
	configs := make(map[string]apistructs.PipelineCmsConfigValue)

	configs[keyPrefix+"url"] = apistructs.PipelineCmsConfigValue{
		Value:       repo.URL,
		EncryptInDB: false,
		Type:        apistructs.PipelineCmsConfigTypeKV,
		Operations:  &apistructs.PipelineCmsConfigOperations{CanDownload: false, CanEdit: false, CanDelete: false},
		Comment:     "nexus repo url",
	}

	return configs
}

func generateUserPipelineCmConfigs(username, plainPassword, keyPrefix string) map[string]apistructs.PipelineCmsConfigValue {
	configs := make(map[string]apistructs.PipelineCmsConfigValue)

	configs[keyPrefix+"username"] = apistructs.PipelineCmsConfigValue{
		Value:       username,
		EncryptInDB: false,
		Type:        apistructs.PipelineCmsConfigTypeKV,
		Operations:  &apistructs.PipelineCmsConfigOperations{CanDownload: false, CanEdit: false, CanDelete: false},
		Comment:     "nexus repo username",
	}
	configs[keyPrefix+"password"] = apistructs.PipelineCmsConfigValue{
		Value:       plainPassword,
		EncryptInDB: true,
		Type:        apistructs.PipelineCmsConfigTypeKV,
		Operations:  &apistructs.PipelineCmsConfigOperations{CanDownload: false, CanEdit: false, CanDelete: false},
		Comment:     "nexus repo password",
	}

	return configs
}
