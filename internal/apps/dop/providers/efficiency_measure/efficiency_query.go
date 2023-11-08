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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"gorm.io/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

type PersonalEfficiencyRow struct {
	UserID             string `json:"userID" ch:"userID"`
	UserName           string `json:"userName" ch:"userName"`
	UserEmail          string `json:"userEmail" ch:"userEmail"`
	UserNickname       string `json:"userNickname" ch:"userNickname"`
	OrgID              string `json:"orgID" ch:"orgID"`
	OrgName            string `json:"orgName" ch:"orgName"`
	ProjectID          string `json:"projectID" ch:"projectID"`
	UserPosition       string `json:"userPosition" ch:"userPosition"`
	UserPositionLevel  string `json:"userPositionLevel" ch:"userPositionLevel"`
	JobStatus          string `json:"jobStatus" ch:"jobStatus"`
	ProjectName        string `json:"projectName" ch:"projectName"`
	ProjectDisplayName string `json:"projectDisplayName" ch:"projectDisplayName"`

	RequirementTotal           float64 `json:"rangeRequirementTotal" ch:"rangeRequirementTotal"`
	WorkingRequirementTotal    float64 `json:"rangeWorkingRequirementTotal" ch:"rangeWorkingRequirementTotal"`
	PendingRequirementTotal    float64 `json:"rangePendingRequirementTotal" ch:"rangePendingRequirementTotal"`
	TaskTotal                  float64 `json:"rangeTaskTotal" ch:"rangeTaskTotal"`
	WorkingTaskTotal           float64 `json:"rangeWorkingTaskTotal" ch:"rangeWorkingTaskTotal"`
	PendingTaskTotal           float64 `json:"rangePendingTaskTotal" ch:"rangePendingTaskTotal"`
	BugTotal                   float64 `json:"rangeBugTotal" ch:"rangeBugTotal"`
	OwnerBugTotal              float64 `json:"rangeOwnerBugTotal" ch:"rangeOwnerBugTotal"`
	PendingBugTotal            float64 `json:"rangePendingBugTotal" ch:"rangePendingBugTotal"`
	WorkingBugTotal            float64 `json:"rangeWorkingBugTotal" ch:"rangeWorkingBugTotal"`
	DesignBugTotal             float64 `json:"rangeDesignBugTotal" ch:"rangeDesignBugTotal"`
	ArchitectureBugTotal       float64 `json:"rangeArchitectureBugTotal" ch:"rangeArchitectureBugTotal"`
	SeriousBugTotal            float64 `json:"rangeSeriousBugTotal" ch:"rangeSeriousBugTotal"`
	ReopenBugTotal             float64 `json:"rangeReopenBugTotal" ch:"rangeReopenBugTotal"`
	SubmitBugTotal             float64 `json:"rangeSubmitBugTotal" ch:"rangeSubmitBugTotal"`
	TestCaseTotal              float64 `json:"rangeTestCaseTotal" ch:"rangeTestCaseTotal"`
	FixBugElapsedMinute        float64 `json:"rangeFixBugElapsedMinute" ch:"rangeFixBugElapsedMinute"`
	FixBugEstimateMinute       float64 `json:"rangeFixBugEstimateMinute" ch:"rangeFixBugEstimateMinute"`
	AvgFixBugElapsedMinute     float64 `json:"rangeAvgFixBugElapsedMinute" ch:"rangeAvgFixBugElapsedMinute"`
	AvgFixBugEstimateMinute    float64 `json:"rangeAvgFixBugEstimateMinute" ch:"rangeAvgFixBugEstimateMinute"`
	ResponsibleFuncPointsTotal float64 `json:"rangeResponsibleFuncPointsTotal" ch:"rangeResponsibleFuncPointsTotal"`
	RequirementFuncPointsTotal float64 `json:"rangeRequirementFuncPointsTotal" ch:"rangeRequirementFuncPointsTotal"`
	DevFuncPointsTotal         float64 `json:"rangeDevFuncPointsTotal" ch:"rangeDevFuncPointsTotal"`
	DemandFuncPointsTotal      float64 `json:"rangeDemandFuncPointsTotal" ch:"rangeDemandFuncPointsTotal"`
	TestFuncPointsTotal        float64 `json:"rangeTestFuncPointsTotal" ch:"rangeTestFuncPointsTotal"`
	OnlineBugTotal             float64 `json:"rangeOnlineBugTotal" ch:"rangeOnlineBugTotal"`
	LowLevelBugTotal           float64 `json:"rangeLowLevelBugTotal" ch:"rangeLowLevelBugTotal"`
	OnlineBugRatio             float64 `json:"rangeOnlineBugRatio" ch:"rangeOnlineBugRatio"`
	LowLevelBugRatio           float64 `json:"rangeLowLevelBugRatio" ch:"rangeLowLevelBugRatio"`
	ResolvedBugTotal           float64 `json:"rangeResolvedBugTotal" ch:"rangeResolvedBugTotal"`
	ActualMandayTotal          float64 `json:"rangeActualMandayTotal" ch:"rangeActualMandayTotal"`
	RequirementDefectDensity   float64 `json:"rangeRequirementDefectDensity" ch:"rangeRequirementDefectDensity"`
	DemandDefectDensity        float64 `json:"rangeDemandDefectDensity" ch:"rangeDemandDefectDensity"`
	DevDefectDensity           float64 `json:"rangeDevDefectDensity" ch:"rangeDevDefectDensity"`
	BugDefectDensity           float64 `json:"rangeBugDefectDensity" ch:"rangeBugDefectDensity"`
	DemandProductPDR           float64 `json:"rangeDemandProductPDR" ch:"rangeDemandProductPDR"`
	DevProductPDR              float64 `json:"rangeDevProductPDR" ch:"rangeDevProductPDR"`
	TestProductPDR             float64 `json:"rangeTestProductPDR" ch:"rangeTestProductPDR"`
	ProjectFuncPointsTotal     float64 `json:"rangeProjectFuncPointsTotal" ch:"rangeProjectFuncPointsTotal"`
	PointParticipationRatio    float64 `json:"rangePointParticipationRatio" ch:"rangePointParticipationRatio"`
	ProductRequirementTotal    float64 `json:"rangeProductRequirementTotal" ch:"rangeProductRequirementTotal"`

	LastRequirementTotal           float64 `json:"requirementTotal" ch:"requirementTotal"`
	LastWorkingRequirementTotal    float64 `json:"workingRequirementTotal" ch:"workingRequirementTotal"`
	LastPendingRequirementTotal    float64 `json:"pendingRequirementTotal" ch:"pendingRequirementTotal"`
	LastTaskTotal                  float64 `json:"taskTotal" ch:"taskTotal"`
	LastWorkingTaskTotal           float64 `json:"workingTaskTotal" ch:"workingTaskTotal"`
	LastPendingTaskTotal           float64 `json:"pendingTaskTotal" ch:"pendingTaskTotal"`
	LastBugTotal                   float64 `json:"bugTotal" ch:"bugTotal"`
	LastOwnerBugTotal              float64 `json:"ownerBugTotal" ch:"ownerBugTotal"`
	LastPendingBugTotal            float64 `json:"pendingBugTotal" ch:"pendingBugTotal"`
	LastWorkingBugTotal            float64 `json:"workingBugTotal" ch:"workingBugTotal"`
	LastDesignBugTotal             float64 `json:"designBugTotal" ch:"designBugTotal"`
	LastArchitectureBugTotal       float64 `json:"architectureBugTotal" ch:"architectureBugTotal"`
	LastSeriousBugTotal            float64 `json:"seriousBugTotal" ch:"seriousBugTotal"`
	LastReopenBugTotal             float64 `json:"reopenBugTotal" ch:"reopenBugTotal"`
	LastSubmitBugTotal             float64 `json:"submitBugTotal" ch:"submitBugTotal"`
	LastTestCaseTotal              float64 `json:"testCaseTotal" ch:"testCaseTotal"`
	LastFixBugElapsedMinute        float64 `json:"fixBugElapsedMinute" ch:"fixBugElapsedMinute"`
	LastFixBugEstimateMinute       float64 `json:"fixBugEstimateMinute" ch:"fixBugEstimateMinute"`
	LastAvgFixBugElapsedMinute     float64 `json:"avgFixBugElapsedMinute" ch:"avgFixBugElapsedMinute"`
	LastAvgFixBugEstimateMinute    float64 `json:"avgFixBugEstimateMinute" ch:"avgFixBugEstimateMinute"`
	LastResponsibleFuncPointsTotal float64 `json:"responsibleFuncPointsTotal" ch:"responsibleFuncPointsTotal"`
	LastRequirementFuncPointsTotal float64 `json:"requirementFuncPointsTotal" ch:"requirementFuncPointsTotal"`
	LastDevFuncPointsTotal         float64 `json:"devFuncPointsTotal" ch:"devFuncPointsTotal"`
	LastDemandFuncPointsTotal      float64 `json:"demandFuncPointsTotal" ch:"demandFuncPointsTotal"`
	LastTestFuncPointsTotal        float64 `json:"testFuncPointsTotal" ch:"testFuncPointsTotal"`
	LastOnlineBugTotal             float64 `json:"onlineBugTotal" ch:"onlineBugTotal"`
	LastLowLevelBugTotal           float64 `json:"lowLevelBugTotal" ch:"lowLevelBugTotal"`
	LastOnlineBugRatio             float64 `json:"onlineBugRatio" ch:"onlineBugRatio"`
	LastLowLevelBugRatio           float64 `json:"lowLevelBugRatio" ch:"lowLevelBugRatio"`
	LastResolvedBugTotal           float64 `json:"resolvedBugTotal" ch:"resolvedBugTotal"`
	LastActualMandayTotal          float64 `json:"actualMandayTotal" ch:"actualMandayTotal"`
	LastRequirementDefectDensity   float64 `json:"requirementDefectDensity" ch:"requirementDefectDensity"`
	LastDemandDefectDensity        float64 `json:"demandDefectDensity" ch:"demandDefectDensity"`
	LastDevDefectDensity           float64 `json:"devDefectDensity" ch:"devDefectDensity"`
	LastBugDefectDensity           float64 `json:"bugDefectDensity" ch:"bugDefectDensity"`
	LastDemandProductPDR           float64 `json:"demandProductPDR" ch:"demandProductPDR"`
	LastDevProductPDR              float64 `json:"devProductPDR" ch:"devProductPDR"`
	LastTestProductPDR             float64 `json:"testProductPDR" ch:"testProductPDR"`
	LastProjectFuncPointsTotal     float64 `json:"projectFuncPointsTotal" ch:"projectFuncPointsTotal"`
	LastPointParticipationRatio    float64 `json:"pointParticipationRatio" ch:"pointParticipationRatio"`
	LastProductRequirementTotal    float64 `json:"productRequirementTotal" ch:"productRequirementTotal"`
}

