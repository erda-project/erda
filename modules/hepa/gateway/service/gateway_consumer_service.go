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
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/hepa/common"
	"github.com/erda-project/erda/modules/hepa/common/util"
	. "github.com/erda-project/erda/modules/hepa/common/vars"
	"github.com/erda-project/erda/modules/hepa/config"
	"github.com/erda-project/erda/modules/hepa/gateway/assembler"
	gw "github.com/erda-project/erda/modules/hepa/gateway/dto"
	"github.com/erda-project/erda/modules/hepa/kong"
	kongDto "github.com/erda-project/erda/modules/hepa/kong/dto"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
	db "github.com/erda-project/erda/modules/hepa/repository/service"
)

type GatewayConsumerServiceImpl struct {
	consumerDb     db.GatewayConsumerService
	policyDb       db.GatewayPolicyService
	apiDb          db.GatewayApiService
	consumerApiDb  db.GatewayConsumerApiService
	azDb           db.GatewayAzInfoService
	kongDb         db.GatewayKongInfoService
	consumerApiBiz GatewayConsumerApiService
	kongAssembler  assembler.GatewayKongAssembler
	dbAssembler    assembler.GatewayDbAssembler
	globalBiz      GatewayGlobalService
}

func NewGatewayConsumerServiceImpl() (*GatewayConsumerServiceImpl, error) {
	consumerDb, _ := db.NewGatewayConsumerServiceImpl()
	policyDb, _ := db.NewGatewayPolicyServiceImpl()
	apiDb, _ := db.NewGatewayApiServiceImpl()
	azDb, _ := db.NewGatewayAzInfoServiceImpl()
	kongDb, _ := db.NewGatewayKongInfoServiceImpl()
	consumerApiDb, _ := db.NewGatewayConsumerApiServiceImpl()
	consumerApiBiz, _ := NewGatewayConsumerApiServiceImpl()
	globalBiz, _ := NewGatewayGlobalServiceImpl()

	return &GatewayConsumerServiceImpl{
		consumerDb:     consumerDb,
		policyDb:       policyDb,
		apiDb:          apiDb,
		consumerApiDb:  consumerApiDb,
		azDb:           azDb,
		kongDb:         kongDb,
		consumerApiBiz: consumerApiBiz,
		kongAssembler:  assembler.GatewayKongAssemblerImpl{},
		dbAssembler:    assembler.GatewayDbAssemblerImpl{},
		globalBiz:      globalBiz,
	}, nil
}

func (impl GatewayConsumerServiceImpl) CreateDefaultConsumer(orgId, projectId, env, az string) (*orm.GatewayConsumer, *orm.ConsumerAuthConfig, StandardErrorCode, error) {
	return impl.createConsumer(orgId, projectId, env, az,
		impl.consumerDb.GetDefaultConsumerName(&orm.GatewayConsumer{
			OrgId:     orgId,
			ProjectId: projectId,
			Env:       env,
			Az:        az,
		}))
}

func (impl GatewayConsumerServiceImpl) getCredentialList(kongAdapter kong.KongAdapter, consumerId string) (map[string]kongDto.KongCredentialListDto, error) {
	kCredentials, err := kongAdapter.GetCredentialList(consumerId, orm.KEYAUTH)
	if err != nil {
		kCredentials = &kongDto.KongCredentialListDto{}
	}
	oCredentials, err := kongAdapter.GetCredentialList(consumerId, orm.OAUTH2)
	if err != nil {
		oCredentials = &kongDto.KongCredentialListDto{}
	}
	return map[string]kongDto.KongCredentialListDto{
		orm.KEYAUTH: *kCredentials,
		orm.OAUTH2:  *oCredentials,
	}, nil
}

