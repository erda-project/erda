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
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	commonPb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda-proto-go/core/hepa/endpoint_api/pb"
	projPb "github.com/erda-project/erda-proto-go/core/project/pb"
	runtimePb "github.com/erda-project/erda-proto-go/orchestrator/runtime/pb"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/common/vars"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/k8s"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/kong"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	repositoryService "github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/endpoint_api"
	endpointApiImpl "github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/endpoint_api/impl"
	"github.com/erda-project/erda/pkg/common/apis"
	erdaErr "github.com/erda-project/erda/pkg/common/errors"
)

const (
	invalidReasonProjectIsInvalid           = "package's project is invalid"
	invalidReasonPackageRuntimeIsInvalid    = "package's runtime is invalid"
	invalidReasonPackageAPIRuntimeIsInvalid = "package_api's runtime is invalid"
	invalidReasonInnerAddrIsInvalid         = "package_api's redirect inner address is invalid"

	invalidTypePackage    = "package"
	invalidTypePackageAPI = "package_api"
)

var (
	innerAddrRegexp, _ = regexp.Compile(`^.+\..+\.svc\.cluster\.local$`)
	innerAddrSuffix    = ".svc.cluster.local"
)

// endpointApiService implements pb.EndpointApiServiceServer
type endpointApiService struct {
	projCli               projPb.ProjectServer
	runtimeCli            runtimePb.RuntimeServiceServer
	runtimeService        repositoryService.GatewayRuntimeServiceService
	gatewayRouteService   repositoryService.GatewayRouteService
	gatewayServiceService repositoryService.GatewayServiceService
	kongInfoService       repositoryService.GatewayKongInfoService
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

func (s *endpointApiService) ListInvalidEndpointApi(ctx context.Context, req *pb.ListInvalidEndpointApiReq) (*pb.ListInvalidEndpointApiResp, error) {
	var result pb.ListInvalidEndpointApiResp
	kongInfo, err := s.kongInfoService.GetKongInfo(&orm.GatewayKongInfo{Az: req.ClusterName})
	if err != nil {
		return nil, err
	}
	err = s.rangeInvalidEndpointApi(ctx, req.GetClusterName(), func(item *pb.ListInvalidEndpointApiItem) {
		if item.GetType() == invalidTypePackageAPI && kongInfo != nil && kongInfo.KongAddr != "" {
			if item.GetKongRouteID() != "" {
				item.RouteDeleting = "curl -sIL -w '%{http_code}\n' -X DELETE " + kongInfo.KongAddr + "/routes/" + item.GetKongRouteID()
			}
			if item.GetKongServiceID() != "" {
				item.ServiceDeleting = "curl -sIL -w '%{http_code}\n' -X DELETE " + kongInfo.KongAddr + "/routes/" + item.GetKongServiceID()
			}
		}
		result.Total += 1
		switch item.GetInvalidReason() {
		case invalidReasonProjectIsInvalid:
			result.TotalProjectIsInvalid += 1
		case invalidReasonPackageRuntimeIsInvalid, invalidReasonPackageAPIRuntimeIsInvalid:
			result.TotalRuntimeIsInvalid += 1
		case invalidReasonInnerAddrIsInvalid:
			result.TotalInnerAddrIsInvalid += 1
		}
		result.List = append(result.List, item)
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *endpointApiService) ClearInvalidEndpointApi(ctx context.Context, req *pb.ListInvalidEndpointApiReq) (*commonPb.VoidResponse, error) {
	l := logrus.WithField("func", "clearInvalidEndpointApi")
	service := endpoint_api.Service.Clone(ctx)
	kongInfo, err := s.kongInfoService.GetKongInfo(&orm.GatewayKongInfo{Az: req.ClusterName})
	if err != nil {
		return nil, err
	}
	kongAdapter := kong.NewKongAdapter(kongInfo.KongAddr)
	err = s.rangeInvalidEndpointApi(ctx, req.GetClusterName(), func(item *pb.ListInvalidEndpointApiItem) {
		if item.GetType() == invalidTypePackage {
			l.Infof("delete package: %+v", item)
			if _, err := s.DeleteEndpoint(ctx, &pb.DeleteEndpointRequest{
				PackageId: item.GetPackageID(),
			}); err != nil {
				l.WithError(err).WithField("package id", item.GetPackageID()).Warnln("failed to DeleteEndpoint")
			}
		}
		if item.GetType() == invalidTypePackageAPI {
			l.Infof("delete package api: %+v", item)
			if _, err := s.DeleteEndpointApi(ctx, &pb.DeleteEndpointApiRequest{
				PackageId: item.GetPackageID(),
				ApiId:     item.GetPackageApiID(),
			}); err != nil {
				l.WithError(err).
					WithField("package id", item.GetPackageID()).
					WithField("package api id", item.GetPackageApiID()).
					Warnln("failed to DeleteEndpointApi")
			}
			if item.GetKongRouteID() != "" || item.GetKongServiceID() != "" {
				if err := service.(*endpointApiImpl.GatewayOpenapiServiceImpl).DeleteKongApi(kongAdapter, item.GetPackageApiID()); err != nil {
					l.WithError(err).WithField("packageAPIID", item.GetPackageApiID()).Warnln("failed to DeleteKongApi")
				}
			}
		}
	})
	if err != nil {
		return nil, err
	}

	return new(commonPb.VoidResponse), nil
}

func (s *endpointApiService) rangeInvalidEndpointApi(ctx context.Context, clusterName string, f func(item *pb.ListInvalidEndpointApiItem)) error {
	l := logrus.WithField("func", "rangeInvalidEndpointApi")
	if clusterName == "" {
		return errors.New("invalid clusterName")
	}

	k8SAdapter, err := k8s.NewAdapter(clusterName)
	if err != nil {
		return errors.Wrap(err, "failed to k8s.NewAdapter")
	}
	k8SServices, err := k8SAdapter.ListAllServices("")
	if err != nil {
		return errors.Wrap(err, "failed to k8SAdapter.ListAllServices")
	}
	var k8SServicesAddresses = make(map[string]struct{})
	for _, service := range k8SServices {
		k8SServicesAddresses[service.Name+"."+service.Namespace] = struct{}{}
	}

	// list all packages
	eas := endpoint_api.Service.Clone(ctx)
	packages, err := eas.ListAllPackages()
	if err != nil {
		l.WithError(err).Errorln("failed to ListAllPackages")
		return err
	}
	if len(packages) == 0 {
		l.Warnln("no packages found")
		return nil
	}

	var projectPackages = make(map[string][]orm.GatewayPackage)
	for _, pkg := range packages {
		if pkg.DiceClusterName == clusterName && pkg.DiceProjectId != "" {
			projectPackages[pkg.DiceProjectId] = append(projectPackages[pkg.DiceProjectId], pkg)
		}
	}
	for projectID := range projectPackages {
		l = l.WithField("projectID", projectID)
		id, err := strconv.ParseUint(projectID, 10, 32)
		if err != nil {
			l.WithError(err).Warnln("projectID can not be parsed to uint")
			continue
		}

		packages := projectPackages[projectID]
		// collect the package if it's project is invalid
		resp, err := s.projCli.CheckProjectExist(ctx, &projPb.CheckProjectExistReq{Id: id})
		if err == nil && !resp.GetOk() {
			for _, pkg := range packages {
				packageApis, _ := eas.ListPackageAllApis(pkg.Id)
				for _, packageApi := range packageApis {
					item := s.adjustInvalidPackageAPIItem(pkg, packageApi, invalidReasonProjectIsInvalid)
					f(item)
				}

				item := &pb.ListInvalidEndpointApiItem{
					InvalidReason: invalidReasonProjectIsInvalid,
					Type:          invalidTypePackage,
					ProjectID:     projectID,
					PackageID:     pkg.Id,
				}
				f(item)
			}
			continue
		}

		for _, pkg := range packages {
			l := l.WithField("tb_gateway_package_api.id", pkg.Id)
			if pkg.RuntimeServiceId != "" {
				l := l.WithField("runtimeServiceID", pkg.RuntimeServiceId)
				runtimeService, err := s.runtimeService.Get(pkg.RuntimeServiceId)
				if err != nil {
					l.WithError(err).Errorln("failed to runtimeService.Get")
					return err
				}
				if runtimeService != nil {
					if runtimeID, err := strconv.ParseUint(runtimeService.RuntimeId, 10, 32); err == nil && runtimeID != 0 {
						if resp, err := s.runtimeCli.CheckRuntimeExist(ctx, &runtimePb.CheckRuntimeExistReq{Id: runtimeID}); err != nil && resp != nil && !resp.GetOk() {
							item := &pb.ListInvalidEndpointApiItem{
								InvalidReason: invalidReasonPackageRuntimeIsInvalid,
								Type:          invalidTypePackage,
								ProjectID:     pkg.DiceProjectId,
								PackageID:     pkg.Id,
								PackageApiID:  "",
								RuntimeID:     runtimeService.RuntimeId,
							}
							f(item)
						}
					}
				}
			}

			packageApis, err := eas.ListPackageAllApis(pkg.Id)
			if err != nil {
				l.WithError(err).Warnln("failed to ListPackageAllApis")
				continue
			}
			for _, packageApi := range packageApis {
				if packageApi.RuntimeServiceId != "" {
					l := l.WithField("packageApi.RuntimeServiceId", packageApi.RuntimeServiceId)
					runtimeService, err := s.runtimeService.Get(packageApi.RuntimeServiceId)
					if err != nil {
						l.WithError(err).Errorln("failed to runtimeService.Get")
						return err
					}
					if runtimeService != nil {
						if runtimeID, err := strconv.ParseUint(runtimeService.RuntimeId, 10, 32); err == nil && runtimeID != 0 {
							if resp, err := s.runtimeCli.CheckRuntimeExist(ctx, &runtimePb.CheckRuntimeExistReq{Id: runtimeID}); err != nil && resp != nil && !resp.GetOk() {
								item := s.adjustInvalidPackageAPIItem(pkg, packageApi, invalidReasonPackageAPIRuntimeIsInvalid)
								f(item)
							}
						}
					}
				}

				// if redirect address is inner host, check if it is invalid
				if redirectAddr, err := url.Parse(packageApi.RedirectAddr); err == nil {
					if ok := innerAddrRegexp.MatchString(redirectAddr.Hostname()); ok {
						if _, ok := k8SServicesAddresses[strings.TrimSuffix(redirectAddr.Hostname(), innerAddrSuffix)]; !ok {
							item := s.adjustInvalidPackageAPIItem(pkg, packageApi, invalidReasonInnerAddrIsInvalid)
							f(item)
						}
					}
				}
			}
		}
	}

	return nil
}

func (s *endpointApiService) adjustInvalidPackageAPIItem(pkg orm.GatewayPackage, packageApi orm.GatewayPackageApi, reason string) *pb.ListInvalidEndpointApiItem {
	item := &pb.ListInvalidEndpointApiItem{
		InvalidReason: reason,
		Type:          invalidTypePackageAPI,
		ProjectID:     pkg.DiceProjectId,
		PackageID:     pkg.Id,
		PackageApiID:  packageApi.Id,
		InnerHostname: "",
		KongRouteID:   "",
		KongServiceID: "",
		ClusterName:   pkg.DiceClusterName,
	}
	if redirectAddr, _ := url.Parse(packageApi.RedirectAddr); redirectAddr != nil {
		item.InnerHostname = redirectAddr.Hostname()
	}
	kongRoute, _ := s.gatewayRouteService.GetByApiId(packageApi.Id)
	if kongRoute == nil {
		kongRoute, _ = s.gatewayRouteService.GetByApiId(packageApi.DiceApiId)
	}
	kongService, _ := s.gatewayServiceService.GetByApiId(packageApi.Id)
	if kongService == nil {
		kongService, _ = s.gatewayServiceService.GetByApiId(packageApi.DiceApiId)
	}
	if kongRoute != nil {
		item.KongRouteID = kongRoute.RouteId
	}
	if kongService != nil {
		item.KongServiceID = kongService.ServiceId
	}
	return item
}
