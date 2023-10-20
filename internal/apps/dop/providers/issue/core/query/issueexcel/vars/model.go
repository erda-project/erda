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

package vars

import (
	"time"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
)

// excel format:
//
// issues sheet:
// 通用字段 | 需求专属字段               | 任务专属字段             | 缺陷专属字段
// -       | 待办事项 | 自定义字段       | 任务类型  | 自定义字段   | 负责人 | 引入源 | 重新打开次数 | 自定义字段
// -                 | 字段1,字段2(动态) |          | 字段1,字段2  |       |        |             | 字段1,字段2
//
// 自定义字段 sheet
// 事项类型 | JSON
//
// 标签 sheet
// name | color
//
// 用户信息 sheet
// ID | Nick | UserInfo(JSON)
type (
	IssueSheetModel struct {
		Common          IssueSheetModelCommon          `excel:"Common"`
		RequirementOnly IssueSheetModelRequirementOnly `excel:"RequirementOnly"`
		TaskOnly        IssueSheetModelTaskOnly        `excel:"TaskOnly"`
		BugOnly         IssueSheetModelBugOnly         `excel:"BugOnly"`
	}

	IssueSheetModelCommon struct {
		ID                 uint64 `excel:"ID"`
		IterationName      string `excel:"IterationName"`
		IssueType          pb.IssueTypeEnum_Type
		IssueTitle         string
		Content            string
		State              string
		Priority           pb.IssuePriorityEnum_Priority
		Complexity         pb.IssueComplexityEnum_Complextity
		Severity           pb.IssueSeverityEnum_Severity
		CreatorName        string
		AssigneeName       string
		CreatedAt          *time.Time
		PlanStartedAt      *time.Time
		PlanFinishedAt     *time.Time
		StartAt            *time.Time
		FinishAt           *time.Time
		EstimateTime       string
		Labels             []string
		ConnectionIssueIDs []int64 // L264 转为 -264

		LineNum int // 行号，用于错误和告警跟踪
	}
	IssueSheetModelRequirementOnly struct {
		InclusionIssueIDs []int64
		CustomFields      []ExcelCustomField
	}
	IssueSheetModelTaskOnly struct {
		TaskType     string
		CustomFields []ExcelCustomField
	}
	IssueSheetModelBugOnly struct {
		OwnerName    string
		Source       string
		ReopenCount  int32
		CustomFields []ExcelCustomField
	}

	ExcelCustomField struct {
		Title string
		Value string
	}
)
