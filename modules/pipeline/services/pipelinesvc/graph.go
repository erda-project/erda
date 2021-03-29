package pipelinesvc

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

func (s *PipelineSvc) PipelineYmlGraph(req *apistructs.PipelineYmlParseGraphRequest) (*apistructs.PipelineYml, error) {
	graph, err := pipelineyml.ConvertToGraphPipelineYml([]byte(req.PipelineYmlContent))
	if err != nil {
		return nil, apierrors.ErrParsePipelineYml.InvalidParameter(err)
	}

	if graph == nil {
		return graph, nil
	}

	// 设置logo和名称
	s.loadGraphActionNameAndLogo(graph)

	return graph, nil
}

func (s *PipelineSvc) loadGraphActionNameAndLogo(graph *apistructs.PipelineYml) {
	var extensionSearchRequest = apistructs.ExtensionSearchRequest{}
	extensionSearchRequest.YamlFormat = true
	for _, stage := range graph.Stages {
		for _, action := range stage {
			extensionSearchRequest.Extensions = append(extensionSearchRequest.Extensions, action.Type)
		}
	}

	resultMap, err := s.bdl.SearchExtensions(extensionSearchRequest)
	if err != nil {
		logrus.Errorf("pipelineYmlGraph to SearchExtensions error: %v", err)
		return
	}
	for _, stage := range graph.Stages {
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

			action.DisplayName = actionSpec.DisplayName
			action.LogoUrl = actionSpec.LogoUrl
			action.Description = actionSpec.Desc
		}
	}
}
