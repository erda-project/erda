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
