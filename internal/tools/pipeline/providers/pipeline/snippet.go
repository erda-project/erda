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
	"sort"
	"sync"

	"github.com/appscode/go/strings"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v2"

	"github.com/erda-project/erda/internal/tools/pipeline/spec"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/encoding/jsonparse"
	"github.com/erda-project/erda/pkg/expression"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/pipeline_snippet_client"
	"github.com/erda-project/erda/pkg/strutil"
)

func (s *pipelineService) QueryPipelineSnippet(ctx context.Context, req *pb.PipelineSnippetQueryRequest) (*pb.PipelineSnippetQueryResponse, error) {
	data, err := s.QueryDetails(req)
	if err != nil {
		return nil, apierrors.ErrQuerySnippetYaml.InternalError(err)
	}
	return &pb.PipelineSnippetQueryResponse{Data: data}, nil
}

// handleQueryPipelineYamlBySnippetConfigs 统一查询 snippetConfigs
func (s *pipelineService) QueryDetails(req *pb.PipelineSnippetQueryRequest) (map[string]*pb.SnippetQueryDetail, error) {

	configs := req.SnippetConfigs
	if configs == nil || len(configs) <= 0 {
		return nil, fmt.Errorf("snippetConfigs is empty")
	}

	for i := range configs {
		if configs[i].Alias == "" {
			return nil, fmt.Errorf("config lost param alias")
		}
		if configs[i].Name == "" {
			return nil, fmt.Errorf("%s config lost param name", configs[i].Alias)
		}
		if configs[i].Source == "" {
			return nil, fmt.Errorf("%s config lost param source", configs[i].Alias)
		}
	}

	var snippetConfigs []*pb.SnippetDetailQuery
	for i := range configs {
		if configs[i].Source == apistructs.ActionSourceType {
			continue
		}
		snippetConfigs = append(snippetConfigs, configs[i])
	}

	result, err := s.HandleQueryPipelineYamlBySnippetConfigs(snippetConfigs)
	if err != nil {
		return nil, err
	}

	snippetDetailMap := make(map[string]*pb.SnippetQueryDetail)
	for i := range configs {
		// 假如是 action 类型就跳过
		if configs[i].Source == apistructs.ActionSourceType {
			continue
		}

		pipelineYmlContext, find := result[s.ConvertSnippetConfig2String(configs[i])]
		if !find {
			return nil, fmt.Errorf("not find snippet: %v", configs[i])
		}

		graph, err := pipelineyml.ConvertToGraphPipelineYml([]byte(pipelineYmlContext))
		if err != nil {
			return nil, err
		}

		var detail pb.SnippetQueryDetail
		// outputs
		for _, output := range graph.Outputs {
			detail.Outputs = append(detail.Outputs, expression.GenOutputRef(configs[i].Alias, output.Name))
		}

		// params
		for _, param := range graph.Params {
			detailParam := &basepb.PipelineParam{
				Name:     param.Name,
				Required: param.Required,
				Default:  param.Default,
				Desc:     param.Desc,
				Type:     param.Type,
			}
			detail.Params = append(detail.Params, detailParam)
		}

		snippetDetailMap[configs[i].Alias] = &detail
	}

	// action source 运行
	for i := range configs {
		// 假如不是 action 类型就跳过
		if configs[i].Source != apistructs.ActionSourceType {
			continue
		}

		detail, err := s.getActionDetail(configs[i])
		if err != nil {
			return nil, err
		}

		if detail != nil {
			snippetDetailMap[configs[i].Alias] = detail
		}
	}

	return snippetDetailMap, nil
}

