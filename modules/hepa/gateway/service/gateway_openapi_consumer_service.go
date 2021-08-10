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
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cloudapi"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/hepa/bundle"
	"github.com/erda-project/erda/modules/hepa/common"
	"github.com/erda-project/erda/modules/hepa/common/util"
	"github.com/erda-project/erda/modules/hepa/common/vars"
	. "github.com/erda-project/erda/modules/hepa/common/vars"
	gw "github.com/erda-project/erda/modules/hepa/gateway/dto"
	"github.com/erda-project/erda/modules/hepa/kong"
	kongDto "github.com/erda-project/erda/modules/hepa/kong/dto"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
	db "github.com/erda-project/erda/modules/hepa/repository/service"
	"github.com/erda-project/erda/pkg/crypto/uuid"
)

type GatewayOpenapiConsumerServiceImpl struct {
	packageDb      db.GatewayPackageService
	packageApiDb   db.GatewayPackageApiService
	consumerDb     db.GatewayConsumerService
	azDb           db.GatewayAzInfoService
	kongDb         db.GatewayKongInfoService
	packageInDb    db.GatewayPackageInConsumerService
	packageApiInDb db.GatewayPackageApiInConsumerService
	ruleBiz        GatewayOpenapiRuleService
	ReqCtx         *gin.Context
}

func NewGatewayOpenapiConsumerServiceImpl() (*GatewayOpenapiConsumerServiceImpl, error) {
	packageDb, _ := db.NewGatewayPackageServiceImpl()
	packageApiDb, _ := db.NewGatewayPackageApiServiceImpl()
	consumerDb, _ := db.NewGatewayConsumerServiceImpl()
	azDb, _ := db.NewGatewayAzInfoServiceImpl()
	kongDb, _ := db.NewGatewayKongInfoServiceImpl()
	packageInDb, _ := db.NewGatewayPackageInConsumerServiceImpl()
	packageApiInDb, _ := db.NewGatewayPackageApiInConsumerServiceImpl()
	ruleBiz, _ := NewGatewayOpenapiRuleServiceImpl()
	return &GatewayOpenapiConsumerServiceImpl{
		consumerDb:     consumerDb,
		azDb:           azDb,
		kongDb:         kongDb,
		packageInDb:    packageInDb,
		packageApiInDb: packageApiInDb,
		packageDb:      packageDb,
		packageApiDb:   packageApiDb,
		ruleBiz:        ruleBiz,
	}, nil
}

func (impl GatewayOpenapiConsumerServiceImpl) GetKongConsumerName(consumer *orm.GatewayConsumer) string {
	if consumer.ProjectId == "" {
		return consumer.ConsumerName
	}
	return fmt.Sprintf("%s.%s.%s.%s:%s", consumer.OrgId, consumer.ProjectId, consumer.Env, consumer.Az, consumer.ConsumerName)
}

func (impl GatewayOpenapiConsumerServiceImpl) getCredentialList(kongAdapter kong.KongAdapter, consumerId string) (map[string]kongDto.KongCredentialListDto, error) {
	kCredentials, err := kongAdapter.GetCredentialList(consumerId, orm.KEYAUTH)
	if err != nil {
		kCredentials = &kongDto.KongCredentialListDto{
			Data: []kongDto.KongCredentialDto{},
		}
	}
	oCredentials, err := kongAdapter.GetCredentialList(consumerId, orm.OAUTH2)
	if err != nil {
		oCredentials = &kongDto.KongCredentialListDto{
			Data: []kongDto.KongCredentialDto{},
		}
	}
	sCredentials, err := kongAdapter.GetCredentialList(consumerId, orm.SIGNAUTH)
	if err != nil {
		sCredentials = &kongDto.KongCredentialListDto{
			Data: []kongDto.KongCredentialDto{},
		}
	}
	hCredentials, err := kongAdapter.GetCredentialList(consumerId, orm.HMACAUTH)
	if err != nil {
		hCredentials = &kongDto.KongCredentialListDto{
			Data: []kongDto.KongCredentialDto{},
		}
	}
	return map[string]kongDto.KongCredentialListDto{
		orm.KEYAUTH:  *kCredentials,
		orm.OAUTH2:   *oCredentials,
		orm.SIGNAUTH: *sCredentials,
		orm.HMACAUTH: *hCredentials,
	}, nil
}

func (impl GatewayOpenapiConsumerServiceImpl) createCredential(kongAdapter kong.KongAdapter, pluginName string, consumerId string, config *kongDto.KongCredentialDto) (*kongDto.KongCredentialDto, error) {
	req := &kongDto.KongCredentialReqDto{}
	req.ConsumerId = consumerId
	req.PluginName = pluginName
	req.Config = config
	if pluginName == orm.HMACAUTH {
		enabled, err := kongAdapter.CheckPluginEnabled(pluginName)
		if err != nil {
			return nil, err
		}
		if !enabled {
			return &kongDto.KongCredentialDto{}, nil
		}
	}
	return kongAdapter.CreateCredential(req)
}

func (impl GatewayOpenapiConsumerServiceImpl) CreateClientConsumer(clientName, clientId, clientSecret, clusterName string) (consumer *orm.GatewayConsumer, err error) {
	dao, err := impl.consumerDb.GetByAny(&orm.GatewayConsumer{
		ConsumerName: clientName,
		Az:           clusterName,
		Type:         orm.APIM_CLIENT_CONSUMER,
	})
	if err != nil {
		return
	}
	if dao != nil {
		err = errors.Errorf("consumer:%s already exist in cluster:%s", clientName, clusterName)
		return
	}
	consumerId, err := util.GenUniqueId()
	if err != nil {
		return
	}
	consumer = &orm.GatewayConsumer{
		BaseRow:      orm.BaseRow{Id: consumerId},
		ClientId:     clientId,
		ConsumerName: clientName,
		Az:           clusterName,
		Type:         orm.APIM_CLIENT_CONSUMER,
	}
	kongInfo, err := impl.kongDb.GetKongInfo(&orm.GatewayKongInfo{
		Az: clusterName,
	})
	if err != nil {
		return
	}
	kongAdapter := kong.NewKongAdapter(kongInfo.KongAddr)
	reqDto := &kongDto.KongConsumerReqDto{
		Username: clientName,
		CustomId: consumerId,
	}
	respDto, err := kongAdapter.CreateConsumer(reqDto)
	if err != nil {
		return
	}
	consumer.ConsumerId = respDto.Id
	err = impl.consumerDb.Insert(consumer)
	if err != nil {
		return
	}
	err = kongAdapter.CreateAclGroup(respDto.Id, reqDto.Username)
	if err != nil {
		return
	}
	_, err = impl.createCredential(kongAdapter, orm.KEYAUTH, respDto.Id, &kongDto.KongCredentialDto{
		Key: clientId,
	})
	if err != nil {
		return
	}
	_, err = impl.createCredential(kongAdapter, orm.OAUTH2, respDto.Id,
		&kongDto.KongCredentialDto{
			Name:         clientName,
			RedirectUrl:  []string{"http://none"},
			ClientId:     clientId,
			ClientSecret: clientSecret,
		})
	if err != nil {
		return
	}
	_, err = impl.createCredential(kongAdapter, orm.SIGNAUTH, respDto.Id, &kongDto.KongCredentialDto{
		Key:    clientId,
		Secret: clientSecret,
	})
	if err != nil {
		return
	}
	_, err = impl.createCredential(kongAdapter, orm.HMACAUTH, respDto.Id, &kongDto.KongCredentialDto{
		Key:    clientId,
		Secret: clientSecret,
	})
	return
}

