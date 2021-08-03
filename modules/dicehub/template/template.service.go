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

package template

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"

	pb "github.com/erda-project/erda-proto-go/core/dicehub/template/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dicehub/service/apierrors"
	"github.com/erda-project/erda/modules/dicehub/template/db"
)

type templateService struct {
	p  *provider
	db *db.TemplateDB
}

func (s *templateService) ApplyPipelineTemplate(ctx context.Context, req *pb.PipelineTemplateApplyRequest) (*pb.PipelineTemplateCreateResponse, error) {
	var templateSpec *pb.PipelineTemplateSpec
	if err := yaml.Unmarshal([]byte(req.Spec), &templateSpec); err != nil {
		return nil, apierrors.ErrCreateExtension.InternalError(err)
	}

	if err := checkTemplateSpec(templateSpec); err != nil {
		return nil, apierrors.ErrCreateExtension.InternalError(err)
	}

	createRequest := pb.PipelineTemplateCreateRequest{
		Name:           templateSpec.Name,
		ScopeType:      req.ScopeType,
		ScopeID:        req.ScopeID,
		Spec:           req.Spec,
		Desc:           templateSpec.Desc,
		LogoUrl:        templateSpec.LogoUrl,
		Version:        templateSpec.Version,
		DefaultVersion: templateSpec.DefaultVersion,
	}

	resp, err := s.Create(&createRequest)
	if err != nil {
		return nil, err
	}
	return &pb.PipelineTemplateCreateResponse{Data: resp}, nil
}

func (s *templateService) QueryPipelineTemplates(ctx context.Context, req *pb.PipelineTemplateQueryRequest) (*pb.PipelineTemplateQueryResponse, error) {
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	if req.PageNo == 0 {
		req.PageNo = 1
	}

	queryDicePipelineTemplate := db.DicePipelineTemplate{
		Name:      req.Name,
		ScopeId:   req.ScopeID,
		ScopeType: req.ScopeType,
	}
	dbTemplates, total, err := s.db.QueryByPipelineTemplates(&queryDicePipelineTemplate, req.PageSize, req.PageNo)
	if err != nil {
		return nil, apierrors.ErrQueryPipelineTemplate.InternalError(err)
	}

	var result []*pb.PipelineTemplate

	if dbTemplates == nil {
		return nil, apierrors.ErrQueryPipelineTemplate.InternalError(err)
	}

	for _, v := range dbTemplates {
		result = append(result, v.ToApiData())
	}

	return &pb.PipelineTemplateQueryResponse{
		Data: &pb.PipelineTemplateQueryResponseData{
			Data:  result,
			Total: total,
		},
	}, nil
}

func (s *templateService) RenderPipelineTemplate(ctx context.Context, req *pb.PipelineTemplateRenderRequest) (*pb.PipelineTemplateRenderResponse, error) {
	name, err := url.QueryUnescape(req.Name)
	if err != nil {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InvalidParameter("name")
	}
	req.Name = name

	if req.Params == nil {
		req.Params = map[string]*structpb.Value{}
	}

	if req.TemplateVersion != int32(TemplateVersionV1) && req.TemplateVersion != int32(TemplateVersionV2) {
		req.TemplateVersion = int32(TemplateVersionV1)
	}

	result, err := s.RenderPipelineTemplateRequest(req)

	if err != nil {
		return nil, apierrors.ErrRenderPipelineTemplate.InternalError(err)
	}

	return &pb.PipelineTemplateRenderResponse{Data: result}, nil
}

func (s *templateService) RenderPipelineTemplateBySpec(ctx context.Context, req *pb.PipelineTemplateRenderSpecRequest) (*pb.PipelineTemplateRenderResponse, error) {
	if req.Params == nil {
		req.Params = map[string]*structpb.Value{}
	}

	if req.TemplateVersion != int32(TemplateVersionV1) && req.TemplateVersion != int32(TemplateVersionV2) {
		req.TemplateVersion = int32(TemplateVersionV1)
	}
	result, err := s.RenderPipelineTemplateBySpecRequest(req)
	if err != nil {
		return nil, apierrors.ErrRenderPipelineTemplate.InternalError(err)
	}

	return &pb.PipelineTemplateRenderResponse{Data: result}, nil
}

func (s *templateService) GetPipelineTemplateVersion(ctx context.Context, req *pb.PipelineTemplateVersionGetRequest) (*pb.PipelineTemplateVersionGetResponse, error) {
	name, err := url.QueryUnescape(req.Name)
	if err != nil {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InvalidParameter("name")
	}
	req.Name = name

	result, err := s.getPipelineTemplateVersion(req)

	if err != nil {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InternalError(err)
	}

	return &pb.PipelineTemplateVersionGetResponse{Data: result}, nil
}