func (impl GatewayConsumerServiceImpl) UpdateConsumerInfo(consumerId string, consumerInfo *gw.ConsumerDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	consumer, err := impl.consumerDb.GetById(consumerId)
	if err != nil || consumer == nil {
		log.Errorf("get consumer failed, err:%+v", err)
		return res.SetReturnCode(CONSUMER_NOT_EXIST)
	}
	kongAdapter := kong.NewKongAdapterForConsumer(consumer)
	credentialListMap, err := impl.getCredentialList(kongAdapter, consumer.ConsumerId)
	if err != nil {
		log.Errorf("get credential list failed, err:%+v", err)
		return res
	}
	newAuth := consumerInfo.AuthConfig
	adds := map[string][]kongDto.KongCredentialDto{}
	dels := map[string][]kongDto.KongCredentialDto{}
	for _, item := range newAuth.Auths {
		var oldCredentials []kongDto.KongCredentialDto
		credentialList := item.AuthData
		oldCredentialList, ok := credentialListMap[item.AuthType]
		if !ok {
			log.Warnf("can't find auth type[%s], need adds", item.AuthType)
			adds[item.AuthType] = credentialList.Data
			continue
		}
		oldCredentials = oldCredentialList.Data

		for _, credential := range credentialList.Data {
			needAdd := true
			for i := len(oldCredentials) - 1; i >= 0; i-- {
				if oldCredentials[i].Id == credential.Id {
					needAdd = false
					oldCredentials = append(oldCredentials[:i], oldCredentials[i+1:]...)
					break
				}
			}
			if needAdd {
				adds[item.AuthType] = append(adds[item.AuthType], credential)
			}
		}
		dels[item.AuthType] = oldCredentials
	}
	for authType, credentials := range adds {
		for _, credential := range credentials {
			if authType == orm.OAUTH2 {
				if credential.RedirectUrl == nil {
					credential.RedirectUrl = []string{"http://none"}
				} else {
					if url, ok := credential.RedirectUrl.(string); ok {
						if !strings.HasPrefix(url, "http") {
							url = "http://" + url
						}
						credential.RedirectUrl = []string{url}
					} else if urls, ok := credential.RedirectUrl.([]string); ok {
						for i := 0; i < len(urls); i++ {
							if urls[i] == "" {
								urls[i] = "http://none"
							} else if urls[i] == "http://" || urls[i] == "https://" {
								urls[i] += "none"
							}
						}
					}
				}
			}
			_, err = impl.createCredential(kongAdapter, authType, consumer.ConsumerId, &credential)
			if err != nil {
				log.Errorf("create credential failed, err:%+v", err)
				cstr, _ := json.Marshal(credential)
				return res.SetErrorInfo(&common.ErrInfo{
					Code: "创建失败",
					Msg: fmt.Sprintf("创建凭证失败,请检查是否已存在,凭证类型:%s, 凭证内容:%s",
						authType, cstr),
				})
			}
		}
	}
	for authType, credentials := range dels {
		for _, credential := range credentials {
			err = kongAdapter.DeleteCredential(consumer.ConsumerId, authType, credential.Id)
			if err != nil {
				log.Errorf("delete credential failed, err:%+v", err)
				return res
			}
		}
	}
	credentialListMap, err = impl.getCredentialList(kongAdapter, consumer.ConsumerId)
	if err != nil {
		log.Errorf("get credential list failed, err:%+v", err)
		return res
	}
	dto := &gw.ConsumerDto{
		ConsumerCredentialsDto: gw.ConsumerCredentialsDto{
			ConsumerId:   consumerId,
			ConsumerName: consumer.ConsumerName,
			AuthConfig: &orm.ConsumerAuthConfig{
				Auths: []orm.AuthItem{
					{
						AuthTips: orm.KeyAuthTips,
						AuthType: orm.KEYAUTH,
						AuthData: credentialListMap[orm.KEYAUTH],
					},
					{
						AuthType: orm.OAUTH2,
						AuthData: credentialListMap[orm.OAUTH2],
					},
				},
			},
		},
	}
	return res.SetSuccessAndData(dto)
}

