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

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/task"
	"github.com/erda-project/erda/pkg/strutil"
)

func (j *JobImpl) Delete(job apistructs.Job) error {
	var (
		ok  bool
		err error
	)
	if err := j.js.Get(context.Background(), makeJobKey(job.Namespace, job.Name), &job); err != nil {
		if strutil.HasSuffixes(err.Error(), NotFoundSuffix) {
			logrus.Warnf("job %s in %s not found", job.Name, job.Namespace)
			return nil
		}
		return err
	}
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
