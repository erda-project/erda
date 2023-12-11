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

package project_report

import (
	"context"
	"strconv"
	"strings"
	"time"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/dao"
	"github.com/erda-project/erda/pkg/arrays"
	"github.com/erda-project/erda/pkg/crypto/uuid"
)

func (p *provider) refreshBasicIterations() error {
	_, orgs, err := p.Org.ListOrgs(context.Background(), []int64{}, &orgpb.ListOrgRequest{
		PageNo:   1,
		PageSize: 9999,
	}, true)
	if err != nil {
		p.Log.Errorf("failed to get all orgs, err: %v", err)
		return err
	}
	for i := range orgs {
		p.orgSet.Set(orgs[i].ID, orgs[i])
	}

	projects, err := p.projDB.GetAllProjects()
	if err != nil {
		p.Log.Errorf("failed to get all projects, err: %v", err)
		return err
	}
	for i := range projects {
		p.projectSet.Set(projects[i].ID, &projects[i])
	}
	iterations, _, err := p.iterationDB.PagingIterations(apistructs.IterationPagingRequest{
		PageNo:   1,
		PageSize: 99999,
	})
	if err != nil {
		p.Log.Errorf("failed to get all iterations, err: %v", err)
		return err
	}

	for i := range iterations {
		projectDto := p.projectSet.Get(int64(iterations[i].ProjectID))
		if projectDto == nil {
			p.Log.Errorf("failed to find project for iteration, iterationID: %d, projectID: %d",
				iterations[i].ID, iterations[i].ProjectID)
			continue
		}
		orgDto := p.orgSet.Get(uint64(projectDto.OrgID))
		if orgDto == nil {
			p.Log.Errorf("failed to find org for project: %s, projectID: %d, orgID: %d",
				projectDto.Name, projectDto.ID, projectDto.OrgID)
			continue
		}
		labels, _ := p.getLabelDetails(iterations[i])
		if iter := p.iterationSet.Get(iterations[i].ID); iter != nil {
			iter.Iteration = &iterations[i]
			iter.Labels = labels
		} else {
			p.iterationSet.Set(iterations[i].ID, &IterationInfo{
				Iteration:  &iterations[i],
				ProjectDto: projectDto,
				OrgDto:     orgDto,
				Labels:     labels,
			})
		}
	}

	return nil
}

func (p *provider) getLabelDetails(iteration dao.Iteration) ([]string, []apistructs.ProjectLabel) {
	lrs, _ := p.iterationDB.GetLabelRelationsByRef(apistructs.LabelTypeIteration, strconv.FormatUint(iteration.ID, 10))
	labelIDs := make([]uint64, 0, len(lrs))
	for _, v := range lrs {
		labelIDs = append(labelIDs, v.LabelID)
	}
	projectLrs, _ := p.iterationDB.GetLabelRelationsByRef(apistructs.LabelTypeProject, strconv.FormatUint(iteration.ProjectID, 10))
	for _, v := range projectLrs {
		labelIDs = append(labelIDs, v.LabelID)
	}
	var labelNames []string
	var labels []apistructs.ProjectLabel
	labels, _ = p.bdl.ListLabelByIDs(labelIDs)
	labelNames = make([]string, 0, len(labels))
	for _, v := range labels {
		labelNames = append(labelNames, v.Name)
	}
	return labelNames, labels
}

func (p *provider) metricFieldsEtcdKey(iterID uint64) string {
	return p.Cfg.IterationMetricEtcdPrefixKey + strconv.FormatUint(iterID, 10)
}

func (p *provider) checkIterationNumberFields() {
	p.iterationSet.Iterate(func(key string, value interface{}) error {
		iter := value.(*IterationInfo)
		if iter.IterationMetricFields == nil || !iter.IterationMetricFields.IsValid() {
			fields, err := p.getIterationFields(iter)
			if err != nil {
				p.Log.Errorf("failed to generate iteration fields, iterationID: %d, err: %v", iter.Iteration.ID, err)
				return nil
			}
			iter.IterationMetricFields = fields
		}
		return nil
	})
}

func (p *provider) getIterationFields(iter *IterationInfo) (*IterationMetricFields, error) {
	if iter.IterationMetricFields != nil && iter.IterationMetricFields.IsValid() {
		return iter.IterationMetricFields, nil
	}

	fields := &IterationMetricFields{}
	if err := p.Js.Get(context.Background(), p.metricFieldsEtcdKey(iter.Iteration.ID), fields); err == nil && fields.IsValid() {
		return fields, nil
	}

	if !p.Election.IsLeader() {
		return nil, nil
	}

	fields, err := p.calIterationFields(iter)
	if err != nil {
		return nil, err
	}
	p.Js.Put(context.Background(), p.metricFieldsEtcdKey(iter.Iteration.ID), fields)
	return fields, nil
}

