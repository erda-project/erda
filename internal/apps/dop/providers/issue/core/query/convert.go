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

package query

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/common"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/pkg/filehelper"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	propertyDateFormat = "2006-01-02T15:04:05+08:00"
)

// Convert: dao.Issue -> apistructs.Issue
func (p *provider) Convert(model dao.Issue, identityInfo *commonpb.IdentityInfo) (*pb.Issue, error) {
	issue := p.ConvertWithoutButton(model, true, nil, true, nil)
	button, err := p.GenerateButton(model, identityInfo, nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	issue.IssueButton = button

	subscribers, err := p.db.GetIssueSubscribersByIssueID(issue.Id)
	if err != nil {
		return nil, err
	}
	for _, v := range subscribers {
		issue.Subscribers = append(issue.Subscribers, v.UserID)
	}
	if identityInfo != nil && identityInfo.OrgID != "" {
		orgID, err := strconv.ParseUint(identityInfo.OrgID, 10, 64)
		if err != nil {
			return nil, err
		}
		propertyInstances, err := p.GetIssuePropertyInstance(&pb.GetIssuePropertyInstanceRequest{
			IssueID:   issue.Id,
			OrgID:     int64(orgID),
			ScopeType: apistructs.ProjectScopeType,
			ScopeID:   strconv.FormatUint(issue.ProjectID, 10),
		})
		if err != nil {
			return nil, err
		}
		issue.PropertyInstances = propertyInstances.Property
	}

	return issue, nil
}

func (p *provider) BatchConvert(models []dao.Issue, issueTypes []string) ([]*pb.Issue, error) {
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

	buttons, err := p.GenerateButtonMap(models[0].ProjectID, issueTypes)
	if err != nil {
		return nil, err
	}

	// 查询项目下所有 labels
	resp, err := p.bdl.ListLabel(apistructs.ProjectLabelListRequest{
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
	var issueIDs []uint64
	for _, model := range models {
		issueIDs = append(issueIDs, model.ID)
	}
	issueLabelIDMap, err := p.db.BatchQueryIssueLabelIDMap(issueIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to batch query issue label relations, err: %v", err)
	}
	if issueLabelIDMap == nil {
		issueLabelIDMap = make(map[uint64][]uint64)
	}

	// 赋值回 models
	var lock sync.RWMutex
	issues := make([]pb.Issue, 0, len(models))
	var wait sync.WaitGroup
	for _, model := range models {
		model := model
		wait.Add(1)
		go func() {
			defer wait.Done()
			issue := p.ConvertWithoutButton(model, false, issueLabelIDMap[model.ID], false, labelMap)

			issue.IssueButton = buttons[model.Type][model.State]
			lock.Lock()
			issues = append(issues, *issue)
			lock.Unlock()
		}()
	}
	wait.Wait()
	// issues 重新排序
	issueMap := make(map[int64]*pb.Issue)
	for i := range issues {
		issueMap[issues[i].Id] = &issues[i]
	}
	results := make([]*pb.Issue, 0, len(models))
	for _, model := range models {
		results = append(results, issueMap[int64(model.ID)])
	}
	return results, nil
}

// ConvertWithoutButton 不包括 button
func (p *provider) ConvertWithoutButton(model dao.Issue,
	needQueryLabelRef bool, labelIDs []uint64,
	needQueryLabels bool, projectLabels map[uint64]apistructs.ProjectLabel,
) *pb.Issue {
	// 标签
	if needQueryLabelRef {
		lrs, _ := p.db.GetLabelRelationsByRef(pb.ProjectLabelTypeEnum_issue.String(), strconv.FormatUint(model.ID, 10))
		labelIDs = make([]uint64, 0, len(lrs))
		for _, v := range lrs {
			labelIDs = append(labelIDs, v.LabelID)
		}
	}
	var labelNames []string
	var labels []apistructs.ProjectLabel
	if needQueryLabels {
		labels, _ = p.bdl.ListLabelByIDs(labelIDs)
		labelNames = make([]string, 0, len(labels))
		for _, v := range labels {
			labelNames = append(labelNames, v.Name)
		}
	} else {
		for _, labelID := range labelIDs {
			label, ok := projectLabels[labelID]
			if ok {
				labelNames = append(labelNames, label.Name)
				labels = append(labels, label)
			}
		}
	}

	var manHour pb.IssueManHour
	_ = json.Unmarshal([]byte(model.ManHour), &manHour)

	var bugStage, taskType = "", ""
	if model.Type == pb.IssueTypeEnum_TASK.String() {
		taskType = model.Stage
	} else if model.Type == pb.IssueTypeEnum_BUG.String() {
		bugStage = model.Stage
	}
	issue := &pb.Issue{
		Id:           int64(model.ID),
		CreatedAt:    timestamppb.New(model.CreatedAt),
		UpdatedAt:    timestamppb.New(model.UpdatedAt),
		ProjectID:    model.ProjectID,
		IterationID:  model.IterationID,
		Type:         pb.IssueTypeEnum_Type(pb.IssueTypeEnum_Type_value[model.Type]),
		Title:        model.Title,
		Content:      filehelper.FilterAPIFileUrl(model.Content),
		State:        model.State,
		Priority:     pb.IssuePriorityEnum_Priority(pb.IssuePriorityEnum_Priority_value[model.Priority]),
		Complexity:   pb.IssueComplexityEnum_Complextity(pb.IssueComplexityEnum_Complextity_value[model.Complexity]),
		Severity:     pb.IssueSeverityEnum_Severity(pb.IssueSeverityEnum_Severity_value[model.Severity]),
		Creator:      model.Creator,
		Assignee:     model.Assignee,
		Labels:       labelNames,
		LabelDetails: getPbLabelDetails(labels),
		IssueManHour: cleanManHour(manHour),
		Source:       model.Source,
		Owner:        model.Owner,
		BugStage:     bugStage,
		TaskType:     taskType,
		ReopenCount:  int32(model.ReopenCount),
	}
	if model.PlanStartedAt != nil {
		issue.PlanStartedAt = timestamppb.New(*model.PlanStartedAt)
	}
	if model.PlanFinishedAt != nil {
		issue.PlanFinishedAt = timestamppb.New(*model.PlanFinishedAt)
	}
	if model.FinishTime != nil {
		issue.FinishTime = timestamppb.New(*model.FinishTime)
	}
	if model.AppID != nil {
		issue.AppID = *model.AppID
	}
	if model.RequirementID != nil {
		issue.RequirementID = *model.RequirementID
	}
	return issue
}

func cleanManHour(imh pb.IssueManHour) *pb.IssueManHour {
	imh.ThisElapsedTime = 0
	imh.StartTime = ""
	imh.WorkContent = ""
	return &imh
}

func getPbLabelDetails(labels []apistructs.ProjectLabel) []*pb.ProjectLabel {
	l := make([]*pb.ProjectLabel, 0, len(labels))
	for _, i := range labels {
		l = append(l, GetPbLabelDetail(i))
	}
	return l
}

func GetPbLabelDetail(l apistructs.ProjectLabel) *pb.ProjectLabel {
	return &pb.ProjectLabel{
		Id:        l.ID,
		Name:      l.Name,
		Type:      pb.ProjectLabelTypeEnum_ProjectLabelType(pb.ProjectLabelTypeEnum_ProjectLabelType_value[string(l.Type)]),
		Color:     l.Color,
		ProjectID: l.ProjectID,
		Creator:   l.Creator,
		CreatedAt: timestamppb.New(l.CreatedAt),
		UpdatedAt: timestamppb.New(l.UpdatedAt),
	}
}

type IssueStage struct {
	Type  string
	Value string
}

func (p *provider) getIssueExportDataI18n(locale, i18nKey string) []string {
	l := p.bdl.GetLocale(locale)
	t := l.Get(i18nKey)
	return strutil.Split(t, ",")
}

type PropertyEnumPair struct {
	PropertyID int64
	ValueID    int64
}

func GetCustomPropertyColumnValue(pro *pb.IssuePropertyIndex, relations []dao.IssuePropertyRelation, mp map[PropertyEnumPair]string, users map[string]apistructs.Member) string {
	if pro == nil || len(relations) == 0 {
		return ""
	}
	if mp == nil {
		mp = make(map[PropertyEnumPair]string)
	}
	// according to the field type, put the data into the table
	if !common.IsOptions(pro.PropertyType.String()) {
		for _, rel := range relations {
			if rel.PropertyID != pro.PropertyID {
				continue
			}
			switch pro.PropertyType {
			case pb.PropertyTypeEnum_Date:
				date, err := time.Parse(propertyDateFormat, rel.ArbitraryValue)
				if err != nil {
					logrus.Errorf("failed to parse date: %s by format: %s", rel.ArbitraryValue, propertyDateFormat)
					return rel.ArbitraryValue
				}
				return date.Format("2006-01-02 15:04:05")
			case pb.PropertyTypeEnum_Person:
				if username, ok := users[rel.ArbitraryValue]; ok {
					return username.Nick
				}
			default:
			}
			return rel.ArbitraryValue
		}
	} else if pro.PropertyType == pb.PropertyTypeEnum_Select {
		for _, rel := range relations {
			if rel.PropertyID == pro.PropertyID {
				return mp[PropertyEnumPair{PropertyID: pro.PropertyID, ValueID: rel.PropertyValueID}]
			}
		}
	} else if pro.PropertyType == pb.PropertyTypeEnum_MultiSelect || pro.PropertyType == pb.PropertyTypeEnum_CheckBox {
		// for multiple selection type, all selected options are concatenated into a string and put into the table
		var str []string
		for _, rel := range relations {
			if rel.PropertyID == pro.PropertyID {
				str = append(str, mp[PropertyEnumPair{PropertyID: pro.PropertyID, ValueID: rel.PropertyValueID}])
			}
		}
		return strutil.Join(str, ",")
	}
	// add empty value for this column
	return ""
}
