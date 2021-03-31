package job

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/task"
	"github.com/erda-project/erda/pkg/strutil"
)

func (j *JobImpl) Stop(namespace, name string) error {
	job := apistructs.Job{}
	ctx := context.Background()
	if err := j.js.Get(ctx, makeJobKey(namespace, name), &job); err != nil {
		if strutil.HasSuffixes(err.Error(), NotFoundSuffix) {
			return err
		}
		return err
	}

	if _, err := j.handleJobTask(ctx, &job, task.TaskDestroy); err != nil {
		return err
	}
	return nil
}
