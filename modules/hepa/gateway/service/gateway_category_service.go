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

package service

import (
	"encoding/json"
	"strconv"

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

type GatewayCategoryServiceImpl struct {
	apiDb         db.GatewayApiService
	azDb          db.GatewayAzInfoService
	serviceDb     db.GatewayServiceService
	routeDb       db.GatewayRouteService
	policyDb      db.GatewayPolicyService
	consumerDb    db.GatewayConsumerService
	pluginDb      db.GatewayPluginInstanceService
	kongAssembler assembler.GatewayKongAssembler
	dbAssembler   assembler.GatewayDbAssembler
}

func NewGatewayCategoryServiceImpl() (*GatewayCategoryServiceImpl, error) {
	apiDb, _ := db.NewGatewayApiServiceImpl()
	azDb, _ := db.NewGatewayAzInfoServiceImpl()
	serviceDb, _ := db.NewGatewayServiceServiceImpl()
	routeDb, _ := db.NewGatewayRouteServiceImpl()
	policyDb, _ := db.NewGatewayPolicyServiceImpl()
	consumerDb, _ := db.NewGatewayConsumerServiceImpl()
	pluginDb, _ := db.NewGatewayPluginInstanceServiceImpl()
	return &GatewayCategoryServiceImpl{
		apiDb:         apiDb,
		azDb:          azDb,
		serviceDb:     serviceDb,
		routeDb:       routeDb,
		policyDb:      policyDb,
		consumerDb:    consumerDb,
		pluginDb:      pluginDb,
		kongAssembler: assembler.GatewayKongAssemblerImpl{},
		dbAssembler:   assembler.GatewayDbAssemblerImpl{},
	}, nil
}

func (impl GatewayCategoryServiceImpl) verifyPolicyParam(category string, createDto *gw.PolicyCreateDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	categoryEnum := GetPolicyCategory(category)
	if categoryEnum == nil {
		log.Error("categoryEnum is nil")
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	if createDto == nil || createDto.IsEmpty() {
		log.Errorf("invalid createDto[%+v]", createDto)
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	if createDto.DisplayName == "" {
		createDto.DisplayName = createDto.PolicyName
	}
	return res.SetSuccessAndData(categoryEnum)
}

func (impl GatewayCategoryServiceImpl) buildConfig(config map[string]interface{}) (map[string]interface{}, error) {
	res := map[string]interface{}{}
	if qpsObj, ok := config["qps"]; ok {
		if qps, ok := qpsObj.(float64); ok {
			res["second"] = qps
		} else {
			return nil, errors.Errorf("qps from config[%+v] transfer failed", config)
		}
	}
	if qpmObj, ok := config["qpm"]; ok {
		if qpm, ok := qpmObj.(float64); ok {
			res["minute"] = qpm
		} else {
			return nil, errors.Errorf("qpm from config[%+v] transfer failed", config)
		}
	}
	if qphObj, ok := config["qph"]; ok {
		if qph, ok := qphObj.(float64); ok {
			res["hour"] = qph
		} else {
			return nil, errors.Errorf("qph from config[%+v] transfer failed", config)
		}
	}
	if qpdObj, ok := config["qpd"]; ok {
		if qpd, ok := qpdObj.(float64); ok {
			res["day"] = qpd
		} else {
			return nil, errors.Errorf("qpd from config[%+v] transfer failed", config)
		}
	}
	// use local policy
	res["policy"] = "local"
	return res, nil
}

func (impl GatewayCategoryServiceImpl) buildDetail(config map[string]interface{}) (map[string]interface{}, error) {
	res := map[string]interface{}{}
	if qpsObj, ok := config["second"]; ok {
		if qps, ok := qpsObj.(float64); ok {
			res["qps"] = qps
		} else {
			return nil, errors.Errorf("qps from config[%+v] transfer failed", config)
		}
	}
	if qpmObj, ok := config["minute"]; ok {
		if qpm, ok := qpmObj.(float64); ok {
			res["qpm"] = qpm
		} else {
			return nil, errors.Errorf("qpm from config[%+v] transfer failed", config)
		}
	}
	if qphObj, ok := config["hour"]; ok {
		if qph, ok := qphObj.(float64); ok {
			res["qph"] = qph
		} else {
			return nil, errors.Errorf("qph from config[%+v] transfer failed", config)
		}
	}
	if qpdObj, ok := config["day"]; ok {
		if qpd, ok := qpdObj.(float64); ok {
			res["qpd"] = qpd
		} else {
			return nil, errors.Errorf("qpd from config[%+v] transfer failed", config)
		}
	}
	return res, nil
}

func (impl GatewayCategoryServiceImpl) CreatePolicy(category string, createDto *gw.PolicyCreateDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	ret := UNKNOW_ERROR
	var err error
	var az string
	var consumer *orm.GatewayConsumer
	verifyRes := impl.verifyPolicyParam(category, createDto)
	if !verifyRes.Success {
		return res.SetErrorInfo(verifyRes.Err)
	}
	categoryEnum, ok := verifyRes.Data.(*PolicyCategory)
	if !ok {
		err = errors.New("transfer to PolicyCategory failed")
		goto errorHappened
	}
	if createDto.OrgId == 0 || len(createDto.ProjectId) == 0 {
		log.Errorf("invalid createDto[%+v]", createDto)
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	az, err = impl.azDb.GetAz(&orm.GatewayAzInfo{
		OrgId:     strconv.Itoa(createDto.OrgId),
		ProjectId: createDto.ProjectId,
		Env:       createDto.Env,
	})
	if err != nil {
		log.Error(err)
		return res.SetReturnCode(CLUSTER_NOT_EXIST)
	}
	consumer, err = impl.consumerDb.GetDefaultConsumer(&orm.GatewayConsumer{
		OrgId:     strconv.Itoa(createDto.OrgId),
		ProjectId: createDto.ProjectId,
		Env:       createDto.Env,
		Az:        az,
	})
	if err != nil {
		err = errors.WithStack(err)
		goto errorHappened
	}
	if consumer == nil {
		err = errors.New("consumer is nil")
		ret = CONSUMER_NOT_EXIST
		goto errorHappened
	}
	{
		var policy = new(orm.GatewayPolicy)
		var config map[string]interface{}
		var configJson []byte
		policy, err = impl.policyDb.GetByPolicyName(createDto.PolicyName, consumer.Id)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
		if policy != nil {
			err = errors.Errorf("gatewayPolicy[%+v] exist", policy)
			ret = POLICY_EXIST
			goto errorHappened
		}
		config, err = impl.buildConfig(createDto.Config)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
		config["CARRIER"] = categoryEnum.Carrier
		configJson, err = json.Marshal(config)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
		policy = &orm.GatewayPolicy{
			PolicyName:  createDto.PolicyName,
			DisplayName: createDto.DisplayName,
			Category:    categoryEnum.Name,
			Description: categoryEnum.CnName,
			PluginName:  categoryEnum.Plugin,
			Config:      configJson,
			ConsumerId:  consumer.Id,
		}
		err = impl.policyDb.Insert(policy)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
		return res.SetSuccessAndData(&gw.PolicyCreateRespDto{PolicyId: policy.Id})
	}
errorHappened:
	log.Errorf("error happened:%+v", err)
	return res.SetReturnCode(ret)
}
func (impl GatewayCategoryServiceImpl) UpdatePolicy(policyId string, category string, createDto *gw.PolicyCreateDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	ret := UNKNOW_ERROR
	var err error
	var policy *orm.GatewayPolicy
	var config map[string]interface{}
	verifyRes := impl.verifyPolicyParam(category, createDto)
	if !verifyRes.Success {
		return res.SetErrorInfo(verifyRes.Err)
	}
	categoryEnum, ok := verifyRes.Data.(*PolicyCategory)
	if !ok {
		err = errors.New("transfer to PolicyCategory failed")
		goto errorHappened
	}
	if len(policyId) == 0 {
		err = errors.New("policyId is nil")
		ret = PARAMS_IS_NULL
		goto errorHappened
	}
	policy, err = impl.policyDb.GetById(policyId)
	if err != nil {
		err = errors.WithStack(err)
		goto errorHappened
	}
	if policy == nil {
		err = errors.New("policy is nil")
		ret = POLICY_NOT_EXIST
		goto errorHappened
	}
	{
		var configJson []byte
		var plugins []orm.GatewayPluginInstance
		config, err = impl.buildConfig(createDto.Config)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
		config["CARRIER"] = categoryEnum.Carrier
		configJson, err = json.Marshal(config)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
		policy.Config = configJson
		policy.PolicyName = createDto.PolicyName
		policy.DisplayName = createDto.DisplayName
		plugins, err = impl.pluginDb.SelectByPolicyId(policyId)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
		for _, oldPlugin := range plugins {
			var route *orm.GatewayRoute
			var service *orm.GatewayService
			var gatewayApi *orm.GatewayApi
			consumer := &orm.GatewayConsumer{}
			route, err = impl.routeDb.GetById(oldPlugin.RouteId)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
			if route == nil {
				log.Errorf("route[%s] already deleted, but policy[%s] relies",
					oldPlugin.RouteId, policyId)
				continue
			}
			service, err = impl.serviceDb.GetById(oldPlugin.ServiceId)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
			if service == nil {
				log.Errorf("service[%s] alreadt deleted, but policy[%s] relies",
					oldPlugin.ServiceId, policyId)
				continue
			}
			gatewayApi, err = impl.apiDb.GetById(service.ApiId)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
			if gatewayApi == nil {
				log.Errorf("api[%s] alreadt deleted, but service[%s] relies",
					service.ApiId, service.Id)
				continue

			}
			if len(oldPlugin.ConsumerId) != 0 {
				consumer, _ = impl.consumerDb.GetById(oldPlugin.ConsumerId)
				if consumer == nil {
					consumer = &orm.GatewayConsumer{}
				}
			}
			var pluginReq *kongDto.KongPluginReqDto
			var pluginResp *kongDto.KongPluginRespDto
			pluginReq, err = impl.kongAssembler.BuildKongPluginReqDto(oldPlugin.PluginId, policy, service.ServiceId, route.RouteId, consumer.ConsumerId)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
			pluginResp, err = kong.NewKongAdapterByConsumerId(impl.consumerDb, gatewayApi.ConsumerId).PutPlugin(pluginReq)
			if err != nil {
				log.Errorf("skip update since error:%s", errors.WithStack(err))
				continue
			}
			if pluginResp.Id != oldPlugin.PluginId {
				oldPlugin.PluginId = pluginResp.Id
				_ = impl.pluginDb.Update(&oldPlugin)
			}
		}
		err = impl.policyDb.Update(policy)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
	}
	{
		var detailConfig map[string]interface{}
		detailConfig, err = impl.buildDetail(config)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
		policyDto := &gw.PolicyDto{
			Category:    policy.Category,
			PolicyId:    policy.Id,
			DisplayName: policy.DisplayName,
			CreateAt:    policy.CreateTime.Format("2006-01-02T15:04:05"),
			Config:      detailConfig,
			PolicyName:  policy.PolicyName,
		}
		return res.SetSuccessAndData(policyDto)
	}
errorHappened:
	log.Errorf("error happened:%+v", err)
	return res.SetReturnCode(ret)

}
func (impl GatewayCategoryServiceImpl) DeletePolicy(policyId string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	ret := UNKNOW_ERROR
	var err error
	var plugins []orm.GatewayPluginInstance
	var policy *orm.GatewayPolicy
	if len(policyId) == 0 {
		err = errors.New("empty policyId")
		ret = PARAMS_IS_NULL
		goto errorHappened
	}
	policy, err = impl.policyDb.GetById(policyId)
	if err != nil {
		err = errors.WithStack(err)
		goto errorHappened
	}
	if policy == nil {
		err = errors.New("policy is nil")
		ret = POLICY_NOT_EXIST
		goto errorHappened
	}
	if plugins, err = impl.pluginDb.SelectByPolicyId(policyId); err != nil {
		err = errors.WithStack(err)
		goto errorHappened
	} else if len(plugins) != 0 {
		err = errors.Errorf("plugins of policyId[%s] is not empty", policyId)
		ret = DELETE_POLICY_FAIL
		goto errorHappened
	}
	err = impl.policyDb.DeleteById(policyId)
	if err != nil {
		err = errors.WithStack(err)
		goto errorHappened
	}
	return res.SetSuccessAndData(true)
errorHappened:
	log.Errorf("error happened:%+v", err)
	return res.SetReturnCode(ret)
}
func (impl GatewayCategoryServiceImpl) GetCategoryInfo(category string, orgId string, projectId string, env string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	ret := UNKNOW_ERROR
	var err error
	var az string
	var consumer *orm.GatewayConsumer
	categoryEnum := GetPolicyCategory(category)
	if len(category) == 0 || len(orgId) == 0 || len(projectId) == 0 {
		ret = PARAMS_IS_NULL
		err = errors.Errorf("invalid args category[%s] orgId[%s] projectId[%s]",
			category, orgId, projectId)
		goto errorHappened
	}
	az, err = impl.azDb.GetAz(&orm.GatewayAzInfo{
		OrgId:     orgId,
		ProjectId: projectId,
		Env:       env,
	})
	if err != nil {
		log.Error(err)
		return res.SetReturnCode(CLUSTER_NOT_EXIST)
	}
	consumer, err = impl.consumerDb.GetDefaultConsumer(&orm.GatewayConsumer{
		OrgId:     orgId,
		ProjectId: projectId,
		Env:       env,
		Az:        az,
	})
	if err != nil {
		err = errors.WithStack(err)
		goto errorHappened
	}
	if consumer == nil {
		err = errors.New("consumer is nil")
		ret = CONSUMER_NOT_EXIST
		goto errorHappened
	}
	{
		var policies []orm.GatewayPolicy
		if categoryEnum == nil {
			log.Info("categoryEnum is nil")
			policies, err = impl.policyDb.SelectByCategory(category)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
		} else {
			policies, err = impl.policyDb.SelectByCategoryAndConsumer(category, consumer.Id)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
		}
		categoryInfoDto := gw.CategoryInfoDto{}
		if len(policies) > 0 {
			categoryInfoDto.Category = category
			categoryInfoDto.Description = policies[0].Description
			policyDtos := []gw.PolicyDto{}
			var detailConfig map[string]interface{}
			for _, policy := range policies {
				config := map[string]interface{}{}
				err = json.Unmarshal([]byte(policy.Config), &config)
				if err != nil {
					err = errors.WithStack(err)
					goto errorHappened
				}
				detailConfig, err = impl.buildDetail(config)
				if err != nil {
					err = errors.WithStack(err)
					goto errorHappened
				}
				policyDto := gw.PolicyDto{
					Category:    policy.Category,
					PolicyId:    policy.Id,
					DisplayName: policy.DisplayName,
					CreateAt:    policy.CreateTime.Format("2006-01-02T15:04:05"),
					PolicyName:  policy.PolicyName,
					Config:      detailConfig,
				}
				policyDtos = append(policyDtos, policyDto)
			}
			categoryInfoDto.PolicyList = policyDtos
		}
		return res.SetSuccessAndData(categoryInfoDto)
	}

errorHappened:
	log.Errorf("error happened:%+v", err)
	return res.SetReturnCode(ret)

}
