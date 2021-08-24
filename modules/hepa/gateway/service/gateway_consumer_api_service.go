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

package service

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/hepa/common"
	. "github.com/erda-project/erda/modules/hepa/common/vars"
	"github.com/erda-project/erda/modules/hepa/gateway/assembler"
	gw "github.com/erda-project/erda/modules/hepa/gateway/dto"
	"github.com/erda-project/erda/modules/hepa/kong"
	kongDto "github.com/erda-project/erda/modules/hepa/kong/dto"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
	db "github.com/erda-project/erda/modules/hepa/repository/service"
)

type GatewayConsumerApiServiceImpl struct {
	consumerApiDb db.GatewayConsumerApiService
	pluginDb      db.GatewayPluginInstanceService
	consumerDb    db.GatewayConsumerService
	apiDb         db.GatewayApiService
	policyDb      db.GatewayPolicyService
	serviceDb     db.GatewayServiceService
	routeDb       db.GatewayRouteService
	kongAssembler assembler.GatewayKongAssembler
	dbAssembler   assembler.GatewayDbAssembler
}

func NewGatewayConsumerApiServiceImpl() (*GatewayConsumerApiServiceImpl, error) {
	consumerApiDb, _ := db.NewGatewayConsumerApiServiceImpl()
	pluginDb, _ := db.NewGatewayPluginInstanceServiceImpl()
	consumerDb, _ := db.NewGatewayConsumerServiceImpl()
	apiDb, _ := db.NewGatewayApiServiceImpl()
	policyDb, _ := db.NewGatewayPolicyServiceImpl()
	serviceDb, _ := db.NewGatewayServiceServiceImpl()
	routeDb, _ := db.NewGatewayRouteServiceImpl()
	return &GatewayConsumerApiServiceImpl{
		consumerApiDb: consumerApiDb,
		pluginDb:      pluginDb,
		consumerDb:    consumerDb,
		apiDb:         apiDb,
		policyDb:      policyDb,
		serviceDb:     serviceDb,
		routeDb:       routeDb,
		kongAssembler: assembler.GatewayKongAssemblerImpl{},
		dbAssembler:   assembler.GatewayDbAssemblerImpl{},
	}, nil
}

func (impl GatewayConsumerApiServiceImpl) Create(consumerId string, apiId string) (string, error) {
	if len(consumerId) == 0 || len(apiId) == 0 {
		return "", errors.New(ERR_INVALID_ARG)
	}
	consumer, err := impl.consumerDb.GetById(consumerId)
	if err != nil {
		return "", errors.WithStack(err)
	}
	if consumer == nil {
		return "", errors.Errorf("consumer[%s] not exists", consumerId)
	}
	gatewayApi, err := impl.apiDb.GetById(apiId)
	if err != nil {
		return "", errors.WithStack(err)
	}
	if gatewayApi == nil {
		return "", errors.Errorf("gatewayApi[%s] not exists", apiId)
	}
	aclInstance, err := impl.pluginDb.GetByPluginNameAndApiId("acl", apiId)
	log.Debugf("%+v", aclInstance) // output for debug
	if err != nil {
		return "", errors.WithStack(err)
	}
	if aclInstance != nil {
		aclGroup, err := impl.getAclGroup(apiId)
		log.Debugf("%+v", aclGroup) // output for debug

		if err != nil {
			return "", errors.WithStack(err)
		}
		aclGroup[consumerId] = true

		kongAdapter := kong.NewKongAdapterForConsumer(consumer)
		err = impl.updateAclGroup(kongAdapter, aclGroup, aclInstance.PluginId)
		if err != nil {
			return "", errors.WithStack(err)
		}
	}
	gatewayConsumerApi, err := impl.consumerApiDb.GetByConsumerAndApi(consumerId, apiId)
	if err != nil {
		return "", errors.WithStack(err)
	}
	if gatewayConsumerApi == nil {
		consumerApi := &orm.GatewayConsumerApi{
			ConsumerId: consumerId,
			ApiId:      apiId,
		}
		_ = impl.consumerApiDb.Insert(consumerApi)
		return consumerApi.Id, nil
	}
	return gatewayConsumerApi.Id, nil
}

