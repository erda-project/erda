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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/bdl"
	"github.com/erda-project/erda/internal/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

type ProjectReportRow struct {
	RequirementTotal             float64   `json:"requirementTotal" ch:"requirementTotal"`
	BugTotal                     float64   `json:"bugTotal" ch:"bugTotal"`
	TaskTotal                    float64   `json:"taskTotal" ch:"taskTotal"`
	ResponsibleFuncPointsTotal   float64   `json:"responsibleFuncPointsTotal" ch:"responsibleFuncPointsTotal"`
	RequirementFuncPointsTotal   float64   `json:"requirementFuncPointsTotal" ch:"requirementFuncPointsTotal"`
	DevFuncPointsTotal           float64   `json:"devFuncPointsTotal" ch:"devFuncPointsTotal"`
	DemandFuncPointsTotal        float64   `json:"demandFuncPointsTotal" ch:"demandFuncPointsTotal"`
	TestFuncPointsTotal          float64   `json:"testFuncPointsTotal" ch:"testFuncPointsTotal"`
	BudgetMandayTotal            float64   `json:"budgetMandayTotal" ch:"budgetMandayTotal"`
	TaskEstimatedMinute          float64   `json:"taskEstimatedMinute" ch:"taskEstimatedMinute"`
	TaskEstimatedManday          float64   `json:"taskEstimatedManday" ch:"taskEstimatedManday"`
	ActualMandayTotal            float64   `json:"actualMandayTotal" ch:"actualMandayTotal"`
	RequirementDoneRate          float64   `json:"requirementDoneRate" ch:"requirementDoneRate"`
	TaskDoneTotal                float64   `json:"taskDoneTotal" ch:"taskDoneTotal"`
	TaskDoneRate                 float64   `json:"taskDoneRate" ch:"taskDoneRate"`
	TaskEstimatedDayGtOneTotal   float64   `json:"taskEstimatedDayGtOneTotal" ch:"taskEstimatedDayGtOneTotal"`
	TaskEstimatedDayGtTwoTotal   float64   `json:"taskEstimatedDayGtTwoTotal" ch:"taskEstimatedDayGtTwoTotal"`
	TaskEstimatedDayGtThreeTotal float64   `json:"taskEstimatedDayGtThreeTotal" ch:"taskEstimatedDayGtThreeTotal"`
	UnfinishedAssigneeTotal      float64   `json:"unfinishedAssigneeTotal" ch:"unfinishedAssigneeTotal"`
	RequirementDoneTotal         float64   `json:"requirementDoneTotal" ch:"requirementDoneTotal"`
	RequirementAssociatedTotal   float64   `json:"requirementAssociatedTotal" ch:"requirementAssociatedTotal"`
	RequirementAssociatedRate    float64   `json:"requirementAssociatedRate" ch:"requirementAssociatedRate"`
	RequirementUnassignedTotal   float64   `json:"requirementUnassignedTotal" ch:"requirementUnassignedTotal"`
	RequirementUnassignedRate    float64   `json:"requirementUnassignedRate" ch:"requirementUnassignedRate"`
	TaskUnassignedTotal          float64   `json:"taskUnassignedTotal" ch:"taskUnassignedTotal"`
	BugUndoneTotal               float64   `json:"bugUndoneTotal" ch:"bugUndoneTotal"`
	BugDoneRate                  float64   `json:"bugDoneRate" ch:"bugDoneRate"`
	BugSeriousTotal              float64   `json:"bugSeriousTotal" ch:"bugSeriousTotal"`
	BugSeriousRate               float64   `json:"bugSeriousRate" ch:"bugSeriousRate"`
	BugDemandDesignTotal         float64   `json:"bugDemandDesignTotal" ch:"bugDemandDesignTotal"`
	BugDemandDesignRate          float64   `json:"bugDemandDesignRate" ch:"bugDemandDesignRate"`
	BugOnlineTotal               float64   `json:"bugOnlineTotal" ch:"bugOnlineTotal"`
	BugOnlineRate                float64   `json:"bugOnlineRate" ch:"bugOnlineRate"`
	BugReopenTotal               float64   `json:"bugReopenTotal" ch:"bugReopenTotal"`
	BugReopenRate                float64   `json:"bugReopenRate" ch:"bugReopenRate"`
	TaskAssociatedTotal          float64   `json:"taskAssociatedTotal" ch:"taskAssociatedTotal"`
	TaskAssociatedRate           float64   `json:"taskAssociatedRate" ch:"taskAssociatedRate"`
	BugLowLevelTotal             float64   `json:"bugLowLevelTotal" ch:"bugLowLevelTotal"`
	BugLowLevelRate              float64   `json:"bugLowLevelRate" ch:"bugLowLevelRate"`
	IterationCompletedRate       float64   `json:"iterationCompletedRate" ch:"iterationCompletedRate"`
	TaskWorkingTotal             float64   `json:"taskWorkingTotal" ch:"taskWorkingTotal"`
	BugWontfixTotal              float64   `json:"bugWontfixTotal" ch:"bugWontfixTotal"`
	IterationAssigneeTotal       float64   `json:"iterationAssigneeTotal" ch:"iterationAssigneeTotal"`
	IterationEstimatedDayTotal   float64   `json:"iterationEstimatedDayTotal" ch:"iterationEstimatedDayTotal"`
	ProjectName                  string    `json:"projectName" ch:"projectName"`
	ProjectDisplayName           string    `json:"projectDisplayName" ch:"projectDisplayName"`
	ProjectID                    string    `json:"projectID" ch:"projectID"`
	Timestamp                    time.Time `json:"timestamp" ch:"timestamp"`
	EmpProjectCode               string    `json:"empProjectCode" ch:"empProjectCode"`
	BeginDate                    string    `json:"beginDate" ch:"beginDate"`
	EndDate                      string    `json:"endDate" ch:"endDate"`
	ActualEndDate                string    `json:"actualEndDate" ch:"actualEndDate"`
	ProjectAssignees             []string  `json:"projectAssignees" ch:"projectAssignees"`
}