func (impl GatewayConsumerServiceImpl) createConsumer(orgId, projectId, env, az, consumerName string) (*orm.GatewayConsumer, *orm.ConsumerAuthConfig, StandardErrorCode, error) {
	ret := UNKNOW_ERROR
	consumer, err := impl.consumerDb.GetByName(consumerName)
	var respDto *kongDto.KongConsumerRespDto
	var kongAdapter kong.KongAdapter
	if err != nil {
		err = errors.WithStack(err)
		goto errorHappened
	}
	if consumer != nil {
		ret = CONSUMER_EXIST
		err = errors.Errorf("consumer[%s] alreay exist", consumerName)
		goto errorHappened
	}
	{
		var customId string
		customId, _ = util.GenUniqueId()
		var gatewayConsumer *orm.GatewayConsumer
		var keyAuth, oauth2 *kongDto.KongCredentialDto
		var consumerAuthConfig *orm.ConsumerAuthConfig
		reqDto := &kongDto.KongConsumerReqDto{
			Username: consumerName,
			CustomId: customId,
		}
		kongAdapter = kong.NewKongAdapterForProject(az, env, projectId)
		if !kongAdapter.KongExist() {
			err = errors.Errorf("no kong in az[%s]", az)
			ret = KONG_NOT_EXIST
			goto errorHappened
		}
		respDto, err = kongAdapter.CreateConsumer(reqDto)
		if err != nil {
			ret = CONSUMER_EXIST
			goto errorHappened
		}
		gatewayConsumer = &orm.GatewayConsumer{
			BaseRow:      orm.BaseRow{Id: customId},
			OrgId:        orgId,
			ProjectId:    projectId,
			Env:          env,
			Az:           az,
			ConsumerId:   respDto.Id,
			ConsumerName: consumerName,
		}
		keyAuth, err = impl.createCredential(kongAdapter, orm.KEYAUTH, respDto.Id, nil)
		if err != nil {
			goto errorHappened
		}
		oauth2, err = impl.createCredential(kongAdapter, orm.OAUTH2, respDto.Id,
			&kongDto.KongCredentialDto{
				Name:        "App",
				RedirectUrl: []string{"http://none"},
			})
		if err != nil {
			goto errorHappened
		}
		consumerAuthConfig = &orm.ConsumerAuthConfig{
			Auths: []orm.AuthItem{
				{
					AuthTips: orm.KeyAuthTips,
					AuthType: orm.KEYAUTH,
					AuthData: kongDto.KongCredentialListDto{
						Total: 1,
						Data:  []kongDto.KongCredentialDto{*keyAuth},
					},
				},
				{
					AuthType: orm.OAUTH2,
					AuthData: kongDto.KongCredentialListDto{
						Total: 1,
						Data:  []kongDto.KongCredentialDto{*oauth2},
					},
				},
			},
		}
		err = impl.consumerDb.Insert(gatewayConsumer)
		if err != nil {
			goto errorHappened
		}
		err = kongAdapter.CreateAclGroup(respDto.Id, customId)
		if err != nil {
			goto errorHappened
		}
		return gatewayConsumer, consumerAuthConfig, ret, nil
	}
errorHappened:
	if respDto != nil {
		_ = kongAdapter.DeleteConsumer(respDto.Id)
	}
	return nil, nil, ret, err
}

func (impl GatewayConsumerServiceImpl) CreateConsumer(createDto *gw.ConsumerCreateDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	var ret StandardErrorCode
	var err error = nil
	var orgId, az, consumerName string
	var gatewayConsumer *orm.GatewayConsumer
	var consumerAuthConfig *orm.ConsumerAuthConfig
	if createDto == nil || createDto.IsEmpty() {
		ret = PARAMS_IS_NULL
		err = errors.Errorf("invalid createDto[%+v]", createDto)
		goto errorHappened
	}
	if createDto.ConsumerName == "default" {
		ret = CONSUMER_EXIST
		goto errorHappened
	}
	orgId = strconv.Itoa(createDto.OrgId)
	az, err = impl.azDb.GetAz(&orm.GatewayAzInfo{
		OrgId:     orgId,
		ProjectId: createDto.ProjectId,
		Env:       createDto.Env,
	})
	if err != nil {
		ret = CLUSTER_NOT_EXIST
		goto errorHappened
	}
	consumerName = orgId + "_" + createDto.ProjectId + "_" + createDto.Env + "_" + az + "_" + createDto.ConsumerName
	gatewayConsumer, consumerAuthConfig, ret, err = impl.createConsumer(orgId, createDto.ProjectId, createDto.Env, az, consumerName)
	if err != nil {
		goto errorHappened
	}
	return res.SetSuccessAndData(&gw.CreateConsumerResp{
		ConsumerId:   gatewayConsumer.Id,
		ConsumerName: gatewayConsumer.ConsumerName,
		AuthConfig:   consumerAuthConfig,
	})
errorHappened:
	log.Errorf("error happened:%+v", err)
	return res.SetReturnCode(ret)
}

