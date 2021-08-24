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

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// IssueTemplate issue 事件模板, key 为 language
var IssueTemplate = map[string]map[IssueStreamType]string{
	"zh": {
		ISTCreate:               `该事件由 {{.UserName}} 创建`,
		ISTComment:              `{{.Comment}}`,
		ISTRelateMR:             `mrInfo: {{.MRInfo}}`,
		ISTAssign:               `该事件分派给 "{{.UserName}}" 处理`,
		ISTTransferState:        `该事件状态自 "{{.CurrentState}}" 迁移至 "{{.NewState}}"`,
		ISTChangeTitle:          `该事件标题自 "{{.CurrentTitle}}" 更新为 "{{.NewTitle}}"`,
		ISTChangePlanStartedAt:  `该事件计划开始时间自 "{{.CurrentPlanStartedAt}}" 调整为 "{{.NewPlanStartedAt}}"`,
		ISTChangePlanFinishedAt: `该事件计划结束时间自 "{{.CurrentPlanFinishedAt}}" 调整为 "{{.NewPlanFinishedAt}}"`,
		ISTChangeAssignee:       `该事件处理人由 "{{.CurrentAssignee}}" 变更为 "{{.NewAssignee}}"`,
		ISTChangeIteration:      `该事件迭代由 "{{.CurrentIteration}}" 变更为 "{{.NewIteration}}"`,
		ISTChangeManHour:        `该事件工时信息由【预估时间：{{.CurrentEstimateTime}}，已用时间：{{.CurrentElapsedTime}}，剩余时间：{{.CurrentRemainingTime}}，开始时间：{{.CurrentStartTime}}，工作内容：{{.CurrentWorkContent}}】变更为【预估时间：{{.NewEstimateTime}}，已用时间：{{.NewElapsedTime}}，剩余时间：{{.NewRemainingTime}}，开始时间：{{.NewStartTime}}，工作内容：{{.NewWorkContent}}】`,
		ISTChangeOwner:          `该事件责任人由 "{{.CurrentOwner}}" 变更为 "{{.NewOwner}}"`,
		ISTChangeTaskType:       `该事件任务类型由 "{{.CurrentStage}}" 变更为 "{{.NewStage}}"`,
		ISTChangeBugStage:       `该事件引入源由 "{{.CurrentStage}}" 变更为 "{{.NewStage}}"`,
		ISTChangePriority:       `该事件优先级由 "{{.CurrentPriority}}" 变更为 "{{.NewPriority}}"`,
		ISTChangeComplexity:     `该事件复杂度由 "{{.CurrentComplexity}}" 变更为 "{{.NewComplexity}}"`,
		ISTChangeSeverity:       `该事件严重程度由 "{{.CurrentSeverity}}" 变更为 "{{.NewSeverity}}"`,
		ISTChangeContent:        `该事件内容发生变更`,
		ISTChangeLabel:          `该事件标签发生变更`,
	},
	`en`: {
		ISTCreate:               `{{.UserName}} created`,
		ISTComment:              `{{.Comment}}`,
		ISTRelateMR:             `mrInfo: {{.MRInfo}}`,
		ISTAssign:               `assigned to "{{.UserName}}"`,
		ISTTransferState:        `transfer state from "{{.CurrentState}}" to "{{.NewState}}"`,
		ISTChangeTitle:          `change title "{{.CurrentTitle}}" to "{{.NewTitle}}"`,
		ISTChangePlanStartedAt:  `adjust Planned Start Time from "{{.CurrentPlanStartedAt}}" to "{{.NewPlanStartedAt}}"`,
		ISTChangePlanFinishedAt: `adjust Planned Finished Time from "{{.CurrentPlanFinishedAt}}" to "{{.NewPlanFinishedAt}}"`,
		ISTChangeAssignee:       `adjust Assignee from "{{.CurrentAssignee}}" to "{{.NewAssignee}}"`,
		ISTChangeIteration:      `adjust Iteration from "{{.CurrentIteration}}" to "{{.NewIteration}}"`,
		ISTChangeManHour:        `adjust man-hour from【EstimateTime: {{.CurrentEstimateTime}}, ElapsedTime: {{.CurrentElapsedTime}}, RemainingTime: {{.CurrentRemainingTime}}, StartTime: {{.CurrentStartTime}}, WorkContent: {{.CurrentWorkContent}}】to【EstimateTime: {{.NewEstimateTime}}, ElapsedTime: {{.NewElapsedTime}}, RemainingTime: {{.NewRemainingTime}}, StartTime: {{.NewStartTime}}, WorkContent: {{.NewWorkContent}}】`,
	},
}

