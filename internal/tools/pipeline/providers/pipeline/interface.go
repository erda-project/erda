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

package pipeline

import (
	"context"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/internal/tools/pipeline/pkg/action_info"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

type Interface interface {
	pb.PipelineServiceServer

	Detail(ctx context.Context, pipelineID uint64) (*pb.PipelineDetailDTO, error)
	List(req *pb.PipelinePagingRequest) (*pb.PipelineListResponseData, error)
	PreCheck(p *spec.Pipeline, stages []spec.PipelineStage, userID string, autoRun bool) error
	DealPipelineCallbackOfAction(data []byte) (err error)
	CreatePipelineGraph(p *spec.Pipeline) (newStages []spec.PipelineStage, err error)
	CreateV2(ctx context.Context, req *pb.PipelineCreateRequestV2) (*spec.Pipeline, error)
	ConvertPipelineBase(p spec.PipelineBase) basepb.PipelineDTO
	ConvertPipeline(p *spec.Pipeline) *basepb.PipelineDTO
	Convert2PagePipeline(p *spec.Pipeline) *pb.PagePipeline
	BatchConvert2PagePipeline(pipelines []spec.Pipeline) []*pb.PagePipeline
	ValidateCreateRequest(req *pb.PipelineCreateRequestV2) error
	MakePipelineFromRequestV2(req *pb.PipelineCreateRequestV2) (*spec.Pipeline, error)
	GetYmlActionTasks(pipelineYml *pipelineyml.PipelineYml, p *spec.Pipeline, dbStages []spec.PipelineStage, passedDataWhenCreate *action_info.PassedDataWhenCreate) []spec.PipelineTask
	MergePipelineYmlTasks(pipelineYml *pipelineyml.PipelineYml, dbTasks []spec.PipelineTask, p *spec.Pipeline, dbStages []spec.PipelineStage, passedDataWhenCreate *action_info.PassedDataWhenCreate) (mergeTasks []spec.PipelineTask, err error)
	HandleQueryPipelineYamlBySnippetConfigs(sourceSnippetConfigs []*pb.SnippetDetailQuery) (map[string]string, error)
	ConvertSnippetConfig2String(snippetConfig *pb.SnippetDetailQuery) string
	MakeSnippetPipeline4Create(p *spec.Pipeline, snippetTask *spec.PipelineTask, yamlContent string) (*spec.Pipeline, error)
}
