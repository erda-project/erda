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
