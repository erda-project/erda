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

package graph

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v2"

	"github.com/erda-project/erda-infra/base/logs"
	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"github.com/erda-project/erda-proto-go/core/pipeline/graph/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/i18n"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/strutil"
)

type graphService struct {
	p   *provider
	bdl *bundle.Bundle
	log logs.Logger
}

func (s *graphService) PipelineYmlGraph(ctx context.Context, req *pb.PipelineYmlGraphRequest) (*pb.PipelineYmlGraphResponse, error) {
	graph, err := pipelineyml.ConvertToGraphPipelineYml([]byte(req.PipelineYmlContent))
	if err != nil {
		return nil, apierrors.ErrParsePipelineYml.InvalidParameter(err)
	}
	if graph == nil {
		return &pb.PipelineYmlGraphResponse{
			Data: graph,
		}, nil
	}
	lang := apis.GetLang(ctx)
	if lang == "" {
		lang = i18n.ZH
	}
	if err := s.loadGraphActionNameAndLogo(graph, lang); err != nil {
		return nil, apierrors.ErrParsePipelineYml.InternalError(err)
	}
	return &pb.PipelineYmlGraphResponse{Data: graph}, nil
}

func (s *graphService) loadGraphActionNameAndLogo(graph *basepb.PipelineYml, lang string) error {
	stages := graph.Stages
	var graphStages [][]*basepb.PipelineYmlAction
	stageBytes, err := stages.MarshalJSON()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(stageBytes, &graphStages); err != nil {
		return err
	}
	var extensionSearchRequest = apistructs.ExtensionSearchRequest{}
	extensionSearchRequest.YamlFormat = true
	for _, stage := range graphStages {
		for _, action := range stage {
			if action.Type == apistructs.ActionTypeSnippet {
				continue
			}
			extensionSearchRequest.Extensions = append(extensionSearchRequest.Extensions, action.Type)
		}
	}
	if extensionSearchRequest.Extensions != nil {
		extensionSearchRequest.Extensions = strutil.DedupSlice(extensionSearchRequest.Extensions, true)
	}

	resultMap, err := s.bdl.SearchExtensions(extensionSearchRequest)
	if err != nil {
		s.log.Errorf("pipelineYmlGraph to SearchExtensions error: %v", err)
		return err
	}
	stageList := make([]interface{}, 0)
	for _, stage := range graphStages {
		actionStage := make([]interface{}, 0)
		for _, action := range stage {
			if action.Type == pipelineyml.Snippet {
				action.LogoUrl = pipelineyml.SnippetLogo
				action.DisplayName = pipelineyml.SnippetDisplayName
				action.Description = pipelineyml.SnippetDesc
				continue
			}

			version, ok := resultMap[action.Type]
			if !ok {
				continue
			}

			specYmlStr, ok := version.Spec.(string)
			if !ok {
				continue
			}

			var actionSpec apistructs.ActionSpec
			if err := yaml.Unmarshal([]byte(specYmlStr), &actionSpec); err != nil {
				logrus.Errorf("pipelineYmlGraph Unmarshal spec error: %v", err)
				continue
			}

			action.DisplayName = actionSpec.GetLocaleDisplayName(lang)
			action.LogoUrl = actionSpec.LogoUrl
			action.Description = actionSpec.GetLocaleDesc(lang)
			actionValue, err := convertAction2Value(action)
			if err != nil {
				return err
			}
			actionStage = append(actionStage, actionValue.AsInterface())
		}
		stageList = append(stageList, actionStage)
	}
	newStages, err := structpb.NewList(stageList)
	if err != nil {
		return err
	}
	graph.Stages = newStages
	return nil
}

func convertAction2Value(action *basepb.PipelineYmlAction) (*structpb.Value, error) {
	var dat interface{}
	byteDat, err := action.MarshalJSON()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(byteDat, &dat); err != nil {
		return nil, err
	}
	return structpb.NewValue(dat)
}