func (impl GatewayOpenapiConsumerServiceImpl) CreateConsumer(args *gw.DiceArgsDto, dto *gw.OpenConsumerDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if dto.Type == "" {
		dto.Type = gw.CT_PRO
	}
	if args.OrgId == "" || args.ProjectId == "" || args.Env == "" || dto.Name == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	var kongConsumerName string
	var consumer *orm.GatewayConsumer
	var unique bool
	var kongInfo *orm.GatewayKongInfo
	var kongAdapter kong.KongAdapter
	var reqDto *kongDto.KongConsumerReqDto
	var respDto *kongDto.KongConsumerRespDto
	var customId string
	var err error
	auditCtx := map[string]interface{}{}
	auditCtx["consumer"] = dto.Name
	key, _ := util.GenUniqueId()
	secret, _ := util.GenUniqueId()
	defer func() {
		audit := common.MakeAuditInfo(impl.ReqCtx, common.ScopeInfo{
			ProjectId: args.ProjectId,
			Workspace: args.Env,
		}, apistructs.CreateGatewayConsumerTemplate, res, auditCtx)
		if audit != nil {
			err = bundle.Bundle.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: *audit})
			if err != nil {
				log.Errorf("create audit failed, err:%+v", err)
			}
		}
	}()
	az, err := impl.azDb.GetAz(&orm.GatewayAzInfo{
		Env:       args.Env,
		OrgId:     args.OrgId,
		ProjectId: args.ProjectId,
	})
	if err != nil {
		goto failed
	}
	customId, err = util.GenUniqueId()
	if err != nil {
		goto failed
	}
	consumer = &orm.GatewayConsumer{
		BaseRow:      orm.BaseRow{Id: customId},
		ConsumerName: dto.Name,
		OrgId:        args.OrgId,
		ProjectId:    args.ProjectId,
		Env:          args.Env,
		Az:           az,
		Description:  dto.Description,
	}
	unique, err = impl.consumerDb.CheckUnique(consumer)
	if err != nil {
		goto failed
	}
	if !unique {
		res.SetReturnCode(CONSUMER_EXIST)
		err = errors.Errorf("consumer %s already exist", dto.Name)
		goto failed
	}
	kongInfo, err = impl.kongDb.GetKongInfo(&orm.GatewayKongInfo{
		Az:        az,
		ProjectId: args.ProjectId,
		Env:       args.Env,
	})
	if err != nil {
		goto failed
	}
	kongAdapter = kong.NewKongAdapter(kongInfo.KongAddr)
	kongConsumerName = impl.GetKongConsumerName(consumer)
	reqDto = &kongDto.KongConsumerReqDto{
		Username: kongConsumerName,
		CustomId: customId,
	}
	respDto, err = kongAdapter.CreateConsumer(reqDto)
	if err != nil {
		res.SetReturnCode(CONSUMER_EXIST)
		goto failed
	}
	consumer.ConsumerId = respDto.Id
	err = impl.consumerDb.Insert(consumer)
	if err != nil {
		goto failed
	}
	err = kongAdapter.CreateAclGroup(respDto.Id, kongConsumerName)
	if err != nil {
		goto failed
	}
	_, err = impl.createCredential(kongAdapter, orm.KEYAUTH, respDto.Id,
		&kongDto.KongCredentialDto{
			Key: key,
		})
	if err != nil {
		goto failed
	}
	_, err = impl.createCredential(kongAdapter, orm.OAUTH2, respDto.Id,
		&kongDto.KongCredentialDto{
			Name:         "App",
			RedirectUrl:  []string{"http://none"},
			ClientId:     key,
			ClientSecret: secret,
		})
	if err != nil {
		goto failed
	}
	_, err = impl.createCredential(kongAdapter, orm.SIGNAUTH, respDto.Id,
		&kongDto.KongCredentialDto{
			Key:    key,
			Secret: secret,
		})
	if err != nil {
		goto failed
	}
	_, err = impl.createCredential(kongAdapter, orm.HMACAUTH, respDto.Id,
		&kongDto.KongCredentialDto{
			Key:    key,
			Secret: secret,
		})
	if err != nil {
		goto failed
	}
	return res.SetSuccessAndData(consumer.Id)
failed:
	log.Errorf("error happened, err:%+v", err)
	return res

}

func (impl GatewayOpenapiConsumerServiceImpl) openConsumerDto(dao *orm.GatewayConsumer) *gw.OpenConsumerInfoDto {
	return &gw.OpenConsumerInfoDto{
		Id:       dao.Id,
		CreateAt: dao.CreateTime.Format("2006-01-02T15:04:05"),
		OpenConsumerDto: gw.OpenConsumerDto{
			Name:        dao.ConsumerName,
			Description: dao.Description,
		},
	}
}

func (impl GatewayOpenapiConsumerServiceImpl) GetConsumers(args *gw.GetOpenConsumersDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if args.OrgId == "" || args.ProjectId == "" || args.Env == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	options := args.GenSelectOptions()
	pageInfo := common.NewPage2(args.PageSize, args.PageNo)
	var list []gw.OpenConsumerInfoDto
	var daos []orm.GatewayConsumer
	var ok bool
	var defaultName string
	var az string
	page, err := impl.consumerDb.GetPage(options, pageInfo)
	if err != nil {
		goto failed
	}
	daos, ok = page.Result.([]orm.GatewayConsumer)
	if !ok {
		err = errors.New("type convert failed")
		goto failed
	}
	az, err = impl.azDb.GetAz(&orm.GatewayAzInfo{
		Env:       args.Env,
		OrgId:     args.OrgId,
		ProjectId: args.ProjectId,
	})
	if err != nil {
		goto failed
	}
	defaultName = impl.consumerDb.GetDefaultConsumerName(&orm.GatewayConsumer{
		OrgId:     args.OrgId,
		ProjectId: args.ProjectId,
		Env:       args.Env,
		Az:        az,
	})
	for _, dao := range daos {
		if strings.HasSuffix(strings.ToLower(dao.ConsumerName), strings.ToLower(defaultName)) {
			continue
		}
		list = append(list, *impl.openConsumerDto(&dao))
	}
	return res.SetSuccessAndData(common.NewPages(list, pageInfo.TotalNum))

failed:
	log.Errorf("error happened, err:%+v", err)
	return res

}

