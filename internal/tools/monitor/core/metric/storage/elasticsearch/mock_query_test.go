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
	"reflect"

	"github.com/golang/mock/gomock"
	"github.com/olivere/elastic"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
	tsql "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/es-tsql"
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
func (m *MockQuery) AppendBoolFilter(arg0 string, arg1 interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AppendBoolFilter", arg0, arg1)
}

// AppendBoolFilter indicates an expected call of AppendBoolFilter.
func (mr *MockQueryMockRecorder) AppendBoolFilter(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AppendBoolFilter", reflect.TypeOf((*MockQuery)(nil).AppendBoolFilter), arg0, arg1)
}

// Context mocks base method.
func (m *MockQuery) Context() tsql.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(tsql.Context)
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

// ParseResult mocks base method.
func (m *MockQuery) ParseResult(arg0 *elastic.SearchResult) (*model.Data, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParseResult", arg0)
	ret0, _ := ret[0].(*model.Data)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ParseResult indicates an expected call of ParseResult.
func (mr *MockQueryMockRecorder) ParseResult(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseResult", reflect.TypeOf((*MockQuery)(nil).ParseResult), arg0)
}

// SearchSource mocks base method.
func (m *MockQuery) SearchSource() *elastic.SearchSource {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchSource")
	ret0, _ := ret[0].(*elastic.SearchSource)
	return ret0
}

// SearchSource indicates an expected call of SearchSource.
func (mr *MockQueryMockRecorder) SearchSource() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchSource", reflect.TypeOf((*MockQuery)(nil).SearchSource))
}

// SetAllColumnsCallback mocks base method.
func (m *MockQuery) SetAllColumnsCallback(arg0 func(int64, int64, []*model.Source) ([]*model.Column, error)) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetAllColumnsCallback", arg0)
}

// SetAllColumnsCallback indicates an expected call of SetAllColumnsCallback.
func (mr *MockQueryMockRecorder) SetAllColumnsCallback(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAllColumnsCallback", reflect.TypeOf((*MockQuery)(nil).SetAllColumnsCallback), arg0)
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
