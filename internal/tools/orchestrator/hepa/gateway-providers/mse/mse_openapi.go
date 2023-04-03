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

package mse

import (
	"strings"

	mseopenapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	mseclient "github.com/alibabacloud-go/mse-20190531/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/pkg/errors"

	. "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	mseCommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	mseplugins "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/plugins"
)

// CreateMSEClientByAPI 创建客户端
func (impl *MseAdapterImpl) CreateMSEClientByAPI() (client *mseclient.Client, err error) {
	if impl.AccessKeyID == "" || impl.AccessKeySecret == "" {
		return nil, errors.Errorf("Aliyun AccessKeyId or AccessKeySecret not set, please set them in env by ALIYUN_ACCESS_KEY_ID and ALIYUN_ACCESS_KEY_SECRET.")
	}

	if impl.GatewayEndpoint == "" {
		return nil, errors.Errorf("Aliyun Mse Gateway UniqueId not set, please set it in env by ALIYUN_MSE_GATEWAY_DOMAIN.")
	}

	config := &mseopenapi.Config{
		// 必填，您的 AccessKey ID
		AccessKeyId: &impl.AccessKeyID,
		// 必填，您的 AccessKey Secret
		AccessKeySecret: &impl.AccessKeySecret,
	}
	// MSE Gateway Endpoint
	config.Endpoint = &impl.GatewayEndpoint
	client = &mseclient.Client{}
	client, err = mseclient.NewClient(config)
	return client, err
}

// GetMSEGatewayByAPI 获取网关详情
func (impl *MseAdapterImpl) GetMSEGatewayByAPI() (*mseclient.GetGatewayResponseBodyData, error) {
	if impl.GatewayUniqueID == "" {
		return nil, errors.Errorf("Aliyun Mse Gateway UniqueId not set, please set it in env by ALIYUN_MSE_GATEWAY_ID.")
	}

	client, err := impl.CreateMSEClientByAPI()
	if err != nil {
		return nil, err
	}

	getGatewayRequest := &mseclient.GetGatewayRequest{
		GatewayUniqueId: &impl.GatewayUniqueID,
	}
	runtime := &util.RuntimeOptions{}

	gatewayResponse, err := client.GetGatewayWithOptions(getGatewayRequest, runtime)
	if err != nil {
		return nil, err
	}

	if gatewayResponse.Body == nil || !(*gatewayResponse.Body.Success) {
		return nil, errors.Errorf("GetGatewayWithOptions get nil reponse body or request failed, message: %v", *gatewayResponse.Body.Message)
	}

	return gatewayResponse.Body.Data, nil
}

// GetMSEPluginsByAPI 获取网关插件列表
func (impl *MseAdapterImpl) GetMSEPluginsByAPI(name *string, category *int32, enableOnly *bool) ([]*mseclient.GetPluginsResponseBodyData, error) {
	if impl.GatewayUniqueID == "" {
		return nil, errors.Errorf("Aliyun Mse Gateway UniqueId not set, please set it in env by ALIYUN_MSE_GATEWAY_ID.")
	}

	client, err := impl.CreateMSEClientByAPI()
	if err != nil {
		return nil, err
	}

	getPluginsRequest := &mseclient.GetPluginsRequest{
		GatewayUniqueId: &impl.GatewayUniqueID,
		Name:            name,
		Category:        category,
		EnableOnly:      enableOnly,
	}
	runtime := &util.RuntimeOptions{}

	gatewayPluginsResponse, err := client.GetPluginsWithOptions(getPluginsRequest, runtime)
	if err != nil {
		return nil, err
	}
	if gatewayPluginsResponse.Body == nil || !(*gatewayPluginsResponse.Body.Success) {
		return nil, errors.Errorf("GetPluginsWithOptions get nil reponse body or request failed, message: %v", *gatewayPluginsResponse.Body.Message)
	}

	return gatewayPluginsResponse.Body.Data, nil
}