func (p *provider) calIterationFields(iter *IterationInfo) (*IterationMetricFields, error) {
	fields := &IterationMetricFields{
		CalculatedAt: time.Now(),
		UUID:         uuid.New(),
	}
	var doneReqStateIDs []int64
	doneReqStates, err := p.issueDB.GetIssuesStatesByTypes(&apistructs.IssueStatesRequest{
		ProjectID: iter.Iteration.ProjectID,
		IssueType: []apistructs.IssueType{apistructs.IssueTypeRequirement},
		StateBelongs: []apistructs.IssueStateBelong{
			apistructs.IssueStateBelongDone,
			apistructs.IssueStateBelongClosed,
		},
	})
	if err != nil {
		return nil, err
	}
	for i := range doneReqStates {
		doneReqStateIDs = append(doneReqStateIDs, int64(doneReqStates[i].ID))
	}

	var iterManHour apistructs.IssueManHour
	iterManHour.FromString(iter.Iteration.ManHour)
	fields.IterationEstimatedDayTotal = float64(iterManHour.EstimateTime) / 480

	var doneTaskStateIDs []int64
	doneTaskStates, err := p.issueDB.GetIssuesStatesByTypes(&apistructs.IssueStatesRequest{
		ProjectID: iter.Iteration.ProjectID,
		IssueType: []apistructs.IssueType{apistructs.IssueTypeTask},
		StateBelongs: []apistructs.IssueStateBelong{
			apistructs.IssueStateBelongDone,
			apistructs.IssueStateBelongClosed,
		},
	})
	if err != nil {
		return nil, err
	}
	for i := range doneTaskStates {
		doneTaskStateIDs = append(doneTaskStateIDs, int64(doneTaskStates[i].ID))
	}

	var workingTaskStateIDs []uint64
	workingTaskStates, err := p.issueDB.GetIssuesStatesByTypes(&apistructs.IssueStatesRequest{
		ProjectID: iter.Iteration.ProjectID,
		IssueType: []apistructs.IssueType{apistructs.IssueTypeTask},
		StateBelongs: []apistructs.IssueStateBelong{
			apistructs.IssueStateBelongWorking,
		},
	})
	if err != nil {
		return nil, err
	}
	for i := range workingTaskStates {
		workingTaskStateIDs = append(workingTaskStateIDs, workingTaskStates[i].ID)
	}

	var doneBugStateIDs []int64
	doneBugStates, err := p.issueDB.GetIssuesStatesByTypes(&apistructs.IssueStatesRequest{
		ProjectID: iter.Iteration.ProjectID,
		IssueType: []apistructs.IssueType{apistructs.IssueTypeBug},
		StateBelongs: []apistructs.IssueStateBelong{
			apistructs.IssueStateBelongClosed,
		},
	})
	if err != nil {
		return nil, err
	}
	for i := range doneBugStates {
		doneBugStateIDs = append(doneBugStateIDs, int64(doneBugStates[i].ID))
	}

	var wontfixBugStateIDs []uint64
	wontfixBugStates, err := p.issueDB.GetIssuesStatesByTypes(&apistructs.IssueStatesRequest{
		ProjectID: iter.Iteration.ProjectID,
		IssueType: []apistructs.IssueType{apistructs.IssueTypeBug},
		StateBelongs: []apistructs.IssueStateBelong{
			apistructs.IssueStateBelongWontfix,
		},
	})
	if err != nil {
		return nil, err
	}
	for i := range wontfixBugStates {
		wontfixBugStateIDs = append(wontfixBugStateIDs, wontfixBugStates[i].ID)
	}

	wontfixBugTotal, wontfixBugTotalIDs, err := p.issueDB.GetIssueNumByStates(iter.Iteration.ID, apistructs.IssueTypeBug, wontfixBugStateIDs)
	if err != nil {
		return nil, err
	}
	fields.BugWontfixTotal = wontfixBugTotal
	fields.BugWontfixTotalIDs = wontfixBugTotalIDs

	iterationSummary := p.issueDB.GetIssueSummary(int64(iter.Iteration.ID), doneTaskStateIDs, doneBugStateIDs, doneReqStateIDs)

	fields.TaskTotal = uint64(iterationSummary.Task.UnDone + iterationSummary.Task.Done)
	fields.TaskTotalIDs = append(iterationSummary.TaskDoneCountIDs, iterationSummary.TaskUnDoneCountIDs...)
	fields.RequirementTotal = uint64(iterationSummary.Requirement.UnDone + iterationSummary.Requirement.Done)
	fields.RequirementTotalIDs = append(iterationSummary.ReqDoneCountIDs, iterationSummary.ReqUnDoneCountIDs...)
	fields.RequirementDoneTotal = uint64(iterationSummary.Requirement.Done)
	fields.RequirementDoneTotalIDs = iterationSummary.ReqDoneCountIDs
	fields.TaskDoneTotalIDs = iterationSummary.TaskDoneCountIDs
	fields.BugDoneTotalIDs = iterationSummary.BugDoneCountIDs
	fields.BugUndoneTotalIDs = iterationSummary.BugUnDoneCountIDs

	requirementInclusionTaskNum, requirementInclusionTaskNumIDs, err := p.issueDB.GetRequirementInclusionTaskNum(iter.Iteration.ID)
	if err != nil {
		return nil, err
	}
	fields.RequirementAssociatedTaskTotal = requirementInclusionTaskNum
	fields.RequirementAssociatedTaskTotalIDs = requirementInclusionTaskNumIDs

	workingTaskTotal, workingTaskTotalIDs, err := p.issueDB.GetIssueNumByStates(iter.Iteration.ID, apistructs.IssueTypeTask, workingTaskStateIDs)
	if err != nil {
		return nil, err
	}
	fields.TaskWorkingTotal = workingTaskTotal
	fields.TaskWorkingTotalIDs = workingTaskTotalIDs

	totalTaskEstimateTime, err := p.issueDB.GetIssuesManHour(
		apistructs.IssuesStageRequest{
			StatisticRange: "iteration",
			RangeID:        int64(iter.Iteration.ID),
			IssueType:      apistructs.IssueTypeTask,
		})
	if err != nil {
		return nil, err
	}
	fields.TaskEstimatedMinute = uint64(totalTaskEstimateTime.SumEstimateTime)
	fields.TaskEstimatedDayGtOneTotal = uint64(totalTaskEstimateTime.EstimateManDayGtOneDayNum)
	fields.TaskEstimatedDayGtTwoTotal = uint64(totalTaskEstimateTime.EstimateManDayGtTwoDayNum)
	fields.TaskEstimatedDayGtThreeTotal = uint64(totalTaskEstimateTime.EstimateManDayGtThreeDayNum)

	if fields.RequirementTotal > 0 {
		fields.RequirementCompleteSchedule = float64(iterationSummary.Requirement.Done) / float64(fields.RequirementTotal)
		fields.RequirementAssociatedPercent = float64(requirementInclusionTaskNum) / float64(fields.RequirementTotal)
	}
	taskBeInclusionReqNum, taskBeInclusionReqNumIDs, err := p.issueDB.GetTaskConnRequirementNum(iter.Iteration.ID)
	if err != nil {
		return nil, err
	}
	haveUndoneTaskAssignees, err := p.issueDB.GetHaveUndoneTaskAssigneeNum(iter.Iteration.ID, iter.Iteration.ProjectID, doneTaskStateIDs)
	if err != nil {
		return nil, err
	}
	fields.IterationAssigneeNum = uint64(len(haveUndoneTaskAssignees))
	fields.IterationAssignees = haveUndoneTaskAssignees

	projectUndoneTaskAssignees, err := p.issueDB.GetHaveUndoneTaskAssigneeNum(0, iter.Iteration.ProjectID, doneTaskStateIDs)
	if err != nil {
		return nil, err
	}
	fields.ProjectAssigneeNum = uint64(len(projectUndoneTaskAssignees))
	fields.TaskDoneTotal = uint64(iterationSummary.Task.Done)
	fields.TaskBeInclusionRequirementTotal = taskBeInclusionReqNum
	fields.TaskBeInclusionRequirementTotalIDs = taskBeInclusionReqNumIDs

	if fields.TaskTotal > 0 {
		fields.TaskUnAssociatedTotal = fields.TaskTotal - taskBeInclusionReqNum
		fields.TaskUnAssociatedTotalIDs = iterationSummary.TaskUnDoneCountIDs
		fields.TaskCompleteSchedule = float64(iterationSummary.Task.Done) / float64(fields.TaskTotal)
		fields.TaskAssociatedPercent = float64(taskBeInclusionReqNum) / float64(fields.TaskTotal)
	}

	fields.BugTotal = uint64(iterationSummary.Bug.UnDone+iterationSummary.Bug.Done) - wontfixBugTotal
	fields.BugWithWonfixTotal = uint64(iterationSummary.Bug.UnDone + iterationSummary.Bug.Done)
	fields.BugWithWonfixTotalIDs = append(iterationSummary.BugDoneCountIDs, iterationSummary.BugUnDoneCountIDs...)
	fields.BugTotalIDs = arrays.DifferenceSet(fields.BugWithWonfixTotalIDs, fields.BugWontfixTotalIDs)
	seriousBugNum, seriousBugNumIDs, err := p.issueDB.GetSeriousBugNum(iter.Iteration.ID)
	if err != nil {
		return nil, err
	}
	demandDesignBugNum, demandDesignBugNumIDs, err := p.issueDB.GetDemandDesignBugNum(iter.Iteration.ID)
	if err != nil {
		return nil, err
	}
	reopenBugNum, _, reopenBugNumIDs, err := p.issueDB.BugReopenCount(iter.Iteration.ProjectID, []uint64{iter.Iteration.ID})
	fields.SeriousBugTotal = seriousBugNum
	fields.SeriousBugTotalIDs = seriousBugNumIDs
	fields.DemandDesignBugTotal = demandDesignBugNum
	fields.DemandDesignBugTotalIDs = demandDesignBugNumIDs
	fields.ReopenBugTotal = reopenBugNum
	fields.ReopenBugTotalIDs = reopenBugNumIDs
	fields.BugDoneTotal = uint64(iterationSummary.Bug.Done)
	if fields.BugTotal > 0 {
		fields.BugCompleteSchedule = float64(iterationSummary.Bug.Done) / float64(fields.BugTotal)
		fields.SeriousBugPercent = float64(seriousBugNum) / float64(fields.BugTotal)
		fields.DemandDesignBugPercent = float64(demandDesignBugNum) / float64(fields.BugTotal)
		fields.ReopenBugPercent = float64(reopenBugNum) / float64(fields.BugTotal+reopenBugNum)
		fields.BugUndoneTotal = fields.BugTotal - fields.BugDoneTotal
	}

	doneTaskElapsedMinute, err := p.issueDB.GetIssuesManHour(
		apistructs.IssuesStageRequest{
			StatisticRange: "iteration",
			RangeID:        int64(iter.Iteration.ID),
			StateIDs:       doneTaskStateIDs,
			IssueType:      apistructs.IssueTypeTask,
		})
	if err != nil {
		return nil, err
	}
	fields.TaskElapsedMinute = uint64(doneTaskElapsedMinute.SumElapsedTime)

	return fields, nil
}

