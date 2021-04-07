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

package aliyun_resources

import (
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"
)

func GetProjectClusterName(ctx Context, projid string, workspace string) (clusterName, projectName string, err error) {
	// parse project id
	projID, err := strconv.ParseUint(projid, 10, 64)
	if err != nil {
		logrus.Errorf("parse project id failed, project id: %s, error:%v", projid, err)
		return
	}

	// get project info
	proj, err := ctx.Bdl.GetProject(projID)
	if err != nil {
		logrus.Errorf("get project info failed, project id: %s, error:%v", projid, err)
		return
	}
	projectName = proj.Name

	// get cluster name
	clusterName = proj.ClusterConfig[workspace]
	if clusterName == "" && workspace != "" {
		err = fmt.Errorf("get cluster name failed, empty clustr name, project id: %s, workspace:%s", projid, workspace)
		logrus.Errorf("%s, project cluster config:%+v", err.Error(), proj.ClusterConfig)
		return
	}
	return
}
