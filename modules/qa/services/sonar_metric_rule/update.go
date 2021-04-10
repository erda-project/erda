// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package sonar_metric_rule

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/httpserver"
)

func (svc *Service) Update(req apistructs.SonarMetricRulesUpdateRequest) (httpserver.Responser, error) {

	dbRule, err := svc.db.GetSonarMetricRules(req.ID)
	if err != nil {
		return nil, err
	}
	if dbRule.ScopeType != req.ScopeType {
		return nil, fmt.Errorf("not find this sonarMetricRule")
	}
	if dbRule.ScopeID != req.ScopeID {
		return nil, fmt.Errorf("not find this sonarMetricRule")
	}
	dbRule.MetricValue = req.MetricValue
	dbRule.Description = req.Description

	if err := checkAndTruncatedMetricValue(dbRule); err != nil {
		return nil, err
	}
	if err := svc.db.UpdateSonarMetricRules(dbRule); err != nil {
		return nil, err
	}

	return httpserver.OkResp(dbRule)
}
