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

package mysql

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon/mock"
	mysqlv1 "github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon/mysql/v1"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

var sg = &apistructs.ServiceGroup{
	Dice: apistructs.Dice{
		ID: "mock-mysql",
		Labels: map[string]string{
			"USE_OPERATOR": "mysql",
		},
		Services: []apistructs.Service{
			{
				Name: "mysql",
				Volumes: []apistructs.Volume{
					{
						ID:         "mysql-data",
						VolumePath: "/var/lib/mysql",
						SCVolume: apistructs.SCVolume{
							StorageClassName: "ssd",
							Capacity:         1000,
						},
					},
				},
				Resources: apistructs.Resources{
					Cpu: 2,
					Mem: 4096,
				},
				Scale: 2,
				Env: map[string]string{
					"MYSQL_ROOT_PASSWORD": "123",
					"ADDON_ID":            "addonxxx",
					"DICE_CLUSTER_NAME":   "erda",
					"DICE_ORG_NAME":       "erda",
				},
			},
		},
	},
}

var mockResourceRequirements = corev1.ResourceRequirements{
	Limits: corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("2"),
		corev1.ResourceMemory: resource.MustParse("4096Mi"),
	},
	Requests: corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("2048Mi"),
	},
}

func TestMysqlOperator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	namespaceUtil := mock.NewMockNamespaceUtil(ctrl)
	overCommitUtil := mock.NewMockOverCommitUtil(ctrl)
	overCommitUtil.EXPECT().ResourceOverCommit(sg.Services[0].Resources).
		Return(mockResourceRequirements).AnyTimes()

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

func TestConvert(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	overCommitUtil := mock.NewMockOverCommitUtil(ctrl)
	overCommitUtil.EXPECT().ResourceOverCommit(sg.Services[0].Resources).
		Return(mockResourceRequirements).AnyTimes()

	op := &MysqlOperator{overcommit: overCommitUtil}

	got := op.Convert(sg)
	mysql, ok := got.(*mysqlv1.Mysql)
	assert.Equal(t, true, ok)
	assert.Equal(t, "erda", mysql.Spec.Labels["DICE_CLUSTER_NAME"])
}