func (impl GatewayConsumerServiceImpl) createCredential(kongAdapter kong.KongAdapter, pluginName string, consumerId string, config *kongDto.KongCredentialDto) (*kongDto.KongCredentialDto, error) {
	req := &kongDto.KongCredentialReqDto{}
	req.ConsumerId = consumerId
	req.PluginName = pluginName
	req.Config = config
	return kongAdapter.CreateCredential(req)
}

func (impl GatewayConsumerServiceImpl) GetConsumerInfo(consumerId string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	consumer, err := impl.consumerDb.GetById(consumerId)
	if err != nil || consumer == nil {
		log.Errorf("get consumer failed, err:%+v", err)
		return res.SetReturnCode(CONSUMER_NOT_EXIST)
	}
	kongAdapter := kong.NewKongAdapterForConsumer(consumer)
	credentialListMap, err := impl.getCredentialList(kongAdapter, consumer.ConsumerId)
	if err != nil {
		log.Errorf("get credential list failed, err:%+v", err)
		return res
	}
	dto := &gw.ConsumerDto{
		ConsumerCredentialsDto: gw.ConsumerCredentialsDto{
			ConsumerId:   consumerId,
			ConsumerName: consumer.ConsumerName,
			AuthConfig: &orm.ConsumerAuthConfig{
				Auths: []orm.AuthItem{
					{
						AuthTips: orm.KeyAuthTips,
						AuthType: orm.KEYAUTH,
						AuthData: credentialListMap[orm.KEYAUTH],
					},
					{
						AuthType: orm.OAUTH2,
						AuthData: credentialListMap[orm.OAUTH2],
					},
				},
			},
		},
	}
	return res.SetSuccessAndData(dto)
}

func (impl GatewayConsumerServiceImpl) GetProjectConsumerInfo(orgId string, projectId string, env string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	ret := UNKNOW_ERROR
	var err error
	if len(orgId) == 0 || len(projectId) == 0 {
		err = errors.Errorf("invalid args: orgId[%s] projectId[%s]", orgId, projectId)
		ret = PARAMS_IS_NULL
		goto errorHappened
	}
	{
		var az string
		var consumer = new(orm.GatewayConsumer)
		var dto *gw.ConsumerDto
		var kongInfo *orm.GatewayKongInfo
		var outerAddr string
		az, err = impl.azDb.GetAz(&orm.GatewayAzInfo{
			OrgId:     orgId,
			ProjectId: projectId,
			Env:       env,
		})
		if err != nil {
			ret = CLUSTER_NOT_EXIST
			goto errorHappened
		}
		if az == "" {
			err = errors.New("empty az")
			ret = CLUSTER_NOT_EXIST
			goto errorHappened
		}
		consumer, err = impl.consumerDb.GetDefaultConsumer(&orm.GatewayConsumer{
			OrgId:     orgId,
			ProjectId: projectId,
			Env:       env,
			Az:        az,
		})
		if err != nil {
			goto errorHappened
		}
		if consumer == nil {
			consumer, _, ret, err = impl.CreateDefaultConsumer(orgId, projectId, env, az)
			if err != nil {
				goto errorHappened
			}
		}
		kongInfo, err = impl.kongDb.GetKongInfo(&orm.GatewayKongInfo{
			Az:        az,
			ProjectId: projectId,
			Env:       env,
		})
		if err != nil {
			goto errorHappened
		}
		outerAddr = kongInfo.Endpoint
		if !strings.EqualFold(env, ENV_TYPE_PROD) {
			outerAddr = strings.ToLower(env + config.ServerConf.SubDomainSplit + outerAddr)
		}
		dto = &gw.ConsumerDto{
			ConsumerCredentialsDto: gw.ConsumerCredentialsDto{
				ConsumerId:   consumer.Id,
				ConsumerName: consumer.ConsumerName,
			},
			Endpoint: gw.EndPoint{
				OuterAddr: outerAddr,
				InnerAddr: impl.globalBiz.GetServiceAddr(consumer.Env),
				InnerTips: "",
			},
			GatewayInstance: kongInfo.AddonInstanceId,
			ClusterName:     az,
		}
		return res.SetSuccessAndData(dto)
	}

errorHappened:
	log.Errorf("error happened:%+v", err)
	return res.SetReturnCode(ret)

}