func (s *templateService) QueryPipelineTemplateVersions(ctx context.Context, req *pb.PipelineTemplateVersionQueryRequest) (*pb.PipelineTemplateVersionQueryResponse, error) {
	name, err := url.QueryUnescape(req.Name)
	if err != nil {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InvalidParameter("name")
	}
	req.Name = name

	result, err := s.queryPipelineTemplateVersions(req)

	if err != nil {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InternalError(err)
	}

	return &pb.PipelineTemplateVersionQueryResponse{Data: result}, nil
}

func (s *templateService) QuerySnippetYml(ctx context.Context, req *pb.QuerySnippetYmlRequest) (*pb.QuerySnippetYmlResponse, error) {
	getPipelineTemplateRequest := pb.PipelineTemplateVersionGetRequest{
		Name:      req.Name,
		ScopeType: req.Source,
		ScopeID:   req.Labels[apistructs.LabelDiceSnippetScopeID],
		Version:   req.Labels[apistructs.LabelChooseSnippetVersion],
	}

	templateVersion, err := s.getPipelineTemplateVersion(&getPipelineTemplateRequest)
	if err != nil {
		return nil, apierrors.QuerySnippetYml.InternalError(err)
	}

	var templateAction pb.PipelineTemplateSpec
	if err := yaml.Unmarshal([]byte(templateVersion.Spec), &templateAction); err != nil {
		logrus.Errorf("Unmarshal specYaml error: %v, yaml: %s", err, templateVersion.Spec)
		return nil, apierrors.QuerySnippetYml.InternalError(err)
	}

	return &pb.QuerySnippetYmlResponse{Data: templateAction.Template}, nil
}