// ISTParam issue stream template params, 字段名称须与模板内占位符匹配
type ISTParam struct {
	Comment     string `json:",omitempty"` // 评论内容
	CommentTime string `json:",omitempty"` // comment time
	UserName    string `json:",omitempty"` // 用户名

	MRInfo MRCommentInfo `json:",omitempty"` // MR 类型评论内容

	CurrentState string `json:",omitempty"` // 当前状态
	NewState     string `json:",omitempty"` // 新状态

	CurrentTitle string `json:",omitempty"` // 当前标题
	NewTitle     string `json:",omitempty"` // 新标题

	CurrentPlanStartedAt string `json:",omitempty"` // 当前计划开始时间
	NewPlanStartedAt     string `json:",omitempty"` // 新计划开始时间

	CurrentPlanFinishedAt string `json:",omitempty"` // 当前计划结束时间
	NewPlanFinishedAt     string `json:",omitempty"` // 新计划结束时间

	CurrentAssignee string `json:",omitempty"` // 当前处理人
	NewAssignee     string `json:",omitempty"` // 新处理人

	CurrentIteration string `json:",omitempty"` // 当前迭代
	NewIteration     string `json:",omitempty"` // 新迭代

	CurrentEstimateTime  string `json:",omitempty"` //当前预估时间
	CurrentElapsedTime   string `json:",omitempty"` //当前已用时间
	CurrentRemainingTime string `json:",omitempty"` //当前剩余时间
	CurrentStartTime     string `json:",omitempty"` //当前开始时间
	CurrentWorkContent   string `json:",omitempty"` //当前工作内容
	NewEstimateTime      string `json:",omitempty"` //新预估时间
	NewElapsedTime       string `json:",omitempty"` //新已用时间
	NewRemainingTime     string `json:",omitempty"` //新剩余时间
	NewStartTime         string `json:",omitempty"` //新开始时间
	NewWorkContent       string `json:",omitempty"` //新工作内容

	CurrentOwner string `json:",omitempty"` // 当前责任人
	NewOwner     string `json:",omitempty"` // 新责任人

	CurrentStage string `json:",omitempty"` // 当前任务类型/引用源
	NewStage     string `json:",omitempty"` // 新任务类型/引用源

	CurrentPriority string `json:",omitempty"` // 当前优先级
	NewPriority     string `json:",omitempty"` // 新优先级

	CurrentComplexity string `json:",omitempty"` // 当前复杂度
	NewComplexity     string `json:",omitempty"` // 新复杂度

	CurrentSeverity string `json:",omitempty"` // 当前严重性
	NewSeverity     string `json:",omitempty"` // 新严重性

	CurrentContent string `json:",omitempty"` // 当前内容
	NewContent     string `json:",omitempty"` // 新内容

	CurrentLabel string `json:",omitempty"` // 当前标签
	NewLabel     string `json:",omitempty"` // 新标签
}

func (p ISTParam) Value() (driver.Value, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, errors.Errorf("failed to marshal ISTParam, err: %v", err)
	}
	return string(b), nil
}

func (p *ISTParam) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return errors.Errorf("invalid scan source for ISTParam")
	}
	if len(v) == 0 {
		return nil
	}
	if err := json.Unmarshal(v, p); err != nil {
		return err
	}
	return nil
}

// GetIssueStreamTemplate 获取事件流模板
func GetIssueStreamTemplate(locale string, ist IssueStreamType) (string, error) {
	if locale != "zh" && locale != "en" {
		return "", errors.Errorf("invalid locale %v", locale)
	}

	v, ok := IssueTemplate[locale][ist]
	if !ok {
		return "", errors.Errorf("issue stream template not found")
	}
	return v, nil
}

func (p *ISTParam) Localize(locale string) *ISTParam {
	//// CurrentState
	//
	//p.CurrentState = IssueState(p.CurrentState).Desc(locale)
	//
	//// NewStatue
	//p.NewState = IssueState(p.NewState).Desc(locale)

	return p
}
