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

// Package apierrors 定义了错误列表
package apierrors

import (
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

const (
	MissingRequestBody = "request body"
	MissingOrgID       = "orgID"
	MissingAssetID     = "assetID"
)

var (
	CreateAPIAsset  = err("ErrCreateAPIAsset", "创建 API 资料失败")
	GetAPIAsset     = err("ErrGetAPIAsset", "查询 API 资料失败")
	UpdateAPIAsset  = err("ErrUpdateAPIAsset", "修改 API 资料失败")
	PagingAPIAssets = err("ErrPagingAPIAssets", "分页查询 API 资料失败")
	DeleteAPIAsset  = err("ErrDeleteAPIAsset", "删除 API 资料失败")

	CreateAPIAssetVersion  = err("ErrCreateAPIAssetVersion", "创建 API 资料版本失败")
	PagingAPIAssetVersions = err("ErrPagingAPIAssetVersions", "获取 API 资料版本列表失败")
	GetAPIAssetVersion     = err("ErrGetAPIAssetVersion", "查询 API 资料版本详情失败")
	UpdateAssetVersion     = err("ErrUpdateAssetVersion", "修改 API 资料版本失败")
	DeleteAPIAssetVersion  = err("ErrDeleteAPIAssetVersion", "删除 API 资料详情失败")

	ValidateAPISpec        = err("ErrValidateAPISpec", "校验 API Spec 失败")
	GetAPIAssetVersionSpec = err("GetAPIAssetVersionSpec", "查询 API 资料版本 Spec 失败")

	ValidateAPIInstance = err("ErrValidateAPIInstance", "校验 API 实例失败")
	CreateAPIInstance   = err("ErrCreateAPIInstance", "创建 API 实例失败")
	ListAPIInstances    = err("ListAPIInstances", "查询 API 实例列表失败")

	PagingSwaggerVersion = err("ErrPagingSwaggerVersion", "查询版本树失败")

	CreateInstantiation = err("ErrCreateInstantiation", "实例化失败")
	GetInstantiations   = err("ErrGetInstantiations", "查询实例化记录失败")
	UpdateInstantiation = err("ErrUpdateInstantiation", "更新实例化记录失败")
	ListRuntimeServices = err("ErrListRuntimeServices", "列举应用下 Runtime Service 失败")

	DownloadSpecText = err("ErrDownloadSpecText", "下载 Swagger 文本失败")

	CreateClient       = err("ErrCreateClient", "创建客户端失败")
	ListClients        = err("ErrGetClients", "查询客户端失败")
	GetClient          = err("ErrGetClient", "查询客户端详情")
	ListSwaggerClients = err("ErrListSwaggerClients", "查询 SwaggerVersion 下的客户端列表失败")
	UpdateClient       = err("ErrUpdateClient", "修改客户端失败")
	DeleteClient       = err("ErrDeleteClient", "删除客户端失败")

	CreateContract      = err("ErrCreateContract", "创建合约失败")
	ListContracts       = err("ErrListContracts", "查询合约列表失败")
	GetContract         = err("ErrGetContract", "查询合约详情失败")
	ListContractRecords = err("ErrGetContractRecords", "查询合约操作记录失败")
	UpdateContract      = err("ErrUpdateContract", "更新合约失败")
	DeleteContract      = err("ErrDeleteContract", "删除调用申请记录失败")

	CreateAccess = err("ErrCreateAccess", "创建访问管理条目失败")
	ListAccess   = err("ErrListAccess", "查询访问管理列表失败")
	GetAccess    = err("ErrGetAccess", "查询访问管理条目失败")
	DeleteAccess = err("ErrDeleteAccess", "删除访问管理条目失败")
	UpdateAccess = err("ErrUpdateAccess", "更新访问管理条目失败")

	ListAPIGateways = err("ErrListAPIGateways", "获取 API Gateway 列表失败")

	AttemptExecuteAPITest = err("ErrAttemptExecuteAPITTest", "执行接口测试失败")

	ListSLAs  = err("ErrListSLAs", "查询 SLA 列表失败")
	CreateSLA = err("ErrCreateSLAs", "创建 SLA 失败")
	GetSLA    = err("ErrGetSLA", "查询 SLA 失败")
	DeleteSLA = err("ErrDeleteSLA", "删除 SLA 失败")
	UpdateSLA = err("ErrUpdateSLA", "修改 SLA 失败")

	CreateNode        = err("ErrCreateNode", "创建节点失败")
	DeleteNode        = err("ErrDeleteNode", "删除节点失败")
	UpdateNode        = err("ErrUpdateNode", "更新节点失败")
	MoveNode          = err("ErrMoveNode", "移动节点失败")
	CopyNode          = err("ErrCopyNode", "复制节点失败")
	ListChildrenNodes = err("ErrListChildrenNodes", "列举子节点失败")
	GetNodeDetail     = err("ErrGetNodeDetail", "查询节点详情失败")
	GetNodeInfo       = err("ErrGetNodeInfo", "查询 Gittar 节点信息失败")

	WsUpgrade = err("ErrWsUpgrade", "建立连接失败")

	ListSchemas = err("ErrListSchemas", "查询 schema 列表失败")

	SearchOperations = err("ErrSearchOperations", "搜索失败")
	GetOperation     = err("GetOperation", "查询接口详情失败")
)

func err(template, defaultValue string) *errorresp.APIError {
	return errorresp.New(errorresp.WithTemplateMessage(template, defaultValue))
}
