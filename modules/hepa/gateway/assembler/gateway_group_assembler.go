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

package assembler

import (
	gw "github.com/erda-project/erda/modules/hepa/gateway/dto"
)

type GatewayGroupAssemblerImpl struct {
}

func (GatewayGroupAssemblerImpl) GroupInfo2Dto(infos []gw.GatewayGroupInfo) ([]gw.GwApiGroupDto, error) {
	dtos := []gw.GwApiGroupDto{}
	for _, info := range infos {
		group := info.GatewayGroup
		policies := info.Policies
		if group.IsEmpty() {
			continue
		}
		dto := &gw.GwApiGroupDto{}
		dto.GroupId = group.Id
		dto.GroupName = group.GroupName
		dto.DisplayName = group.DispalyName
		dto.CreateAt = group.CreateTime.Format("2006-01-02T15:04:05")
		if len(policies) == 0 {
			dtos = append(dtos, *dto)
			continue
		}
		for _, policy := range policies {
			policyDto := &gw.GwApiPolicyDto{}
			policyDto.PolicyId = policy.Id
			policyDto.Category = policy.Category
			policyDto.DisplayName = policy.DisplayName
			dto.Policies = append(dto.Policies, *policyDto)
		}
		dtos = append(dtos, *dto)
	}
	return dtos, nil
}
