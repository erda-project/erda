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

package dao

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/strutil"
)

// Issue .
type Issue struct {
	dbengine.BaseModel

	PlanStartedAt  *time.Time                 // 计划开始时间
	PlanFinishedAt *time.Time                 // 计划结束时间
	ProjectID      uint64                     // 所属项目 ID
	IterationID    int64                      // 所属迭代 ID
	AppID          *uint64                    // 所属应用 ID
	RequirementID  *int64                     // 所属需求 ID
	Type           apistructs.IssueType       // issue 类型
	Title          string                     // 标题
	Content        string                     // 内容
	State          int64                      // 状态
	Priority       apistructs.IssuePriority   // 优先级
	Complexity     apistructs.IssueComplexity // 复杂度
	Severity       apistructs.IssueSeverity   // 严重程度
	Creator        string                     // issue 创建者 ID
	Assignee       string                     // 分配到 issue 的人，即当前处理人
	Source         string                     // issue创建的来源，目前只有工单使用
	ManHour        string                     // 工时信息
	External       bool                       // 用来区分是通过ui还是bundle创建的
	Deleted        bool                       // 是否已删除
	Stage          string                     // bug阶段 or 任务类型 的值
	Owner          string                     // 负责人

	FinishTime *time.Time // 实际结束时间
}

func (Issue) TableName() string {
	return "dice_issues"
}

// GetCanUpdateFields 获取所有可以被主动更新的字段
func (i *Issue) GetCanUpdateFields() map[string]interface{} {
	return map[string]interface{}{
		"plan_started_at":  i.PlanStartedAt,
		"plan_finished_at": i.PlanFinishedAt,
		"app_id":           i.AppID,
		"title":            i.Title,
		"content":          i.Content,
		"state":            i.State,
		"priority":         i.Priority,
		"complexity":       i.Complexity,
		"severity":         i.Severity,
		"assignee":         i.Assignee,
		"iteration_id":     i.IterationID,
		"man_hour":         i.ManHour,
		"source":           i.Source,
		"stage":            i.Stage,
		"owner":            i.Owner,
		"finish_time":      i.FinishTime,
	}
}

// CreateIssue 创建
func (client *DBClient) CreateIssue(issue *Issue) error {
	return client.Create(issue).Error
}

// UpdateIssue 更新
func (client *DBClient) UpdateIssue(id uint64, fields map[string]interface{}) error {
	issue := Issue{}
	issue.ID = id

	return client.Debug().Model(&issue).Updates(fields).Error
}

// UpdateIssues 批量更新 issue
func (client *DBClient) UpdateIssues(requirementID uint64, fields map[string]interface{}) error {
	return client.Debug().Model(Issue{}).
		Where("requirement_id = ?", requirementID).
		Where("type = ?", apistructs.IssueTypeTask).
		Updates(fields).Error
}

// UpdateIssueType 转换issueType
func (client *DBClient) UpdateIssueType(issue *Issue) error {
	return client.Model(Issue{}).Save(issue).Error
}

// GetBatchUpdateIssues 获取待批量更新记录，生成活动记录
func (client *DBClient) GetBatchUpdateIssues(req *apistructs.IssueBatchUpdateRequest) ([]Issue, error) {
	var issues []Issue
	sql := client.Model(Issue{}).Where("type = ?", req.Type).Where("deleted = 0")
	if len(req.CurrentIterationIDs) > 0 {
		sql = sql.Where("iteration_id in (?)", req.CurrentIterationIDs)
	}
	if !req.All && len(req.IDs) > 0 {
		sql = sql.Where("id in (?)", req.IDs)
	}
	if req.All && req.Mine {
		sql = sql.Where("assignee = ?", req.UserID)
	}
	if err := sql.Find(&issues).Error; err != nil {
		return nil, err
	}

	return issues, nil
}

