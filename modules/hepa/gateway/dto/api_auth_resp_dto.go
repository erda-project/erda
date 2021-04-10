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
	log "github.com/sirupsen/logrus"
)

type ApiAuthData struct {
	Access bool `json:"access"`
}

type ApiAuthErr struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

type ApiAuthRespDto struct {
	Success bool        `json:"success"`
	Data    ApiAuthData `json:"data"`
	Err     ApiAuthErr  `json:"err"`
}

func (dto ApiAuthRespDto) HasPermission() bool {
	if !dto.Success {
		log.Errorf("auth resp failed, code:%s, msg:%s", dto.Err.Code, dto.Err.Msg)
		return false
	}
	return dto.Data.Access
}
