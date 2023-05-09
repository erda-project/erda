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
	"encoding/json"
	"strconv"
	"time"

	mseclient "github.com/alibabacloud-go/mse-20190531/v3/client"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/msp/resource/utils"
	. "github.com/erda-project/erda/internal/tools/orchestrator/hepa/common/vars"
	gateway_providers "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers"
	. "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	mseplugins "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/plugins"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/hepautils"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type MseAdapterImpl struct {
	Bdl          *bundle.Bundle
	ProviderName string
	// Aliyun AccessKey ID
	AccessKeyID string
	// Aliyun AccessKey Secret
	AccessKeySecret string
	// Aliyun Mse Gateway unique ID ("gw-eeab5d74e29d435f87bcxxxxxxxxxxxxx")
	GatewayUniqueID string
	// Aliyun Mse Gateway EndPoint ("mse.cn-hangzhou.aliyuncs.com")
	GatewayEndpoint string
	// 用于映射不同集群的插件列表
	ClusterName string
}

func NewMseAdapter(az string) (gateway_providers.GatewayAdapter, error) {
	adapter := &MseAdapterImpl{
		Bdl: bundle.New(bundle.WithScheduler(),
			bundle.WithOrchestrator(),
			bundle.WithHepa(),
			bundle.WithCMDB(),
			bundle.WithErdaServer(),
			bundle.WithPipeline(),
			bundle.WithMonitor(),
			bundle.WithCollector(),
			bundle.WithHTTPClient(httpclient.New(httpclient.WithTimeout(time.Second*10, time.Second*60))),
		),
		ProviderName:    common.MseProviderName,
		GatewayEndpoint: common.MseDefaultServerEndpoint,
		ClusterName:     az,
	}

	configmap, err := adapter.Bdl.QueryClusterInfo(az)
	if err != nil {
		log.Errorf("bundle QueryClusterInfo for cluster %s failed: %v\n", az, err)
		return nil, err
	}

	config := map[string]string{}
	err = utils.JsonConvertObjToType(configmap, &config)
	if err != nil {
		log.Errorf("JsonConvertObjToType cluster %s  configmap failed: %v\n", az, err)
		return nil, err
	}

	accessKeyID, ok := config["ALIYUN_ACCESS_KEY_ID"]
	if !ok || accessKeyID == "" {
		log.Errorf("ALIYUN_ACCESS_KEY_ID not set in cluster %s  configmap", az)
		return nil, errors.Errorf("ALIYUN_ACCESS_KEY_ID not set in cluster %s  configmap", az)
	}
	adapter.AccessKeyID = accessKeyID

	accessKeySecret, ok := config["ALIYUN_ACCESS_KEY_SECRET"]
	if !ok || accessKeySecret == "" {
		log.Errorf("ALIYUN_ACCESS_KEY_SECRET not set in cluster %s  configmap", az)
		return nil, errors.Errorf("ALIYUN_ACCESS_KEY_SECRET not set in cluster %s  configmap", az)
	}
	adapter.AccessKeySecret = accessKeySecret

	gatewayUniqueID, ok := config["ALIYUN_MSE_GATEWAY_ID"]
	if !ok || gatewayUniqueID == "" {
		log.Errorf("ALIYUN_MSE_GATEWAY_ID not set in cluster %s  configmap", az)
		return nil, errors.Errorf("ALIYUN_MSE_GATEWAY_ID not set in cluster %s  configmap", az)
	}
	adapter.GatewayUniqueID = gatewayUniqueID

	mseGatewayEndpoint, ok := config["ALIYUN_MSE_GATEWAY_ENDPOINT"]
	if ok && mseGatewayEndpoint != "" {
		adapter.GatewayEndpoint = mseGatewayEndpoint
	}

	// 避免动态添加自定义插件之后无法感知，因此每次 adapter 初始化都刷新一遍对应集群的插件 ID 列表
	if _, ok = common.MapClusterNameToMSEPluginNameToPluginID[az]; !ok {
		common.MapClusterNameToMSEPluginNameToPluginID[az] = make(map[string]*int64)
	}
	plugins, err := adapter.GetMSEPluginsByAPI(nil, nil, nil)
	if err != nil {
		log.Errorf("get mse plugin list for cluster %s failed: %v", az, err)
		return nil, errors.Errorf("ALIYUN_MSE_GATEWAY_ID not set in cluster %s  configmap", az)
	}
	for _, plugin := range plugins {
		common.MapClusterNameToMSEPluginNameToPluginID[az][*plugin.Name] = plugin.Id
		log.Debugf("cluster %s MSE Gateay Plugin %s ID=%d ", az, *plugin.Name, *plugin.Id)
	}

	return adapter, nil
}

