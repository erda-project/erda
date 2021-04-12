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
	"encoding/json"

	"github.com/pkg/errors"
)

type RegisterDto struct {
	OrgId       string          `json:"orgId"`
	ProjectId   string          `json:"projectId"`
	Workspace   string          `json:"workspace"`
	ClusterName string          `json:"clusterName"`
	AppId       string          `json:"appId"`
	AppName     string          `json:"appName"`
	RuntimeId   string          `json:"runtimeId"`
	RuntimeName string          `json:"runtimeName"`
	ServiceName string          `json:"serviceName"`
	ServiceAddr string          `json:"serviceAddr"`
	Swagger     json.RawMessage `json:"swagger"`
}

func (dto RegisterDto) CheckValid() error {
	if dto.OrgId == "" || dto.ProjectId == "" || dto.Workspace == "" || dto.ClusterName == "" || dto.AppId == "" || dto.ServiceName == "" || dto.Swagger == nil {
		return errors.Errorf("invalid RegisterDto: %+v", dto)
	}
	return nil
}

type RegisterRespDto struct {
	ApiRegisterId string `json:"apiRegisterId"`
}

type RegisterStatusDto struct {
	Completed bool `json:"completed"`
}