func (impl GatewayOpenapiConsumerServiceImpl) GetConsumersName(args *gw.GetOpenConsumersDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if args.OrgId == "" || args.ProjectId == "" || args.Env == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	options := args.GenSelectOptions()
	daos, err := impl.consumerDb.SelectByOptions(options)
	list := []gw.OpenConsumerInfoDto{}
	var defaultName string
	var az string
	if err != nil {
		goto failed
	}
	az, err = impl.azDb.GetAz(&orm.GatewayAzInfo{
		Env:       args.Env,
		OrgId:     args.OrgId,
		ProjectId: args.ProjectId,
	})
	if err != nil {
		goto failed
	}
	defaultName = impl.consumerDb.GetDefaultConsumerName(&orm.GatewayConsumer{
		OrgId:     args.OrgId,
		ProjectId: args.ProjectId,
		Env:       args.Env,
		Az:        az,
	})
	for _, dao := range daos {
		if dao.ConsumerName == defaultName {
			continue
		}
		list = append(list, *impl.openConsumerDto(&dao))
	}
	return res.SetSuccessAndData(list)
failed:
	log.Errorf("error happened, err:%+v", err)
	return res

}

func (impl GatewayOpenapiConsumerServiceImpl) UpdateConsumer(id string, dto *gw.OpenConsumerDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if id == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	dao, err := impl.consumerDb.GetById(id)
	if err != nil {
		goto failed
	}
	if dao == nil {
		err = errors.New("consumer not exist")
		goto failed
	}
	dao.Description = dto.Description
	err = impl.consumerDb.Update(dao)
	if err != nil {
		goto failed
	}
	return res.SetSuccessAndData(impl.openConsumerDto(dao))
failed:
	log.Errorf("error happened, err:%+v", err)
	return res

}

func (impl GatewayOpenapiConsumerServiceImpl) DeleteConsumer(id string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if id == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	var kongInfo *orm.GatewayKongInfo
	var kongAdapter kong.KongAdapter
	var consumer *orm.GatewayConsumer
	var err error
	auditCtx := map[string]interface{}{}
	defer func() {
		if consumer == nil {
			return
		}
		auditCtx["consumer"] = consumer.ConsumerName
		audit := common.MakeAuditInfo(impl.ReqCtx, common.ScopeInfo{
			ProjectId: consumer.ProjectId,
			Workspace: consumer.Env,
		}, apistructs.DeleteGatewayConsumerTemplate, res, auditCtx)
		if audit != nil {
			err = bundle.Bundle.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: *audit})
			if err != nil {
				log.Errorf("create audit failed, err:%+v", err)
			}
		}
	}()

	consumer, err = impl.consumerDb.GetById(id)
	if err != nil {
		goto failed
	}
	if consumer == nil {
		return res.SetSuccessAndData(true)
	}
	err = impl.packageInDb.DeleteByConsumerId(id)
	if err != nil {
		goto failed
	}
	err = impl.packageApiInDb.DeleteByConsumerId(id)
	if err != nil {
		goto failed
	}
	kongInfo, err = impl.kongDb.GetKongInfo(&orm.GatewayKongInfo{
		Az:        consumer.Az,
		ProjectId: consumer.ProjectId,
		Env:       consumer.Env,
	})
	if err != nil {
		goto failed
	}
	kongAdapter = kong.NewKongAdapter(kongInfo.KongAddr)
	err = kongAdapter.DeleteConsumer(consumer.ConsumerId)
	if err != nil {
		goto failed
	}
	go func() {
		defer util.DoRecover()
		if consumer.CloudapiAppId != "" {
			req := cloudapi.CreateDeleteAppRequest()
			resp := cloudapi.CreateDeleteAppResponse()
			req.AppId = requests.Integer(consumer.CloudapiAppId)
			req.SecurityToken = uuid.UUID()
			err = bundle.Bundle.DoRemoteAliyunAction(consumer.OrgId, consumer.Az, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, req, resp)
			if err != nil {
				log.Errorf("error happened: %+v", err)
			}
		}
	}()
	err = impl.consumerDb.DeleteById(id)
	if err != nil {
		goto failed
	}
	return res.SetSuccessAndData(true)
failed:
	log.Errorf("error happened, err:%+v", err)
	return res

}

func (impl GatewayOpenapiConsumerServiceImpl) GetConsumerCredentials(id string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if id == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	var kongInfo *orm.GatewayKongInfo
	var kongAdapter kong.KongAdapter
	var credentialListMap map[string]kongDto.KongCredentialListDto
	consumer, err := impl.consumerDb.GetById(id)
	if err != nil {
		goto failed
	}
	if consumer == nil {
		err = errors.New("consumer not exist")
		goto failed
	}
	kongInfo, err = impl.kongDb.GetKongInfo(&orm.GatewayKongInfo{
		Az:        consumer.Az,
		ProjectId: consumer.ProjectId,
		Env:       consumer.Env,
	})
	if err != nil {
		goto failed
	}
	kongAdapter = kong.NewKongAdapter(kongInfo.KongAddr)
	credentialListMap, err = impl.getCredentialList(kongAdapter, consumer.ConsumerId)
	if err != nil {
		goto failed
	}
	return res.SetSuccessAndData(gw.ConsumerCredentialsDto{
		ConsumerName: consumer.ConsumerName,
		ConsumerId:   id,
		AuthConfig: &orm.ConsumerAuthConfig{
			Auths: []orm.AuthItem{
				{
					AuthTips: orm.KeyAuthTips,
					AuthType: orm.KEYAUTH,
					AuthData: credentialListMap[orm.KEYAUTH],
				},
				{
					AuthTips: orm.SignAuthTips,
					AuthType: orm.SIGNAUTH,
					AuthData: credentialListMap[orm.SIGNAUTH],
				},
				{
					AuthType: orm.OAUTH2,
					AuthData: credentialListMap[orm.OAUTH2],
				},
				{
					AuthType: orm.HMACAUTH,
					AuthData: credentialListMap[orm.HMACAUTH],
				},
			},
		},
	})
failed:
	log.Errorf("error happened, err:%+v", err)
	return res

}

