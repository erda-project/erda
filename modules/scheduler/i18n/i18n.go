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

package i18n

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func InitI18N() {
	message.SetString(language.SimplifiedChinese, "ImagePullFailed", "拉取镜像失败")
	message.SetString(language.SimplifiedChinese, "Unschedulable", "调度失败")
	message.SetString(language.SimplifiedChinese, "InsufficientResources", "资源不足")
	message.SetString(language.SimplifiedChinese, "ProbeFailed", "健康检查失败")
	message.SetString(language.SimplifiedChinese, "ContainerCannotRun", "容器无法启动")
}
