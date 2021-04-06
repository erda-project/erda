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

package servicegroup

import (
	"context"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/impl/clusterinfo"
	"github.com/erda-project/erda/modules/scheduler/task"
)

func (s ServiceGroupImpl) Create(req apistructs.ServiceGroupCreateV2Request) (apistructs.ServiceGroup, error) {
	sg, err := convertServiceGroupCreateV2Request(req, s.clusterinfo)
	if err != nil {
		return apistructs.ServiceGroup{}, err
	}
	if err := s.js.Put(context.Background(), mkServiceGroupKey(sg.Type, sg.ID), sg); err != nil {
		return apistructs.ServiceGroup{}, err
	}
	sg.Labels = appendServiceTags(sg.Labels, sg.Executor)
	if _, err := s.handleServiceGroup(context.Background(), &sg, task.TaskCreate); err != nil {
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