func (impl GatewayOpenapiConsumerServiceImpl) UpdateConsumerCredentials(id string, dto *gw.ConsumerCredentialsDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if id == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	var kongInfo *orm.GatewayKongInfo
	var consumerDto *gw.ConsumerCredentialsDto
	var kongAdapter kong.KongAdapter
	var credentialListMap map[string]kongDto.KongCredentialListDto
	newAuth := dto.AuthConfig
	adds := map[string][]kongDto.KongCredentialDto{}
	dels := map[string][]kongDto.KongCredentialDto{}
	var consumer *orm.GatewayConsumer
	var err error
	auditCtx := map[string]interface{}{}
	defer func() {
		if consumer == nil {
			return
		}
		auditCtx["consumer"] = consumer.ConsumerName
		audit := common.MakeAuditInfo(impl.ReqCtx, common.ScopeInfo{
			ProjectId: consumer.ProjectId,
			Workspace: consumer.Env,
		}, apistructs.UpdateGatewayConsumerTemplate, res, auditCtx)
		if audit != nil {
			err = bundle.Bundle.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: *audit})
			if err != nil {
				log.Errorf("create audit failed, err:%+v", err)
			}
		}
	}()
	consumer, err = impl.consumerDb.GetById(id)
	if err != nil {
		goto failed
	}
	if consumer == nil {
		err = errors.New("consumer not exist")
		goto failed
	}
	kongInfo, err = impl.kongDb.GetKongInfo(&orm.GatewayKongInfo{
		Az:        consumer.Az,
		ProjectId: consumer.ProjectId,
		Env:       consumer.Env,
	})
	if err != nil {
		goto failed
	}
	kongAdapter = kong.NewKongAdapter(kongInfo.KongAddr)
	credentialListMap, err = impl.getCredentialList(kongAdapter, consumer.ConsumerId)
	if err != nil {
		goto failed
	}
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
	for authType, credentials := range dels {
		for _, credential := range credentials {
			err = kongAdapter.DeleteCredential(consumer.ConsumerId, authType, credential.Id)
			if err != nil {
				goto failed
			}
		}
	}
	for authType, credentials := range adds {
		var hmacEnabled bool
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
			if authType == orm.HMACAUTH {
				hmacEnabled, err = kongAdapter.CheckPluginEnabled(orm.HMACAUTH)
				if err != nil {
					goto failed
				}
				if !hmacEnabled {
					log.Errorf("create hmac credential ignored, since not enabled")
					continue
				}
			}
			_, err = impl.createCredential(kongAdapter, authType, consumer.ConsumerId, &credential)
			if err != nil {
				log.Errorf("create credential failed, err:%+v", err)
				cstr, _ := json.Marshal(credential)
				return res.SetErrorInfo(&common.ErrInfo{
					Msg: fmt.Sprintf("create credential failed, already used, type:%s, content:%s",
						authType, cstr),
				})
			}
		}
	}
	credentialListMap, err = impl.getCredentialList(kongAdapter, consumer.ConsumerId)
	if err != nil {
		goto failed
	}
	consumerDto = &gw.ConsumerCredentialsDto{
		ConsumerId:   id,
		ConsumerName: consumer.ConsumerName,
		AuthConfig: &orm.ConsumerAuthConfig{
			Auths: []orm.AuthItem{
				{
					AuthTips: orm.KeyAuthTips,
					AuthType: orm.KEYAUTH,
					AuthData: credentialListMap[orm.KEYAUTH],
				},
				{
					AuthTips: orm.SignAuthTips,
					AuthType: orm.SIGNAUTH,
					AuthData: credentialListMap[orm.SIGNAUTH],
				},
				{
					AuthType: orm.OAUTH2,
					AuthData: credentialListMap[orm.OAUTH2],
				},
				{
					AuthType: orm.HMACAUTH,
					AuthData: credentialListMap[orm.HMACAUTH],
				},
			},
		},
	}
	return res.SetSuccessAndData(consumerDto)

failed:
	log.Errorf("error happened, err:%+v", err)
	return res

}

func (impl GatewayOpenapiConsumerServiceImpl) GetPackageApiAcls(packageId string, packageApiId string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	var packageRes *common.StandardResult
	var consumers []orm.GatewayConsumer
	var defaultName string
	var pack *orm.GatewayPackage
	result := []gw.PackageAclInfoDto{}
	selected := []gw.PackageAclInfoDto{}
	unselect := []gw.PackageAclInfoDto{}
	selectMap := map[string]bool{}
	var apiAclRules []gw.OpenapiRuleInfo
	if packageId == "" || packageApiId == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	consumerIn, err := impl.packageApiInDb.SelectByPackageApi(packageId, packageApiId)
	if err != nil {
		goto failed
	}
	apiAclRules, err = impl.ruleBiz.GetApiRules(packageApiId, gw.ACL_RULE)
	if err != nil {
		goto failed
	}
	packageRes = impl.GetPackageAcls(packageId)
	if len(apiAclRules) == 0 {
		return packageRes
	}
	pack, err = impl.packageDb.Get(packageId)
	if err != nil {
		goto failed
	}
	if pack == nil {
		err = errors.New("package not exist")
		goto failed
	}
	consumers, err = impl.consumerDb.SelectByAny(&orm.GatewayConsumer{
		OrgId:     pack.DiceOrgId,
		ProjectId: pack.DiceProjectId,
		Env:       pack.DiceEnv,
		Az:        pack.DiceClusterName,
	})
	if err != nil {
		goto failed
	}
	consumers = impl.adjustConsumers(pack, consumers)
	for _, in := range consumerIn {
		selectMap[in.ConsumerId] = true
	}
	defaultName = impl.consumerDb.GetDefaultConsumerName(&orm.GatewayConsumer{
		OrgId:     pack.DiceOrgId,
		ProjectId: pack.DiceProjectId,
		Env:       pack.DiceEnv,
		Az:        pack.DiceClusterName,
	})
	log.Debugf("default consumer name: %s", defaultName) // output for debug
	for _, consumer := range consumers {
		if strings.HasSuffix(strings.ToLower(consumer.ConsumerName), strings.ToLower(defaultName)) {
			continue
		}
		dto := gw.PackageAclInfoDto{
			OpenConsumerInfoDto: gw.OpenConsumerInfoDto{
				OpenConsumerDto: gw.OpenConsumerDto{
					Name:        consumer.ConsumerName,
					Description: consumer.Description,
				},
				Id:       consumer.Id,
				CreateAt: consumer.CreateTime.Format("2006-01-02T15:04:05"),
			},
		}
		if _, exist := selectMap[consumer.Id]; exist {
			dto.Selected = true
			selected = append(selected, dto)
		} else {
			dto.Selected = false
			unselect = append(unselect, dto)
		}
	}
	result = append(result, selected...)
	result = append(result, unselect...)
	return res.SetSuccessAndData(result)

failed:
	log.Errorf("error happened, err:%+v", err)
	return res.SetErrorInfo(&common.ErrInfo{
		Msg: errors.Cause(err).Error(),
	})

}

func (impl GatewayOpenapiConsumerServiceImpl) adjustConsumers(pack *orm.GatewayPackage, consumers []orm.GatewayConsumer) []orm.GatewayConsumer {
	if pack.AuthType != gw.AT_ALIYUN_APP {
		return consumers
	}
	for i := len(consumers) - 1; i >= 0; i-- {
		if consumers[i].CloudapiAppId == "" {
			consumers = append(consumers[:i], consumers[i+1:]...)
		}
	}
	return consumers

}

