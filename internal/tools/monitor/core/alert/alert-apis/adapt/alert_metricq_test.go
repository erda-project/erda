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

package adapt

import (
	"context"
	"net/http"
	"net/url"
	"reflect"

	"github.com/golang/mock/gomock"
	"github.com/olivere/elastic"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric/query/chartmeta"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric/query/metricq"
	queryv1 "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/query/v1"
)

// MockQueryer is a mock of Queryer interface.
type MockQueryer struct {
	ctrl     *gomock.Controller
	recorder *MockQueryerMockRecorder
}

// MockQueryerMockRecorder is the mock recorder for MockQueryer.
type MockQueryerMockRecorder struct {
	mock *MockQueryer
}

// NewMockQueryer creates a new mock instance.
func NewMockQueryer(ctrl *gomock.Controller) *MockQueryer {
	mock := &MockQueryer{ctrl: ctrl}
	mock.recorder = &MockQueryerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockQueryer) EXPECT() *MockQueryerMockRecorder {
	return m.recorder
}

// Charts mocks base method.
func (m *MockQueryer) Charts(langCodes i18n.LanguageCodes, typ string) []*chartmeta.ChartMeta {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Charts", langCodes, typ)
	ret0, _ := ret[0].([]*chartmeta.ChartMeta)
	return ret0
}

// Charts indicates an expected call of Charts.
func (mr *MockQueryerMockRecorder) Charts(langCodes, typ interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Charts", reflect.TypeOf((*MockQueryer)(nil).Charts), langCodes, typ)
}

// Client mocks base method.
func (m *MockQueryer) Client() *elastic.Client {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Client")
	ret0, _ := ret[0].(*elastic.Client)
	return ret0
}

// Client indicates an expected call of Client.
func (mr *MockQueryerMockRecorder) Client() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Client", reflect.TypeOf((*MockQueryer)(nil).Client))
}

// GetSingleAggregationMeta mocks base method.
func (m *MockQueryer) GetSingleAggregationMeta(langCodes i18n.LanguageCodes, mode, name string) (*pb.Aggregation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSingleAggregationMeta", langCodes, mode, name)
	ret0, _ := ret[0].(*pb.Aggregation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSingleAggregationMeta indicates an expected call of GetSingleAggregationMeta.
func (mr *MockQueryerMockRecorder) GetSingleAggregationMeta(langCodes, mode, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSingleAggregationMeta", reflect.TypeOf((*MockQueryer)(nil).GetSingleAggregationMeta), langCodes, mode, name)
}

// GetSingleMetricsMeta mocks base method.
func (m *MockQueryer) GetSingleMetricsMeta(langCodes i18n.LanguageCodes, scope, scopeID, metric string) (*pb.MetricMeta, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSingleMetricsMeta", langCodes, scope, scopeID, metric)
	ret0, _ := ret[0].(*pb.MetricMeta)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSingleMetricsMeta indicates an expected call of GetSingleMetricsMeta.
func (mr *MockQueryerMockRecorder) GetSingleMetricsMeta(langCodes, scope, scopeID, metric interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSingleMetricsMeta", reflect.TypeOf((*MockQueryer)(nil).GetSingleMetricsMeta), langCodes, scope, scopeID, metric)
}

// Handle mocks base method.
func (m *MockQueryer) Handle(r *http.Request) interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Handle", r)
	ret0, _ := ret[0].(interface{})
	return ret0
}

// Handle indicates an expected call of Handle.
func (mr *MockQueryerMockRecorder) Handle(r interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Handle", reflect.TypeOf((*MockQueryer)(nil).Handle), r)
}

// HandleV1 mocks base method.
func (m *MockQueryer) HandleV1(r *http.Request, params *metricq.QueryParams) interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HandleV1", r, params)
	ret0, _ := ret[0].(interface{})
	return ret0
}

// HandleV1 indicates an expected call of HandleV1.
func (mr *MockQueryerMockRecorder) HandleV1(r, params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleV1", reflect.TypeOf((*MockQueryer)(nil).HandleV1), r, params)
}

// Indices mocks base method.
func (m *MockQueryer) Indices(metrics, clusters []string, start, end int64) []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Indices", metrics, clusters, start, end)
	ret0, _ := ret[0].([]string)
	return ret0
}

// Indices indicates an expected call of Indices.
func (mr *MockQueryerMockRecorder) Indices(metrics, clusters, start, end interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Indices", reflect.TypeOf((*MockQueryer)(nil).Indices), metrics, clusters, start, end)
}

