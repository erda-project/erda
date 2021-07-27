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

package pipelinesvc

import (
	"fmt"
	"strconv"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	queuepb "github.com/erda-project/erda-proto-go/core/pipeline/queue/pb"
	"github.com/erda-project/erda/apistructs"
)

func (s *PipelineSvc) validateQueueFromLabels(req *basepb.PipelineCreateRequest) (*queuepb.Queue, error) {
	var foundBindQueueID bool
	var bindQueueIDStr string
	for k, v := range req.Labels {
		if k == apistructs.LabelBindPipelineQueueID {
			foundBindQueueID = true
			bindQueueIDStr = v
			break
		}
	}
	if !foundBindQueueID {
		return nil, nil
	}
	// parse queue id
	queueID, err := strconv.ParseUint(bindQueueIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bindQueueID: %s, err: %v", bindQueueIDStr, err)
	}
	// query queue
	queue, err := s.queueManage.GetPipelineQueue(queueID)
	if err != nil {
		return nil, err
	}
	// check queue is matchable
	if err := checkQueueValidateWithPipelineCreateReq(req, queue); err != nil {
		return nil, err
	}
	req.BindQueue = queue

	return queue, nil
}

func checkQueueValidateWithPipelineCreateReq(req *basepb.PipelineCreateRequest, queue *queuepb.Queue) error {
	// pipeline source
	if queue.PipelineSource != req.PipelineSource {
		return fmt.Errorf("invalid queue: pipeline source not match: %s(req) vs %s(queue)", req.PipelineSource, queue.PipelineSource)
	}
	// cluster name
	if queue.ClusterName != req.ClusterName {
		return fmt.Errorf("invalid queue: cluster name not match: %s(req) vs %s(queue)", req.ClusterName, queue.ClusterName)
	}

	return nil
}
