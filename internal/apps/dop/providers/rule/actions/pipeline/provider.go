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
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	definitionpb "github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	sourcepb "github.com/erda-project/erda-proto-go/core/pipeline/source/pb"
	"github.com/erda-project/erda-proto-go/dop/projectpipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/providers/projectpipeline"
	"github.com/erda-project/erda/internal/apps/dop/services/apierrors"
	"github.com/erda-project/erda/internal/apps/dop/services/branchrule"
	"github.com/erda-project/erda/internal/apps/dop/services/pipeline"
	"github.com/erda-project/erda/internal/pkg/diceworkspace"
	"github.com/erda-project/erda/pkg/common/apis"
)

type provider struct {
	bdl                *bundle.Bundle
	ProjectPipelineSvc *projectpipeline.ProjectPipelineService `autowired:"erda.dop.projectpipeline.ProjectPipelineService"`
	PipelineSource     sourcepb.SourceServiceServer            `autowired:"erda.core.pipeline.source.SourceService" required:"true"`
	PipelineDefinition definitionpb.DefinitionServiceServer    `autowired:"erda.core.pipeline.definition.DefinitionService" required:"true"`
	pipeline           *pipeline.Pipeline
	branchRule         *branchrule.BranchRule
}

func (p *provider) WithPipelineSvc(svc *pipeline.Pipeline) {
	p.pipeline = svc
}

func (p *provider) WithBranchRule(branchRule *branchrule.BranchRule) {
	p.branchRule = branchRule
}

type Interface interface {
	CreatePipeline(env map[string]interface{}) (string, error)
	WithPipelineSvc(svc *pipeline.Pipeline)
	WithBranchRule(branchRule *branchrule.BranchRule)
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bdl = bundle.New(bundle.WithPipeline())
	return nil
}

type PipelineConfig struct {
	PipelineYml    string
	PipelineYmlStr string
	AppID          uint64
	RefName        string
	UserID         string
	Path           string
	FileName       string
}

func (p *provider) CreatePipeline(env map[string]interface{}) (string, error) {
	gitPush, ok := env["git_push"]
	if !ok {
		return "", fmt.Errorf("empty git_push config")
	}
	v, ok := gitPush.(map[string]interface{})["pipelineConfig"]
	if !ok {
		return "", fmt.Errorf("empty pipeline config")
	}
	// parse pipeline config
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	var req PipelineConfig
	if err := json.Unmarshal(b, &req); err != nil {
		return "", err
	}

	app, err := p.bdl.GetApp(req.AppID)
	if err != nil {
		return "", err
	}

	// create pipeline
	ctx := context.Background()
	definitionID, err := p.GetPipelineDefinitionID(apis.WithUserIDContext(ctx, req.UserID), app, req.RefName, req.Path, req.FileName, req.PipelineYmlStr)
	if err != nil {
		logrus.Errorf("failed to bind definition %v", err)
	}

	reqPipeline := &apistructs.PipelineCreateRequest{
		AppID:              req.AppID,
		Branch:             req.RefName,
		Source:             apistructs.PipelineSourceDice,
		PipelineYmlSource:  apistructs.PipelineYmlSourceGittar,
		PipelineYmlContent: req.PipelineYmlStr,
		AutoRun:            true,
		UserID:             req.UserID,
	}
	v2, err := p.pipeline.ConvertPipelineToV2(reqPipeline)
	if err != nil {
		logrus.Errorf("error convert to pipelineV2 %s, (%+v)", req.PipelineYmlStr, err)
		return "", err
	}
	projectRules, err := p.branchRule.Query(apistructs.ProjectScope, int64(app.ProjectID))
	if err != nil {
		return "", apierrors.ErrFetchConfigNamespace.InternalError(err)
	}
	validBranch := diceworkspace.GetValidBranchByGitReference(reqPipeline.Branch, projectRules)
	workspace := validBranch.Workspace
	v2.ForceRun = true
	v2.DefinitionID = definitionID
	v2.PipelineYmlName = fmt.Sprintf("%d/%s/%s/%s", reqPipeline.AppID, workspace, req.RefName, strings.TrimPrefix(req.PipelineYml, "/"))

	resp, err := p.pipeline.CreatePipelineV2(v2)
	if err != nil {
		logrus.Errorf("create pipeline failed, pipeline: %s, (%+v)", req.PipelineYmlStr, err)
		return "", err
	}

	// store created pipeline id
	return strconv.FormatUint(resp.ID, 10), nil
}

func (p *provider) GetPipelineDefinitionID(ctx context.Context, app *apistructs.ApplicationDTO, branch, path, name, strPipelineYml string) (definitionID string, err error) {
	if app == nil {
		return "", nil
	}
	sourceList, err := p.PipelineSource.List(ctx, &sourcepb.PipelineSourceListRequest{
		Remote:     fmt.Sprintf("%v/%v/%v", app.OrgName, app.ProjectName, app.Name),
		Ref:        branch,
		Name:       name,
		Path:       path,
		SourceType: apistructs.SourceTypeErda,
	})
	if err != nil {
		return "", err
	}

	var source *sourcepb.PipelineSource
	for _, v := range sourceList.Data {
		source = v
		break
	}
	if source != nil {
		definitionList, err := p.PipelineDefinition.List(ctx, &definitionpb.PipelineDefinitionListRequest{
			SourceIDList: []string{source.ID},
			Location:     apistructs.MakeLocation(app, apistructs.PipelineTypeCICD),
		})
		if err != nil {
			return "", nil
		}

		for _, definition := range definitionList.Data {
			return definition.ID, nil
		}
	}

	const sourceType = "erda"
	projectPipeline, err := p.ProjectPipelineSvc.Create(ctx, &pb.CreateProjectPipelineRequest{
		ProjectID:  app.ProjectID,
		Name:       projectpipeline.MakeProjectPipelineName(strPipelineYml, name),
		AppID:      app.ID,
		SourceType: sourceType,
		Ref:        branch,
		Path:       path,
		FileName:   name,
	})
	if err != nil {
		return "", err
	}
	return projectPipeline.ProjectPipeline.ID, nil
}

func init() {
	servicehub.Register("erda.dop.rule.action.pipeline", &servicehub.Spec{
		Services: []string{"erda.core.rule.action.pipeline"},
		Types:    []reflect.Type{reflect.TypeOf((*Interface)(nil)).Elem()},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
