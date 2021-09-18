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

package logic

import (
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/schedule/executorconfig"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy"
	"github.com/erda-project/erda/pkg/strutil"
)

func GetScheduleInfo(cluster apistructs.ClusterInfo, executorName, executorKind string, jobFromUser apistructs.JobFromUser) (apistructs.ScheduleInfo2, apistructs.ScheduleInfo, error) {
	job := apistructs.Job{
		JobFromUser: jobFromUser,
	}

	enableTag := true
	if cluster.SchedConfig != nil {
		enableTag = cluster.SchedConfig.EnableTag
	}
	// TODO migrate enableWorkspace, cluster-manager should control this param
	enableWorkspace := true
	if cluster.SchedConfig != nil && cluster.SchedConfig.EnableWorkspace != nil {
		enableWorkspace = *cluster.SchedConfig.EnableWorkspace
	}
	configs := executorconfig.ExecutorWholeConfigs{
		BasicConfig: map[string]string{
			"ENABLETAG":        strconv.FormatBool(enableTag),
			"ENABLE_WORKSPACE": strconv.FormatBool(enableWorkspace),
		},
	}
	scheduleInfo2, scheduleInfo, _, err := schedulepolicy.LabelFilterChain(&configs, executorName, strutil.ToUpper(executorKind), job)
	if err != nil {
		return apistructs.ScheduleInfo2{}, apistructs.ScheduleInfo{}, fmt.Errorf("failed to get task scheduleInfo2, err: %v", err)
	}
	return scheduleInfo2, scheduleInfo, nil
}
