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
	"github.com/erda-project/erda/modules/qa/dao"
	"github.com/erda-project/erda/pkg/httpserver"
)

func (svc *Service) BatchDelete(req *apistructs.SonarMetricRulesBatchDeleteRequest) (httpserver.Responser, error) {
	var rules []dao.QASonarMetricRules
	if req.IDs == nil || len(req.IDs) <= 0 {
		return nil, fmt.Errorf("batch delete id list is empry")
	}

	for _, v := range req.IDs {
		rules = append(rules, dao.QASonarMetricRules{
			ID:        v,
			ScopeType: req.ScopeType,
			ScopeID:   req.ScopeID,
		})
	}

	if err := svc.db.BatchDeleteSonarMetricRules(rules); err != nil {
		return nil, err
	}

	return httpserver.OkResp("success")
}