var (
	metricGroup = "project_report"
)

var (
	basicSql = `
SELECT
    requirementTotal,
    bugTotal,
    taskTotal,
    responsibleFuncPointsTotal,
    requirementFuncPointsTotal,
    devFuncPointsTotal,
    demandFuncPointsTotal,
    testFuncPointsTotal,
    budgetMandayTotal,
    taskEstimatedMinute,
    taskEstimatedManday,
    actualMandayTotal,
    taskDoneTotal,
    taskDoneRate,
    taskEstimatedDayGtOneTotal,
    taskEstimatedDayGtTwoTotal,
    taskEstimatedDayGtThreeTotal,
    requirementDoneRate,
    last_value(projectAssignees) as projectAssignees,
    toFloat64(length(projectAssignees)) as unfinishedAssigneeTotal,
    requirementDoneTotal,
    requirementAssociatedTotal,
    requirementAssociatedRate,
    requirementUnassignedTotal,
    requirementUnassignedRate,
    taskUnassignedTotal,
    bugUndoneTotal,
    bugDoneRate,
    bugSeriousTotal,
    bugSeriousRate,
    bugDemandDesignTotal,
    bugDemandDesignRate,
    bugOnlineTotal,
    bugOnlineRate,
    bugReopenTotal,
    bugReopenRate,
    taskAssociatedTotal,
    taskAssociatedRate,
    bugLowLevelTotal,
    bugLowLevelRate,
    iterationCompletedRate,
    taskWorkingTotal,
    bugWontfixTotal,
    iterationAssigneeTotal,
    iterationEstimatedDayTotal,
    projectName,
    projectDisplayName,
    projectID,
    timestamp,
    empProjectCode,
    last_value(beginDate) as beginDate,
    last_value(endDate) as endDate,
    last_value(actualEndDate) as actualEndDate
FROM
    (
    SELECT
        sum(requirement_total) as requirementTotal,
        sum(bug_total) as bugTotal,
        sum(task_total) as taskTotal,
    	sum(responsible_func_points_total) as responsibleFuncPointsTotal,
    	sum(requirement_func_points_total) as requirementFuncPointsTotal,
    	sum(dev_func_points_total) as devFuncPointsTotal,
    	sum(demand_func_points_total) as demandFuncPointsTotal,
    	sum(test_func_points_total) as testFuncPointsTotal,
        sum(budget_manday_total) as budgetMandayTotal,
        sum(task_estimated_minute) as taskEstimatedMinute,
    	sum(task_estimated_minute) / 480 as taskEstimatedManday,
        sum(actual_manday_total) as actualMandayTotal,
        sum(task_done_total) as taskDoneTotal,
        if(sum(task_total) > 0, sum(task_done_total) / sum(task_total), 0) as taskDoneRate,
        sum(task_estimated_day_gt_one_total) as taskEstimatedDayGtOneTotal,
        sum(task_estimated_day_gt_two_total) as taskEstimatedDayGtTwoTotal,
        sum(task_estimated_day_gt_three_total) as taskEstimatedDayGtThreeTotal,
        last_value(unfinished_assignee_total) as unfinishedAssigneeTotal,
        sum(requirement_done_total) as requirementDoneTotal,
    	if(sum(requirement_total) > 0, sum(requirement_done_total) / sum(requirement_total), 0) as requirementDoneRate,
        sum(requirement_associated_total) as requirementAssociatedTotal,
        if(sum(requirement_total) > 0, sum(requirement_associated_total) / sum(requirement_total), 0) as requirementAssociatedRate,
        sum(requirement_unassigned_total) as requirementUnassignedTotal,
        if(sum(requirement_total) > 0, sum(requirement_unassigned_total) / sum(requirement_total), 0) as requirementUnassignedRate,
        sum(task_unassigned_total) as taskUnassignedTotal,
        sum(bug_undone_total) as bugUndoneTotal,
        if(sum(bug_total) > 0, (sum(bug_total)-sum(bug_undone_total)) / sum(bug_total), 0) as bugDoneRate,
        sum(bug_serious_total) as bugSeriousTotal,
        if(sum(bug_total) > 0, sum(bug_serious_total) / sum(bug_total), 0) as bugSeriousRate,
        sum(bug_demand_design_total) as bugDemandDesignTotal,
        if(sum(bug_total) > 0, sum(bug_demand_design_total) / sum(bug_total), 0) as bugDemandDesignRate,
        sum(bug_online_total) as bugOnlineTotal,
        if(sum(bug_total) > 0, sum(bug_online_total) / sum(bug_total), 0) as bugOnlineRate,
        sum(bug_reopen_total) as bugReopenTotal,
        if((sum(bug_total)+sum(bug_reopen_total)) > 0, sum(bug_reopen_total) / (sum(bug_total)+sum(bug_reopen_total)), 0) as bugReopenRate,
        sum(task_associated_total) as taskAssociatedTotal,
        if(sum(task_total) > 0, sum(task_associated_total) / sum(task_total), 0) as taskAssociatedRate,
    	sum(bug_low_level_total) as bugLowLevelTotal,
        if(sum(bug_total) > 0, sum(bug_low_level_total) / sum(bug_total), 0) as bugLowLevelRate,
    	sum(task_working_total) as taskWorkingTotal,
        if(sum(task_total) > 0, (4*sum(task_done_total)+sum(task_working_total))/(4*sum(task_total)), 0) as iterationCompletedRate,
    	sum(bug_wontfix_total) as bugWontfixTotal,
    	sum(iteration_unfinished_assignee_total) as iterationAssigneeTotal,
        sum(iteration_estimated_day_total) as iterationEstimatedDayTotal,
        if(last_value(emp_project_begin_date) > 0, toString(fromUnixTimestamp(toInt64(last_value(emp_project_begin_date)))), '') as beginDate,
        if(last_value(emp_project_end_date) > 0, toString(fromUnixTimestamp(toInt64(last_value(emp_project_end_date)))), '') as endDate,
        if(last_value(emp_project_actual_end_date) > 0, toString(fromUnixTimestamp(toInt64(last_value(emp_project_actual_end_date)))), '') as actualEndDate,
        groupUniqArrayArray(iteration_assignees) as projectAssignees,
        projectName,
        projectDisplayName,
        projectID,
        timestamp,
        empProjectCode
    FROM 
    (
        %s
    )
    GROUP BY
        projectName,
        projectDisplayName,
        projectID,
        timestamp,
        empProjectCode
)
WHERE
    projectID != ''
`
	basicSqlGroup = `
GROUP BY
    requirementTotal,
    bugTotal,
    taskTotal,
    responsibleFuncPointsTotal,
    requirementFuncPointsTotal,
    devFuncPointsTotal,
    demandFuncPointsTotal,
    testFuncPointsTotal,
    budgetMandayTotal,
    taskEstimatedMinute,
    taskEstimatedManday,
    actualMandayTotal,
    requirementDoneRate,
    taskDoneTotal,
    taskDoneRate,
    taskEstimatedDayGtOneTotal,
    taskEstimatedDayGtTwoTotal,
    taskEstimatedDayGtThreeTotal,
    requirementDoneTotal,
    requirementAssociatedTotal,
    requirementAssociatedRate,
    requirementUnassignedTotal,
    requirementUnassignedRate,
    taskUnassignedTotal,
    bugUndoneTotal,
    bugDoneRate,
    bugSeriousTotal,
    bugSeriousRate,
    bugDemandDesignTotal,
    bugDemandDesignRate,
    bugOnlineTotal,
    bugOnlineRate,
    bugReopenTotal,
    bugReopenRate,
    taskAssociatedTotal,
    taskAssociatedRate,
    bugLowLevelTotal,
    bugLowLevelRate,
    taskWorkingTotal,
    iterationCompletedRate,
    bugWontfixTotal,
    iterationAssigneeTotal,
    iterationEstimatedDayTotal,
    projectName,
    projectDisplayName,
    projectID,
    timestamp,
    empProjectCode
ORDER BY
    timestamp ASC,
    projectID ASC
`
	lastValueBasicSql = `SELECT 
            last_value(projectName) as projectName,
            last_value(projectDisplayName) as projectDisplayName,
            projectID as projectID,
            iteration_id as iteration_id,
            last_value(empProjectCode) as empProjectCode,
            toStartOfInterval(timestamp, INTERVAL 1 day) as timestamp,
            last_value(requirement_total) as requirement_total,
            last_value(bug_total) as bug_total,
            last_value(task_total) as task_total,
            last_value(responsible_func_points_total) as responsible_func_points_total,
            last_value(requirement_func_points_total) as requirement_func_points_total,
            last_value(dev_func_points_total) as dev_func_points_total,
            last_value(demand_func_points_total) as demand_func_points_total,
            last_value(test_func_points_total) as test_func_points_total,
            last_value(budget_manday_total) as budget_manday_total,
            last_value(task_estimated_minute) as task_estimated_minute,
            last_value(actual_manday_total) as actual_manday_total,
            last_value(task_done_total) as task_done_total,
            last_value(task_working_total) as task_working_total,
            last_value(task_estimated_day_gt_one_total) as task_estimated_day_gt_one_total,
            last_value(task_estimated_day_gt_two_total) as task_estimated_day_gt_two_total,
            last_value(task_estimated_day_gt_three_total) as task_estimated_day_gt_three_total,
            last_value(unfinished_assignee_total) as unfinished_assignee_total,
            last_value(requirement_done_total) as requirement_done_total,
            last_value(requirement_associated_total) as requirement_associated_total,
            last_value(requirement_unassigned_total) as requirement_unassigned_total,
            last_value(task_unassigned_total) as task_unassigned_total,
            last_value(bug_undone_total) as bug_undone_total,
            last_value(bug_serious_total) as bug_serious_total,
            last_value(bug_demand_design_total) as bug_demand_design_total,
            last_value(bug_online_total) as bug_online_total,
            last_value(bug_reopen_total) as bug_reopen_total,
            last_value(task_associated_total) as task_associated_total,
            last_value(bug_low_level_total) as bug_low_level_total,
            last_value(bug_wontfix_total) as bug_wontfix_total,
            last_value(iteration_unfinished_assignee_total) as iteration_unfinished_assignee_total,
            last_value(iteration_estimated_day_total) as iteration_estimated_day_total,
            last_value(emp_project_begin_date) as emp_project_begin_date,
            last_value(emp_project_end_date) as emp_project_end_date,
            last_value(emp_project_actual_end_date) as emp_project_actual_end_date,
            last_value(iteration_assignees) as iteration_assignees
        FROM (
        %s
        )
		GROUP BY
            projectID,
            iteration_id,
            timestamp
`
	dataSourceSql = `SELECT
                tag_values[indexOf(tag_keys,'project_name')] as projectName,
	            tag_values[indexOf(tag_keys,'project_display_name')] as projectDisplayName,
	            tag_values[indexOf(tag_keys,'project_id')] as projectID,
	            tag_values[indexOf(tag_keys,'iteration_id')] as iteration_id,
	            tag_values[indexOf(tag_keys,'emp_project_code')] as empProjectCode,
	            timestamp,
	            number_field_values[indexOf(number_field_keys,'iteration_requirement_total')] as requirement_total,
	            number_field_values[indexOf(number_field_keys,'iteration_bug_total')] as bug_total,
	            number_field_values[indexOf(number_field_keys,'iteration_task_total')] as task_total,
	            number_field_values[indexOf(number_field_keys,'iteration_responsible_func_points_total')] as responsible_func_points_total,
	            number_field_values[indexOf(number_field_keys,'iteration_requirement_func_points_total')] as requirement_func_points_total,
	            number_field_values[indexOf(number_field_keys,'iteration_dev_func_points_total')] as dev_func_points_total,
	            number_field_values[indexOf(number_field_keys,'iteration_demand_func_points_total')] as demand_func_points_total,
	            number_field_values[indexOf(number_field_keys,'iteration_test_func_points_total')] as test_func_points_total,
	            number_field_values[indexOf(number_field_keys,'emp_project_budget_manday_total')] as budget_manday_total,
	            number_field_values[indexOf(number_field_keys,'iteration_task_estimated_minute')] as task_estimated_minute,
	            number_field_values[indexOf(number_field_keys,'emp_project_actual_manday_total')] as actual_manday_total,
	            number_field_values[indexOf(number_field_keys,'iteration_task_done_total')] as task_done_total,
	            number_field_values[indexOf(number_field_keys,'iteration_task_working_total')] as task_working_total,
                number_field_values[indexOf(number_field_keys,'iteration_task_estimated_day_gt_one_total')] as task_estimated_day_gt_one_total,
                number_field_values[indexOf(number_field_keys,'iteration_task_estimated_day_gt_two_total')] as task_estimated_day_gt_two_total,
                number_field_values[indexOf(number_field_keys,'iteration_task_estimated_day_gt_three_total')] as task_estimated_day_gt_three_total,
	            number_field_values[indexOf(number_field_keys,'project_assignee_total')] as unfinished_assignee_total,
	            number_field_values[indexOf(number_field_keys,'iteration_requirement_done_total')] as requirement_done_total,
	            number_field_values[indexOf(number_field_keys,'iteration_requirement_associated_task_total')] as requirement_associated_total,
	            number_field_values[indexOf(number_field_keys,'requirement_unassigned_total')] as requirement_unassigned_total,
	            number_field_values[indexOf(number_field_keys,'iteration_task_unassociated_total')] as task_unassigned_total,
	            number_field_values[indexOf(number_field_keys,'iteration_bug_undone_total')] as bug_undone_total,
	            number_field_values[indexOf(number_field_keys,'iteration_serious_bug_total')] as bug_serious_total,
	            number_field_values[indexOf(number_field_keys,'iteration_demand_design_bug_total')] as bug_demand_design_total,
	            number_field_values[indexOf(number_field_keys,'online_bug_total')] as bug_online_total,
	            number_field_values[indexOf(number_field_keys,'iteration_reopen_bug_total')] as bug_reopen_total,
	            number_field_values[indexOf(number_field_keys,'iteration_task_inclusion_requirement_total')] as task_associated_total,
	            number_field_values[indexOf(number_field_keys,'low_level_bug_total')] as bug_low_level_total,
	            number_field_values[indexOf(number_field_keys,'iteration_bug_wontfix_total')] as bug_wontfix_total,
	            number_field_values[indexOf(number_field_keys,'iteration_assignee_total')] as iteration_unfinished_assignee_total,
                number_field_values[indexOf(number_field_keys,'iteration_estimated_day_total')] as iteration_estimated_day_total,
                number_field_values[indexOf(number_field_keys,'emp_project_begin_date')] as emp_project_begin_date,
                number_field_values[indexOf(number_field_keys,'emp_project_end_date')] as emp_project_end_date,
                number_field_values[indexOf(number_field_keys,'emp_project_actual_end_date')] as emp_project_actual_end_date,
                if(tag_values[indexOf(tag_keys,'iteration_assignees')] != '', splitByString(',',tag_values[indexOf(tag_keys,'iteration_assignees')]), []) as iteration_assignees
            FROM monitor.external_metrics_all
            WHERE
                metric_group= ? 
                AND timestamp >= ? 
                AND timestamp <= ? 
                AND tag_values[indexOf(tag_keys,'org_id')] = '?'`

	dataSourceOrderBy = `
            ORDER BY
                timestamp ASC`

	lastValueGroupSql = `
        GROUP BY 
            tag_values[indexOf(tag_keys,'project_id')],
            tag_values[indexOf(tag_keys,'project_name')],
            tag_values[indexOf(tag_keys,'project_display_name')],
            timestamp,
            tag_values[indexOf(tag_keys,'iteration_id')],
            tag_values[indexOf(tag_keys,'emp_project_code')]  
		ORDER BY
            timestamp ASC
`
)

