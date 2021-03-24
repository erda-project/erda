package statusutil

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

// CalculatePipelineStatusV2
func CalculatePipelineStatusV2(tasks []*spec.PipelineTask) apistructs.PipelineStatus {
	total := len(tasks)
	var successNum int
	var failedNum int
	var pauseNum int
	var bornNum int
	var runningNum int
	var analyzedNum int

	for _, task := range tasks {
		if task.Status.IsSuccessStatus() ||
			(task.Status.IsFailedStatus() && task.Extra.AllowFailure) ||
			task.Status == apistructs.PipelineStatusDisabled {
			successNum++
			continue
		}
		if task.Status.IsFailedStatus() {
			failedNum++
			continue
		}
		if task.Status == apistructs.PipelineStatusAnalyzed {
			analyzedNum++
			continue
		}
		if task.Status == apistructs.PipelineStatusBorn {
			bornNum++
			continue
		}
		if task.Status == apistructs.PipelineStatusPaused {
			pauseNum++
			continue
		}
		if task.Status == apistructs.PipelineStatusDisabled {
			successNum++
			continue
		}
		runningNum++
	}

	switch total {
	case 0:
		return apistructs.PipelineStatusSuccess
	case analyzedNum:
		return apistructs.PipelineStatusAnalyzed
	case bornNum + analyzedNum:
		return apistructs.PipelineStatusRunning
	case successNum:
		return apistructs.PipelineStatusSuccess
	case successNum + pauseNum + bornNum + analyzedNum:
		if pauseNum > 0 {
			return apistructs.PipelineStatusPaused
		}
		return apistructs.PipelineStatusRunning
	case successNum + failedNum + bornNum + analyzedNum:
		if failedNum == 0 {
			return apistructs.PipelineStatusRunning
		}
		return apistructs.PipelineStatusFailed
	default:
		return apistructs.PipelineStatusRunning
	}
}

// CalculatePipelineTaskAllDone 计算 pipeline 下的所有任务是否都已完毕
func CalculatePipelineTaskAllDone(tasks []*spec.PipelineTask) bool {
	return CalculatePipelineStatusV2(tasks).IsEndStatus()
}
