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

// DomainListRequest 域名查询请求
type DomainListRequest struct {
	// 应用实例 ID
	RuntimeID uint64 `path:"runtimeId"`
}

// DomainListResponse 域名查询响应
type DomainListResponse struct {
	Header
	Data DomainGroup `json:"data"`
}

// DomainUpdateRequest 域名更新请求
type DomainUpdateRequest struct {
	// 应用实例 ID
	RuntimeID uint64 `path:"runtimeId"`
	Body      DomainGroup
}

// DomainUpdateResponse 域名更新响应
type DomainUpdateResponse struct {
	Header
	Data DomainGroup `json:"data"`
}

type Domain struct {
	AppName      string `json:"appName"`
	DomainID     uint64 `json:"domainId"` // Deprecated
	Domain       string `json:"domain"`
	DomainType   string `json:"domainType"`
	CustomDomain string `json:"customDomain"`
	RootDomain   string `json:"rootDomain"` // Deprecated
	UseHttps     bool   `json:"useHttps"`   // Deprecated
}

type DomainGroup = map[string][]*Domain