func (p *provider) wrapBadRequest(rw http.ResponseWriter, err error) {
	httpserver.WriteErr(rw, strconv.FormatInt(int64(http.StatusBadRequest), 10), err.Error())
}

func (p *provider) queryProjectReport(rw http.ResponseWriter, r *http.Request) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		p.wrapBadRequest(rw, err)
		return
	}
	req := &apistructs.ProjectReportRequest{}
	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.WithError(err).Errorln("failed to read request body")
		p.wrapBadRequest(rw, err)
		return
	}
	if err := json.Unmarshal(bodyData, req); err != nil {
		p.wrapBadRequest(rw, err)
		return
	}
	if req.OrgID == 0 {
		orgID, err := user.GetOrgID(r)
		if err != nil {
			p.wrapBadRequest(rw, fmt.Errorf("missing orgID"))
			return
		}
		req.OrgID = orgID
	}
	if err := p.checkPermission(req, identityInfo); err != nil {
		p.wrapBadRequest(rw, err)
		return
	}
	if err := checkQueryRequest(req); err != nil {
		p.wrapBadRequest(rw, err)
		return
	}
	if !req.IsAdmin && len(req.ProjectIDs) == 0 {
		httpserver.WriteData(rw, []*ProjectReportRow{})
		return
	}
	dataSourceBasic := p.DB.Explain(dataSourceSql, metricGroup, req.Start, req.End, req.OrgID)
	lastValueWhereSql := p.genLastValueWhereSql(req)
	if lastValueWhereSql != "" {
		dataSourceBasic += " " + lastValueWhereSql
	}
	lastValueBasic := fmt.Sprintf(lastValueBasicSql, dataSourceBasic+dataSourceOrderBy)
	basic := fmt.Sprintf(basicSql, lastValueBasic)
	basicWhereSql := p.genBasicWhereSql(req)
	if basicWhereSql != " " {
		basic += basicWhereSql
	}
	basic += basicSqlGroup
	rows, err := p.Clickhouse.Client().Query(r.Context(), basic)
	if err != nil {
		p.wrapBadRequest(rw, err)
		return
	}
	defer rows.Close()
	ans := make([]*ProjectReportRow, 0)
	for rows.Next() {
		row := &ProjectReportRow{}
		if err := rows.ScanStruct(row); err != nil {
			p.wrapBadRequest(rw, err)
			return
		}
		ans = append(ans, row)
	}

	httpserver.WriteData(rw, ans)
}