func (impl *MseAdapterImpl) GatewayProviderExist() bool {
	//TODO:
	// check Deployment mse-ingress-controller/ack-mse-ingress-controller in k8s cluster
	if impl == nil || impl.ProviderName != common.MseProviderName {
		return false
	}
	_, err := impl.GetMSEGatewayByAPI()
	if err != nil {
		log.Errorf("get MSE gateway info failed: %v", err)
		return false
	}
	return true
}

func (impl *MseAdapterImpl) GetVersion() (string, error) {
	log.Debugf("GetVersion is not implemented in Mse gateway provider.")
	return common.MseVersion, nil
}

func (impl *MseAdapterImpl) CheckPluginEnabled(pluginName string) (bool, error) {
	if pluginName == "" {
		return false, errors.Errorf("plugin name not set")
	}

	id, ok := common.MapClusterNameToMSEPluginNameToPluginID[impl.ClusterName][pluginName]
	if !ok {
		log.Debugf("plugin %s not support by MSE Gateway now", pluginName)
		return true, nil
	}

	pl, err := impl.GetMSEPluginConfigByIDByAPI(id)
	if err != nil {
		return false, err
	}

	if pl == nil {
		return false, errors.Errorf("plugin %s config not found for MSE Gateway", pluginName)
	}
	// 不检测插件是否启用，只检查插件是否支持，因为开启后续对插件的更新，都显式设置其为启用状态
	return true, nil
}

func (impl *MseAdapterImpl) CreateConsumer(req *ConsumerReqDto) (*ConsumerRespDto, error) {
	log.Debugf("CreateConsumer is not really implemented in Mse gateway provider.")
	if impl == nil {
		return nil, errors.New("mse can't be attached")
	}
	if req == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}

	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate id for CreateConsumer")
	}

	return &ConsumerRespDto{
		CustomId:  req.CustomId,
		CreatedAt: time.Now().Unix(),
		Id:        uuid.String(),
	}, nil
}

func (impl *MseAdapterImpl) DeleteConsumer(id string) error {
	log.Debugf("DeleteConsumer is not implemented in Mse gateway provider.")
	return nil
}

func (impl *MseAdapterImpl) CreateOrUpdateRoute(req *RouteReqDto) (*RouteRespDto, error) {
	log.Debugf("CreateOrUpdateRoute is not really implemented in Mse gateway provider.")
	timeNow := time.Now()
	defer func() {
		log.Infof("*MseAdapterImpl.CreateOrUpdateRoute costs %dms", time.Now().Sub(timeNow).Milliseconds())
	}()

	if impl == nil {
		return nil, errors.New("kong can't be attached")
	}
	if req == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	req.Adjust(Versioning(impl))
	for i := 0; i < len(req.Paths); i++ {
		pth, err := hepautils.RenderKongUri(req.Paths[i])
		if err != nil {
			return nil, errors.Wrap(err, "failed to render service path")
		}
		req.Paths[i] = pth
	}

	routeId := ""
	if len(req.RouteId) != 0 {
		routeId = req.RouteId
	} else {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate id for Route")
		}
		req.RouteId = uuid.String()
	}

	return &RouteRespDto{
		Id:        routeId,
		Protocols: req.Protocols,
		Methods:   req.Methods,
		Hosts:     req.Hosts,
		Paths:     req.Paths,
		Service:   Service{Id: req.Service.Id},
	}, nil
}

func (impl *MseAdapterImpl) DeleteRoute(routeId string) error {
	log.Debugf("DeleteRoute is not implemented in Mse gateway provider.")
	return nil
}