// MetricGroup mocks base method.
func (m *MockQueryer) MetricGroup(langCodes i18n.LanguageCodes, scope, scopeID, group, mode, format string, appendTags bool) (*pb.MetricGroup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetricGroup", langCodes, scope, scopeID, group, mode, format, appendTags)
	ret0, _ := ret[0].(*pb.MetricGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetricGroup indicates an expected call of MetricGroup.
func (mr *MockQueryerMockRecorder) MetricGroup(langCodes, scope, scopeID, group, mode, format, appendTags interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricGroup", reflect.TypeOf((*MockQueryer)(nil).MetricGroup), langCodes, scope, scopeID, group, mode, format, appendTags)
}

// MetricGroups mocks base method.
func (m *MockQueryer) MetricGroups(langCodes i18n.LanguageCodes, scope, scopeID, mode string) ([]*pb.Group, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetricGroups", langCodes, scope, scopeID, mode)
	ret0, _ := ret[0].([]*pb.Group)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetricGroups indicates an expected call of MetricGroups.
func (mr *MockQueryerMockRecorder) MetricGroups(langCodes, scope, scopeID, mode interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricGroups", reflect.TypeOf((*MockQueryer)(nil).MetricGroups), langCodes, scope, scopeID, mode)
}

// MetricMeta mocks base method.
func (m *MockQueryer) MetricMeta(langCodes i18n.LanguageCodes, scope, scopeID string, _pb ...string) ([]*pb.MetricMeta, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{langCodes, scope, scopeID}
	for _, a := range _pb {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "MetricMeta", varargs...)
	ret0, _ := ret[0].([]*pb.MetricMeta)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetricMeta indicates an expected call of MetricMeta.
func (mr *MockQueryerMockRecorder) MetricMeta(langCodes, scope, scopeID interface{}, pb ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{langCodes, scope, scopeID}, pb...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricMeta", reflect.TypeOf((*MockQueryer)(nil).MetricMeta), varargs...)
}

// MetricNames mocks base method.
func (m *MockQueryer) MetricNames(langCodes i18n.LanguageCodes, scope, scopeID string) ([]*pb.NameDefine, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetricNames", langCodes, scope, scopeID)
	ret0, _ := ret[0].([]*pb.NameDefine)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetricNames indicates an expected call of MetricNames.
func (mr *MockQueryerMockRecorder) MetricNames(langCodes, scope, scopeID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricNames", reflect.TypeOf((*MockQueryer)(nil).MetricNames), langCodes, scope, scopeID)
}

// Query mocks base method.
func (m *MockQueryer) Query(ctx context.Context, tsql, statement string, params map[string]interface{}, options url.Values) (*model.ResultSet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Query", ctx, tsql, statement, params, options)
	ret0, _ := ret[0].(*model.ResultSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query.
func (mr *MockQueryerMockRecorder) Query(ctx, tsql, statement, params, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockQueryer)(nil).Query), ctx, tsql, statement, params, options)
}

// QueryWithFormat mocks base method.
func (m *MockQueryer) QueryWithFormat(ctx context.Context, tsql, statement, format string, langCodes i18n.LanguageCodes, params map[string]interface{}, filters []*model.Filter, options url.Values) (*model.ResultSet, interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryWithFormat", ctx, tsql, statement, format, langCodes, params, filters, options)
	ret0, _ := ret[0].(*model.ResultSet)
	ret1, _ := ret[1].(interface{})
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// QueryWithFormat indicates an expected call of QueryWithFormat.
func (mr *MockQueryerMockRecorder) QueryWithFormat(ctx, tsql, statement, format, langCodes, params, filters, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryWithFormat", reflect.TypeOf((*MockQueryer)(nil).QueryWithFormat), ctx, tsql, statement, format, langCodes, params, filters, options)
}

// QueryWithFormatV1 mocks base method.
func (m *MockQueryer) QueryWithFormatV1(qlang, statement, format string, langCodes i18n.LanguageCodes) (*queryv1.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryWithFormatV1", qlang, statement, format, langCodes)
	ret0, _ := ret[0].(*queryv1.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryWithFormatV1 indicates an expected call of QueryWithFormatV1.
func (mr *MockQueryerMockRecorder) QueryWithFormatV1(qlang, statement, format, langCodes interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryWithFormatV1", reflect.TypeOf((*MockQueryer)(nil).QueryWithFormatV1), qlang, statement, format, langCodes)
}

// RegeistMetricMeta mocks base method.
func (m *MockQueryer) RegeistMetricMeta(scope, scopeID, group string, metrics ...*pb.MetricMeta) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{scope, scopeID, group}
	for _, a := range metrics {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "RegeistMetricMeta", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegeistMetricMeta indicates an expected call of RegeistMetricMeta.
func (mr *MockQueryerMockRecorder) RegeistMetricMeta(scope, scopeID, group interface{}, metrics ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{scope, scopeID, group}, metrics...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegeistMetricMeta", reflect.TypeOf((*MockQueryer)(nil).RegeistMetricMeta), varargs...)
}

// UnregeistMetricMeta mocks base method.
func (m *MockQueryer) UnregeistMetricMeta(scope, scopeID, group string, metrics ...string) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{scope, scopeID, group}
	for _, a := range metrics {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UnregeistMetricMeta", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// UnregeistMetricMeta indicates an expected call of UnregeistMetricMeta.
func (mr *MockQueryerMockRecorder) UnregeistMetricMeta(scope, scopeID, group interface{}, metrics ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{scope, scopeID, group}, metrics...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnregeistMetricMeta", reflect.TypeOf((*MockQueryer)(nil).UnregeistMetricMeta), varargs...)
}
