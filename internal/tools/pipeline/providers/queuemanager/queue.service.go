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

package queuemanager

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-proto-go/core/pipeline/queue/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/strutil"
)

func (p *provider) CreateQueue(ctx context.Context, req *pb.QueueCreateRequest) (*pb.QueueCreateResponse, error) {
	if err := p.ValidateQueueCreateRequest(req); err != nil {
		return nil, apierrors.ErrCreatePipelineQueue.InvalidParameter(err)
	}
	queue, err := p.dbClient.CreatePipelineQueue(req)
	if err != nil {
		return nil, apierrors.ErrCreatePipelineQueue.InternalError(err)
	}
	return &pb.QueueCreateResponse{Data: queue}, nil
}

func (p *provider) GetQueue(ctx context.Context, req *pb.QueueGetRequest) (*pb.QueueGetResponse, error) {
	queue, exist, err := p.dbClient.GetPipelineQueue(req.QueueID)
	if err != nil {
		return nil, apierrors.ErrGetPipelineQueue.InternalError(err)
	}
	if !exist {
		return nil, apierrors.ErrGetPipelineQueue.NotFound()
	}
	queue.Usage = p.DistributedQueryQueueUsage(ctx, queue)
	return &pb.QueueGetResponse{Data: queue}, nil
}

func (p *provider) PagingQueue(ctx context.Context, req *pb.QueuePagingRequest) (*pb.QueuePagingResponse, error) {
	pagingResult, err := p.dbClient.PagingPipelineQueues(req)
	if err != nil {
		return nil, apierrors.ErrPagingPipelineQueues.InternalError(err)
	}
	return pagingResult, nil
}

func (p *provider) UpdateQueue(ctx context.Context, req *pb.QueueUpdateRequest) (*pb.QueueUpdateResponse, error) {
	if err := p.ValidateQueueUpdateRequest(req); err != nil {
		return nil, apierrors.ErrUpdatePipelineQueue.InvalidParameter(err)
	}
	queue, err := p.dbClient.UpdatePipelineQueue(req)
	if err != nil {
		return nil, apierrors.ErrUpdatePipelineQueue.InternalError(err)
	}
	p.DistributedUpdateQueue(ctx, req.QueueID)
	return &pb.QueueUpdateResponse{Data: queue}, nil
}

func (p *provider) DeleteQueue(ctx context.Context, req *pb.QueueDeleteRequest) (*pb.QueueDeleteResponse, error) {
	err := p.dbClient.DeletePipelineQueue(req.QueueID)
	if err != nil {
		return nil, apierrors.ErrDeletePipelineQueue.InternalError(err)
	}
	return &pb.QueueDeleteResponse{}, nil
}

// ValidateQueueCreateRequest validate and handle request.
func (p *provider) ValidateQueueCreateRequest(req *pb.QueueCreateRequest) error {
	// name
	if err := strutil.Validate(req.Name, strutil.MinLenValidator(1), strutil.MaxRuneCountValidator(191)); err != nil {
		return fmt.Errorf("invalid name: %v", err)
	}
	// source
	if req.PipelineSource == "" {
		return fmt.Errorf("missing pipelineSource")
	}
	if !apistructs.PipelineSource(req.PipelineSource).Valid() {
		return fmt.Errorf("invalid pipelineSource: %s", req.PipelineSource)
	}
	// clusterName
	if req.ClusterName == "" {
		return fmt.Errorf("missing clusterName")
	}
	// strategy
	if req.ScheduleStrategy == "" {
		req.ScheduleStrategy = apistructs.PipelineQueueDefaultScheduleStrategy.String()
	}
	// mode
	if req.Mode == "" {
		req.Mode = apistructs.PipelineQueueDefaultMode.String()
	}
	if !apistructs.PipelineQueueMode(req.Mode).IsValid() {
		return fmt.Errorf("invalid mode: %s", req.Mode)
	}
	// scheduleStrategy
	if !apistructs.ScheduleStrategyInsidePipelineQueue(req.ScheduleStrategy).IsValid() {
		return fmt.Errorf("invalid schedule strategy: %s", req.ScheduleStrategy)
	}
	// priority
	if req.Priority == 0 {
		req.Priority = apistructs.PipelineQueueDefaultPriority
	}
	if req.Priority < 0 {
		return fmt.Errorf("priority must > 0")
	}
	// concurrency
	if req.Concurrency == 0 {
		req.Concurrency = apistructs.PipelineQueueDefaultConcurrency
	}
	if req.Concurrency < 0 {
		return fmt.Errorf("concurrency must > 0")
	}
	// max cpu
	if req.MaxCPU < 0 {
		return fmt.Errorf("max cpu must >= 0")
	}
	// max memoryMB
	if req.MaxMemoryMB < 0 {
		return fmt.Errorf("max memory(MB) must >= 0")
	}
	return nil
}

// ValidateQueueUpdateRequest update queue request.
func (p *provider) ValidateQueueUpdateRequest(req *pb.QueueUpdateRequest) error {
	// id
	if req.QueueID == 0 {
		return fmt.Errorf("missing queue id")
	}
	// pipeline source
	if req.PipelineSource != "" {
		return fmt.Errorf("cannot change queue's source")
	}

	return nil
}
