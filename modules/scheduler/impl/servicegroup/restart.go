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

package servicegroup

import (
	"context"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/task"
)

func (s ServiceGroupImpl) Restart(namespace string, name string) error {
	sg := apistructs.ServiceGroup{}
	if err := s.js.Get(context.Background(), mkServiceGroupKey(namespace, name), &sg); err != nil {
		return err
	}
	sg.Extra[LastRestartTimeKey] = time.Now().String()
	sg.LastModifiedTime = time.Now().Unix()

	if err := s.js.Put(context.Background(), mkServiceGroupKey(sg.Type, sg.ID), &sg); err != nil {
		return err
	}

	sg.Labels = appendServiceTags(sg.Labels, sg.Executor)

	if _, err := s.handleServiceGroup(context.Background(), &sg, task.TaskUpdate); err != nil {
		return err
	}
	return nil
}
