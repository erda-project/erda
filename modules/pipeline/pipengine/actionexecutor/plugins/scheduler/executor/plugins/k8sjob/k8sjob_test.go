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

package k8sjob

import (
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/logic"
	"github.com/erda-project/erda/pkg/k8sclient"
)

func Test_isBuildkitHit(t *testing.T) {
	utCount := 10000

	tests := []struct {
		Rate   int
		Result int
	}{
		{
			Rate: 0,
		},
		{
			Rate: 10,
		},
		{
			Rate: 50,
		},
		{
			Rate: 80,
		},
		{
			Rate: 100,
		},
	}

	for _, test := range tests {
		for i := 0; i < utCount; i++ {
			if isBuildkitHit(test.Rate) {
				test.Result++
			}
		}
		t.Logf("rate: %d, result: %v", test.Rate, test.Result*100/utCount)
	}
}

func Test_generateKubeJob(t *testing.T) {
	defer monkey.UnpatchAll()

	monkey.Patch(k8sclient.New, func(_ string) (*k8sclient.K8sClient, error) {
		return nil, nil
	})

	monkey.Patch(logic.GetCLusterInfo, func(clusterName string) (map[string]string, error) {
		return map[string]string{
			"BUILDKIT_ENABLE": "false",
		}, nil
	})

	j, err := New("fake-job", "fake-cluster", apistructs.ClusterInfo{})
	assert.NoError(t, err)
	assert.Equal(t, string(j.Name()), "fake-job")

	_, err = j.generateKubeJob(apistructs.JobFromUser{
		Name:      "fake-job",
		Namespace: metav1.NamespaceDefault,
	})
	assert.NoError(t, err)
}
