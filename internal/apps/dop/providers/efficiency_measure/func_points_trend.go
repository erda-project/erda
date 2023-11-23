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
	"time"

	"gorm.io/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

type FuncPointTrendRow struct {
	OrgID                      string    `json:"orgID" ch:"orgID"`
	UserID                     string    `json:"userID" ch:"userID"`
	ResponsibleFuncPointsTotal float64   `json:"responsibleFuncPointsTotal" ch:"responsibleFuncPointsTotal"`
	ProductPDR                 float64   `json:"productPDR" ch:"productPDR"`
	Timestamp                  time.Time `json:"timestamp" ch:"timestamp"`
}

func (p *provider) queryFuncPointTrend(rw http.ResponseWriter, r *http.Request) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		p.wrapBadRequest(rw, err)
		return
	}
	req := &apistructs.FuncPointTrendRequest{}
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

	rawSql := p.makeFuncPointTrendSql(req)
	rows, err := p.Clickhouse.Client().Query(r.Context(), rawSql)
	if err != nil {
		p.wrapBadRequest(rw, err)
		return
	}
	defer rows.Close()
	ans := make([]*FuncPointTrendRow, 0)
	for rows.Next() {
		row := &FuncPointTrendRow{}
		if err := rows.ScanStruct(row); err != nil {
			p.wrapBadRequest(rw, err)
			return
		}
		ans = append(ans, row)
	}
	httpserver.WriteData(rw, ans)
}

func (p *provider) makeFuncPointTrendSql(req *apistructs.FuncPointTrendRequest) string {
	dataSql := p.DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		tx = tx.Table("monitor.external_metrics_all").
			Select(`last_value(tag_values[indexOf(tag_keys,'org_name')]) as orgName,
            last_value(tag_values[indexOf(tag_keys,'user_name')]) as userName,
            last_value(tag_values[indexOf(tag_keys,'app_name')]) as appName,
            tag_values[indexOf(tag_keys,'project_id')] as projectID,
            tag_values[indexOf(tag_keys,'org_id')] as orgID,
            tag_values[indexOf(tag_keys,'user_id')] as userID,
            toStartOfInterval(timestamp, INTERVAL 1 day) as timestamp,
            max(number_field_values[indexOf(number_field_keys,'personal_responsible_func_points_total')]) as responsibleFuncPointsTotal,
            max(number_field_values[indexOf(number_field_keys,'emp_user_actual_manday_total')]) as actualMandayTotal`).
			Where("metric_group = 'performance_measure'").
			Where("timestamp >= ?", req.Start).
			Where("timestamp <= ?", req.End).
			Where("tag_values[indexOf(tag_keys,'org_id')] = '?'", req.OrgID).
			Where("tag_values[indexOf(tag_keys,'user_id')] = '?'", req.UserID).
			Group("orgID").
			Group("userID").
			Group("projectID").
			Group("timestamp")
		if len(req.ProjectIDs) > 0 {
			tx = tx.Where("tag_values[indexOf(tag_keys,'project_id')] in (?)", strutil.ToStrSlice(req.ProjectIDs))
		}
		return tx.Find(&[]FuncPointTrendRow{})
	})
	basicSql := p.DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		tx = tx.Table(fmt.Sprintf("(%s)", dataSql)).
			Select(`orgID,
    userID,
	timestamp,
    sum(responsibleFuncPointsTotal) as responsibleFuncPointsTotal,
    if(responsibleFuncPointsTotal > 0, sum(actualMandayTotal) * 8 / responsibleFuncPointsTotal, 0) as productPDR`).
			Group("orgID").
			Group("userID").
			Group("timestamp").
			Order("timestamp ASC")
		return tx.Find(&[]FuncPointTrendRow{})
	})
	return basicSql
}
