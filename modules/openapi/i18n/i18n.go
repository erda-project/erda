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
	"net/http"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/erda-project/erda/pkg/strutil"
)

func InitI18N() {
	// issue state
	message.SetString(language.SimplifiedChinese, "OPEN", "待处理")
	message.SetString(language.SimplifiedChinese, "WORKING", "进行中")
	message.SetString(language.SimplifiedChinese, "TESTING", "测试中")
	message.SetString(language.SimplifiedChinese, "DONE", "已完成")
	message.SetString(language.SimplifiedChinese, "RESOLVED", "已解决")
	message.SetString(language.SimplifiedChinese, "REOPEN", "重新打开")
	message.SetString(language.SimplifiedChinese, "WONTFIX", "无需修复")
	message.SetString(language.SimplifiedChinese, "DUP", "重复提交")
	message.SetString(language.SimplifiedChinese, "CLOSED", "已关闭")
	// issue expire state
	message.SetString(language.SimplifiedChinese, "Undefined", "未指定")
	message.SetString(language.SimplifiedChinese, "Expired", "已到期")
	message.SetString(language.SimplifiedChinese, "ExpireIn1Day", "今天到期")
	message.SetString(language.SimplifiedChinese, "ExpireIn2Days", "明天到期")
	message.SetString(language.SimplifiedChinese, "ExpireIn7Days", "7天内到期")
	message.SetString(language.SimplifiedChinese, "ExpireIn30Days", "30天内到期")
	message.SetString(language.SimplifiedChinese, "ExpireInFuture", "未来")
	// issue expire priority
	message.SetString(language.SimplifiedChinese, "URGENT", "紧急")
	message.SetString(language.SimplifiedChinese, "HIGH", "高")
	message.SetString(language.SimplifiedChinese, "NORMAL", "中")
	message.SetString(language.SimplifiedChinese, "LOW", "低")
	message.SetString(language.SimplifiedChinese, "MoveToPriority", "转移至优先级")
	// issue expire state
	message.SetString(language.SimplifiedChinese, "MoveTo", "转移至")
	message.SetString(language.SimplifiedChinese, "MoveOut", "移出迭代")
	// issue expore assignee
	message.SetString(language.SimplifiedChinese, "MoveToAssignee", "转移至处理人")
	// issue expire custom
	message.SetString(language.SimplifiedChinese, "MoveToCustom", "转移至看板")
	// msg tips
	message.SetString(language.SimplifiedChinese, "Confirm to move out iteration", "确认移出迭代")
	message.SetString(language.SimplifiedChinese, "Confirm to delete board", "确认删除看板")
	message.SetString(language.SimplifiedChinese, "Confirm to update board", "确认更新看板")
	message.SetString(language.SimplifiedChinese, "The number of boards cannot exceed 15", "创建的看板数量不能超过15")
	message.SetString(language.SimplifiedChinese, "No permission to operate", "您暂无权限进行此操作")
}

func I18nPrinter(r *http.Request) *message.Printer {
	lang := r.Header.Get("Lang")
	p := message.NewPrinter(language.English)
	if strutil.Contains(lang, "zh") {
		p = message.NewPrinter(language.SimplifiedChinese)
	}
	return p
}
