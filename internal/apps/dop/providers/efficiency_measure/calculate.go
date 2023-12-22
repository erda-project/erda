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

package efficiency_measure

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/crypto/uuid"
)

func (p *provider) metricFieldsEtcdKey(id uint64) string {
	return p.Cfg.PerformanceMetricEtcdPrefixKey + strconv.FormatUint(id, 10)
}

func (p *provider) getPersonalMetricFields(personalInfo *PersonalPerformanceInfo) (*personalMetricField, error) {
	if personalInfo.metricFields != nil && personalInfo.metricFields.IsValid() {
		return personalInfo.metricFields, nil
	}

	fields := &personalMetricField{}
	if err := p.Js.Get(context.Background(), p.metricFieldsEtcdKey(personalInfo.userProject.ID), fields); err == nil && fields.IsValid() {
		return fields, nil
	}

	if !p.Election.IsLeader() {
		return nil, nil
	}

	fields, err := p.calPersonalFields(personalInfo)
	if err != nil {
		return nil, err
	}
	p.Js.Put(context.Background(), p.metricFieldsEtcdKey(personalInfo.userProject.ID), fields)
	return fields, nil
}

func (p *provider) calPersonalFields(personalInfo *PersonalPerformanceInfo) (*personalMetricField, error) {
	fields := &personalMetricField{
		CalculatedAt: time.Now(),
		UUID:         uuid.New(),
	}
	userID, projectID := personalInfo.userProject.UserID, personalInfo.userProject.ProjectID

	wontfixStateIDS, err := p.getStateIDS(projectID, apistructs.IssueStateBelongWontfix)
	if err != nil {
		return nil, err
	}
	demandBugTotal, demandBugTotalIDs, err := p.issueDB.GetBugCountByUserID(userID, projectID, wontfixStateIDS, p.Cfg.DemandStageList, nil, false, 0)
	if err != nil {
		return nil, err
	}
	fields.DemandDesignBugTotal = demandBugTotal
	fields.DemandDesignBugTotalIDs = demandBugTotalIDs

	architectureDesignTotal, architectureDesignTotalIDs, err := p.issueDB.GetBugCountByUserID(userID, projectID, wontfixStateIDS, p.Cfg.ArchitectureStageList, nil, false, 0)
	if err != nil {
		return nil, err
	}
	fields.ArchitectureDesignBugTotal = architectureDesignTotal
	fields.ArchitectureDesignBugTotalIDs = architectureDesignTotalIDs

	seriousBugTotal, seriousBugTotalIDs, err := p.issueDB.GetBugCountByUserID(userID, projectID, wontfixStateIDS, nil, []string{string(apistructs.IssueSeverityFatal), string(apistructs.IssueSeveritySerious)}, false, 0)
	if err != nil {
		return nil, err
	}
	fields.SeriousBugTotal = seriousBugTotal
	fields.SeriousBugTotalIDs = seriousBugTotalIDs

	reopenBugTotal, reopenBugTotalIDs, err := p.issueDB.GetBugCountByUserID(userID, projectID, wontfixStateIDS, nil, nil, true, 0)
	if err != nil {
		return nil, err
	}
	fields.ReopenBugTotal = reopenBugTotal
	fields.ReopenBugTotalIDs = reopenBugTotalIDs

	submitBugTotal, submitBugTotalIDs, err := p.issueDB.GetBugCountByUserID(0, projectID, wontfixStateIDS, nil, nil, false, userID)
	if err != nil {
		return nil, err
	}
	fields.SubmitBugTotal = submitBugTotal
	fields.SubmitBugTotalIDs = submitBugTotalIDs

	testCaseTotal, testCaseTotalIDs, err := p.getTestCaseNum(userID, projectID)
	if err != nil {
		return nil, err
	}
	fields.CreateTestCaseTotal = testCaseTotal
	fields.CreateTestCaseTotalIDs = testCaseTotalIDs

	requirementTotal, requirementTotalIDs, err := p.issueDB.GetIssueNumByStatesAndUserID(0, userID, projectID, apistructs.IssueTypeRequirement, nil, wontfixStateIDS)
	if err != nil {
		return nil, err
	}
	fields.RequirementTotal = requirementTotal
	fields.RequirementTotalIDs = requirementTotalIDs

	workingStateIDS, err := p.getStateIDS(projectID, apistructs.IssueStateBelongWorking)
	if err != nil {
		return nil, err
	}
	openStateIDS, err := p.getStateIDS(projectID, apistructs.IssueStateBelongOpen)
	if err != nil {
		return nil, err
	}
	workingRequirementTotal, workingRequirementTotalIDs, err := p.issueDB.GetIssueNumByStatesAndUserID(0, userID, projectID, apistructs.IssueTypeRequirement, workingStateIDS, nil)
	if err != nil {
		return nil, err
	}
	fields.WorkingRequirementTotal = workingRequirementTotal
	fields.WorkingRequirementTotalIDs = workingRequirementTotalIDs

	openRequirementTotal, openRequirementTotalIDs, err := p.issueDB.GetIssueNumByStatesAndUserID(0, userID, projectID, apistructs.IssueTypeRequirement, openStateIDS, nil)
	if err != nil {
		return nil, err
	}
	fields.PendingRequirementTotal = openRequirementTotal
	fields.PendingRequirementTotalIDs = openRequirementTotalIDs

	bugTotal, bugTotalIDs, err := p.issueDB.GetIssueNumByStatesAndUserID(0, userID, projectID, apistructs.IssueTypeBug, nil, wontfixStateIDS)
	if err != nil {
		return nil, err
	}
	fields.BugTotal = bugTotal
	fields.BugTotalIDs = bugTotalIDs

	ownerBugTotal, ownerBugTotalIDs, err := p.issueDB.GetIssueNumByStatesAndUserID(userID, 0, projectID, apistructs.IssueTypeBug, nil, wontfixStateIDS)
	if err != nil {
		return nil, err
	}
	fields.OwnerBugTotal = ownerBugTotal
	fields.OwnerBugTotalIDs = ownerBugTotalIDs

	workingBugTotal, workingBugTotalIDs, err := p.issueDB.GetIssueNumByStatesAndUserID(0, userID, projectID, apistructs.IssueTypeBug, workingStateIDS, nil)
	if err != nil {
		return nil, err
	}
	fields.WorkingBugTotal = workingBugTotal
	fields.WorkingBugTotalIDs = workingBugTotalIDs

	pendingBugTotal, pendingBugTotalIDs, err := p.issueDB.GetIssueNumByStatesAndUserID(0, userID, projectID, apistructs.IssueTypeBug, openStateIDS, nil)
	if err != nil {
		return nil, err
	}
	fields.PendingBugTotal = pendingBugTotal
	fields.PendingBugTotalIDs = pendingBugTotalIDs

	taskTotal, taskTotalIDs, err := p.issueDB.GetIssueNumByStatesAndUserID(0, userID, projectID, apistructs.IssueTypeTask, nil, wontfixStateIDS)
	if err != nil {
		return nil, err
	}
	fields.TaskTotal = taskTotal
	fields.TaskTotalIDs = taskTotalIDs

	workingTaskTotal, workingTaskTotalIDs, err := p.issueDB.GetIssueNumByStatesAndUserID(0, userID, projectID, apistructs.IssueTypeTask, workingStateIDS, nil)
	if err != nil {
		return nil, err
	}
	fields.WorkingTaskTotal = workingTaskTotal
	fields.WorkingTaskTotalIDs = workingTaskTotalIDs

	pendingTaskTotal, pendingTaskTotalIDs, err := p.issueDB.GetIssueNumByStatesAndUserID(0, userID, projectID, apistructs.IssueTypeTask, openStateIDS, nil)
	if err != nil {
		return nil, err
	}
	fields.PendingTaskTotal = pendingTaskTotal
	fields.PendingTaskTotalIDs = pendingTaskTotalIDs

	resolvedStateIDS, err := p.getStateIDS(projectID, apistructs.IssueStateBelongResolved)
	if err != nil {
		return nil, err
	}
	closedStateIDS, err := p.getStateIDS(projectID, apistructs.IssueStateBelongClosed)
	if err != nil {
		return nil, err
	}
	endStateIDS := make([]int64, 0)
	for _, id := range append(resolvedStateIDS, closedStateIDS...) {
		endStateIDS = append(endStateIDS, int64(id))
	}
	bugManHour, err := p.issueDB.GetIssuesManHour(apistructs.IssuesStageRequest{
		Owner:          userID,
		IssueType:      apistructs.IssueTypeBug,
		StatisticRange: "project",
		RangeID:        int64(projectID),
		StateIDs:       endStateIDS,
	})
	fields.AvgFixBugElapsedMinute = bugManHour.AvgElapsedMinute
	fields.AvgFixBugEstimateMinute = bugManHour.AvgEstimateMinute
	fields.TotalFixFixBugElapsedMinute = float64(bugManHour.TotalElapsedMinute)
	fields.TotalFixBugEstimateMinute = float64(bugManHour.TotalEstimateMinute)
	fields.ResolvedBugTotal = float64(bugManHour.Total)

	return fields, nil
}