func (impl GatewayConsumerServiceImpl) UpdateConsumerApi(dto *gw.ConsumerEditDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	ret := UNKNOW_ERROR
	var err error
	var consumer *orm.GatewayConsumer
	var consumerApiList []orm.GatewayConsumerApi
	consumerApis := map[string]string{}
	if dto == nil || len(dto.ConsumerId) == 0 {
		err = errors.Errorf("invalid dto[%+v]", dto)
		ret = PARAMS_IS_NULL
		goto errorHappened
	}
	consumer, err = impl.consumerDb.GetById(dto.ConsumerId)
	if err != nil {
		err = errors.WithStack(err)
		goto errorHappened
	}
	if consumer == nil {
		err = errors.New("consumer is nil")
		ret = CONSUMER_NOT_EXIST
		goto errorHappened
	}
	consumerApiList, err = impl.consumerApiDb.SelectByConsumer(dto.ConsumerId)
	if err != nil {
		err = errors.WithStack(err)
		goto errorHappened
	}
	for _, consumerApi := range consumerApiList {
		contains := false
		var cApi *orm.GatewayConsumerApi
		for i := len(dto.ApiList) - 1; i >= 0; i-- {
			apiId := dto.ApiList[i]
			if apiId == consumerApi.ApiId {
				contains = true
				cApi, err = impl.consumerApiDb.GetByConsumerAndApi(dto.ConsumerId, apiId)
				if err != nil {
					log.Errorf("get consumer api failed consumer[%s] api[%s]", dto.ConsumerId, apiId)
				} else {
					dto.ApiList = append(dto.ApiList[:i], dto.ApiList[i+1:]...)
					consumerApis[apiId] = cApi.Id
				}
				break
			}
		}
		if !contains {
			err = impl.consumerApiBiz.Delete(consumerApi.Id)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
			continue
		}
	}
	for _, apiId := range util.UniqStringSlice(dto.ApiList) {
		var consumerApiId string
		consumerApiId, err = impl.consumerApiBiz.Create(dto.ConsumerId, apiId)
		consumerApis[apiId] = consumerApiId
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
	}
	err = impl.consumerDb.Update(consumer)
	if err != nil {
		err = errors.WithStack(err)
		goto errorHappened
	}
	return res.SetSuccessAndData(gw.ConsumerEditRespDto{ConsumerApis: consumerApis})
errorHappened:
	log.Errorf("error happened:%+v", err)
	return res.SetReturnCode(ret)
}

