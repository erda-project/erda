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

package queue

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/erda-project/erda-proto-go/core/pipeline/queue/pb"
	"github.com/erda-project/erda/apistructs"
)

type projectQueueLabelKey string

func (p projectQueueLabelKey) String() string {
	return string(p)
}

const (
	projectQueueLabelKeyProjectName projectQueueLabelKey = "projectName"
	projectQueueLabelKeyWorkspace   projectQueueLabelKey = "workspace"
	projectQueueLabelKeyProjectID   projectQueueLabelKey = "projectID"
)

type Interface interface {
	// IdempotentGetProjectLevelQueue returns project level queue
	// if not exists, automatically create one by given workspace and project resource config
	IdempotentGetProjectLevelQueue(workspace string, project *apistructs.ProjectDTO) (*pb.Queue, error)

	InjectQueueManager(queueManager pb.QueueServiceServer)
}

func (p *provider) createProjectLevelQueue(workspace string, project *apistructs.ProjectDTO) (*pb.Queue, error) {
	resourceConfig := project.ResourceConfig.GetWSConfig(workspace)
	if resourceConfig == nil {
		return nil, fmt.Errorf("workspace: %s resource info not found", workspace)
	}
	queueResource, err := p.calculateProjectResource(workspace, project)
	if err != nil {
		return nil, err
	}
	queue, err := p.QueueManager.CreateQueue(context.Background(), &pb.QueueCreateRequest{
		Name:             p.makeProjectLevelQueueName(workspace, project.Name),
		ClusterName:      resourceConfig.ClusterName,
		PipelineSource:   apistructs.PipelineSourceDice.String(),
		ScheduleStrategy: apistructs.ScheduleStrategyInsidePipelineQueueOfFIFO.String(),
		MaxCPU:           queueResource.MaxCPU,
		MaxMemoryMB:      queueResource.MaxMemoryMB,
		Concurrency:      queueResource.Concurrency,
		Labels: map[string]string{
			projectQueueLabelKeyProjectName.String(): project.Name,
			projectQueueLabelKeyWorkspace.String():   workspace,
			projectQueueLabelKeyProjectID.String():   strconv.FormatUint(project.ID, 10),
		},
	})
	if err != nil {
		p.Log.Errorf("failed to create project level queue: %v", err)
		return nil, err
	}
	return queue.Data, nil
}

func (p *provider) IdempotentGetProjectLevelQueue(workspace string, project *apistructs.ProjectDTO) (*pb.Queue, error) {
	clusterName, ok := project.ClusterConfig[strings.ToUpper(workspace)]
	if !ok {
		return nil, fmt.Errorf("workspace: %s cluster not found", workspace)
	}
	queuePaging, err := p.QueueManager.PagingQueue(context.Background(), &pb.QueuePagingRequest{
		Name:             p.makeProjectLevelQueueName(workspace, project.Name),
		ClusterName:      clusterName,
		PipelineSources:  []string{apistructs.PipelineSourceDice.String()},
		ScheduleStrategy: apistructs.ScheduleStrategyInsidePipelineQueueOfFIFO.String(),
		MustMatchLabels:  makeMustMatchLabels(workspace, project),
		PageSize:         1,
		PageNo:           1,
	})
	if err != nil {
		return nil, err
	}
	if len(queuePaging.Queues) > 0 {
		return queuePaging.Queues[0], nil
	}
	return p.createProjectLevelQueue(workspace, project)
}

func (p *provider) InjectQueueManager(queueManager pb.QueueServiceServer) {
	p.QueueManager = queueManager
}

func (p *provider) makeProjectLevelQueueName(workspace string, projectName string) string {
	return fmt.Sprintf("project-%s-%s", projectName, workspace)
}

func makeMustMatchLabels(workspace string, project *apistructs.ProjectDTO) []string {
	return []string{
		fmt.Sprintf("%s=%s", projectQueueLabelKeyProjectName, project.Name),
		fmt.Sprintf("%s=%s", projectQueueLabelKeyWorkspace, workspace),
		fmt.Sprintf("%s=%d", projectQueueLabelKeyProjectID, project.ID),
	}
}
