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

package apierrors

import (
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
)

var (
	ErrCreateRelease                   = err("ErrCreateRelease", "创建Release失败")
	ErrUpdateRelease                   = err("ErrUpdateRelease", "更新Release失败")
	ErrDeleteRelease                   = err("ErrDeleteRelease", "删除Release失败")
	ErrGetRelease                      = err("ErrGetRelease", "获取Release失败")
	ErrListRelease                     = err("ErrListRelease", "获取Release列表失败")
	ErrGetYAML                         = err("ErrGetYAML", "获取Dice YAML失败")
	ErrGetIosPlist                     = err("ErrGetIosPlist", "获取Ios Plist文件失败")
	ErrCreateImage                     = err("ErrCreateImage", "添加镜像失败")
	ErrUpdateImage                     = err("ErrUpdateImage", "更新镜像失败")
	ErrDeleteImage                     = err("ErrDeleteImage", "删除镜像失败")
	ErrGetImage                        = err("ErrGetImage", "获取镜像失败")
	ErrListImage                       = err("ErrListImage", "获取镜像列表失败")
	ErrCreateExtension                 = err("ErrCreateExtension", "添加扩展失败")
	ErrQueryExtension                  = err("ErrQueryExtension", "查询扩展失败")
	ErrCreateExtensionVersion          = err("ErrCreateExtensionVersion", "添加扩展版本失败")
	ErrQueryExtensionVersion           = err("ErrQueryExtensionVersion", "查询扩展版本失败")
	ErrCreatePipelineTemplate          = err("ErrCreateTemplate", "添加模板失败")
	ErrQueryPipelineTemplate           = err("ErrQueryTemplate", "查询模板失败")
	ErrCreatePipelineTemplateVersion   = err("ErrCreateTemplateVersion", "添加模板版本失败")
	ErrQueryPipelineTemplateVersion    = err("ErrQueryTemplateVersion", "查询模板版本失败")
	ErrRenderPipelineTemplate          = err("ErrQueryTemplateVersion", "模板渲染失败")
	ErrQueryPublishItem                = err("ErrQueryPublishItem", "查询发布内容失败")
	ErrCreatePublishItem               = err("ErrCreatePublishItem", "创建发布内容失败")
	ErrGetPublishItem                  = err("ErrGetPublishItem", "获取发布内容详情失败")
	ErrUpdatePublishItem               = err("ErrUpdatePublishItem", "更新发布内容失败")
	ErrDeletePublishItem               = err("ErrDeletePublishItem", "删除发布内容失败")
	ErrQueryPublishItemVersion         = err("ErrQueryPublishItemVersion", "查询发布版本失败")
	ErrCreatePublishItemVersion        = err("ErrCreatePublishItemVersion", "创建发布版本失败")
	ErrCreateOffLinePublishItemVersion = err("ErrCreateOffLinePublishItemVersion", "推送离线包失败")
	ErrUpdatePublishItemVersion        = err("ErrUpdatePublishItemVersion", "更新发布版本失败")
	ErrDeletePublishItemVersion        = err("ErrDeletePublishItemVersion", "删除发布版本失败")
	ErrSetPublishItemVersionStatus     = err("ErrSetPublishItemVersionStatus", "更新版本状态失败")
	ErrGetMonitorKeys                  = err("ErrGetMonitorKeys", "获取监控key失败")

	QuerySnippetYml = err("ErrQuerySnippetYml", "查询嵌套Yml文件失败")

	ErrCreateBlacklist = err("ErrCreateBlacklist", "添加黑名单失败")
	ErrGetBlacklist    = err("ErrCreateBlacklist", "查询黑名单失败")
	ErrDeleteBlacklist = err("ErrDeleteBlacklist", "删除黑名单失败")

	ErrCreateEraselist = err("ErrCreateEraselist", "添加数据擦除失败")

	ErrSecurity    = err("ErrSecurity", "security error")
	ErrUpdateErase = err("ErrUpdateErase", "request error")

	ErrSratisticsErrList       = err("ErrSratisticsErrList", "获取错误列表失败")
	ErrSratisticsErrTrend      = err("ErrSratisticsErrTrend", "获取错误趋势失败")
	ErrSratisticsTotalTrend    = err("ErrSratisticsTotalTrend", "获取整体趋势失败")
	ErrSratisticsVersionDetail = err("ErrSratisticsVersionDetail", "获取版本详情明细数据失败")
	ErrSratisticsChannelDetail = err("ErrSratisticsChannelDetail", "获取渠道详情明细数据失败")

	ErrCrashRateList = err("ErrCrashRateList", "获取崩溃率失败")
)

func err(template, defaultValue string) *errorresp.APIError {
	return errorresp.New(errorresp.WithTemplateMessage(template, defaultValue))
}
