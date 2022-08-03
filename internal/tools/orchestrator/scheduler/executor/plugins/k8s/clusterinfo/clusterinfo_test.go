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

package clusterinfo

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/configmap"
)

func TestParseNetportalURL(t *testing.T) {
	var (
		testURL1 = "inet://127.0.0.1/master.mesos"
		testURL2 = "inet://127.0.0.2?ssl=on&direct=on/master.mesos/service/marathon?test=on"
	)

	netportal, err := parseNetportalURL(testURL1)
	assert.Nil(t, err)
	assert.Equal(t, "inet://127.0.0.1", netportal)

	netportal, err = parseNetportalURL(testURL2)
	assert.Nil(t, err)
	assert.Equal(t, "inet://127.0.0.2?ssl=on&direct=on", netportal)
}

func TestLoad(t *testing.T) {
	cm := configmap.New()
	ci := ClusterInfo{
		ConfigMap: cm,
	}

	defer monkey.UnpatchAll()

	monkey.PatchInstanceMethod(reflect.TypeOf(cm), "Get", func(c *configmap.ConfigMap, namespace string, key string) (*corev1.ConfigMap, error) {
		switch key {
		case apistructs.ConfigMapNameOfClusterInfo:
			return &corev1.ConfigMap{
				Data: map[string]string{
					"DICE_SSH_USER": "fake",
				},
			}, nil
		case apistructs.ConfigMapNameOfAddons:
			return &corev1.ConfigMap{
				Data: map[string]string{
					"REGISTRY_ADDR": "addon-registry.default.svc.cluster.local:5000",
				},
			}, nil
		default:
			return nil, errors.Errorf("configmap %s not found", key)
		}
	})

	assert.NoError(t, ci.Load())
}
