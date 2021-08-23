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

package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/hepa/common"
	"github.com/erda-project/erda/modules/hepa/gateway/dto"
	"github.com/erda-project/erda/modules/hepa/gateway/service"
)

type GatewayController struct {
	apiService         service.GatewayApiService
	consumerService    service.GatewayConsumerService
	categoryService    service.GatewayCategoryService
	consumerApiService service.GatewayConsumerApiService
	mockService        service.GatewayMockService
	upstreamService    service.GatewayUpstreamService
	upstreamLbService  service.GatewayUpstreamLbService
	globalService      service.GatewayGlobalService
	zoneService        service.GatewayZoneService
	apiPolicyService   service.GatewayApiPolicyService
	runtimeService     service.GatewayRuntimeServiceService
	domainService      service.GatewayDomainService
}

func NewGatewayController() (*GatewayController, error) {
	apiService, _ := service.NewGatewayApiServiceImpl()
	consumerService, _ := service.NewGatewayConsumerServiceImpl()
	categoryService, _ := service.NewGatewayCategoryServiceImpl()
	consumerApiService, _ := service.NewGatewayConsumerApiServiceImpl()
	mockService, _ := service.NewGatewayMockServiceImpl()
	upstreamService, _ := service.NewGatewayUpstreamServiceImpl(consumerService, apiService)
	upstreamLbService, _ := service.NewGatewayUpstreamLbServiceImpl()
	globalService, _ := service.NewGatewayGlobalServiceImpl()
	zoneService, _ := service.NewGatewayZoneServiceImpl()
	apiPolicyService, _ := service.NewGatewayApiPolicyServiceImpl()
	runtimeService, _ := service.NewGatewayRuntimeServiceServiceImpl()
	domainService, _ := service.NewGatewayDomainServiceImpl()
	return &GatewayController{
		apiService:         apiService,
		consumerService:    consumerService,
		categoryService:    categoryService,
		consumerApiService: consumerApiService,
		mockService:        mockService,
		upstreamService:    upstreamService,
		upstreamLbService:  upstreamLbService,
		globalService:      globalService,
		zoneService:        zoneService,
		apiPolicyService:   apiPolicyService,
		runtimeService:     runtimeService,
		domainService:      domainService,
	}, nil
}

func (ctl GatewayController) Register() {
	BindRawApi(DICE_HEALTH, "GET", ctl.GetDiceHealth())
	BindApi(TENANT_GROUP, "GET", ctl.GetTenantGroup())
	BindApi(COMPONENT_INGRESS, "PUT", ctl.CreateOrUpdateComponentIngress())
	BindApi(RUNTIME_SERVICE, "PUT", ctl.TouchRuntime(), ctl.TouchRuntimeComplete())
	BindApi(RUNTIME_SERVICE_DELETE, "DELETE", ctl.DeleteRuntime())
	BindApi(TENANTS, "POST", ctl.CreateTenant())
	BindApi(GATEWAY_APP_LIST, "GET", ctl.GetRegisterApps())
	//	BindApi(GATEWAY_UI_TYPE, "GET", ctl.GetClusterUIType())
	BindApi(API_GATEWAY_API, "GET", ctl.GetApiInfos())
	BindApi(API_GATEWAY_API, "POST", ctl.CreateApi())
	BindApi(API_GATEWAY_API_ID, "DELETE", ctl.DeleteApi())
	BindApi(API_GATEWAY_API_ID, "PATCH", ctl.UpdateApi())
	BindApi(API_GATEWAY_CATEGORY, "PUT", ctl.SetCategoryInfo())
	BindApi(API_GATEWAY_CATEGORY, "GET", ctl.GetCategoryInfo())
	BindApi(API_GATEWAY_CATEGORY, "POST", ctl.CreatePolicy())
	BindApi(API_GATEWAY_CATEGORY_ID, "PATCH", ctl.UpdatePolicy())
	BindApi(API_GATEWAY_CATEGORY_ID, "DELETE", ctl.DeletePolicy())
	BindApi(GATEWAY_PROJECT_CONSUMER_INFO, "GET", ctl.GetProjectConsumerInfo())
	BindApi(GATEWAY_CONSUMER_LIST, "GET", ctl.GetConsumerList())
	BindApi(GATEWAY_CONSUMER_CREATE, "POST", ctl.CreateConsumer())
	BindApi(GATEWAY_CONSUMER_API_EDIT, "PATCH", ctl.EditConsumerApi())
	BindApi(GATEWAY_CONSUMER_DELETE, "DELETE", ctl.DeleteConsumer())
	BindApi(GATEWAY_CONSUMER_INFO, "GET", ctl.GetConsumer())
	BindApi(GATEWAY_CONSUMER_UPDATE, "PATCH", ctl.UpdateConsumer())
	BindApi(GATEWAY_CONSUMER_API_INFO, "PATCH", ctl.UpdateConsumerApi())

	BindApi(API_MOCK_REGISTER, "POST", ctl.RegisterMockApi())
	BindApi(API_MOCK_CALL, "POST", ctl.CallMockApi())
	BindApi(UPSTREAM_REGISTER, "PUT", ctl.UpstreamRegister())
	BindApi(UPSTREAM_REGISTER_ASYNC, "PUT", ctl.UpstreamRegisterAsync(),
		ctl.UpstreamValidAsync())
	BindApi(UPSTREAM_TARGET_ONLINE, "PUT", ctl.UpstreamTargetOnline())
	BindApi(UPSTREAM_TARGET_OFFLINE, "PUT", ctl.UpstreamTargetOffline())
	BindApi(HEALTH_CHECK, "GET", ctl.HealthCheck())

	BindApi(DOMAINS, "GET", ctl.GetDomains())
}

