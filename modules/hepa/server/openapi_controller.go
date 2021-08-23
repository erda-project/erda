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
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/hepa/common/util"
	"github.com/erda-project/erda/modules/hepa/gateway/dto"
	"github.com/erda-project/erda/modules/hepa/gateway/exdto"
	"github.com/erda-project/erda/modules/hepa/gateway/service"
	"github.com/erda-project/erda/pkg/discover"
)

type OpenapiController struct {
	api      service.GatewayOpenapiService
	consumer service.GatewayOpenapiConsumerService
	rule     service.GatewayOpenapiRuleService
	runtime  service.GatewayRuntimeServiceService
	domain   service.GatewayDomainService
	client   service.GatewayOrgClientService
	global   service.GatewayGlobalService
}

func NewOpenapiController() (*OpenapiController, error) {
	api, _ := service.NewGatewayOpenapiServiceImpl()
	consumer, _ := service.NewGatewayOpenapiConsumerServiceImpl()
	rule, _ := service.NewGatewayOpenapiRuleServiceImpl()
	runtime, _ := service.NewGatewayRuntimeServiceServiceImpl()
	domain, _ := service.NewGatewayDomainServiceImpl()
	client, _ := service.NewGatewayOrgClientServiceImpl()
	global, _ := service.NewGatewayGlobalServiceImpl()
	return &OpenapiController{
		api:      api,
		consumer: consumer,
		rule:     rule,
		runtime:  runtime,
		domain:   domain,
		client:   client,
		global:   global,
	}, nil
}

func (ctl OpenapiController) Register() {
	BindOpenApi(FEATURES, "GET", ctl.GetGatewayFeatures())
	BindOpenApi(CLIENTS, "POST", ctl.CreateClient())
	BindOpenApi(CLIENT, "DELETE", ctl.DeleteClient())
	BindOpenApi(CLIENTAUTH, "GET", ctl.GetClientCredentials())
	BindOpenApi(CLIENTAUTH, "PATCH", ctl.UpdateClientCredentials())
	BindOpenApi(CLIENTACL, "POST", ctl.GrantClientPackage())
	BindOpenApi(CLIENTACL, "DELETE", ctl.RevokeClientPackage())
	BindOpenApi(CLIENTLIMIT, "PUT", ctl.CreateOrUpdateClientLimits())

	BindOpenApi(PACKAGEROOTAPI, "PUT", ctl.TouchPackageRootApi())

	BindOpenApi(SERVICE_API_PREFIX, "GET", ctl.GetServiceApiPrefix())
	BindOpenApi(CLOUDAPI_INFO, "GET", ctl.GetCloudapiInfo())
	BindOpenApi(PACKAGE_ALIYUN_BIND, "GET", ctl.GetPackageAliyunBind())
	BindOpenApi(PACKAGE_ALIYUN_BIND, "POST", ctl.SetPackageAliyunBind())
	BindOpenApi(CONSUMER_ALIYUN_AUTH, "GET", ctl.GetCloudapiCredential())
	BindOpenApi(CONSUMER_ALIYUN_AUTH, "POST", ctl.SetCloudapiCredential(false))
	BindOpenApi(CONSUMER_ALIYUN_AUTH_ASYNC, "POST", ctl.SetCloudapiCredential(true))
	BindOpenApi(CONSUMER_ALIYUN_AUTH, "DELETE", ctl.DeleteCloudapiCredential())

	BindOpenApi(RUNTIME_DOMAIN, "GET", ctl.GetRuntimeDomains())
	BindOpenApi(RUNTIME_SERVICE_DOMAIN, "PUT", ctl.UpdateRuntimeServiceDomain())
	BindOpenApi(TENANT_DOMAIN, "GET", ctl.GetTenantDomains())
	BindOpenApi(SERVICE_RUNTIME, "GET", ctl.GetServiceRuntimes())

	BindOpenApi(PACKAGES, "POST", ctl.CreatePackage())
	BindOpenApi(PACKAGES, "GET", ctl.GetPackages())

	BindOpenApi(PACKAGE, "GET", ctl.GetPackage())
	BindOpenApi(PACKAGE, "DELETE", ctl.DeletePackage())
	BindOpenApi(PACKAGE, "PATCH", ctl.UpdatePackage())

	BindOpenApi(PACKAGEAPIS, "GET", ctl.GetPackageApis())
	BindOpenApi(PACKAGEAPIS, "POST", ctl.CreatePackageApi())

	BindOpenApi(PACKAGEAPI, "DELETE", ctl.DeletePackageApi())
	BindOpenApi(PACKAGEAPI, "PATCH", ctl.UpdatePackageApi())

	BindOpenApi(PACKAGEACL, "POST", ctl.UpdatePackageAcl())
	BindOpenApi(PACKAGEACL, "GET", ctl.GetPackageAcl())

	BindOpenApi(PACKAGEAPIACL, "POST", ctl.UpdatePackageApiAcl())
	BindOpenApi(PACKAGEAPIACL, "GET", ctl.GetPackageApiAcl())

	BindOpenApi(CONSUMERS, "POST", ctl.CreateConsumer())
	BindOpenApi(CONSUMERS, "GET", ctl.GetConsumers())

	BindOpenApi(CONSUMER, "DELETE", ctl.DeleteConsumer())
	BindOpenApi(CONSUMER, "PATCH", ctl.UpdateConsumer())

	BindOpenApi(CONSUMERACL, "GET", ctl.GetConsumerAcl())
	BindOpenApi(CONSUMERACL, "POST", ctl.UpdateConsumerAcl())

	BindOpenApi(CONSUMERAUTH, "GET", ctl.GetConsumerAuth())
	BindOpenApi(CONSUMERAUTH, "POST", ctl.UpdateConsumerAuth())

	BindOpenApi(LIMITS, "GET", ctl.GetLimits())
	BindOpenApi(LIMITS, "POST", ctl.CreateLimit())

	BindOpenApi(LIMIT, "DELETE", ctl.DeleteLimit())
	BindOpenApi(LIMIT, "PATCH", ctl.UpdateLimit())

	BindOpenApi(PACKAGESNAME, "GET", ctl.GetPackagesName())
	BindOpenApi(CONSUMERSNAME, "GET", ctl.GetConsumersName())

	BindOpenApi(METRICS, "GET", ctl.GetMetrics())
}

