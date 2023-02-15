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

package dto

import (
	"github.com/erda-project/erda-proto-go/core/hepa/endpoint_api/pb"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
)

type PackageInfoDto struct {
	Id       string `json:"id"`
	CreateAt string `json:"createAt"`
	PackageDto
}

type SortBySceneList []PackageInfoDto

func (list SortBySceneList) Len() int { return len(list) }

func (list SortBySceneList) Swap(i, j int) { list[i], list[j] = list[j], list[i] }

func (list SortBySceneList) Less(i, j int) bool {
	if list[i].Scene == list[j].Scene {
		if list[i].Scene == orm.HubScene {
			return list[i].Name < list[j].Name
		}
		return list[i].CreateAt > list[j].CreateAt
	}
	for _, scene := range []string{orm.UnityScene, orm.HubScene, orm.WebapiScene, orm.OpenapiScene} {
		if list[i].Scene == scene {
			return true
		}
		if list[j].Scene == scene {
			return false
		}
	}
	return list[i].CreateAt > list[j].CreateAt
}

func (dto PackageInfoDto) ToEndpoint() *pb.Endpoint {
	return &pb.Endpoint{
		Id:              dto.Id,
		CreateAt:        dto.CreateAt,
		Name:            dto.Name,
		BindDomain:      dto.BindDomain,
		AuthType:        dto.AuthType,
		AclType:         dto.AclType,
		Scene:           dto.Scene,
		Description:     dto.Description,
		GatewayProvider: dto.GatewayProvider,
	}
}

func FromEndpoint(ep *pb.Endpoint) *PackageDto {
	return &PackageDto{
		Name:        ep.Name,
		BindDomain:  ep.BindDomain,
		AuthType:    ep.AuthType,
		AclType:     ep.AclType,
		Scene:       ep.Scene,
		Description: ep.Description,
	}
}
