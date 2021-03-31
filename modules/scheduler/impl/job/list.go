package job

import (
	"context"

	"github.com/erda-project/erda/apistructs"
)

func (j *JobImpl) List(namespace string) ([]apistructs.Job, error) {
	jobs := make([]apistructs.Job, 0, 100)
	ctx := context.Background()
	err := j.js.ForEach(ctx, makeJobKey(namespace, ""), apistructs.Job{}, func(key string, v interface{}) error {
		job := v.(*apistructs.Job)
		update, err := j.fetchJobStatus(ctx, job)
		if err != nil {
			return err
		}
		if update {
			if err := j.js.Put(ctx, makeJobKey(job.Namespace, job.Name), job); err != nil {
				return err
			}
		}
		jobs = append(jobs, *job)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return jobs, nil
}
