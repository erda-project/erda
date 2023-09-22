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

package performance_measure

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

type PersonalContributorRow struct {
	OrgID            string  `json:"orgID" ch:"orgID"`
	UserEmail        string  `json:"userEmail" ch:"userEmail"`
	ProjectName      string  `json:"projectName" ch:"projectName"`
	CommitTotal      float64 `json:"commitTotal" ch:"commitTotal"`
	FileChangedTotal float64 `json:"fileChangedTotal" ch:"fileChangedTotal"`
	AdditionTotal    float64 `json:"additionTotal" ch:"additionTotal"`
	DeletionTotal    float64 `json:"deletionTotal" ch:"deletionTotal"`
}

type PersonalActualMandayRow struct {
	OrgID             string  `json:"orgID" ch:"orgID"`
	OrgName           string  `json:"orgName" ch:"orgName"`
	UserID            string  `json:"userID" ch:"userID"`
	UserName          string  `json:"userName" ch:"userName"`
	ProjectID         string  `json:"projectID" ch:"projectID"`
	ProjectName       string  `json:"projectName" ch:"projectName"`
	ProjectCode       string  `json:"projectCode" ch:"projectCode"`
	ActualMandayTotal float64 `json:"actualMandayTotal" ch:"actualMandayTotal"`
}

func (p *provider) queryPersonalContributors(rw http.ResponseWriter, r *http.Request) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		p.wrapBadRequest(rw, err)
		return
	}
	req := &apistructs.PersonalContributorRequest{}
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
	if len(req.UserEmail) == 0 {
		p.wrapBadRequest(rw, fmt.Errorf("missing user email"))
		return
	}

	contributorSql := p.makeContributorBasicSql(req)
	contributorRows, err := p.Clickhouse.Client().Query(r.Context(), contributorSql)
	if err != nil {
		p.wrapBadRequest(rw, err)
		return
	}
	defer contributorRows.Close()
	contributors := make([]*PersonalContributorRow, 0)
	for contributorRows.Next() {
		row := &PersonalContributorRow{}
		if err := contributorRows.ScanStruct(row); err != nil {
			p.wrapBadRequest(rw, err)
			return
		}
		contributors = append(contributors, row)
	}

	actualMandaySql := p.makeActualManDaySql(req)
	mandayRows, err := p.Clickhouse.Client().Query(r.Context(), actualMandaySql)
	if err != nil {
		p.wrapBadRequest(rw, err)
		return
	}
	defer mandayRows.Close()
	mandays := make([]*PersonalActualMandayRow, 0)
	for mandayRows.Next() {
		row := &PersonalActualMandayRow{}
		if err := mandayRows.ScanStruct(row); err != nil {
			p.wrapBadRequest(rw, err)
			return
		}
		mandays = append(mandays, row)
	}

	httpserver.WriteData(rw, map[string]interface{}{
		"codeContributors":   contributors,
		"mandayContributors": mandays,
	})
}

func (p *provider) makeContributorBasicSql(req *apistructs.PersonalContributorRequest) string {
	dataSql := p.DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		tx = tx.Table("monitor.external_metrics_all").
			Select(`last_value(tag_values[indexOf(tag_keys,'org_name')]) as orgName,
            last_value(tag_values[indexOf(tag_keys,'user_name')]) as userName,
            last_value(tag_values[indexOf(tag_keys,'app_name')]) as appName,
            tag_values[indexOf(tag_keys,'app_id')] as appID,
            last_value(tag_values[indexOf(tag_keys,'project_name')]) as projectName,
            tag_values[indexOf(tag_keys,'repo_id')] as repoID,
            tag_values[indexOf(tag_keys,'project_id')] as projectID,
            tag_values[indexOf(tag_keys,'org_id')] as orgID,
            tag_values[indexOf(tag_keys,'user_email')] as userEmail,
            toStartOfInterval(timestamp, INTERVAL 1 day) as timestamp,
            last_value(number_field_values[indexOf(number_field_keys,'personal_daily_commits_total')]) as dailyCommitTotal,
            last_value(number_field_values[indexOf(number_field_keys,'personal_daily_files_changed_total')]) as dailyFileChangedTotal,
            last_value(number_field_values[indexOf(number_field_keys,'personal_daily_addition_total')]) as dailyAdditionTotal,
            last_value(number_field_values[indexOf(number_field_keys,'personal_daily_deletion_total')]) as dailyDeletionTotal`).
			Where("metric_group='personal_contributor'").
			Where("timestamp >= ?", req.Start).
			Where("timestamp <= ?", req.End).
			Where("tag_values[indexOf(tag_keys,'org_id')] = '?'", req.OrgID).
			Where("tag_values[indexOf(tag_keys,'user_email')] = ?", req.UserEmail).
			Group("orgID").
			Group("userEmail").
			Group("projectID").
			Group("appID").
			Group("repoID").
			Group("timestamp")
		if len(req.ProjectIDs) > 0 {
			tx = tx.Where("tag_values[indexOf(tag_keys,'org_id')] in (?)", req.ProjectIDs)
		}
		return tx.Find(&[]PersonalContributorRow{})
	})
	basicSql := p.DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		selectSql := `orgID,
    userEmail,
    sum(dailyCommitTotal) as commitTotal,
    sum(dailyFileChangedTotal) as fileChangedTotal,
    sum(dailyAdditionTotal) as additionTotal,
    sum(dailyDeletionTotal) as deletionTotal`
		if req.GroupByProject {
			selectSql += ", last_value(projectName) as projectName"
		}
		tx = tx.Table(fmt.Sprintf("(%s)", dataSql)).
			Select(selectSql).
			Group("orgID").
			Group("userEmail")
		if req.GroupByProject {
			tx = tx.Group("projectID")
		}
		return tx.Find(&[]PersonalContributorRow{})
	})
	return basicSql
}

func (p *provider) makeActualManDaySql(req *apistructs.PersonalContributorRequest) string {
	basicSq := p.DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		tx = tx.Table("monitor.external_metrics_all").
			Select(`last_value(tag_values[indexOf(tag_keys,'org_name')]) as orgName,
    last_value(tag_values[indexOf(tag_keys,'user_name')]) as userName,
    last_value(tag_values[indexOf(tag_keys,'project_name')]) as projectName,
	last_value(tag_values[indexOf(tag_keys,'emp_project_code')]) as projectCode,
    tag_values[indexOf(tag_keys,'project_id')] as projectID,
    tag_values[indexOf(tag_keys,'org_id')] as orgID,
    tag_values[indexOf(tag_keys,'user_id')] as userID,
    max(number_field_values[indexOf(number_field_keys,'emp_user_actual_manday_total')]) as actualMandayTotal`).
			Where("metric_group='performance_measure'").
			Where("timestamp >= ?", req.Start).
			Where("timestamp <= ?", req.End).
			Where("tag_values[indexOf(tag_keys,'org_id')] = '?'", req.OrgID).
			Where("tag_values[indexOf(tag_keys,'user_id')] = '?'", req.UserID).
			Where("tag_values[indexOf(tag_keys,'emp_project_code')] != ''").
			Group("orgID, projectID, userID")
		return tx.Find(&[]PersonalActualMandayRow{})
	})
	return basicSq
}
