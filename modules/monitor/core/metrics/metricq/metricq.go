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

package metricq

import (
	"net/http"

	"github.com/erda-project/erda-infra/providers/i18n"

	"github.com/erda-project/erda/modules/monitor/core/metrics"
	indexmanager "github.com/erda-project/erda/modules/monitor/core/metrics/index"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq/chartmeta"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq/metricmeta"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq/query"
	queryv1 "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/query/v1"
	"github.com/olivere/elastic"
)

// InfluxQL tsql
const InfluxQL = "influxql"

// QueryParams .
type QueryParams struct {
	Scope     string `param:"scope"`
	Aggregate string `param:"aggregate"`
	Format    string `query:"format"`
	Query     string `query:"q"`
	QL        string `query:"ql"`
}

// Queryer .
type Queryer interface {
	query.Queryer

	Client() *elastic.Client
	Indices(metrics []string, clusters []string, start, end int64) []string
	Handle(r *http.Request) interface{}
	MetricNames(langCodes i18n.LanguageCodes, scope, scopeID string) ([]*metrics.NameDefine, error)
	MetricMeta(langCodes i18n.LanguageCodes, scope, scopeID string, metrics ...string) ([]*metrics.MetricMeta, error)
	GetSingleMetricsMeta(langCodes i18n.LanguageCodes, scope, scopeID, metric string) (*metrics.MetricMeta, error)
	GetSingleAggregationMeta(langCodes i18n.LanguageCodes, mode, name string) (*metricmeta.Aggregation, error)
	RegeistMetricMeta(scope, scopeID, group string, metrics ...*metrics.MetricMeta) error
	UnregeistMetricMeta(scope, scopeID, group string, metrics ...string) error
	MetricGroups(langCodes i18n.LanguageCodes, scope, scopeID, mode string) ([]*metricmeta.Group, error)
	MetricGroup(langCodes i18n.LanguageCodes, scope, scopeID, group, mode, format string, appendTags bool) (*metricmeta.GroupDetail, error)

	// v1
	queryv1.Queryer
	HandleV1(r *http.Request, params *QueryParams) interface{}
	Charts(langCodes i18n.LanguageCodes, typ string) []*chartmeta.ChartMeta
}

// Q Queryer .
var Q Queryer

type metricq struct {
	query.Queryer
	index   indexmanager.Index
	meta    *metricmeta.Manager
	handler func(r *http.Request) interface{}

	queryv1   queryv1.Queryer
	charts    *chartmeta.Manager
	handlerV1 func(r *http.Request, params *QueryParams) interface{}
}

// Client .
func (q *metricq) Client() *elastic.Client {
	return q.index.Client()
}

// Indices .
func (q *metricq) Indices(metrics []string, clusters []string, start, end int64) []string {
	return q.index.GetReadIndices(metrics, clusters, start, end)
}

// Handle .
func (q *metricq) Handle(r *http.Request) interface{} {
	return q.handler(r)
}

// MetricMeta .
func (q *metricq) MetricNames(langCodes i18n.LanguageCodes, scope, scopeID string) ([]*metrics.NameDefine, error) {
	return q.meta.MetricNames(langCodes, scope, scopeID)
}

// MetricMeta .
func (q *metricq) MetricMeta(langCodes i18n.LanguageCodes, scope, scopeID string, metrics ...string) ([]*metrics.MetricMeta, error) {
	return q.meta.MetricMeta(langCodes, scope, scopeID, metrics...)
}

func (q *metricq) GetSingleMetricsMeta(langCodes i18n.LanguageCodes, scope, scopeID, metric string) (*metrics.MetricMeta, error) {
	return q.meta.GetSingleMetricsMeta(langCodes, scope, scopeID, metric)
}

func (q *metricq) GetSingleAggregationMeta(langCodes i18n.LanguageCodes, mode, name string) (*metricmeta.Aggregation, error) {
	return q.meta.GetSingleAggregationMeta(langCodes, mode, name)
}

// RegeistMetricMeta .
func (q *metricq) RegeistMetricMeta(scope, scopeID, group string, metrics ...*metrics.MetricMeta) error {
	return q.meta.RegeistMetricMeta(scope, scopeID, group, metrics...)
}

// UnregeistMetricMeta .
func (q *metricq) UnregeistMetricMeta(scope, scopeID, group string, metrics ...string) error {
	return q.meta.UnregeistMetricMeta(scope, scopeID, group, metrics...)
}

// MetricGroups .
func (q *metricq) MetricGroups(langCodes i18n.LanguageCodes, scope, scopeID, mode string) ([]*metricmeta.Group, error) {
	return q.meta.MetricGroups(langCodes, scope, scopeID, mode)
}

// MetricGroup .
func (q *metricq) MetricGroup(langCodes i18n.LanguageCodes, scope, scopeID, group, mode, format string, appendTags bool) (*metricmeta.GroupDetail, error) {
	return q.meta.MetricGroup(langCodes, scope, scopeID, group, mode, format, appendTags)
}

func (q *metricq) QueryWithFormatV1(qlang, statement, format string, langCodes i18n.LanguageCodes) (*queryv1.Response, error) {
	return q.queryv1.QueryWithFormatV1(qlang, statement, format, langCodes)
}

// Charts .
func (q *metricq) Charts(langCodes i18n.LanguageCodes, typ string) []*chartmeta.ChartMeta {
	return q.charts.ChartMetaList(langCodes, typ)
}

// HandleV1 .
func (q *metricq) HandleV1(r *http.Request, params *QueryParams) interface{} {
	return q.handlerV1(r, params)
}
