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

package actionagent

import (
	"os"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/clusterinfo"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
)

var s *provider

func init() {
	// etcd
	js, err := jsonstore.New()
	if err != nil {
		panic(err)
	}
	etcdClient, err := etcd.New()
	if err != nil {
		panic(err)
	}
	// bundle
	os.Setenv("CMDB_ADDR", "cmdb.marathon.l4lb.thisdcos.directory:9093")
	os.Setenv("AGENT_IMAGE_FILE_PATH", "/opt/action/agent")
	bdl := bundle.New(
		bundle.WithCMDB(),
		bundle.WithCMP(),
	)
	s = &provider{accessibleCache: js, etcdctl: etcdClient, bdl: bdl}
}

func TestActionAgentSvc_Ensure(t *testing.T) {
	agentImage := "registry.cn-hangzhou.aliyuncs.com/dice/action-agent:3.4.0-20190715-78211b9c4c"
	agentMD5 := "771821eb0aeab82dc963446a3da381aa"
	var cluster = apistructs.ClusterInfoData{
		"DICE_CLUSTER_TYPE": "k8s",
	}
	err := s.Ensure(cluster, agentImage, agentMD5)
	assert.NoError(t, err)
}

func TestRunScript(t *testing.T) {
	eadaCluster1 := apistructs.ClusterInfoData{
		apistructs.DICE_CLUSTER_TYPE:    "edas",
		apistructs.EDASJOB_CLUSTER_NAME: "cluster2",
	}
	k8sCluster2 := apistructs.ClusterInfoData{
		apistructs.DICE_CLUSTER_TYPE: "edas",
		apistructs.DICE_ROOT_DOMAIN:  "openapi.dice.io",
		apistructs.DICE_PROTOCOL:     "https",
		apistructs.DICE_HTTPS_PORT:   "443",
	}
	pm1 := monkey.Patch(clusterinfo.GetClusterInfoByName, func(name string) (apistructs.ClusterInfo, error) {
		return apistructs.ClusterInfo{CM: k8sCluster2}, nil
	})
	defer pm1.Unpatch()
	bdl := bundle.New()
	pm2 := monkey.Patch(bundle.New, func(options ...bundle.Option) *bundle.Bundle {
		return bdl
	})
	defer pm2.Unpatch()
	pm3 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "RunSoldierScript", func(bdl *bundle.Bundle, scriptName string, params map[string]string) error {
		return nil
	})
	defer pm3.Unpatch()
	err := RunScript(eadaCluster1, "download_file_from_image", map[string]string{})
	assert.NoError(t, err)
}
