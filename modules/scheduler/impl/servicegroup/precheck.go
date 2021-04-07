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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/task"
)

func (s ServiceGroupImpl) Precheck(req apistructs.ServiceGroupPrecheckRequest) (apistructs.ServiceGroupPrecheckData, error) {
	sg, err := convertServiceGroupCreateV2Request(apistructs.ServiceGroupCreateV2Request(req), s.clusterinfo)
	if err != nil {
		return apistructs.ServiceGroupPrecheckData{}, err
	}
	t, err := s.handleServiceGroup(context.Background(), &sg, task.TaskPrecheck)
	if err != nil {
		return apistructs.ServiceGroupPrecheckData{}, err
	}

	return t.Extra.(apistructs.ServiceGroupPrecheckData), nil
}
