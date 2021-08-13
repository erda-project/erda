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
	// issue
	message.SetString(language.SimplifiedChinese, "Sprint", "迭代")
	message.SetString(language.SimplifiedChinese, "All", "全部")
	message.SetString(language.SimplifiedChinese, "Choose Sprint", "选择迭代")
	message.SetString(language.SimplifiedChinese, "Title", "标题")
	message.SetString(language.SimplifiedChinese, "Please enter title or ID", "请输入标题或ID")
	message.SetString(language.SimplifiedChinese, "Label", "标签")
	message.SetString(language.SimplifiedChinese, "Please choose label", "请选择标签")
	message.SetString(language.SimplifiedChinese, "Priority", "优先级")
	message.SetString(language.SimplifiedChinese, "Choose Priorities", "选择优先级")
	message.SetString(language.SimplifiedChinese, "Severity", "严重程度")
	message.SetString(language.SimplifiedChinese, "Choose Severity", "选择严重程度")
	message.SetString(language.SimplifiedChinese, "FATAL", "致命")
	message.SetString(language.SimplifiedChinese, "SERIOUS", "严重")
	message.SetString(language.SimplifiedChinese, "NORMAL", "一般")
	message.SetString(language.SimplifiedChinese, "SLIGHT", "轻微")
	message.SetString(language.SimplifiedChinese, "SUGGEST", "建议")
	message.SetString(language.SimplifiedChinese, "Creator", "创建人")
	message.SetString(language.SimplifiedChinese, "Choose Yourself", "选择自己")
	message.SetString(language.SimplifiedChinese, "Assignee", "处理人")
	message.SetString(language.SimplifiedChinese, "Responsible Person", "责任人")
	message.SetString(language.SimplifiedChinese, "EPIC", "里程碑")
	message.SetString(language.SimplifiedChinese, "Task Type", "任务类型")
	message.SetString(language.SimplifiedChinese, "Import Source", "引入源")
	message.SetString(language.SimplifiedChinese, "Demand Design", "需求设计")
	message.SetString(language.SimplifiedChinese, "Architecture Design", "架构设计")
	message.SetString(language.SimplifiedChinese, "Code Development", "代码研发")
	message.SetString(language.SimplifiedChinese, "Created At", "创建日期")
	message.SetString(language.SimplifiedChinese, "Deadline", "截止日期")
	message.SetString(language.SimplifiedChinese, "Closed At", "关闭日期")
	message.SetString(language.SimplifiedChinese, "REQUIREMENT", "需求")
	message.SetString(language.SimplifiedChinese, "BUG", "缺陷")
	message.SetString(language.SimplifiedChinese, "TASK", "任务")
	message.SetString(language.SimplifiedChinese, "Create Issue", "新建事项")
	message.SetString(language.SimplifiedChinese, "Create Requirement", "新建需求")
	message.SetString(language.SimplifiedChinese, "Create Bug", "新建缺陷")
	message.SetString(language.SimplifiedChinese, "Create Task", "新建任务")
	message.SetString(language.SimplifiedChinese, "The Gantt chart of the event can only be displayed properly if the deadline and estimated time are correctly entered",
		"事项的甘特图只有确保正确输入截止日期、预计时间才能正常显示")
	message.SetString(language.SimplifiedChinese, "#gray#gray#>gray#: Represents the remaining time period of the deadline",
		"#gray#灰色#>gray#：代表事项截止日期的剩余时间段")
	message.SetString(language.SimplifiedChinese, "#blue#blue#>blue#: Represents the time period from the start time of the issue to the current or the issue completion date",
		"#blue#蓝色#>blue#：代表从事项开始时间到当前/事项完成日期的时间段")
	message.SetString(language.SimplifiedChinese, "#red#read#>red#: Represents the timeout period from the due date to the current or completion date of the issue",
		"#red#红色#>red#：代表截止日期到当前/事项完成日期的超时时间段")
	message.SetString(language.SimplifiedChinese, "List", "列表")
	message.SetString(language.SimplifiedChinese, "Board", "看板")
	message.SetString(language.SimplifiedChinese, "Gantt Chart", "甘特图")
	message.SetString(language.SimplifiedChinese, "Board View", "看板视图")
	message.SetString(language.SimplifiedChinese, "Custom", "自定义")
	message.SetString(language.SimplifiedChinese, "State", "状态")
}

func I18nPrinter(r *http.Request) *message.Printer {
	lang := r.Header.Get("Lang")
	p := message.NewPrinter(language.English)
	if strutil.Contains(lang, "zh") {
		p = message.NewPrinter(language.SimplifiedChinese)
	}
	return p
}
