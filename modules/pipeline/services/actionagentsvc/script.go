// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
