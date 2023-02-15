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
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	. "github.com/erda-project/erda/internal/tools/orchestrator/hepa/common/vars"
	gateway_providers "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers"
	. "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/kong/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/hepautils"
)

const (
	Mse_Version                               = "mse-1.0.5"
	Mse_Provider_Name                         = "MSE"
	MseBurstMultiplier                        = "2"
	Mse_Ingress_Controller_ACK_Namespace      = "mse-ingress-controller"
	Mse_Ingress_Controller_ACK_DeploymentName = "ack-mse-ingress-controller"
	Mse_Need_Drop_Annotation                  = "need_drop_annotation"
)

type MseAdapterImpl struct {
	ProviderName string
}

func NewMseAdapter() gateway_providers.GatewayAdapter {
	return &MseAdapterImpl{
		ProviderName: Mse_Provider_Name,
	}
}

func (impl *MseAdapterImpl) GatewayProviderExist() bool {
	//TODO:
	// check Deployment mse-ingress-controller/ack-mse-ingress-controller in k8s cluster
	if impl == nil || impl.ProviderName != Mse_Provider_Name {
		return false
	}
	return true
}

func (impl *MseAdapterImpl) GetVersion() (string, error) {
	log.Debugf("GetVersion is not implemented in Mse gateway provider.")
	return Mse_Version, nil
}

func (impl *MseAdapterImpl) CheckPluginEnabled(pluginName string) (bool, error) {
	log.Debugf("CheckPluginEnabled is not implemented in Mse gateway provider.")
	return true, nil
}

func (impl *MseAdapterImpl) CreateConsumer(req *KongConsumerReqDto) (*KongConsumerRespDto, error) {
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

	return &KongConsumerRespDto{
		CustomId:  req.CustomId,
		CreatedAt: time.Now().Unix(),
		Id:        uuid.String(),
	}, nil
}

func (impl *MseAdapterImpl) DeleteConsumer(id string) error {
	log.Debugf("DeleteConsumer is not implemented in Mse gateway provider.")
	return nil
}

func (impl *MseAdapterImpl) CreateOrUpdateRoute(req *KongRouteReqDto) (*KongRouteRespDto, error) {
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

	return &KongRouteRespDto{
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

func (impl *MseAdapterImpl) UpdateRoute(req *KongRouteReqDto) (*KongRouteRespDto, error) {
	return &KongRouteRespDto{}, nil
}

func (impl *MseAdapterImpl) CreateOrUpdateService(req *KongServiceReqDto) (*KongServiceRespDto, error) {
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
	return &KongServiceRespDto{
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

func (impl *MseAdapterImpl) DeletePluginIfExist(req *KongPluginReqDto) error {
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

func (impl *MseAdapterImpl) CreateOrUpdatePluginById(req *KongPluginReqDto) (*KongPluginRespDto, error) {
	log.Debugf("CreateOrUpdatePluginById is not really implemented in Mse gateway provider.")
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

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	return &KongPluginRespDto{
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
	}, nil
}

func (impl *MseAdapterImpl) GetPlugin(req *KongPluginReqDto) (*KongPluginRespDto, error) {
	log.Debugf("GetPlugin is not  really implemented in Mse gateway provider.")
	if req == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	return &KongPluginRespDto{
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

func (impl *MseAdapterImpl) CreateOrUpdatePlugin(req *KongPluginReqDto) (*KongPluginRespDto, error) {
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

func (impl *MseAdapterImpl) AddPlugin(req *KongPluginReqDto) (*KongPluginRespDto, error) {
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

	return &KongPluginRespDto{
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

func (impl *MseAdapterImpl) PutPlugin(req *KongPluginReqDto) (*KongPluginRespDto, error) {
	log.Debugf("PutPlugin is not implemented in Mse gateway provider.")
	return nil, nil
}

func (impl *MseAdapterImpl) UpdatePlugin(req *KongPluginReqDto) (*KongPluginRespDto, error) {
	log.Debugf("UpdatePlugin is not really implemented in Mse gateway provider.")
	if impl == nil {
		return nil, errors.New("mse can't be attached")
	}
	if req == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	return &KongPluginRespDto{
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
	log.Debugf("RemovePlugin is not implemented in Mse gateway provider.")
	return nil
}

func (impl *MseAdapterImpl) CreateCredential(req *KongCredentialReqDto) (*KongCredentialDto, error) {
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
	return &KongCredentialDto{
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

func (impl *MseAdapterImpl) DeleteCredential(consumerId, pluginName, credentialId string) error {
	log.Debugf("DeleteCredential is not implemented in Mse gateway provider.")
	return nil
}

func (impl *MseAdapterImpl) GetCredentialList(consumerId, pluginName string) (*KongCredentialListDto, error) {
	log.Debugf("GetCredentialList is not implemented in Mse gateway provider.")
	return &KongCredentialListDto{}, nil
}

func (impl *MseAdapterImpl) CreateAclGroup(consumerId, customId string) error {
	log.Debugf("CreateAclGroup is not implemented in Mse gateway provider.")
	return nil
}

func (impl *MseAdapterImpl) CreateUpstream(req *KongUpstreamDto) (*KongUpstreamDto, error) {
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

func (impl *MseAdapterImpl) GetUpstreamStatus(upstreamId string) (*KongUpstreamStatusRespDto, error) {
	log.Debugf("GetUpstreamStatus is not really implemented in Mse gateway provider.")
	return &KongUpstreamStatusRespDto{
		Data: []KongTargetDto{},
	}, nil
}

func (impl *MseAdapterImpl) AddUpstreamTarget(upstreamId string, req *KongTargetDto) (*KongTargetDto, error) {
	log.Debugf("AddUpstreamTarget is not really implemented in Mse gateway provider.")
	if upstreamId == "" || req == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate id for UpstreamTarget")
	}

	return &KongTargetDto{
		Id:         uuid.String(),
		Target:     req.Target,
		Weight:     req.Weight,
		UpstreamId: upstreamId,
		CreatedAt:  time.Now().Unix(),
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

func (impl *MseAdapterImpl) GetRoutes() ([]KongRouteRespDto, error) {
	log.Debugf("GetRoutes is not implemented in Mse gateway provider.")
	return []KongRouteRespDto{}, nil
}

func (impl *MseAdapterImpl) GetRoutesWithTag(tag string) ([]KongRouteRespDto, error) {
	log.Debugf("GetRoutesWithTag is not implemented in Mse gateway provider.")
	return []KongRouteRespDto{}, nil
}
