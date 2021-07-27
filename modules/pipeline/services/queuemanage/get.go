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
	queuepb "github.com/erda-project/erda-proto-go/core/pipeline/queue/pb"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
)

// GetPipelineQueue get a pipeline queue by id.
func (qm *QueueManage) GetPipelineQueue(queueID uint64) (*queuepb.Queue, error) {
	queue, exist, err := qm.dbClient.GetPipelineQueue(queueID)
	if err != nil {
		return nil, apierrors.ErrGetPipelineQueue.InternalError(err)
	}
	if !exist {
		return nil, apierrors.ErrGetPipelineQueue.NotFound()
	}
	return queue, nil
}
