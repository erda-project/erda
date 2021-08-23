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

package actionagentsvc

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func RunScript(clusterInfo apistructs.ClusterInfoData, scriptName string, params map[string]string) error {

	// 如果是 edas 集群且指定了打包集群，则去对应集群下载
	jobClusterName := clusterInfo.Get(apistructs.EDASJOB_CLUSTER_NAME)
	if clusterInfo.IsEDAS() && jobClusterName != "" {
		jobClusterInfo, err := bundle.New(bundle.WithScheduler()).QueryClusterInfo(jobClusterName)
		if err != nil {
			return err
		}
		return RunScript(jobClusterInfo, scriptName, params)
	}

	return bundle.New(bundle.WithSoldier(clusterInfo.MustGetPublicURL("soldier"))).RunSoldierScript(scriptName, params)
}
