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
	"github.com/erda-project/erda/pkg/httpserver"
)

func (svc *Service) Get(ID int64) (httpserver.Responser, error) {

	dbRule, err := svc.db.GetSonarMetricRules(ID)
	if err != nil {
		return nil, err
	}

	apiRule := dbRule.ToApi()
	setMetricKeyDesc(apiRule, nil)

	return httpserver.OkResp(apiRule)
}
