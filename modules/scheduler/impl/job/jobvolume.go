package job

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/task"
)

func (j *JobImpl) CreateJobVolume(req apistructs.JobVolume) (string, error) {
	ctx := context.Background()

	result, err := j.handleJobVolumeTask(ctx, &req, task.TaskJobVolumeCreate)
	if err != nil {
		return "", err
	}
	id := result.Extra.(string)
	return id, nil
}