func (impl GatewayConsumerApiServiceImpl) Delete(id string) error {
	if len(id) == 0 {
		return errors.New(ERR_INVALID_ARG)
	}
	consumerApi, err := impl.consumerApiDb.GetById(id)
	if err != nil {
		return errors.WithStack(err)
	}
	if consumerApi == nil {
		return errors.Errorf("consumerApi of id[%s] is nil", id)
	}
	kongAdapter := kong.NewKongAdapterByConsumerId(impl.consumerDb, consumerApi.ConsumerId)
	aclInstance, err := impl.pluginDb.GetByPluginNameAndApiId("acl", consumerApi.ApiId)
	if err != nil {
		return errors.WithStack(err)
	}
	if aclInstance != nil {
		var aclGroup map[string]bool
		aclGroup, err = impl.getAclGroup(consumerApi.ApiId)
		if err != nil {
			return errors.WithStack(err)
		}
		delete(aclGroup, consumerApi.ConsumerId)
		kongAdapter := kong.NewKongAdapterByConsumerId(impl.consumerDb, consumerApi.ConsumerId)
		err = impl.updateAclGroup(kongAdapter, aclGroup, aclInstance.PluginId)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	if len(consumerApi.Policies) > 0 {
		policyList := []string{}
		err = json.Unmarshal([]byte(consumerApi.Policies), &policyList)
		if err != nil {
			return errors.WithStack(err)
		}
		for _, policyId := range policyList {
			cond := &orm.GatewayPluginInstance{
				PolicyId:   policyId,
				ApiId:      consumerApi.ApiId,
				ConsumerId: consumerApi.ConsumerId,
			}
			var plugin = new(orm.GatewayPluginInstance)
			plugin, err = impl.pluginDb.GetByAny(cond)
			if err != nil {
				return errors.WithStack(err)
			}
			if plugin == nil {
				log.Warnf("plugin is nil, cond[%+v]", cond)
				continue
			}
			err = kongAdapter.RemovePlugin(plugin.PluginId)
			if err != nil {
				return errors.WithStack(err)
			}
			err = impl.pluginDb.DeleteById(plugin.Id)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}
	err = impl.consumerApiDb.DeleteById(id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
func (impl GatewayConsumerApiServiceImpl) UpdateConsumerApi(reqDto *gw.ConsumerApiReqDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	ret := UNKNOW_ERROR
	var err error
	if reqDto == nil || len(reqDto.List) == 0 {
		log.Warnf("invalid reqDto[%+v]", reqDto)
		return res.SetSuccessAndData(true)
	}
	list := reqDto.List
	for _, consumerApiDto := range list {
		var policiesJson []byte
		var consumerApi *orm.GatewayConsumerApi
		var consumer *orm.GatewayConsumer
		var service *orm.GatewayService
		var route *orm.GatewayRoute
		var kongAdapter kong.KongAdapter
		adds := []orm.GatewayPolicy{}
		updates := []orm.GatewayPolicy{}
		dels := []orm.GatewayPolicy{}
		oldPolicyMap := map[string]orm.GatewayPolicy{}
		var newPolicyMap map[string]orm.GatewayPolicy
		consumerApi, err = impl.consumerApiDb.GetById(consumerApiDto.ConsumerApiId)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
		if consumerApi == nil {
			continue
		}
		consumer, err = impl.consumerDb.GetById(consumerApi.ConsumerId)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
		if consumer == nil {
			continue
		}
		kongAdapter = kong.NewKongAdapterForConsumer(consumer)
		service, err = impl.serviceDb.GetByApiId(consumerApi.ApiId)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
		if service == nil {
			continue
		}
		route, err = impl.routeDb.GetByApiId(consumerApi.ApiId)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
		if route == nil {
			continue
		}
		if len(consumerApi.Policies) > 0 {
			policyList := []string{}
			err = json.Unmarshal([]byte(consumerApi.Policies), &policyList)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
			oldPolicyMap, err = impl.getPolicyMap(policyList)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
		}
		newPolicyMap, err = impl.getPolicyMap(consumerApiDto.Policies)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
		for key, value := range newPolicyMap {
			oldValue, exist := oldPolicyMap[key]
			if !exist {
				adds = append(adds, value)
				continue
			}
			if value.Id == oldValue.Id {
				delete(oldPolicyMap, key)
				continue
			}
			updates = append(updates, value)
			delete(oldPolicyMap, key)
		}
		for _, value := range oldPolicyMap {
			dels = append(dels, value)
		}
		for _, policy := range adds {
			var addResp *kongDto.KongPluginRespDto
			var plugin *orm.GatewayPluginInstance
			pluginParams := assembler.PluginParams{
				PolicyId:   policy.Id,
				ServiceId:  service.Id,
				RouteId:    route.Id,
				ConsumerId: consumer.Id,
				ApiId:      consumerApi.ApiId,
			}
			var addReq *kongDto.KongPluginReqDto
			addReq, err = impl.kongAssembler.BuildKongPluginReqDto("", &policy, service.ServiceId, route.RouteId, consumer.ConsumerId)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
			addResp, err = kongAdapter.AddPlugin(addReq)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
			if addResp == nil {
				ret = UPDATE_API_PLUGIN_FAIL
				err = errors.New("addResp is nil")
				goto errorHappened
			}
			plugin, err = impl.dbAssembler.Resp2GatewayPluginInstance(addResp, pluginParams)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
			err = impl.pluginDb.Insert(plugin)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
		}
		for _, policy := range updates {
			var updateReq *kongDto.KongPluginReqDto
			var updateResp *kongDto.KongPluginRespDto
			var plugin *orm.GatewayPluginInstance
			var oldPlugin *orm.GatewayPluginInstance
			pluginParams := assembler.PluginParams{
				PolicyId:   policy.Id,
				ServiceId:  service.Id,
				RouteId:    route.Id,
				ConsumerId: consumer.Id,
				ApiId:      consumerApi.ApiId,
			}
			cond := &orm.GatewayPluginInstance{
				PluginName: policy.PluginName,
				ApiId:      consumerApi.ApiId,
				ConsumerId: consumerApi.ConsumerId,
			}
			oldPlugin, err = impl.pluginDb.GetByAny(cond)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
			if oldPlugin == nil {
				continue
			}
			updateReq, err = impl.kongAssembler.BuildKongPluginReqDto(oldPlugin.PluginId, &policy, service.ServiceId, route.RouteId, consumer.ConsumerId)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
			if updateReq == nil {
				ret = UPDATE_API_PLUGIN_FAIL
				err = errors.New("updateReq is nil")
				goto errorHappened
			}
			updateResp, err = kongAdapter.PutPlugin(updateReq)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
			if updateResp == nil {
				ret = UPDATE_API_PLUGIN_FAIL
				err = errors.New("updateResp is nil")
				goto errorHappened
			}
			plugin, err = impl.dbAssembler.Resp2GatewayPluginInstance(updateResp, pluginParams)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
			plugin.Id = oldPlugin.Id
			err = impl.pluginDb.Update(plugin)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
		}
		for _, policy := range dels {
			var oldPlugin *orm.GatewayPluginInstance
			cond := &orm.GatewayPluginInstance{
				PluginName: policy.PluginName,
				ApiId:      consumerApi.ApiId,
				ConsumerId: consumerApi.ConsumerId,
			}
			oldPlugin, err = impl.pluginDb.GetByAny(cond)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
			if oldPlugin == nil {
				continue
			}
			err = kongAdapter.RemovePlugin(oldPlugin.PluginId)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
			err = impl.pluginDb.DeleteById(oldPlugin.Id)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
		}
		policiesJson, err = json.Marshal(consumerApiDto.Policies)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
		consumerApi.Policies = string(policiesJson)
		err = impl.consumerApiDb.Update(consumerApi)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
	}
	return res.SetSuccessAndData(true)
errorHappened:
	log.Errorf("error happened:%+v", err)
	return res.SetReturnCode(ret)
}

func (impl GatewayConsumerApiServiceImpl) updateAclGroup(kongAdapter kong.KongAdapter, aclGroup map[string]bool, pluginId string) error {
	if pluginId == "" {
		return errors.New(ERR_INVALID_ARG)
	}
	var buffer bytes.Buffer
	for group := range aclGroup {
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(group)
	}
	wl := buffer.String()
	config := map[string]interface{}{}
	if len(wl) == 0 {
		wl = ","
	}
	// consumers := []string{}
	// for consumer, _ := range aclGroup {
	// 	consumers = append(consumers, consumer)
	// }
	// config["whitelist"] = consumers
	config["whitelist"] = wl
	req := &kongDto.KongPluginReqDto{
		PluginId: pluginId,
		Name:     "acl",
		Config:   config,
	}
	_, err := kongAdapter.UpdatePlugin(req)
	return err
}

func (impl GatewayConsumerApiServiceImpl) getAclGroup(apiId string) (map[string]bool, error) {
	set := map[string]bool{}
	consumerApiList, err := impl.consumerApiDb.SelectByApi(apiId)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, api := range consumerApiList {
		set[api.ConsumerId] = true
	}
	return set, nil
}

func (impl GatewayConsumerApiServiceImpl) getPolicyMap(policyList []string) (map[string]orm.GatewayPolicy, error) {
	policyMap := map[string]orm.GatewayPolicy{}
	for _, policyId := range policyList {
		if len(policyId) == 0 {
			continue
		}
		policy, err := impl.policyDb.GetById(policyId)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if policy == nil {
			log.Errorf("policy[%s] not exists", policyId)
			continue
		}
		policyMap[policy.PluginName] = *policy
	}
	return policyMap, nil
}
