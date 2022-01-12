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

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/impl/clusterinfo"
	"github.com/erda-project/erda/modules/scheduler/task"
)

func (s ServiceGroupImpl) Create(req apistructs.ServiceGroupCreateV2Request) (apistructs.ServiceGroup, error) {
	sg, err := convertServiceGroupCreateV2Request(req, s.clusterinfo)
	if err != nil {
		return apistructs.ServiceGroup{}, errors.Errorf("failed to convert sg createV2Request, err: %v", err)
	}

	if err := s.js.Put(context.Background(), mkServiceGroupKey(sg.Type, sg.ID), sg); err != nil {
		logrus.Errorf("failed to put sg to jsonStore, err: %v", err)
		return apistructs.ServiceGroup{}, err
	}

	sg.Labels = appendServiceTags(sg.Labels, sg.Executor)
	if _, err := s.handleServiceGroup(context.Background(), &sg, task.TaskCreate); err != nil {
		logrus.Errorf("failed to handle sg, err: %v", err)
		return apistructs.ServiceGroup{}, err
	}

	return sg, err
}

func convertServiceGroupCreateV2Request(req apistructs.ServiceGroupCreateV2Request, clusterinfo clusterinfo.ClusterInfo) (apistructs.ServiceGroup, error) {
	sg, err := convertServiceGroup(req, clusterinfo)
	if err != nil {
		return apistructs.ServiceGroup{}, err
	}
	sg.CreatedTime = time.Now().Unix()
	sg.LastModifiedTime = sg.CreatedTime
	sg.Status = apistructs.StatusCreated
	return sg, nil
}