// BatchUpdateIssues 批量更新 issue
func (client *DBClient) BatchUpdateIssues(req *apistructs.IssueBatchUpdateRequest) error {
	sql := client.Model(Issue{}).Where("type = ?", req.Type)
	if len(req.CurrentIterationIDs) > 0 {
		sql = sql.Where("iteration_id in (?)", req.CurrentIterationIDs)
	}
	if req.All && req.Mine {
		sql = sql.Where("assignee = ?", req.UserID)
	}
	if !req.All && len(req.IDs) > 0 {
		sql = sql.Where("id in (?)", req.IDs)
	}

	var issue Issue
	if req.State != 0 {
		issue.State = req.State
	}
	if req.Owner != "" {
		issue.Owner = req.Owner
	}
	if req.Assignee != "" {
		issue.Assignee = req.Assignee
	}
	if req.NewIterationID != 0 {
		issue.IterationID = req.NewIterationID
	}

	return sql.Updates(issue).Error
}

// DeleteIssue 删除
func (client *DBClient) DeleteIssue(id uint64) error {
	return client.Model(Issue{}).Where("id = ?", id).Update("deleted", 1).Error
}

// DeleteIssuesByRequirement .
func (client *DBClient) DeleteIssuesByRequirement(requirementID uint64) error {
	return client.Model(Issue{}).Where("requirement_id = ?", requirementID).Update("deleted", 1).Error
}

