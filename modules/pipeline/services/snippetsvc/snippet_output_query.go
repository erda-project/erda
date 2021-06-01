// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package snippetsvc

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/expression"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/pipeline_snippet_client"
)

func (s *SnippetSvc) QueryDetails(req *apistructs.SnippetQueryDetailsRequest) (map[string]apistructs.SnippetQueryDetail, error) {

	configs := req.SnippetConfigs
	if configs == nil || len(configs) <= 0 {
		return nil, fmt.Errorf("snippetConfigs is empty")
	}

	for _, config := range configs {
		if config.Alias == "" {
			return nil, fmt.Errorf("config lost param alias")
		}
		if config.Name == "" {
			return nil, fmt.Errorf("%s config lost param name", config.Alias)
		}
		if config.Source == "" {
			return nil, fmt.Errorf("%s config lost param source", config.Alias)
		}
	}

	var snippetConfigs []apistructs.SnippetConfig
	for _, config := range configs {
		if config.Source == apistructs.ActionSourceType {
			continue
		}
		snippetConfigs = append(snippetConfigs, config.SnippetConfig)
	}
	yamlResults, err := pipeline_snippet_client.BatchGetSnippetPipelineYml(snippetConfigs)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, yamlResult := range yamlResults {
		result[yamlResult.Config.ToString()] = yamlResult.Yml
	}

	snippetDetailMap := make(map[string]apistructs.SnippetQueryDetail)
	for _, config := range configs {
		// 假如是 action 类型就跳过
		if config.Source == apistructs.ActionSourceType {
			continue
		}

		pipelineYmlContext, find := result[config.SnippetConfig.ToString()]
		if !find {
			return nil, fmt.Errorf("not find snippet: %v", config)
		}

		graph, err := pipelineyml.ConvertToGraphPipelineYml([]byte(pipelineYmlContext))
		if err != nil {
			return nil, err
		}

		var detail apistructs.SnippetQueryDetail
		// outputs
		for _, output := range graph.Outputs {
			detail.Outputs = append(detail.Outputs, expression.GenOutputRef(config.Alias, output.Name))
		}

		// params
		for _, param := range graph.Params {
			detail.Params = append(detail.Params, param)
		}

		snippetDetailMap[config.Alias] = detail
	}

	// action source 运行
	for _, config := range configs {
		// 假如不是 action 类型就跳过
		if config.Source != apistructs.ActionSourceType {
			continue
		}

		detail, err := s.getActionDetail(config)
		if err != nil {
			return nil, err
		}

		if detail != nil {
			snippetDetailMap[config.Alias] = *detail
		}
	}

	return snippetDetailMap, nil
}

func (s *SnippetSvc) getActionDetail(config apistructs.SnippetDetailQuery) (*apistructs.SnippetQueryDetail, error) {

	var detail = &apistructs.SnippetQueryDetail{}

	req := apistructs.ExtensionVersionGetRequest{
		Name:       config.Name,
		Version:    config.Labels[apistructs.LabelActionVersion],
		YamlFormat: true,
	}
	version, err := s.bdl.GetExtensionVersion(req)
	if err != nil {
		return nil, err
	}

	specYml, ok := version.Spec.(string)
	if !ok {
		return nil, fmt.Errorf("action %s spec not string", config.Alias)
	}

	actionSpec := apistructs.ActionSpec{}
	err = yaml.Unmarshal([]byte(specYml), &actionSpec)
	if err != nil {
		return nil, err
	}

	for _, output := range actionSpec.Outputs {
		detail.Outputs = append(detail.Outputs, expression.GenOutputRef(config.Alias, output.Name))
	}

	actionJson := config.Labels[apistructs.LabelActionJson]
	var action apistructs.PipelineYmlAction
	err = json.Unmarshal([]byte(actionJson), &action)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to parse actionJson %s", config.Alias))
	}

	for _, matchOutputs := range actionSpec.OutputsFromParams {
		switch matchOutputs.Type {
		case apistructs.JqActionMatchOutputType:
			outputs, err := handlerActionOutputsWithJq(&action, matchOutputs.Expression)
			if err != nil {
				return nil, err
			}
			for _, output := range outputs {
				detail.Outputs = append(detail.Outputs, expression.GenOutputRef(config.Alias, output))
			}
		}
	}

	return detail, nil
}