func (ctl OpenapiController) GetGatewayFeatures() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.global.GetGatewayFeatures(c.Param("clusterName"))
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

func (ctl OpenapiController) CreateClient() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.client.Create(c.Query("orgId"), c.Query("clientName"))
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

func (ctl OpenapiController) DeleteClient() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.client.Delete(c.Param("clientId"))
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

func (ctl OpenapiController) GetClientCredentials() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.client.GetCredentials(c.Param("clientId"))
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

func (ctl OpenapiController) UpdateClientCredentials() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.client.UpdateCredentials(c.Param("clientId"), c.Query("clientSecret"))
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

func (ctl OpenapiController) GrantClientPackage() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.client.GrantPackage(c.Param("clientId"), c.Param("packageId"))
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

func (ctl OpenapiController) RevokeClientPackage() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.client.RevokePackage(c.Param("clientId"), c.Param("packageId"))
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

func (ctl OpenapiController) GetMetrics() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		if c.Query("projectId") == "" {
			return http.StatusForbidden, []byte("")
		}
		path := c.Request.URL.Path
		path = strings.Replace(path, "/api/gateway/openapi/metrics/charts", "/api/metrics", 1)
		path += "?" + c.Request.URL.RawQuery
		log.Infof("dashboard proxy url:%s", path)
		code, body, err := util.CommonRequest("GET", discover.Monitor()+path, nil)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, []byte("")
		}
		return code, body
	}
}

func (ctl OpenapiController) GetRuntimeDomains() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		runtimeId := c.Param("runtimeId")
		resp := ctl.domain.GetRuntimeDomains(runtimeId)
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

func (ctl OpenapiController) UpdateRuntimeServiceDomain() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		runtimeId := c.Param("runtimeId")
		serviceName := c.Param("serviceName")
		orgId := c.GetHeader("Org-ID")
		reqDto := dto.ServiceDomainReqDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		server, _ := service.NewGatewayDomainServiceImpl()
		server.ReqCtx = c
		resp := server.UpdateRuntimeServiceDomain(orgId, runtimeId, serviceName,
			&reqDto)
		resp.SwitchLang(c)
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

