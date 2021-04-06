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
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/pipeline_snippet_client"
)

const apiTest = "api-test"

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

	snippetDetailMap := make(map[string]apistructs.SnippetQueryDetail)

	for _, config := range configs {

		// 假如是 action 类型就跳过
		if config.Source == apistructs.SnippetActionSourceType {
			continue
		}

		snippetConfig := pipelineyml.HandleSnippetConfigLabel(&pipelineyml.SnippetConfig{
			Name:   config.Name,
			Source: config.Source,
			Labels: config.Labels,
		}, nil)

		config.Labels = snippetConfig.Labels
		pipelineYmlContext, err := pipeline_snippet_client.GetSnippetPipelineYml(config.SnippetConfig)

		if err != nil {
			return nil, err
		}

		graph, err := pipelineyml.ConvertToGraphPipelineYml([]byte(pipelineYmlContext))
		if err != nil {
			return nil, err
		}

		var detail apistructs.SnippetQueryDetail

		// outputs
		for _, output := range graph.Outputs {
			detail.Outputs = append(detail.Outputs, getOutputRef(config.Alias, output.Name))
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
		if config.Source != apistructs.SnippetActionSourceType {
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

func getOutputRef(alias, outputName string) string {
	return fmt.Sprintf("${%s:OUTPUT:%s}", alias, outputName)
}

func (s *SnippetSvc) getActionDetail(config apistructs.SnippetDetailQuery) (*apistructs.SnippetQueryDetail, error) {

	var detail apistructs.SnippetQueryDetail

	// config 的类型是 api-test 的类型，就需要额外的去拿 out_params
	if config.Name == apiTest {

		actionJson := config.Labels[apistructs.LabelActionJson]
		var action apistructs.PipelineYmlAction
		err := json.Unmarshal([]byte(actionJson), &action)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to parse actionJson %s", config.Alias))
		}

		params := action.Params
		if params != nil {

			outParamsBytes, err := json.Marshal(action.Params["out_params"])
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("failed to parse %s actionJson out_params %s", apiTest, config.Alias))
			}

			var outParams []apistructs.APIOutParam
			err = json.Unmarshal(outParamsBytes, &outParams)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("failed to unparse %s actionJson out_params %s", apiTest, config.Alias))
			}

			if outParams != nil {
				for _, out := range outParams {
					detail.Outputs = append(detail.Outputs, getOutputRef(config.Alias, out.Key))
				}
			}

		}
	}

	// 其他 action 需要去查询其 action 的 output
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

	outputs := actionSpec.Outputs
	if outputs == nil {
		return &detail, nil
	}

	for _, output := range outputs {
		detail.Outputs = append(detail.Outputs, getOutputRef(config.Alias, output.Name))
	}

	return &detail, nil
}
