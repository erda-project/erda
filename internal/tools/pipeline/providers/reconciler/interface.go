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

package reconciler

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/pkg/action_info"
	"github.com/erda-project/erda/internal/tools/pipeline/services/actionagentsvc"
	spec2 "github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

type Interface interface {
	ReconcileOnePipeline(ctx context.Context, pipelineID uint64)

	// InjectLegacyFields TODO decouple it
	InjectLegacyFields(f *PipelineSvcFuncs, actionAgentSvc *actionagentsvc.ActionAgentSvc)
}

type PipelineSvcFuncs struct {
	CronNotExecuteCompensate                func(id uint64) error
	MergePipelineYmlTasks                   func(pipelineYml *pipelineyml.PipelineYml, dbTasks []spec2.PipelineTask, p *spec2.Pipeline, dbStages []spec2.PipelineStage, passedDataWhenCreate *action_info.PassedDataWhenCreate) (mergeTasks []spec2.PipelineTask, err error)
	HandleQueryPipelineYamlBySnippetConfigs func(sourceSnippetConfigs []apistructs.SnippetConfig) (map[string]string, error)
	MakeSnippetPipeline4Create              func(p *spec2.Pipeline, snippetTask *spec2.PipelineTask, yamlContent string) (*spec2.Pipeline, error)
	CreatePipelineGraph                     func(p *spec2.Pipeline) (stages []spec2.PipelineStage, err error)
	PreCheck                                func(p *spec2.Pipeline, stages []spec2.PipelineStage, userID string, autoRun bool) error
}

func (r *provider) InjectLegacyFields(f *PipelineSvcFuncs, actionAgentSvc *actionagentsvc.ActionAgentSvc) {
	r.pipelineSvcFuncs = f
	r.actionAgentSvc = actionAgentSvc
}
