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

package apistructs

import "time"

// ProjectCreateRequest POST /api/projects 创建项目请求结构
type ProjectCreateRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Logo        string `json:"logo"`

	// 项目级别的dd回调地址
	DdHook string `json:"ddHook"`
	Desc   string `json:"desc"`

	// 创建者的用户id
	Creator string `json:"creator"` // TODO deprecated

	// 组织id
	OrgID       uint64 `json:"orgId"`
	ClusterID   uint64 `json:"clusterId"`   // TODO deprecated
	ClusterName string `json:"clusterName"` // TODO deprecated

	// 项目各环境集群配置
	ClusterConfig map[string]string `json:"clusterConfig"`
	// 项目回滚点配置
	RollbackConfig map[string]int `json:"rollbackConfig"`
	// +required 单位: c
	CpuQuota float64 `json:"cpuQuota"`
	// +required 单位: GB
	MemQuota float64 `json:"memQuota"`
	// +required 项目模版
	Template ProjectTemplate `json:"template"`
}

// ProjectCreateResponse POST /api/projects 创建项目响应结构
type ProjectCreateResponse struct {
	Header
	Data uint64 `json:"data"`
}

// ProjectTemplate 项目模版
type ProjectTemplate string

const (
	DevopsTemplate ProjectTemplate = "DevOps"
)

// GetProjectFunctionsByTemplate 根据项目模版获取对应的项目功能
func (pt ProjectTemplate) GetProjectFunctionsByTemplate() map[ProjectFunction]bool {
	switch pt {
	case DevopsTemplate:
		return map[ProjectFunction]bool{PrjCooperativeFunc: true, PrjTestManagementFunc: true, PrjCodeQualityFunc: true,
			PrjCodeBaseFunc: true, PrjBranchRuleFunc: true, PrjCICIDFunc: true, PrjProductLibManagementFunc: true,
			PrjNotifyFunc: true}
	}

	return nil
}

// ProjectUpdateRequest PUT /api/projects/{projectId} 更新项目请求结构
type ProjectUpdateRequest struct {
	ProjectID uint64            `json:"-" path:"projectId"`
	Body      ProjectUpdateBody `json:"body"`
}

// ProjectUpdateBody 更新项目请求body
type ProjectUpdateBody struct {
	// 路径上有可以不传
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"` // TODO 废弃displayName字段
	Logo        string `json:"logo"`
	Desc        string `json:"desc"`
	DdHook      string `json:"ddHook"`
	IsPublic    bool   `json:"isPublic"` // 是否公开项目

	// 项目各环境集群配置
	ClusterConfig map[string]string `json:"clusterConfig"`

	// 项目回滚点配置
	RollbackConfig map[string]int `json:"rollbackConfig"`

	// +required 单位: c
	CpuQuota float64 `json:"cpuQuota"`
	// +required 单位: GB
	MemQuota float64 `json:"memQuota"`
}

// ProjectUpdateResponse PUT /api/projects/{projectId} 更新项目响应结构
type ProjectUpdateResponse struct {
	Header
	Data interface{} `json:"data"`
}

// ProjectDeleteRequest DELETE /api/projects/{projectId} 删除项目请求结构
type ProjectDeleteRequest struct {
	ProjectID uint64 `path:"projectId"`
}

// ProjectDeleteResponse DELETE /api/projects/{projectId} 删除项目响应结构
type ProjectDeleteResponse struct {
	Header
	Data ProjectDTO `json:"data"`
}

// ProjectDetailRequest GET /api/projects/{projectIdOrName} 项目详情请求结构
type ProjectDetailRequest struct {
	// 支持项目id/项目名查询
	ProjectIDOrName string `path:"projectIdOrName"`

	// 当传入projectName时，需要传入orgId或orgName
	OrgID uint64 `query:"orgId"`

	// 当传入projectName时，需要传入orgId或orgName
	OrgName uint64 `query:"orgName"`
}

// ProjectDetailResponse GET /api/projects/{projectIdOrName} 项目详情响应结构
// 由于与删除project时产生审计事件所需要的返回一样，所以删除project时也用这个接收返回
type ProjectDetailResponse struct {
	Header
	Data ProjectDTO `json:"data"`
}

// ProjectListRequest GET /api/projects 查询项目请求
type ProjectListRequest struct {
	OrgID uint64 `query:"orgId"`

	// 对项目名进行like查询
	Query string `query:"q"`
	Name  string `query:"name"` //project name

	// 排序支持activeTime,memQuota和cpuQuota
	OrderBy string `query:"orderBy"`
	// 是否升序
	Asc bool `query:"asc"`
	// 是否只展示已加入的项目
	Joined   bool `query:"joined"` // TODO refactor
	PageNo   int  `query:"pageNo"`
	PageSize int  `query:"pageSize"`
	// 是否只显示公开项目
	IsPublic bool `query:"isPublic"`
}

// ProjectListResponse GET /api/projects 查询项目响应
type ProjectListResponse struct {
	Header
	Data PagingProjectDTO `json:"data"`
}

// PagingProjectDTO 查询项目响应Body
type PagingProjectDTO struct {
	Total int          `json:"total"`
	List  []ProjectDTO `json:"list"`
}