func (impl *MseAdapterImpl) UpdateRoute(req *RouteReqDto) (*RouteRespDto, error) {
	return &RouteRespDto{}, nil
}

func (impl *MseAdapterImpl) CreateOrUpdateService(req *ServiceReqDto) (*ServiceRespDto, error) {
	log.Debugf("CreateOrUpdateService is not really implemented in Mse gateway provider.")
	timeNow := time.Now()
	defer func() {
		log.Infof("*MseAdapterImpl.CreateOrUpdateService costs %dms", time.Now().Sub(timeNow).Milliseconds())
	}()

	if impl == nil {
		return nil, errors.New("kong can't be attached")
	}
	if req == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	pth, err := hepautils.RenderKongUri(req.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render service path")
	}
	req.Path = pth
	serviceId := ""
	if len(req.ServiceId) != 0 {
		serviceId = req.ServiceId
	} else {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate id for Service")
		}
		serviceId = uuid.String()
	}

	port := 80
	if req.Port > 0 {
		port = req.Port
	}
	return &ServiceRespDto{
		Id:       serviceId,
		Name:     req.Name,
		Protocol: req.Protocol,
		Host:     req.Host,
		Port:     port,
		Path:     req.Path,
	}, nil
}

func (impl *MseAdapterImpl) DeleteService(serviceId string) error {
	log.Debugf("DeletePluginIfExist is not implemented in Mse gateway provider.")
	return nil
}

func (impl *MseAdapterImpl) DeletePluginIfExist(req *PluginReqDto) error {
	log.Debugf("DeletePluginIfExist is not really implemented in Mse gateway provider.")
	enabled, err := impl.CheckPluginEnabled(req.Name)
	if err != nil {
		return err
	}
	if !enabled {
		log.Warnf("plugin %s not enabled, req:%+v", req.Name, req)
		return nil
	}
	exist, err := impl.GetPlugin(req)
	if err != nil {
		return err
	}
	if exist == nil {
		return nil
	}
	return impl.RemovePlugin(exist.Id)
}