func (p *provider) wrapBadRequest(rw http.ResponseWriter, err error) {
	httpserver.WriteErr(rw, strconv.FormatInt(int64(http.StatusBadRequest), 10), err.Error())
}

func (p *provider) queryPersonalEfficiency(rw http.ResponseWriter, r *http.Request) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		p.wrapBadRequest(rw, err)
		return
	}
	req := &apistructs.PersonalEfficiencyRequest{}
	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
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
	if req.UserID == 0 {
		userID, err := strconv.ParseUint(identityInfo.UserID, 10, 64)
		if err != nil {
			p.wrapBadRequest(rw, fmt.Errorf("invalid userID: %s", identityInfo.UserID))
			return
		}
		req.UserID = userID
	}
	if err := checkQueryRequest(req); err != nil {
		p.wrapBadRequest(rw, err)
		return
	}

	rawSql := p.makeEfficiencyBasicSql(req)
	rows, err := p.Clickhouse.Client().Query(r.Context(), rawSql)
	if err != nil {
		p.wrapBadRequest(rw, err)
		return
	}
	defer rows.Close()
	ans := make([]*PersonalEfficiencyRow, 0)
	for rows.Next() {
		row := &PersonalEfficiencyRow{}
		if err := rows.ScanStruct(row); err != nil {
			p.wrapBadRequest(rw, err)
			return
		}
		ans = append(ans, row)
	}
	httpserver.WriteData(rw, ans)
}

