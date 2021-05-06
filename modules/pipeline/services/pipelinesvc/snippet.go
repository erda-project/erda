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

	"github.com/appscode/go/strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/pipeline_snippet_client"
	"github.com/erda-project/erda/pkg/strutil"
)

// handleQueryPipelineYamlBySnippetConfigs 统一查询 snippetConfigs
func (s *PipelineSvc) handleQueryPipelineYamlBySnippetConfigs(sourceSnippetConfigs []apistructs.SnippetConfig) (map[string]string, error) {

	sourceSupportAsyncBatch := map[string]struct{}{
		apistructs.PipelineSourceAutoTest.String(): {},
	}

	batchSourceSnippetConfigs := []apistructs.SnippetConfig{}
	singleSourceSnippetConfigs := []apistructs.SnippetConfig{}
	for _, snippetConfig := range sourceSnippetConfigs {
		if _, supportAsync := sourceSupportAsyncBatch[snippetConfig.Source]; supportAsync {
			batchSourceSnippetConfigs = append(batchSourceSnippetConfigs, snippetConfig)
		} else {
			singleSourceSnippetConfigs = append(singleSourceSnippetConfigs, snippetConfig)
		}
	}

	var wait sync.WaitGroup
	var lock sync.Mutex
	var errInfo error
	var yamlResults = make(map[string]string)

	wait.Add(1)
	go func() {
		defer wait.Done()
		batchResults, err := s.batchQueryPipelineYAMLBySnippetConfigs(batchSourceSnippetConfigs)
		if err != nil {
			errInfo = err
			return
		}

		lock.Lock()
		defer lock.Unlock()
		for k, v := range batchResults {
			yamlResults[k] = v
		}
	}()

	wait.Add(1)
	go func() {
		defer wait.Done()
		for _, singleConfig := range singleSourceSnippetConfigs {
			yml, err := s.queryPipelineYAMLBySnippetConfig(&singleConfig)
			if err != nil {
				errInfo = err
				return
			}

			lock.Lock()
			yamlResults[singleConfig.ToString()] = yml
			lock.Unlock()
		}
	}()

	wait.Wait()
	if errInfo != nil {
		return nil, errInfo
	}

	var errMsgs []string
	for _, snippetConfig := range sourceSnippetConfigs {
		yml, ok := yamlResults[snippetConfig.ToString()]
		if !ok {
			errMsgs = append(errMsgs, "source: %s, name: %s", snippetConfig.Source, snippetConfig.Name)
		}
		if strings.IsEmpty(&yml) {
			errMsgs = append(errMsgs, "source: %s, name: %s", snippetConfig.Source, snippetConfig.Name)
		}
	}
	if len(errMsgs) > 0 {
		return nil, fmt.Errorf("not found yaml for snippet configs: %s", strutil.Join(errMsgs, ", ", true))
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
func (s *PipelineSvc) batchQueryPipelineYAMLBySnippetConfigs(snippetConfigs []apistructs.SnippetConfig) (map[string]string, error) {
	yamlResults, err := pipeline_snippet_client.BatchGetSnippetPipelineYml(snippetConfigs)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, yamlResult := range yamlResults {
		result[yamlResult.Config.ToString()] = yamlResult.Yml
	}
	return result, nil
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
