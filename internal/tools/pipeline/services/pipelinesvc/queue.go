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

package pipelinesvc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/erda-project/erda-proto-go/core/pipeline/queue/pb"
	"github.com/erda-project/erda/apistructs"
)

func (s *PipelineSvc) validateQueueFromLabels(req *apistructs.PipelineCreateRequestV2) (*pb.Queue, error) {
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
	queueRes, err := s.queueManage.GetQueue(context.Background(), &pb.QueueGetRequest{QueueID: queueID})
	if err != nil {
		return nil, err
	}
	// check queue is matchable
	if err := checkQueueValidateWithPipelineCreateReq(req, queueRes.Data); err != nil {
		return nil, err
	}
	req.BindQueue = queueRes.Data

	return queueRes.Data, nil
}

func checkQueueValidateWithPipelineCreateReq(req *apistructs.PipelineCreateRequestV2, queue *pb.Queue) error {
	// pipeline source
	if queue.PipelineSource != req.PipelineSource.String() {
		return fmt.Errorf("invalid queue: pipeline source not match: %s(req) vs %s(queue)", req.PipelineSource, queue.PipelineSource)
	}
	// cluster name
	if queue.ClusterName != req.ClusterName {
		return fmt.Errorf("invalid queue: cluster name not match: %s(req) vs %s(queue)", req.ClusterName, queue.ClusterName)
	}

	return nil
}
