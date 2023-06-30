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

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
)

func Test_convertRedis(t *testing.T) {
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
	ro := &RedisOperator{
		overcommit: &mockOverCommitUtil{},
	}
	redis := ro.convertRedis(svc, affinity)
	assert.Equal(t, redisExporterImage, redis.Exporter.Image)
	assert.Equal(t, "ignore-warnings ARM64-COW-BUG", redis.CustomConfig[0])
}

type mockOverCommitUtil struct{}

func (m *mockOverCommitUtil) CPUOvercommit(limit float64) float64 {
	return limit
}

func (m *mockOverCommitUtil) MemoryOvercommit(limit int) int {
	return limit
}
