// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package endpoint_api

import (
	context "context"

	"github.com/pkg/errors"

	pb "github.com/erda-project/erda-proto-go/core/hepa/endpoint_api/pb"
	"github.com/erda-project/erda/modules/hepa/common/vars"
	"github.com/erda-project/erda/modules/hepa/gateway/dto"
	"github.com/erda-project/erda/modules/hepa/services/endpoint_api"
	erdaErr "github.com/erda-project/erda/pkg/common/errors"
)

type endpointApiService struct {
	p *provider
}

func (s *endpointApiService) GetEndpointsName(ctx context.Context, req *pb.GetEndpointsNameRequest) (resp *pb.GetEndpointsNameResponse, err error) {
	service := endpoint_api.Service.Clone(ctx)
	reqDto := &dto.GetPackagesDto{}
	reqDto.ProjectId = req.ProjectId
	reqDto.Env = req.Env
	endpointDtos, err := service.GetPackagesName(reqDto)
	if err != nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, errors.Cause(err).Error())
		return
	}
	endpoints := []*pb.Endpoint{}
	for _, ep := range endpointDtos {
		endpoints = append(endpoints, &pb.Endpoint{
			Id:          ep.Id,
			CreateAt:    ep.CreateAt,
			Name:        ep.Name,
			BindDomain:  ep.BindDomain,
			AuthType:    ep.AuthType,
			AclType:     ep.AclType,
			Scene:       ep.Scene,
			Description: ep.Description,
		})
	}
	resp = &pb.GetEndpointsNameResponse{
		Data: endpoints,
	}
	return
}

