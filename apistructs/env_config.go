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

import "time"

// EnvConfig 环境变量配置
type EnvConfig struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	// ENV, FILE
	ConfigType string `json:"configType"`
	Comment    string `json:"comment"`
	Status     string `json:"status"`
	Source     string `json:"source"`
	Type       string `json:"type"` // dice-file/kv
	Encrypt    bool   `json:"encrypt"`
	// Operations 配置项操作，若为 nil，则使用默认配置: canDownload=false, canEdit=true, canDelete=true
	Operations *PipelineCmsConfigOperations `json:"operations"`
	CreateTime time.Time                    `json:"createTime,omitempty"`
	UpdateTime time.Time                    `json:"updateTime,omitempty"`
}

// EnvConfigAddOrUpdateRequest 配置新增/更新请求 POST /api/config
type EnvConfigAddOrUpdateRequest struct {
	Configs []EnvConfig `json:"configs"`
}

// EnvConfigFetchRequest namespace 配置获取请求
type EnvConfigFetchRequest struct {
	Namespace string // required
	Decrypt   bool   // optional, default false

	AutoCreateIfNotExist bool                   // optional, default false
	CreateReq            NamespaceCreateRequest // 当 AutoCreateIfNotExist == true 时需要
}

// EnvConfigFetchResponse namespace 配置获取响应
type EnvConfigFetchResponse struct {
	Header
	Data []EnvConfig `json:"data"`
}

// EnvConfigPublishResponse 发布配置
type EnvConfigPublishResponse struct {
	Header
}

// EnvMultiConfigFetchRequest 获取多个 namespace 配置请求
type EnvMultiConfigFetchRequest struct {
	NamespaceParams []NamespaceParam `json:"namespaceParams"`
}

// NamespaceParam namespace 参数信息
type NamespaceParam struct {
	NamespaceName string `json:"namespace_name"`
	Decrypt       bool   `json:"decrypt"`
}

// EnvMultiConfigFetchResponse 多个 namespace 配置响应
type EnvMultiConfigFetchResponse struct {
	Header
	Data map[string][]EnvConfig `json:"data"`
}
