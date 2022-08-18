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
	"fmt"

	"github.com/c2h5oh/datasize"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/numeral"
)

var (
	defaultZeroFloat64Resource = float64(0)
	defaultZeroInt64Resource   = int64(0)
)

type ProjectQueueResource struct {
	Concurrency int64
	MaxCPU      float64
	MaxMemoryMB float64
}

// calculateConcurrency calculates the queue concurrency for the given project.
// each application is allocated two concurrency
func (p *provider) calculateConcurrency(project *apistructs.ProjectDTO) (int64, error) {
	apps, err := p.bdl.CountAppByProID(project.ID)
	if err != nil {
		return 0, err
	}
	concurrency := apps * 2
	return numeral.MaxInt64([]int64{concurrency, p.Cfg.DefaultQueueConcurrency, defaultZeroInt64Resource}), nil
}

func (p *provider) calculateMaxCpu(resourceConfig *apistructs.ResourceConfigInfo) (float64, error) {
	return numeral.MaxFloat64([]float64{p.Cfg.DefaultQueueCPU, resourceConfig.CPUQuota, defaultZeroFloat64Resource}), nil
}

func (p *provider) calculateMaxMemoryMB(resourceConfig *apistructs.ResourceConfigInfo) (float64, error) {
	memGB := datasize.GB * datasize.ByteSize(resourceConfig.MemQuota)
	return numeral.MaxFloat64([]float64{p.Cfg.DefaultQueueMemoryMB, memGB.MBytes(), defaultZeroFloat64Resource}), nil
}

// calculateProjectResource calculates the queue resource for the given project.
// TODO: calculate the real resource situation of the cluster to allocate queues
func (p *provider) calculateProjectResource(workspace string, project *apistructs.ProjectDTO) (*ProjectQueueResource, error) {
	resourceConfig := project.ResourceConfig.GetWSConfig(workspace)
	if resourceConfig == nil {
		return nil, fmt.Errorf("workspace: %s resource info not found", workspace)
	}
	concurrency, err := p.calculateConcurrency(project)
	if err != nil {
		return nil, err
	}
	maxCPU, err := p.calculateMaxCpu(resourceConfig)
	if err != nil {
		return nil, err
	}
	maxMemoryMB, err := p.calculateMaxMemoryMB(resourceConfig)
	if err != nil {
		return nil, err
	}
	return &ProjectQueueResource{
		Concurrency: concurrency,
		MaxCPU:      maxCPU,
		MaxMemoryMB: maxMemoryMB,
	}, nil
}
