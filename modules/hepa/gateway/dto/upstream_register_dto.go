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

package dto

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type UpstreamRegisterDto struct {
	Az            string           `json:"az"`
	UpstreamName  string           `json:"-"`
	DiceAppId     string           `json:"diceAppId"`
	DiceService   string           `json:"diceService"`
	RuntimeName   string           `json:"runtimeName"`
	RuntimeId     string           `json:"runtimeId"`
	AppName       string           `json:"appName"`
	ServiceAlias  string           `json:"serviceName"`
	OrgId         string           `json:"orgId"`
	ProjectId     string           `json:"projectId"`
	Env           string           `json:"env"`
	ApiList       []UpstreamApiDto `json:"apiList"`
	OldRegisterId *int             `json:"registerId"`
	RegisterId    string           `json:"registerTag"`
	PathPrefix    *string          `json:"pathPrefix"`
}

func (dto *UpstreamRegisterDto) Init() bool {
	if len(dto.AppName) == 0 || len(dto.OrgId) == 0 || len(dto.ProjectId) == 0 || len(dto.ApiList) == 0 ||
		len(dto.Env) == 0 || (dto.OldRegisterId == nil && dto.RegisterId == "") {
		log.Errorf("invalid dto:%+v", dto)
		return false
	}
	dto.UpstreamName = dto.AppName
	if dto.DiceService == "" {
		dto.DiceService = dto.ServiceAlias
	}
	if dto.ServiceAlias != "" {
		dto.UpstreamName += "/" + dto.ServiceAlias
	}
	if dto.RuntimeId != "" {
		dto.UpstreamName += "/" + dto.RuntimeId
	}
	if dto.RuntimeName == "" {
		if dto.PathPrefix == nil {
			path := strings.ToLower("/" + dto.UpstreamName)
			dto.PathPrefix = &path
		} else {
			path := "/" + strings.TrimPrefix(*dto.PathPrefix, "/")
			path = strings.TrimSuffix(path, "/")
			dto.PathPrefix = &path
		}
	} else {
		dto.PathPrefix = nil
	}
	if dto.OldRegisterId != nil && dto.RegisterId == "" {
		dto.RegisterId = strconv.Itoa(*dto.OldRegisterId)
	}
	address := dto.ApiList[0].Address
	for i := len(dto.ApiList) - 1; i >= 0; i-- {
		if !dto.ApiList[i].IsCustom {
			dto.ApiList = append(dto.ApiList[:i], dto.ApiList[i+1:]...)
		}
	}
	if len(dto.ApiList) == 0 {
		dto.ApiList = []UpstreamApiDto{
			{
				Path:        "/",
				GatewayPath: "/",
				Method:      "",
				Address:     address,
				IsInner:     false,
				Name:        "/",
			},
		}
	}
	return true
}
