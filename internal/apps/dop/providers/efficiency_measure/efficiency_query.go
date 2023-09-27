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
)

type PersonalEfficiencyRow struct {
	UserID                     string  `json:"userID" ch:"userID"`
	UserName                   string  `json:"userName" ch:"userName"`
	UserEmail                  string  `json:"userEmail" ch:"userEmail"`
	UserNickname               string  `json:"userNickname" ch:"userNickname"`
	OrgID                      string  `json:"orgID" ch:"orgID"`
	OrgName                    string  `json:"orgName" ch:"orgName"`
	UserPosition               string  `json:"userPosition" ch:"userPosition"`
	UserPositionLevel          string  `json:"userPositionLevel" ch:"userPositionLevel"`
	JobStatus                  string  `json:"jobStatus" ch:"jobStatus"`
	ProjectName                string  `json:"projectName" ch:"projectName"`
	ProjectDisplayName         string  `json:"projectDisplayName" ch:"projectDisplayName"`
	RequirementTotal           float64 `json:"requirementTotal" ch:"requirementTotal"`
	WorkingRequirementTotal    float64 `json:"workingRequirementTotal" ch:"workingRequirementTotal"`
	PendingRequirementTotal    float64 `json:"pendingRequirementTotal" ch:"pendingRequirementTotal"`
	TaskTotal                  float64 `json:"taskTotal" ch:"taskTotal"`
	WorkingTaskTotal           float64 `json:"workingTaskTotal" ch:"workingTaskTotal"`
	PendingTaskTotal           float64 `json:"pendingTaskTotal" ch:"pendingTaskTotal"`
	BugTotal                   float64 `json:"bugTotal" ch:"bugTotal"`
	OwnerBugTotal              float64 `json:"ownerBugTotal" ch:"ownerBugTotal"`
	PendingBugTotal            float64 `json:"pendingBugTotal" ch:"pendingBugTotal"`
	WorkingBugTotal            float64 `json:"workingBugTotal" ch:"workingBugTotal"`
	DesignBugTotal             float64 `json:"designBugTotal" ch:"designBugTotal"`
	ArchitectureBugTotal       float64 `json:"architectureBugTotal" ch:"architectureBugTotal"`
	SeriousBugTotal            float64 `json:"seriousBugTotal" ch:"seriousBugTotal"`
	ReopenBugTotal             float64 `json:"reopenBugTotal" ch:"reopenBugTotal"`
	SubmitBugTotal             float64 `json:"submitBugTotal" ch:"submitBugTotal"`
	TestCaseTotal              float64 `json:"testCaseTotal" ch:"testCaseTotal"`
	FixBugElapsedMinute        float64 `json:"fixBugElapsedMinute" ch:"fixBugElapsedMinute"`
	FixBugEstimateMinute       float64 `json:"fixBugEstimateMinute" ch:"fixBugEstimateMinute"`
	AvgFixBugElapsedMinute     float64 `json:"avgFixBugElapsedMinute" ch:"avgFixBugElapsedMinute"`
	AvgFixBugEstimateMinute    float64 `json:"avgFixBugEstimateMinute" ch:"avgFixBugEstimateMinute"`
	ResponsibleFuncPointsTotal float64 `json:"responsibleFuncPointsTotal" ch:"responsibleFuncPointsTotal"`
	RequirementFuncPointsTotal float64 `json:"requirementFuncPointsTotal" ch:"requirementFuncPointsTotal"`
	DevFuncPointsTotal         float64 `json:"devFuncPointsTotal" ch:"devFuncPointsTotal"`
	DemandFuncPointsTotal      float64 `json:"demandFuncPointsTotal" ch:"demandFuncPointsTotal"`
	TestFuncPointsTotal        float64 `json:"testFuncPointsTotal" ch:"testFuncPointsTotal"`
	OnlineBugTotal             float64 `json:"onlineBugTotal" ch:"onlineBugTotal"`
	LowLevelBugTotal           float64 `json:"lowLevelBugTotal" ch:"lowLevelBugTotal"`
	OnlineBugRatio             float64 `json:"onlineBugRatio" ch:"onlineBugRatio"`
	LowLevelBugRatio           float64 `json:"lowLevelBugRatio" ch:"lowLevelBugRatio"`
	ResolvedBugTotal           float64 `json:"resolvedBugTotal" ch:"resolvedBugTotal"`
	ActualMandayTotal          float64 `json:"actualMandayTotal" ch:"actualMandayTotal"`
	RequirementDefectDensity   float64 `json:"requirementDefectDensity" ch:"requirementDefectDensity"`
	DemandDefectDensity        float64 `json:"demandDefectDensity" ch:"demandDefectDensity"`
	DevDefectDensity           float64 `json:"devDefectDensity" ch:"devDefectDensity"`
	BugDefectDensity           float64 `json:"bugDefectDensity" ch:"bugDefectDensity"`
	DemandProductPDR           float64 `json:"demandProductPDR" ch:"demandProductPDR"`
	DevProductPDR              float64 `json:"devProductPDR" ch:"devProductPDR"`
	TestProductPDR             float64 `json:"testProductPDR" ch:"testProductPDR"`
	ProjectFuncPointsTotal     float64 `json:"projectFuncPointsTotal" ch:"projectFuncPointsTotal"`
	PointParticipationRatio    float64 `json:"pointParticipationRatio" ch:"pointParticipationRatio"`
	ProductRequirementTotal    float64 `json:"productRequirementTotal" ch:"productRequirementTotal"`
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
			tx = tx.Where("tag_values[indexOf(tag_keys,'project_id')] in (?)", req.ProjectIDs)
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
	       last_value(number_field_values[indexOf(number_field_keys,'personal_requirement_total')]) as requirementTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_working_requirement_total')]) as workingRequirementTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_pending_requirement_total')]) as pendingRequirementTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_task_total')]) as taskTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_working_task_total')]) as workingTaskTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_pending_task_total')]) as pendingTaskTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_bug_total')]) as bug_total,
           last_value(number_field_values[indexOf(number_field_keys,'personal_owner_bug_total')]) as owner_bug_total,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_pending_bug_total')]) as pendingBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_working_bug_total')]) as workingBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_demand_design_bug_total')]) as designBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_architecture_design_bug_total')]) as architectureBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_serious_bug_total')]) as seriousBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_reopen_bug_total')]) as reopenBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_submit_bug_total')]) as submitBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_test_case_total')]) as testCaseTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_fix_bug_elapsed_minute_total')]) as fix_bug_elapsed_minute,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_fix_bug_estimate_minute_total')]) as fix_bug_estimate_minute,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_responsible_func_points_total')]) as responsibleFuncPointsTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_requirement_func_points_total')]) as requirementFuncPointsTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_dev_func_points_total')]) as devFuncPointsTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_demand_func_points_total')]) as demandFuncPointsTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_test_func_points_total')]) as testFuncPointsTotal,
           last_value(number_field_values[indexOf(number_field_keys,'project_func_points_total')]) as projectFuncPointsTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'personal_online_bug_total')]) as onlineBugTotal,
           last_value(number_field_values[indexOf(number_field_keys,'personal_product_requirement_total')]) as productRequirementTotal,
           last_value(number_field_values[indexOf(number_field_keys,'personal_low_level_bug_total')]) as lowLevelBugTotal,
           last_value(number_field_values[indexOf(number_field_keys,'personal_resolved_bug_total')]) as resolvedBugTotal,
	       last_value(number_field_values[indexOf(number_field_keys,'emp_user_actual_manday_total')]) as actualMandayTotal`)
		tx = tx.Group("orgID, userID, projectID")
		return tx.Find(&[]PersonalEfficiencyRow{})
	})
	basicSql := p.DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		tx = tx.Table(fmt.Sprintf("(%s)", dataSql)).Select(`last_value(orgName) as orgName,
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
    sum(requirementTotal) as requirementTotal,
    sum(workingRequirementTotal) as workingRequirementTotal,
    sum(pendingRequirementTotal) as pendingRequirementTotal,
    sum(taskTotal) as taskTotal,
    sum(workingTaskTotal) as workingTaskTotal,
    sum(pendingTaskTotal) as pendingTaskTotal,
    sum(bug_total) as bugTotal,
    sum(pendingBugTotal) as pendingBugTotal,
    sum(workingBugTotal) as workingBugTotal,
    sum(designBugTotal) as designBugTotal,
    sum(architectureBugTotal) as architectureBugTotal,
    sum(seriousBugTotal) as seriousBugTotal,
    sum(reopenBugTotal) as reopenBugTotal,
    sum(submitBugTotal) as submitBugTotal,
    sum(testCaseTotal) as testCaseTotal,
    sum(owner_bug_total) as ownerBugTotal,
    sum(fix_bug_elapsed_minute) as fixBugElapsedMinute,
    sum(fix_bug_estimate_minute) as fixBugEstimateMinute,
    sum(resolvedBugTotal) as resolvedBugTotal,
    if(resolvedBugTotal > 0, sum(fix_bug_elapsed_minute) / resolvedBugTotal, 0) as avgFixBugElapsedMinute,
    if(resolvedBugTotal > 0, sum(fix_bug_estimate_minute) / resolvedBugTotal, 0) as avgFixBugEstimateMinute,
    sum(responsibleFuncPointsTotal) as responsibleFuncPointsTotal,
    sum(requirementFuncPointsTotal) as requirementFuncPointsTotal,
    sum(devFuncPointsTotal) as devFuncPointsTotal,
    sum(demandFuncPointsTotal) as demandFuncPointsTotal,
    sum(testFuncPointsTotal) as testFuncPointsTotal,
    sum(onlineBugTotal) as onlineBugTotal,
    sum(lowLevelBugTotal) as lowLevelBugTotal,
    sum(actualMandayTotal) as actualMandayTotal,
	sum(projectFuncPointsTotal) as projectFuncPointsTotal,
    sum(productRequirementTotal) as productRequirementTotal,
    if(ownerBugTotal > 0, onlineBugTotal / ownerBugTotal, 0) as onlineBugRatio,
    if(ownerBugTotal > 0, lowLevelBugTotal / ownerBugTotal, 0) as lowLevelBugRatio,
    if(projectFuncPointsTotal > 0, responsibleFuncPointsTotal / projectFuncPointsTotal, 0) as pointParticipationRatio,
    if(demandFuncPointsTotal > 0, designBugTotal / demandFuncPointsTotal, 0) as requirementDefectDensity,
    if(demandFuncPointsTotal > 0, architectureBugTotal / demandFuncPointsTotal, 0) as demandDefectDensity,
    if(devFuncPointsTotal > 0, ownerBugTotal / devFuncPointsTotal, 0) as devDefectDensity,
    if(testFuncPointsTotal > 0, onlineBugTotal / testFuncPointsTotal, 0) as bugDefectDensity,
    if(demandFuncPointsTotal > 0, actualMandayTotal * 8 / demandFuncPointsTotal, 0) as demandProductPDR,
    if(devFuncPointsTotal > 0, actualMandayTotal * 8 / devFuncPointsTotal, 0) as devProductPDR,
    if(testFuncPointsTotal > 0, actualMandayTotal * 8 / testFuncPointsTotal, 0) as testProductPDR`).
			Group("orgID, userID")
		if req.GroupByProject {
			tx = tx.Group("projectID")
		}
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
