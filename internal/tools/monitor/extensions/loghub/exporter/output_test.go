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

package exporter

import (
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	mocklogger "github.com/erda-project/erda/pkg/mock"
)

// -go:generate mockgen -destination=./mock_log.go -package exporter github.com/erda-project/erda-infra/base/logs Logger
// -go:generate mockgen -destination=./mock_output.go -package exporter -source=./interface.go Output
func TestInvoke_WithLogAnalyticsPattern_Should_Call_Output(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogger.NewMockLogger(ctrl)
	output := NewMockOutput(ctrl)

	output.EXPECT().Write(gomock.Any(), gomock.Any()).Return(nil)

	c := &consumer{
		filters: map[string]string{
			"monitor_log_key": "n4e4d034460114086b2a2b203312f5522",
		},
		log:    logger,
		output: output,
	}

	topic := "topic"
	err := c.Invoke(nil, []byte(`{
		"source": "container",
		"id": "3eb75b2ba0d1560c6148f3023e63c16915e32b077857591dbdb42beca98d997f",
		"stream": "stdout",
		"content": "\u001b[37mDEBU\u001b[0m[2021-08-24 09:50:02.404177939] service: core-services endpoint acquired: core-services.project-387-test.svc.cluster.local:9526 ",
		"offset": 8403051,
		"timestamp": 1629769802404,
		"tags": {
			"container_name": "dop",
			"dice_application_id": "5880",
			"dice_application_name": "erda",
			"dice_cluster_name": "erda-hongkong",
			"dice_org_id": "100060",
			"dice_project_id": "387",
			"dice_project_name": "erda-project",
			"dice_runtime_id": "12496",
			"dice_runtime_name": "develop",
			"dice_service_name": "dop",
			"dice_workspace": "test",
			"module": "erda-project/erda/dop",
			"monitor_log_key": "n4e4d034460114086b2a2b203312f5522",
			"origin": "dice",
			"pod_name": "dop-a525e02f6c-55959d6dbf-7sq8q",
			"pod_namespace": "project-387-test",
			"terminus_key": "t6f7b240844ad47cd8473c30da36ae5dd",
			"terminus_log_key": "n4e4d034460114086b2a2b203312f5522"
		},
		"labels":{
			"container_name": "dop"
		},
		"uniId": "5"
	}`), &topic, time.Now())

	if err != nil {
		t.Errorf("should not throw error")
	}
}

func TestInvoke_WithLogServicePattern_Should_Call_Output(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogger.NewMockLogger(ctrl)
	output := NewMockOutput(ctrl)

	output.EXPECT().Write(gomock.Any(), gomock.Any()).Return(nil)

	c := &consumer{
		filters: map[string]string{
			"msp_env_id":     "n4e4d034460114086b2a2b203312f5522",
			"msp_log_attach": "",
		},
		log:    logger,
		output: output,
	}

	topic := "topic"
	err := c.Invoke(nil, []byte(`{
		"source": "container",
		"id": "3eb75b2ba0d1560c6148f3023e63c16915e32b077857591dbdb42beca98d997f",
		"stream": "stdout",
		"content": "\u001b[37mDEBU\u001b[0m[2021-08-24 09:50:02.404177939] service: core-services endpoint acquired: core-services.project-387-test.svc.cluster.local:9526 ",
		"offset": 8403051,
		"timestamp": 1629769802404,
		"tags": {
			"container_name": "dop",
			"dice_application_id": "5880",
			"dice_application_name": "erda",
			"dice_cluster_name": "erda-hongkong",
			"dice_org_id": "100060",
			"dice_project_id": "387",
			"dice_project_name": "erda-project",
			"dice_runtime_id": "12496",
			"dice_runtime_name": "develop",
			"dice_service_name": "dop",
			"dice_workspace": "test",
			"module": "erda-project/erda/dop",
			"origin": "dice",
			"pod_name": "dop-a525e02f6c-55959d6dbf-7sq8q",
			"pod_namespace": "project-387-test",
			"terminus_key": "t6f7b240844ad47cd8473c30da36ae5dd",
			"msp_env_id": "n4e4d034460114086b2a2b203312f5522",
			"msp_log_attach": "true"
		},
		"labels":{
			"container_name": "dop"
		},
		"uniId": "5"
	}`), &topic, time.Now())

	if err != nil {
		t.Errorf("should not throw error")
	}
}