func (ctl OpenapiController) GetTenantDomains() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		projectId := c.Query("projectId")
		env := c.Query("env")
		server, _ := service.NewGatewayDomainServiceImpl()
		server.ReqCtx = c
		resp := server.GetTenantDomains(projectId, env)
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

func (ctl OpenapiController) GetServiceRuntimes() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		projectId := c.Query("projectId")
		env := c.Query("env")
		app := c.Query("app")
		service := c.Query("service")
		resp := ctl.runtime.GetServiceRuntimes(projectId, env, app, service)
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

func (ctl OpenapiController) CreatePackage() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		args := dto.NewDiceArgsDto(c)
		reqDto := dto.PackageDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		server, _ := service.NewGatewayOpenapiServiceImpl()
		server.ReqCtx = c
		resp := server.CreatePackage(&args, &reqDto)
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

func (ctl OpenapiController) GetPackages() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		args := dto.GetPackagesDto{
			DiceArgsDto: dto.NewDiceArgsDto(c),
			Domain:      c.Query("domain"),
		}
		resp := ctl.api.GetPackages(&args)
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

func (ctl OpenapiController) GetPackagesName() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		args := dto.GetPackagesDto{DiceArgsDto: dto.NewDiceArgsDto(c)}
		resp := ctl.api.GetPackagesName(&args)
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

func (ctl OpenapiController) GetPackage() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.api.GetPackage(c.Param("packageId"))
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

func (ctl OpenapiController) DeletePackage() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		server, _ := service.NewGatewayOpenapiServiceImpl()
		server.ReqCtx = c
		resp := server.DeletePackage(c.Param("packageId"))
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

func (ctl OpenapiController) UpdatePackage() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.PackageDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		server, _ := service.NewGatewayOpenapiServiceImpl()
		server.ReqCtx = c
		resp := server.UpdatePackage(c.Param("packageId"), &reqDto)
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

func (ctl OpenapiController) CreatePackageApi() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.OpenapiDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		server, _ := service.NewGatewayOpenapiServiceImpl()
		server.ReqCtx = c
		resp := server.CreatePackageApi(c.Param("packageId"), &reqDto)
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

func (ctl OpenapiController) TouchPackageRootApi() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.OpenapiDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.api.TouchPackageRootApi(c.Param("packageId"), &reqDto)
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

func (ctl OpenapiController) GetPackageApis() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		args := dto.NewGetOpenapiDto(c)
		resp := ctl.api.GetPackageApis(c.Param("packageId"), &args)
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

func (ctl OpenapiController) DeletePackageApi() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		server, _ := service.NewGatewayOpenapiServiceImpl()
		server.ReqCtx = c
		resp := server.DeletePackageApi(c.Param("packageId"), c.Param("apiId"))
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

func (ctl OpenapiController) UpdatePackageApi() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.OpenapiDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		server, _ := service.NewGatewayOpenapiServiceImpl()
		server.ReqCtx = c
		resp := server.UpdatePackageApi(c.Param("packageId"), c.Param("apiId"), &reqDto)
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

func (ctl OpenapiController) GetPackageAcl() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.consumer.GetPackageAcls(c.Param("packageId"))
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

func (ctl OpenapiController) UpdatePackageAcl() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.PackageAclsDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.consumer.UpdatePackageAcls(c.Param("packageId"), &reqDto)
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

func (ctl OpenapiController) GetPackageApiAcl() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.consumer.GetPackageApiAcls(c.Param("packageId"), c.Param("apiId"))
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

func (ctl OpenapiController) UpdatePackageApiAcl() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.PackageAclsDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.consumer.UpdatePackageApiAcls(c.Param("packageId"), c.Param("apiId"), &reqDto)
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

func (ctl OpenapiController) GetConsumers() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		args := dto.GetOpenConsumersDto{DiceArgsDto: dto.NewDiceArgsDto(c)}
		resp := ctl.consumer.GetConsumers(&args)
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

func (ctl OpenapiController) GetConsumersName() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		args := dto.GetOpenConsumersDto{DiceArgsDto: dto.NewDiceArgsDto(c)}
		resp := ctl.consumer.GetConsumersName(&args)
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

