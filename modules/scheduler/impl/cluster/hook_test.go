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

package cluster

import (
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/impl/cluster/clusterutil"
	"github.com/erda-project/erda/pkg/jsonstore"
)

func TestSetDefaultClusterConfig(t *testing.T) {
	c1 := ClusterInfo{
		ClusterName: "x1",
		Kind:        "MARATHON",
		Options: map[string]string{
			"ADDR":      "http://dcos.x1.cn/service/marathon",
			"ENABLETAG": "true",
		},
	}

	setDefaultClusterConfig(&c1)
	assert.Equal(t, "true", c1.Options["ENABLETAG"])
	assert.Equal(t, "true", c1.Options["ENABLE_ORG"])
	assert.Equal(t, "true", c1.Options["ENABLE_WORKSPACE"])

	c2 := ClusterInfo{
		ClusterName: "x2",
		Kind:        "MARATHON",
		Options: map[string]string{
			"ADDR": "http://dcos.x2.cn/service/marathon",
		},
	}

	setDefaultClusterConfig(&c2)
	assert.Equal(t, "true", c2.Options["ENABLETAG"])
	assert.Equal(t, "true", c2.Options["ENABLE_ORG"])
	assert.Equal(t, "true", c2.Options["ENABLE_WORKSPACE"])

	c3 := ClusterInfo{
		ClusterName: "x3",
		Kind:        "METRONOME",
		Options: map[string]string{
			"ADDR": "http://dcos.x3.cn/service/metronome",
		},
	}

	setDefaultClusterConfig(&c3)
	assert.Equal(t, "true", c3.Options["ENABLETAG"])
	assert.Equal(t, "true", c3.Options["ENABLE_ORG"])
	assert.Equal(t, "true", c3.Options["ENABLE_WORKSPACE"])
}

func TestCreateEdasExecutor(t *testing.T) {
	cl := ClusterImpl{
		js: nil,
	}
	ce := apistructs.ClusterEvent{
		Content: apistructs.ClusterInfo{
			ID:             0,
			Name:           "fake-cluster",
			DisplayName:    "fake-cluster",
			Type:           "fake-cluster",
			CloudVendor:    "fake-cluster",
			Logo:           "fake-cluster",
			Description:    "fake-cluster",
			WildcardDomain: "fake-cluster",
			SchedConfig: &apistructs.ClusterSchedConfig{
				MasterURL:                "fake-cluster",
				AuthType:                 "fake-cluster",
				AuthUsername:             "fake-cluster",
				AuthPassword:             "fake-cluster",
				CACrt:                    "fake-cluster",
				ClientCrt:                "fake-cluster",
				ClientKey:                "fake-cluster",
				EnableTag:                false,
				EdasConsoleAddr:          "fake-cluster",
				AccessKey:                "fake-cluster",
				AccessSecret:             "fake-cluster",
				ClusterID:                "fake-cluster",
				RegionID:                 "fake-cluster",
				LogicalRegionID:          "fake-cluster",
				K8sAddr:                  "fake-cluster",
				RegAddr:                  "fake-cluster:5000",
				CPUSubscribeRatio:        "fake-cluster",
				DevCPUSubscribeRatio:     "fake-cluster",
				TestCPUSubscribeRatio:    "fake-cluster",
				StagingCPUSubscribeRatio: "fake-cluster",
			},
			OpsConfig:    nil,
			System:       nil,
			ManageConfig: nil,
		},
	}
	generateExecutorByClusterPatch := monkey.Patch(clusterutil.GenerateExecutorByCluster, func(cluster, executorType string) string {
		return "fake-cluster"
	})
	createEdasExectorPatch := monkey.Patch(createEdasExector, func(js jsonstore.JsonStore, key string, c string) error {
		return nil
	})
	createPatch := monkey.Patch(create, func(js jsonstore.JsonStore, key string, c ClusterInfo) error {
		return nil
	})
	defer createEdasExectorPatch.Unpatch()
	defer generateExecutorByClusterPatch.Unpatch()
	defer createPatch.Unpatch()

	err := cl.createEdasExecutor(&ce)
	assert.Equal(t, err, nil)
}
