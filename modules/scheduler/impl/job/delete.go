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
	"github.com/erda-project/erda/modules/scheduler/task"
)

func (j *JobImpl) Delete(job apistructs.Job) error {
	var (
		ok  bool
		err error
	)
	if _, err = j.handleJobTask(context.Background(), &job, task.TaskRemove); err != nil {
		return err
	}
	if ok, err = j.js.Notfound(context.Background(), makeJobKey(job.Namespace, job.Name)); err != nil {
		return err
	}
	if ok {
		if err = j.js.Remove(context.Background(), makeJobKey(job.Namespace, job.Name), &job); err != nil {
			return err
		}
	}
	return nil
}
