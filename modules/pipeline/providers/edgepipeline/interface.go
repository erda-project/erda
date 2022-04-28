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

package edgepipeline

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/pipelinesvc"
)

type Interface interface {
	CreatePipeline(ctx context.Context, req *apistructs.PipelineCreateRequestV2) (*apistructs.PipelineDTO, error)
	InjectLegacyFields(pSvc *pipelinesvc.PipelineSvc)
}

func (p *provider) InjectLegacyFields(pipelineSvc *pipelinesvc.PipelineSvc) {
	p.pipelineSvc = pipelineSvc
}

func (p *provider) CreatePipeline(ctx context.Context, req *apistructs.PipelineCreateRequestV2) (*apistructs.PipelineDTO, error) {
	isEdge := p.EdgePipelineRegister.ShouldDispatchToEdge(req.PipelineSource.String(), req.ClusterName)
	if !isEdge {
		pipeline, err := p.pipelineSvc.CreateV2(ctx, req)
		if err != nil {
			return nil, err
		}
		return p.pipelineSvc.ConvertPipeline(pipeline), nil
	}
	edgeBundle, err := p.EdgePipelineRegister.GetEdgeBundleByClusterName(req.ClusterName)
	if err != nil {
		return nil, err
	}
	pipelineDto, err := edgeBundle.CreatePipeline(req)
	if err != nil {
		return nil, err
	}
	return pipelineDto, nil
}
