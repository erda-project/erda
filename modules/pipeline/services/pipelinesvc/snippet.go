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

package pipelinesvc

import (
	"fmt"
	"sync"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/pipeline_snippet_client"
	"github.com/erda-project/erda/pkg/strutil"
)

// handleQueryPipelineYamlBySnippetConfigs 统一查询 snippetConfigs
func (s *PipelineSvc) handleQueryPipelineYamlBySnippetConfigs(sourceSnippetConfigs map[string]map[string]apistructs.SnippetConfig) (map[string]map[string]string, error) {
	// 支持批量查询的 source
	sourceSupportAsyncBatch := map[string]struct{}{
		apistructs.PipelineSourceAutoTest.String(): {},
	}

	// 不同 source 并行处理
	batchSourceSnippetConfigs := make(map[string]map[string]apistructs.SnippetConfig)
	singleSourceSnippetConfigs := make(map[string]map[string]apistructs.SnippetConfig)
	for source, snippetConfigMap := range sourceSnippetConfigs {
		// 若 source 支持批量查询，则批量查询
		if _, supportAsync := sourceSupportAsyncBatch[source]; supportAsync {
			batchSourceSnippetConfigs[source] = snippetConfigMap
		} else { // 不支持，则单个查询
			singleSourceSnippetConfigs[source] = snippetConfigMap
		}
	}

	// TODO 同一 source 考虑全局并发度，防止把客户端打满造成 timeout

	yamlResults := make(map[string]map[string]string)

	// 异步批量查询异步调用
	var wg sync.WaitGroup
	var batchErrs []error
	var lock sync.Mutex
	for i := range batchSourceSnippetConfigs {
		wg.Add(1)

		batchConfigs := batchSourceSnippetConfigs[i]
		go func() {
			defer wg.Done()

			var snippetConfigs []apistructs.SnippetConfig
			for _, cfg := range batchConfigs {
				snippetConfigs = append(snippetConfigs, cfg)
			}
			batchResults, err := s.batchQueryPipelineYAMLBySnippetConfigs(snippetConfigs)
			if err != nil {
				lock.Lock()
				batchErrs = append(batchErrs, err)
				lock.Unlock()
				return
			}

			lock.Lock()
			for source, nameYamls := range batchResults {
				for name, yml := range nameYamls {
					if _, ok := yamlResults[source]; !ok {
						yamlResults[source] = make(map[string]string)
					}
					yamlResults[source][name] = yml
				}
			}
			lock.Unlock()
		}()
	}
	wg.Wait()
	if len(batchErrs) > 0 {
		return nil, strutil.FlatErrors(batchErrs, ", ")
	}

	// 不支持批量查询的仍单个调用
	for i := range singleSourceSnippetConfigs {
		singleConfigs := singleSourceSnippetConfigs[i]
		for _, sc := range singleConfigs {
			yml, err := s.queryPipelineYAMLBySnippetConfig(&sc)
			if err != nil {
				return nil, err
			}
			if _, ok := yamlResults[sc.Source]; !ok {
				yamlResults[sc.Source] = make(map[string]string)
			}
			yamlResults[sc.Source][sc.Name] = yml
		}
	}

	return yamlResults, nil
}

// queryPipelineYAMLBySnippetConfig 根据 snippetConfig 查询对应的 pipeline yaml
func (s *PipelineSvc) queryPipelineYAMLBySnippetConfig(cfg *apistructs.SnippetConfig) (string, error) {
	return pipeline_snippet_client.GetSnippetPipelineYml(apistructs.SnippetConfig{
		Name:   cfg.Name,
		Source: cfg.Source,
		Labels: cfg.Labels,
	})
}

// batchQueryPipelineYAMLBySnippetConfigs 根据 source snippetConfig 批量查询对应的 pipeline yaml
func (s *PipelineSvc) batchQueryPipelineYAMLBySnippetConfigs(snippetConfigs []apistructs.SnippetConfig) (map[string]map[string]string, error) {
	yamlResults, err := pipeline_snippet_client.BatchGetSnippetPipelineYml(snippetConfigs)
	if err != nil {
		return nil, err
	}
	sourceSnippetConfigMap := make(map[string]map[string]string)
	for _, yamlResult := range yamlResults {
		source := yamlResult.Config.Source
		if sourceSnippetConfigMap[source] == nil {
			sourceSnippetConfigMap[source] = make(map[string]string)
		}
		name := yamlResult.Config.Name
		sourceSnippetConfigMap[source][name] = yamlResult.Yml
	}
	// 校验传入的请求是否都有返回，没有则报错
	notFoundMap := make(map[string]string) // key: source, value: name
	for _, sc := range snippetConfigs {
		if m, ok := sourceSnippetConfigMap[sc.Source]; !ok {
			notFoundMap[sc.Source] = sc.Name
		} else {
			if _, ok := m[sc.Name]; !ok {
				notFoundMap[sc.Source] = sc.Name
			}
		}
	}
	if len(notFoundMap) > 0 {
		var errMsgs []string
		for source, name := range notFoundMap {
			errMsgs = append(errMsgs, "source: %s, name: %s", source, name)
		}
		return nil, fmt.Errorf("not found yaml for snippet configs: %s", strutil.Join(errMsgs, ", ", true))
	}
	return sourceSnippetConfigMap, nil
}

// createSnippetPipeline4Create 为 snippetTask 创建流水线对象
func (s *PipelineSvc) makeSnippetPipeline4Create(p *spec.Pipeline, snippetTask *spec.PipelineTask, yamlContent string) (*spec.Pipeline, error) {
	snippetConfig := snippetTask.Extra.Action.SnippetConfig
	// runParams
	var runParams []apistructs.PipelineRunParam
	for k, v := range snippetTask.Extra.Action.Params {
		runParams = append(runParams, apistructs.PipelineRunParam{Name: k, Value: v})
	}
	// transfer snippetTask to pipeline create request
	snippetPipelineCreateReq := apistructs.PipelineCreateRequestV2{
		PipelineYml:            yamlContent,
		ClusterName:            snippetTask.Extra.ClusterName,
		PipelineYmlName:        snippetConfig.Name,
		PipelineSource:         apistructs.PipelineSource(snippetConfig.Source),
		Labels:                 p.Labels,
		NormalLabels:           p.NormalLabels,
		Envs:                   p.Snapshot.Envs,
		ConfigManageNamespaces: p.Extra.ConfigManageNamespaces,
		AutoRunAtOnce:          false,
		RunParams:              runParams,
		IdentityInfo:           p.GenIdentityInfo(),
	}
	if err := s.validateCreateRequest(&snippetPipelineCreateReq); err != nil {
		return nil, apierrors.ErrCreateSnippetPipeline.InternalError(err)
	}
	snippetP, err := s.makePipelineFromRequestV2(&snippetPipelineCreateReq)
	if err != nil {
		return nil, err
	}
	snippetP.IsSnippet = true
	snippetP.ParentPipelineID = &p.ID
	snippetP.ParentTaskID = &snippetTask.ID
	snippetP.Extra.SnippetChain = append(p.Extra.SnippetChain, p.ID)
	return snippetP, nil
}
