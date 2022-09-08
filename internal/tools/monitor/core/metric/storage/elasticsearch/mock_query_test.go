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

package elasticsearch

import (
	"context"
	"reflect"

	"github.com/golang/mock/gomock"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
	es_tsql "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/es-tsql"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric/query/metricmeta"
)

// MockQuery is a mock of Query interface.
type MockQuery struct {
	ctrl     *gomock.Controller
	recorder *MockQueryMockRecorder
}

// MockQueryMockRecorder is the mock recorder for MockQuery.
type MockQueryMockRecorder struct {
	mock *MockQuery
}

// NewMockQuery creates a new mock instance.
func NewMockQuery(ctrl *gomock.Controller) *MockQuery {
	mock := &MockQuery{ctrl: ctrl}
	mock.recorder = &MockQueryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockQuery) EXPECT() *MockQueryMockRecorder {
	return m.recorder
}

// AppendBoolFilter mocks base method.
func (m *MockQuery) AppendBoolFilter(key string, value interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AppendBoolFilter", key, value)
}

// AppendBoolFilter indicates an expected call of AppendBoolFilter.
func (mr *MockQueryMockRecorder) AppendBoolFilter(key, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AppendBoolFilter", reflect.TypeOf((*MockQuery)(nil).AppendBoolFilter), key, value)
}

// Context mocks base method.
func (m *MockQuery) Context() es_tsql.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(es_tsql.Context)
	return ret0
}

// Context indicates an expected call of Context.
func (mr *MockQueryMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockQuery)(nil).Context))
}

// Debug mocks base method.
func (m *MockQuery) Debug() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Debug")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Debug indicates an expected call of Debug.
func (mr *MockQueryMockRecorder) Debug() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockQuery)(nil).Debug))
}

// Kind mocks base method.
func (m *MockQuery) Kind() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Kind")
	ret0, _ := ret[0].(string)
	return ret0
}

// Kind indicates an expected call of Kind.
func (mr *MockQueryMockRecorder) Kind() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Kind", reflect.TypeOf((*MockQuery)(nil).Kind))
}

// OrgName mocks base method.
func (m *MockQuery) OrgName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OrgName")
	ret0, _ := ret[0].(string)
	return ret0
}

// OrgName indicates an expected call of OrgName.
func (mr *MockQueryMockRecorder) OrgName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OrgName", reflect.TypeOf((*MockQuery)(nil).OrgName))
}

// ParseResult mocks base method.
func (m *MockQuery) ParseResult(ctx context.Context, resp interface{}) (*model.Data, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParseResult", ctx, resp)
	ret0, _ := ret[0].(*model.Data)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ParseResult indicates an expected call of ParseResult.
func (mr *MockQueryMockRecorder) ParseResult(ctx, resp interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseResult", reflect.TypeOf((*MockQuery)(nil).ParseResult), ctx, resp)
}

// SearchSource mocks base method.
func (m *MockQuery) SearchSource() interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchSource")
	ret0, _ := ret[0].(interface{})
	return ret0
}

// SearchSource indicates an expected call of SearchSource.
func (mr *MockQueryMockRecorder) SearchSource() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchSource", reflect.TypeOf((*MockQuery)(nil).SearchSource))
}

// Sources mocks base method.
func (m *MockQuery) Sources() []*model.Source {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sources")
	ret0, _ := ret[0].([]*model.Source)
	return ret0
}

// Sources indicates an expected call of Sources.
func (mr *MockQueryMockRecorder) Sources() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sources", reflect.TypeOf((*MockQuery)(nil).Sources))
}

// SubSearchSource mocks base method.
func (m *MockQuery) SubSearchSource() interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubSearchSource")
	ret0, _ := ret[0].(interface{})
	return ret0
}

// SubSearchSource indicates an expected call of SubSearchSource.
func (mr *MockQueryMockRecorder) SubSearchSource() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubSearchSource", reflect.TypeOf((*MockQuery)(nil).SubSearchSource))
}

// TerminusKey mocks base method.
func (m *MockQuery) TerminusKey() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TerminusKey")
	ret0, _ := ret[0].(string)
	return ret0
}

// TerminusKey indicates an expected call of TerminusKey.
func (mr *MockQueryMockRecorder) TerminusKey() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TerminusKey", reflect.TypeOf((*MockQuery)(nil).TerminusKey))
}

// Timestamp mocks base method.
func (m *MockQuery) Timestamp() (int64, int64) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Timestamp")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(int64)
	return ret0, ret1
}

// Timestamp indicates an expected call of Timestamp.
func (mr *MockQueryMockRecorder) Timestamp() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Timestamp", reflect.TypeOf((*MockQuery)(nil).Timestamp))
}

// MockParser is a mock of Parser interface.
type MockParser struct {
	ctrl     *gomock.Controller
	recorder *MockParserMockRecorder
}

// MockParserMockRecorder is the mock recorder for MockParser.
type MockParserMockRecorder struct {
	mock *MockParser
}

// NewMockParser creates a new mock instance.
func NewMockParser(ctrl *gomock.Controller) *MockParser {
	mock := &MockParser{ctrl: ctrl}
	mock.recorder = &MockParserMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockParser) EXPECT() *MockParserMockRecorder {
	return m.recorder
}

// Build mocks base method.
func (m *MockParser) Build() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Build")
	ret0, _ := ret[0].(error)
	return ret0
}

// Build indicates an expected call of Build.
func (mr *MockParserMockRecorder) Build() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Build", reflect.TypeOf((*MockParser)(nil).Build))
}

