// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package apistructs api request/response结构体定义
package apistructs

import (
	"net/url"
	"strconv"
	"time"
)

// ResourceType release可管理的资源类型
type ResourceType string

const (
	// MigrationResourceKey release出来的migration信息key，作用于服务的release->resource信息中
	MigrationResourceKey string = "dice-migration"
)

const (
	// ResourceTypeDiceYml 资源类型为dice.yml
	ResourceTypeDiceYml ResourceType = "diceyml"
	// ResourceTypeAddonYml 资源类型为addon.yml
	ResourceTypeAddonYml ResourceType = "addonyml"
	// ResourceTypeBinary 资源类型为二进制可执行文件
	ResourceTypeBinary ResourceType = "binary"
	// ResourceTypeScript 资源类型为可执行的脚本文件, eg: shell/python/ruby, etc
	ResourceTypeScript ResourceType = "script"
	// ResourceTypeSQL 资源类型为可执行的sql文本
	ResourceTypeSQL ResourceType = "sql"
	// ResourceTypeDataSet 资源类型为数据文本文件
	ResourceTypeDataSet ResourceType = "data"
	// ResourceTypeAndroid android类型文件
	ResourceTypeAndroid ResourceType = "android"
	// ResourceTypeIOS ios类型文件
	ResourceTypeIOS ResourceType = "ios"
	// ResourceTypeMigration 资源类型为migration文件releaseID
	ResourceTypeMigration ResourceType = "migration"
	// ResourceTypeH5 h5类型的资源文件
	ResourceTypeH5 ResourceType = "h5"
)

// ReleaseCreateRequest 创建Release API(POST /api/releases)使用
type ReleaseCreateRequest struct {
	// ReleaseName 任意字符串，便于用户识别，最大长度255，必填
	ReleaseName string `json:"releaseName"`

	// Desc 详细描述此release功能, 选填
	Desc string `json:"desc,omitempty"`

	// Dice 资源类型为diceyml时, 存储dice.yml内容, 选填
	Dice string `json:"dice,omitempty"`

	// Addon addon注册时，release包含dice.yml与addon.yml，选填
	Addon string `json:"addon,omitempty"`

	// Labels 用于release分类，描述release类别，map类型, 最大长度1000, 选填
	Labels map[string]string `json:"labels,omitempty"`

	// Version 存储release版本信息, 同一企业同一项目同一应用下唯一，最大长度100，选填
	Version string `json:"version,omitempty"`

	// OrgID 企业标识符，描述release所属企业，选填
	OrgID int64 `json:"orgId,omitempty"`

	// ProjectID 项目标识符，描述release所属项目，选填
	ProjectID int64 `json:"projectId,omitempty"`

	// ApplicationID 应用标识符，描述release所属应用，选填
	ApplicationID int64 `json:"applicationId,omitempty"`

	// ProjectName 项目名称，描述release所属项目，选填
	ProjectName string `json:"projectName,omitempty"`

	// ApplicationName 应用名称，描述release所属应用，选填
	ApplicationName string `json:"applicationName,omitempty"`

	// UserID 用户标识符, 描述release所属用户，最大长度50，选填
	UserID string `json:"userId,omitempty"`

	// ClusterName 集群名称，描述release所属集群，最大长度80，选填
	ClusterName string `json:"clusterName,omitempty"`

	// Resources release包含的资源，包含类型、名称、资源存储路径, 为兼容现有diceyml，先选填
	Resources []ReleaseResource `json:"resources,omitempty"`

	// CrossCluster 跨集群
	CrossCluster bool `json:"crossCluster,omitempty"`
}

// ReleaseResource release资源结构
type ReleaseResource struct {
	// Type 资源类型, 必填

	// 资源类型
	Type ResourceType `json:"type"`
	// Name 资源名称，选填, eg: init.sql/upgrade.sql

	// 资源名称
	Name string `json:"name"`
	// URL 资源URL, 可直接wget获取到资源, 选填(当type为diceyml, 资源作为release的dice字段在mysql存储)

	// 资源URL, 可wget获取
	URL string `json:"url"`

	Meta map[string]interface{} `json:"meta"`
}

