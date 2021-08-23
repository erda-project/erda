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

package metricq

import (
	"net/http"

	"github.com/olivere/elastic"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	indexmanager "github.com/erda-project/erda/modules/core/monitor/metric/index"
	"github.com/erda-project/erda/modules/core/monitor/metric/query/chartmeta"
	"github.com/erda-project/erda/modules/core/monitor/metric/query/metricmeta"
	"github.com/erda-project/erda/modules/core/monitor/metric/query/query"
	queryv1 "github.com/erda-project/erda/modules/core/monitor/metric/query/query/v1"
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
	MetricNames(langCodes i18n.LanguageCodes, scope, scopeID string) ([]*pb.NameDefine, error)
	MetricMeta(langCodes i18n.LanguageCodes, scope, scopeID string, pb ...string) ([]*pb.MetricMeta, error)
	GetSingleMetricsMeta(langCodes i18n.LanguageCodes, scope, scopeID, metric string) (*pb.MetricMeta, error)
	GetSingleAggregationMeta(langCodes i18n.LanguageCodes, mode, name string) (*pb.Aggregation, error)
	RegeistMetricMeta(scope, scopeID, group string, metrics ...*pb.MetricMeta) error
	UnregeistMetricMeta(scope, scopeID, group string, metrics ...string) error
	MetricGroups(langCodes i18n.LanguageCodes, scope, scopeID, mode string) ([]*pb.Group, error)
	MetricGroup(langCodes i18n.LanguageCodes, scope, scopeID, group, mode, format string, appendTags bool) (*pb.MetricGroup, error)

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
func (q *metricq) MetricNames(langCodes i18n.LanguageCodes, scope, scopeID string) ([]*pb.NameDefine, error) {
	return q.meta.MetricNames(langCodes, scope, scopeID)
}

// MetricMeta .
func (q *metricq) MetricMeta(langCodes i18n.LanguageCodes, scope, scopeID string, metrics ...string) ([]*pb.MetricMeta, error) {
	return q.meta.MetricMeta(langCodes, scope, scopeID, metrics...)
}

func (q *metricq) GetSingleMetricsMeta(langCodes i18n.LanguageCodes, scope, scopeID, metric string) (*pb.MetricMeta, error) {
	return q.meta.GetSingleMetricsMeta(langCodes, scope, scopeID, metric)
}

func (q *metricq) GetSingleAggregationMeta(langCodes i18n.LanguageCodes, mode, name string) (*pb.Aggregation, error) {
	return q.meta.GetSingleAggregationMeta(langCodes, mode, name)
}

// RegeistMetricMeta .
func (q *metricq) RegeistMetricMeta(scope, scopeID, group string, metrics ...*pb.MetricMeta) error {
	return q.meta.RegeistMetricMeta(scope, scopeID, group, metrics...)
}

// UnregeistMetricMeta .
func (q *metricq) UnregeistMetricMeta(scope, scopeID, group string, metrics ...string) error {
	return q.meta.UnregeistMetricMeta(scope, scopeID, group, metrics...)
}

// MetricGroups .
func (q *metricq) MetricGroups(langCodes i18n.LanguageCodes, scope, scopeID, mode string) ([]*pb.Group, error) {
	return q.meta.MetricGroups(langCodes, scope, scopeID, mode)
}

// MetricGroup .
func (q *metricq) MetricGroup(langCodes i18n.LanguageCodes, scope, scopeID, group, mode, format string, appendTags bool) (*pb.MetricGroup, error) {
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
