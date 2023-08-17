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
func (m *MockQueryer) Charts(arg0 i18n.LanguageCodes, arg1 string) []*chartmeta.ChartMeta {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Charts", arg0, arg1)
	ret0, _ := ret[0].([]*chartmeta.ChartMeta)
	return ret0
}

// Charts indicates an expected call of Charts.
func (mr *MockQueryerMockRecorder) Charts(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Charts", reflect.TypeOf((*MockQueryer)(nil).Charts), arg0, arg1)
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

// ExternalHandle mocks base method.
func (m *MockQueryer) ExternalHandle(arg0 *http.Request) interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExternalHandle", arg0)
	ret0, _ := ret[0].(interface{})
	return ret0
}

// ExternalHandle indicates an expected call of ExternalHandle.
func (mr *MockQueryerMockRecorder) ExternalHandle(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExternalHandle", reflect.TypeOf((*MockQueryer)(nil).ExternalHandle), arg0)
}

// GetSingleAggregationMeta mocks base method.
func (m *MockQueryer) GetSingleAggregationMeta(arg0 i18n.LanguageCodes, arg1, arg2 string) (*pb.Aggregation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSingleAggregationMeta", arg0, arg1, arg2)
	ret0, _ := ret[0].(*pb.Aggregation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSingleAggregationMeta indicates an expected call of GetSingleAggregationMeta.
func (mr *MockQueryerMockRecorder) GetSingleAggregationMeta(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSingleAggregationMeta", reflect.TypeOf((*MockQueryer)(nil).GetSingleAggregationMeta), arg0, arg1, arg2)
}

// GetSingleMetricsMeta mocks base method.
func (m *MockQueryer) GetSingleMetricsMeta(arg0 i18n.LanguageCodes, arg1, arg2, arg3 string) (*pb.MetricMeta, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSingleMetricsMeta", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*pb.MetricMeta)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSingleMetricsMeta indicates an expected call of GetSingleMetricsMeta.
func (mr *MockQueryerMockRecorder) GetSingleMetricsMeta(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSingleMetricsMeta", reflect.TypeOf((*MockQueryer)(nil).GetSingleMetricsMeta), arg0, arg1, arg2, arg3)
}

// Handle mocks base method.
func (m *MockQueryer) Handle(arg0 *http.Request) interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Handle", arg0)
	ret0, _ := ret[0].(interface{})
	return ret0
}

// Handle indicates an expected call of Handle.
func (mr *MockQueryerMockRecorder) Handle(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Handle", reflect.TypeOf((*MockQueryer)(nil).Handle), arg0)
}

// HandleV1 mocks base method.
func (m *MockQueryer) HandleV1(arg0 *http.Request, arg1 *metricq.QueryParams) interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HandleV1", arg0, arg1)
	ret0, _ := ret[0].(interface{})
	return ret0
}

// HandleV1 indicates an expected call of HandleV1.
func (mr *MockQueryerMockRecorder) HandleV1(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleV1", reflect.TypeOf((*MockQueryer)(nil).HandleV1), arg0, arg1)
}

// Indices mocks base method.
func (m *MockQueryer) Indices(arg0, arg1 []string, arg2, arg3 int64) []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Indices", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]string)
	return ret0
}

// Indices indicates an expected call of Indices.
func (mr *MockQueryerMockRecorder) Indices(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Indices", reflect.TypeOf((*MockQueryer)(nil).Indices), arg0, arg1, arg2, arg3)
}

