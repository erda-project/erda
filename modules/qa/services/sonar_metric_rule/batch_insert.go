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
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/dao"
	"github.com/erda-project/erda/pkg/httpserver"
)

// Create 创建测试集
func (svc *Service) BatchInsert(req *apistructs.SonarMetricRulesBatchInsertRequest) (httpserver.Responser, error) {
	var rules []*dao.QASonarMetricRules
	if req.Metrics == nil || len(req.Metrics) <= 0 {
		return nil, fmt.Errorf("insert metrics is empry")
	}
	for _, metric := range req.Metrics {
		insertRule := &dao.QASonarMetricRules{
			ScopeID:     req.ScopeID,
			ScopeType:   req.ScopeType,
			MetricKeyID: metric.MetricKeyID,
			MetricValue: metric.MetricValue,
			Description: metric.Description,
		}
		if err := checkAndTruncatedMetricValue(insertRule); err != nil {
			return nil, err
		}
		rules = append(rules, insertRule)
	}

	if err := svc.db.BatchInsertSonarMetricRules(rules); err != nil {
		return nil, err
	}

	return httpserver.OkResp(rules)
}

// 校验 key 和 Operational 是否符合规范
func checkAndTruncatedMetricValue(rule *dao.QASonarMetricRules) error {
	metricKey := apistructs.SonarMetricKeys[rule.MetricKeyID]
	if metricKey == nil {
		return fmt.Errorf("not find this metricKey %d", rule.MetricKeyID)
	}

	switch metricKey.ValueType {
	case "WORK_DUR":
		_, err := strconv.Atoi(rule.MetricValue)
		if err != nil {
			return fmt.Errorf("%s type value can not be %s, write value like int", metricKey.MetricKeyDesc, rule.MetricValue)
		}
	case "RATING":
		if rule.MetricValue != "1" && rule.MetricValue != "2" && rule.MetricValue != "3" && rule.MetricValue != "4" {
			return fmt.Errorf("%s type value can not be %s, write value like A", metricKey.MetricKeyDesc, rule.MetricValue)
		}
	case "PERCENT":
		value, err := strconv.ParseFloat(rule.MetricValue, 64)
		if err != nil {
			return fmt.Errorf("%s type value can not be %s, write value like float", metricKey.MetricKeyDesc, rule.MetricValue)
		}
		if metricKey.DecimalScale > 0 {
			rule.MetricValue = fmt.Sprintf("%."+strconv.Itoa(metricKey.DecimalScale)+"f", value)
		}
	case "MILLISEC":
		_, err := strconv.Atoi(rule.MetricValue)
		if err != nil {
			return fmt.Errorf("%s type value can not be %s, write value like int", metricKey.MetricKeyDesc, rule.MetricValue)
		}
	case "INT":
		_, err := strconv.Atoi(rule.MetricValue)
		if err != nil {
			return fmt.Errorf("%s type value can not be %s, write value like int", metricKey.MetricKeyDesc, rule.MetricValue)
		}
	case "FLOAT":
		_, err := strconv.ParseFloat(rule.MetricValue, 64)
		if err != nil {
			return fmt.Errorf("%s type value can not be %s, write value like float", metricKey.MetricKeyDesc, rule.MetricValue)
		}
	}

	return nil
}
