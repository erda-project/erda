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

package projectpipeline

import (
	"context"

	"github.com/erda-project/erda-infra/base/logs"
	dpb "github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	sourcepb "github.com/erda-project/erda-proto-go/core/pipeline/source/pb"
	"github.com/erda-project/erda-proto-go/dop/projectpipeline/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/providers/projectpipeline/deftype"
	"github.com/erda-project/erda/modules/dop/services/permission"
	"github.com/erda-project/erda/modules/dop/services/pipeline"
)

type ProjectPipelineService struct {
	logger             logs.Logger
	db                 *dao.DBClient
	bundle             *bundle.Bundle
	pipelineSourceType ProjectSourceType

	pipelineSvc        *pipeline.Pipeline
	PipelineSource     sourcepb.SourceServiceServer
	PipelineDefinition dpb.DefinitionServiceServer
	Permission         *permission.Permission
}

func (p *ProjectPipelineService) WithPipelineSvc(svc *pipeline.Pipeline) {
	p.pipelineSvc = svc
}

func (p *ProjectPipelineService) WithPermissionSvc(permission *permission.Permission) {
	p.Permission = permission
}

type Service interface {
	Create(ctx context.Context, params *pb.CreateProjectPipelineRequest) (*pb.CreateProjectPipelineResponse, error)
	List(ctx context.Context, params deftype.ProjectPipelineList) ([]*dpb.PipelineDefinition, int64, error)
	Delete(ctx context.Context, params deftype.ProjectPipelineDelete) (*deftype.ProjectPipelineDeleteResult, error)
	Update(ctx context.Context, params deftype.ProjectPipelineUpdate) (*deftype.ProjectPipelineUpdateResult, error)
	SetPrimary(ctx context.Context, params deftype.ProjectPipelineCategory) (*dpb.PipelineDefinitionUpdateResponse, error)
	UnSetPrimary(ctx context.Context, params deftype.ProjectPipelineCategory) (*dpb.PipelineDefinitionUpdateResponse, error)
	ListApp(ctx context.Context, params *pb.ListAppRequest) (*pb.ListAppResponse, error)
	ListPipelineYmlList(ctx context.Context, req *pb.ListAppPipelineYmlRequest) (*pb.ListAppPipelineYmlResponse, error)

	Run(ctx context.Context, params deftype.ProjectPipelineRun) (*deftype.ProjectPipelineRunResult, error)
	BatchRun(ctx context.Context, params deftype.ProjectPipelineBatchRun) (*deftype.ProjectPipelineBatchRunResult, error)
	Cancel(ctx context.Context, params deftype.ProjectPipelineCancel) (*deftype.ProjectPipelineCancelResult, error)
	FailRerun(ctx context.Context, params deftype.ProjectPipelineFailRerun) (*deftype.ProjectPipelineFailRerunResult, error)
	Rerun(ctx context.Context, params deftype.ProjectPipelineRerun) (*deftype.ProjectPipelineRerunResult, error)
	StartCron(ctx context.Context, params deftype.ProjectPipelineStartCron) (*deftype.ProjectPipelineStartCronResult, error)
	EndCron(ctx context.Context, params deftype.ProjectPipelineEndCron) (*deftype.ProjectPipelineEndCronResult, error)
	ListExecHistory(ctx context.Context, params deftype.ProjectPipelineListExecHistory) (*deftype.ProjectPipelineListExecHistoryResult, error)
}
