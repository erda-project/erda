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

package queuemanage

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
)

// UpdatePipelineQueue update queue all fields by id.
func (qm *QueueManage) UpdatePipelineQueue(req apistructs.PipelineQueueUpdateRequest) (*apistructs.PipelineQueue, error) {
	// validate
	if err := req.Validate(); err != nil {
		return nil, apierrors.ErrUpdatePipelineQueue.InvalidParameter(err)
	}

	// do update
	queue, err := qm.dbClient.UpdatePipelineQueue(req)
	if err != nil {
		return nil, apierrors.ErrUpdatePipelineQueue.InternalError(err)
	}

	return queue, nil
}