// IssueStateCount 需求下任务统计
func (client *DBClient) IssueStateCount(requirementIDs []int64) ([]apistructs.RequirementGroupResult, error) {
	var result []apistructs.RequirementGroupResult
	if err := client.Raw("select requirement_id as id, state, COUNT(*) as count from dice_issues where type = ? and requirement_id in (?) and deleted = ? group by requirement_id, state",
		apistructs.IssueTypeTask, requirementIDs, 0).Scan(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

// PagingIssues 分页查询issue
// queryIDs = true 表示不管req.IDs是否为空都要根据ID查询
func (client *DBClient) PagingIssues(req apistructs.IssuePagingRequest, queryIDs bool) ([]Issue, uint64, error) {
	var (
		total  uint64
		issues []Issue
	)
	// 有了待办事件后，迭代id一定不能等于0
	cond := Issue{}
	if req.ProjectID > 0 {
		cond.ProjectID = req.ProjectID
	}
	if req.AppID != nil && *req.AppID > 0 {
		cond.AppID = req.AppID
	}
	if req.RequirementID != nil && *req.RequirementID > 0 {
		cond.RequirementID = req.RequirementID
	}

	sql := client.Debug()
	if req.CustomPanelID != 0 {
		joinSQL := "LEFT OUTER JOIN dice_issue_panel on dice_issues.id=dice_issue_panel.issue_id"
		sql = sql.Joins(joinSQL).Where("relation = ?", req.CustomPanelID)
	}
	sql = sql.Where(cond).Where("deleted = ?", 0)
	if len(req.IDs) > 0 {
		sql = sql.Where("dice_issues.id IN (?)", req.IDs)
	} else if queryIDs {
		return nil, 0, nil
	}
	if len(req.IterationIDs) > 0 {
		sql = sql.Where("iteration_id in (?)", req.IterationIDs)
	}
	if len(req.Type) > 0 {
		sql = sql.Where("type IN (?)", req.Type)
	}
	if len(req.Creators) > 0 {
		sql = sql.Where("creator IN (?)", req.Creators)
	}
	if len(req.Assignees) > 0 {
		sql = sql.Where("assignee IN (?)", req.Assignees)
	}
	if len(req.Priority) > 0 {
		sql = sql.Where("priority IN (?)", req.Priority)
	}
	if len(req.Complexity) > 0 {
		sql = sql.Where("complexity IN (?)", req.Complexity)
	}
	if len(req.Severity) > 0 {
		sql = sql.Where("severity IN (?)", req.Severity)
	}
	if len(req.State) > 0 {
		sql = sql.Where("state IN (?)", req.State)
	}
	if len(req.Owner) > 0 {
		sql = sql.Where("owner IN (?)", req.Owner)
	}
	if len(req.BugStage) > 0 {
		sql = sql.Where("stage IN (?)", req.BugStage)
	} else if len(req.TaskType) > 0 {
		sql = sql.Where("stage IN (?)", req.TaskType)
	}
	if len(req.ExceptIDs) > 0 {
		sql = sql.Not("id", req.ExceptIDs)
	}
	if req.StartCreatedAt > 0 {
		startCreatedAt := time.Unix(req.StartCreatedAt/1000, 0)
		sql = sql.Where("dice_issues.created_at >= ?", startCreatedAt)
	}
	if req.EndCreatedAt > 0 {
		endCreatedAt := time.Unix(req.EndCreatedAt/1000, 0)
		sql = sql.Where("dice_issues.created_at <= ?", endCreatedAt)
	}
	if req.IsEmptyPlanFinishedAt {
		sql = sql.Where("plan_finished_at IS NULL")
	}
	if req.StartFinishedAt > 0 && !req.IsEmptyPlanFinishedAt {
		startFinishedAt := time.Unix(req.StartFinishedAt/1000, 0)
		sql = sql.Where("plan_finished_at >= ?", startFinishedAt)
	}
	if req.EndFinishedAt > 0 && !req.IsEmptyPlanFinishedAt {
		endFinishedAt := time.Unix(req.EndFinishedAt/1000, 0)
		sql = sql.Where("plan_finished_at <= ?", endFinishedAt)
	}

	if req.StartClosedAt > 0 {
		startClosedAt := time.Unix(req.StartClosedAt/1000, 0)
		sql = sql.Where("finish_time >= ?", startClosedAt)
	}
	if req.EndClosedAt > 0 {
		endClosedAt := time.Unix(req.EndClosedAt/1000, 0)
		sql = sql.Where("finish_time <= ?", endClosedAt)
	}

	if req.Title != "" {
		title := strings.ReplaceAll(req.Title, "%", "\\%")
		if _, err := strutil.Atoi64(title); err == nil {
			sql = sql.Where("title LIKE ? OR dice_issues.id LIKE ?", "%"+title+"%", "%"+title+"%")
		} else {
			sql = sql.Where("title LIKE ?", "%"+title+"%")
		}
	}
	if req.Source != "" {
		sql = sql.Where("source LIKE ?", "%"+req.Source+"%")
	}
	if req.External {
		sql = sql.Where("external = 1")
	} else {
		sql = sql.Where("external = 0")
	}
	if req.OrderBy != "" {
		if req.Asc {
			sql = sql.Order(fmt.Sprintf("%s", req.OrderBy))
		} else {
			sql = sql.Order(fmt.Sprintf("%s DESC", req.OrderBy))
		}
	} else {
		sql = sql.Order("dice_issues.id DESC")
	}

	offset := (req.PageNo - 1) * req.PageSize
	if err := sql.Offset(offset).Limit(req.PageSize).Find(&issues).
		// reset offset & limit before count
		Offset(0).Limit(-1).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return issues, total, nil
}

// GetIssue 查询事件
func (client *DBClient) GetIssue(id int64) (Issue, error) {
	var issue Issue
	err := client.Where("deleted = 0").Where("id = ?", id).First(&issue).Error
	return issue, err
}

// GetIssue 查询事件
func (client *DBClient) GetIssueWithOutDelete(id int64) (Issue, error) {
	var issue Issue
	err := client.First(&issue, id).Error
	return issue, err
}

func (client *DBClient) ListIssueByIDs(issueIDs []int64) ([]Issue, error) {
	var issues []Issue
	if err := client.Where("`id` IN (?)", issueIDs).Find(&issues).Error; err != nil {
		return nil, err
	}
	return issues, nil
}

// ListIssue 查询事件列表
func (client *DBClient) ListIssue(req apistructs.IssueListRequest) ([]Issue, error) {
	var issues []Issue
	cond := Issue{}
	if req.ProjectID > 0 {
		cond.ProjectID = req.ProjectID
	}
	if req.IterationID > 0 {
		cond.IterationID = req.IterationID
	}
	if req.AppID != nil && *req.AppID > 0 {
		cond.AppID = req.AppID
	}
	if req.RequirementID != nil && *req.RequirementID > 0 {
		cond.RequirementID = req.RequirementID
	}

	sql := client.Where(cond)
	if len(req.Type) > 0 {
		sql.Where("type in (?)", req.Type)
	}
	if len(req.Assignees) > 0 {
		sql = sql.Where("assignee IN (?)", req.Assignees)
	}
	if len(req.State) > 0 {
		sql = sql.Where("state IN (?)", req.State)
	}
	sql = sql.Where("deleted = ?", 0).Order("id DESC")

	if req.OnlyIDResult {
		sql.Select("id")
	}

	sql = sql.Find(&issues)
	if err := sql.Error; err != nil {
		return nil, err
	}

	return issues, nil
}

// GetIssueSummary 获取迭代相关的issue统计信息
func (client *DBClient) GetIssueSummary(iterationID int64, task, bug, requirement []int64) apistructs.ISummary {
	var reqDoneCount, reqUnDoneCount, taskDoneCount, taskUnDoneCount, bugDoneCount, bugUnDoneCount int
	// 已完成需求数
	client.Model(Issue{}).Where("type = ?", apistructs.IssueTypeRequirement).Where("deleted = ?", 0).
		Where("state in (?)", requirement).Where("iteration_id = ?", iterationID).Count(&reqDoneCount)
	// 总需求数
	client.Model(Issue{}).Where("type = ?", apistructs.IssueTypeRequirement).Where("deleted = ?", 0).
		Where("iteration_id = ?", iterationID).Count(&reqUnDoneCount)
	// 已完成任务数
	client.Model(Issue{}).Where("type = ?", apistructs.IssueTypeTask).Where("deleted = ?", 0).
		Where("state in (?)", task).Where("iteration_id = ?", iterationID).Count(&taskDoneCount)
	// 总任务数
	client.Model(Issue{}).Where("type = ?", apistructs.IssueTypeTask).Where("deleted = ?", 0).
		Where("iteration_id = ?", iterationID).Count(&taskUnDoneCount)
	// 已完成bug数
	client.Model(Issue{}).Where("type = ?", apistructs.IssueTypeBug).Where("deleted = ?", 0).
		Where("state in (?)", bug).Where("iteration_id = ?", iterationID).Count(&bugDoneCount)
	// 总bug数
	client.Model(Issue{}).Where("type = ?", apistructs.IssueTypeBug).Where("deleted = ?", 0).
		Where("iteration_id = ?", iterationID).Count(&bugUnDoneCount)

	return apistructs.ISummary{
		Requirement: apistructs.ISummaryState{
			Done:   reqDoneCount,
			UnDone: reqUnDoneCount - reqDoneCount,
		},
		Task: apistructs.ISummaryState{
			Done:   taskDoneCount,
			UnDone: taskUnDoneCount - taskDoneCount,
		},
		Bug: apistructs.ISummaryState{
			Done:   bugDoneCount,
			UnDone: bugUnDoneCount - bugDoneCount,
		},
	}
}

// GetIssueByIssueIDs 通过issueid获取issue
func (client *DBClient) GetIssueByIssueIDs(issueIDs []uint64) ([]Issue, error) {
	var issues []Issue
	if err := client.Where("deleted = 0").Where("id in ( ? )", issueIDs).Find(&issues).Error; err != nil {
		return nil, err
	}

	return issues, nil
}

// GetIssueByState 通过状态获取issue
func (client *DBClient) GetIssueByState(state int64) (*Issue, error) {
	var issues Issue
	if err := client.Where("deleted = 0").Where("state = ?", state).Find(&issues).Error; err != nil {
		return nil, err
	}
	return &issues, nil
}

// IssueStateCount2 需求关联的任务状态统计
func (client *DBClient) IssueStateCount2(issueIDs []uint64) ([]apistructs.RequirementGroupResult, error) {
	var result []apistructs.RequirementGroupResult
	if err := client.Raw("select id, state from dice_issues where id in (?) and deleted = ?",
		issueIDs, 0).Scan(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

// GetIssueManHourSum 事件下所有的任务总和
func (client *DBClient) GetIssueManHourSum(req apistructs.IssuesStageRequest) (apistructs.IssueManHourSumResponse, error) {
	var (
		issues []Issue
		ans          = make(map[string]int64)
		sum    int64 = 0
	)
	sql := client.Table("dice_issues")
	if len(req.StatisticRange) > 0 {
		if req.StatisticRange == "project" {
			sql = sql.Where("project_id in (?)", req.RangeID)
		}
		if req.StatisticRange == "iteration" {
			sql = sql.Where("iteration_id in (?)", req.RangeID)
		}
	}
	if err := sql.Where("deleted = ?", 0).Where("type = ?", apistructs.IssueTypeTask).Find(&issues).Error; err != nil {
		return apistructs.IssueManHourSumResponse{}, err
	}
	// 工时，单位与数据库一致 （人分）

	for _, each := range issues {
		ret := apistructs.IssueManHour{}
		if each.ManHour == "" {
			continue
		}
		err := json.Unmarshal([]byte(each.ManHour), &ret)
		if err != nil {
			return apistructs.IssueManHourSumResponse{}, err
		}
		ans[each.Stage] += ret.ElapsedTime
		sum += ret.ElapsedTime
	}
	return apistructs.IssueManHourSumResponse{
		DesignManHour:    ans["design"],
		DevManHour:       ans["dev"],
		TestManHour:      ans["test"],
		ImplementManHour: ans["implement"],
		DeployManHour:    ans["deploy"],
		OperatorManHour:  ans["operator"],
		SumManHour:       sum,
	}, nil
}

// GetIssueByRange 通过迭代或项目获取issue Bug
func (client *DBClient) GetIssueBugByRange(req apistructs.IssuesStageRequest) ([]Issue, float32, error) {
	var (
		issue []Issue
		total float32
	)
	sql := client.Table("dice_issues").Where("deleted = ?", 0).Where("type = ?", apistructs.IssueTypeBug)
	if len(req.StatisticRange) > 0 {
		if req.StatisticRange == "project" {
			sql = sql.Where("project_id in (?)", req.RangeID)
		}
		if req.StatisticRange == "iteration" {
			sql = sql.Where("iteration_id in (?)", req.RangeID)
		}
	}
	if err := sql.Scan(&issue).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	return issue, total, nil
}

// GetReceiversByIssueID get receivers of issue event
func (client *DBClient) GetReceiversByIssueID(issueID int64) ([]string, error) {
	var receivers []string
	issue, err := client.GetIssue(issueID)
	if err != nil {
		return nil, err
	}
	subscribers, err := client.GetIssueSubscribersSliceByIssueID(issueID)
	if err != nil {
		return nil, err
	}
	receivers = append(receivers, issue.Assignee)
	receivers = append(receivers, subscribers...)

	return strutil.DedupSlice(receivers), nil
}

// GetIssueNumByPros query by IssuePagingRequest and group by project_id to get special issue num
func (client *DBClient) GetIssueNumByPros(projectIDS []uint64, req apistructs.IssuePagingRequest) ([]apistructs.IssueNum, error) {
	var (
		res []apistructs.IssueNum
	)

	sql := client.Table("dice_issues").Select("count(id) as issue_num, project_id").Where("deleted = ?", 0)
	if len(req.IDs) > 0 {
		sql = sql.Where("dice_issues.id IN (?)", req.IDs)
	}
	if len(req.IterationIDs) > 0 {
		sql = sql.Where("iteration_id in (?)", req.IterationIDs)
	}
	if len(req.Type) > 0 {
		sql = sql.Where("type IN (?)", req.Type)
	}
	if len(req.Creators) > 0 {
		sql = sql.Where("creator IN (?)", req.Creators)
	}
	if len(req.Assignees) > 0 {
		sql = sql.Where("assignee IN (?)", req.Assignees)
	}
	if len(req.Priority) > 0 {
		sql = sql.Where("priority IN (?)", req.Priority)
	}
	if len(req.Complexity) > 0 {
		sql = sql.Where("complexity IN (?)", req.Complexity)
	}
	if len(req.Severity) > 0 {
		sql = sql.Where("severity IN (?)", req.Severity)
	}
	if len(req.State) > 0 {
		sql = sql.Where("state IN (?)", req.State)
	}
	if len(req.Owner) > 0 {
		sql = sql.Where("owner IN (?)", req.Owner)
	}
	if len(req.BugStage) > 0 {
		sql = sql.Where("stage IN (?)", req.BugStage)
	} else if len(req.TaskType) > 0 {
		sql = sql.Where("stage IN (?)", req.TaskType)
	}
	if len(req.ExceptIDs) > 0 {
		sql = sql.Not("id", req.ExceptIDs)
	}
	if req.StartCreatedAt > 0 {
		startCreatedAt := time.Unix(req.StartCreatedAt/1000, 0)
		sql = sql.Where("dice_issues.created_at >= ?", startCreatedAt)
	}
	if req.EndCreatedAt > 0 {
		endCreatedAt := time.Unix(req.EndCreatedAt/1000, 0)
		sql = sql.Where("dice_issues.created_at <= ?", endCreatedAt)
	}
	if req.IsEmptyPlanFinishedAt {
		sql = sql.Where("plan_finished_at IS NULL")
	}
	if req.StartFinishedAt > 0 && !req.IsEmptyPlanFinishedAt {
		startFinishedAt := time.Unix(req.StartFinishedAt/1000, 0)
		sql = sql.Where("plan_finished_at >= ?", startFinishedAt)
	}
	if req.EndFinishedAt > 0 && !req.IsEmptyPlanFinishedAt {
		endFinishedAt := time.Unix(req.EndFinishedAt/1000, 0)
		sql = sql.Where("plan_finished_at <= ?", endFinishedAt)
	}

	if req.StartClosedAt > 0 {
		startClosedAt := time.Unix(req.StartClosedAt/1000, 0)
		sql = sql.Where("finish_time >= ?", startClosedAt)
	}
	if req.EndClosedAt > 0 {
		endClosedAt := time.Unix(req.EndClosedAt/1000, 0)
		sql = sql.Where("finish_time <= ?", endClosedAt)
	}
	if len(projectIDS) > 0 {
		sql = sql.Where("project_id IN (?)", projectIDS)
	}

	if req.Title != "" {
		title := strings.ReplaceAll(req.Title, "%", "\\%")
		if _, err := strutil.Atoi64(title); err == nil {
			sql = sql.Where("title LIKE ? OR dice_issues.id LIKE ?", "%"+title+"%", "%"+title+"%")
		} else {
			sql = sql.Where("title LIKE ?", "%"+title+"%")
		}
	}
	if req.Source != "" {
		sql = sql.Where("source LIKE ?", "%"+req.Source+"%")
	}
	if req.External {
		sql = sql.Where("external = 1")
	} else {
		sql = sql.Where("external = 0")
	}

	if err := sql.Group("project_id").Find(&res).Error; err != nil {
		return nil, err
	}

	return res, nil
}