// ReleaseCreateResponse 创建 release API响应数据结构
type ReleaseCreateResponse struct {
	Header
	Data ReleaseCreateResponseData `json:"data"`
}

// ReleaseCreateResponseData 创建 release 实际返回数据
type ReleaseCreateResponseData struct {
	ReleaseID string `json:"releaseId"`
}

// ReleaseUpdateRequest 更新 release API(PUT /api/releases/{releaseId})使用
type ReleaseUpdateRequest struct {
	ReleaseID string                   `json:"-" path:"releaseId"`
	Body      ReleaseUpdateRequestData `json:"body"`
}

// ReleaseUpdateRequestData 更新 release 请求数据结构
type ReleaseUpdateRequestData struct {
	Version string `json:"version,omitempty"`
	Desc    string `json:"desc,omitempty"`
	// 以下信息主要为了version覆盖使用，找出之前的version清除

	// 企业标识
	OrgID int64 `json:"orgId"`

	// 项目Id
	ProjectID int64 `json:"projectId"`

	// 应用Id
	ApplicationID int64 `json:"applicationId"`
}

// ReleaseUpdateResponse 更新 release API 响应数据结构
type ReleaseUpdateResponse struct {
	Header
	Data string `json:"data"` // Update succ
}

// ReleaseReferenceUpdateRequest 更新Reference API(/api/releases/{releaseId}/reference/actions/change)使用
type ReleaseReferenceUpdateRequest struct {
	ReleaseID string `json:"-" path:"releaseId"`
	Increase  bool   `json:"increase"` // true:reference+1  false:reference-1
}

// ReleaseDeleteRequest 删除 release API(DELETE /api/releases/{releaseId})使用
type ReleaseDeleteRequest struct {
	ReleaseID string `json:"-" path:"releaseId"`
}

// ReleaseDeleteResponse 删除 release API响应数据结构
type ReleaseDeleteResponse struct {
	Header
	Data string `json:"data"` // Delete succ
}

// ReleaseGetRequest release详情 API(GET /api/releases/{releaseId})使用
type ReleaseGetRequest struct {
	ReleaseID string `json:"-" path:"releaseId"`
}

// ReleaseGetResponse release 详情API响应数据结构
type ReleaseGetResponse struct {
	Header
	Data ReleaseGetResponseData `json:"data"`
}