func (s *templateService) Create(request *pb.PipelineTemplateCreateRequest) (*pb.PipelineTemplate, error) {
	if request.Name == "" {
		return nil, apierrors.ErrCreateExtension.InvalidParameter("name")
	}

	if request.ScopeID == "" {
		return nil, apierrors.ErrCreateExtension.InvalidParameter("scopeID")
	}

	if request.ScopeType == "" {
		return nil, apierrors.ErrCreateExtension.InvalidParameter("scopeType")
	}

	if request.Spec == "" {
		return nil, apierrors.ErrCreateExtension.InvalidParameter("spec")
	}

	if request.Version == "" {
		request.Version = "latest"
	}

	if request.DefaultVersion == "" {
		request.DefaultVersion = request.Version
	}

	dbPipelineTemplate, err := s.db.GetPipelineTemplate(request.Name, request.ScopeType, request.ScopeID)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	// already exists
	var saveTemplate *db.DicePipelineTemplate
	if dbPipelineTemplate != nil {
		saveTemplate = dbPipelineTemplate
		saveTemplate.Desc = request.Desc
		saveTemplate.LogoUrl = request.LogoUrl
		saveTemplate.DefaultVersion = request.DefaultVersion
		if err := s.db.UpdatePipelineTemplate(saveTemplate); err != nil {
			tx.Rollback()
			return nil, errors.New(fmt.Sprintf("update pipeline template error. saveTemplate %v", saveTemplate))
		}
	} else {
		saveTemplate = &db.DicePipelineTemplate{
			ScopeId:        request.ScopeID,
			ScopeType:      request.ScopeType,
			Name:           request.Name,
			Desc:           request.Desc,
			LogoUrl:        request.LogoUrl,
			DefaultVersion: request.DefaultVersion,
		}
		if err := s.db.CreatePipelineTemplate(saveTemplate); err != nil {
			tx.Rollback()
			return nil, errors.New(fmt.Sprintf("create pipeline template error. saveTemplate %v", saveTemplate))
		}
	}

	dbPipelineTemplateVersion, err := s.db.GetPipelineTemplateVersion(request.Version, saveTemplate.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if dbPipelineTemplateVersion != nil {
		dbPipelineTemplateVersion.Spec = request.Spec
		dbPipelineTemplateVersion.Readme = request.Readme
		if err := s.db.UpdatePipelineTemplateVersion(dbPipelineTemplateVersion); err != nil {
			tx.Rollback()
			return nil, errors.New(fmt.Sprintf("update pipeline template version error. saveTemplateVersion %v", dbPipelineTemplateVersion))
		}
	} else {
		dbPipelineTemplateVersion = &db.DicePipelineTemplateVersion{
			Name:       saveTemplate.Name,
			Version:    request.Version,
			Spec:       request.Spec,
			Readme:     request.Readme,
			TemplateId: saveTemplate.ID,
		}
		if err := s.db.CreatePipelineTemplateVersion(dbPipelineTemplateVersion); err != nil {
			tx.Rollback()
			return nil, errors.New(fmt.Sprintf("create pipeline template version error. saveTemplateVersion %v", dbPipelineTemplateVersion))
		}
	}
	data := saveTemplate.ToApiData()
	data.Version = request.Version
	tx.Commit()
	return data, err
}

func (s *templateService) queryPipelineTemplateVersions(request *pb.PipelineTemplateVersionQueryRequest) ([]*pb.PipelineTemplateVersion, error) {

	if request.Name == "" {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InvalidParameter("name")
	}
	if request.ScopeID == "" {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InvalidParameter("scopeID")
	}
	if request.ScopeType == "" {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InvalidParameter("scopeType")
	}

	dbTemplate, err := s.db.GetPipelineTemplate(request.Name, request.ScopeType, request.ScopeID)
	if err != nil || dbTemplate == nil {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InternalError(err)
	}

	version := db.DicePipelineTemplateVersion{
		TemplateId: dbTemplate.ID,
	}

	dbVersions, err := s.db.QueryPipelineTemplateVersions(&version)
	if err != nil {
		return nil, apierrors.ErrQueryPipelineTemplateVersion.InternalError(err)
	}

	var result []*pb.PipelineTemplateVersion
	for _, v := range dbVersions {
		result = append(result, v.ToApiData())
	}
	return result, nil
}

func (s *templateService) RenderPipelineTemplateRequest(request *pb.PipelineTemplateRenderRequest) (*pb.PipelineTemplateRender, error) {
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

	if request.TemplateVersion == int32(TemplateVersionV2) {
		if request.Alias == "" {
			return nil, apierrors.ErrRenderPipelineTemplate.InvalidParameter("alias")
		}
	}

	getPipelineTemplateRequest := &pb.PipelineTemplateVersionGetRequest{
		Name:      request.Name,
		ScopeType: request.ScopeType,
		ScopeID:   request.ScopeID,
		Version:   request.Version,
	}

	templateVersion, err := s.getPipelineTemplateVersion(getPipelineTemplateRequest)
	if err != nil {
		return nil, apierrors.ErrRenderPipelineTemplate.InternalError(err)
	}

	if templateVersion == nil {
		return nil, apierrors.ErrRenderPipelineTemplate.InternalError(errors.New(fmt.Sprintf(" not find template version. request %v", request)))
	}

	var templateAction pb.PipelineTemplateSpec
	if err := yaml.Unmarshal([]byte(templateVersion.Spec), &templateAction); err != nil {
		logrus.Errorf("Unmarshal specYaml error: %v, yaml: %s", err, templateVersion.Spec)
		return nil, err
	}

	pipelineYaml, outputs, err := DoRenderTemplateWithFormatV2(request.Params, &templateAction, request.Alias, TemplateVersion(request.TemplateVersion))
	if err != nil {
		return nil, apierrors.ErrRenderPipelineTemplate.InternalError(err)
	}

	render := pb.PipelineTemplateRender{
		PipelineYaml: pipelineYaml,
		Version:      templateVersion,
		Outputs:      outputs,
	}

	return &render, nil
}

func (s *templateService) RenderPipelineTemplateBySpecRequest(request *pb.PipelineTemplateRenderSpecRequest) (*pb.PipelineTemplateRender, error) {

	if request.Spec == nil {
		return nil, apierrors.ErrRenderPipelineTemplate.InvalidParameter("spec")
	}
	if request.Alias == "" {
		return nil, apierrors.ErrRenderPipelineTemplate.InvalidParameter("alias")
	}

	params := request.Spec.Params
	if params != nil {
		for _, v := range params {
			if err := CheckPipelineParam(v); err != nil {
				return nil, apierrors.ErrRenderPipelineTemplate.InvalidParameter(fmt.Sprintf("snippet %s definition params name can not empty", request.Spec.Name))
			}
		}
	}

	pipelineYaml, outputs, err := DoRenderTemplateWithFormatV2(request.Params, request.Spec, request.Alias, TemplateVersion(request.TemplateVersion))
	if err != nil {
		return nil, apierrors.ErrRenderPipelineTemplate.InternalError(err)
	}

	render := pb.PipelineTemplateRender{
		PipelineYaml: pipelineYaml,
		Version:      nil,
		Outputs:      outputs,
	}
	return &render, nil
}

func (s *templateService) getPipelineTemplateVersion(request *pb.PipelineTemplateVersionGetRequest) (*pb.PipelineTemplateVersion, error) {
	if request.Name == "" {
		return nil, apierrors.ErrCreateExtension.InvalidParameter("name")
	}
	if request.ScopeID == "" {
		return nil, apierrors.ErrCreateExtension.InvalidParameter("scopeID")
	}
	if request.ScopeType == "" {
		return nil, apierrors.ErrCreateExtension.InvalidParameter("scopeType")
	}
	if request.Version == "" {
		return nil, apierrors.ErrCreateExtension.InvalidParameter("Version")
	}

	dbTemplate, err := s.db.GetPipelineTemplate(request.Name, request.ScopeType, request.ScopeID)
	if err != nil {
		return nil, apierrors.ErrCreateExtension.InternalError(err)
	}

	if dbTemplate == nil {
		return nil, nil
	}

	dbVersion, err := s.db.GetPipelineTemplateVersion(request.Version, dbTemplate.ID)
	if err != nil {
		return nil, apierrors.ErrCreateExtension.InternalError(err)
	}

	if dbVersion == nil {
		return nil, nil
	}

	return dbVersion.ToApiData(), nil
}
