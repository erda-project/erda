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
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/task"
	"github.com/erda-project/erda/pkg/strutil"
)

func (j *JobImpl) Delete(namespace, name string) error {
	job := apistructs.Job{}
	ctx := context.Background()
	// 多次删除后job为空结构体, jsonstore的remove接口可以再添加一个返回值判断job是否被填充
	if err := j.js.Remove(ctx, makeJobKey(namespace, name), &job); err != nil {
		return err
	}

	if len(job.Name) == 0 {
		return fmt.Errorf("job not found: namespace: %v, name:%v", namespace, name)
	}

	if _, err := j.handleJobTask(ctx, &job, task.TaskRemove); err != nil {
		return err
	}

	return nil
}

func (j *JobImpl) DeletePipelineJobs(namespace string) error {
	if namespace == "default" || namespace == "kube-system" || namespace == "kube-public" {
		return fmt.Errorf("can't delete jobs, namespace: %s", namespace)
	}
	ctx := context.Background()
	job := apistructs.Job{}
	if err := j.js.Get(ctx, makeJobKey(namespace, ""), &job); err != nil {
		return err
	}
	job.Name = ""
	// k8s 实现能直接删除所有同一namespace下的job，metronome 不支持，则要一个一个删除所有job
	if _, err := j.handleJobTask(ctx, &job, task.TaskRemove); err != nil {
		if strutil.Contains(err.Error(), "metronome not support delete pipeline jobs") {
			keys, err := j.js.ListKeys(ctx, makeJobKey(namespace, ""))
			if err != nil {
				return err
			}

			for _, k := range keys {
				splited := strutil.Split(k, "/")
				if err := j.Delete(splited[len(splited)-2], splited[len(splited)-1]); err != nil {
					return err
				}
			}
			return nil
		}
		return err
	}

	if _, err := j.js.PrefixRemove(ctx, makeJobKey(namespace, "")); err != nil {
		return err
	}

	return nil
}