func (impl GatewayConsumerServiceImpl) GetConsumerList(orgId string, projectId string, env string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	ret := UNKNOW_ERROR
	var err error
	var az, defaultName string
	var cond *orm.GatewayConsumer
	consumerAllResp := &gw.ConsumerAllResp{}
	var consumerList []orm.GatewayConsumer
	consumerDtos := []gw.ConsumerInfoDto{}
	if len(orgId) == 0 || len(projectId) == 0 || len(env) == 0 {
		err = errors.New("invalid arg")
		ret = CONSUMER_PARAMS_MISS
		goto errorHappened
	}
	az, err = impl.azDb.GetAz(&orm.GatewayAzInfo{
		OrgId:     orgId,
		ProjectId: projectId,
		Env:       env,
	})
	if err != nil {
		ret = CLUSTER_NOT_EXIST
		goto errorHappened
	}
	cond = &orm.GatewayConsumer{
		OrgId:     orgId,
		ProjectId: projectId,
		Env:       env,
		Az:        az,
	}
	consumerList, err = impl.consumerDb.SelectByAny(cond)
	if err != nil {
		goto errorHappened
	}
	defaultName = impl.consumerDb.GetDefaultConsumerName(cond)
	for _, consumer := range consumerList {
		var consumerDto *gw.ConsumerInfoDto
		if consumer.ConsumerName == defaultName {
			continue
		}
		consumerDto, err = impl.dbAssembler.BuildConsumerInfo(&consumer)
		var consumerApis []orm.GatewayConsumerApi
		consumerApiDtos := []gw.ConsumerApiInfoDto{}
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
		consumerApis, err = impl.consumerApiDb.SelectByConsumer(consumer.Id)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
		if len(consumerApis) == 0 {
			consumerDtos = append(consumerDtos, *consumerDto)
			continue
		}
		for _, consumerApi := range consumerApis {
			var api *orm.GatewayApi
			var consumerApiDto *gw.ConsumerApiInfoDto
			api, err = impl.apiDb.GetById(consumerApi.ApiId)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
			if api == nil {
				log.Warnf("api of id[%s] is nil", consumerApi.ApiId)
				continue
			}
			consumerApiDto, err = impl.dbAssembler.BuildConsumerApiInfo(&consumerApi, api)
			if err != nil {
				err = errors.WithStack(err)
				goto errorHappened
			}
			if len(consumerApi.Policies) > 0 {
				policyList := []string{}
				err = json.Unmarshal([]byte(consumerApi.Policies), &policyList)
				if err != nil {
					err = errors.WithStack(err)
					goto errorHappened
				}
				if len(policyList) > 0 {
					var policiesList []orm.GatewayPolicy
					policiesList, err = impl.policyDb.SelectInIds(policyList...)
					if err != nil {
						err = errors.WithStack(err)
						goto errorHappened
					}
					for _, policy := range policiesList {
						if policy.Category == TRAFFIC_CONTROL.Name {
							var policyDto *gw.ConsumerApiPolicyInfoDto
							policyDto, err = impl.dbAssembler.BuildConsumerApiPolicyInfo(&policy)
							if err != nil {
								err = errors.WithStack(err)
								goto errorHappened
							}
							consumerApiDto.RateLimitPolicy = *policyDto

						}
					}
				}
			}
			consumerApiDtos = append(consumerApiDtos, *consumerApiDto)
		}
		consumerDto.ConsumerApiInfos = consumerApiDtos
		consumerDtos = append(consumerDtos, *consumerDto)
	}
	consumerAllResp.Consumers = consumerDtos
	return res.SetSuccessAndData(consumerAllResp)
errorHappened:
	log.Errorf("error happened:%+v", err)
	return res.SetReturnCode(ret)

}
func (impl GatewayConsumerServiceImpl) DeleteConsumer(consumerId string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	ret := UNKNOW_ERROR
	var err error
	var consumer *orm.GatewayConsumer
	var consumerApiList []orm.GatewayConsumerApi
	if len(consumerId) == 0 {
		ret = PARAMS_IS_NULL
		err = errors.New("consumerId is empty")
		goto errorHappened
	}
	consumer, err = impl.consumerDb.GetById(consumerId)
	if err != nil {
		err = errors.WithStack(err)
		goto errorHappened
	}
	if consumer == nil {
		ret = CONSUMER_NOT_EXIST
		err = errors.New("consumer is nil")
		goto errorHappened
	}
	consumerApiList, err = impl.consumerApiDb.SelectByConsumer(consumerId)
	if err != nil {
		err = errors.WithStack(err)
		goto errorHappened
	}
	for _, api := range consumerApiList {
		err = impl.consumerApiBiz.Delete(api.Id)
		if err != nil {
			err = errors.WithStack(err)
			goto errorHappened
		}
	}
	err = impl.consumerDb.DeleteById(consumerId)
	if err != nil {
		err = errors.WithStack(err)
		goto errorHappened
	}
	_ = kong.NewKongAdapterForConsumer(consumer).DeleteConsumer(consumer.ConsumerId)
	return res.SetSuccessAndData(true)
errorHappened:
	log.Errorf("error happened:%+v", err)
	return res.SetReturnCode(ret)
}
