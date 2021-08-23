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