func (impl GatewayOpenapiConsumerServiceImpl) GetPackageAcls(id string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if id == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	var consumers []orm.GatewayConsumer
	var consumerIn []orm.GatewayPackageInConsumer
	var defaultName string
	result := []gw.PackageAclInfoDto{}
	selected := []gw.PackageAclInfoDto{}
	unselect := []gw.PackageAclInfoDto{}
	selectMap := map[string]bool{}
	pack, err := impl.packageDb.Get(id)
	if err != nil {
		goto failed
	}
	if pack == nil {
		err = errors.New("package not exist")
		goto failed
	}
	consumers, err = impl.consumerDb.SelectByAny(&orm.GatewayConsumer{
		OrgId:     pack.DiceOrgId,
		ProjectId: pack.DiceProjectId,
		Env:       pack.DiceEnv,
		Az:        pack.DiceClusterName,
	})
	if err != nil {
		goto failed
	}
	consumers = impl.adjustConsumers(pack, consumers)
	consumerIn, err = impl.packageInDb.SelectByPackage(id)
	if err != nil {
		goto failed
	}
	for _, in := range consumerIn {
		selectMap[in.ConsumerId] = true
	}
	defaultName = impl.consumerDb.GetDefaultConsumerName(&orm.GatewayConsumer{
		OrgId:     pack.DiceOrgId,
		ProjectId: pack.DiceProjectId,
		Env:       pack.DiceEnv,
		Az:        pack.DiceClusterName,
	})
	log.Debugf("default consumer name: %s", defaultName) // output for debug
	for _, consumer := range consumers {
		if strings.HasSuffix(strings.ToLower(consumer.ConsumerName), strings.ToLower(defaultName)) {
			continue
		}
		dto := gw.PackageAclInfoDto{
			OpenConsumerInfoDto: gw.OpenConsumerInfoDto{
				OpenConsumerDto: gw.OpenConsumerDto{
					Name:        consumer.ConsumerName,
					Description: consumer.Description,
				},
				Id:       consumer.Id,
				CreateAt: consumer.CreateTime.Format("2006-01-02T15:04:05"),
			},
		}
		if _, exist := selectMap[consumer.Id]; exist {
			dto.Selected = true
			selected = append(selected, dto)
		} else {
			dto.Selected = false
			unselect = append(unselect, dto)
		}
	}
	result = append(result, selected...)
	result = append(result, unselect...)
	return res.SetSuccessAndData(result)
failed:
	log.Errorf("error happened, err:%+v", err)
	return res

}

func (impl GatewayOpenapiConsumerServiceImpl) UpdatePackageApiAcls(packageId, packageApiId string, dto *gw.PackageAclsDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if packageId == "" || packageApiId == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	var consumerIn []orm.GatewayPackageApiInConsumer
	var api *orm.GatewayPackageApi
	var consumer *orm.GatewayConsumer
	selectMap := map[string]bool{}
	pack, err := impl.packageDb.Get(packageId)
	if err != nil {
		goto failed
	}
	if pack == nil {
		err = errors.New("package not exist")
		goto failed
	}
	api, err = impl.packageApiDb.Get(packageApiId)
	if err != nil {
		goto failed
	}
	if api == nil {
		err = errors.New("package api not exist")
		goto failed
	}
	consumerIn, err = impl.packageApiInDb.SelectByPackageApi(packageId, packageApiId)
	if err != nil {
		goto failed
	}
	for _, in := range consumerIn {
		consumer, err = impl.consumerDb.GetById(in.ConsumerId)
		if err != nil {
			goto failed
		}
		if consumer != nil && consumer.Type == orm.APIM_CLIENT_CONSUMER {
			continue
		}
		selectMap[in.ConsumerId] = false
	}
	for _, consumerId := range dto.Consumers {
		if _, selected := selectMap[consumerId]; selected {
			selectMap[consumerId] = true
			continue
		}
		go func(consumerId string) {
			defer util.DoRecover()
			err = impl.grantCloudapi(consumerId, pack, api)
			if err != nil {
				log.Errorf("error happened: %+v", err)
			}
		}(consumerId)
		err = impl.packageApiInDb.Insert(&orm.GatewayPackageApiInConsumer{
			ConsumerId:   consumerId,
			PackageId:    packageId,
			PackageApiId: packageApiId,
		})
		if err != nil {
			goto failed
		}

	}
	for consumerId, selected := range selectMap {
		if !selected {
			go func(consumerId string) {
				defer util.DoRecover()
				err = impl.cancelCloudapi(consumerId, pack, api)
				if err != nil {
					log.Errorf("error happened: %+v", err)
				}
			}(consumerId)
			err = impl.packageApiInDb.Delete(packageId, packageApiId, consumerId)
			if err != nil {
				goto failed
			}
		}
	}
	err = impl.touchPackageApiAclRules(packageId, packageApiId)
	if err != nil {
		goto failed
	}
	return res.SetSuccessAndData(true)
failed:
	log.Errorf("error happened, err:%+v", err)
	return res

}