func (impl *MseAdapterImpl) CreateOrUpdatePluginById(req *PluginReqDto) (*PluginRespDto, error) {
	log.Infof("CreateOrUpdatePluginById with Req: %+v\n", *req)
	if impl == nil {
		return nil, errors.New("mse can't be attached")
	}
	if req == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	if len(req.Id) != 0 {
		req.CreatedAt = time.Now().Unix() * 1000
	}
	if len(req.PluginId) == 0 {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate id for CreateOrUpdatePluginById")
		}
		req.PluginId = uuid.String()
	}
	// enabled 设置网关一直启用状态
	enabled := true

	// The application scope of the plug-in.
	// *   0: global
	// *   1: domain names
	// *   2: routes
	var configLevel int32 = 0 // 目前(2023.02) 只支持设置为 0

	pResp := &PluginRespDto{
		Id:         req.PluginId,
		ServiceId:  req.ServiceId,
		RouteId:    req.RouteId,
		ConsumerId: req.ConsumerId,
		Route:      req.Route,
		Service:    req.Service,
		Consumer:   req.Consumer,
		Name:       req.Name,
		Config:     req.Config,
		Enabled:    enabled,
		CreatedAt:  time.Now().Unix(),
		PolicyId:   req.Id,
	}

	pluginId, ok := common.MapClusterNameToMSEPluginNameToPluginID[impl.ClusterName][req.Name]
	if !ok {
		log.Debugf("plugin %s not support in MSE Gateway", req.Name)
		return pResp, nil
	}
	pluginConfig, err := impl.GetMSEPluginConfigByIDByAPI(pluginId)
	if err != nil {
		return nil, errors.Errorf("failed to get plugin %s config for CreateOrUpdatePluginById, error: %v", req.Name, err)
	}

	// map[configLevel]GetPluginConfigResponseBodyDataGatewayConfigList
	confList := make(map[string][]mseclient.GetPluginConfigResponseBodyDataGatewayConfigList)
	if pluginConfig != nil && len(pluginConfig.GatewayConfigList) > 0 {
		// Step 1: 如果插未启用，则先启用插件
		for _, cf := range pluginConfig.GatewayConfigList {
			if cf.ConfigLevel == nil {
				cfLevel := mseplugins.MsePluginConfigLevelGlobalNumber
				cf.ConfigLevel = &cfLevel
			}
			if cf.Enable == nil {
				enablePlugin := false
				cf.Enable = &enablePlugin
			}
			if *cf.ConfigLevel == mseplugins.MsePluginConfigLevelGlobalNumber && *cf.Enable == false {
				// Step 1: 如果插未启用，则先启用插件
				// 启用 插件，对应才会有插件配置 ID，否则因为无插件配置 ID 会导致后续无法更新插件
				defaultConfig := ""
				switch req.Name {
				case common.MsePluginKeyAuth:
					defaultConfig = mseplugins.MseDefaultKeyAuthConfig
				case common.MsePluginHmacAuth:
					defaultConfig = mseplugins.MseDefaultHmacAuthConfig
				case common.MsePluginParaSignAuth:
					defaultConfig = mseplugins.MseDefaultParaSignAuthConfig
				case common.MsePluginIP:
					defaultConfig = mseplugins.MseDefaultErdaIPConfig
				}
				resp, err := impl.UpdateMSEPluginConfigByIDByAPI(pluginId, nil, &defaultConfig, &configLevel, &enabled)
				if err != nil {
					log.Errorf("failed to enable plugin %s for CreateOrUpdatePluginById, error: %v", req.Name, err)
					return nil, errors.Errorf("failed to enable plugin %s for CreateOrUpdatePluginById, error: %v", req.Name, err)
				}
				if resp.Data != nil {
					log.Debugf("plugin %s ConfigID=%d", req.Name, *resp.Data)
				}

				// Step 2: 重新获取插件配置
				pluginConfig, err = impl.GetMSEPluginConfigByIDByAPI(pluginId)
				if err != nil {
					return nil, errors.Errorf("failed to get plugin %s config for CreateOrUpdatePluginById, error: %v", req.Name, err)
				}
				break
			}
		}
	} else {
		return nil, errors.Errorf("failed to get plugin %s config for CreateOrUpdatePluginById: no config found for this plugin", req.Name)
	}

	if pluginConfig != nil && len(pluginConfig.GatewayConfigList) > 0 {
		for _, cf := range pluginConfig.GatewayConfigList {
			log.Infof("cf.ConfigLevel=%d  cf.Config=%s", *cf.ConfigLevel, *cf.Config)
			mapKey := ""
			switch *cf.ConfigLevel {
			case mseplugins.MsePluginConfigLevelDomainNumber:
				mapKey = mseplugins.MsePluginConfigLevelDomain
			case mseplugins.MsePluginConfigLevelRouteNumber:
				mapKey = mseplugins.MsePluginConfigLevelRoute
			default:
				mapKey = mseplugins.MsePluginConfigLevelGlobal
			}
			if _, exist := confList[mapKey]; !exist {
				confList[mapKey] = make([]mseclient.GetPluginConfigResponseBodyDataGatewayConfigList, 0)
			}
			confList[mapKey] = append(confList[mapKey], *cf)
		}
	} else {
		return nil, errors.Errorf("failed to get plugin %s config for CreateOrUpdatePluginById: no config found for this plugin", req.Name)
	}

	routeName, err := impl.GetMSEGatewayRouteNameByZoneName(req.ZoneName, nil)
	if err != nil {
		return nil, errors.Errorf("GetMSEGatewayRouteNameByZoneName failed for plugin %s for zone_name %s CreateOrUpdatePluginById, error: %v", req.Name, req.ZoneName, err)
	}

	req.MSERouteName = routeName
	config, configId, err := impl.createPluginConfig(req, confList)
	if err != nil {
		return nil, errors.Errorf("createPluginConfig failed for plugin %s for CreateOrUpdatePluginById, error: %v", req.Name, err)
	}
	_, err = impl.UpdateMSEPluginConfigByIDByAPI(pluginId, &configId, &config, &configLevel, &enabled)
	if err != nil {
		return nil, errors.Errorf("failed to update plugin %s config for CreateOrUpdatePluginById, error: %v", req.Name, err)
	}

	return pResp, nil
}

