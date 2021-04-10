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
	"strings"

	log "github.com/sirupsen/logrus"
)

type UpstreamApiDto struct {
	Path        string      `json:"path"`
	GatewayPath string      `json:"gatewayPath"`
	Method      string      `json:"method"`
	Address     string      `json:"address"`
	IsInner     bool        `json:"isInner"`
	IsCustom    bool        `json:"isCustom"`
	Doc         interface{} `json:"doc"`
	Name        string      `json:"name"`
}

func (dto *UpstreamApiDto) Init() bool {
	if len(dto.Path) == 0 || dto.Path[0] != '/' {
		log.Errorf("invalid path:%+v", dto)
		return false
	}
	if len(dto.Address) == 0 || !strings.HasPrefix(dto.Address, "http") {
		log.Errorf("invalid address:%+v", dto)
		return false
	}
	if len(dto.Name) == 0 {
		dto.Name = dto.Method + dto.Path
	}
	if dto.GatewayPath == "" {
		dto.GatewayPath = dto.Path
	}
	return true
}