// GetMSEPluginConfigByIDByAPI 获取网关指定 ID 的插件的配置信息
func (impl *MseAdapterImpl) GetMSEPluginConfigByIDByAPI(pluginId *int64) (*mseclient.GetPluginConfigResponseBodyData, error) {
	if impl.GatewayUniqueID == "" {
		return nil, errors.Errorf("Aliyun Mse Gateway UniqueId not set, please set it in env by ALIYUN_MSE_GATEWAY_ID.")
	}
	if pluginId == nil {
		return nil, errors.Errorf("need set pluginId")
	}

	client, err := impl.CreateMSEClientByAPI()
	if err != nil {
		return nil, err
	}

	getPluginConfigRequest := &mseclient.GetPluginConfigRequest{
		GatewayUniqueId: &impl.GatewayUniqueID,
		PluginId:        pluginId,
	}
	runtime := &util.RuntimeOptions{}

	getPluginConfigResponse, err := client.GetPluginConfigWithOptions(getPluginConfigRequest, runtime)
	if err != nil {
		return nil, err
	}
	if getPluginConfigResponse.Body == nil || !(*getPluginConfigResponse.Body.Success) {
		return nil, errors.Errorf("GetPluginConfigWithOptions get nil reponse body or request failed, message: %v", *getPluginConfigResponse.Body.Message)
	}

	return getPluginConfigResponse.Body.Data, nil
}

// UpdateMSEPluginConfigByIDByAPI 获取网关指定 ID 的插件的配置信息
func (impl *MseAdapterImpl) UpdateMSEPluginConfigByIDByAPI(pluginID *int64, configID *int64, config *string, configLevel *int32, enable *bool) (*mseclient.UpdatePluginConfigResponseBody, error) {
	if impl.GatewayUniqueID == "" {
		return nil, errors.Errorf("Aliyun Mse Gateway UniqueId not set, please set it in env by ALIYUN_MSE_GATEWAY_ID.")
	}
	if pluginID == nil {
		return nil, errors.Errorf("need set pluginId")
	}

	client, err := impl.CreateMSEClientByAPI()
	if err != nil {
		return nil, err
	}

	updatePluginConfigRequest := &mseclient.UpdatePluginConfigRequest{
		GatewayUniqueId: &impl.GatewayUniqueID,
		PluginId:        pluginID,
		Config:          config,
		ConfigLevel:     configLevel,
		Enable:          enable,
		Id:              configID,
	}

	runtime := &util.RuntimeOptions{}

	updatePluginConfigResponse, err := client.UpdatePluginConfigWithOptions(updatePluginConfigRequest, runtime)
	if err != nil {
		return nil, err
	}
	if updatePluginConfigResponse.Body == nil {
		return nil, errors.Errorf("UpdatePluginConfigWithOptions get nil reponse body")
	}

	if !*(updatePluginConfigResponse.Body.Success) {
		return nil, errors.Errorf("UpdatePluginConfigWithOptions failed with message: %v", *updatePluginConfigResponse.Body.Message)
	}

	return updatePluginConfigResponse.Body, nil
}

// ListMSEGatewayRoutesByAPI 获取指定网关下的路由列表，支持按域名筛选
func (impl *MseAdapterImpl) ListMSEGatewayRoutesByAPI(domainName *string, pageNumber *int32, pageSize *int32) (*mseclient.ListGatewayRouteResponseBody, error) {
	if impl.GatewayUniqueID == "" {
		return nil, errors.Errorf("Aliyun Mse Gateway UniqueId not set, please set it in env by ALIYUN_MSE_GATEWAY_ID.")
	}

	client, err := impl.CreateMSEClientByAPI()
	if err != nil {
		return nil, err
	}

	gatewayUniqueId := impl.GatewayUniqueID
	filters := &mseclient.ListGatewayRouteRequestFilterParams{
		DomainName:      domainName,
		GatewayUniqueId: &gatewayUniqueId,
		//RouteOrder:       nil,
		//Status:           nil,
	}
	listGatewayRouteRequest := &mseclient.ListGatewayRouteRequest{
		FilterParams: filters,
		//OrderItem:    orderItem,
		PageNumber: pageNumber,
		PageSize:   pageSize,
	}
	runtime := &util.RuntimeOptions{}
	listGatewayRouteResponse, err := client.ListGatewayRouteWithOptions(listGatewayRouteRequest, runtime)
	if err != nil {
		return nil, err
	}

	if listGatewayRouteResponse.Body == nil || !(*listGatewayRouteResponse.Body.Success) {
		return nil, errors.Errorf("GetPluginConfigWithOptions get nil reponse body or request failed, message: %v", *listGatewayRouteResponse.Body.Message)
	}

	return listGatewayRouteResponse.Body, nil
}