func (impl GatewayOpenapiConsumerServiceImpl) touchPackageApiAclRules(packageId, packageApiId string) error {
	pack, err := impl.packageDb.Get(packageId)
	if err != nil {
		return err
	}
	if pack == nil {
		return errors.New("package not found")
	}
	consumers, err := impl.GetConsumersOfPackageApi(packageId, packageApiId)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	for _, consumer := range consumers {
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(impl.GetKongConsumerName(&consumer))
	}
	wl := buffer.String()
	config := map[string]interface{}{}
	if wl == "" {
		wl = ","
	}
	config["whitelist"] = wl
	aclRules, err := impl.ruleBiz.GetApiRules(packageApiId, gw.ACL_RULE)
	if err != nil {
		return err
	}
	for _, rule := range aclRules {
		rule.Config = config
		_, err = impl.ruleBiz.UpdateRule(rule.Id, &rule.OpenapiRule)
		if err != nil {
			return err
		}
	}
	if len(aclRules) == 0 {
		newAclRule := &gw.OpenapiRule{
			PackageId:    packageId,
			PackageApiId: packageApiId,
			PluginName:   gw.ACL,
			Category:     gw.ACL_RULE,
			Config:       config,
			Enabled:      true,
			Region:       gw.API_RULE,
		}
		err = impl.ruleBiz.CreateRule(DiceInfo{
			ProjectId: pack.DiceProjectId,
			Env:       pack.DiceEnv,
			Az:        pack.DiceClusterName,
		}, newAclRule, nil)
		if err != nil {
			return err
		}
		err = impl.ruleBiz.SetPackageKongPolicies(pack, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl GatewayOpenapiConsumerServiceImpl) getGroupApis(orgId string, pack *orm.GatewayPackage) ([]string, error) {
	var apis []string
	if pack.AuthType != gw.AT_ALIYUN_APP {
		return nil, nil
	}
	resp := cloudapi.CreateDescribeApisResponse()
	var pageCount, pageNum int
	for {
		pageNum++
		req := cloudapi.CreateDescribeApisRequest()
		req.GroupId = pack.CloudapiGroupId
		req.PageSize = requests.NewInteger(100)
		req.PageNumber = requests.NewInteger(pageNum)
		err := bundle.Bundle.DoRemoteAliyunAction(orgId, pack.DiceClusterName, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, req, resp)
		if err != nil {
			return nil, err
		}
		if resp.TotalCount == 0 {
			break
		}
		for _, api := range resp.ApiSummarys.ApiSummary {
			apis = append(apis, api.ApiId)
		}
		if pageCount == 0 {
			pageCount = (resp.TotalCount + 99) / 100
		}
		if pageNum == pageCount {
			break
		}
	}
	return apis, nil

}

func (impl GatewayOpenapiConsumerServiceImpl) splitApis(apis []string, size int) [][]string {
	var res [][]string
	var curSlice []string
	for _, api := range apis {
		if len(curSlice) == size {
			res = append(res, curSlice)
			curSlice = []string{}
			continue
		}
		curSlice = append(curSlice, api)
	}
	return res
}

func (impl GatewayOpenapiConsumerServiceImpl) grantCloudapiGroup(consumerId string, pack *orm.GatewayPackage) error {
	consumer, err := impl.consumerDb.GetById(consumerId)
	if err != nil {
		return err
	}
	if consumer == nil {
		err = errors.New("consumer not exist")
		return err
	}
	if pack.AuthType != gw.AT_ALIYUN_APP || consumer.CloudapiAppId == "" {
		return nil
	}
	apis, err := impl.getGroupApis(consumer.OrgId, pack)
	if err != nil {
		return err
	}
	slices := impl.splitApis(apis, 100)
	for _, apiSlice := range slices {
		req := cloudapi.CreateSetApisAuthoritiesRequest()
		req.AppId = requests.Integer(consumer.CloudapiAppId)
		req.ApiIds = strings.Join(apiSlice, ",")
		req.GroupId = pack.CloudapiGroupId
		req.StageName = "RELEASE"
		resp := cloudapi.CreateSetApisAuthoritiesResponse()
		err = bundle.Bundle.DoRemoteAliyunAction(consumer.OrgId, pack.DiceClusterName, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, req, resp)
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl GatewayOpenapiConsumerServiceImpl) cancelCloudapiGroup(consumerId string, pack *orm.GatewayPackage) error {
	consumer, err := impl.consumerDb.GetById(consumerId)
	if err != nil {
		return err
	}
	if consumer == nil {
		err = errors.New("consumer not exist")
		return err
	}
	if pack.AuthType != gw.AT_ALIYUN_APP || consumer.CloudapiAppId == "" {
		return nil
	}
	apis, err := impl.getGroupApis(consumer.OrgId, pack)
	if err != nil {
		return err
	}
	slices := impl.splitApis(apis, 100)
	for _, apiSlice := range slices {
		req := cloudapi.CreateRemoveApisAuthoritiesRequest()
		req.AppId = requests.Integer(consumer.CloudapiAppId)
		req.ApiIds = strings.Join(apiSlice, ",")
		req.GroupId = pack.CloudapiGroupId
		req.StageName = "RELEASE"
		resp := cloudapi.CreateRemoveApisAuthoritiesResponse()
		err = bundle.Bundle.DoRemoteAliyunAction(consumer.OrgId, pack.DiceClusterName, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, req, resp)
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl GatewayOpenapiConsumerServiceImpl) grantCloudapi(consumerId string, pack *orm.GatewayPackage, api *orm.GatewayPackageApi) error {
	consumer, err := impl.consumerDb.GetById(consumerId)
	if err != nil {
		return err
	}
	if consumer == nil {
		err = errors.New("consumer not exist")
		return err
	}
	if pack.AuthType != gw.AT_ALIYUN_APP || consumer.CloudapiAppId == "" || api.CloudapiApiId == "" {
		return nil
	}
	req := cloudapi.CreateSetApisAuthoritiesRequest()
	req.AppId = requests.Integer(consumer.CloudapiAppId)
	req.ApiIds = api.CloudapiApiId
	req.GroupId = pack.CloudapiGroupId
	req.StageName = "RELEASE"
	resp := cloudapi.CreateSetApisAuthoritiesResponse()
	err = bundle.Bundle.DoRemoteAliyunAction(consumer.OrgId, pack.DiceClusterName, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, req, resp)
	if err != nil {
		return err
	}

	return nil
}

func (impl GatewayOpenapiConsumerServiceImpl) cancelCloudapi(consumerId string, pack *orm.GatewayPackage, api *orm.GatewayPackageApi) error {
	consumer, err := impl.consumerDb.GetById(consumerId)
	if err != nil {
		return err
	}
	if consumer == nil {
		err = errors.New("consumer not exist")
		return err
	}
	if pack.AuthType != gw.AT_ALIYUN_APP || consumer.CloudapiAppId == "" || api.CloudapiApiId == "" {
		return nil
	}
	req := cloudapi.CreateRemoveApisAuthoritiesRequest()
	req.AppId = requests.Integer(consumer.CloudapiAppId)
	req.ApiIds = api.CloudapiApiId
	req.GroupId = pack.CloudapiGroupId
	req.StageName = "RELEASE"
	resp := cloudapi.CreateRemoveApisAuthoritiesResponse()
	err = bundle.Bundle.DoRemoteAliyunAction(consumer.OrgId, pack.DiceClusterName, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, req, resp)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayOpenapiConsumerServiceImpl) UpdatePackageAcls(id string, dto *gw.PackageAclsDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if id == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	var consumerIn []orm.GatewayPackageInConsumer
	var consumer *orm.GatewayConsumer
	selectMap := map[string]bool{}
	pack, err := impl.packageDb.Get(id)
	if err != nil {
		goto failed
	}
	if pack == nil {
		err = errors.New("package not exist")
		goto failed
	}
	consumerIn, err = impl.packageInDb.SelectByPackage(id)
	if err != nil {
		goto failed
	}
	for _, in := range consumerIn {
		consumer, err = impl.consumerDb.GetById(in.ConsumerId)
		if err != nil {
			goto failed
		}
		if consumer != nil && consumer.Type == orm.APIM_CLIENT_CONSUMER {
			continue
		}
		selectMap[in.ConsumerId] = false
	}
	for _, consumerId := range dto.Consumers {
		if _, selected := selectMap[consumerId]; selected {
			selectMap[consumerId] = true
			continue
		}
		err = impl.packageInDb.Insert(&orm.GatewayPackageInConsumer{
			ConsumerId: consumerId,
			PackageId:  id,
		})
		if err != nil {
			goto failed
		}
		go func(consumerId string) {
			defer util.DoRecover()
			err = impl.grantCloudapiGroup(consumerId, pack)
			if err != nil {
				log.Errorf("error happened: %+v", err)
			}
		}(consumerId)
	}
	for consumerId, selected := range selectMap {
		if !selected {
			err = impl.packageInDb.Delete(id, consumerId)
			if err != nil {
				goto failed
			}
			go func(consumerId string) {
				defer util.DoRecover()
				err = impl.cancelCloudapiGroup(consumerId, pack)
				if err != nil {
					log.Errorf("error happened: %+v", err)
				}
			}(consumerId)
		}
	}
	err = impl.updatePackageAclRules(id)
	if err != nil {
		goto failed
	}
	return res.SetSuccessAndData(true)
failed:
	log.Errorf("error happened, err:%+v", err)
	return res

}

func (impl GatewayOpenapiConsumerServiceImpl) adjustPackages(consumer *orm.GatewayConsumer, packages []orm.GatewayPackage) []orm.GatewayPackage {
	if consumer.CloudapiAppId != "" {
		return packages
	}
	for i := len(packages) - 1; i >= 0; i-- {
		if packages[i].AuthType == gw.AT_ALIYUN_APP {
			packages = append(packages[:i], packages[i+1:]...)
		}
	}
	return packages
}

func (impl GatewayOpenapiConsumerServiceImpl) GetConsumerAcls(id string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if id == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	var packages []orm.GatewayPackage
	var packageIn []orm.GatewayPackageInConsumer
	result := []gw.ConsumerAclInfoDto{}
	selected := []gw.ConsumerAclInfoDto{}
	unselect := []gw.ConsumerAclInfoDto{}
	selectMap := map[string]bool{}
	consumer, err := impl.consumerDb.GetById(id)
	if err != nil {
		goto failed
	}
	if consumer == nil {
		err = errors.New("consumer not exist")
		goto failed
	}
	packages, err = impl.packageDb.SelectByAny(&orm.GatewayPackage{
		DiceProjectId:   consumer.ProjectId,
		DiceEnv:         consumer.Env,
		DiceClusterName: consumer.Az,
		Scene:           orm.OPENAPI_SCENE,
	})
	if err != nil {
		goto failed
	}
	packages = impl.adjustPackages(consumer, packages)
	packageIn, err = impl.packageInDb.SelectByConsumer(id)
	if err != nil {
		goto failed
	}
	for _, in := range packageIn {
		selectMap[in.PackageId] = true
	}
	for _, pack := range packages {
		dto := gw.ConsumerAclInfoDto{
			PackageInfoDto: gw.PackageInfoDto{
				PackageDto: gw.PackageDto{
					Name:        pack.PackageName,
					Description: pack.Description,
				},
				Id:       pack.Id,
				CreateAt: pack.CreateTime.Format("2006-01-02T15:04:05"),
			},
		}
		if _, exist := selectMap[pack.Id]; exist {
			dto.Selected = true
			selected = append(selected, dto)
		} else {
			dto.Selected = false
			unselect = append(unselect, dto)
		}
	}
	result = append(result, selected...)
	result = append(result, unselect...)
	return res.SetSuccessAndData(result)
failed:
	log.Errorf("error happened, err:%+v", err)
	return res

}

func (impl GatewayOpenapiConsumerServiceImpl) updatePackageAclRules(packageId string) error {
	consumers, err := impl.GetConsumersOfPackage(packageId)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	for _, consumer := range consumers {
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(impl.GetKongConsumerName(&consumer))
	}
	wl := buffer.String()
	config := map[string]interface{}{}
	if wl == "" {
		wl = ","
	}
	config["whitelist"] = wl
	aclRules, err := impl.ruleBiz.GetPackageRules(packageId, nil, gw.ACL_RULE)
	if err != nil {
		return err
	}
	for _, rule := range aclRules {
		rule.Config = config
		_, err = impl.ruleBiz.UpdateRule(rule.Id, &rule.OpenapiRule)
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl GatewayOpenapiConsumerServiceImpl) GrantPackageToConsumer(consumerId, packageId string) error {
	pack, err := impl.packageDb.Get(packageId)
	if err != nil {
		return err
	}
	if pack == nil {
		return errors.New("package not found")
	}
	packageIn, err := impl.packageInDb.SelectByConsumer(consumerId)
	if err != nil {
		return err
	}
	for _, in := range packageIn {
		if in.PackageId == packageId {
			// granted
			return nil
		}
	}
	err = impl.packageInDb.Insert(&orm.GatewayPackageInConsumer{
		ConsumerId: consumerId,
		PackageId:  packageId,
	})
	if err != nil {
		return err
	}
	go func() {
		defer util.DoRecover()
		err = impl.grantCloudapiGroup(consumerId, pack)
		if err != nil {
			log.Errorf("error happened: %+v", err)
		}
	}()
	err = impl.updatePackageAclRules(packageId)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayOpenapiConsumerServiceImpl) RevokePackageFromConsumer(consumerId, packageId string) error {
	pack, err := impl.packageDb.Get(packageId)
	if err != nil {
		return err
	}
	if pack == nil {
		return errors.New("package not found")
	}
	packageIn, err := impl.packageInDb.SelectByConsumer(consumerId)
	if err != nil {
		return err
	}
	granted := false
	for _, in := range packageIn {
		if in.PackageId == packageId {
			granted = true
			break
		}
	}
	if !granted {
		return nil
	}
	err = impl.packageInDb.Delete(packageId, consumerId)
	if err != nil {
		return err
	}
	go func() {
		defer util.DoRecover()
		err = impl.cancelCloudapiGroup(consumerId, pack)
		if err != nil {
			log.Errorf("error happened: %+v", err)
		}
	}()
	err = impl.updatePackageAclRules(packageId)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayOpenapiConsumerServiceImpl) UpdateConsumerAcls(id string, dto *gw.ConsumerAclsDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if id == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	var packageIn []orm.GatewayPackageInConsumer
	selectMap := map[string]bool{}
	consumer, err := impl.consumerDb.GetById(id)
	if err != nil {
		goto failed
	}
	if consumer == nil {
		err = errors.New("consumer not exist")
		goto failed
	}
	packageIn, err = impl.packageInDb.SelectByConsumer(id)
	if err != nil {
		goto failed
	}
	for _, in := range packageIn {
		selectMap[in.PackageId] = false
	}
	for _, packageId := range dto.Packages {
		var pack *orm.GatewayPackage
		pack, err = impl.packageDb.Get(packageId)
		if err != nil {
			goto failed
		}
		if pack == nil {
			continue
		}
		if _, selected := selectMap[packageId]; selected {
			selectMap[packageId] = true
			continue
		}
		err = impl.packageInDb.Insert(&orm.GatewayPackageInConsumer{
			ConsumerId: consumer.Id,
			PackageId:  packageId,
		})
		if err != nil {
			goto failed
		}
		go func(consumerId string) {
			defer util.DoRecover()
			err = impl.grantCloudapiGroup(consumerId, pack)
			if err != nil {
				log.Errorf("error happened: %+v", err)
			}
		}(consumer.Id)
		err = impl.updatePackageAclRules(packageId)
		if err != nil {
			goto failed
		}
	}
	for packageId, selected := range selectMap {
		if !selected {
			var pack *orm.GatewayPackage
			pack, err = impl.packageDb.Get(packageId)
			if err != nil {
				goto failed
			}
			if pack == nil {
				continue
			}
			err = impl.packageInDb.Delete(packageId, consumer.Id)
			if err != nil {
				goto failed
			}
			go func(consumerId string) {
				defer util.DoRecover()
				err = impl.cancelCloudapiGroup(consumerId, pack)
				if err != nil {
					log.Errorf("error happened: %+v", err)
				}
			}(consumer.Id)
			err = impl.updatePackageAclRules(packageId)
			if err != nil {
				goto failed
			}
		}
	}
	return res.SetSuccessAndData(true)
failed:
	log.Errorf("error happened, err:%+v", err)
	return res

}

func (impl GatewayOpenapiConsumerServiceImpl) GetConsumersOfPackage(packageId string) ([]orm.GatewayConsumer, error) {
	consumerIds, err := impl.packageInDb.SelectByPackage(packageId)
	if err != nil {
		return nil, err
	}
	var consumers []orm.GatewayConsumer
	for _, consumerId := range consumerIds {
		consumer, err := impl.consumerDb.GetById(consumerId.ConsumerId)
		if err != nil {
			return nil, err
		}
		consumers = append(consumers, *consumer)
	}
	return consumers, nil
}

func (impl GatewayOpenapiConsumerServiceImpl) GetConsumersOfPackageApi(packageId, packageApiId string) ([]orm.GatewayConsumer, error) {
	consumerIds, err := impl.packageApiInDb.SelectByPackageApi(packageId, packageApiId)
	if err != nil {
		return nil, err
	}
	var consumers []orm.GatewayConsumer
	for _, consumerId := range consumerIds {
		consumer, err := impl.consumerDb.GetById(consumerId.ConsumerId)
		if err != nil {
			return nil, err
		}
		consumers = append(consumers, *consumer)
	}
	packageConsumers, err := impl.GetConsumersOfPackage(packageId)
	if err != nil {
		return nil, err
	}
	for _, consumer := range packageConsumers {
		if consumer.Type == orm.APIM_CLIENT_CONSUMER {
			consumers = append(consumers, consumer)
		}
	}
	return consumers, nil
}

// 获取阿里云APP凭证
func (impl GatewayOpenapiConsumerServiceImpl) GetCloudapiAppCredential(consumerId string) (res *common.StandardResult) {
	var err error
	res = &common.StandardResult{Success: false}
	defer func() {
		if err != nil {
			log.Errorf("error happened: %+v", err)
			res.SetErrorInfo(&common.ErrInfo{
				Msg: errors.Cause(err).Error(),
			})
		}
	}()
	consumer, err := impl.consumerDb.GetById(consumerId)
	if err != nil {
		return
	}
	if consumer == nil {
		err = errors.Errorf("consumer not found")
		return
	}
	if consumer.CloudapiAppId == "" {
		res.SetSuccessAndData([]gw.AliyunAppCredential{})
		return
	}
	req := cloudapi.CreateDescribeAppSecurityRequest()
	req.AppId = requests.Integer(consumer.CloudapiAppId)
	req.SecurityToken = uuid.UUID()
	resp := cloudapi.CreateDescribeAppSecurityResponse()
	err = bundle.Bundle.DoRemoteAliyunAction(consumer.OrgId, consumer.Az, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, req, resp)
	if err != nil {
		return
	}
	res.SetSuccessAndData([]gw.AliyunAppCredential{
		{
			AppKey:    resp.AppKey,
			AppSecret: resp.AppSecret,
		},
	})
	return
}

// 生成阿里云APP凭证
func (impl GatewayOpenapiConsumerServiceImpl) SetCloudapiAppCredential(consumerId string, async ...bool) (res *common.StandardResult) {
	var err error
	res = &common.StandardResult{Success: false}
	var needAsync bool
	if len(async) != 0 && async[0] {
		needAsync = true
	}
	defer func() {
		if err != nil {
			log.Errorf("error happened: %+v", err)
			res.SetErrorInfo(&common.ErrInfo{
				Msg: errors.Cause(err).Error(),
			})
		}
	}()
	existRes := impl.GetCloudapiAppCredential(consumerId)
	if !existRes.Success {
		err = errors.Errorf("get crendential failed, err:%+v", existRes.Err.Msg)
		return
	}
	if existData, ok := existRes.Data.([]gw.AliyunAppCredential); ok {
		if len(existData) != 0 {
			res.SetSuccessAndData(existData)
		}
	}
	consumer, err := impl.consumerDb.GetById(consumerId)
	if err != nil {
		return
	}
	if !needAsync {
		req := cloudapi.CreateCreateAppRequest()
		req.AppName = fmt.Sprintf("dice_%s", consumer.Id[:20])
		req.SecurityToken = uuid.UUID()
		req.Description = fmt.Sprintf("project:%s workspace:%s consumerName:%s", consumer.ProjectId, consumer.Env, consumer.ConsumerName)
		resp := cloudapi.CreateCreateAppResponse()
		err = bundle.Bundle.DoRemoteAliyunAction(consumer.OrgId, consumer.Az, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, req, resp)
		if err != nil {
			return
		}
		consumer.CloudapiAppId = strconv.FormatInt(resp.AppId, 10)
		err = impl.consumerDb.Update(consumer)
		if err != nil {
			return
		}
		time.Sleep(3 * time.Second)
		sreq := cloudapi.CreateDescribeAppSecurityRequest()
		sreq.AppId = requests.Integer(consumer.CloudapiAppId)
		sreq.SecurityToken = uuid.UUID()
		sresp := cloudapi.CreateDescribeAppSecurityResponse()
		err = bundle.Bundle.DoRemoteAliyunAction(consumer.OrgId, consumer.Az, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, sreq, sresp)
		if err != nil {
			return
		}
		res.SetSuccessAndData([]gw.AliyunAppCredential{
			{
				AppKey:    sresp.AppKey,
				AppSecret: sresp.AppSecret,
			},
		})
		return
	}
	res.SetSuccessAndData([]gw.AliyunAppCredential{})
	go func() {
		defer util.DoRecover()
		req := cloudapi.CreateCreateAppRequest()
		req.AppName = fmt.Sprintf("dice_%s", consumer.Id[:20])
		req.SecurityToken = uuid.UUID()
		req.Description = fmt.Sprintf("project:%s workspace:%s consumerName:%s", consumer.ProjectId, consumer.Env, consumer.ConsumerName)
		resp := cloudapi.CreateCreateAppResponse()
		verr := bundle.Bundle.DoRemoteAliyunAction(consumer.OrgId, consumer.Az, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, req, resp)
		if verr != nil {
			log.Errorf("error happened: %+v", verr)
			return
		}
		consumer.CloudapiAppId = strconv.FormatInt(resp.AppId, 10)
		verr = impl.consumerDb.Update(consumer)
		if verr != nil {
			log.Errorf("error happened: %+v", verr)
			return
		}
		sreq := cloudapi.CreateDescribeAppSecurityRequest()
		sreq.AppId = requests.Integer(consumer.CloudapiAppId)
		sreq.SecurityToken = uuid.UUID()
		sresp := cloudapi.CreateDescribeAppSecurityResponse()
		verr = bundle.Bundle.DoRemoteAliyunAction(consumer.OrgId, consumer.Az, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, sreq, sresp)
		if verr != nil {
			log.Errorf("error happened: %+v", verr)
			return
		}
	}()
	return

}

// 删除阿里云APP凭证
func (impl GatewayOpenapiConsumerServiceImpl) DeleteCloudapiAppCredential(consumerId string) (res *common.StandardResult) {
	var err error
	res = &common.StandardResult{Success: false}
	defer func() {
		if err != nil {
			log.Errorf("error happened: %+v", err)
			res.SetErrorInfo(&common.ErrInfo{
				Msg: errors.Cause(err).Error(),
			})
		}
	}()
	consumer, err := impl.consumerDb.GetById(consumerId)
	if err != nil {
		return
	}
	if consumer.CloudapiAppId == "" {
		res.SetSuccessAndData(true)
		return
	}
	req := cloudapi.CreateDeleteAppRequest()
	req.AppId = requests.Integer(consumer.CloudapiAppId)
	req.SecurityToken = uuid.UUID()
	resp := cloudapi.CreateDeleteAppResponse()
	err = bundle.Bundle.DoRemoteAliyunAction(consumer.OrgId, consumer.Az, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, req, resp)
	if err != nil {
		return
	}
	consumer.CloudapiAppId = ""
	err = impl.consumerDb.Update(consumer)
	if err != nil {
		return
	}
	res.SetSuccessAndData(true)
	return
}