func (impl *MseAdapterImpl) GetPlugin(req *PluginReqDto) (*PluginRespDto, error) {
	if req == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	pluginID, ok := common.MapClusterNameToMSEPluginNameToPluginID[impl.ClusterName][req.Name]
	if !ok {
		log.Debugf("plugin %s not support in MSE Gateway", req.Name)
		return &PluginRespDto{
			Id:         req.PluginId,
			ServiceId:  req.ServiceId,
			RouteId:    req.RouteId,
			ConsumerId: req.ConsumerId,
			Route:      req.Route,
			Service:    req.Service,
			Consumer:   req.Consumer,
			Name:       req.Name,
			Config:     req.Config,
			Enabled:    enabled,
			CreatedAt:  req.CreatedAt,
			PolicyId:   req.Id,
		}, nil
	}

	pluginConfig, err := impl.GetMSEPluginConfigByIDByAPI(pluginID)
	if err != nil {
		return nil, errors.Errorf("failed to get plugin %s config for CreateOrUpdatePluginById, error: %v", req.Name, err)
	}

	// 插件名称到插件配置列表的映射: map[pluginName]GatewayConfigList
	//config := make(map[string][]*mseclient.GetPluginConfigResponseBodyDataGatewayConfigList, 0)
	config := make(map[string]interface{}, 0)
	if pluginConfig != nil && len(pluginConfig.GatewayConfigList) > 0 {
		config[req.Name] = pluginConfig.GatewayConfigList
	}

	return &PluginRespDto{
		Id:         req.PluginId,
		ServiceId:  req.ServiceId,
		RouteId:    req.RouteId,
		ConsumerId: req.ConsumerId,
		Route:      req.Route,
		Service:    req.Service,
		Consumer:   req.Consumer,
		Name:       req.Name,
		Config:     config,
		Enabled:    enabled,
		CreatedAt:  req.CreatedAt,
		PolicyId:   req.Id,
	}, nil
}

func (impl *MseAdapterImpl) CreateOrUpdatePlugin(req *PluginReqDto) (*PluginRespDto, error) {
	log.Debugf("CreateOrUpdatePlugin is not implemented in Mse gateway provider.")
	timeNow := time.Now()
	defer func() {
		log.Infof("*MseAdapterImpl.CreateOrUpdatePlugin costs %dms", time.Now().Sub(timeNow).Milliseconds())
	}()

	enabled, err := impl.CheckPluginEnabled(req.Name)
	if err != nil {
		return nil, err
	}
	if !enabled {
		log.Warnf("plugin %s not enabled, req:%+v", req.Name, req)
		return nil, nil
	}
	exist, err := impl.GetPlugin(req)
	if err != nil {
		return nil, err
	}
	if exist == nil {
		return impl.AddPlugin(req)
	}
	req.Id = exist.Id
	req.PluginId = exist.Id
	return impl.PutPlugin(req)
}

func (impl *MseAdapterImpl) AddPlugin(req *PluginReqDto) (*PluginRespDto, error) {
	log.Debugf("AddPlugin is not implemented in Mse gateway provider.")
	if impl == nil {
		return nil, errors.New("mse can't be attached")
	}
	if req == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	enabled, err := impl.CheckPluginEnabled(req.Name)
	if err != nil {
		return nil, err
	}
	if !enabled {
		log.Warnf("plugin %s not enabled, req:%+v", req.Name, req)
		return nil, nil
	}
	pluginEnabled := true
	if req.Enabled != nil {
		pluginEnabled = *req.Enabled
	}

	return &PluginRespDto{
		Id:         req.Id,
		ServiceId:  req.ServiceId,
		RouteId:    req.RouteId,
		ConsumerId: req.ConsumerId,
		Route:      req.Route,
		Service:    req.Service,
		Consumer:   req.Consumer,
		Name:       req.Name,
		Config:     req.Config,
		Enabled:    pluginEnabled,
		CreatedAt:  time.Now().Unix(),
		PolicyId:   req.PluginId,
	}, nil
}