func (p *provider) checkPermission(req *apistructs.ProjectReportRequest, identityInfo apistructs.IdentityInfo) error {
	if identityInfo.IsInternalClient() {
		return nil
	}
	isOrgManager, err := bdl.IsManager(identityInfo.UserID, apistructs.OrgScope, req.OrgID)
	if err != nil {
		return err
	}
	if isOrgManager && req.IsAdmin {
		return nil
	}
	myProjectIDs, err := p.bdl.GetMyManagedProjectIDs(req.OrgID, identityInfo.UserID)
	if err != nil {
		return err
	}
	if len(req.ProjectIDs) == 0 {
		req.ProjectIDs = myProjectIDs
		return nil
	}

	myProjectIDSet := make(map[uint64]struct{})
	for _, id := range myProjectIDs {
		myProjectIDSet[id] = struct{}{}
	}
	for _, id := range req.ProjectIDs {
		if _, ok := myProjectIDSet[id]; !ok {
			return fmt.Errorf("permission denied, projectID: %d", id)
		}
	}
	return nil
}

func (p *provider) genLastValueWhereSql(req *apistructs.ProjectReportRequest) string {
	var projectIDSql, iterationIDSql, labelQuerySql string
	if len(req.ProjectIDs) > 0 {
		projectIDSql = fmt.Sprintf("AND tag_values[indexOf(tag_keys,'project_id')] IN (%s)", strings.Join(strutil.ToStrSlice(req.ProjectIDs, true), ","))
	}
	if len(req.IterationIDs) > 0 {
		iterationIDSql = fmt.Sprintf("AND tag_values[indexOf(tag_keys,'iteration_id')] IN (%s)", strings.Join(strutil.ToStrSlice(req.IterationIDs, true), ","))
	}
	for _, query := range req.LabelQuerys {
		if query.Key == labelProjectName {
			labelQuerySql += p.DB.Explain("AND (tag_values[indexOf(tag_keys,?)] like ? or tag_values[indexOf(tag_keys,?)] like ?) ", labelProjectName, fmt.Sprintf("%%%s%%", query.Val), labelProjectDisplayName, fmt.Sprintf("%%%s%%", query.Val))
			continue
		}
		if query.Operation == "like" {
			labelQuerySql += p.DB.Explain("AND tag_values[indexOf(tag_keys,?)] like ? ", query.Key, fmt.Sprintf("%%%s%%", query.Val))
			continue
		}
		labelQuerySql += p.DB.Explain(fmt.Sprintf("AND tag_values[indexOf(tag_keys,?)] %s ? ", query.Operation), query.Key, query.Val)
	}
	return projectIDSql + " " + iterationIDSql + " " + labelQuerySql
}

func (p *provider) genBasicWhereSql(req *apistructs.ProjectReportRequest) string {
	var sql string
	for _, operation := range req.Operations {
		sql += p.DB.Explain(fmt.Sprintf("AND ? %s %f ", operation.Operation, operation.Val), operation.Key)
	}
	return sql
}

func checkQueryRequest(req *apistructs.ProjectReportRequest) error {
	if req.OrgID == 0 {
		return fmt.Errorf("orgID required")
	}
	if req.Start == "" {
		return fmt.Errorf("startTime required")
	}
	if req.End == "" {
		return fmt.Errorf("endTime required")
	}
	for _, operation := range req.Operations {
		if !apistructs.IsValidOperator(operation.Operation) {
			return fmt.Errorf("invalid operation %s", operation.Operation)
		}
	}
	for _, query := range req.LabelQuerys {
		if !apistructs.IsValidLabelOperator(query.Operation) {
			return fmt.Errorf("invalid operation %s", query.Operation)
		}
	}
	return nil
}