// ReleaseGetResponseData release 详情API实际返回数据
type ReleaseGetResponseData struct {
	ReleaseID   string            `json:"releaseId"`
	ReleaseName string            `json:"releaseName"`
	Diceyml     string            `json:"diceyml"`
	Desc        string            `json:"desc,omitempty"`
	Addon       string            `json:"addon,omitempty"`
	Resources   []ReleaseResource `json:"resources,omitempty"`
	Images      []string          `json:"images,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Version     string            `json:"version,omitempty"`

	// CrossCluster 是否可以跨集群
	CrossCluster bool `json:"crossCluster,omitempty"`

	// 当前被部署次数
	Reference int64 `json:"reference"`

	// 企业标识
	OrgID int64 `json:"orgId"`

	// 项目Id
	ProjectID int64 `json:"projectId"`

	// 应用Id
	ApplicationID int64 `json:"applicationId"`

	// 项目Name
	ProjectName string `json:"projectName"`

	// 应用Name
	ApplicationName string `json:"applicationName"`

	// 操作用户Id
	UserID string `json:"userId,omitempty"`

	// 集群名称
	ClusterName string    `json:"clusterName"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ReleaseListRequest release列表 API(GET /api/releases)使用
type ReleaseListRequest struct {
	// 查询参数，releaseId/releaseName/version
	Query string `json:"-" query:"q"` // 查询参数，可根据releaseId/releaseName/version模糊匹配

	// release 名字精确匹配
	ReleaseName string `json:"-" query:"releaseName"`

	// 集群名称
	Cluster string `json:"-" query:"cluster"`

	// 分支名
	Branch string `json:"-" query:"branchName"`

	// 只列出有 version 的 release
	IsVersion bool `json:"-" query:"isVersion"`

	// 跨集群
	CrossCluster *bool `json:"-" query:"crossCluster"`

	// 跨集群或指定集群
	CrossClusterOrSpecifyCluster *string `json:"-" query:"crossClusterOrSpecifyCluster"`

	// 应用Id
	ApplicationID int64 `json:"-" query:"applicationId"`

	// 项目ID
	ProjectID int64 `json:"-" query:"projectId"`

	// 开始时间, ms
	StartTime int64 `json:"-" query:"startTime"`

	// 结束时间,ms
	EndTime int64 `json:"-" query:"endTime"`

	// 分页大小,默认值20
	PageSize int64 `json:"-" query:"pageSize"`

	// 当前页号，默认值1
	PageNum int64 `json:"-" query:"pageNo"`
}

func (req ReleaseListRequest) ConvertToQueryParams() url.Values {
	values := make(url.Values)
	if req.Query != "" {
		values.Add("q", req.Query)
	}
	if req.ReleaseName != "" {
		values.Add("releaseName", req.ReleaseName)
	}
	if req.Cluster != "" {
		values.Add("cluster", req.Cluster)
	}
	if req.CrossCluster != nil {
		values.Add("crossCluster", strconv.FormatBool(*req.CrossCluster))
	}
	if req.CrossClusterOrSpecifyCluster != nil {
		values.Add("crossClusterOrSpecifyCluster", *req.CrossClusterOrSpecifyCluster)
	}
	if req.ApplicationID > 0 {
		values.Add("applicationId", strconv.FormatInt(req.ApplicationID, 10))
	}
	if req.ProjectID > 0 {
		values.Add("projectId", strconv.FormatInt(req.ProjectID, 10))
	}
	if req.StartTime > 0 {
		values.Add("startTime", strconv.FormatInt(req.StartTime, 10))
	}
	if req.EndTime > 0 {
		values.Add("endTime", strconv.FormatInt(req.EndTime, 10))
	}
	if req.PageSize > 0 {
		values.Add("pageSize", strconv.FormatInt(req.PageSize, 10))
	}
	if req.PageNum > 0 {
		values.Add("pageNo", strconv.FormatInt(req.PageNum, 10))
	}
	if req.IsVersion {
		values.Add("isVersion", "true")
	}
	if req.Branch != "" {
		values.Add("branchName", req.Branch)
	}
	return values
}

// ReleaseListResponse release 列表API响应数据结构
type ReleaseListResponse struct {
	Header
	Data ReleaseListResponseData `json:"data"`
}

// ReleaseListResponseData release 列表API实际响应数据
type ReleaseListResponseData struct {
	// release总数，用于分页
	Total    int64                    `json:"total"`
	Releases []ReleaseGetResponseData `json:"list"`
}

// ReleaseNameListRequest releaseName列表请求
type ReleaseNameListRequest struct {
	// 应用Id
	ApplicationID int64 `json:"-" query:"applicationId"`
}

// ReleaseNameListResponse releaseName列表响应
type ReleaseNameListResponse struct {
	Header
	Data []string `json:"data"`
}

// ReleasePullRequest release dice.yml内容获取API(GET /api/releases/<releaseId>/actions/pull)
type ReleasePullRequest struct {
	ReleaseID string `json:"-" path:"releaseId"`
}

// ReleaseGetDiceYmlRequest release 请求 dice.yml 格式
type ReleaseGetDiceYmlRequest struct {
	ReleaseID string `json:"-" path:"releaseId"`
}

// ReleaseEventData release 事件格式
type ReleaseEventData ReleaseGetResponseData