func (impl *MseAdapterImpl) PutPlugin(req *PluginReqDto) (*PluginRespDto, error) {
	log.Debugf("PutPlugin is not implemented in Mse gateway provider.")
	return nil, nil
}

// UpdatePlugin 更新 req.Name 对应的插件的配置
func (impl *MseAdapterImpl) UpdatePlugin(req *PluginReqDto) (*PluginRespDto, error) {
	if impl == nil {
		return nil, errors.New("mse can't be attached")
	}
	if req == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}

	if req.Name == "" {
		return nil, errors.New("plugin name not set")
	}

	// enabled 设置网关一直启用状态
	enabled := true
	pluginId, ok := common.MapClusterNameToMSEPluginNameToPluginID[impl.ClusterName][req.Name]
	if !ok {
		log.Debugf("plugin %s not support in MSE Gateway", req.Name)
		return &PluginRespDto{
			Id:         req.Id,
			ServiceId:  req.ServiceId,
			RouteId:    req.RouteId,
			ConsumerId: req.ConsumerId,
			Route:      req.Route,
			Service:    req.Service,
			Consumer:   req.Consumer,
			Name:       req.Name,
			Config:     req.Config,
			Enabled:    enabled,
			CreatedAt:  req.CreatedAt,
			PolicyId:   req.PluginId,
		}, nil
	}

	if len(req.Config) == 0 {
		return nil, errors.New("config info not set")
	}

	confList, ok := req.Config[req.Name]
	if !ok {
		return nil, errors.Errorf("plugin %s config not set", req.Name)
	}

	clist, ok := confList.([]*mseclient.GetPluginConfigResponseBodyDataGatewayConfigList)
	if !ok {
		return nil, errors.Errorf("not valid mse gateway config list for plugin %s: %+v", req.Name, confList)
	}

	var cf *mseclient.GetPluginConfigResponseBodyDataGatewayConfigList = nil
	for idx := range clist {
		// 因为当前(2022.02.23) MSE 网关插件配置只支持全局配置（ConfigLevel 为 0）
		if *clist[idx].ConfigLevel == 0 {
			cf = clist[idx]
		}
	}
	if cf == nil {
		return nil, errors.Errorf("no mse gateway global config list for plugin %s: %+v", req.Name, confList)
	}

	_, err := impl.UpdateMSEPluginConfigByIDByAPI(pluginId, cf.Id, cf.Config, cf.ConfigLevel, &enabled)
	if err != nil {
		return nil, errors.Errorf("call UpdateMSEPluginConfigByIDByAPI for plugin %s with config [%s] failed, error: %+v", req.Name, *cf.Config, err)
	}

	return &PluginRespDto{
		Id:         req.Id,
		ServiceId:  req.ServiceId,
		RouteId:    req.RouteId,
		ConsumerId: req.ConsumerId,
		Route:      req.Route,
		Service:    req.Service,
		Consumer:   req.Consumer,
		Name:       req.Name,
		Config:     req.Config,
		Enabled:    enabled,
		CreatedAt:  req.CreatedAt,
		PolicyId:   req.PluginId,
	}, nil
}

func (impl *MseAdapterImpl) RemovePlugin(id string) error {
	log.Debugf("RemovePlugin is not needed in Mse gateway provider when Mse Gateway use global config.")
	return nil
}

func (impl *MseAdapterImpl) CreateCredential(req *CredentialReqDto) (*CredentialDto, error) {
	log.Debugf("CreateCredential is not really implemented in Mse gateway provider.")
	if impl == nil {
		return nil, errors.New("mse can't be attached")
	}
	if req == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	if req.PluginName == "hmac-auth" {
		req.Config.ToHmacReq()
	}
	return &CredentialDto{
		ConsumerId:   req.ConsumerId,
		CreatedAt:    time.Now().Unix(),
		Id:           req.Config.Id,
		Key:          req.Config.Key,
		RedirectUrl:  req.Config.RedirectUrl,
		RedirectUrls: req.Config.RedirectUrls,
		Name:         req.PluginName,
		ClientId:     req.Config.ClientId,
		ClientSecret: req.Config.ClientSecret,
		Secret:       req.Config.Secret,
		Username:     req.Config.Username,
	}, nil
}

