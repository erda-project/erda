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

// Package apierrors api的错误返回信息
package apierrors

import (
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

var (
	// ErrReleaseCallback 回调函数错误信息
	ErrReleaseCallback    = err("ErrReleaseCallback", "release gittar hook回调失败")
	ErrRepoMrCallback     = err("ErrRepoMrCallback", "repo mr hook回调失败")
	ErrRepoBranchCallback = err("ErrRepoBranchCallback", "repo branch hook回调失败")
	ErrIssueCallback      = err("ErrIssueCallback", "issue callback hook 回调失败")

	ErrDealCDPCallback = err("ErrDealCDPCallback", "cdp hook回调失败")

	ErrGetCICDTaskLog      = err("ErrGetCICDTaskLog", "查询 CICD 任务日志失败")
	ErrDownloadCICDTaskLog = err("ErrDownloadCICDTaskLog", "下载 CICD 任务日志失败")

	ErrCheckPermission = err("ErrCheckPermission", "权限校验失败")

	ErrGetUser    = err("ErrGetUser", "获取用户信息失败，请登录")
	ErrGetApp     = err("ErrGetApp", "获取应用信息失败")
	ErrGetProject = err("ErrGetProject", "获取项目失败")

	ErrCreatePipeline         = err("ErrCreatePipeline", "创建流水线失败")
	ErrListPipeline           = err("ErrListPipeline", "获取流水线列表失败")
	ErrListPipelineYml        = err("ErrListPipelineYml", "获取流水线配置列表失败")
	ErrListInvokedCombos      = err("ErrListInvokedCombos", "获取流水线侧边栏信息失败")
	ErrFetchPipelineByAppInfo = err("ErrFetchPipelineByAppInfo", "获取流水线信息失败")
	ErrGetPipeline            = err("ErrGetPipeline", "获取流水线失败")
	ErrGetPipelineBranchRule  = err("ErrGetPipeline", "获取流水线对应分支规则失败")
	ErrOperatePipeline        = err("ErrOperatePipeline", "操作流水线失败")
	ErrRunPipeline            = err("ErrRunPipeline", "启动流水线失败")
	ErrCancelPipeline         = err("ErrCancelPipeline", "取消流水线失败")
	ErrRerunFailedPipeline    = err("ErrRerunFailedPipeline", "重试失败节点失败")
	ErrRerunPipeline          = err("ErrRerunPipeline", "重试全流程失败")
	ErrCreateCheckRun         = err("ErrCreateCheckRun", "创建流水线失败")

	ErrFetchConfigNamespace  = err("ErrFetchConfigNamespace", "获取私有配置命名空间失败")
	ErrMakeConfigNamespace   = err("ErrMakeConfigNamespace", "创建私有配置命名空间失败")
	ErrGetBranchWorkspaceMap = err("ErrGetBranchWorkspaceMap", "获取分支与环境映射关系失败")
	ErrGetGittarTag          = err("ErrGetGittarTag", "获取Tag信息失败")
	ErrGetGittarBranch       = err("ErrGetGittarBranch", "获取Branch信息失败")
	ErrGetGittarCommit       = err("ErrGetGittarCommit", "获取Commit信息失败")
	ErrGetGittarRepoFile     = err("ErrGetGittarRepoFile", "获取仓库文件失败")

	ErrCreatePipelineCron = err("ErrCreatePipelineCron", "创建流水线定时配置失败")
	ErrPagingPipelineCron = err("ErrPagingPipelineCron", "分页获取流水线定时配置失败")
	ErrStartPipelineCron  = err("ErrStartPipelineCron", "启动定时流水线失败")
	ErrStopPipelineCron   = err("ErrStopPipelineCron", "停止定时流水线失败")
	ErrDeletePipelineCron = err("ErrDeletePipelineCron", "删除流水线定时配置失败")

	ErrAddEnvConfig          = err("ErrAddEnvConfig", "添加环境变量配置失败")
	ErrUpdateEnvConfig       = err("ErrUpdateEnvConfig", "更新环境变量配置失败")
	ErrDeleteEnvConfig       = err("ErrDeleteEnvConfig", "删除环境变量配置失败")
	ErrGetEnvConfig          = err("ErrGetEnvConfig", "获取环境变量配置失败")
	ErrGetNamespaceEnvConfig = err("ErrGetNamespaceEnvConfig", "获取指定namespace环境变量配置失败")

	ErrDeletePipelineCmsNs              = err("ErrDeletePipelineCmsNs", "删除流水线配置管理命名空间失败")
	ErrCreateOrUpdatePipelineCmsConfigs = err("ErrUpdatePipelineCmsConfigs", "创建或更新流水线配置管理配置失败")
	ErrDeletePipelineCmsConfigs         = err("ErrDeletePipelineCmsConfigs", "删除流水线配置管理配置失败")
	ErrGetPipelineCmsConfigs            = err("ErrGetPipelineCmsConfigs", "查询流水线配置管理配置失败")

	ErrGetSnippetYaml = err("ErrGetSnippetYaml", "获取 snippet yml 失败")

	ErrCreateGittarFileTreeNode        = err("ErrCreateGittarFileTreeNode", "创建应用目录树节点失败")
	ErrDeleteGittarFileTreeNode        = err("ErrDeleteGittarFileTreeNode", "删除应用目录树节点失败")
	ErrUpdateGittarSetBasicInfo        = err("ErrUpdateGittarSetBasicInfo", "更新应用目录树节点基础信息失败")
	ErrMoveGittarFileTreeNode          = err("ErrMoveGittarFileTreeNode", "移动应用目录树节点失败")
	ErrCopyGittarFileTreeNode          = err("ErrCopyGittarFileTreeNode", "复制应用目录树节点失败")
	ErrGetGittarFileTreeNode           = err("ErrGetGittarFileTreeNode", "查询应用目录树节点详情失败")
	ErrListGittarFileTreeNodes         = err("ErrListGittarFileTreeNodes", "查询应用目录树节点列表失败")
	ErrListGittarFileTreeNodeHistory   = err("ErrListGittarFileTreeNodeHistory", "查询应用目录树节点历史列表失败")
	ErrFuzzySearchGittarFileTreeNodes  = err("ErrFuzzySearchGittarFileTreeNodes", "模糊搜索应用目录树节点失败")
	ErrSaveGittarFileTreeNodePipeline  = err("ErrSaveGittarFileTreeNodePipeline", "保存应用流水线失败")
	ErrFindGittarFileTreeNodeAncestors = err("ErrFindGittarFileTreeNodeAncestors", "应用目录树节点寻祖失败")
)

func err(template, defaultValue string) *errorresp.APIError {
	return errorresp.New(errorresp.WithTemplateMessage(template, defaultValue))
}
