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

package pipelinetemplate

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda-proto-go/dop/pipelinetemplate/pb"
	"github.com/erda-project/erda/apistructs"
	dbclient "github.com/erda-project/erda/internal/apps/dop/providers/pipelinetemplate/db"
	"github.com/erda-project/erda/internal/apps/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

func (p *ServiceImpl) ApplyPipelineTemplate(ctx context.Context, request *pb.PipelineTemplateApplyRequest) (*pb.PipelineTemplateCreateResponse, error) {
	var templateSpec pb.PipelineTemplateSpec
	if err := yaml.Unmarshal([]byte(request.Spec), &templateSpec); err != nil {
		return nil, apierrors.ErrCreatePipelineTemplate.InternalError(err)
	}

	if err := p.checkTemplateSpec(templateSpec); err != nil {
		return nil, apierrors.ErrCreatePipelineTemplate.InternalError(err)
	}

	createRequest := pb.PipelineTemplateCreateRequest{
		Name:           templateSpec.Name,
		ScopeType:      request.ScopeType,
		ScopeID:        request.ScopeID,
		Spec:           request.Spec,
		Desc:           templateSpec.Desc,
		LogoUrl:        templateSpec.LogoUrl,
		Version:        templateSpec.Version,
		DefaultVersion: templateSpec.DefaultVersion,
	}

	template, err := p.Create(ctx, &createRequest)
	if err != nil {
		return nil, apierrors.ErrCreatePipelineTemplate.InternalError(err)
	}

	return &pb.PipelineTemplateCreateResponse{
		Data: template,
	}, nil
}

