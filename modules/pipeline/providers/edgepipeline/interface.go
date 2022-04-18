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
	"fmt"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/services/pipelinesvc"
	"github.com/erda-project/erda/pkg/clusterdialer"
	"github.com/erda-project/erda/pkg/discover"
)

type Interface interface {
	CreatePipeline(ctx context.Context, req *apistructs.PipelineCreateRequestV2) (*apistructs.PipelineDTO, error)
	InjectLegacyFields(pSvc *pipelinesvc.PipelineSvc)
}

func (p *provider) InjectLegacyFields(pipelineSvc *pipelinesvc.PipelineSvc) {
	p.pipelineSvc = pipelineSvc
}

func (p *provider) ShouldDispatchToEdge(source, clusterName string) bool {
	if clusterName == "" {
		return false
	}
	if p.Cfg.ClusterName == clusterName {
		return false
	}
	var findInWhitelist bool
	for _, whiteListSource := range p.Cfg.AllowedSources {
		if strings.HasPrefix(source, whiteListSource) {
			findInWhitelist = true
			break
		}
	}
	if !findInWhitelist {
		return false
	}
	isEdge, err := p.bdl.IsClusterDialerClientRegistered(clusterName, apistructs.ClusterDialerClientTypePipeline.String())
	if !isEdge || err != nil {
		return false
	}
	return true
}

func (p *provider) GetDialContextByClusterName(clusterName string) clusterdialer.DialContextFunc {
	clusterKey := apistructs.ClusterDialerClientTypePipeline.MakeClientKey(clusterName)
	return clusterdialer.DialContext(clusterKey)
}

func (p *provider) GetEdgeBundleByClusterName(clusterName string) (*bundle.Bundle, error) {
	edgeDial := p.GetDialContextByClusterName(clusterName)
	edgeDetail, err := p.bdl.GetClusterDialerClientData(apistructs.ClusterDialerClientTypePipeline.String(), clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to get edge bundle for cluster %s, err: %v", clusterName, err)
	}
	pipelineAddr := edgeDetail.Get(apistructs.ClusterDialerDataKeyPipelineAddr)
	return bundle.New(bundle.WithDialContext(edgeDial), bundle.WithCustom(discover.EnvPipeline, pipelineAddr)), nil
}

func (p *provider) CreatePipeline(ctx context.Context, req *apistructs.PipelineCreateRequestV2) (*apistructs.PipelineDTO, error) {
	isEdge := p.ShouldDispatchToEdge(req.PipelineSource.String(), req.ClusterName)
	if !isEdge {
		pipeline, err := p.pipelineSvc.CreateV2(ctx, req)
		if err != nil {
			return nil, err
		}
		return p.pipelineSvc.ConvertPipeline(pipeline), nil
	}
	edgeBundle, err := p.GetEdgeBundleByClusterName(req.ClusterName)
	if err != nil {
		return nil, err
	}
	pipelineDto, err := edgeBundle.CreatePipeline(req)
	if err != nil {
		return nil, err
	}
	return pipelineDto, nil
}