func (ctl GatewayController) GetDomains() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.domainService.GetOrgDomainInfo(dto.NewDiceArgsDto(c), &dto.ManageDomainReq{
			Domain:      c.Query("domain"),
			ClusterName: c.Query("clusterName"),
			Type:        dto.DomainType(c.Query("type")),
		})
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) CreateOrUpdateComponentIngress() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := apistructs.ComponentIngressUpdateRequest{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.domainService.CreateOrUpdateComponentIngress(reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) TouchRuntime() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.RuntimeServiceReqDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.runtimeService.TouchRuntime(c, &reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) TouchRuntimeComplete() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.RuntimeServiceReqDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.runtimeService.TouchRuntimeComplete(c, &reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) DeleteRuntime() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		runtimeId := c.Param("runtimeId")
		resp := ctl.runtimeService.DeleteRuntime(runtimeId)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) GetTenantGroup() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		projectId := c.Query("projectId")
		env := c.Query("env")
		resp := ctl.globalService.GetTenantGroup(projectId, env)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) CreateTenant() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.TenantDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.globalService.CreateTenant(&reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) GetRegisterApps() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		projectId := c.Query("projectId")
		env := c.Query("env")
		resp := ctl.runtimeService.GetRegisterAppInfo(projectId, env)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) GetClusterUIType() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		orgId := c.Query("orgId")
		projectId := c.Query("projectId")
		env := c.Query("env")
		resp := ctl.globalService.GetClusterUIType(orgId, projectId, env)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) UpstreamTargetOnline() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.UpstreamLbDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		if reqDto.IsInvalid() {
			log.Errorf("invalid req:%+v", reqDto)
			return http.StatusBadRequest, []byte("invalid req")
		}
		resp := ctl.upstreamLbService.UpstreamTargetOnline(&reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) UpstreamTargetOffline() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.UpstreamLbDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		if reqDto.IsInvalid() {
			log.Errorf("invalid req:%+v", reqDto)
			return http.StatusBadRequest, []byte("invalid req")
		}
		resp := ctl.upstreamLbService.UpstreamTargetOffline(&reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) UpstreamValidAsync() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.UpstreamRegisterDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		if !reqDto.Init() {
			log.Errorf("invalid dto:%+v", reqDto)
			return http.StatusBadRequest, []byte("invalid request")
		}
		for i := 0; i < len(reqDto.ApiList); i++ {
			apiDto := &reqDto.ApiList[i]
			if !apiDto.Init() {
				log.Errorf("invalid api:%+v", *apiDto)
				return http.StatusBadRequest, []byte("invalid request")
			}
		}
		resp := ctl.upstreamService.UpstreamValidAsync(c, &reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) UpstreamRegisterAsync() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.UpstreamRegisterDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		if !reqDto.Init() {
			log.Errorf("invalid dto:%+v", reqDto)
			return http.StatusBadRequest, []byte("invalid request")
		}
		for i := 0; i < len(reqDto.ApiList); i++ {
			apiDto := &reqDto.ApiList[i]
			if !apiDto.Init() {
				log.Errorf("invalid api:%+v", *apiDto)
				return http.StatusBadRequest, []byte("invalid request")
			}
		}
		resp := ctl.upstreamService.UpstreamRegisterAsync(c, &reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) UpstreamRegister() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.UpstreamRegisterDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		if !reqDto.Init() {
			log.Errorf("invalid dto:%+v", reqDto)
			return http.StatusBadRequest, []byte("invalid request")
		}
		for i := 0; i < len(reqDto.ApiList); i++ {
			apiDto := &reqDto.ApiList[i]
			if !apiDto.Init() {
				log.Errorf("invalid api:%+v", *apiDto)
				return http.StatusBadRequest, []byte("invalid request")
			}
		}
		resp := ctl.upstreamService.UpstreamRegister(&reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) RegisterMockApi() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.MockInfoDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.mockService.RegisterMockApi(&reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) CallMockApi() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.CallMockDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.mockService.CallMockApi(reqDto.HeadKey, reqDto.PathUrl, reqDto.Method)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) CreateConsumer() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.ConsumerCreateDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.consumerService.CreateConsumer(&reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) EditConsumerApi() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.ConsumerEditDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.consumerService.UpdateConsumerApi(&reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) DeleteConsumer() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		consumerId := c.Param("consumerId")
		resp := ctl.consumerService.DeleteConsumer(consumerId)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) UpdateConsumerApi() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.ConsumerApiReqDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.consumerApiService.UpdateConsumerApi(&reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) GetProjectConsumerInfo() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		orgId := c.Query("orgId")
		projectId := c.Query("projectId")
		env := c.Query("env")
		resp := ctl.consumerService.GetProjectConsumerInfo(orgId, projectId, env)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) GetConsumer() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.consumerService.GetConsumerInfo(c.Param("consumerId"))
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) UpdateConsumer() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.ConsumerDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.consumerService.UpdateConsumerInfo(c.Param("consumerId"), &reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) GetConsumerList() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		orgId := c.Query("orgId")
		projectId := c.Query("projectId")
		env := c.Query("env")
		resp := ctl.consumerService.GetConsumerList(orgId, projectId, env)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) CreateApi() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.ApiReqDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		server, _ := service.NewGatewayApiServiceImpl()
		server.ReqCtx = c
		resp := server.CreateApi(&reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) GetApiInfos() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil {
			log.Warnf("atoi failed page[%s]", c.Query("page"))
			page = 1
		}
		size, err := strconv.Atoi(c.DefaultQuery("size", "20"))
		if err != nil {
			log.Warnf("atoi failed size[%s]", c.Query("size"))
			size = 20
		}
		needAuth, err := strconv.Atoi(c.DefaultQuery("needAuth", "0"))
		if err != nil {
			log.Warnf("atoi failed needAuth[%s]", c.Query("needAuth"))
			needAuth = 0
		}
		getDto := &dto.GetApisDto{
			From:         c.Query("from"),
			Method:       c.Query("method"),
			DiceApp:      c.Query("diceApp"),
			DiceService:  c.Query("diceService"),
			RuntimeId:    c.Query("runtimeId"),
			ApiPath:      c.Query("apiPath"),
			RegisterType: c.Query("registerType"),
			NetType:      c.Query("netType"),
			NeedAuth:     needAuth,
			SortField:    c.Query("sortField"),
			SortType:     c.Query("sortType"),
			OrgId:        c.Query("orgId"),
			ProjectId:    c.Query("projectId"),
			Env:          c.Query("env"),
			Size:         int64(size),
			Page:         int64(page),
		}
		resp := ctl.apiService.GetApiInfos(getDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) DeleteApi() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		apiId := c.Param("apiId")
		server, _ := service.NewGatewayApiServiceImpl()
		server.ReqCtx = c
		resp := server.DeleteApi(apiId)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) UpdateApi() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		apiId := c.Param("apiId")
		reqDto := dto.ApiReqDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		reqDto.From = c.Query("from")
		server, _ := service.NewGatewayApiServiceImpl()
		server.ReqCtx = c
		resp := server.UpdateApi(apiId, &reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) GetCategoryInfo() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		category := c.Param("category")
		orgId := c.Query("orgId")
		projectId := c.Query("projectId")
		env := c.Query("env")
		var resp *common.StandardResult
		if category == "auth" || category == "trafficControl" {
			resp = ctl.categoryService.GetCategoryInfo(category, orgId, projectId, env)
		} else {
			packageId := c.Query("packageId")
			packageApiId := c.Query("apiId")
			resp = ctl.apiPolicyService.GetPolicyConfig(category, packageId, packageApiId)
		}
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) SetCategoryInfo() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		category := c.Param("category")
		packageId := c.Query("packageId")
		packageApiId := c.Query("apiId")
		server, _ := service.NewGatewayApiPolicyServiceImpl()
		server.ReqCtx = c
		resp := server.SetPolicyConfig(category, packageId, packageApiId, reqBody)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) CreatePolicy() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		category := c.Param("category")
		reqDto := dto.PolicyCreateDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.categoryService.CreatePolicy(category, &reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) UpdatePolicy() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		category := c.Param("category")
		policyId := c.Param("policyId")
		reqDto := dto.PolicyCreateDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.categoryService.UpdatePolicy(policyId, category, &reqDto)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) DeletePolicy() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		policyId := c.Param("policyId")
		resp := ctl.categoryService.DeletePolicy(policyId)
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) HealthCheck() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := common.StandardResult{Success: true}
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		if !resp.Success {
			return http.StatusBadRequest, respJson
		}
		return http.StatusOK, respJson
	}
}

func (ctl GatewayController) GetDiceHealth() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.globalService.GetDiceHealth()
		respJson, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("encode response failed")
		}
		return http.StatusOK, respJson
	}
}
