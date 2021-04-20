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

package pipelinesvc

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/loop"
	"github.com/erda-project/erda/pkg/strutil"
)

// retryQueryClusterInfo query cluster info, retry if tcp error.
func (s *PipelineSvc) retryQueryClusterInfo(clusterName string, pipelineID uint64) (apistructs.ClusterInfoData, error) {
	// no need retry, cluster name is invalid
	if clusterName == "" {
		return apistructs.ClusterInfoData{}, fmt.Errorf("empty cluster name")
	}

	var result apistructs.ClusterInfoData
	var queryErr error

	// 2, 4, 8, 16, 30
	_ = loop.New(loop.WithInterval(time.Second*1), loop.WithDeclineRatio(2), loop.WithDeclineLimit(time.Second*30), loop.WithMaxTimes(5)).
		Do(func() (abort bool, err error) {
			clusterInfo, err := s.bdl.QueryClusterInfo(clusterName)
			if err != nil {
				// need retry if tcp error
				if strutil.Contains(strings.ToLower(err.Error()),
					"dial tcp", "timeout") {
					logrus.Errorf("failed to query cluster info, will retry, clusterName: %s, pipelineID: %d, err: %v", clusterName, pipelineID, err)
					return false, err
				}
				// abnormal error, no need retry
				logrus.Errorf("failed to query cluster info, won't retry, clusterName: %s, pipelineID: %d, err: %v", clusterName, pipelineID, err)
				queryErr = err
				return true, err
			}
			result = clusterInfo
			return true, nil
		})

	return result, queryErr
}
