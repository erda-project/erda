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
