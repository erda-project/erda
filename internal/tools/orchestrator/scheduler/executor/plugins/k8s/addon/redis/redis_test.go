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

package redis

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon/mock"
)

func Test_convertRedis(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := apistructs.Service{
		Env: map[string]string{
			"DICE_ORG_ID": "1",
		},
		Image: "redis:6.2.10",
		Resources: apistructs.Resources{
			Cpu:    0.1,
			Mem:    1024,
			MaxCPU: 0.1,
			MaxMem: 1024,
		},
	}
	affinity := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{},
	}

	overCommitUtil := mock.NewMockOverCommitUtil(ctrl)
	overCommitUtil.EXPECT().ResourceOverCommit(svc.Resources).
		Return(corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0.1"),
				corev1.ResourceMemory: resource.MustParse("1024Mi"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0.1"),
				corev1.ResourceMemory: resource.MustParse("1024Mi"),
			},
		}).AnyTimes()

	ro := &RedisOperator{
		overcommit: overCommitUtil,
	}
	redis := ro.convertRedis(svc, affinity)
	assert.Equal(t, redisExporterImage, redis.Exporter.Image)
	assert.Equal(t, "ignore-warnings ARM64-COW-BUG", redis.CustomConfig[0])
}
