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

// CloudAccountInfo 云账号信息
type CloudAccountInfo struct {
	ID            int64  `json:"accoundID"`
	CloudProvider string `json:"cloudProvider"`
	Name          string `json:"name"`
	OrgID         int64  `json:"orgID"`
}

// CloudAccountCreateRequest POST /api/cloud-accounts 创建账号请求结构
type CloudAccountCreateRequest struct {
	CloudProvider   string `json:"cloudProvider"`
	Name            string `json:"name"`
	AccessKeyID     string `json:"accessKeyID"`
	AccessKeySecret string `json:"accessKeySecret"`
}

// CloudAccountCreateResponse POST /api/cloud-account 创建账号返回结构
type CloudAccountCreateResponse struct {
	Header
	Data CloudAccountInfo `json:"data"`
}

// CloudAccountUpdateRequest PUT /api/cloud-accounts/{accountID} 更新云账号信息
type CloudAccountUpdateRequest struct {
	AccountID       uint64 `json:"-" path:"accountID"`
	CloudProvider   string `json:"cloudProvider"`
	Name            string `json:"name"`
	AccessKeyID     string `json:"accessKeyID"`
	AccessKeySecret string `json:"accessKeySecret"`
}

// CloudAccountUpdateResponse PUT /api/cloud-accounts/{accountID} 更新云账号响应结构
type CloudAccountUpdateResponse struct {
	Header
	Data CloudAccountInfo `json:"data"`
}

// CloudAccountDeleteResponse DELETE /api/cloud-accounts/{acountId} 删除云账号响应结构
type CloudAccountDeleteResponse struct {
	Header
	Data uint64 `json:"data"`
}

// CloudAccountListResponse GET /api/cloud-accounts 获取云账号列表
type CloudAccountListResponse struct {
	Header
	Data []CloudAccountInfo `json:"data"`
}

type CloudAccountAllInfo struct {
	CloudAccountInfo
	AccessKeyID     string `json:"accessKeyID"`
	AccessKeySecret string `json:"accessKeySecret"`
}

// CloudAccountGetResponse GET /api/cloud-accounts 获取云账号列表, 仅内网使用
type CloudAccountGetResponse struct {
	Header
	Data CloudAccountAllInfo `json:"data"`
}