func TestInvoke_WithNilLabels_Should_Call_Output(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogger.NewMockLogger(ctrl)
	output := NewMockOutput(ctrl)

	output.EXPECT().Write(gomock.Any(), gomock.Any()).Return(nil)

	c := &consumer{
		filters: map[string]string{
			"msp_env_id":     "n4e4d034460114086b2a2b203312f5522",
			"msp_log_attach": "",
		},
		log:    logger,
		output: output,
	}

	topic := "topic"
	err := c.Invoke(nil, []byte(`{
		"source": "container",
		"id": "3eb75b2ba0d1560c6148f3023e63c16915e32b077857591dbdb42beca98d997f",
		"stream": "stdout",
		"content": "\u001b[37mDEBU\u001b[0m[2021-08-24 09:50:02.404177939] service: core-services endpoint acquired: core-services.project-387-test.svc.cluster.local:9526 ",
		"offset": 8403051,
		"timestamp": 1629769802404,
		"tags": {
			"container_name": "dop",
			"dice_application_id": "5880",
			"dice_application_name": "erda",
			"dice_cluster_name": "erda-hongkong",
			"dice_org_id": "100060",
			"dice_project_id": "387",
			"dice_project_name": "erda-project",
			"dice_runtime_id": "12496",
			"dice_runtime_name": "develop",
			"dice_service_name": "dop",
			"dice_workspace": "test",
			"module": "erda-project/erda/dop",
			"origin": "dice",
			"pod_name": "dop-a525e02f6c-55959d6dbf-7sq8q",
			"pod_namespace": "project-387-test",
			"terminus_key": "t6f7b240844ad47cd8473c30da36ae5dd",
			"msp_env_id": "n4e4d034460114086b2a2b203312f5522",
			"msp_log_attach": "true"
		},
		"uniId": "5"
	}`), &topic, time.Now())

	if err != nil {
		t.Errorf("should not throw error")
	}
}

func TestInvoke_WithNoneExistsFilterKey_Should_Not_Call_Output(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogger.NewMockLogger(ctrl)

	c := &consumer{
		filters: map[string]string{
			"msp_env_id":      "n4e4d034460114086b2a2b203312f5522",
			"_not_exist_key_": "",
		},
		log:    logger,
		output: nil,
	}

	topic := "topic"
	err := c.Invoke(nil, []byte(`{
		"source": "container",
		"id": "3eb75b2ba0d1560c6148f3023e63c16915e32b077857591dbdb42beca98d997f",
		"stream": "stdout",
		"content": "\u001b[37mDEBU\u001b[0m[2021-08-24 09:50:02.404177939] service: core-services endpoint acquired: core-services.project-387-test.svc.cluster.local:9526 ",
		"offset": 8403051,
		"timestamp": 1629769802404,
		"tags": {
			"container_name": "dop",
			"dice_application_id": "5880",
			"dice_application_name": "erda",
			"dice_cluster_name": "erda-hongkong",
			"dice_org_id": "100060",
			"dice_project_id": "387",
			"dice_project_name": "erda-project",
			"dice_runtime_id": "12496",
			"dice_runtime_name": "develop",
			"dice_service_name": "dop",
			"dice_workspace": "test",
			"module": "erda-project/erda/dop",
			"origin": "dice",
			"pod_name": "dop-a525e02f6c-55959d6dbf-7sq8q",
			"pod_namespace": "project-387-test",
			"terminus_key": "t6f7b240844ad47cd8473c30da36ae5dd",
			"msp_env_id": "n4e4d034460114086b2a2b203312f5522",
			"msp_log_attach": "true"
		},
		"labels":{
			"container_name": "dop"
		},
		"uniId": "5"
	}`), &topic, time.Now())

	if err != nil {
		t.Errorf("should not throw error")
	}
}

