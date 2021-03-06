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
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/task"
	"github.com/erda-project/erda/pkg/strutil"
)

func (j *JobImpl) Start(namespace, name string, env map[string]string) (apistructs.Job, error) {
	job := apistructs.Job{}
	ctx := context.Background()
	if err := j.js.Get(ctx, makeJobKey(namespace, name), &job); err != nil {
		if strutil.HasSuffixes(err.Error(), NotFoundSuffix) {
			return apistructs.Job{}, err
		}
		return apistructs.Job{}, err
	}

	// update job status
	job.LastModify = time.Now().String()
	job.LastStartTime = time.Now().Unix()
	job.Status = apistructs.StatusUnschedulable
	if err := j.js.Put(ctx, makeJobKey(job.Namespace, job.Name), job); err != nil {
		return apistructs.Job{}, err
	}

	// update job request content(currently only update env)
	if job.Env == nil {
		job.Env = make(map[string]string)
	}
	for k, v := range env {
		job.Env[k] = v
	}

	// build job match tags & exclude tags
	job.Labels = appendJobTags(job.Labels)

	result, err := j.handleJobTask(ctx, &job, task.TaskCreate)
	if err != nil {
		return apistructs.Job{}, err
	}
	// get job id from flink/spark server
	if job.Kind == string(apistructs.Flink) &&
		job.Kind == string(apistructs.Spark) {
		if result.Extra != nil {
			job.ID = result.Extra.(string)
			// update job id to etcd
			if err := j.js.Put(ctx, makeJobKey(job.Namespace, job.Name), job); err != nil {
				return apistructs.Job{}, err
			}
		} else {
			logrus.Errorf("[alert] can't get job id, %+v", job)
		}
	}

	if job, ok := result.Extra.(apistructs.Job); ok {
		return job, nil
	}

	return apistructs.Job{}, nil
}