func (s *pipelineService) getActionDetail(config *pb.SnippetDetailQuery) (*pb.SnippetQueryDetail, error) {

	var detail = &pb.SnippetQueryDetail{}

	var actionVersion = s.actionMgr.MakeActionTypeVersion(&pipelineyml.Action{
		Type:    pipelineyml.ActionType(config.Name),
		Version: config.Labels[apistructs.LabelActionVersion],
	})
	extensionSearchRequest := apistructs.ExtensionSearchRequest{
		Extensions: []string{actionVersion},
		YamlFormat: true,
	}

	actions, err := s.bdl.SearchExtensions(extensionSearchRequest)
	if err != nil {
		return nil, err
	}
	version, ok := actions[actionVersion]
	if !ok {
		return nil, nil
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

// HandleQueryPipelineYamlBySnippetConfigs unified query snippetConfigs
func (s *pipelineService) HandleQueryPipelineYamlBySnippetConfigs(sourceSnippetConfigs []*pb.SnippetDetailQuery) (map[string]string, error) {

	sourceSupportAsyncBatch := map[string]struct{}{
		apistructs.PipelineSourceAutoTest.String(): {},
	}

	batchSourceSnippetConfigs := []*pb.SnippetDetailQuery{}
	singleSourceSnippetConfigs := []*pb.SnippetDetailQuery{}
	for i := range sourceSnippetConfigs {
		if _, supportAsync := sourceSupportAsyncBatch[sourceSnippetConfigs[i].Source]; supportAsync {
			batchSourceSnippetConfigs = append(batchSourceSnippetConfigs, sourceSnippetConfigs[i])
		} else {
			singleSourceSnippetConfigs = append(singleSourceSnippetConfigs, sourceSnippetConfigs[i])
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
		for i := range singleSourceSnippetConfigs {
			yml, err := s.queryPipelineYAMLBySnippetConfig(singleSourceSnippetConfigs[i])
			if err != nil {
				errInfo = err
				return
			}

			lock.Lock()
			yamlResults[s.ConvertSnippetConfig2String(singleSourceSnippetConfigs[i])] = yml
			lock.Unlock()
		}
	}()

	wait.Wait()
	if errInfo != nil {
		return nil, errInfo
	}

	var errMsgs []string
	for i := range sourceSnippetConfigs {
		yml, ok := yamlResults[s.ConvertSnippetConfig2String(sourceSnippetConfigs[i])]
		if !ok {
			errMsgs = append(errMsgs, "source: %s, name: %s", sourceSnippetConfigs[i].Source, sourceSnippetConfigs[i].Name)
		}
		if strings.IsEmpty(&yml) {
			errMsgs = append(errMsgs, "source: %s, name: %s", sourceSnippetConfigs[i].Source, sourceSnippetConfigs[i].Name)
		}
	}
	if len(errMsgs) > 0 {
		return nil, fmt.Errorf("not found yaml for snippet configs: %s", strutil.Join(errMsgs, ", ", true))
	}

	return yamlResults, nil
}

// queryPipelineYAMLBySnippetConfig 根据 snippetConfig 查询对应的 pipeline yaml
func (s *pipelineService) queryPipelineYAMLBySnippetConfig(cfg *pb.SnippetDetailQuery) (string, error) {
	return pipeline_snippet_client.GetSnippetPipelineYml(&pb.SnippetDetailQuery{
		Name:   cfg.Name,
		Source: cfg.Source,
		Labels: cfg.Labels,
	})
}

// batchQueryPipelineYAMLBySnippetConfigs 根据 source snippetConfig 批量查询对应的 pipeline yaml
func (s *pipelineService) batchQueryPipelineYAMLBySnippetConfigs(snippetConfigs []*pb.SnippetDetailQuery) (map[string]string, error) {
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

// createSnippetPipeline4Create create pipeline object for snippetTask
func (s *pipelineService) MakeSnippetPipeline4Create(p *spec.Pipeline, snippetTask *spec.PipelineTask, yamlContent string) (*spec.Pipeline, error) {
	snippetConfig := snippetTask.Extra.Action.SnippetConfig
	// runParams
	var runParams []*basepb.PipelineRunParam
	for k, v := range snippetTask.Extra.Action.Params {
		paramValue, err := structpb.NewValue(v)
		if err != nil {
			return nil, err
		}
		runParams = append(runParams, &basepb.PipelineRunParam{Name: k, Value: paramValue})
	}
	// avoid concurrent problem
	labels := make(map[string]string, len(p.Labels))
	for k, v := range p.Labels {
		labels[k] = v
	}
	// transfer snippetTask to pipeline create request
	for k, v := range snippetConfig.Labels {
		labels[k] = v
	}
	identityInfo := p.GenIdentityInfo()
	snippetPipelineCreateReq := pb.PipelineCreateRequestV2{
		PipelineYml:            yamlContent,
		ClusterName:            snippetTask.Extra.ClusterName,
		PipelineYmlName:        snippetConfig.Name,
		PipelineSource:         snippetConfig.Source,
		Labels:                 labels,
		NormalLabels:           p.NormalLabels,
		Envs:                   p.Snapshot.Envs,
		ConfigManageNamespaces: p.Extra.ConfigManageNamespaces,
		AutoRunAtOnce:          false,
		RunParams:              runParams,
		UserID:                 identityInfo.UserID,
		InternalClient:         identityInfo.InternalClient,
	}
	if err := s.ValidateCreateRequest(&snippetPipelineCreateReq); err != nil {
		return nil, apierrors.ErrCreateSnippetPipeline.InternalError(err)
	}
	snippetP, err := s.MakePipelineFromRequestV2(&snippetPipelineCreateReq)
	if err != nil {
		return nil, err
	}
	snippetP.IsSnippet = true
	snippetP.ParentPipelineID = &p.ID
	snippetP.ParentTaskID = &snippetTask.ID
	snippetP.Extra.SnippetChain = append(p.Extra.SnippetChain, p.ID)
	return snippetP, nil
}

func (s *pipelineService) orderSnippetDetailQuery(snippetConfig *pb.SnippetDetailQuery) *apistructs.SnippetConfigOrder {
	if snippetConfig == nil {
		return nil
	}

	var snippetLabels apistructs.SnippetLabels
	if len(snippetConfig.Labels) > 0 {
		for k, v := range snippetConfig.Labels {
			snippetLabels = append(snippetLabels, apistructs.SnippetLabel{
				Key:   k,
				Value: v,
			})
		}
		sort.Sort(snippetLabels)
	}

	var order = apistructs.SnippetConfigOrder{
		Source:        snippetConfig.Source,
		Name:          snippetConfig.Name,
		SnippetLabels: snippetLabels,
	}
	return &order
}

func (s *pipelineService) ConvertSnippetConfig2String(snippetConfig *pb.SnippetDetailQuery) string {
	if snippetConfig == nil {
		return ""
	}
	return jsonparse.JsonOneLine(s.orderSnippetDetailQuery(snippetConfig))
}
