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

package issue

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/filehelper"
	"github.com/erda-project/erda/pkg/strutil"
)

// BatchConvert 批量转换
func (svc *Issue) BatchConvert(models []dao.Issue, issueTypes []apistructs.IssueType, identityInfo apistructs.IdentityInfo) ([]apistructs.Issue, error) {
	if len(models) == 0 {
		return nil, nil
	}
	// todo 老的issue button 先不删除了。等新的issue button map 权限加上后再删除老的代码
	// 批量处理按钮
	// buttons := make(map[dao.Issue][]apistructs.IssueStateButton)
	// for _, model := range models {
	// 	buttons[model] = nil
	// }
	// if err := svc.batchGenerateButton(buttons, identityInfo); err != nil {
	// 	return nil, err
	// }

	buttons, err := svc.genrateButtonMap(models[0].ProjectID, issueTypes)
	if err != nil {
		return nil, err
	}

	// 查询项目下所有 labels
	resp, err := svc.bdl.ListLabel(apistructs.ProjectLabelListRequest{
		ProjectID: models[0].ProjectID,
		Type:      apistructs.LabelTypeIssue,
		PageNo:    1,
		PageSize:  1000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list project labels, err: %v", err)
	}
	labels := resp.List
	labelMap := make(map[uint64]apistructs.ProjectLabel, len(labels))
	for _, label := range labels {
		labelMap[uint64(label.ID)] = label
	}
	// 批量获取 models 的 labels
	var issueIDs []int64
	for _, model := range models {
		issueIDs = append(issueIDs, int64(model.ID))
	}
	issueLabelIDMap, err := svc.db.BatchQueryIssueLabelIDMap(issueIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to batch query issue label relations, err: %v", err)
	}
	if issueLabelIDMap == nil {
		issueLabelIDMap = make(map[uint64][]uint64)
	}

	// 赋值回 models
	var lock sync.RWMutex
	issues := make([]apistructs.Issue, 0, len(models))
	var wait sync.WaitGroup
	for _, model := range models {
		model := model
		wait.Add(1)
		go func() {
			defer wait.Done()
			issue := svc.ConvertWithoutButton(model, false, issueLabelIDMap[uint64(model.ID)], false, labelMap)

			issue.IssueButton = buttons[model.Type][model.State]
			lock.Lock()
			issues = append(issues, *issue)
			lock.Unlock()
		}()
	}
	wait.Wait()
	// issues 重新排序
	issueMap := make(map[int64]*apistructs.Issue)
	for i := range issues {
		issueMap[issues[i].ID] = &issues[i]
	}
	results := make([]apistructs.Issue, 0, len(models))
	for _, model := range models {
		results = append(results, *issueMap[int64(model.ID)])
	}
	return results, nil
}

// Convert: dao.Issue -> apistructs.Issue
func (svc *Issue) Convert(model dao.Issue, identityInfo apistructs.IdentityInfo) (*apistructs.Issue, error) {
	issue := svc.ConvertWithoutButton(model, true, nil, true, nil)
	button, err := svc.generateButton(model, identityInfo, nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	issue.IssueButton = button

	subscribers, err := svc.db.GetIssueSubscribersByIssueID(issue.ID)
	if err != nil {
		return nil, err
	}
	for _, v := range subscribers {
		issue.Subscribers = append(issue.Subscribers, v.UserID)
	}

	return issue, nil
}

// ConvertWithoutButton 不包括 button
func (svc *Issue) ConvertWithoutButton(model dao.Issue,
	needQueryLabelRef bool, labelIDs []uint64,
	needQueryLabels bool, projectLabels map[uint64]apistructs.ProjectLabel,
) *apistructs.Issue {
	// 标签
	if needQueryLabelRef {
		lrs, _ := svc.db.GetLabelRelationsByRef(apistructs.LabelTypeIssue, model.ID)
		labelIDs = make([]uint64, 0, len(lrs))
		for _, v := range lrs {
			labelIDs = append(labelIDs, v.LabelID)
		}
	}
	var labelNames []string
	if needQueryLabels {
		labels, _ := svc.bdl.ListLabelByIDs(labelIDs)
		labelNames = make([]string, 0, len(labels))
		for _, v := range labels {
			labelNames = append(labelNames, v.Name)
		}
	} else {
		for _, labelID := range labelIDs {
			label, ok := projectLabels[labelID]
			if ok {
				labelNames = append(labelNames, label.Name)
			}
		}
	}

	var manHour apistructs.IssueManHour
	json.Unmarshal([]byte(model.ManHour), &manHour)

	var bugStage, taskType = "", ""
	if model.Type == apistructs.IssueTypeTask {
		taskType = model.Stage
	} else if model.Type == apistructs.IssueTypeBug {
		bugStage = model.Stage
	}
	return &apistructs.Issue{
		ID:             int64(model.ID),
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
		PlanStartedAt:  model.PlanStartedAt,
		PlanFinishedAt: model.PlanFinishedAt,
		ProjectID:      model.ProjectID,
		IterationID:    model.IterationID,
		AppID:          model.AppID,
		RequirementID:  model.RequirementID,
		Type:           model.Type,
		Title:          model.Title,
		Content:        filehelper.FilterAPIFileUrl(model.Content),
		State:          model.State,
		Priority:       model.Priority,
		Complexity:     model.Complexity,
		Severity:       model.Severity,
		Creator:        model.Creator,
		Assignee:       model.Assignee,
		Labels:         labelNames,
		ManHour:        manHour.Clean(),
		Source:         model.Source,
		Owner:          model.Owner,
		BugStage:       bugStage,
		TaskType:       taskType,
		FinishTime:     model.FinishTime,
	}
}

func (svc *Issue) convertIssueToExcelList(issues []apistructs.Issue, property []apistructs.IssuePropertyIndex, projectID uint64, isDownload bool, stageMap map[issueStage]string) ([][]string, error) {
	// 默认字段列名
	r := [][]string{{"ID", "标题", "内容", "状态", "创建人", "处理人", "负责人", "任务类型或缺陷引入源", "优先级", "所属迭代", "复杂度", "严重程度", "标签", "类型", "截止时间", "创建时间"}}
	// 自定义字段列名
	for _, pro := range property {
		r[0] = append(r[0], pro.DisplayName)
	}
	// 下载模版
	if isDownload {
		return r, nil
	}
	type pair struct {
		PropertyID int64
		valueID    int64
	}
	// 构建自定义字段枚举值map
	mp := make(map[pair]string)
	for _, v := range property {
		if v.PropertyType.IsOptions() == true {
			for _, val := range v.EnumeratedValues {
				mp[pair{PropertyID: v.PropertyID, valueID: val.ID}] = val.Name
			}
		}
	}
	// 状态名map
	stateMap := make(map[int64]string)
	states, err := svc.db.GetIssuesStatesByProjectID(projectID, "")
	if err != nil {
		return nil, err
	}
	for _, v := range states {
		stateMap[int64(v.ID)] = v.Name
	}
	// 迭代map
	iterationMap := make(map[int64]string)
	iterations, _, err := svc.db.PagingIterations(apistructs.IterationPagingRequest{
		PageNo:    1,
		PageSize:  10000,
		ProjectID: projectID,
	})
	if err != nil {
		return nil, err
	}
	for _, v := range iterations {
		iterationMap[int64(v.ID)] = v.Title
	}
	iterationMap[-1] = "待办事项"
	// 自定义字段map
	var issueIDs []int64
	for _, v := range issues {
		issueIDs = append(issueIDs, v.ID)
	}
	propertyMap := make(map[int64][]dao.IssuePropertyRelation)
	properties, err := svc.db.PagingPropertyRelationByIDs(issueIDs)
	for _, v := range properties {
		propertyMap[v.IssueID] = append(propertyMap[v.IssueID], v)
	}
	for index, i := range issues {
		planFinishedAt := ""
		if i.PlanFinishedAt != nil {
			planFinishedAt = i.PlanFinishedAt.Format("2006-01-02 15:04:05")
		}
		iterationName := iterationMap[i.IterationID]
		stage := issueStage{
			Type:  i.Type,
			Value: i.GetStage(),
		}
		r = append(r, append([]string{
			strconv.FormatInt(i.ID, 10),
			i.Title,
			i.Content,
			stateMap[i.State],
			i.Creator,
			i.Assignee,
			i.Owner,
			stageMap[stage],
			map[apistructs.IssuePriority]string{
				apistructs.IssuePriorityLow:    "低",
				apistructs.IssuePriorityHigh:   "高",
				apistructs.IssuePriorityNormal: "中",
				apistructs.IssuePriorityUrgent: "紧急",
			}[i.Priority],
			iterationName,
			map[apistructs.IssueComplexity]string{
				apistructs.IssueComplexityEasy:   "容易",
				apistructs.IssueComplexityNormal: "中",
				apistructs.IssueComplexityHard:   "复杂",
			}[i.Complexity],
			map[apistructs.IssueSeverity]string{
				apistructs.IssueSeverityFatal:   "致命",
				apistructs.IssueSeveritySerious: "严重",
				apistructs.IssueSeverityNormal:  "一般",
				apistructs.IssueSeveritySlight:  "轻微",
				apistructs.IssueSeverityLow:     "建议",
			}[i.Severity],
			strutil.Join(i.Labels, ",", true),
			map[apistructs.IssueType]string{
				apistructs.IssueTypeRequirement: "需求",
				apistructs.IssueTypeTask:        "任务",
				apistructs.IssueTypeBug:         "缺陷",
				apistructs.IssueTypeEpic:        "史诗",
			}[i.Type],
		},
			[]string{
				planFinishedAt,
				i.CreatedAt.Format("2006-01-02 15:04:05"),
			}...),
		)
		relations := propertyMap[i.ID]
		// 获取每个自定义字段的值
		for _, pro := range property {
			if err != nil {
				return nil, err
			}
			// 根据字段类型将数据放入表格
			if pro.PropertyType.IsOptions() == false {
				for _, rel := range relations {
					if rel.PropertyID == pro.PropertyID {
						r[index+1] = append(r[index+1], rel.ArbitraryValue)
						break
					}
				}
			} else if pro.PropertyType == apistructs.PropertyTypeSelect {
				for _, rel := range relations {
					if rel.PropertyID == pro.PropertyID {
						r[index+1] = append(r[index+1], mp[pair{PropertyID: pro.PropertyID, valueID: rel.PropertyValueID}])
						break
					}
				}
			} else if pro.PropertyType == apistructs.PropertyTypeMultiSelect || pro.PropertyType == apistructs.PropertyTypeCheckBox {
				// 多选类型的全部已选项的名字拼接成一个字符串放入表格
				var str []string
				for _, rel := range relations {
					if rel.PropertyID == pro.PropertyID {
						str = append(str, mp[pair{PropertyID: pro.PropertyID, valueID: rel.PropertyValueID}])
					}
				}
				r[index+1] = append(r[index+1], strutil.Join(str, ","))
			}
		}
	}
	return r, nil
}

func (svc *Issue) decodeFromExcelFile(req apistructs.IssueImportExcelRequest, r io.Reader, properties []apistructs.IssuePropertyIndex) ([]apistructs.Issue,
	[]apistructs.IssuePropertyRelationCreateRequest, []int, []int, []string, int, error) {
	var (
		falseExcel, excelIndex []int
		falseReason            []string
		allIssue               []apistructs.Issue
		allInstance            []apistructs.IssuePropertyRelationCreateRequest
	)
	sheets, err := excel.Decode(r)
	if err != nil {
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("failed to decode excel, err: %v", err)
	}
	if len(sheets) == 0 {
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("not found sheet")
	}
	rows := sheets[0]
	// 校验：至少有1行 title
	if len(rows) < 1 {
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("invalid title format")
	}
	falseExcel = append(falseExcel, 0)
	falseReason = append(falseReason, "错误原因")
	// 获取状态
	states, err := svc.db.GetIssuesStatesByProjectID(req.ProjectID, apistructs.IssueType(req.Type))
	if err != nil {
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("failed to get state, err: %v", err)
	}
	stateMap := make(map[string]int64) // key: state  value: id
	for _, s := range states {
		stateMap[s.Name] = int64(s.ID)
	}
	// 获取迭代信息
	iterations, err := svc.db.FindIterations(uint64(req.ProjectID))
	if err != nil {
		return nil, nil, nil, nil, nil, 0, err
	}
	iterationMap := make(map[string]int64) // key: iterationName value: iterationID
	for _, it := range iterations {
		iterationMap[it.Title] = int64(it.ID)
	}
	iterationMap["待办事项"] = -1
	// 获取自定义字段
	type propertyValue struct {
		PropertyID int64
		Value      string
	}
	propertyNameMap := make(map[string]apistructs.IssuePropertyIndex) // key: propertyName value: property
	propertyMap := make(map[propertyValue]int64)                      // key: propertyID+value  value: valueID
	for _, pro := range properties {
		propertyNameMap[pro.PropertyName] = pro
		if pro.PropertyType.IsOptions() == true {
			for _, val := range pro.EnumeratedValues {
				propertyMap[propertyValue{pro.PropertyID, val.Name}] = val.ID
			}
		}
	}
	// 第一行是列名,之后每行都是一个事件
	for i, row := range rows[1:] {
		issue := apistructs.Issue{
			Title:     row[1],
			Content:   row[2],
			ProjectID: uint64(req.ProjectID),
		}
		if stateMap[row[3]] != 0 {
			issue.State = stateMap[row[3]]
		} else {
			falseExcel = append(falseExcel, i+1)
			falseReason = append(falseReason, "无法找到该状态")
			continue
		}
		issue.Creator = row[4]
		issue.Assignee = row[5]
		issue.Owner = row[6]
		issue.TaskType = row[7]
		issue.BugStage = row[7]
		issue.Priority = issue.Priority.GetEnName(row[8])
		if val, ok := iterationMap[row[9]]; !ok {
			falseExcel = append(falseExcel, i+1)
			falseReason = append(falseReason, "无法找到该迭代")
			continue
		} else {
			issue.IterationID = val
		}
		issue.Complexity = issue.Complexity.GetEnName(row[10])
		issue.Severity = issue.Severity.GetEnName(row[11])
		issue.Labels = strutil.Split(row[12], ",", true)
		issue.Type = issue.Type.GetEnName(row[13])
		if row[14] != "" {
			finishedTime, err := time.Parse("2006-01-02 15:04:05", row[14])
			if err != nil {
				falseExcel = append(falseExcel, i+1)
				falseReason = append(falseReason, "无法解析任务结束时间")
				continue
			}
			issue.PlanFinishedAt = &finishedTime
		}

		// firstLine[15]是创建时间，跳过
		// 获取自定义字段
		relation := apistructs.IssuePropertyRelationCreateRequest{
			OrgID:     req.OrgID,
			ProjectID: int64(req.ProjectID),
		}
		for indexx, line := range row[16:] {
			index := indexx + 16
			// 获取字段名对应的字段
			instance := apistructs.IssuePropertyInstance{
				IssuePropertyIndex: propertyNameMap[rows[0][index]],
			}
			if !instance.PropertyType.IsOptions() {
				instance.ArbitraryValue = line
			} else {
				values := strutil.Split(line, ",", true)
				for _, val := range values {
					instance.Values = append(instance.Values, propertyMap[propertyValue{instance.PropertyID, val}])
				}
			}
			relation.Property = append(relation.Property, instance)
		}
		allIssue = append(allIssue, issue)
		allInstance = append(allInstance, relation)
		excelIndex = append(excelIndex, i+1)
	}

	return allIssue, allInstance, falseExcel, excelIndex, falseReason, len(rows) - 1, nil
}
