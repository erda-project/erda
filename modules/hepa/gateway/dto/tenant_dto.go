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

type TenantDto struct {
	Id              string `json:"id"`
	TenantGroup     string `json:"tenantGroup"`
	Az              string `json:"az"`
	Env             string `json:"env"`
	ProjectId       string `json:"projectId"`
	ProjectName     string `json:"projectName"`
	AdminAddr       string `json:"adminAddr"`
	GatewayEndpoint string `json:"gatewayEndpoint"`
	InnerAddr       string `json:"innerAddr"`
	ServiceName     string `json:"serviceName"`
	InstanceId      string `json:"instanceId"`
}