// ProjectDTO 项目结构
type ProjectDTO struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	DDHook      string `json:"ddHook"`
	OrgID       uint64 `json:"orgId"`
	Creator     string `json:"creator"`
	Logo        string `json:"logo"`
	Desc        string `json:"desc"`

	// 项目所有者
	Owners []string `json:"owners"`
	// 项目活跃时间
	ActiveTime string `json:"activeTime"`
	// 用户是否已加入项目
	Joined bool `json:"joined"`

	// 当前用户是否可以解封该 project (目前只有 /api/projects/actions/list-my-projects api 有这个值)
	CanUnblock *bool `json:"canUnblock"`
	// 解封状态: unblocking | unblocked (目前只有 /api/projects/actions/list-my-projects api 有这个值)
	BlockStatus string `json:"blockStatus"`

	// 当前用户是否可以管理该 project (目前只有 /api/projects/actions/list-my-projects api 有这个值)
	CanManage bool `json:"CanManage"`
	IsPublic  bool `json:"isPublic"`

	// 项目统计信息
	Stats ProjectStats `json:"stats"`

	// 项目资源使用
	ProjectResourceUsage

	// 项目各环境集群配置
	ClusterConfig  map[string]string `json:"clusterConfig"`
	RollbackConfig map[string]int    `json:"rollbackConfig"`
	CpuQuota       float64           `json:"cpuQuota"`
	MemQuota       float64           `json:"memQuota"`

	// 项目创建时间
	CreatedAt time.Time `json:"createdAt"`

	// 项目更新时间
	UpdatedAt time.Time `json:"updatedAt"`

	// Project type
	Type string `json:"type"`
}

// ProjectResourceUsage 项目资源使用
type ProjectResourceUsage struct {
	CpuServiceUsed float64 `json:"cpuServiceUsed"`
	MemServiceUsed float64 `json:"memServiceUsed"`
	CpuAddonUsed   float64 `json:"cpuAddonUsed"`
	MemAddonUsed   float64 `json:"memAddonUsed"`
}

// ProjectFillQuotaResponse 项目填充配额响应
type ProjectFillQuotaResponse struct {
	Header
	Data string `json:"data"`
}

// ProjectStats 项目统计
type ProjectStats struct {
	// 应用数
	CountApplications int `json:"countApplications"`
	// 总成员数
	CountMembers int `json:"countMembers"`

	// new states
	// 总应用数
	TotalApplicationsCount int `json:"totalApplicationsCount"`
	// 总成员数
	TotalMembersCount int `json:"totalMembersCount"`
	// 总迭代数
	TotalIterationsCount int `json:"totalIterationsCount"`
	// 进行中的迭代数
	RunningIterationsCount int `json:"runningIterationsCount"`
	// 规划中的迭代数
	PlanningIterationsCount int `json:"planningIterationsCount"`
	// 总预计工时
	TotalManHourCount float64 `json:"totalManHourCount"`
	// 总已记录工时
	UsedManHourCount float64 `json:"usedManHourCount"`
	// 总规划工时
	PlanningManHourCount float64 `json:"planningManHourCount"`
	// 已解决bug数
	DoneBugCount int64 `json:"doneBugCount"`
	// 总bug数
	TotalBugCount int64 `json:"totalBugCount"`
	// bug解决率·
	DoneBugPercent float64 `json:"doneBugPercent"`
}

// ProjectFunction 项目功能
type ProjectFunction string

const (
	PrjCooperativeFunc          ProjectFunction = "projectCooperative"   // 项目协同
	PrjTestManagementFunc       ProjectFunction = "testManagement"       // 测试管理
	PrjCodeQualityFunc          ProjectFunction = "codeQuality"          // 代码质量
	PrjCodeBaseFunc             ProjectFunction = "codeBase"             // 代码仓库
	PrjBranchRuleFunc           ProjectFunction = "branchRule"           // 分支规则
	PrjCICIDFunc                ProjectFunction = "cicd"                 // 持续集成
	PrjProductLibManagementFunc ProjectFunction = "productLibManagement" // 制品库管理
	PrjNotifyFunc               ProjectFunction = "Projectnotify"        // 通知通知组
)

// ProjectFunctionSetRequest 项目功能开关设置请求
type ProjectFunctionSetRequest struct {
	ProjectID       uint64                   `json:"projectId"`       // 项目id，必传参数
	ProjectFunction map[ProjectFunction]bool `json:"projectFunction"` // 项目功能开关配置
}

// ProjectFunctionSetResponse 项目功能开关设置响应
type ProjectFunctionSetResponse struct {
	Header
	Data string `json:"data"`
}

// ProjectActiveTimeUpdateRequest 项目活跃时间更新请求
type ProjectActiveTimeUpdateRequest struct {
	ProjectID  uint64    `json:"projectId"`  // 项目id，必传参数
	ActiveTime time.Time `json:"activeTime"` // 活跃时间
}

// ProjectActiveTimeUpdateResponse 项目活跃时间更新响应
type ProjectActiveTimeUpdateResponse struct {
	Header
	Data string `json:"data"`
}

// ProjectNameSpaceInfoResponse 项目级命名空间响应
type ProjectNameSpaceInfoResponse struct {
	Header
	Data ProjectNameSpaceInfo `json:"data"`
}

// ProjectNameSpaceInfo 项目级命名空间信息
type ProjectNameSpaceInfo struct {
	Enabled    bool              `json:"enabled"`
	Namespaces map[string]string `json:"namespaces"`
}

type MyProjectIDsResponse struct {
	Header
	Data []uint64 `json:"data"`
}

type GetProjectIDListByStatesRequest struct {
	StateReq IssuePagingRequest `json:"stateReq"`
	ProIDs   []uint64           `json:"proIDs"`
}

type GetProjectIDListByStatesResponse struct {
	Header

	Data GetProjectIDListByStatesData `json:"data"`
}

type GetProjectIDListByStatesData struct {
	Total int          `json:"total"`
	List  []ProjectDTO `json:"list"`
}

type GetAllProjectsResponse struct {
	Header

	Data []ProjectDTO `json:"data"`
}
