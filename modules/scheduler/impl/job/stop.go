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
