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

package canal

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon/mock"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

var sg = &apistructs.ServiceGroup{
	Dice: apistructs.Dice{
		ID: "mock-canal",
		Labels: map[string]string{
			"USE_OPERATOR": "canal",
		},
		Services: []apistructs.Service{
			{
				Name: "canal",
				Resources: apistructs.Resources{
					Cpu: 3,
					Mem: 3072,
				},
				Scale: 2,
				Env: map[string]string{
					"CANAL_DESTINATION":             "example",
					"canal.instance.master.address": "mock-mysql.svc.cluster.local:3306",
					"canal.instance.dbUsername":     "erda",
					"canal.instance.dbPassword":     "password",
				},
			},
		},
	},
}

var sgCanalAdmin = &apistructs.ServiceGroup{
	Dice: apistructs.Dice{
		ID: "mock-canal",
		Labels: map[string]string{
			"USE_OPERATOR": "canal",
		},
		Services: []apistructs.Service{
			{
				Name: "canal",
				Resources: apistructs.Resources{
					Cpu: 1,
					Mem: 2048,
				},
				Scale: 2,
				Env: map[string]string{
					"canal.admin.manager":        "127.0.0.1:8089",
					"spring.datasource.address":  "mock-mysql.svc.cluster.local:3306",
					"spring.datasource.username": "erda",
					"spring.datasource.password": "",
				},
			},
		},
	},
}

var mockResourceRequirements = corev1.ResourceRequirements{
	Limits: corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("3"),
		corev1.ResourceMemory: resource.MustParse("3072Mi"),
	},
	Requests: corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("3"),
		corev1.ResourceMemory: resource.MustParse("3072Mi"),
	},
}

var mockAdminResourceRequirements = corev1.ResourceRequirements{
	Limits: corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("2048Mi"),
	},
	Requests: corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("2048Mi"),
	},
}

func TestCanalOperator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	namespaceUtil := mock.NewMockNamespaceUtil(ctrl)
	overCommitUtil := mock.NewMockOverCommitUtil(ctrl)
	overCommitUtil.EXPECT().ResourceOverCommit(sg.Services[0].Resources).
		Return(mockResourceRequirements).AnyTimes()
	overCommitUtil.EXPECT().ResourceOverCommit(sgCanalAdmin.Services[0].Resources).
		Return(mockAdminResourceRequirements).AnyTimes()

	k8sUtil := mock.NewMockK8SUtil(ctrl)
	k8sUtil.EXPECT().GetK8SAddr().Return("mock-k8s-addr").AnyTimes()

	mo := New(k8sUtil, namespaceUtil, overCommitUtil, nil, nil, httpclient.New())

	t.Run("Test Name and NamespacedName", func(t *testing.T) {
		assert.NotPanics(t, func() { mo.Name(sg) })
		assert.NotPanics(t, func() { mo.NamespacedName(sg) })
	})

	t.Run("Test IsSupported", func(t *testing.T) {
		assert.NotPanics(t, func() { mo.IsSupported() })
	})

	t.Run("Test Validate", func(t *testing.T) {
		assert.NotPanics(t, func() { mo.Validate(sg) })
	})

	t.Run("Test Convert", func(t *testing.T) {
		assert.NotPanics(t, func() { mo.Convert(sg) })
	})

	t.Run("Test CRUD Operations", func(t *testing.T) {
		assert.NotPanics(t, func() { mo.Create(sg) })
		assert.NotPanics(t, func() { mo.Inspect(sg) })
		assert.NotPanics(t, func() { mo.Update(sg) })
		assert.NotPanics(t, func() { mo.Remove(sg) })
	})
}

func TestCanalOperatorAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	namespaceUtil := mock.NewMockNamespaceUtil(ctrl)
	overCommitUtil := mock.NewMockOverCommitUtil(ctrl)
	overCommitUtil.EXPECT().ResourceOverCommit(sg.Services[0].Resources).
		Return(mockResourceRequirements).AnyTimes()
	overCommitUtil.EXPECT().ResourceOverCommit(sgCanalAdmin.Services[0].Resources).
		Return(mockAdminResourceRequirements).AnyTimes()

	k8sUtil := mock.NewMockK8SUtil(ctrl)
	k8sUtil.EXPECT().GetK8SAddr().Return("mock-k8s-addr").AnyTimes()

	mo := New(k8sUtil, namespaceUtil, overCommitUtil, nil, nil, httpclient.New())

	t.Run("Test Name and NamespacedName", func(t *testing.T) {
		assert.NotPanics(t, func() { mo.Name(sg) })
		assert.NotPanics(t, func() { mo.NamespacedName(sg) })
	})

	t.Run("Test IsSupported", func(t *testing.T) {
		assert.NotPanics(t, func() { mo.IsSupported() })
	})

	t.Run("Test Validate", func(t *testing.T) {
		assert.NotPanics(t, func() { mo.Validate(sg) })
	})

	t.Run("Test Convert", func(t *testing.T) {
		assert.NotPanics(t, func() { mo.Convert(sg) })
	})

	t.Run("Test CRUD Operations", func(t *testing.T) {
		assert.NotPanics(t, func() { mo.Create(sg) })
		assert.NotPanics(t, func() { mo.Inspect(sg) })
		assert.NotPanics(t, func() { mo.Update(sg) })
		assert.NotPanics(t, func() { mo.Remove(sg) })
	})
}