func (impl *MseAdapterImpl) DeleteCredential(consumerId, pluginName, credentialStr string) error {
	log.Debugf("DeleteCredential is not implemented in Mse gateway provider.")

	_, ok := common.MapClusterNameToMSEPluginNameToPluginID[impl.ClusterName][pluginName]
	if !ok {
		log.Debugf("plugin %s not support in MSE Gateway", pluginName)
		return nil
	}

	pluginConf, err := impl.GetPlugin(&PluginReqDto{
		Name: pluginName,
	})
	if err != nil {
		log.Errorf("update mse plugin %s when get plugin conf failed: %v\n", pluginName, err)
		return err
	}

	pluginConfig, ok := pluginConf.Config[pluginName]
	if !ok {
		return nil
	}

	credential := CredentialDto{}
	err = json.Unmarshal([]byte(credentialStr), &credential)

	confList, err := mseplugins.UpdatePluginConfigWhenDeleteCredential(pluginName, credential, pluginConfig)
	if err != nil {
		log.Errorf("update mse plugin %s when delete credential failed: %v\n", pluginName, err)
		return err
	}

	if confList != nil {
		newConfig := make(map[string]interface{})
		newConfig[pluginName] = confList

		_, err = impl.UpdatePlugin(&PluginReqDto{
			Name:   pluginName,
			Config: newConfig,
		})
		if err != nil {
			log.Errorf("update mse plugin %s with config %v failed: %v\n", pluginName, newConfig, err)
			return err
		}
	}

	return nil
}

func (impl *MseAdapterImpl) GetCredentialList(consumerId, pluginName string) (*CredentialListDto, error) {
	log.Debugf("GetCredentialList is not implemented in Mse gateway provider.")
	return &CredentialListDto{}, nil
}

func (impl *MseAdapterImpl) CreateAclGroup(consumerId, customId string) error {
	log.Debugf("CreateAclGroup is not implemented in Mse gateway provider.")
	return nil
}

func (impl *MseAdapterImpl) CreateUpstream(req *UpstreamDto) (*UpstreamDto, error) {
	log.Debugf("CreateUpstream is not really implemented in Mse gateway provider.")
	if impl == nil {
		return nil, errors.New("mse can't be attached")
	}
	if req == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	if req.Id == "" {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate id for CreateUpstream")
		}
		req.Id = uuid.String()
	}
	return req, nil
}

func (impl *MseAdapterImpl) GetUpstreamStatus(upstreamId string) (*UpstreamStatusRespDto, error) {
	log.Debugf("GetUpstreamStatus is not really implemented in Mse gateway provider.")
	return &UpstreamStatusRespDto{
		Data: []TargetDto{},
	}, nil
}

func (impl *MseAdapterImpl) AddUpstreamTarget(upstreamId string, req *TargetDto) (*TargetDto, error) {
	log.Debugf("AddUpstreamTarget is not really implemented in Mse gateway provider.")
	if upstreamId == "" || req == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate id for UpstreamTarget")
	}

	return &TargetDto{
		Id:         uuid.String(),
		Target:     req.Target,
		Weight:     req.Weight,
		UpstreamId: upstreamId,
		CreatedAt:  json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
		Health:     req.Health,
	}, nil
}

func (impl *MseAdapterImpl) DeleteUpstreamTarget(upstreamId, targetId string) error {
	log.Debugf("DeleteUpstreamTarget is not implemented in Mse gateway provider.")
	return nil
}

func (impl *MseAdapterImpl) TouchRouteOAuthMethod(id string) error {
	log.Debugf("TouchRouteOAuthMethod is not implemented in Mse gateway provider.")
	return nil
}

func (impl *MseAdapterImpl) GetRoutes() ([]RouteRespDto, error) {
	log.Debugf("GetRoutes is not implemented in Mse gateway provider.")
	return []RouteRespDto{}, nil
}

func (impl *MseAdapterImpl) GetRoutesWithTag(tag string) ([]RouteRespDto, error) {
	log.Debugf("GetRoutesWithTag is not implemented in Mse gateway provider.")
	return []RouteRespDto{}, nil
}