// GetOrgName mocks base method.
func (m *MockParser) GetOrgName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrgName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetOrgName indicates an expected call of GetOrgName.
func (mr *MockParserMockRecorder) GetOrgName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrgName", reflect.TypeOf((*MockParser)(nil).GetOrgName))
}

// GetTerminusKey mocks base method.
func (m *MockParser) GetTerminusKey() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTerminusKey")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetTerminusKey indicates an expected call of GetTerminusKey.
func (mr *MockParserMockRecorder) GetTerminusKey() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTerminusKey", reflect.TypeOf((*MockParser)(nil).GetTerminusKey))
}

// Metrics mocks base method.
func (m *MockParser) Metrics() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Metrics")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Metrics indicates an expected call of Metrics.
func (mr *MockParserMockRecorder) Metrics() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Metrics", reflect.TypeOf((*MockParser)(nil).Metrics))
}

// ParseQuery mocks base method.
func (m *MockParser) ParseQuery(ctx context.Context, kind string) ([]es_tsql.Query, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParseQuery", ctx, kind)
	ret0, _ := ret[0].([]es_tsql.Query)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ParseQuery indicates an expected call of ParseQuery.
func (mr *MockParserMockRecorder) ParseQuery(ctx, kind interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseQuery", reflect.TypeOf((*MockParser)(nil).ParseQuery), ctx, kind)
}

// SetFilter mocks base method.
func (m *MockParser) SetFilter(filter []*model.Filter) (es_tsql.Parser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetFilter", filter)
	ret0, _ := ret[0].(es_tsql.Parser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SetFilter indicates an expected call of SetFilter.
func (mr *MockParserMockRecorder) SetFilter(filter interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetFilter", reflect.TypeOf((*MockParser)(nil).SetFilter), filter)
}

// SetMaxTimePoints mocks base method.
func (m *MockParser) SetMaxTimePoints(points int64) es_tsql.Parser {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetMaxTimePoints", points)
	ret0, _ := ret[0].(es_tsql.Parser)
	return ret0
}

// SetMaxTimePoints indicates an expected call of SetMaxTimePoints.
func (mr *MockParserMockRecorder) SetMaxTimePoints(points interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetMaxTimePoints", reflect.TypeOf((*MockParser)(nil).SetMaxTimePoints), points)
}

// SetMeta mocks base method.
func (m *MockParser) SetMeta(arg0 metricmeta.MetricMeta) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetMeta", arg0)
}

// SetMeta indicates an expected call of SetMeta.
func (mr *MockParserMockRecorder) SetMeta(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetMeta", reflect.TypeOf((*MockParser)(nil).SetMeta), arg0)
}

// SetOrgName mocks base method.
func (m *MockParser) SetOrgName(org string) es_tsql.Parser {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetOrgName", org)
	ret0, _ := ret[0].(es_tsql.Parser)
	return ret0
}

// SetOrgName indicates an expected call of SetOrgName.
func (mr *MockParserMockRecorder) SetOrgName(org interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetOrgName", reflect.TypeOf((*MockParser)(nil).SetOrgName), org)
}

// SetOriginalTimeUnit mocks base method.
func (m *MockParser) SetOriginalTimeUnit(unit es_tsql.TimeUnit) es_tsql.Parser {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetOriginalTimeUnit", unit)
	ret0, _ := ret[0].(es_tsql.Parser)
	return ret0
}

// SetOriginalTimeUnit indicates an expected call of SetOriginalTimeUnit.
func (mr *MockParserMockRecorder) SetOriginalTimeUnit(unit interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetOriginalTimeUnit", reflect.TypeOf((*MockParser)(nil).SetOriginalTimeUnit), unit)
}

// SetParams mocks base method.
func (m *MockParser) SetParams(params map[string]interface{}) es_tsql.Parser {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetParams", params)
	ret0, _ := ret[0].(es_tsql.Parser)
	return ret0
}

// SetParams indicates an expected call of SetParams.
func (mr *MockParserMockRecorder) SetParams(params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetParams", reflect.TypeOf((*MockParser)(nil).SetParams), params)
}

// SetTargetTimeUnit mocks base method.
func (m *MockParser) SetTargetTimeUnit(unit es_tsql.TimeUnit) es_tsql.Parser {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetTargetTimeUnit", unit)
	ret0, _ := ret[0].(es_tsql.Parser)
	return ret0
}

// SetTargetTimeUnit indicates an expected call of SetTargetTimeUnit.
func (mr *MockParserMockRecorder) SetTargetTimeUnit(unit interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTargetTimeUnit", reflect.TypeOf((*MockParser)(nil).SetTargetTimeUnit), unit)
}

// SetTerminusKey mocks base method.
func (m *MockParser) SetTerminusKey(terminusKey string) es_tsql.Parser {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetTerminusKey", terminusKey)
	ret0, _ := ret[0].(es_tsql.Parser)
	return ret0
}

// SetTerminusKey indicates an expected call of SetTerminusKey.
func (mr *MockParserMockRecorder) SetTerminusKey(terminusKey interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTerminusKey", reflect.TypeOf((*MockParser)(nil).SetTerminusKey), terminusKey)
}

// SetTimeKey mocks base method.
func (m *MockParser) SetTimeKey(key string) es_tsql.Parser {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetTimeKey", key)
	ret0, _ := ret[0].(es_tsql.Parser)
	return ret0
}

// SetTimeKey indicates an expected call of SetTimeKey.
func (mr *MockParserMockRecorder) SetTimeKey(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTimeKey", reflect.TypeOf((*MockParser)(nil).SetTimeKey), key)
}
