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

package task_uuid

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func MakeJobIDSliceWithLoopedTimes(action *spec.PipelineTask) []string {
	var JobIDSlice []string
	if isLoop(action) {
		for i := apistructs.TaskLoopTimeBegin; i <= int(action.Extra.LoopOptions.LoopedTimes); i++ {
			JobIDSlice = append(JobIDSlice, parseUUID(action.Extra.UUID, i))
		}
		return JobIDSlice
	}
	JobIDSlice = append(JobIDSlice, action.Extra.UUID)
	return JobIDSlice
}

func parseUUID(uuid string, index int) string {
	return fmt.Sprintf("%s-loop-%d", uuid, index)
}

func MakeJobID(action *spec.PipelineTask) string {
	if isLoop(action) {
		return parseUUID(action.Extra.UUID, int(action.Extra.LoopOptions.LoopedTimes))
	}
	return action.Extra.UUID
}

func isLoop(action *spec.PipelineTask) bool {
	return action.Extra.LoopOptions != nil && action.Extra.LoopOptions.CalculatedLoop != nil && action.Extra.LoopOptions.CalculatedLoop.Strategy.MaxTimes > 0
}