func (s *endpointApiService) GetEndpoints(ctx context.Context, req *pb.GetEndpointsRequest) (resp *pb.GetEndpointsResponse, err error) {
	service := endpoint_api.Service.Clone(ctx)
	reqDto := &dto.GetPackagesDto{
		DiceArgsDto: dto.DiceArgsDto{
			ProjectId: req.ProjectId,
			Env:       req.Env,
			PageNo:    req.PageNo,
			PageSize:  req.PageSize,
			SortField: req.SortField,
			SortType:  req.SortType,
		},
		Domain: req.Domain,
	}
	if reqDto.PageSize == 0 {
		reqDto.PageSize = 20
	}
	if reqDto.PageNo == 0 {
		reqDto.PageNo = 1
	}
	pageQuery, err := service.GetPackages(reqDto)
	if err != nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, errors.Cause(err).Error())
		return
	}
	resp = &pb.GetEndpointsResponse{
		Data: pageQuery.ToPbPage(),
	}
	return
}
func (s *endpointApiService) GetEndpoint(ctx context.Context, req *pb.GetEndpointRequest) (resp *pb.GetEndpointResponse, err error) {
	service := endpoint_api.Service.Clone(ctx)
	ep, err := service.GetPackage(req.PackageId)
	if err != nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, errors.Cause(err).Error())
		return
	}
	resp = &pb.GetEndpointResponse{
		Data: ep.ToEndpoint(),
	}
	return
}
func (s *endpointApiService) CreateEndpoint(ctx context.Context, req *pb.CreateEndpointRequest) (resp *pb.CreateEndpointResponse, err error) {
	service := endpoint_api.Service.Clone(ctx)
	if req.Endpoint == nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, "endpoint is empty")
		return
	}
	ep, existName, err := service.CreatePackage(&dto.DiceArgsDto{
		ProjectId: req.ProjectId,
		Env:       req.Env,
	}, dto.FromEndpoint(req.Endpoint))
	if existName != "" {
		err = erdaErr.NewAlreadyExistsError(existName)
		return
	}
	if err != nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, errors.Cause(err).Error())
		return
	}
	resp = &pb.CreateEndpointResponse{
		Data: ep.ToEndpoint(),
	}
	return
}
func (s *endpointApiService) UpdateEndpoint(ctx context.Context, req *pb.UpdateEndpointRequest) (resp *pb.UpdateEndpointResponse, err error) {
	service := endpoint_api.Service.Clone(ctx)
	if req.Endpoint == nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, "endpoint is empty")
		return
	}
	ep, err := service.UpdatePackage(req.PackageId, dto.FromEndpoint(req.Endpoint))
	if err != nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, errors.Cause(err).Error())
		return
	}
	resp = &pb.UpdateEndpointResponse{
		Data: ep.ToEndpoint(),
	}
	return
}
func (s *endpointApiService) DeleteEndpoint(ctx context.Context, req *pb.DeleteEndpointRequest) (resp *pb.DeleteEndpointResponse, err error) {
	service := endpoint_api.Service.Clone(ctx)
	result, err := service.DeletePackage(req.PackageId)
	if err != nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, errors.Cause(err).Error())
		return
	}
	resp = &pb.DeleteEndpointResponse{
		Data: result,
	}
	return
}
func (s *endpointApiService) GetEndpointApis(ctx context.Context, req *pb.GetEndpointApisRequest) (resp *pb.GetEndpointApisResponse, err error) {
	service := endpoint_api.Service.Clone(ctx)
	reqDto := &dto.GetOpenapiDto{}
	reqDto.ApiPath = req.ApiPath
	reqDto.DiceApp = req.DiceApp
	reqDto.DiceService = req.DiceService
	reqDto.Method = req.Method
	reqDto.Origin = req.Origin
	reqDto.PageNo = req.PageNo
	reqDto.PageSize = req.PageSize
	reqDto.SortField = req.SortField
	reqDto.SortType = req.SortType
	if reqDto.PageNo == 0 {
		reqDto.PageNo = 1
	}
	if reqDto.PageSize == 0 {
		reqDto.PageSize = 20
	}
	pageQuery, err := service.GetPackageApis(req.PackageId, reqDto)
	if err != nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, errors.Cause(err).Error())
		return
	}
	resp = &pb.GetEndpointApisResponse{
		Data: pageQuery.ToPbPage(),
	}
	return
}
func (s *endpointApiService) CreateEndpointApi(ctx context.Context, req *pb.CreateEndpointApiRequest) (resp *pb.CreateEndpointApiResponse, err error) {
	service := endpoint_api.Service.Clone(ctx)
	if req.EndpointApi == nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, "endpoint api is empty")
		return
	}
	result, exist, err := service.CreatePackageApi(req.PackageId, dto.FromEndpointApi(req.EndpointApi))
	if exist {
		err = erdaErr.NewAlreadyExistsError("api")
		return
	}
	if err != nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, errors.Cause(err).Error())
		return
	}
	resp = &pb.CreateEndpointApiResponse{
		Data: result,
	}
	return
}
func (s *endpointApiService) UpdateEndpointApi(ctx context.Context, req *pb.UpdateEndpointApiRequest) (resp *pb.UpdateEndpointApiResponse, err error) {
	service := endpoint_api.Service.Clone(ctx)
	if req.EndpointApi == nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, "endpoint api is empty")
		return
	}
	result, exist, err := service.UpdatePackageApi(req.PackageId, req.ApiId, dto.FromEndpointApi(req.EndpointApi))
	if exist {
		err = erdaErr.NewAlreadyExistsError("api")
		return
	}
	if err != nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, errors.Cause(err).Error())
		return
	}
	resp = &pb.UpdateEndpointApiResponse{
		Data: result.ToEndpointApi(),
	}
	return
}
func (s *endpointApiService) DeleteEndpointApi(ctx context.Context, req *pb.DeleteEndpointApiRequest) (resp *pb.DeleteEndpointApiResponse, err error) {
	service := endpoint_api.Service.Clone(ctx)
	result, err := service.DeletePackageApi(req.PackageId, req.ApiId)
	if err != nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, errors.Cause(err).Error())
		return
	}
	resp = &pb.DeleteEndpointApiResponse{
		Data: result,
	}
	return
}
func (s *endpointApiService) ChangeEndpointRoot(ctx context.Context, req *pb.ChangeEndpointRootRequest) (resp *pb.ChangeEndpointRootResponse, err error) {
	service := endpoint_api.Service.Clone(ctx)
	if req.EndpointApi == nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, "endpoint api is empty")
		return
	}
	result, err := service.TouchPackageRootApi(req.PackageId, dto.FromEndpointApi(req.EndpointApi))
	if err != nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, errors.Cause(err).Error())
		return
	}
	resp = &pb.ChangeEndpointRootResponse{
		Data: result,
	}
	return
}
