package job

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/strutil"
)

func (j *JobImpl) Inspect(namespace, name string) (apistructs.Job, error) {
	job := apistructs.Job{}
	ctx := context.Background()

	if err := j.js.Get(ctx, makeJobKey(namespace, name), &job); err != nil {
		if strutil.HasSuffixes(err.Error(), NotFoundSuffix) {
			return apistructs.Job{}, err
		}
		return apistructs.Job{}, err
	}

	update, err := j.fetchJobStatus(ctx, &job)
	if err != nil {
		return apistructs.Job{}, err
	}
	if update {
		if err := j.js.Put(ctx, makeJobKey(namespace, name), &job); err != nil {
			return apistructs.Job{}, err
		}
	}
	return job, nil

}
