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

package errorbox

import (
	"strconv"
)

// FindRuntimeByPipelineID 根据 pipeline Id获取 runtime id
func (eb *ErrorBox) FindRuntimeByPipelineID(pipelineID uint64) ([]string, error) {
	pipelineDetail, err := eb.bdl.GetPipeline(pipelineID)
	if err != nil {
		return nil, err
	}

	var resourceIDs []string
	for _, stage := range pipelineDetail.PipelineStages {
		for _, task := range stage.PipelineTasks {
			for _, metadata := range task.Result.Metadata {
				if metadata.Name == "runtimeID" {
					resourceIDs = append(resourceIDs, metadata.Value)
				}
			}
		}
	}

	return resourceIDs, nil
}

// FindAddonByRuntimeID 根据 runtime ID 获取 addon id
func (eb *ErrorBox) FindAddonByRuntimeID(runtimeID uint64) ([]string, error) {
	addons, err := eb.bdl.ListAddonByRuntimeID(strconv.FormatUint(runtimeID, 10))
	if err != nil {
		return nil, err
	}

	var resourceIDs []string
	for _, addon := range addons.Data {
		resourceIDs = append(resourceIDs, addon.RealInstanceID)
	}

	return resourceIDs, nil
}