func (p *provider) makeEfficiencyBasicSql(req *apistructs.PersonalEfficiencyRequest) string {
	sourceSql := p.DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		tx = tx.Table("monitor.external_metrics_all").Select("*").Where("metric_group='performance_measure'").
			Where("timestamp >= ?", req.Start).
			Where("timestamp <= ?", req.End).
			Where("tag_values[indexOf(tag_keys,'org_id')] = '?'", req.OrgID).
			Order("timestamp ASC")
		if req.UserID != 0 {
			tx = tx.Where("tag_values[indexOf(tag_keys,'user_id')] = '?'", req.UserID)
		}
		if len(req.ProjectIDs) > 0 {
			tx = tx.Where("tag_values[indexOf(tag_keys,'project_id')] in (?)", strutil.ToStrSlice(req.ProjectIDs))
		}
		for _, query := range req.LabelQuerys {
			if query.Operation == "like" {
				tx = tx.Where(fmt.Sprintf("tag_values[indexOf(tag_keys,'%s')] like '%%%s%%'", query.Key, query.Val))
				continue
			}
			tx = tx.Where(fmt.Sprintf("tag_values[indexOf(tag_keys,'%s')] %s '%s'", query.Key, query.Operation, query.Val))
		}
		return tx.Find(&[]PersonalEfficiencyRow{})
	})
	dataSql := p.DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		tx = tx.Table(fmt.Sprintf("(%s)", sourceSql)).
			Select(`last_value(tag_values[indexOf(tag_keys,'org_name')]) as orgName,
            last_value(tag_values[indexOf(tag_keys,'user_name')]) as userName,
            last_value(tag_values[indexOf(tag_keys,'user_nickname')]) as userNickname,
            last_value(tag_values[indexOf(tag_keys,'user_email')]) as userEmail,
            max(tag_values[indexOf(tag_keys,'emp_user_position')]) as userPosition,
            max(tag_values[indexOf(tag_keys,'emp_user_position_level')]) as userPositionLevel,
            max(tag_values[indexOf(tag_keys,'emp_job_status')]) as jobStatus,
            last_value(tag_values[indexOf(tag_keys,'project_name')]) as projectName,
            last_value(tag_values[indexOf(tag_keys,'project_display_name')]) as projectDisplayName,
	       tag_values[indexOf(tag_keys,'org_id')] as orgID,
	       tag_values[indexOf(tag_keys,'user_id')] as userID,
	       tag_values[indexOf(tag_keys,'project_id')] as projectID,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_requirement_total')]) as firstRequirementTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_requirement_total')]) as lastRequirementTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_working_requirement_total')]) as firstWorkingRequirementTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_working_requirement_total')]) as workingRequirementTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_pending_requirement_total')]) as firstPendingRequirementTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_pending_requirement_total')]) as pendingRequirementTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_task_total')]) as firstTaskTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_task_total')]) as taskTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_working_task_total')]) as firstWorkingTaskTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_working_task_total')]) as workingTaskTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_pending_task_total')]) as firstPendingTaskTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_pending_task_total')]) as pendingTaskTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_bug_total')]) as first_bug_total,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_bug_total')]) as bug_total,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_owner_bug_total')]) as first_owner_bug_total,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_owner_bug_total')]) as owner_bug_total,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_pending_bug_total')]) as firstPendingBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_pending_bug_total')]) as pendingBugTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_working_bug_total')]) as firstWorkingBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_working_bug_total')]) as workingBugTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_demand_design_bug_total')]) as firstDesignBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_demand_design_bug_total')]) as designBugTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_architecture_design_bug_total')]) as firstArchitectureBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_architecture_design_bug_total')]) as architectureBugTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_serious_bug_total')]) as firstSeriousBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_serious_bug_total')]) as seriousBugTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_reopen_bug_total')]) as firstReopenBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_reopen_bug_total')]) as reopenBugTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_submit_bug_total')]) as firstSubmitBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_submit_bug_total')]) as submitBugTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_test_case_total')]) as firstTestCaseTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_test_case_total')]) as testCaseTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_fix_bug_elapsed_minute_total')]) as first_fix_bug_elapsed_minute,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_fix_bug_elapsed_minute_total')]) as fix_bug_elapsed_minute,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_fix_bug_estimate_minute_total')]) as first_fix_bug_estimate_minute,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_fix_bug_estimate_minute_total')]) as fix_bug_estimate_minute,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_responsible_func_points_total')]) as firstResponsibleFuncPointsTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_responsible_func_points_total')]) as responsibleFuncPointsTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_requirement_func_points_total')]) as firstRequirementFuncPointsTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_requirement_func_points_total')]) as requirementFuncPointsTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_dev_func_points_total')]) as firstDevFuncPointsTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_dev_func_points_total')]) as devFuncPointsTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_demand_func_points_total')]) as firstDemandFuncPointsTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_demand_func_points_total')]) as demandFuncPointsTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_test_func_points_total')]) as firstTestFuncPointsTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_test_func_points_total')]) as testFuncPointsTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'project_func_points_total')]) as firstProjectFuncPointsTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'project_func_points_total')]) as projectFuncPointsTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_online_bug_total')]) as firstOnlineBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_online_bug_total')]) as onlineBugTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_product_requirement_total')]) as firstProductRequirementTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_product_requirement_total')]) as productRequirementTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_low_level_bug_total')]) as firstLowLevelBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_low_level_bug_total')]) as lowLevelBugTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'personal_resolved_bug_total')]) as firstResolvedBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_resolved_bug_total')]) as resolvedBugTotal,
	       first_value(number_field_values[indexOf(number_field_keys,'emp_user_actual_manday_total')]) as firstActualMandayTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'emp_user_actual_manday_total')]) as actualMandayTotal`)
		tx = tx.Group("orgID, userID, projectID")
		return tx.Find(&[]PersonalEfficiencyRow{})
	})
	basicSql := p.DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		selectSql := `last_value(orgName) as orgName,
    last_value(userName) as userName,
    last_value(userEmail) as userEmail,
    last_value(userNickname) as userNickname,
    orgID,
    userID,
    max(userPosition) as userPosition,
    max(userPositionLevel) as userPositionLevel,
    max(jobStatus) as jobStatus,
    last_value(projectName) as projectName,
    last_value(projectDisplayName) as projectDisplayName,
    sum(lastRequirementTotal) as requirementTotal,
    if(requirementTotal - sum(firstRequirementTotal) > 0, requirementTotal - sum(firstRequirementTotal), 0) as rangeRequirementTotal,
    sum(workingRequirementTotal) as workingRequirementTotal,
    if(workingRequirementTotal - sum(firstWorkingRequirementTotal) > 0, workingRequirementTotal - sum(firstWorkingRequirementTotal), 0) as rangeWorkingRequirementTotal,
    sum(pendingRequirementTotal) as pendingRequirementTotal,
    if(pendingRequirementTotal - sum(firstPendingRequirementTotal) > 0, pendingRequirementTotal - sum(firstPendingRequirementTotal), 0) as rangePendingRequirementTotal,
    sum(taskTotal) as taskTotal,
    if(taskTotal - sum(firstTaskTotal) > 0, taskTotal - sum(firstTaskTotal), 0) as rangeTaskTotal,
    sum(workingTaskTotal) as workingTaskTotal,
    if(workingTaskTotal - sum(firstWorkingTaskTotal) > 0, workingTaskTotal - sum(firstWorkingTaskTotal), 0) as rangeWorkingTaskTotal,
    sum(pendingTaskTotal) as pendingTaskTotal,
    if(pendingTaskTotal - sum(firstPendingTaskTotal) > 0, pendingTaskTotal - sum(firstPendingTaskTotal), 0) as rangePendingTaskTotal,
    sum(bug_total) as bugTotal,
    if(bugTotal - sum(first_bug_total) > 0, bugTotal - sum(first_bug_total), 0) as rangeBugTotal,
    sum(pendingBugTotal) as pendingBugTotal,
    if(pendingBugTotal - sum(firstPendingBugTotal) > 0, pendingBugTotal - sum(firstPendingBugTotal), 0) as rangePendingBugTotal,
    sum(workingBugTotal) as workingBugTotal,
    if(workingBugTotal - sum(firstWorkingBugTotal) > 0, workingBugTotal - sum(firstWorkingBugTotal), 0) as rangeWorkingBugTotal,
    sum(designBugTotal) as designBugTotal,
    if(designBugTotal - sum(firstDesignBugTotal) > 0, designBugTotal - sum(firstDesignBugTotal), 0) as rangeDesignBugTotal,
    sum(architectureBugTotal) as architectureBugTotal,
    if(architectureBugTotal - sum(firstArchitectureBugTotal) > 0, architectureBugTotal - sum(firstArchitectureBugTotal), 0) as rangeArchitectureBugTotal,
    sum(seriousBugTotal) as seriousBugTotal,
    if(seriousBugTotal - sum(firstSeriousBugTotal) > 0, seriousBugTotal - sum(firstSeriousBugTotal), 0) as rangeSeriousBugTotal,
    sum(reopenBugTotal) as reopenBugTotal,
    if(reopenBugTotal - sum(firstReopenBugTotal) > 0, reopenBugTotal - sum(firstReopenBugTotal), 0) as rangeReopenBugTotal,
    sum(submitBugTotal) as submitBugTotal,
    if(submitBugTotal - sum(firstSubmitBugTotal) > 0, submitBugTotal - sum(firstSubmitBugTotal), 0) as rangeSubmitBugTotal,
    sum(testCaseTotal) as testCaseTotal,
    if(testCaseTotal - sum(firstTestCaseTotal) > 0, testCaseTotal - sum(firstTestCaseTotal), 0) as rangeTestCaseTotal,
    sum(owner_bug_total) as ownerBugTotal,
    if(ownerBugTotal - sum(first_owner_bug_total) > 0, ownerBugTotal - sum(first_owner_bug_total), 0) as rangeOwnerBugTotal,
    sum(fix_bug_elapsed_minute) as fixBugElapsedMinute,
    if(fixBugElapsedMinute - sum(first_fix_bug_elapsed_minute) > 0, fixBugElapsedMinute - sum(first_fix_bug_elapsed_minute), 0) as rangeFixBugElapsedMinute,
    sum(fix_bug_estimate_minute) as fixBugEstimateMinute,
    if(fixBugEstimateMinute - sum(first_fix_bug_estimate_minute) > 0, fixBugEstimateMinute - sum(first_fix_bug_estimate_minute), 0) as rangeFixBugEstimateMinute,
    sum(resolvedBugTotal) as resolvedBugTotal,
    if(resolvedBugTotal - sum(firstResolvedBugTotal) > 0, resolvedBugTotal - sum(firstResolvedBugTotal), 0) as rangeResolvedBugTotal,
    if(rangeResolvedBugTotal > 0, rangeFixBugElapsedMinute / rangeResolvedBugTotal, 0) as rangeAvgFixBugElapsedMinute,
    if(resolvedBugTotal > 0, sum(fix_bug_elapsed_minute) / resolvedBugTotal, 0) as avgFixBugElapsedMinute,
    if(rangeResolvedBugTotal > 0, rangeFixBugEstimateMinute / rangeResolvedBugTotal, 0) as rangeAvgFixBugEstimateMinute,
    if(resolvedBugTotal > 0, sum(fix_bug_estimate_minute) / resolvedBugTotal, 0) as avgFixBugEstimateMinute,
    if(responsibleFuncPointsTotal - sum(firstResponsibleFuncPointsTotal) > 0, responsibleFuncPointsTotal - sum(firstResponsibleFuncPointsTotal), 0) as rangeResponsibleFuncPointsTotal,
    sum(responsibleFuncPointsTotal) as responsibleFuncPointsTotal,
    if(requirementFuncPointsTotal - sum(firstRequirementFuncPointsTotal) > 0, requirementFuncPointsTotal - sum(firstRequirementFuncPointsTotal), 0) as rangeRequirementFuncPointsTotal,
    sum(requirementFuncPointsTotal) as requirementFuncPointsTotal,
    if(devFuncPointsTotal - sum(firstDevFuncPointsTotal) > 0, devFuncPointsTotal - sum(firstDevFuncPointsTotal), 0) as rangeDevFuncPointsTotal,
    sum(devFuncPointsTotal) as devFuncPointsTotal,
    if(demandFuncPointsTotal - sum(firstDemandFuncPointsTotal) > 0, demandFuncPointsTotal - sum(firstDemandFuncPointsTotal), 0) as rangeDemandFuncPointsTotal,
    sum(demandFuncPointsTotal) as demandFuncPointsTotal,
    if(testFuncPointsTotal - sum(firstTestFuncPointsTotal) > 0, testFuncPointsTotal - sum(firstTestFuncPointsTotal), 0) as rangeTestFuncPointsTotal,
    sum(testFuncPointsTotal) as testFuncPointsTotal,
    if(onlineBugTotal - sum(firstOnlineBugTotal) > 0, onlineBugTotal - sum(firstOnlineBugTotal), 0) as rangeOnlineBugTotal,
    sum(onlineBugTotal) as onlineBugTotal,
    if(lowLevelBugTotal - sum(firstLowLevelBugTotal) > 0, lowLevelBugTotal - sum(firstLowLevelBugTotal), 0) as rangeLowLevelBugTotal,
    sum(lowLevelBugTotal) as lowLevelBugTotal,
    sum(actualMandayTotal) as actualMandayTotal,
    if(actualMandayTotal - sum(firstActualMandayTotal) > 0, actualMandayTotal - sum(firstActualMandayTotal), 0) as rangeActualMandayTotal,
    if(projectFuncPointsTotal - sum(firstProjectFuncPointsTotal) > 0, projectFuncPointsTotal - sum(firstProjectFuncPointsTotal), 0) as rangeProjectFuncPointsTotal,
    sum(projectFuncPointsTotal) as projectFuncPointsTotal,
    if(productRequirementTotal - sum(firstProductRequirementTotal) > 0, productRequirementTotal - sum(firstProductRequirementTotal), 0) as rangeProductRequirementTotal,
    sum(productRequirementTotal) as productRequirementTotal,
    if(rangeOwnerBugTotal > 0, rangeOnlineBugTotal / rangeOwnerBugTotal, 0) as rangeOnlineBugRatio,
    if(ownerBugTotal > 0, onlineBugTotal / ownerBugTotal, 0) as onlineBugRatio,
    if(rangeOwnerBugTotal > 0, rangeLowLevelBugTotal / rangeOwnerBugTotal, 0) as rangeLowLevelBugRatio,
    if(ownerBugTotal > 0, lowLevelBugTotal / ownerBugTotal, 0) as lowLevelBugRatio,
    if(rangeProjectFuncPointsTotal > 0, rangeResponsibleFuncPointsTotal / rangeProjectFuncPointsTotal, 0) as rangePointParticipationRatio,
    if(projectFuncPointsTotal > 0, responsibleFuncPointsTotal / projectFuncPointsTotal, 0) as pointParticipationRatio,
    if(rangeDemandFuncPointsTotal > 0, rangeDesignBugTotal / rangeDemandFuncPointsTotal, 0) as rangeRequirementDefectDensity,
    if(demandFuncPointsTotal > 0, designBugTotal / demandFuncPointsTotal, 0) as requirementDefectDensity,
    if(rangeDemandFuncPointsTotal > 0, rangeArchitectureBugTotal / rangeDemandFuncPointsTotal, 0) as rangeDemandDefectDensity,
    if(demandFuncPointsTotal > 0, architectureBugTotal / demandFuncPointsTotal, 0) as demandDefectDensity,
    if(rangeDevFuncPointsTotal > 0, rangeOwnerBugTotal / rangeDevFuncPointsTotal, 0) as rangeDevDefectDensity,
    if(devFuncPointsTotal > 0, ownerBugTotal / devFuncPointsTotal, 0) as devDefectDensity,
    if(rangeTestFuncPointsTotal > 0, rangeOnlineBugTotal / rangeTestFuncPointsTotal, 0) as rangeBugDefectDensity,
    if(testFuncPointsTotal > 0, onlineBugTotal / testFuncPointsTotal, 0) as bugDefectDensity,
    if(rangeDemandFuncPointsTotal > 0, rangeActualMandayTotal * 8 / rangeDemandFuncPointsTotal, 0) as rangeDemandProductPDR,
    if(demandFuncPointsTotal > 0, actualMandayTotal * 8 / demandFuncPointsTotal, 0) as demandProductPDR,
    if(rangeDevFuncPointsTotal > 0, rangeActualMandayTotal * 8 / rangeDevFuncPointsTotal, 0) as rangeDevProductPDR,
    if(devFuncPointsTotal > 0, actualMandayTotal * 8 / devFuncPointsTotal, 0) as devProductPDR,
    if(rangeTestFuncPointsTotal > 0, rangeActualMandayTotal * 8 / rangeTestFuncPointsTotal, 0) as rangeTestProductPDR,
    if(testFuncPointsTotal > 0, actualMandayTotal * 8 / testFuncPointsTotal, 0) as testProductPDR`
		if req.GroupByProject {
			selectSql += ", projectID"
			tx = tx.Group("projectID")
		}
		tx = tx.Table(fmt.Sprintf("(%s)", dataSql)).Select(selectSql).
			Group("orgID, userID")
		return tx.Find(&[]PersonalEfficiencyRow{})
	})
	return basicSql
}

func checkQueryRequest(req *apistructs.PersonalEfficiencyRequest) error {
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