// MetricGroup mocks base method.
func (m *MockQueryer) MetricGroup(arg0 i18n.LanguageCodes, arg1, arg2, arg3, arg4, arg5 string, arg6 bool) (*pb.MetricGroup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetricGroup", arg0, arg1, arg2, arg3, arg4, arg5, arg6)
	ret0, _ := ret[0].(*pb.MetricGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetricGroup indicates an expected call of MetricGroup.
func (mr *MockQueryerMockRecorder) MetricGroup(arg0, arg1, arg2, arg3, arg4, arg5, arg6 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricGroup", reflect.TypeOf((*MockQueryer)(nil).MetricGroup), arg0, arg1, arg2, arg3, arg4, arg5, arg6)
}

// MetricGroups mocks base method.
func (m *MockQueryer) MetricGroups(arg0 i18n.LanguageCodes, arg1, arg2, arg3 string) ([]*pb.Group, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetricGroups", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]*pb.Group)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetricGroups indicates an expected call of MetricGroups.
func (mr *MockQueryerMockRecorder) MetricGroups(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricGroups", reflect.TypeOf((*MockQueryer)(nil).MetricGroups), arg0, arg1, arg2, arg3)
}

// MetricMeta mocks base method.
func (m *MockQueryer) MetricMeta(arg0 i18n.LanguageCodes, arg1, arg2 string, arg3 ...string) ([]*pb.MetricMeta, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "MetricMeta", varargs...)
	ret0, _ := ret[0].([]*pb.MetricMeta)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetricMeta indicates an expected call of MetricMeta.
func (mr *MockQueryerMockRecorder) MetricMeta(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricMeta", reflect.TypeOf((*MockQueryer)(nil).MetricMeta), varargs...)
}

// MetricNames mocks base method.
func (m *MockQueryer) MetricNames(arg0 i18n.LanguageCodes, arg1, arg2 string) ([]*pb.NameDefine, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetricNames", arg0, arg1, arg2)
	ret0, _ := ret[0].([]*pb.NameDefine)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetricNames indicates an expected call of MetricNames.
func (mr *MockQueryerMockRecorder) MetricNames(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricNames", reflect.TypeOf((*MockQueryer)(nil).MetricNames), arg0, arg1, arg2)
}

// Query mocks base method.
func (m *MockQueryer) Query(arg0 context.Context, arg1, arg2 string, arg3 map[string]interface{}, arg4 url.Values) (*model.ResultSet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Query", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(*model.ResultSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query.
func (mr *MockQueryerMockRecorder) Query(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockQueryer)(nil).Query), arg0, arg1, arg2, arg3, arg4)
}

// QueryExternalWithFormat mocks base method.
func (m *MockQueryer) QueryExternalWithFormat(arg0 context.Context, arg1, arg2, arg3 string, arg4 i18n.LanguageCodes, arg5 map[string]interface{}, arg6 []*model.Filter, arg7 url.Values) (*model.ResultSet, interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryExternalWithFormat", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	ret0, _ := ret[0].(*model.ResultSet)
	ret1, _ := ret[1].(interface{})
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// QueryExternalWithFormat indicates an expected call of QueryExternalWithFormat.
func (mr *MockQueryerMockRecorder) QueryExternalWithFormat(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryExternalWithFormat", reflect.TypeOf((*MockQueryer)(nil).QueryExternalWithFormat), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
}

// QueryWithFormat mocks base method.
func (m *MockQueryer) QueryWithFormat(arg0 context.Context, arg1, arg2, arg3 string, arg4 i18n.LanguageCodes, arg5 map[string]interface{}, arg6 []*model.Filter, arg7 url.Values) (*model.ResultSet, interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryWithFormat", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	ret0, _ := ret[0].(*model.ResultSet)
	ret1, _ := ret[1].(interface{})
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// QueryWithFormat indicates an expected call of QueryWithFormat.
func (mr *MockQueryerMockRecorder) QueryWithFormat(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryWithFormat", reflect.TypeOf((*MockQueryer)(nil).QueryWithFormat), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
}

// QueryWithFormatV1 mocks base method.
func (m *MockQueryer) QueryWithFormatV1(arg0, arg1, arg2 string, arg3 i18n.LanguageCodes) (*queryv1.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryWithFormatV1", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*queryv1.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryWithFormatV1 indicates an expected call of QueryWithFormatV1.
func (mr *MockQueryerMockRecorder) QueryWithFormatV1(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryWithFormatV1", reflect.TypeOf((*MockQueryer)(nil).QueryWithFormatV1), arg0, arg1, arg2, arg3)
}

// RegeistMetricMeta mocks base method.
func (m *MockQueryer) RegeistMetricMeta(arg0, arg1, arg2 string, arg3 ...*pb.MetricMeta) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "RegeistMetricMeta", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegeistMetricMeta indicates an expected call of RegeistMetricMeta.
func (mr *MockQueryerMockRecorder) RegeistMetricMeta(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegeistMetricMeta", reflect.TypeOf((*MockQueryer)(nil).RegeistMetricMeta), varargs...)
}

// UnregeistMetricMeta mocks base method.
func (m *MockQueryer) UnregeistMetricMeta(arg0, arg1, arg2 string, arg3 ...string) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UnregeistMetricMeta", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// UnregeistMetricMeta indicates an expected call of UnregeistMetricMeta.
func (mr *MockQueryerMockRecorder) UnregeistMetricMeta(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnregeistMetricMeta", reflect.TypeOf((*MockQueryer)(nil).UnregeistMetricMeta), varargs...)
}
