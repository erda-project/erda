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

package apistructs

// AppWorkspaceReleasesGetRequest 查询应用某个环境所有可部署的 release 请求
type AppWorkspaceReleasesGetRequest struct {
	AppID     uint64        `schema:"appID,required"`
	Workspace DiceWorkspace `schema:"workspace,required"`
}

// AppWorkspaceReleasesGetResponse 查询应用某个环境所有可部署的 release 响应
type AppWorkspaceReleasesGetResponse struct {
	Header
	Data AppWorkspaceReleasesGetResponseData `json:"data,omitempty"`
}

// AppWorkspaceReleasesGetResponseData map key: branch, map value: paging releases
type AppWorkspaceReleasesGetResponseData map[string]*ReleaseListResponseData