func (p *provider) iterationLabelsFunc(iter *IterationInfo) map[string]string {
	labels := map[string]string{
		labelIterationID:    strconv.FormatUint(iter.Iteration.ID, 10),
		labelProjectID:      strconv.FormatUint(iter.Iteration.ProjectID, 10),
		labelIterationTitle: iter.Iteration.Title,
		labelIterationAssignees: func(iterInfo *IterationInfo) string {
			if iter.IterationMetricFields == nil {
				return ""
			}
			return strings.Join(iter.IterationMetricFields.IterationAssignees, ",")
		}(iter),
	}
	for i := range iter.Labels {
		label := iter.Labels[i]
		splitLabel := strings.Split(label, ":")
		if len(splitLabel) != 2 {
			continue
		}
		labels[splitLabel[0]] = splitLabel[1]
	}
	projectDto := p.projectSet.Get(int64(iter.Iteration.ProjectID))
	if projectDto == nil {
		return labels
	}
	labels[labelProjectName] = projectDto.Name
	labels[labelProjectDisplayName] = projectDto.DisplayName
	labels[labelOrgID] = strconv.FormatInt(projectDto.OrgID, 10)
	orgDto := p.orgSet.Get(uint64(projectDto.OrgID))
	if orgDto == nil {
		return labels
	}
	labels[labelOrgName] = orgDto.Name
	labels[labelIterationItemUUID] = ""
	return labels
}

func (i *iterationCollector) iterationIDsLabelsFunc(iter *IterationInfo) map[string]string {
	if iter.IterationMetricFields == nil {
		return nil
	}
	labels := map[string]string{
		"uuid":         iter.IterationMetricFields.UUID,
		"metrics_type": "",
		"ids":          "",
	}
	return labels
}
