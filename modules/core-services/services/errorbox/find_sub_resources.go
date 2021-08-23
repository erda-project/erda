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