func (p *provider) getTestCaseNum(creator uint64, projectID uint64) (uint64, []uint64, error) {
	var issueIDs []uint64
	if err := p.projDB.Table("dice_test_cases").Where("creator_id = ?", creator).Where("project_id = ?", projectID).Find(&issueIDs).Error; err != nil {
		return 0, nil, err
	}
	return uint64(len(issueIDs)), issueIDs, nil
}

func (p *provider) getStateIDS(projectID uint64, stateBelong apistructs.IssueStateBelong) ([]uint64, error) {
	var stateIDS []uint64
	switch stateBelong {
	case apistructs.IssueStateBelongWontfix:
		stateIDS = p.propertySet.GetWonfixStateIDs(projectID)
	case apistructs.IssueStateBelongOpen:
		stateIDS = p.propertySet.GetOpenStateIDs(projectID)
	case apistructs.IssueStateBelongWorking:
		stateIDS = p.propertySet.GetWorkingStateIDs(projectID)
	case apistructs.IssueStateBelongClosed:
		stateIDS = p.propertySet.GetClosedStateIDs(projectID)
	case apistructs.IssueStateBelongDone:
		stateIDS = p.propertySet.GetDoneStateIDs(projectID)
	case apistructs.IssueStateBelongResolved:
		stateIDS = p.propertySet.GeResolvedStateIDs(projectID)
	default:
		return nil, fmt.Errorf("unsupported issue state belong: %s", stateBelong)
	}
	if len(stateIDS) > 0 {
		return stateIDS, nil
	}
	states, err := p.issueDB.GetIssuesStatesByTypes(&apistructs.IssueStatesRequest{
		IssueType:    []apistructs.IssueType{apistructs.IssueTypeBug, apistructs.IssueTypeTask, apistructs.IssueTypeRequirement},
		ProjectID:    projectID,
		StateBelongs: []apistructs.IssueStateBelong{stateBelong},
	})
	if err != nil {
		return nil, err
	}
	stateIDS = make([]uint64, 0, len(states))
	for _, v := range states {
		stateIDS = append(stateIDS, v.ID)
	}
	return stateIDS, nil
}