func (p *ServiceImpl) Create(ctx context.Context, request *pb.PipelineTemplateCreateRequest) (*pb.PipelineTemplate, error) {
	if request.Name == "" {
		return nil, apierrors.ErrCreatePipelineTemplate.InvalidParameter("name")
	}

	if request.ScopeID == "" {
		return nil, apierrors.ErrCreatePipelineTemplate.InvalidParameter("scopeID")
	}

	if request.ScopeType == "" {
		return nil, apierrors.ErrCreatePipelineTemplate.InvalidParameter("scopeType")
	}

	if request.Spec == "" {
		return nil, apierrors.ErrCreatePipelineTemplate.InvalidParameter("spec")
	}

	if request.Version == "" {
		request.Version = "latest"
	}

	if request.DefaultVersion == "" {
		request.DefaultVersion = request.Version
	}

	dbPipelineTemplate, err := p.db.GetPipelineTemplate(request.Name, request.ScopeType, request.ScopeID)
	if err != nil {
		return nil, err
	}

	tx := p.db.Begin()
	//已经存在
	var saveTemplate *dbclient.DicePipelineTemplate
	if dbPipelineTemplate != nil {
		saveTemplate = dbPipelineTemplate
		saveTemplate.Desc = request.Desc
		saveTemplate.LogoUrl = request.LogoUrl
		saveTemplate.DefaultVersion = request.DefaultVersion
		if err := p.db.UpdatePipelineTemplate(saveTemplate); err != nil {
			tx.Rollback()
			return nil, errors.New(fmt.Sprintf("update pipeline template error. saveTemplate %v", saveTemplate))
		}
	} else {
		saveTemplate = &dbclient.DicePipelineTemplate{
			ScopeId:        request.ScopeID,
			ScopeType:      request.ScopeType,
			Name:           request.Name,
			Desc:           request.Desc,
			LogoUrl:        request.LogoUrl,
			DefaultVersion: request.DefaultVersion,
		}
		if err := p.db.CreatePipelineTemplate(saveTemplate); err != nil {
			tx.Rollback()
			return nil, errors.New(fmt.Sprintf("create pipeline template error. saveTemplate %v", saveTemplate))
		}
	}

	dbPipelineTemplateVersion, err := p.db.GetPipelineTemplateVersion(request.Version, saveTemplate.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if dbPipelineTemplateVersion != nil {
		dbPipelineTemplateVersion.Spec = request.Spec
		dbPipelineTemplateVersion.Readme = request.Readme
		if err := p.db.UpdatePipelineTemplateVersion(dbPipelineTemplateVersion); err != nil {
			tx.Rollback()
			return nil, errors.New(fmt.Sprintf("update pipeline template version error. saveTemplateVersion %v", dbPipelineTemplateVersion))
		}
	} else {
		dbPipelineTemplateVersion = &dbclient.DicePipelineTemplateVersion{
			Name:       saveTemplate.Name,
			Version:    request.Version,
			Spec:       request.Spec,
			Readme:     request.Readme,
			TemplateId: saveTemplate.ID,
		}
		if err := p.db.CreatePipelineTemplateVersion(dbPipelineTemplateVersion); err != nil {
			tx.Rollback()
			return nil, errors.New(fmt.Sprintf("create pipeline template version error. saveTemplateVersion %v", dbPipelineTemplateVersion))
		}
	}
	data := saveTemplate.ToApiData()
	data.Version = request.Version
	tx.Commit()
	return data, err
}

func (p *ServiceImpl) QueryPipelineTemplates(ctx context.Context, request *pb.PipelineTemplateQueryRequest) (*pb.PipelineTemplateQueryResponse, error) {
	if request.PageSize == 0 {
		request.PageSize = 20
	}
	if request.PageNo == 0 {
		request.PageNo = 1
	}
	if err := p.checkScopeTypeAndScopeId(request.ScopeType, request.ScopeID); err != nil {
		return nil, err
	}
	queryDicePipelineTemplate := dbclient.DicePipelineTemplate{
		Name:      request.Name,
		ScopeId:   request.ScopeID,
		ScopeType: request.ScopeType,
	}
	dbTemplates, total, err := p.db.QueryByPipelineTemplates(&queryDicePipelineTemplate, int(request.PageSize), int(request.PageNo))
	if err != nil {
		return nil, err
	}

	var result []*pb.PipelineTemplate

	if dbTemplates == nil {
		return &pb.PipelineTemplateQueryResponse{
			Data: &pb.PipelineTemplateQueryResponseData{
				Data:  result,
				Total: 0,
			},
		}, nil
	}

	for _, v := range dbTemplates {
		result = append(result, v.ToApiData())
	}

	return &pb.PipelineTemplateQueryResponse{
		Data: &pb.PipelineTemplateQueryResponseData{
			Data:  result,
			Total: int32(total),
		},
	}, nil
}

func (p *ServiceImpl) GetPipelineTemplateVersion(ctx context.Context, request *pb.PipelineTemplateVersionGetRequest) (*pb.PipelineTemplateVersionGetResponse, error) {
	if request.Name == "" {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InvalidParameter("name")
	}
	if request.ScopeID == "" {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InvalidParameter("scopeID")
	}
	if request.ScopeType == "" {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InvalidParameter("scopeType")
	}
	if request.Version == "" {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InvalidParameter("Version")
	}

	dbTemplate, err := p.db.GetPipelineTemplate(request.Name, request.ScopeType, request.ScopeID)
	if err != nil {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InternalError(err)
	}

	if dbTemplate == nil {
		return &pb.PipelineTemplateVersionGetResponse{
			Data: nil,
		}, nil
	}

	dbVersion, err := p.db.GetPipelineTemplateVersion(request.Version, dbTemplate.ID)
	if err != nil {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InternalError(err)
	}

	if dbVersion == nil {
		return &pb.PipelineTemplateVersionGetResponse{
			Data: nil,
		}, nil
	}

	return &pb.PipelineTemplateVersionGetResponse{
		Data: dbVersion.ToApiData(),
	}, nil
}

func (p *ServiceImpl) QueryPipelineTemplateVersions(ctx context.Context, request *pb.PipelineTemplateVersionQueryRequest) (*pb.PipelineTemplateVersionQueryResponse, error) {
	if request.Name == "" {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InvalidParameter("name")
	}
	if request.ScopeID == "" {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InvalidParameter("scopeID")
	}
	if request.ScopeType == "" {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InvalidParameter("scopeType")
	}

	dbTemplate, err := p.db.GetPipelineTemplate(request.Name, request.ScopeType, request.ScopeID)
	if err != nil {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InternalError(err)
	}
	if dbTemplate == nil {
		return &pb.PipelineTemplateVersionQueryResponse{Data: nil}, nil
	}

	version := dbclient.DicePipelineTemplateVersion{
		TemplateId: dbTemplate.ID,
	}

	dbVersions, err := p.db.QueryPipelineTemplateVersions(&version)
	if err != nil {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InternalError(err)
	}
	if dbVersions == nil {
		return &pb.PipelineTemplateVersionQueryResponse{Data: nil}, nil
	}

	var result []*pb.PipelineTemplateVersion
	for _, v := range dbVersions {
		result = append(result, v.ToApiData())
	}
	return &pb.PipelineTemplateVersionQueryResponse{
		Data: result,
	}, nil
}

func (p *ServiceImpl) RenderPipelineTemplate(ctx context.Context, request *pb.PipelineTemplateRenderRequest) (*pb.PipelineTemplateRenderResponse, error) {
	if request.Name == "" {
		return nil, apierrors.ErrRenderPipelineTemplate.InvalidParameter("name")
	}
	if request.ScopeID == "" {
		return nil, apierrors.ErrRenderPipelineTemplate.InvalidParameter("scopeID")
	}
	if request.ScopeType == "" {
		return nil, apierrors.ErrRenderPipelineTemplate.InvalidParameter("scopeType")
	}
	if request.Version == "" {
		return nil, apierrors.ErrRenderPipelineTemplate.InvalidParameter("Version")
	}
	if apistructs.TemplateVersion(request.TemplateVersion) == apistructs.TemplateVersionV2 {
		if request.Alias == "" {
			return nil, apierrors.ErrRenderPipelineTemplate.InvalidParameter("alias")
		}
	}

	if apistructs.TemplateVersion(request.TemplateVersion) == apistructs.TemplateVersionV2 {
		if request.Alias == "" {
			return nil, apierrors.ErrRenderPipelineTemplate.InvalidParameter("alias")
		}
	}

	getPipelineTemplateRequest := pb.PipelineTemplateVersionGetRequest{
		Name:      request.Name,
		ScopeType: request.ScopeType,
		ScopeID:   request.ScopeID,
		Version:   request.Version,
	}

	templateVersion, err := p.GetPipelineTemplateVersion(ctx, &getPipelineTemplateRequest)
	if err != nil {
		return nil, apierrors.ErrRenderPipelineTemplate.InternalError(err)
	}

	if templateVersion.Data == nil {
		return nil, apierrors.ErrRenderPipelineTemplate.InternalError(errors.New(fmt.Sprintf(" not find template version. request %v", request)))
	}

	specYaml := templateVersion.Data.Spec

	pipelineYaml, outputs, err := RenderTemplate(specYaml, request)
	if err != nil {
		return nil, apierrors.ErrRenderPipelineTemplate.InternalError(err)
	}

	render := &pb.PipelineTemplateRender{
		PipelineYaml: pipelineYaml,
		Version:      templateVersion.Data,
		Outputs:      outputs,
	}

	return &pb.PipelineTemplateRenderResponse{
		Data: render,
	}, nil
}

func (p *ServiceImpl) RenderPipelineTemplateBySpec(ctx context.Context, request *pb.PipelineTemplateRenderSpecRequest) (*pb.PipelineTemplateRenderResponse, error) {

	if request.Spec == nil {
		return nil, apierrors.ErrRenderPipelineTemplate.InvalidParameter("spec")
	}
	if request.Alias == "" {
		return nil, apierrors.ErrRenderPipelineTemplate.InvalidParameter("alias")
	}

	params := request.Spec.Params
	if params != nil {
		for _, v := range params {
			if err := p.checkPipelineParam(v); err != nil {
				return nil, apierrors.ErrRenderPipelineTemplate.InvalidParameter(fmt.Sprintf("snippet %s definition params name can not empty", request.Spec.Name))
			}
		}
	}

	pipelineYaml, outputs, err := pipelineyml.DoRenderTemplateWithFormat(convertParamMap(request.Params), request.Spec, request.Alias, apistructs.TemplateVersion(request.TemplateVersion))
	if err != nil {
		return nil, apierrors.ErrRenderPipelineTemplate.InternalError(err)
	}

	render := &pb.PipelineTemplateRender{
		PipelineYaml: pipelineYaml,
		Version:      nil,
		Outputs:      outputs,
	}
	return &pb.PipelineTemplateRenderResponse{
		Data: render,
	}, nil
}

func (p *ServiceImpl) QuerySnippetYml(ctx context.Context, req *pb.QuerySnippetYmlRequest) (*pb.QuerySnippetYmlResponse, error) {
	getPipelineTemplateRequest := &pb.PipelineTemplateVersionGetRequest{
		Name:      req.Name,
		ScopeType: req.Source,
		ScopeID:   req.Labels[apistructs.LabelDiceSnippetScopeID],
		Version:   req.Labels[apistructs.LabelChooseSnippetVersion],
	}

	templateVersion, err := p.GetPipelineTemplateVersion(ctx, getPipelineTemplateRequest)
	if err != nil {
		return nil, err
	}
	if templateVersion.Data == nil {
		return &pb.QuerySnippetYmlResponse{Data: ""}, nil
	}

	var templateAction pb.PipelineTemplateSpec
	if err := yaml.Unmarshal([]byte(templateVersion.Data.Spec), &templateAction); err != nil {
		p.log.Errorf("Unmarshal specYaml error: %v, yaml: %s", err, templateVersion.Data.Spec)
	}

	return &pb.QuerySnippetYmlResponse{
		Data: templateAction.Template,
	}, nil
}

func RenderTemplate(specYaml string, request *pb.PipelineTemplateRenderRequest) (string, []*pb.SnippetFormatOutputs, error) {
	alias := request.Alias

	var templateAction apistructs.PipelineTemplateSpec
	if err := yaml.Unmarshal([]byte(specYaml), &templateAction); err != nil {
		logrus.Errorf("Unmarshal specYaml error: %v, yaml: %s", err, specYaml)
		return "", nil, err
	}
	params := make([]*pb.PipelineParam, 0)
	for _, v := range templateAction.Params {
		defaultValue, err := structpb.NewValue(v.Default)
		if err != nil {
			return "", nil, err
		}
		params = append(params, &pb.PipelineParam{
			Name:     v.Name,
			Required: v.Required,
			Default:  defaultValue,
			Desc:     v.Desc,
			Type:     v.Type,
		})
	}
	outputs := make([]*pb.PipelineOutput, 0)
	for _, v := range templateAction.Outputs {
		outputs = append(outputs, &pb.PipelineOutput{
			Name: v.Name,
			Desc: v.Desc,
			Ref:  v.Ref,
		})
	}
	pbTemplateAction := &pb.PipelineTemplateSpec{
		Name:           templateAction.Name,
		Version:        templateAction.Version,
		Desc:           templateAction.Desc,
		LogoUrl:        templateAction.LogoUrl,
		Params:         params,
		Outputs:        outputs,
		Template:       templateAction.Template,
		DefaultVersion: templateAction.DefaultVersion,
	}

	return pipelineyml.DoRenderTemplateWithFormat(convertParamMap(request.Params), pbTemplateAction, alias, apistructs.TemplateVersion(request.TemplateVersion))
}

func (p *ServiceImpl) checkTemplateSpec(spec pb.PipelineTemplateSpec) error {
	if spec.Name == "" {
		return errors.New("spec name can not empty")
	}

	if spec.Template == "" {
		return errors.New("spec template can not empty")
	}

	if spec.Params != nil {
		for _, v := range spec.Params {
			if err := p.checkPipelineParam(v); err != nil {
				return err
			}
		}
	}

	if spec.Outputs != nil {
		for _, v := range spec.Outputs {
			if err := p.checkPipelineOutput(v); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *ServiceImpl) checkPipelineParam(param *pb.PipelineParam) error {
	if param.Name == "" {
		return errors.New("params name can not empty")
	}

	return nil
}

func (p *ServiceImpl) checkPipelineOutput(output *pb.PipelineOutput) error {
	if output.Name == "" {
		return errors.New("outputs name can not empty")
	}

	if output.Ref == "" {
		return errors.New("outputs ref can not empty")
	}

	return nil
}

func (p *ServiceImpl) checkScopeTypeAndScopeId(scopeType, scopeID string) error {
	if scopeType == "" {
		return apierrors.ErrQueryPipelineTemplate.InvalidParameter("scopeID")
	}
	if scopeID == "" {
		return apierrors.ErrQueryPipelineTemplate.InvalidParameter("scopeType")
	}

	return nil
}

func convertParamMap(param map[string]*structpb.Value) map[string]interface{} {
	if param == nil {
		return map[string]interface{}{}
	}

	res := map[string]interface{}{}
	for k, v := range param {
		res[k] = v.AsInterface()
	}
	return res
}
