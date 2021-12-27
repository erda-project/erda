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

package taskerror

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

type TaskError struct {
	bdl *bundle.Bundle
}

type Option func(*TaskError)

func New(options ...Option) *TaskError {
	eb := &TaskError{}
	for _, op := range options {
		op(eb)
	}
	return eb
}

func WithBundle(bdl *bundle.Bundle) Option {
	return func(eb *TaskError) {
		eb.bdl = bdl
	}
}

func (te *TaskError) List(param *apistructs.ErrorLogListRequest) ([]apistructs.ErrorLog, error) {
	resourceIDs, resourceTypes, err := te.aggregateResources(param.ResourceType, param.ResourceID)
	if err != nil {
		return nil, err
	}
	errLogsReq := apistructs.TaskErrorListRequest{
		ResourceIDS:   resourceIDs,
		ResourceTypes: resourceTypes,
		StartTime:     param.StartTime,
	}

	errLogs, err := te.bdl.ListErrorLog(&errLogsReq)
	if err != nil {
		return nil, err
	}
	return errLogs.List, nil
}

func (te *TaskError) aggregateResources(resourceType apistructs.ErrorResourceType, resourceID string) ([]string,
	[]apistructs.ErrorResourceType, error) {
	resourceTypes, resourceIDs := []apistructs.ErrorResourceType{resourceType}, []string{resourceID}

	switch resourceType {
	case apistructs.PipelineError:
		pipelineID, err := strconv.ParseUint(resourceID, 10, 64)
		if err != nil {
			return nil, nil, err
		}
		runtimeIDs, err := te.FindRuntimeByPipelineID(pipelineID)
		if err != nil {
			return nil, nil, err
		}
		if len(runtimeIDs) != 0 {
			resourceIDs = append(resourceIDs, runtimeIDs...)
			resourceTypes = append(resourceTypes, apistructs.RuntimeError)
		}

		for _, v := range runtimeIDs {
			runtimeID, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return nil, nil, err
			}
			addonIDs, err := te.FindAddonByRuntimeID(runtimeID)
			if err != nil {
				return nil, nil, err
			}
			if len(addonIDs) != 0 {
				resourceIDs = append(resourceIDs, addonIDs...)
				resourceTypes = append(resourceTypes, apistructs.AddonError)
			}
		}
	case apistructs.RuntimeError:
		runtimeID, err := strconv.ParseUint(resourceID, 10, 64)
		if err != nil {
			return nil, nil, err
		}
		addonIDs, err := te.FindAddonByRuntimeID(runtimeID)
		if err != nil {
			return nil, nil, err
		}
		if len(addonIDs) != 0 {
			resourceIDs = append(resourceIDs, addonIDs...)
			resourceTypes = append(resourceTypes, apistructs.AddonError)
		}
		// case apistructs.AddonError:
	}

	return resourceIDs, resourceTypes, nil
}

// FindRuntimeByPipelineID 根据 pipeline Id获取 runtime id
func (te *TaskError) FindRuntimeByPipelineID(pipelineID uint64) ([]string, error) {
	pipelineDetail, err := te.bdl.GetPipeline(pipelineID)
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
func (te *TaskError) FindAddonByRuntimeID(runtimeID uint64) ([]string, error) {
	addons, err := te.bdl.ListAddonByRuntimeID(strconv.FormatUint(runtimeID, 10))
	if err != nil {
		return nil, err
	}

	var resourceIDs []string
	for _, addon := range addons.Data {
		resourceIDs = append(resourceIDs, addon.RealInstanceID)
	}

	return resourceIDs, nil
}
