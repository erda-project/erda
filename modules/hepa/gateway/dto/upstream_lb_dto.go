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

import "strings"

type UpstreamLbDto struct {
	Az              string   `json:"az"`
	LbName          string   `json:"lbName"`
	OrgId           string   `json:"orgId"`
	ProjectId       string   `json:"projectId"`
	Env             string   `json:"env"`
	DeploymentId    int      `json:"deploymentId"`
	HealthcheckPath string   `json:"healthcheckPath"`
	Targets         []string `json:"targets"`
}

func (dto UpstreamLbDto) IsInvalid() bool {
	return dto.Az == "" || dto.LbName == "" || dto.OrgId == "" || dto.ProjectId == "" ||
		dto.Env == "" || dto.DeploymentId == 0 || !strings.HasPrefix(dto.HealthcheckPath, "/") || len(dto.Targets) == 0
}
