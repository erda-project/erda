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

package endpoint_api

import (
	"context"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	commonPb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda-proto-go/core/hepa/endpoint_api/pb"
	projPb "github.com/erda-project/erda-proto-go/core/services/project/pb"
	runtimePb "github.com/erda-project/erda-proto-go/orchestrator/runtime/pb"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/common/vars"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	repositoryService "github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/endpoint_api"
	"github.com/erda-project/erda/pkg/common/apis"
	erdaErr "github.com/erda-project/erda/pkg/common/errors"
)

type endpointApiService struct {
	projCli            projPb.ProjectServer
	runtimeCli         runtimePb.RuntimeServiceServer
	gatewayApiService  repositoryService.GatewayApiService
	upstreamApiService repositoryService.GatewayUpstreamApiService
	upstreamService    repositoryService.GatewayUpstreamService
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
		OrgId:     apis.GetOrgID(ctx),
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
	ep, err := service.UpdatePackage(apis.GetOrgID(ctx), req.PackageId, dto.FromEndpoint(req.Endpoint))
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

func (s *endpointApiService) ClearInvalidEndpointApi(ctx context.Context, _ *commonPb.VoidRequest) (*commonPb.VoidResponse, error) {
	l := logrus.WithField("func", "*endpointApiService.ClearInvalidEndpointApi")
	// list all packages
	service := endpoint_api.Service.Clone(ctx)
	packages, err := service.ListAllPackages()
	if err != nil {
		l.Warnln("failed to ListAllPackages")
		return nil, err
	}
	if len(packages) == 0 {
		l.Warnln("no packages found")
		return new(commonPb.VoidResponse), nil
	}

	// delete the package if it's project is invalid
	var projectPackages = make(map[string][]orm.GatewayPackage)
	for _, package_ := range packages {
		if package_.DiceProjectId != "" {
			projectPackages[package_.DiceProjectId] = append(projectPackages[package_.DiceProjectId], package_)
		}
	}
	for projectID := range projectPackages {
		l = l.WithField("projectID", projectID)
		id, err := strconv.ParseUint(projectID, 10, 32)
		if err != nil {
			l.WithError(err).Warnln("projectID can not be parsed to uint")
			continue
		}
		projResp, err := s.projCli.GetProjectByID(ctx, &projPb.GetProjectByIDReq{Id: id})
		if err != nil {
			l.WithError(err).Warnln("failed to GetProjectByID")
			continue
		}
		if projResp.Status == nil {
			l.Warnln("invalid response: projResp.Status is nil")
			continue
		}
		if !projResp.Status.Success {
			if projResp.Status.Status == commonPb.StatusEnum_not_found {
				for _, package_ := range projectPackages[projectID] {
					deleteEndpointResp, err := s.DeleteEndpoint(ctx, &pb.DeleteEndpointRequest{PackageId: package_.Id})
					l.WithError(err).WithField("packageID", package_.Id).
						WithField("deleteEndpointResp.Data", deleteEndpointResp.Data).
						Infoln("DeleteEndpoint because that the project dose not exist.")
				}
			}
			continue
		}

		// join on tb_gateway_package_api.dice_api_id = tb_gateway_api.id
		//         tb_gateway_api.upstream_api_id = tb_gateway_upstream_api.id
		//         tb_gateway_upstream_api.upstream_id = tb_gateway_api.id
		//         runtimeID = filepath.base(tb_gateway_api.upstream_name)
		for _, package_ := range projectPackages[projectID] {
			l = l.WithField("tb_gateway_package_api", package_.Id)
			packageApis, err := service.ListPackageAllApis(package_.Id)
			if err != nil {
				l.WithError(err).Warnln("failed to ListPackageAllApis")
				continue
			}
			for _, packageApi := range packageApis {
				l = l.WithField("tb_gateway_package_api.dice_api_id", packageApi.DiceApiId)
				if packageApi.DiceApiId == "" {
					l.Warnln("packageApi.DiceApiId is empty")
					continue
				}
				l = l.WithField("tb_gateway_api.id", packageApi.DiceApiId)
				gatewayApi, err := s.gatewayApiService.GetById(packageApi.DiceApiId)
				if err != nil {
					l.WithError(err).Warnf("failed to gatewayApiService.GetById(%s)", packageApi.DiceApiId)
					continue
				}
				// todo: if gatewayApi.redirect_addr is invalid inner address, delete the package_api
				if gatewayApi.UpstreamApiId == "" {
					l.Warnln("gatewayApi.UpstreamApiID is empty")
					continue
				}
				l = l.WithField("tb_gateway_upstream_api.id", gatewayApi.UpstreamApiId)
				upstreamApi, err := s.upstreamApiService.GetById(gatewayApi.UpstreamApiId)
				if err != nil {
					l.WithError(err).Warnf("failed to upstreamApiService.GetById(%s)", gatewayApi.UpstreamApiId)
					continue
				}
				l = l.WithField("tb_gateway_upstream.id", gatewayApi.Id)
				var cond orm.GatewayUpstream
				cond.Id = upstreamApi.UpstreamId
				upstreams, err := s.upstreamService.SelectByAny(&cond)
				if err != nil {
					l.WithError(err).Warnf("failed to upstreamService.SelectByAny(%+v)", cond)
					continue
				}
				if len(upstreams) == 0 {
					l.Warnln("not found any upstreams")
					continue
				}
				upstreamName := upstreams[0].UpstreamName
				l = l.WithField("upstreamName", upstreamName)
				runtimeID, err := strconv.ParseUint(filepath.Base(upstreamName), 10, 32)
				if err != nil || runtimeID == 0 {
					l.WithError(err).Warnln("failed to parse runtime id from upstream name")
					continue
				}
				l = l.WithField("runtimeID", runtimeID)
				runtimeResp, err := s.runtimeCli.GetRuntimeByID(ctx, &runtimePb.GetRuntimeByIDReq{Id: runtimeID})
				if err != nil {
					l.WithError(err).Warnln("failed to GetRuntimeByID")
					continue
				}
				if !runtimeResp.Status.GetSuccess() && runtimeResp.GetStatus().GetStatus() == commonPb.StatusEnum_not_found {
					if _, err := s.DeleteEndpointApi(ctx, &pb.DeleteEndpointApiRequest{ApiId: packageApi.Id}); err != nil {
						l.WithError(err).Warnln("failed to DeleteEndpointApi")
					}
					continue
				}
			}
		}
	}
	return new(commonPb.VoidResponse), nil
}