// GetMSEGatewayRouteNameByZoneName 获取指定网关下的关联到 IngressName 的网关路由
// 不通过官方的 GetGatewayRouteDetail API 直接获取的原因是 RouteId 参数无法获取， ListGatewayRoute 返回的路由列表里，每个路由并不带 ID 信息，只有 Name
func (impl *MseAdapterImpl) GetMSEGatewayRouteNameByZoneName(zoneName string, domainName *string) (string, error) {
	if zoneName == "" {
		return "", nil
	}

	var pageNumber int32 = 0
	var pageSize int32 = 1500

	listRespBody, err := impl.ListMSEGatewayRoutesByAPI(domainName, &pageNumber, &pageSize)
	if err != nil {
		return "", err
	}

	if listRespBody.Data != nil {
		for _, route := range listRespBody.Data.Result {
			if strings.Contains(*route.Name, zoneName) {
				return *route.Name, nil
			}
		}
	}

	for i := 1; i <= int(*listRespBody.Data.TotalSize)/int(pageSize); i++ {
		pageNumber = int32(i)
		listRespBody, err = impl.ListMSEGatewayRoutesByAPI(domainName, &pageNumber, &pageSize)
		if err != nil {
			return "", err
		}

		if listRespBody.Data != nil {
			for _, route := range listRespBody.Data.Result {
				if strings.Contains(*route.Name, zoneName) {
					return *route.Name, nil
				}
			}
		}
	}

	return "", errors.Errorf("not matched route contain %s found in MSE Gateway route list", zoneName)
}

func (impl *MseAdapterImpl) createPluginConfig(req *PluginReqDto, confList map[string][]mseclient.GetPluginConfigResponseBodyDataGatewayConfigList) (string, int64, error) {

	switch req.Name {
	case mseCommon.MsePluginKeyAuth:
		//return fmt.Sprintf("consumers: \n# 注意！该凭证仅做示例使用，请勿用于具体业务，造成安全风险\n- credential: example-key\n  name: consumer1\nkeys:\n- apikey\n- x-api-key\n")
		return mseplugins.CreateKeyAuthConfig(req, confList)
	case mseCommon.MsePluginBasicAuth:
		return "", -1, nil
	case mseCommon.MsePluginHmacAuth:
		return "", -1, nil
	case mseCommon.MsePluginCustomResponse:
		return "", -1, nil
	case mseCommon.MsePluginRequestBlock:
		return "", -1, nil
	case mseCommon.MsePluginBotDetect:
		return "", -1, nil
	case mseCommon.MsePluginKeyRateLimit:
		//return fmt.Sprintf("limit_by_header: x-api-key\nlimit_keys:\n- key: example-key-a\n  #query_per_second: 10\n  query_per_minute: 1\n- key: example-key-b\n  #query_per_second: 1\n  query_per_minute: 1\n")
		return "", -1, nil
	case mseCommon.MsePluginHttp2Misdirect:
		return "", -1, nil
	case mseCommon.MsePluginJwtAuth:
		return "", -1, nil
	case mseCommon.MsePluginHttpRealIP:
		return "", -1, nil
	default:
		return "", -1, nil
	}
}
