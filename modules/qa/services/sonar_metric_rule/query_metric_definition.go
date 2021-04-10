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
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/dao"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/i18n"
)

func (svc *Service) QueryMetricDefinition(req *apistructs.SonarMetricRulesDefinitionListRequest, local *i18n.LocaleResource) (httpserver.Responser, error) {

	rule := dao.QASonarMetricRules{
		ScopeID:   req.ScopeID,
		ScopeType: req.ScopeType,
	}

	// 获取已经新增的 key
	var metricKeysQuery = func(engine *gorm.DB) *gorm.DB {
		var searchMetricKey []int64
		for k := range apistructs.SonarMetricKeys {
			searchMetricKey = append(searchMetricKey, k)
		}
		engine.Where("metric_key_id in (?)", searchMetricKey)
		return engine
	}
	dbRules, err := svc.db.ListSonarMetricRules(&rule, metricKeysQuery)
	if err != nil {
		return nil, err
	}

	// 过滤出用户没有新增的 key
	var sonarMetricKeys []*apistructs.SonarMetricKey
	for key := range apistructs.SonarMetricKeys {
		var find = false
		for _, dbRule := range dbRules {
			if dbRule.MetricKeyID == key {
				find = true
				break
			}
		}
		if !find {
			sonarMetricKeys = append(sonarMetricKeys, apistructs.SonarMetricKeys[key])
		}
	}

	for _, v := range sonarMetricKeys {
		v.FormatValue()

		// 国际化
		localDesc := getMetricKeyLocal(v.MetricKey, local)
		if localDesc != "" {
			v.MetricKeyDesc = localDesc
		}

	}

	return httpserver.OkResp(sonarMetricKeys)
}