func TestInvoke_WithNoneMatchFilterKey_Should_Not_Call_Output(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogger.NewMockLogger(ctrl)

	c := &consumer{
		filters: map[string]string{
			"msp_env_id":     "_not_exists_",
			"msp_log_attach": "",
		},
		log:    logger,
		output: nil,
	}

	topic := "topic"
	err := c.Invoke(nil, []byte(`{
		"source": "container",
		"id": "3eb75b2ba0d1560c6148f3023e63c16915e32b077857591dbdb42beca98d997f",
		"stream": "stdout",
		"content": "\u001b[37mDEBU\u001b[0m[2021-08-24 09:50:02.404177939] service: core-services endpoint acquired: core-services.project-387-test.svc.cluster.local:9526 ",
		"offset": 8403051,
		"timestamp": 1629769802404,
		"tags": {
			"container_name": "dop",
			"dice_application_id": "5880",
			"dice_application_name": "erda",
			"dice_cluster_name": "erda-hongkong",
			"dice_org_id": "100060",
			"dice_project_id": "387",
			"dice_project_name": "erda-project",
			"dice_runtime_id": "12496",
			"dice_runtime_name": "develop",
			"dice_service_name": "dop",
			"dice_workspace": "test",
			"module": "erda-project/erda/dop",
			"origin": "dice",
			"pod_name": "dop-a525e02f6c-55959d6dbf-7sq8q",
			"pod_namespace": "project-387-test",
			"terminus_key": "t6f7b240844ad47cd8473c30da36ae5dd",
			"msp_env_id": "n4e4d034460114086b2a2b203312f5522",
			"msp_log_attach": "true"
		},
		"labels":{
			"container_name": "dop"
		},
		"uniId": "5"
	}`), &topic, time.Now())

	if err != nil {
		t.Errorf("should not throw error")
	}
}

func TestInvoke_WithNilTags_Should_Not_Call_Output(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogger.NewMockLogger(ctrl)
	logger.EXPECT().Warnf(gomock.Any(), gomock.Any())

	c := &consumer{
		filters: map[string]string{
			"msp_env_id":      "n4e4d034460114086b2a2b203312f5522",
			"_not_exist_key_": "",
		},
		log:    logger,
		output: nil,
	}

	topic := "topic"
	err := c.Invoke(nil, []byte(`{
		"source": "container",
		"id": "3eb75b2ba0d1560c6148f3023e63c16915e32b077857591dbdb42beca98d997f",
		"stream": "stdout",
		"content": "\u001b[37mDEBU\u001b[0m[2021-08-24 09:50:02.404177939] service: core-services endpoint acquired: core-services.project-387-test.svc.cluster.local:9526 ",
		"offset": 8403051,
		"timestamp": 1629769802404,
		"uniId": "5"
	}`), &topic, time.Now())

	if err != nil {
		t.Errorf("should not throw error")
	}
}

// MockOutput is a mock of Output interface.
type MockOutput struct {
	ctrl     *gomock.Controller
	recorder *MockOutputMockRecorder
}

// MockOutputMockRecorder is the mock recorder for MockOutput.
type MockOutputMockRecorder struct {
	mock *MockOutput
}

// NewMockOutput creates a new mock instance.
func NewMockOutput(ctrl *gomock.Controller) *MockOutput {
	mock := &MockOutput{ctrl: ctrl}
	mock.recorder = &MockOutputMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOutput) EXPECT() *MockOutputMockRecorder {
	return m.recorder
}

// Write mocks base method.
func (m *MockOutput) Write(key string, data []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", key, data)
	ret0, _ := ret[0].(error)
	return ret0
}

// Write indicates an expected call of Write.
func (mr *MockOutputMockRecorder) Write(key, data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockOutput)(nil).Write), key, data)
}

// MockInterface is a mock of Interface interface.
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface.
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance.
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// NewConsumer mocks base method.
func (m *MockInterface) NewConsumer(arg0 OutputFactory) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewConsumer", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// NewConsumer indicates an expected call of NewConsumer.
func (mr *MockInterfaceMockRecorder) NewConsumer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewConsumer", reflect.TypeOf((*MockInterface)(nil).NewConsumer), arg0)
}
