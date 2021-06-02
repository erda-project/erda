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
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/i18n"
)

func (svc *Service) Paging(req apistructs.SonarMetricRulesPagingRequest, local *i18n.LocaleResource) (httpserver.Responser, error) {
	setDefaultValue(&req)

	paging, err := svc.db.PagingSonarMetricRules(req)
	if err != nil {
		return nil, err
	}

	// 查询系统默认的key
	rule := dao.QASonarMetricRules{
		ScopeID:   "-1",
		ScopeType: "platform",
	}
	defaultDbRules, err := svc.db.ListSonarMetricRules(&rule)
	if err != nil {
		return nil, err
	}

	// 假如数据库没有就设置默认值
	if paging == nil || paging.List == nil {
		paging.List = defaultDbRules
		return httpserver.OkResp(paging)
	}

	// 数据库中有用户的 key, 就在最前面设置上系统默认的 key
	list := paging.List
	dbRules := list.([]dao.QASonarMetricRules)
	defaultDbRules = append(defaultDbRules, dbRules...)

	var rules []*apistructs.SonarMetricRuleDto
	for _, dbRule := range defaultDbRules {
		apiRule := dbRule.ToApi()
		setMetricKeyDesc(apiRule, local)
		rules = append(rules, apiRule)
	}
	paging.List = rules

	return httpserver.OkResp(paging)
}

func setDefaultValue(req *apistructs.SonarMetricRulesPagingRequest) {
	if req.PageNo <= 0 {
		req.PageNo = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 1000
	}
}

func setMetricKeyDesc(apiRule *apistructs.SonarMetricRuleDto, local *i18n.LocaleResource) {
	if apiRule == nil {
		return
	}

	sonarMetricKey := apistructs.SonarMetricKeys[apiRule.MetricKeyID]
	if sonarMetricKey == nil {
		return
	}

	// 国际化
	localDesc := getMetricKeyLocal(apiRule.MetricKey, local)
	if localDesc == "" {
		apiRule.MetricKeyDesc = sonarMetricKey.MetricKeyDesc
	} else {
		apiRule.MetricKeyDesc = localDesc
	}

	apiRule.ValueType = sonarMetricKey.ValueType
	apiRule.DecimalScale = sonarMetricKey.DecimalScale
}

func getMetricKeyLocal(key string, local *i18n.LocaleResource) string {
	if local == nil {
		return ""
	}
	return local.Get(fmt.Sprintf("metric.%s.description", key))
}
