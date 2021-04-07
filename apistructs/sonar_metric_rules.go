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

package apistructs

import (
	"time"
)

const (
	ProjectScopeType = "project"
)

type SonarMetricKey struct {
	ID            int64  `json:"id"`
	MetricKey     string `json:"metricKey"`
	ValueType     string `json:"valueType"`
	Name          string `json:"name"`
	MetricKeyDesc string `json:"metricKeyDesc"`
	Domain        string `json:"domain"`
	Operational   string `json:"operational"`
	Qualitative   bool   `json:"qualitative"`
	Hidden        bool   `json:"hidden"`
	Custom        bool   `json:"custom"`
	DecimalScale  int    `json:"decimalScale"`
	MetricValue   string `json:"metricValue"`
}

func (this *SonarMetricKey) FormatValue() {
	if this == nil {
		return
	}

	this.Operational = GetOperationalValue(this.Operational)
}

func GetOperationalValue(operational string) string {
	if operational == "-1" {
		operational = ">"
	} else if operational == "1" {
		operational = "<"
	}
	return operational
}

var SonarMetricKeys = map[int64]*SonarMetricKey{}

// 分页查询
type SonarMetricRulesPagingRequest struct {
	ScopeType string `json:"scopeType"`
	ScopeID   string `json:"scopeId"`
	PageNo    int    `json:"pageNo"`
	PageSize  int    `json:"pageSize"`
}

// 更新
type SonarMetricRulesUpdateRequest struct {
	ID          int64  `json:"id"`
	Description string `json:"description"`
	MetricValue string `json:"metricValue"`
	ScopeType   string `json:"scopeType"`
	ScopeID     string `json:"scopeId"`
}

// 批量插入
type SonarMetricRulesBatchInsertRequest struct {
	ScopeType string               `json:"scopeType"`
	ScopeID   string               `json:"scopeId"`
	Metrics   []SonarMetricRuleDto `json:"metrics"`
}

type SonarMetricRulesBatchDeleteRequest struct {
	ScopeType string  `json:"scopeType"`
	ScopeID   string  `json:"scopeId"`
	IDs       []int64 `json:"ids"`
}

type SonarMetricRuleDto struct {
	ID            int64     `json:"id"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Description   string    `json:"description"`
	ScopeType     string    `json:"scopeType"`
	ScopeID       string    `json:"scopeId"`
	MetricKey     string    `json:"metricKey"`
	MetricKeyID   int64     `json:"metricKeyId"`
	Operational   string    `json:"operational"`
	MetricValue   string    `json:"metricValue"`
	MetricKeyDesc string    `json:"metricKeyDesc"`
	DecimalScale  int       `json:"decimalScale"`
	ValueType     string    `json:"valueType"`
}

// 删除
type SonarMetricRulesDeleteRequest struct {
	ID        int64  `json:"id"`
	ScopeType string `json:"scopeType"`
	ScopeID   string `json:"scopeId"`
}

//  查询用户未添加 metricKey 和 operational 列表
type SonarMetricRulesDefinitionListRequest struct {
	ScopeType string `json:"scopeType"`
	ScopeID   string `json:"scopeId"`
}

//  查询用户未添加 metricKey 和 operational 列表
type SonarMetricRulesListRequest struct {
	ScopeType string `json:"scopeType"`
	ScopeID   string `json:"scopeId"`
}

type SonarMetricRulesListResp struct {
	Header
	Results []*SonarMetricKey `json:"data"`
}