func (ctl OpenapiController) CreateConsumer() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		args := dto.NewDiceArgsDto(c)
		reqDto := dto.OpenConsumerDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		server, _ := service.NewGatewayOpenapiConsumerServiceImpl()
		server.ReqCtx = c
		resp := server.CreateConsumer(&args, &reqDto)
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

func (ctl OpenapiController) DeleteConsumer() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		server, _ := service.NewGatewayOpenapiConsumerServiceImpl()
		server.ReqCtx = c
		resp := server.DeleteConsumer(c.Param("consumerId"))
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

func (ctl OpenapiController) UpdateConsumer() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.OpenConsumerDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		server, _ := service.NewGatewayOpenapiConsumerServiceImpl()
		server.ReqCtx = c
		resp := server.UpdateConsumer(c.Param("consumerId"), &reqDto)
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

func (ctl OpenapiController) GetConsumerAcl() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.consumer.GetConsumerAcls(c.Param("consumerId"))
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

func (ctl OpenapiController) UpdateConsumerAcl() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.ConsumerAclsDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.consumer.UpdateConsumerAcls(c.Param("consumerId"), &reqDto)
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

func (ctl OpenapiController) GetConsumerAuth() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.consumer.GetConsumerCredentials(c.Param("consumerId"))
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

func (ctl OpenapiController) UpdateConsumerAuth() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.ConsumerCredentialsDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		server, _ := service.NewGatewayOpenapiConsumerServiceImpl()
		server.ReqCtx = c
		resp := server.UpdateConsumerCredentials(c.Param("consumerId"), &reqDto)
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

func (ctl OpenapiController) GetLimits() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		args := dto.NewGetOpenLimitRulesDto(c)
		resp := ctl.rule.GetLimitRules(&args)
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

func (ctl OpenapiController) CreateLimit() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		args := dto.NewDiceArgsDto(c)
		reqDto := dto.OpenLimitRuleDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(err)
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.rule.CreateLimitRule(&args, &reqDto)
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

func (ctl OpenapiController) DeleteLimit() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.rule.DeleteLimitRule(c.Param("ruleId"))
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

func (ctl OpenapiController) UpdateLimit() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.OpenLimitRuleDto{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(errors.Wrapf(err, "reqBody:%s", reqBody))
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.rule.UpdateLimitRule(c.Param("ruleId"), &reqDto)
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

func (ctl OpenapiController) GetServiceApiPrefix() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := dto.ApiPrefixReqDto{
			OrgId:     c.Query("orgId"),
			ProjectId: c.Query("projectId"),
			Env:       c.Query("env"),
			App:       c.Query("app"),
			Service:   c.Query("service"),
			RuntimeId: c.Query("runtimeId"),
		}
		resp := ctl.runtime.GetServiceApiPrefix(&reqDto)
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

func (ctl OpenapiController) SetPackageAliyunBind() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		orgId := c.GetHeader("Org-ID")
		packageId := c.Param("packageId")
		resp := ctl.api.SetCloudapiGroupBind(orgId, packageId)
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

func (ctl OpenapiController) GetPackageAliyunBind() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.api.GetCloudapiGroupBind(c.Param("packageId"))
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

func (ctl OpenapiController) GetCloudapiInfo() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.api.GetCloudapiInfo(c.Query("projectId"), c.Query("env"))
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

func (ctl OpenapiController) GetCloudapiCredential() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.consumer.GetCloudapiAppCredential(c.Param("consumerId"))
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

func (ctl OpenapiController) SetCloudapiCredential(async bool) Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.consumer.SetCloudapiAppCredential(c.Param("consumerId"), async)
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

func (ctl OpenapiController) DeleteCloudapiCredential() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		resp := ctl.consumer.DeleteCloudapiAppCredential(c.Param("consumerId"))
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

func (ctl OpenapiController) CreateOrUpdateClientLimits() Controller {
	return func(c *gin.Context, reqBody []byte) (int, []byte) {
		reqDto := exdto.ChangeLimitsReq{}
		err := json.Unmarshal(reqBody, &reqDto)
		if err != nil {
			log.Error(errors.Wrapf(err, "reqBody:%s", reqBody))
			return http.StatusBadRequest, []byte("parse request failed")
		}
		resp := ctl.client.CreateOrUpdateLimit(c.Param("clientId"), c.Param("packageId"), reqDto)
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
