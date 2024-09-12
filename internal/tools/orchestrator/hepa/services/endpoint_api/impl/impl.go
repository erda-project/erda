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

package impl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/bundle"
	orgCache "github.com/erda-project/erda/internal/tools/orchestrator/hepa/cache/org"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/common"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/common/util"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/config"
	context1 "github.com/erda-project/erda/internal/tools/orchestrator/hepa/context"
	gateway_providers "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers"
	providerDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/kong"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse"
	mseCommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	mseDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/plugins"
	gw "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/hepautils"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/i18n"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/k8s"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	db "github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/api_policy"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/domain"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/endpoint_api"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/global"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/micro_api"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/openapi_consumer"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/openapi_rule"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/runtime_service"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/zone"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

const (
	K8S_SVC_CLUSTER_DOMAIN = ".svc.cluster.local"
)

type GatewayOpenapiServiceImpl struct {
	packageDb       db.GatewayPackageService
	packageApiDb    db.GatewayPackageApiService
	zoneInPackageDb db.GatewayZoneInPackageService
	apiInPackageDb  db.GatewayApiInPackageService
	packageInDb     db.GatewayPackageInConsumerService
	serviceDb       db.GatewayServiceService
	routeDb         db.GatewayRouteService
	consumerDb      db.GatewayConsumerService
	apiDb           db.GatewayApiService
	upstreamApiDb   db.GatewayUpstreamApiService
	azDb            db.GatewayAzInfoService
	kongDb          db.GatewayKongInfoService
	hubInfoDb       db.GatewayHubInfoService
	credentialDb    db.GatewayCredentialService
	apiBiz          *micro_api.GatewayApiService
	zoneBiz         *zone.GatewayZoneService
	ruleBiz         *openapi_rule.GatewayOpenapiRuleService
	consumerBiz     *openapi_consumer.GatewayOpenapiConsumerService
	globalBiz       *global.GatewayGlobalService
	policyBiz       *api_policy.GatewayApiPolicyService
	runtimeDb       db.GatewayRuntimeServiceService
	domainBiz       *domain.GatewayDomainService
	ctx             context.Context
	reqCtx          context.Context
}

var once sync.Once

func NewGatewayOpenapiServiceImpl() error {
	once.Do(
		func() {
			packageDb, _ := db.NewGatewayPackageServiceImpl()
			packageApiDb, _ := db.NewGatewayPackageApiServiceImpl()
			zoneInPackageDb, _ := db.NewGatewayZoneInPackageServiceImpl()
			apiInPackageDb, _ := db.NewGatewayApiInPackageServiceImpl()
			packageInDb, _ := db.NewGatewayPackageInConsumerServiceImpl()
			serviceDb, _ := db.NewGatewayServiceServiceImpl()
			routeDb, _ := db.NewGatewayRouteServiceImpl()
			consumerDb, _ := db.NewGatewayConsumerServiceImpl()
			apiDb, _ := db.NewGatewayApiServiceImpl()
			upstreamApiDb, _ := db.NewGatewayUpstreamApiServiceImpl()
			kongDb, _ := db.NewGatewayKongInfoServiceImpl()
			azDb, _ := db.NewGatewayAzInfoServiceImpl()
			runtimeDb, _ := db.NewGatewayRuntimeServiceServiceImpl()
			hubInfoDb, _ := db.NewGatewayHubInfoServiceImpl()
			credentialDb, _ := db.NewGatewayCredentialServiceImpl()
			endpoint_api.Service = &GatewayOpenapiServiceImpl{
				packageDb:       packageDb,
				packageApiDb:    packageApiDb,
				zoneInPackageDb: zoneInPackageDb,
				apiInPackageDb:  apiInPackageDb,
				packageInDb:     packageInDb,
				serviceDb:       serviceDb,
				routeDb:         routeDb,
				consumerDb:      consumerDb,
				apiDb:           apiDb,
				upstreamApiDb:   upstreamApiDb,
				kongDb:          kongDb,
				azDb:            azDb,
				hubInfoDb:       hubInfoDb,
				credentialDb:    credentialDb,
				zoneBiz:         &zone.Service,
				apiBiz:          &micro_api.Service,
				consumerBiz:     &openapi_consumer.Service,
				ruleBiz:         &openapi_rule.Service,
				globalBiz:       &global.Service,
				policyBiz:       &api_policy.Service,
				runtimeDb:       runtimeDb,
				domainBiz:       &domain.Service,
			}
		})
	return nil
}

func (impl GatewayOpenapiServiceImpl) Clone(ctx context.Context) endpoint_api.GatewayOpenapiService {
	newService := impl
	newService.reqCtx = ctx
	return &newService
}

func (impl GatewayOpenapiServiceImpl) createApiAclRule(aclType, packageId, apiId, clusterName string) (*gw.OpenapiRule, error) {
	rule, err := impl.createAclRule(aclType, packageId, clusterName)
	if err != nil {
		return nil, err
	}
	rule.PackageApiId = apiId
	rule.Region = gw.API_RULE
	return rule, nil
}

func (impl GatewayOpenapiServiceImpl) createAclRule(aclType, packageId, clusterName string) (*gw.OpenapiRule, error) {
	var consumers []orm.GatewayConsumer
	var aclRule *gw.OpenapiRule
	consumers, err := (*impl.consumerBiz).GetConsumersOfPackage(packageId)
	if err != nil {
		return nil, err
	}

	if clusterName == "" {
		err = errors.New("cluster name not provider")
		return nil, err
	}

	aclRule = &gw.OpenapiRule{
		PackageId:  packageId,
		PluginName: gw.ACL,
		Category:   gw.ACL_RULE,
	}

	gatewayProvider, err := impl.GetGatewayProvider(clusterName)
	if err != nil {
		return nil, errors.Errorf("can not detect gateway provider, error: %v", err)
	}

	config := make(map[string]interface{})
	switch gatewayProvider {
	case mseCommon.MseProviderName:
		wlConsumers, err := impl.mseConsumerConfig(consumers)
		if err != nil {
			return nil, err
		}
		if len(wlConsumers) == 0 {
			wlConsumers = append(wlConsumers, mseDto.Consumers{
				Name: plugins.MseDefaultConsumerName,
			})
		}
		config["whitelist"] = wlConsumers
	case "":
		var buffer bytes.Buffer
		for _, consumer := range consumers {
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			buffer.WriteString((*impl.consumerBiz).GetKongConsumerName(&consumer))
		}
		wl := buffer.String()
		if wl == "" {
			wl = ","
		}
		config["whitelist"] = wl
	default:
		return nil, errors.Errorf("unknown gateway provider %s", gatewayProvider)
	}

	aclRule.Config = config
	switch aclType {
	case gw.ACL_ON:
		aclRule.Enabled = true
	case gw.ACL_OFF:
		aclRule.Enabled = false
	default:
		return nil, errors.Errorf("invalid acl type:%s", aclType)
	}
	return aclRule, nil
}

func (impl GatewayOpenapiServiceImpl) mseConsumerConfig(consumers []orm.GatewayConsumer) ([]mseDto.Consumers, error) {
	wlConsumers := make([]mseDto.Consumers, 0)
	for _, consumer := range consumers {
		wlConsumer := mseDto.Consumers{
			Name: fmt.Sprintf("%s.%s.%s.%s:%s", consumer.OrgId, consumer.ProjectId, consumer.Env, consumer.Az, consumer.ConsumerName),
		}

		credentials, err := impl.credentialDb.SelectByConsumerId(consumer.ConsumerId)
		if err != nil {
			return wlConsumers, err
		}

		if len(credentials) == 0 {
			return wlConsumers, errors.Errorf("no credential info found for consumer %s", consumer.ConsumerName)
		}

		for _, credential := range credentials {
			switch credential.PluginName {
			case orm.OAUTH2:
				// TODO: MSE 暂不支持 Oauth2
			case orm.KEYAUTH:
				wlConsumer.Credential = credential.Key
			case orm.SIGNAUTH:
				wlConsumer.Key = credential.Key
				wlConsumer.Secret = credential.Secret
			case orm.HMACAUTH:
				wlConsumer.Key = credential.Key
				wlConsumer.Secret = credential.Secret
			case orm.MSEBasicAuth:
				wlConsumer.Credential = credential.Key
			case orm.MSEJWTAuth:
				if credential.FromParams != "" {
					wlConsumer.FromParams = strings.Split(credential.FromParams, ",")
				}
				if credential.FromCookies != "" {
					wlConsumer.FromCookies = strings.Split(credential.FromCookies, ",")
				}
				if credential.KeepToken == "N" {
					wlConsumer.KeepToken = false
				} else {
					wlConsumer.KeepToken = true
				}

				wlConsumer.ClockSkewSeconds = 60
				if credential.ClockSkewSeconds != "" {
					csSeconds, err := strconv.Atoi(credential.ClockSkewSeconds)
					if err == nil && csSeconds > 0 {
						wlConsumer.ClockSkewSeconds = csSeconds
					}
				}
			}
		}
		wlConsumers = append(wlConsumers, wlConsumer)
	}
	return wlConsumers, nil
}

func (impl GatewayOpenapiServiceImpl) createApiAuthRule(packageId, apiId string, enable ...bool) (*gw.OpenapiRule, error) {
	dao, err := impl.packageDb.Get(packageId)
	if err != nil {
		return nil, err
	}
	if dao == nil {
		return nil, errors.Errorf("endpoint not found, id:%s", packageId)
	}
	rule, err := impl.createAuthRule(dao.AuthType, dao, enable...)
	if err != nil {
		return nil, err
	}
	rule.PackageApiId = apiId
	rule.Region = gw.API_RULE
	return rule, nil
}

func (impl GatewayOpenapiServiceImpl) createAuthRule(authType string, pack *orm.GatewayPackage, enable ...bool) (*gw.OpenapiRule, error) {
	ruleEnabled := true
	if len(enable) > 0 {
		ruleEnabled = enable[0]
	}
	authRule := &gw.OpenapiRule{
		PackageId:  pack.Id,
		PluginName: authType,
		Category:   gw.AUTH_RULE,
		Enabled:    ruleEnabled,
	}

	gatewayProvider, err := impl.GetGatewayProvider(pack.DiceClusterName)
	if err != nil {
		return nil, err
	}

	kongInfo, err := impl.kongDb.GetKongInfo(&orm.GatewayKongInfo{
		Az:        pack.DiceClusterName,
		ProjectId: pack.DiceProjectId,
		Env:       pack.DiceEnv,
	})
	if err != nil {
		return nil, err
	}

	var gatewayAdapter gateway_providers.GatewayAdapter
	switch gatewayProvider {
	case mseCommon.MseProviderName:
		consumers, err := (*impl.consumerBiz).GetConsumersOfPackage(pack.Id)
		if err != nil {
			return nil, err
		}

		gatewayAdapter, err = mse.NewMseAdapter(pack.DiceClusterName)
		if err != nil {
			return nil, err
		}
		pluginName := ""
		switch authType {
		case gw.AT_KEY_AUTH:
			config := map[string]interface{}{}
			wlConsumers, err := impl.mseConsumerConfig(consumers)
			if err != nil {
				return nil, err
			}
			if len(wlConsumers) == 0 {
				wlConsumers = append(wlConsumers, mseDto.Consumers{
					Name: plugins.MseDefaultConsumerName,
				})
			}
			config["whitelist"] = wlConsumers
			authRule.Config = config
			pluginName = mseCommon.MsePluginKeyAuth
		case gw.AT_OAUTH2:
			authRule.Config = gw.OAUTH2_CONFIG
			pluginName = gw.AT_OAUTH2
		case gw.AT_SIGN_AUTH:
			config := map[string]interface{}{}
			wlConsumers, err := impl.mseConsumerConfig(consumers)
			if err != nil {
				return nil, err
			}
			if len(wlConsumers) == 0 {
				wlConsumers = append(wlConsumers, mseDto.Consumers{
					Name:   plugins.MseDefaultConsumerName,
					Key:    plugins.MseDefaultConsumerKey,
					Secret: plugins.MseDefaultConsumerSecret,
				})
			}
			config["whitelist"] = wlConsumers
			authRule.Config = config
			pluginName = mseCommon.MsePluginParaSignAuth
		case gw.AT_HMAC_AUTH:
			config := map[string]interface{}{}
			wlConsumers, err := impl.mseConsumerConfig(consumers)
			if err != nil {
				return nil, err
			}
			if len(wlConsumers) == 0 {
				wlConsumers = append(wlConsumers, mseDto.Consumers{
					Name:   plugins.MseDefaultConsumerName,
					Key:    plugins.MseDefaultConsumerKey,
					Secret: plugins.MseDefaultConsumerSecret,
				})
			}
			config["whitelist"] = wlConsumers
			authRule.Config = config
			pluginName = mseCommon.MsePluginHmacAuth
		case gw.AT_ALIYUN_APP:
			authRule.Config = nil
			pluginName = gw.AT_ALIYUN_APP
			authRule.NotKongPlugin = true
		default:
			return nil, errors.Errorf("invalide auth type:%s", authType)
		}

		enabled, err := gatewayAdapter.CheckPluginEnabled(pluginName)
		if err != nil {
			return nil, err
		}
		if !enabled {
			return nil, errors.Errorf("plugin %s is not support", pluginName)
		}
		authRule.PluginName = pluginName

	case "":
		gatewayAdapter = kong.NewKongAdapter(kongInfo.KongAddr)
		switch authType {
		case gw.AT_KEY_AUTH:
			authRule.Config = gw.KEYAUTH_CONFIG
		case gw.AT_OAUTH2:
			authRule.Config = gw.OAUTH2_CONFIG
		case gw.AT_SIGN_AUTH:
			authRule.Config = gw.SIGNAUTH_CONFIG
		case gw.AT_HMAC_AUTH:
			enabled, err := gatewayAdapter.CheckPluginEnabled(gw.AT_HMAC_AUTH)
			if err != nil {
				return nil, err
			}
			if !enabled {
				return nil, errors.New("hmac-auth plugin is not enabled")
			}
			authRule.Config = gw.HMACAUTH_CONFIG
		case gw.AT_ALIYUN_APP:
			authRule.Config = nil
			authRule.NotKongPlugin = true
		default:
			return nil, errors.Errorf("invalide auth type:%s", authType)
		}
	default:
		return nil, errors.Errorf("Unknown gatewayProvider: %v\n", gatewayProvider)
	}

	return authRule, nil
}

func (impl GatewayOpenapiServiceImpl) CreatePackage(ctx context.Context, args *gw.DiceArgsDto, dto *gw.PackageDto) (result *gw.PackageInfoDto, existName string, err error) {
	ctx = context1.WithLoggerIfWithout(ctx, logrus.StandardLogger())
	l := ctx.(*context1.LogContext).Entry()

	var (
		diceInfo          gw.DiceInfo
		helper            *db.SessionHelper
		pack              *orm.GatewayPackage
		aclRule, authRule *gw.OpenapiRule
		packSession       db.GatewayPackageService
		z                 *orm.GatewayZone
		unique            bool
		domains           []string
	)
	defer func() {
		if err != nil {
			l.Errorf("error happened, err:%+v", err)
			if helper != nil {
				_ = helper.Rollback()
				helper.Close()
			}
		}
	}()
	if args.ProjectId == "" || args.Env == "" {
		err = errors.New("projectId or env is empty")
		return
	}
	dto.Name = strings.TrimSpace(dto.Name)
	for i := 0; i < len(dto.BindDomain); i++ {
		dto.BindDomain[i] = strings.TrimSpace(dto.BindDomain[i])
	}
	dto.BindDomain = util.UniqStringSlice(dto.BindDomain)
	auditMeta := make(map[string]interface{})
	auditMeta["endpoint"] = dto.Name
	auditMeta["domains"] = strings.Join(dto.BindDomain, ", ")
	defer func() {
		audit := common.MakeAuditInfo(impl.reqCtx, common.ScopeInfo{
			ProjectId: args.ProjectId,
			Workspace: args.Env,
		}, apistructs.CreateEndpointTemplate, err, auditMeta)
		if audit != nil {
			berr := bundle.Bundle.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: *audit})
			if berr != nil {
				l.Errorf("create audit failed, err:%+v", berr)
			}
		}
	}()
	orgId := args.OrgId
	if orgId == "" {
		orgId = apis.GetOrgID(impl.reqCtx)
	}
	if orgId == "" {
		orgId, _ = (*impl.globalBiz).GetOrgId(args.ProjectId)
	}
	az, err := impl.azDb.GetAz(&orm.GatewayAzInfo{
		OrgId:     orgId,
		Env:       args.Env,
		ProjectId: args.ProjectId,
	})
	if err != nil {
		return
	}
	if dto.Scene == orm.HubScene {
		return impl.createHubPackage(ctx, args, dto, az)
	}

	err = dto.CheckValid()
	if err != nil {
		return
	}
	diceInfo = gw.DiceInfo{
		OrgId:     args.OrgId,
		ProjectId: args.ProjectId,
		Env:       args.Env,
		Az:        az,
	}
	helper, err = db.NewSessionHelper()
	if err != nil {
		return
	}
	// create self zone
	z, err = (*impl.zoneBiz).CreateZone(zone.ZoneConfig{
		Name:        "package-" + dto.Name,
		ProjectId:   args.ProjectId,
		Env:         args.Env,
		Az:          az,
		Type:        db.ZONE_TYPE_PACKAGE_NEW,
		ServiceName: args.ServiceName,
		Namespace:   args.Namespace,
	}, helper)
	if err != nil {
		return
	}
	// create package in db
	packSession, err = impl.packageDb.NewSession(helper)
	if err != nil {
		return
	}
	pack = &orm.GatewayPackage{
		DiceOrgId:       args.OrgId,
		DiceProjectId:   args.ProjectId,
		DiceEnv:         args.Env,
		DiceClusterName: az,
		ZoneId:          z.Id,
		PackageName:     dto.Name,
		AuthType:        dto.AuthType,
		AclType:         dto.AclType,
		Scene:           dto.Scene,
		Description:     dto.Description,
	}
	unique, err = packSession.CheckUnique(pack)
	if err != nil {
		return
	}
	if !unique {
		err = errors.Errorf("package %s already exists", pack.PackageName)
		existName = pack.PackageName
		return
	}
	err = packSession.Insert(pack)
	if err != nil {
		return
	}
	domains, err = (*impl.domainBiz).TouchPackageDomain(args.OrgId, pack.Id, az, dto.BindDomain, helper)
	if err != nil {
		return
	}
	if dto.Scene == orm.OpenapiScene {
		gatewayProvider, errGetProvider := impl.GetGatewayProvider(az)
		if errGetProvider != nil {
			err = errGetProvider
			return
		}

		switch gatewayProvider {
		case "":
			authRule, err = impl.createAuthRule(dto.AuthType, pack)
			if err != nil {
				return
			}
			authRule.Region = gw.PACKAGE_RULE
			aclRule, err = impl.createAclRule(dto.AclType, pack.Id, az)
			if err != nil {
				return
			}
			aclRule.Region = gw.PACKAGE_RULE
			err = (*impl.ruleBiz).CreateRule(diceInfo, authRule, helper)
			if err != nil {
				return
			}
			err = (*impl.ruleBiz).CreateRule(diceInfo, aclRule, helper)
			if err != nil {
				return
			}

		case mseCommon.MseProviderName:
			// create auth, acl rule
			authRule, err = impl.createAuthRule(dto.AuthType, pack)
			if err != nil {
				return
			}
			authRule.Region = gw.PACKAGE_RULE
			err = (*impl.ruleBiz).CreateRule(diceInfo, authRule, helper)
			if err != nil {
				return
			}

		default:
			err = errors.Errorf("unknown gateway provider %s", gatewayProvider)
			return
		}
		// update zone kong polices
		err = (*impl.ruleBiz).SetPackageKongPolicies(pack, helper)
		if err != nil {
			return
		}
	}
	err = helper.Commit()
	if err != nil {
		return
	}
	helper.Close()
	result = impl.packageDto(pack, domains)
	return
}

func (impl GatewayOpenapiServiceImpl) createHubPackage(ctx context.Context, args *gw.DiceArgsDto, dto *gw.PackageDto, az string) (result *gw.PackageInfoDto, existName string, err error) {
	kongInfos, err := impl.kongDb.SelectByAny(&orm.GatewayKongInfo{Env: args.Env, Az: az})
	if err != nil {
		return nil, "", err
	}
	if len(kongInfos) == 0 {
		return nil, "", errors.New("kong info not found")
	}

	helper, err := db.NewSessionHelper()
	if err != nil {
		return nil, "", err
	}
	hubInfoService, err := impl.hubInfoDb.NewSession(helper)
	if err != nil {
		return nil, "", err
	}
	hubInfos, err := hubInfoService.SelectByAny(&orm.GatewayHubInfo{
		OrgId:           args.OrgId,
		DiceEnv:         args.Env,
		DiceClusterName: az,
	})
	if err != nil {
		return nil, "", err
	}
	for i := range dto.BindDomain {
		dto.BindDomain[i] = strings.TrimSpace(dto.BindDomain[i])
	}
	hepautils.SortDomains(dto.BindDomain)
	hubInfo, ok := hubExists(hubInfos, dto.BindDomain)
	if !ok {
		hubInfo = &orm.GatewayHubInfo{
			OrgId:           args.OrgId,
			DiceEnv:         args.Env,
			DiceClusterName: az,
			BindDomain:      hepautils.SortJoinDomains(dto.BindDomain),
		}
		if err = hubInfoService.Insert(hubInfo); err != nil {
			return nil, "", err
		}
	}

	var tenants = make(map[string]struct{})
	for _, kongInfo := range kongInfos {
		if _, ok := tenants[kongInfo.TenantId]; ok {
			continue
		}
		tenants[kongInfo.TenantId] = struct{}{}
		if orgID, _ := (*impl.globalBiz).GetOrgId(kongInfo.ProjectId); orgID != args.OrgId {
			continue
		}
		if _, err = impl.createOrGetTenantHubPackage(ctx, hubInfo, args.OrgId, args.ProjectId, kongInfo.Env, kongInfo.Az, helper); err != nil {
			return nil, "", err
		}
	}
	pkg, err := impl.packageDb.GetByAny(&orm.GatewayPackage{
		DiceProjectId:   args.ProjectId,
		DiceEnv:         args.Env,
		DiceClusterName: az,
		Scene:           orm.HubScene,
		BindDomain:      hubInfo.BindDomain,
	})
	if err != nil {
		return nil, "", err
	}
	return impl.packageDto(pkg, dto.BindDomain), "", nil

}

func (impl GatewayOpenapiServiceImpl) GetPackage(id string) (dto *gw.PackageInfoDto, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("error happened, err:%+v", err)
		}
	}()
	if id == "" {
		err = errors.New("id is empty")
		return
	}
	dao, err := impl.packageDb.Get(id)
	var domains []string
	if err != nil {
		return
	}
	if dao == nil {
		err = errors.Errorf("package not found, packageId: %s", id)
		return
	}
	domains, err = (*impl.domainBiz).GetPackageDomains(dao.Id)
	if err != nil {
		return
	}
	dto = impl.packageDto(dao, domains)
	return
}

func (impl GatewayOpenapiServiceImpl) GetPackages(ctx context.Context, args *gw.GetPackagesDto) (res common.NewPageQuery, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("error happened: %+v", err)
		}
	}()
	if args.ProjectId == "" || args.Env == "" {
		err = errors.New("projectId or env is empty")
		return
	}
	var domainInfos []orm.GatewayDomain
	var packageIds []string
	if args.Domain != "" {
		domainInfos, err = (*impl.domainBiz).FindDomains(args.Domain, args.ProjectId, args.Env, orm.PrefixMatch, orm.DT_PACKAGE)
		if err != nil {
			return
		}
	}
	for _, domainInfo := range domainInfos {
		if domainInfo.PackageId != "" {
			packageIds = append(packageIds, domainInfo.PackageId)
		}
	}
	pageInfo := common.NewPage2(args.PageSize, args.PageNo)
	options := args.GenSelectOptions()
	if len(packageIds) > 0 {
		options = append(options, orm.SelectOption{
			Type:   orm.Contains,
			Column: "id",
			Value:  packageIds,
		})
	}
	if args.Domain != "" && len(packageIds) == 0 {
		res = common.NewPages(nil, 0)
		return
	}
	var list gw.SortBySceneList
	var daos []orm.GatewayPackage
	var ok bool
	page, err := impl.packageDb.GetPage(options, (*common.Page)(pageInfo))
	if err != nil {
		return
	}
	daos, ok = page.Result.([]orm.GatewayPackage)
	if !ok {
		err = errors.New("type convert failed")
		return
	}

	for _, dao := range daos {
		var dto *gw.PackageInfoDto
		var domains []string
		domains, err = (*impl.domainBiz).GetPackageDomains(dao.Id)
		if err != nil {
			return
		}
		dto = impl.packageDto(&dao, domains)
		list = append(list, *dto)
	}
	sort.Sort(list)
	res = common.NewPages(list, pageInfo.TotalNum)
	return
}

func (impl GatewayOpenapiServiceImpl) ListAllPackages() ([]orm.GatewayPackage, error) {
	return impl.packageDb.SelectByAny(new(orm.GatewayPackage))
}

func (impl GatewayOpenapiServiceImpl) packageDto(dao *orm.GatewayPackage, domains []string) *gw.PackageInfoDto {
	var res gw.PackageInfoDto
	if dao == nil {
		return &res
	}

	gatewayProvider, err := impl.GetGatewayProvider(dao.DiceClusterName)
	if err != nil {
		logrus.Errorf("get gateway provider for cluster %s failed: %v\n", dao.DiceClusterName, err)
		return &res
	}

	res = gw.PackageInfoDto{
		Id:       dao.Id,
		CreateAt: dao.CreateTime.Format("2006-01-02T15:04:05"),
		PackageDto: gw.PackageDto{
			Name:            dao.PackageName,
			BindDomain:      domains,
			AuthType:        dao.AuthType,
			AclType:         dao.AclType,
			Description:     dao.Description,
			Scene:           dao.Scene,
			GatewayProvider: gatewayProvider,
		},
	}
	return &res
}

func (impl GatewayOpenapiServiceImpl) GetPackagesName(args *gw.GetPackagesDto) (list []gw.PackageInfoDto, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("error happened, err:%+v", err)
		}
	}()
	if args.ProjectId == "" || args.Env == "" {
		return
	}
	list = []gw.PackageInfoDto{}
	var daos []orm.GatewayPackage

	orgId := args.OrgId
	if orgId == "" {
		orgId = apis.GetOrgID(impl.reqCtx)
	}
	if orgId == "" {
		orgId, _ = (*impl.globalBiz).GetOrgId(args.ProjectId)
	}
	az, err := impl.azDb.GetAz(&orm.GatewayAzInfo{
		OrgId:     orgId,
		Env:       args.Env,
		ProjectId: args.ProjectId,
	})
	if err != nil {
		return
	}
	daos, err = impl.packageDb.SelectByAny(&orm.GatewayPackage{
		DiceProjectId:   args.ProjectId,
		DiceEnv:         args.Env,
		DiceClusterName: az,
	})
	if err != nil {
		return
	}
	for _, dao := range daos {
		var dto *gw.PackageInfoDto
		var domains []string
		domains, err = (*impl.domainBiz).GetPackageDomains(dao.Id)
		if err != nil {
			return
		}
		dto = impl.packageDto(&dao, domains)
		list = append(list, *dto)
	}
	return
}

func (impl GatewayOpenapiServiceImpl) updatePackageApiHost(pack *orm.GatewayPackage, hosts []string) error {
	var gatewayAdapter gateway_providers.GatewayAdapter
	gatewayProvider, err := impl.GetGatewayProvider(pack.DiceClusterName)
	if err != nil {
		return err
	}

	kongInfo, err := impl.kongDb.GetKongInfo(&orm.GatewayKongInfo{
		Az:        pack.DiceClusterName,
		ProjectId: pack.DiceProjectId,
		Env:       pack.DiceEnv,
	})
	if err != nil {
		return err
	}

	switch gatewayProvider {
	case mseCommon.MseProviderName:
		logrus.Debugf("mse gateway no need update GatewayRoute.")
		gatewayAdapter, err = mse.NewMseAdapter(pack.DiceClusterName)
		if err != nil {
			return err
		}
	case "":
		gatewayAdapter = kong.NewKongAdapter(kongInfo.KongAddr)
	default:
		return errors.Errorf("unknown gateway provider:%v\n", gatewayProvider)
	}
	apis, err := impl.packageApiDb.SelectByAny(&orm.GatewayPackageApi{
		PackageId:    pack.Id,
		RedirectType: gw.RT_URL,
	})
	if err != nil {
		return err
	}
	for _, api := range apis {
		if api.Origin != string(gw.FROM_CUSTOM) && api.Origin != string(gw.FROM_DICEYML) {
			continue
		}
		route, err := impl.routeDb.GetByApiId(api.Id)
		if err != nil {
			return err
		}
		if route == nil {
			return errors.Errorf("can't find route of api:%s", api.Id)
		}
		req := providerDto.NewKongRouteReqDto()
		req.RouteId = route.RouteId
		req.Hosts = hosts
		req.AddTag("package_api_id", api.Id)
		resp, err := gatewayAdapter.UpdateRoute(req)
		if err != nil {
			return err
		}
		routeDao := impl.kongRouteDao(resp, route.ServiceId, api.Id)
		routeDao.Id = route.Id
		err = impl.routeDb.Update(routeDao)
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) UpdatePackage(orgId, id string, dto *gw.PackageDto) (result *gw.PackageInfoDto, err error) {
	var diceInfo gw.DiceInfo
	var dao *orm.GatewayPackage
	var z *orm.GatewayZone
	var apiSession db.GatewayPackageApiService
	var domains []string
	var apis []orm.GatewayPackageApi
	var session *db.SessionHelper
	defer func() {
		if err != nil {
			logrus.Errorf("error happened, err:%+v", err)
			if session != nil {
				_ = session.Rollback()
				session.Close()
			}
		}
	}()
	if id == "" {
		err = errors.New("id is empty")
		return
	}
	needUpdate := false
	auditCtx := map[string]interface{}{}
	auditCtx["endpoint"] = dto.Name
	auditCtx["domains"] = strings.Join(dto.BindDomain, ", ")
	defer func() {
		audit := common.MakeAuditInfo(impl.reqCtx, common.ScopeInfo{
			ProjectId: dao.DiceProjectId,
			Workspace: dao.DiceEnv,
		}, apistructs.UpdateEndpointTemplate, err, auditCtx)
		if audit != nil {
			berr := bundle.Bundle.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: *audit})
			if berr != nil {
				logrus.Errorf("create audit failed, err:%+v", berr)
			}
		}
	}()
	domainHasDiff := false
	session, err = db.NewSessionHelper()
	if err != nil {
		return
	}
	err = dto.CheckValid()
	if err != nil {
		return
	}
	for i := 0; i < len(dto.BindDomain); i++ {
		dto.BindDomain[i] = strings.TrimSpace(dto.BindDomain[i])
	}
	dto.BindDomain = util.UniqStringSlice(dto.BindDomain)
	dao, err = impl.packageDb.Get(id)
	if err != nil {
		return
	}
	if dao.Scene == orm.UnityScene || dao.Scene == orm.HubScene {
		err = errors.Errorf("%s scene package can not be updated", dao.Scene)
		return
	}
	diceInfo = gw.DiceInfo{
		OrgId:     orgId,
		ProjectId: dao.DiceProjectId,
		Env:       dao.DiceEnv,
		Az:        dao.DiceClusterName,
	}
	domainHasDiff, err = (*impl.domainBiz).IsPackageDomainsDiff(id, dao.DiceClusterName, dto.BindDomain, session)
	if err != nil {
		return
	}
	if domainHasDiff {
		needUpdate = true
		domains, err = (*impl.domainBiz).TouchPackageDomain(orgId, id, dao.DiceClusterName, dto.BindDomain, session)
		if err != nil {
			return
		}

		z, err = (*impl.zoneBiz).GetZone(dao.ZoneId, session)
		if err != nil {
			return
		}
		//老类型兼容
		if z != nil && z.Type == db.ZONE_TYPE_PACKAGE {
			_, err = (*impl.zoneBiz).UpdateZoneRoute(dao.ZoneId, zone.ZoneRoute{
				zone.RouteConfig{
					Hosts: domains,
					Path:  "/",
				},
			}, nil, "")
			if err != nil {
				return
			}
		}
		apiSession, err = impl.packageApiDb.NewSession(session)
		if err != nil {
			return
		}
		apis, err = apiSession.SelectByAny(&orm.GatewayPackageApi{PackageId: id})
		if err != nil {
			return
		}
		wg := sync.WaitGroup{}
		for _, api := range apis {
			wg.Add(1)
			go func(apiRefer orm.GatewayPackageApi) {
				defer util.DoRecover()
				logrus.Debugf("begin touch pacakge api zone, api:%+v", apiRefer)
				zoneId, aerr := impl.TouchPackageApiZone(endpoint_api.PackageApiInfo{&apiRefer, domains, dao.DiceProjectId, dao.DiceEnv, dao.DiceClusterName, false})
				if aerr != nil {
					err = aerr
				}
				if apiRefer.ZoneId == "" {
					apiRefer.ZoneId = zoneId
					aerr = apiSession.Update(&apiRefer)
					if aerr != nil {
						err = aerr
					}
				}
				wg.Done()
				logrus.Debugf("complete touch pacakge api zone, api:%+v", apiRefer)
			}(api)
		}
		wg.Wait()
		if err != nil {
			return
		}
		err = impl.updatePackageApiHost(dao, domains)
		if err != nil {
			return
		}
	}
	if dao.AuthType != dto.AuthType {
		needUpdate = true
		var oldAuthRule []gw.OpenapiRuleInfo
		oldAuthRule, err = (*impl.ruleBiz).GetPackageRules(id, session, gw.AUTH_RULE)
		if err != nil {
			return
		}
		for _, rule := range oldAuthRule {
			err = (*impl.ruleBiz).DeleteRule(rule.Id, session)
			if err != nil {
				return
			}
		}
		if dto.AuthType != "" {
			var authRule *gw.OpenapiRule
			authRule, err = impl.createAuthRule(dto.AuthType, dao)
			if err != nil {
				return
			}
			authRule.Region = gw.PACKAGE_RULE
			err = (*impl.ruleBiz).CreateRule(diceInfo, authRule, session)
			if err != nil {
				return
			}
		}
	}
	if dao.AclType != dto.AclType {
		needUpdate = true
		var oldAclRule []gw.OpenapiRuleInfo
		oldAclRule, err = (*impl.ruleBiz).GetPackageRules(id, nil, gw.ACL_RULE)
		if err != nil {
			return
		}
		for _, rule := range oldAclRule {
			err = (*impl.ruleBiz).DeleteRule(rule.Id, session)
			if err != nil {
				return
			}
		}
		if dto.AclType != "" {
			var aclRule *gw.OpenapiRule
			aclRule, err = impl.createAclRule(dto.AclType, id, dao.DiceClusterName)
			if err != nil {
				return
			}
			aclRule.Region = gw.PACKAGE_RULE
			err = (*impl.ruleBiz).CreateRule(diceInfo, aclRule, session)
			if err != nil {
				return
			}
		}
	}
	dao.AuthType = dto.AuthType
	dao.AclType = dto.AclType
	dao.Description = dto.Description
	err = impl.packageDb.Update(dao)
	if err != nil {
		return
	}
	if needUpdate {
		// update zone kong polices
		err = (*impl.ruleBiz).SetPackageKongPolicies(dao, session)
		if err != nil {
			return
		}
	}
	err = session.Commit()
	if err != nil {
		return
	}
	session.Close()
	result = impl.packageDto(dao, dto.BindDomain)
	return
}

func (impl GatewayOpenapiServiceImpl) TryClearRuntimePackage(runtimeService *orm.GatewayRuntimeService, session *db.SessionHelper, force ...bool) error {
	packSession, _ := impl.packageDb.NewSession(session)
	packApiSession, _ := impl.packageApiDb.NewSession(session)
	pack, err := packSession.GetByAny(&orm.GatewayPackage{
		RuntimeServiceId: runtimeService.Id,
	})
	if err != nil {
		return err
	}
	if pack != nil {
		apis, err := packApiSession.SelectByAny(&orm.GatewayPackageApi{
			PackageId: pack.Id,
			Origin:    string(gw.FROM_CUSTOM),
		})
		if err != nil {
			return err
		}
		otherServiceApiExist := false
		for _, api := range apis {
			if api.DiceApp != runtimeService.AppName || api.DiceService != runtimeService.ServiceName {
				otherServiceApiExist = true
			}
		}
		if !otherServiceApiExist || (len(force) > 0 && force[0]) {
			err = impl.deletePackage(pack.Id, session)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type ctxWg struct {
}

func (impl *GatewayOpenapiServiceImpl) deletePackage(id string, session *db.SessionHelper, auditCtx ...map[string]interface{}) error {
	// var consumers []orm.GatewayConsumer
	var apis []orm.GatewayPackageApi
	var dao *orm.GatewayPackage
	var oldDomains []string
	defer func() {
		if len(auditCtx) == 0 {
			return
		}
		if dao == nil {
			return
		}
		auditCtx[0]["projectId"] = dao.DiceProjectId
		auditCtx[0]["workspace"] = dao.DiceEnv
		auditCtx[0]["endpoint"] = dao.PackageName
		auditCtx[0]["domains"] = oldDomains
	}()
	wg := &sync.WaitGroup{}
	impl.ctx = context.WithValue(context.Background(), ctxWg{}, wg)
	packSession, err := impl.packageDb.NewSession(session)
	if err != nil {
		return err
	}
	apiSession, err := impl.packageApiDb.NewSession(session)
	if err != nil {
		return err
	}
	dao, err = packSession.Get(id)
	if err != nil {
		return err
	}
	if dao == nil {
		return nil
	}
	if dao.Scene == orm.UnityScene || dao.Scene == orm.HubScene {
		return errors.Errorf("%s scene can't be deleted", dao.Scene)
	}
	// remove consumer check
	// consumers, err = (*impl.consumerBiz).GetConsumersOfPackage(id)
	// if err != nil {
	// 	return err
	// }
	// if len(consumers) > 0 {
	// 	return errors.New("this endpoint has granted access to other consumers, please cancel first")
	// }
	apis, _ = apiSession.SelectByAny(&orm.GatewayPackageApi{
		PackageId: id,
	})
	for _, api := range apis {
		wg.Add(1)
		_, err := impl.DeletePackageApi(id, api.Id)
		if err != nil {
			return errors.Errorf("delete package api failed, err:%+v", err)
		}
	}
	err = impl.apiInPackageDb.DeleteByPackageId(id)
	if err != nil {
		return err
	}
	err = impl.packageInDb.DeleteByPackageId(id)
	if err != nil {
		return err
	}
	err = (*impl.ruleBiz).DeleteByPackage(dao)
	if err != nil {
		return err
	}
	if dao.ZoneId != "" {
		err = (*impl.zoneBiz).DeleteZone(dao.ZoneId, "")
		if err != nil {
			return err
		}
	}
	oldDomains, err = (*impl.domainBiz).GetPackageDomains(id, session)
	if err != nil {
		return err
	}
	_, err = (*impl.domainBiz).TouchPackageDomain("", id, dao.DiceClusterName, nil, session)
	if err != nil {
		return err
	}
	err = packSession.Delete(id, true)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) DeletePackage(id string) (result bool, err error) {
	var session *db.SessionHelper
	defer func() {
		if err != nil {
			logrus.Errorf("error happened, err:%+v", err)
			if session != nil {
				_ = session.Rollback()
				session.Close()
			}
		}
	}()
	auditCtx := map[string]interface{}{}
	defer func() {
		if auditCtx["projectId"] == nil || auditCtx["workspace"] == nil {
			return
		}
		audit := common.MakeAuditInfo(impl.reqCtx, common.ScopeInfo{
			ProjectId: auditCtx["projectId"].(string),
			Workspace: auditCtx["workspace"].(string),
		}, apistructs.DeleteEndpointTemplate, err, auditCtx)
		if audit != nil {
			berr := bundle.Bundle.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: *audit})
			if berr != nil {
				logrus.Errorf("create audit failed, err:%+v", berr)
			}
		}
	}()
	if id == "" {
		err = errors.New("id is empty")
		return
	}
	session, _ = db.NewSessionHelper()
	err = impl.deletePackage(id, session, auditCtx)
	if err != nil {
		return
	}
	err = session.Commit()
	if err != nil {
		return
	}
	session.Close()
	result = true
	return
}

func (impl GatewayOpenapiServiceImpl) packageApiDao(dto *gw.OpenapiDto) *orm.GatewayPackageApi {
	dao := &orm.GatewayPackageApi{
		ApiPath:      dto.ApiPath,
		Method:       dto.Method,
		DiceApp:      dto.RedirectApp,
		DiceService:  dto.RedirectService,
		RedirectAddr: dto.RedirectAddr,
		RedirectPath: dto.RedirectPath,
		RedirectType: dto.RedirectType,
		Description:  dto.Description,
		Origin:       string(gw.FROM_CUSTOM),
	}
	if dto.AllowPassAuth {
		dao.AclType = gw.ACL_OFF
	}
	if dto.Origin != "" {
		dao.Origin = string(dto.Origin)
	}
	return dao
}

func (impl GatewayOpenapiServiceImpl) createKongServiceReq(dto *gw.OpenapiDto, serviceId ...string) *providerDto.ServiceReqDto {
	i := 0
	reqDto := &providerDto.ServiceReqDto{
		Url:            dto.AdjustRedirectAddr,
		ConnectTimeout: 5000,
		ReadTimeout:    60000,
		WriteTimeout:   60000,
		Retries:        &i,
	}
	if len(serviceId) != 0 {
		reqDto.ServiceId = serviceId[0]
	}
	return reqDto
}

func (impl GatewayOpenapiServiceImpl) kongServiceDao(dto *providerDto.ServiceRespDto, apiId string) *orm.GatewayService {
	return &orm.GatewayService{
		ServiceId:   dto.Id,
		ServiceName: dto.Name,
		Protocol:    dto.Protocol,
		Host:        dto.Host,
		Port:        strconv.Itoa(dto.Port),
		Path:        dto.Path,
		ApiId:       apiId,
	}
}

func (impl GatewayOpenapiServiceImpl) touchKongService(adapter gateway_providers.GatewayAdapter, dto *gw.OpenapiDto, apiId string,
	helper ...*db.SessionHelper) (string, error) {
	var serviceSession db.GatewayServiceService
	var err error
	if len(helper) > 0 {
		serviceSession, err = impl.serviceDb.NewSession(helper[0])
		if err != nil {
			return "", err
		}
	} else {
		serviceSession = impl.serviceDb
	}
	service, err := serviceSession.GetByApiId(apiId)
	if err != nil {
		return "", err
	}
	var req *providerDto.ServiceReqDto
	if service == nil {
		req = impl.createKongServiceReq(dto)
	} else {
		req = impl.createKongServiceReq(dto, service.ServiceId)
	}
	resp, err := adapter.CreateOrUpdateService(req)
	if err != nil {
		return "", err
	}
	serviceDao := impl.kongServiceDao(resp, apiId)
	if err != nil {
		return "", err
	}
	if service == nil {
		err = serviceSession.Insert(serviceDao)
		if err != nil {
			return "", err
		}
	} else {
		serviceDao.Id = service.Id
		err = serviceSession.Update(serviceDao)
		if err != nil {
			return "", err
		}
	}
	return resp.Id, nil
}

func (impl GatewayOpenapiServiceImpl) deleteKongService(adapter gateway_providers.GatewayAdapter, apiId string) error {
	service, err := impl.serviceDb.GetByApiId(apiId)
	if err != nil {
		return err
	}
	if service == nil {
		return nil
	}
	err = adapter.DeleteService(service.ServiceId)
	if err != nil {
		return err
	}
	err = impl.serviceDb.DeleteById(service.Id)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) createKongRouteReq(dto *gw.OpenapiDto, serviceId string, routeId ...string) *providerDto.RouteReqDto {
	reqDto := providerDto.NewKongRouteReqDto()
	if dto.Method != "" {
		reqDto.Methods = []string{dto.Method}
	}
	if dto.ApiPath != "" {
		reqDto.Paths = []string{dto.AdjustPath}
	}
	reqDto.Hosts = dto.Hosts
	reqDto.Service = &providerDto.Service{}
	reqDto.Service.Id = serviceId
	if dto.IsRegexPath {
		ignore := strings.Count(dto.AdjustPath, "^/") + strings.Count(dto.AdjustPath, `\/`)
		reqDto.RegexPriority = strings.Count(dto.AdjustPath, "/") - ignore
	}
	if len(routeId) != 0 {
		reqDto.RouteId = routeId[0]
	}
	return reqDto
}

func (impl GatewayOpenapiServiceImpl) kongRouteDao(dto *providerDto.RouteRespDto, serviceId, apiId string) *orm.GatewayRoute {
	protocols, _ := json.Marshal(dto.Protocols)
	hosts, _ := json.Marshal(dto.Hosts)
	paths, _ := json.Marshal(dto.Paths)
	methods, _ := json.Marshal(dto.Methods)
	return &orm.GatewayRoute{
		RouteId:   dto.Id,
		Protocols: string(protocols),
		Hosts:     string(hosts),
		Paths:     string(paths),
		Methods:   string(methods),
		ServiceId: serviceId,
		ApiId:     apiId,
	}
}

func (impl GatewayOpenapiServiceImpl) touchKongRoute(adapter gateway_providers.GatewayAdapter, dto *gw.OpenapiDto, apiId string,
	helper ...*db.SessionHelper) (string, error) {
	var routeSession db.GatewayRouteService
	var err error
	if len(helper) > 0 {
		routeSession, err = impl.routeDb.NewSession(helper[0])
		if err != nil {
			return "", err
		}
	} else {
		routeSession = impl.routeDb
	}
	route, err := routeSession.GetByApiId(apiId)
	if err != nil {
		return "", err
	}
	var req *providerDto.RouteReqDto
	if route == nil {
		req = impl.createKongRouteReq(dto, dto.ServiceId)
	} else {
		req = impl.createKongRouteReq(dto, dto.ServiceId, route.RouteId)
	}
	req.AddTag("package_api_id", apiId)
	resp, err := adapter.CreateOrUpdateRoute(req)
	if err != nil {
		return "", err
	}
	routeDao := impl.kongRouteDao(resp, dto.ServiceId, apiId)
	if err != nil {
		return "", err
	}
	if route == nil {
		err = routeSession.Insert(routeDao)
		if err != nil {
			return "", err
		}
	} else {
		routeDao.Id = route.Id
		err = routeSession.Update(routeDao)
		if err != nil {
			return "", err
		}
	}
	return resp.Id, nil
}

func (impl GatewayOpenapiServiceImpl) deleteKongRoute(adapter gateway_providers.GatewayAdapter, apiId string) error {
	route, err := impl.routeDb.GetByApiId(apiId)
	if err != nil {
		return err
	}
	if route == nil {
		return nil
	}
	err = adapter.DeleteRoute(route.RouteId)
	if err != nil {
		return err
	}
	err = impl.routeDb.DeleteById(route.Id)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) DeleteKongApi(adapter gateway_providers.GatewayAdapter, apiId string) error {
	err := impl.deleteKongRoute(adapter, apiId)
	if err != nil {
		return err
	}
	err = impl.deleteKongService(adapter, apiId)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) GetGatewayProvider(clusterName string) (string, error) {
	if clusterName == "" {
		return "", errors.Errorf("clusterName is nil")
	}
	_, azInfo, err := impl.azDb.GetAzInfoByClusterName(clusterName)
	if err != nil {
		return "", err
	}

	if azInfo != nil && azInfo.GatewayProvider != "" {
		return azInfo.GatewayProvider, nil
	}
	return "", nil
}

func (impl GatewayOpenapiServiceImpl) createOrUpdatePlugins(adapter gateway_providers.GatewayAdapter, dto *gw.OpenapiDto) error {
	if dto.ServiceRewritePath != "" {
		reqDto := &providerDto.PluginReqDto{
			Name: "path-variable",
			// ServiceId: dto.ServiceId,
			RouteId: dto.RouteId,
			Config: map[string]interface{}{
				"request_regex": dto.AdjustPath,
				"rewrite_path":  dto.ServiceRewritePath,
			},
		}
		if _, ok := adapter.(*mse.MseAdapterImpl); !ok {
			_, err := adapter.CreateOrUpdatePlugin(reqDto)
			if err != nil {
				return err
			}
		}
	}
	if config.ServerConf.HasRouteInfo {
		reqDto := &providerDto.PluginReqDto{
			Name:    "set-route-info",
			RouteId: dto.RouteId,
			Config: map[string]interface{}{
				"project_id": dto.ProjectId,
				"workspace":  strings.ToLower(dto.Env),
				"api_path":   dto.ApiPath,
			},
		}
		if _, ok := adapter.(*mse.MseAdapterImpl); !ok {
			_, err := adapter.CreateOrUpdatePlugin(reqDto)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) SessionCreatePackageApi(id string, dto *gw.OpenapiDto, session *db.SessionHelper, injectRuntimeDomain bool) (bool, string, *apistructs.Audit, error) {
	exist := false
	var pack *orm.GatewayPackage
	var kongInfo *orm.GatewayKongInfo
	var gatewayAdapter gateway_providers.GatewayAdapter
	var dao *orm.GatewayPackageApi
	var unique bool
	var serviceId, routeId string
	var apiSession db.GatewayPackageApiService
	var packSession db.GatewayPackageService
	var err error
	var zoneId string
	var domains []string
	var audit *apistructs.Audit
	var ingressNamespace string
	var consumers []string
	var packageAclInfoDto []gw.PackageAclInfoDto
	gatewayProvider := ""

	needUpdateDomainPolicy := false
	method := dto.Method
	if valid, msg := dto.CheckValid(); !valid {
		return false, "", audit, errors.New(msg)
	}
	packSession, err = impl.packageDb.NewSession(session)
	if err != nil {
		goto failed
	}
	pack, err = packSession.Get(id)
	if err != nil {
		goto failed
	}
	if pack == nil {
		err = errors.New("package not find")
		goto failed
	}
	if method == "" {
		method = "all"
	}
	gatewayProvider, err = impl.GetGatewayProvider(pack.DiceClusterName)
	if err != nil {
		logrus.Errorf("get cluster zaInfo failed: %v", err)
		goto failed
	}
	audit = common.MakeAuditInfo(impl.reqCtx, common.ScopeInfo{
		ProjectId: pack.DiceProjectId,
		Workspace: pack.DiceEnv,
	}, apistructs.CreateRouteTemplate, nil, map[string]interface{}{
		"endpoint": pack.PackageName,
		"method":   method,
		"path":     dto.ApiPath,
	})
	dto.ProjectId = pack.DiceProjectId
	dto.Env = pack.DiceEnv
	if pack.Scene == orm.UnityScene || pack.Scene == orm.HubScene {
		defaultPath, err := (*impl.globalBiz).GenerateDefaultPath(pack.DiceProjectId)
		if err != nil {
			goto failed
		}
		if !strings.HasPrefix(dto.ApiPath, defaultPath) {
			dto.ApiPath = defaultPath + dto.ApiPath
		}
	}
	// get package domain
	domains, err = (*impl.domainBiz).GetPackageDomains(id, session)
	if err != nil {
		goto failed
	}
	if len(domains) == 0 {
		err = errors.New("you need set the endpoint's domain first")
		goto failed
	}
	dto.Hosts = domains
	logrus.Debugf("api dto before adjust %+v", dto) // output for debug
	err = dto.Adjust()
	if err != nil {
		goto failed
	}
	logrus.Debugf("api dto after adjust %+v", *dto) // output for debug
	dao = impl.packageApiDao(dto)
	dao.PackageId = id
	apiSession, err = impl.packageApiDb.NewSession(session)
	if err != nil {
		goto failed
	}
	unique, err = apiSession.CheckUnique(dao)
	if err != nil {
		goto failed
	}
	if !unique {
		err = errors.Errorf("package api already exist, dao:%+v", dao)
		exist = true
		goto failed
	}
	if dto.RedirectType == gw.RT_URL {
		var kongSession db.GatewayKongInfoService
		kongSession, err = impl.kongDb.NewSession(session)
		if err != nil {
			goto failed
		}
		err = apiSession.Insert(dao)
		if err != nil {
			goto failed
		}
		kongInfo, err = kongSession.GetKongInfo(&orm.GatewayKongInfo{
			Az:        pack.DiceClusterName,
			ProjectId: pack.DiceProjectId,
			Env:       pack.DiceEnv,
		})
		if err != nil {
			goto failed
		}
		switch gatewayProvider {
		case mseCommon.MseProviderName:
			gatewayAdapter, err = mse.NewMseAdapter(pack.DiceClusterName)
			if err != nil {
				goto failed
			}
		case "":
			gatewayAdapter = kong.NewKongAdapter(kongInfo.KongAddr)
		default:
			logrus.Errorf("unknown gateway provider:%v\n", gatewayProvider)
			goto failed
		}
		serviceId, err = impl.touchKongService(gatewayAdapter, dto, dao.Id, session)
		if err != nil {
			goto failed
		}
		dto.ServiceId = serviceId
		routeId, err = impl.touchKongRoute(gatewayAdapter, dto, dao.Id, session)
		if err != nil {
			goto failed
		}
		dto.RouteId = routeId
		err = impl.createOrUpdatePlugins(gatewayAdapter, dto)
		if err != nil {
			goto failed
		}
	} else {
		var runtimeSession db.GatewayRuntimeServiceService
		runtimeSession, err = impl.runtimeDb.NewSession(session)
		if err != nil {
			goto failed
		}
		var runtimeService *orm.GatewayRuntimeService
		if dto.RuntimeServiceId == "" {
			runtimeService, err = runtimeSession.GetByAny(&orm.GatewayRuntimeService{
				AppName:     dto.RedirectApp,
				ServiceName: dto.RedirectService,
				RuntimeId:   dto.RedirectRuntimeId,
			})
		} else {
			runtimeService, err = runtimeSession.Get(dto.RuntimeServiceId)
		}
		if err != nil {
			goto failed
		}
		if runtimeService == nil {
			err = errors.New("find runtime service failed")
			goto failed
		}
		ingressNamespace = runtimeService.ProjectNamespace
		dao.RuntimeServiceId = runtimeService.Id
		err = apiSession.Insert(dao)
		if err != nil {
			goto failed
		}
	}
	zoneId, err = impl.TouchPackageApiZone(endpoint_api.PackageApiInfo{dao, domains, pack.DiceProjectId, pack.DiceEnv, pack.DiceClusterName, injectRuntimeDomain}, session)
	if err != nil {
		goto failed
	}
	if dao.ZoneId == "" {
		dao.ZoneId = zoneId
	}
	err = apiSession.Update(dao)
	if err != nil {
		goto failed
	}
	if dto.AllowPassAuth {
		var authRule, aclRule *gw.OpenapiRule
		diceInfo := gw.DiceInfo{
			OrgId:     pack.DiceOrgId,
			ProjectId: pack.DiceProjectId,
			Env:       pack.DiceEnv,
			Az:        pack.DiceClusterName,
		}
		if gatewayProvider == mseCommon.MseProviderName {
			// mse 网关，暂不创建 acl rule
			authRule, err = impl.createApiAuthRule(dao.PackageId, dao.Id, false)
			if err != nil {
				goto failed
			}
			err = (*impl.ruleBiz).CreateRule(diceInfo, authRule, session)
			if err != nil {
				goto failed
			}
		} else {
			aclRule, err = impl.createApiAclRule(gw.ACL_OFF, dao.PackageId, dao.Id, pack.DiceClusterName)
			if err != nil {
				goto failed
			}
			authRule, err = impl.createApiAuthRule(dao.PackageId, dao.Id, false)
			if err != nil {
				goto failed
			}
			err = (*impl.ruleBiz).CreateRule(diceInfo, aclRule, session)
			if err != nil {
				goto failed
			}
			err = (*impl.ruleBiz).CreateRule(diceInfo, authRule, session)
			if err != nil {
				goto failed
			}
		}
		needUpdateDomainPolicy = true
	}
	if needUpdateDomainPolicy {
		err = (*impl.ruleBiz).SetPackageKongPolicies(pack, session)
		if err != nil {
			goto failed
		}
	}
	// MSE 网关 OpenAPI 流量入口中的路由，需要授权
	if gatewayProvider == mseCommon.MseProviderName && pack.Scene == orm.OpenapiScene && dao.Id != "" {
		packageAclInfoDto, err = (*impl.consumerBiz).GetPackageApiAcls(pack.Id, dao.Id)
		if err != nil {
			logrus.Errorf("Update package api authz for mse plugin failed: %v", err)
			goto failed
		}

		if len(packageAclInfoDto) > 0 {
			for _, aclInfo := range packageAclInfoDto {
				if aclInfo.Selected {
					consumers = append(consumers, aclInfo.Id)
				}
			}

			logrus.Errorf("UpdatePackageApiAcls for MSE plugin for pack.Id=%s, dao.Id=%s  consumers=%+v", pack.Id, dao.Id, consumers)
			go impl.UpdatePackageApiAclsWhenCreateApi(pack.Id, dao.Id, consumers)
		}
	}

	return false, dao.Id, audit, nil
failed:
	if serviceId != "" {
		_ = gatewayAdapter.DeleteService(serviceId)
	}
	if routeId != "" {
		_ = gatewayAdapter.DeleteRoute(routeId)
	}
	if zoneId != "" {
		zerr := (*impl.zoneBiz).DeleteZoneRoute(zoneId, ingressNamespace, session)
		if zerr != nil {
			logrus.Errorf("delete zone route failed, err:%+v", zerr)
		}
	}
	return exist, "", audit, err
}

func (impl GatewayOpenapiServiceImpl) UpdatePackageApiAclsWhenCreateApi(packageId, packageApiId string, consumerIds []string) {
	// MSE 网关路由是 MSE 控制器 watch ingress 发现，因此有延时，所以最多等待 30s
	time.Sleep(10 * time.Second)
	for i := 0; i < 3; i++ {
		dao, err := impl.packageApiDb.Get(packageApiId)
		if err != nil && i == 2 {
			logrus.Errorf("UpdatePackageApiAclsWhenCreateApi when get package api failed: %v", err)
		}
		if dao == nil && i == 2 {
			logrus.Errorf("UpdatePackageApiAclsWhenCreateApi when get package api failed: not found")
		}

		authzResult, err := (*impl.consumerBiz).UpdatePackageApiAcls(packageId, packageApiId, &gw.PackageAclsDto{Consumers: consumerIds})
		if err != nil && i == 2 {
			logrus.Errorf("UpdatePackageApiAclsWhenCreateApi when update package api authz failed: %v", err)
		}

		if !authzResult && i == 2 {
			logrus.Errorf("UpdatePackageApiAclsWhenCreateApi when update package api authz failed")
		}
		if authzResult {
			break
		}
		time.Sleep(10 * time.Second)
	}
}

func (impl GatewayOpenapiServiceImpl) touchServiceForExternalService(info endpoint_api.PackageApiInfo, z orm.GatewayZone) (*corev1.Service, error) {
	hostAndPathStr := strings.Split(info.RedirectAddr, "://")[1]
	hostAndPortStr := strings.Split(hostAndPathStr, "/")[0]
	protolHosts := strings.Split(hostAndPortStr, ":")
	HttpPort := 80
	externalName := protolHosts[0]
	if len(protolHosts) > 1 {
		HttpPort, _ = strconv.Atoi(protolHosts[1])
	}

	servicePorts := make([]corev1.ServicePort, 0)
	servicePorts = append(servicePorts, corev1.ServicePort{
		Name:       "target",
		Protocol:   corev1.ProtocolTCP,
		Port:       int32(HttpPort),
		TargetPort: intstr.FromInt(HttpPort),
	})

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.ToLower(z.Name),
			Namespace: "project-" + info.ProjectId + "-" + strings.ToLower(info.Env),
			Labels: map[string]string{
				"packageId":                info.PackageId,
				"erda.gateway.projectId":   info.ProjectId,
				"erda.gateway.appName":     info.DiceApp,
				"erda.gateway.serviceName": info.DiceService,
				"erda.gateway.workspace":   info.Env,
			},
		},
		Spec: corev1.ServiceSpec{
			ExternalName: externalName,
			Type:         corev1.ServiceTypeExternalName,
			Ports:        servicePorts,
		},
	}

	return service, nil
}

func (impl GatewayOpenapiServiceImpl) TouchPackageApiZone(info endpoint_api.PackageApiInfo, session ...*db.SessionHelper) (string, error) {
	var runtimeSession db.GatewayRuntimeServiceService
	var kongSession db.GatewayKongInfoService
	var externalSvc *corev1.Service
	var k8sAdapter k8s.K8SAdapter
	var err error
	if len(session) > 0 {
		runtimeSession, err = impl.runtimeDb.NewSession(session[0])
		if err != nil {
			return "", err
		}
		kongSession, err = impl.kongDb.NewSession(session[0])
		if err != nil {
			return "", err
		}
	} else {
		runtimeSession = impl.runtimeDb
		kongSession = impl.kongDb
	}

	kongInfo, err := kongSession.GetKongInfo(&orm.GatewayKongInfo{
		Az:        info.Az,
		ProjectId: info.ProjectId,
		Env:       info.Env,
	})
	if err != nil {
		return "", err
	}

	useKong := true
	az, azInfo, err := impl.azDb.GetAzInfoByClusterName(info.Az)
	if err != nil {
		return "", err
	}

	if azInfo != nil && azInfo.GatewayProvider != "" {
		useKong = false
	}
	// set route config
	routeConfig := zone.RouteConfig{
		Hosts: info.Hosts,
		Path:  info.ApiPath,
	}
	routeConfig.InjectRuntimeDomain = info.InjectRuntimeDomain
	var runtimeService *orm.GatewayRuntimeService
	if info.RedirectType == gw.RT_SERVICE {
		runtimeService, err = runtimeSession.Get(info.RuntimeServiceId)
		if err != nil {
			return "", err
		}
		if runtimeService == nil {
			return "", errors.Errorf("runtime service not found, err:%+v", err)
		}
		if runtimeService.IsSecurity == 1 {
			routeConfig.BackendProtocol = &[]k8s.BackendProtocl{k8s.HTTPS}[0]
		}
		_, innerHost, err := (*impl.globalBiz).GenerateEndpoint(gw.DiceInfo{
			Az:        info.Az,
			ProjectId: info.ProjectId,
			Env:       info.Env,
		})
		if err != nil {
			return "", err
		}
		routeConfig.UseRegex = true
		routeConfig.RewriteHost = &innerHost
		rewritePath := "/"
		if useKong {
			rewritePath, err = (*impl.globalBiz).GetRuntimeServicePrefix(runtimeService)
			if err != nil {
				return "", err
			}
		}
		// avoid dup slash in redirect path
		if strings.HasSuffix(info.RedirectPath, "/") {
			if routeConfig.Path != "/" {
				routeConfig.Path = strings.TrimSuffix(routeConfig.Path, "/")
				routeConfig.Path += "(/|$)(.*)"
				rewritePath += info.RedirectPath + "$2"
			} else {
				routeConfig.Path += "(.*)"
				rewritePath += info.RedirectPath + "$1"
			}
		} else {
			routeConfig.Path += "(.*)"
			rewritePath += info.RedirectPath + "$1"
		}
		routeConfig.RewritePath = &rewritePath

		runtimeServiceProjectName := kongInfo.ProjectName
		runtimeServiceProjectId := runtimeService.ProjectId
		runtimeServiceClusterName := runtimeService.ClusterName
		runtimeServiceRuntimeId := runtimeService.RuntimeId
		runtimeServiceServiceName := runtimeService.ServiceName
		runtimeServiceAppName := runtimeService.AppName
		routeConfig.Annotations = map[string]*string{
			k8s.Erda_Runtime_Project_Id:  &runtimeServiceProjectId,
			k8s.Erda_Runtime_ProjectName: &runtimeServiceProjectName,
			k8s.Erda_Runtime_ClusterName: &runtimeServiceClusterName,
			k8s.Erda_Runtime_Runtime_Id:  &runtimeServiceRuntimeId,
			k8s.Erda_Runtime_ServiceName: &runtimeServiceServiceName,
			k8s.Erda_Runtime_AppName:     &runtimeServiceAppName,
		}
	} else {
		routeConfig.Annotations = map[string]*string{}
		if useKong {
			routeConfig.BackendProtocol = &[]k8s.BackendProtocl{k8s.HTTPS}[0]
			routeConfig.UseRegex = true
			routeConfig.Path += ".*"
			routeConfig.Annotations = map[string]*string{}
			routeConfig.Annotations[k8s.REWRITE_HOST_KEY] = nil
			routeConfig.Annotations[k8s.REWRITE_PATH_KEY] = nil
		} else {
			rewriteHost := info.RedirectAddr
			if len(strings.Split(info.RedirectAddr, "://")) == 2 {
				rewriteHost = strings.Split(info.RedirectAddr, "://")[1]
			}
			rewriteHost = strings.Split(rewriteHost, "/")[0]
			rewriteHost = strings.Split(rewriteHost, ":")[0]
			routeConfig.Annotations[k8s.REWRITE_HOST_KEY] = &rewriteHost
			rewritePath := info.RedirectPath
			routeConfig.Annotations[k8s.REWRITE_PATH_KEY] = &rewritePath
		}
	}
	if info.ZoneId == "" {
		//create zone
		z, err := (*impl.zoneBiz).CreateZoneWithoutIngress(zone.ZoneConfig{
			ZoneRoute: &zone.ZoneRoute{routeConfig},
			Name:      "api-" + info.Id,
			ProjectId: info.ProjectId,
			Env:       info.Env,
			Az:        info.Az,
			Type:      db.ZONE_TYPE_PACKAGE_API,
		}, session...)
		if err != nil {
			return "", err
		}

		// mse for RedirectType == gw.RT_URL need create External Service
		if !useKong && info.RedirectType == gw.RT_URL {
			k8sAdapter, err = k8s.NewAdapter(info.Az)
			if err != nil {
				return "", err
			}

			externalSvc, err = impl.createOrUpdateService(k8sAdapter, info, *z)
			if err != nil {
				return "", err
			}
		}
		transSucc := false
		if info.PackageId != "" {
			// 表示不是从 SDK 注册的 api
			annotations, locationSnippet, retSession, err := (*impl.policyBiz).SetZoneDefaultPolicyConfig(info.PackageId, runtimeService, z, az, session...)
			if len(session) == 0 {
				defer func() {
					if transSucc {
						_ = retSession.Commit()
					} else {
						_ = retSession.Rollback()
					}
					retSession.Close()
				}()
			}
			if err != nil {
				if externalSvc != nil {
					if delSvcErr := k8sAdapter.DeleteService(externalSvc.Namespace, externalSvc.Name); delSvcErr != nil {
						logrus.Warnf("delete external service %s/%s failed: %v", externalSvc.Namespace, externalSvc.Name, delSvcErr)
					}
				}
				return "", err
			}

			if routeConfig.Annotations == nil {
				routeConfig.Annotations = make(map[string]*string)
			}
			for k, v := range annotations {
				routeConfig.Annotations[k] = v
			}
			routeConfig.LocationSnippet = locationSnippet
		}

		_, err = (*impl.zoneBiz).UpdateZoneRoute(z.Id, zone.ZoneRoute{routeConfig}, runtimeService, info.RedirectType, session...)
		if err != nil {
			if externalSvc != nil {
				if delSvcErr := k8sAdapter.DeleteService(externalSvc.Namespace, externalSvc.Name); delSvcErr != nil {
					logrus.Warnf("delete external service %s/%s failed: %v", externalSvc.Namespace, externalSvc.Name, delSvcErr)
				}
			}
			return "", err
		}
		transSucc = true
		return z.Id, nil
	} else {
		// 对于 MSE 需要更新单独创建的 Service
		if !useKong && info.RedirectType == gw.RT_URL {
			k8sAdapter, err = k8s.NewAdapter(info.Az)
			if err != nil {
				return "", err
			}
			z, err := (*impl.zoneBiz).GetZone(info.ZoneId, session...)
			if err != nil {
				return "", err
			}
			if z != nil {
				externalSvc, err = impl.createOrUpdateService(k8sAdapter, info, *z)
				if err != nil {
					return "", err
				}
			}
		}
	}

	//update zone route
	exist, err := (*impl.zoneBiz).UpdateZoneRoute(info.ZoneId, zone.ZoneRoute{routeConfig}, runtimeService, info.RedirectType)
	if err != nil {
		return "", err
	}
	if !exist {
		z, err := (*impl.zoneBiz).GetZone(info.ZoneId, session...)
		if err != nil {
			return "", err
		}
		if z != nil {
			err = (*impl.policyBiz).RefreshZoneIngress(*z, *az, runtimeService.ProjectNamespace, useKong)
			if err != nil {
				return "", err
			}
		}
	}
	return info.ZoneId, nil
}

func (impl GatewayOpenapiServiceImpl) createOrUpdateService(k8sAdapter k8s.K8SAdapter, info endpoint_api.PackageApiInfo, z orm.GatewayZone) (*corev1.Service, error) {
	svc, err := impl.touchServiceForExternalService(info, z)
	if err != nil {
		return nil, err
	}

	// 转发地址为内部地址，相当于此做一个内部 Service 的 Spec 拷贝
	if strings.Contains(svc.Spec.ExternalName, K8S_SVC_CLUSTER_DOMAIN) {
		err = copyService(svc, k8sAdapter)
		if err != nil {
			return nil, err
		}
	}

	externalSvc, err := k8sAdapter.CreateOrUpdateService(svc)
	if err != nil {
		return nil, err
	}
	return externalSvc, nil
}

func copyService(svc *corev1.Service, k8sAdapter k8s.K8SAdapter) error {
	svcNameAndNamespace := strings.Split(strings.TrimSuffix(svc.Spec.ExternalName, K8S_SVC_CLUSTER_DOMAIN), ".")
	if len(svcNameAndNamespace) <= 1 {
		return errors.Errorf("get svc name and namespace from inner addr %s failed\n", svc.Spec.ExternalName)
	}
	svcNamespace := svcNameAndNamespace[1]
	svcName := svcNameAndNamespace[0]
	innerSvc, err := k8sAdapter.GetServiceByName(svcNamespace, svcName)
	if err != nil {
		logrus.Errorf("GetServiceByName failed:%v\n", err)
		return err
	}
	svc.Spec.ExternalName = ""
	svc.Spec.Ports = innerSvc.Spec.Ports
	svc.Spec.Type = innerSvc.Spec.Type
	svc.Spec.Selector = innerSvc.Spec.Selector
	svc.Spec.SessionAffinity = innerSvc.Spec.SessionAffinity
	return nil
}

func (impl GatewayOpenapiServiceImpl) CreatePackageApi(id string, dto *gw.OpenapiDto) (apiId string, exist bool, err error) {
	var helper *db.SessionHelper
	defer func() {
		if err != nil {
			if helper != nil {
				_ = helper.Rollback()
				helper.Close()
			}
			logrus.Errorf("errorHappened, err:%+v", err)
		}
	}()
	if id == "" || (dto.ApiPath == "" && dto.Method == "") {
		err = errors.New("invalid request")
		return
	}
	var audit *apistructs.Audit
	defer func() {
		if audit != nil && err == nil {
			berr := bundle.Bundle.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: *audit})
			if berr != nil {
				logrus.Errorf("create audit failed, err:%+v", err)
			}
		}
	}()
	helper, err = db.NewSessionHelper()
	if err != nil {
		return
	}
	if exist, apiId, audit, err = impl.SessionCreatePackageApi(id, dto, helper, false); err != nil {
		return
	}
	_ = helper.Commit()
	helper.Close()
	return
}

func (impl GatewayOpenapiServiceImpl) openapiDto(dao *orm.GatewayPackageApi) *gw.OpenapiInfoDto {
	dto := &gw.OpenapiInfoDto{
		ApiId:       dao.Id,
		CreateAt:    dao.CreateTime.Format("2006-01-02T15:04:05"),
		DiceApp:     dao.DiceApp,
		DiceService: dao.DiceService,
		Origin:      gw.Origin(dao.Origin),
		OpenapiDto: gw.OpenapiDto{
			ApiPath:      dao.ApiPath,
			RedirectType: dao.RedirectType,
			Method:       dao.Method,
			Description:  dao.Description,
		},
	}
	if dao.AclType == gw.ACL_OFF {
		dto.AllowPassAuth = true
	}
	if dto.Origin == gw.FROM_DICE || dto.Origin == gw.FROM_SHADOW {
		dto.Mutable = false
	} else {
		dto.Mutable = true
	}
	if dao.RuntimeServiceId != "" {
		dto.RedirectPath = dao.RedirectPath
		runtimeService, err := impl.runtimeDb.Get(dao.RuntimeServiceId)
		if err != nil || runtimeService == nil {
			logrus.Errorf("get runtimeservice failed, err:%+v", err)
			return dto
		}
		dto.RedirectApp = runtimeService.AppName
		dto.RedirectService = runtimeService.ServiceName
		dto.RedirectRuntimeId = runtimeService.RuntimeId
		dto.RedirectRuntimeName = runtimeService.RuntimeName
		return dto
	}
	if dao.RedirectAddr != "" {
		dto.RedirectAddr = dao.RedirectAddr
		scheme_find := strings.Index(dto.RedirectAddr, "://")
		if scheme_find == -1 {
			logrus.Errorf("invalide RedirectAddr %s", dto.RedirectAddr)
			return dto
		}
		dto.RedirectPath = "/"
		slash_find := strings.Index(dto.RedirectAddr[scheme_find+3:], "/")
		if slash_find != -1 {
			dto.RedirectPath = dto.RedirectAddr[slash_find+scheme_find+3:]
			dto.RedirectAddr = dto.RedirectAddr[:slash_find+scheme_find+3]
		}
	}
	return dto
}

func (impl GatewayOpenapiServiceImpl) deleteApiPassAuthRules(apiId string) error {
	var rules []gw.OpenapiRuleInfo
	rules, err := (*impl.ruleBiz).GetApiRules(apiId, gw.AUTH_RULE)
	if err != nil {
		return err
	}
	if len(rules) > 0 {
		err = (*impl.ruleBiz).DeleteRule(rules[0].Id, nil)
		if err != nil {

			return err
		}
	}
	rules, err = (*impl.ruleBiz).GetApiRules(apiId, gw.ACL_RULE)
	if err != nil {
		return err
	}
	if len(rules) > 0 {
		err = (*impl.ruleBiz).DeleteRule(rules[0].Id, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) createApiPassAuthRules(pack *orm.GatewayPackage, apiId string) error {
	var authRule, aclRule *gw.OpenapiRule
	diceInfo := gw.DiceInfo{
		OrgId:     pack.DiceOrgId,
		ProjectId: pack.DiceProjectId,
		Env:       pack.DiceEnv,
		Az:        pack.DiceClusterName,
	}
	aclRule, err := impl.createApiAclRule(gw.ACL_OFF, pack.Id, apiId, pack.DiceClusterName)
	if err != nil {
		return err
	}
	authRule, err = impl.createApiAuthRule(pack.Id, apiId, false)
	if err != nil {
		return err
	}
	err = (*impl.ruleBiz).CreateRule(diceInfo, aclRule, nil)
	if err != nil {
		return err
	}
	err = (*impl.ruleBiz).CreateRule(diceInfo, authRule, nil)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) GetPackageApis(ctx context.Context, id string, args *gw.GetOpenapiDto) (result common.NewPageQuery, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("error happened, err:%+v", err)
		}
	}()
	pageInfo := common.NewPage2(args.PageSize, args.PageNo)
	options := args.GenSelectOptions()
	options = append(options, orm.SelectOption{
		Type:   orm.ExactMatch,
		Column: "package_id",
		Value:  id,
	})
	var list []gw.OpenapiInfoDto
	var daos []orm.GatewayPackageApi
	var ok bool
	pack, err := impl.packageDb.Get(id)
	if err != nil {
		return
	}
	if pack == nil {
		err = errors.Errorf("GatewayPackage not found with id %s\n", id)
		return
	}

	gatewayProvider, err := impl.GetGatewayProvider(pack.DiceClusterName)
	if err != nil {
		logrus.Errorf("get gateway provider for cluster %s failed: %v\n", pack.DiceClusterName, err)
		return
	}
	page, err := impl.packageApiDb.GetPage(options, (*common.Page)(pageInfo))
	if err != nil {
		return
	}
	daos, ok = page.Result.([]orm.GatewayPackageApi)
	if !ok {
		err = errors.New("type convert failed")
		return
	}
	for _, dao := range daos {
		dto := impl.openapiDto(&dao)
		dto.GatewayProvider = gatewayProvider
		list = append(list, *dto)
	}
	result = common.NewPages(list, pageInfo.TotalNum)
	return
}

func (impl GatewayOpenapiServiceImpl) ListPackageAllApis(id string) ([]orm.GatewayPackageApi, error) {
	return impl.packageApiDb.SelectByAny(&orm.GatewayPackageApi{PackageId: id})
}

func (impl GatewayOpenapiServiceImpl) UpdatePackageApi(packageId, apiId string, dto *gw.OpenapiDto) (result *gw.OpenapiInfoDto, exist bool, err error) {
	var dao, updateDao *orm.GatewayPackageApi
	var pack *orm.GatewayPackage
	var kongInfo *orm.GatewayKongInfo
	var gatewayAdapter gateway_providers.GatewayAdapter
	var unique bool
	var serviceId, routeId string
	var domains []string
	needUpdateDomainPolicy := false
	auditCtx := map[string]interface{}{}
	gatewayProvider := ""
	defer func() {
		if err != nil {
			logrus.Errorf("error happened, err:%+v", err)
		}
		if pack == nil {
			return
		}
		method := dto.Method
		if method == "" {
			method = "all"
		}
		auditCtx["endpoint"] = pack.PackageName
		auditCtx["path"] = dto.ApiPath
		auditCtx["method"] = method
		audit := common.MakeAuditInfo(impl.reqCtx, common.ScopeInfo{
			ProjectId: pack.DiceProjectId,
			Workspace: pack.DiceEnv,
		}, apistructs.UpdateRouteTemplate, err, auditCtx)
		if audit != nil {
			err = bundle.Bundle.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: *audit})
			if err != nil {
				logrus.Errorf("create audit failed, err:%+v", err)
			}
		}
	}()
	if packageId == "" || apiId == "" {
		err = errors.New("packageId or apiId is empty")
		return
	}
	if valid, msg := dto.CheckValid(); !valid {
		err = errors.New(msg)
		return
	}
	// get package domain
	domains, err = (*impl.domainBiz).GetPackageDomains(packageId)
	if err != nil {
		return
	}
	dto.Hosts = domains
	pack, err = impl.packageDb.Get(packageId)
	if err != nil {
		return
	}
	if pack == nil {
		err = errors.New("can't find endpoint")
		return
	}
	gatewayProvider, err = impl.GetGatewayProvider(pack.DiceClusterName)
	if err != nil {
		return
	}
	err = dto.Adjust()
	if err != nil {
		return
	}
	dao, err = impl.packageApiDb.Get(apiId)
	if err != nil {
		return
	}
	if dao == nil {
		err = errors.New("api not exist")
		return
	}

	auditCtx["endpoint"] = pack.PackageName
	if pack.Scene == orm.UnityScene || pack.Scene == orm.HubScene {
		var defaultPath string
		defaultPath, err = (*impl.globalBiz).GenerateDefaultPath(pack.DiceProjectId)
		if err != nil {
			return
		}
		if !strings.HasPrefix(dto.ApiPath, defaultPath) {
			dto.ApiPath = defaultPath + dto.ApiPath
		}
	}
	kongInfo, err = impl.kongDb.GetKongInfo(&orm.GatewayKongInfo{
		Az:        pack.DiceClusterName,
		ProjectId: pack.DiceProjectId,
		Env:       pack.DiceEnv,
	})
	if err != nil {
		return
	}
	switch gatewayProvider {
	case mseCommon.MseProviderName:
		gatewayAdapter, err = mse.NewMseAdapter(pack.DiceClusterName)
		if err != nil {
			return
		}
	case "":
		gatewayAdapter = kong.NewKongAdapter(kongInfo.KongAddr)
	default:
		logrus.Errorf("unknown gateway provider:%v\n", gatewayProvider)
		err = errors.Errorf("unknown gateway provider:%v\n", gatewayProvider)
		return
	}
	updateDao = impl.packageApiDao(dto)
	updateDao.ZoneId = dao.ZoneId
	//mutable check
	if dao.Origin == string(gw.FROM_DICE) || dao.Origin == string(gw.FROM_SHADOW) {
		if dto.RedirectAddr != dao.RedirectAddr {
			err = errors.Errorf("redirectaddr not same, old:%s, new:%s", dao.RedirectAddr,
				dto.RedirectAddr)
			return
		}
		if dto.Method != dao.Method {
			err = errors.Errorf("method not same, old:%s, new:%s", dao.Method,
				dto.Method)
			return
		}
	} else if dao.Origin == string(gw.FROM_CUSTOM) || dao.Origin == string(gw.FROM_DICEYML) {
		var zoneId string
		updateDao.PackageId = packageId
		updateDao.Id = apiId
		unique, err = impl.packageApiDb.CheckUnique(updateDao)
		if err != nil {
			return
		}
		if !unique {
			err = errors.Errorf("package api already exist, dao:%+v", updateDao)
			exist = true
			return
		}
		dto.ProjectId = pack.DiceProjectId
		dto.Env = pack.DiceEnv
		if updateDao.RedirectType == gw.RT_URL {
			updateDao.RuntimeServiceId = ""
			updateDao.DiceApp = ""
			updateDao.DiceService = ""
			if dto.RedirectAddr != dao.RedirectAddr || dto.ApiPath != dao.ApiPath || dto.Method != dao.Method {
				serviceId, err = impl.touchKongService(gatewayAdapter, dto, apiId)
				if err != nil {
					return
				}
				dto.ServiceId = serviceId
				routeId, err = impl.touchKongRoute(gatewayAdapter, dto, apiId)
				if err != nil {
					return
				}
				dto.RouteId = routeId
			} else {
				var route *orm.GatewayRoute
				route, err = impl.routeDb.GetByApiId(apiId)
				if err != nil {
					return
				}
				if route == nil {
					err = errors.Errorf("route of api:%s not exist", apiId)
					return
				}
				dto.RouteId = route.RouteId
			}
			err = impl.createOrUpdatePlugins(gatewayAdapter, dto)
			if err != nil {
				return
			}
		} else if updateDao.RedirectType == gw.RT_SERVICE {
			var runtimeService *orm.GatewayRuntimeService
			if dao.RedirectType == gw.RT_URL {
				err = impl.DeleteKongApi(gatewayAdapter, apiId)
				if err != nil {
					return
				}
			}
			updateDao.RedirectAddr = ""
			runtimeSession := impl.runtimeDb
			if dto.RuntimeServiceId == "" {
				runtimeService, err = runtimeSession.GetByAny(&orm.GatewayRuntimeService{
					AppName:     dto.RedirectApp,
					ServiceName: dto.RedirectService,
					RuntimeId:   dto.RedirectRuntimeId,
				})
			} else {
				runtimeService, err = runtimeSession.Get(dto.RuntimeServiceId)
			}
			if err != nil {
				return
			}
			if runtimeService == nil {
				err = errors.New("find runtime service failed")
				return
			}
			updateDao.RuntimeServiceId = runtimeService.Id
		}
		zoneId, err = impl.TouchPackageApiZone(endpoint_api.PackageApiInfo{updateDao, domains, pack.DiceProjectId, pack.DiceEnv, pack.DiceClusterName, false})
		if err != nil {
			return
		}
		updateDao.ZoneId = zoneId
	}
	err = impl.packageApiDb.Update(updateDao)
	if err != nil {
		return
	}
	if dao.AclType != updateDao.AclType {
		needUpdateDomainPolicy = true
		if updateDao.AclType == gw.ACL_NONE {
			err = impl.deleteApiPassAuthRules(apiId)
			if err != nil {
				return
			}
			// 可能在更新路由操作之前，已经手动调整过路由授权，因此需要恢复手动创建的调用方授权
			err = (*impl.consumerBiz).TouchPackageApiAclRules(packageId, apiId)
			if err != nil {
				return
			}

		} else if updateDao.AclType == gw.ACL_OFF {
			// 可能在更新路由操作之前，已经手动调整过路由授权，因此先要清理掉，然后再创建
			err = impl.deleteApiPassAuthRules(apiId)
			if err != nil {
				return
			}

			err = impl.createApiPassAuthRules(pack, apiId)
			if err != nil {
				return
			}
		}
	} else if dao.RedirectType != updateDao.RedirectType {
		needUpdateDomainPolicy = true
		if updateDao.AclType == gw.ACL_OFF {
			err = impl.deleteApiPassAuthRules(apiId)
			if err != nil {
				return
			}
			err = impl.createApiPassAuthRules(pack, apiId)
			if err != nil {
				return
			}
		}
	}
	if needUpdateDomainPolicy {
		err = (*impl.ruleBiz).SetPackageKongPolicies(pack, nil)
		if err != nil {
			return
		}
	}
	result = impl.openapiDto(dao)
	return
}

func (impl *GatewayOpenapiServiceImpl) DeletePackageApi(packageId, apiId string) (result bool, err error) {
	var pack *orm.GatewayPackage
	var dao *orm.GatewayPackageApi
	var kongInfo *orm.GatewayKongInfo
	var gatewayAdapter gateway_providers.GatewayAdapter
	var ingressNamespace string
	auditCtx := map[string]interface{}{}
	defer func() {
		if err != nil {
			logrus.Errorf("error happened, err:%+v", err)
		}
		if pack == nil || dao == nil {
			return
		}
		method := dao.Method
		if method == "" {
			method = "all"
		}
		auditCtx["endpoint"] = pack.PackageName
		auditCtx["path"] = dao.ApiPath
		auditCtx["method"] = method
		audit := common.MakeAuditInfo(impl.reqCtx, common.ScopeInfo{
			ProjectId: pack.DiceProjectId,
			Workspace: pack.DiceEnv,
		}, apistructs.DeleteRouteTemplate, err, auditCtx)
		if audit != nil {
			err = bundle.Bundle.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: *audit})
			if err != nil {
				logrus.Errorf("create audit failed, err:%+v", err)
			}
		}
	}()
	dao, err = impl.packageApiDb.Get(apiId)
	if err != nil {
		return
	}
	if dao == nil {
		result = true
		return
	}
	if dao.RuntimeServiceId != "" {
		runtimeService, err := impl.runtimeDb.Get(dao.RuntimeServiceId)
		if err != nil {
			return result, err
		}
		ingressNamespace = runtimeService.ProjectNamespace
	}
	pack, err = impl.packageDb.Get(packageId)
	if err != nil {
		return
	}

	gatewayProvider := ""
	gatewayProvider, err = impl.GetGatewayProvider(pack.DiceClusterName)
	if err != nil {
		return
	}

	kongInfo, err = impl.kongDb.GetKongInfo(&orm.GatewayKongInfo{
		Az:        pack.DiceClusterName,
		ProjectId: pack.DiceProjectId,
		Env:       pack.DiceEnv,
	})
	if err != nil {
		return
	}

	switch gatewayProvider {
	case mseCommon.MseProviderName:
		gatewayAdapter, err = mse.NewMseAdapter(pack.DiceClusterName)
		if err != nil {
			return
		}
		// 删除 MSE 全局插件配置中，相对于当前要删除的路由的配置项，采用
		aclRules := make([]gw.OpenapiRuleInfo, 0)
		aclRules, err = (*impl.ruleBiz).GetApiRules(apiId, gw.AUTH_RULE)
		if err != nil {
			return
		}

		// 因为是删除路由，路由相关的 mse 插件配置都需要删除，因此对应的 规则的 consumer 只保留默认的 consumer，用于后续逻辑处理
		// 其他 consumer 直接清除掉
		config := map[string]interface{}{}
		wlConsumers := make([]mseDto.Consumers, 0)
		wlConsumers = append(wlConsumers, mseDto.Consumers{
			Name: plugins.MseDefaultConsumerName,
		})

		config["whitelist"] = wlConsumers
		for _, rule := range aclRules {
			rule.Config = config
			_, err = (*impl.ruleBiz).CreateOrUpdatePlugin(gatewayProvider, gatewayAdapter, &rule.OpenapiRule, nil)
			if err != nil {
				return
			}
		}
		// Delete erda-ip erda-sbac erda-csrf plugin config for this api
		if dao.ZoneId != "" {
			var zone *orm.GatewayZone
			zone, err = (*impl.zoneBiz).GetZone(dao.ZoneId)
			if err != nil {
				return
			}
			policyPlugins := []string{mseCommon.MsePluginIP, mseCommon.MsePluginSbac, mseCommon.MsePluginCsrf}
			for _, pluginName := range policyPlugins {
				var pluginReq *providerDto.PluginReqDto
				switch pluginName {
				case mseCommon.MsePluginIP:
					pluginReq = &providerDto.PluginReqDto{
						Name: mseCommon.MsePluginIP,
						Config: map[string]interface{}{
							mseCommon.MseErdaIpIpSource:    mseCommon.MseErdaIpSourceXRealIP,
							mseCommon.MseErdaIpAclType:     mseCommon.MseErdaIpAclBlack,
							mseCommon.MseErdaIpAclList:     []string{},
							mseCommon.MseErdaIpRouteSwitch: false,
						},
						ZoneName: strings.ToLower(zone.Name),
					}
				case mseCommon.MsePluginSbac:
					pluginReq = &providerDto.PluginReqDto{
						Name: mseCommon.MsePluginSbac,
						Config: map[string]interface{}{
							//TODO: 补充相关配置信息
							mseCommon.MseErdaSBACRouteSwitch:            false,
							mseCommon.MseErdaSBACConfigAccessControlAPI: mseCommon.MseErdaSBACAccessControlAPI,
							mseCommon.MseErdaSBACConfigMatchPatterns:    []string{mseCommon.MseErdaSBACConfigDefaultMatchPattern},
							mseCommon.MseErdaSBACConfigHttpMethods: map[string]bool{
								http.MethodGet:     true,
								http.MethodHead:    true,
								http.MethodPost:    true,
								http.MethodPut:     true,
								http.MethodPatch:   true,
								http.MethodDelete:  true,
								http.MethodConnect: true,
								http.MethodOptions: true,
								http.MethodTrace:   true,
							},
							mseCommon.MseErdaSBACConfigWithHeaders: []string{mseCommon.MseErdaSBACConfigDefaultWithHeader},
							mseCommon.MseErdaSBACConfigWithCookie:  false,
						},
						ZoneName: strings.ToLower(zone.Name),
					}
				default:
					// mseCommon.MsePluginCsrf
					pluginReq = &providerDto.PluginReqDto{
						Name: mseCommon.MsePluginCsrf,
						Config: map[string]interface{}{
							//TODO: 补充相关配置信息
							mseCommon.MseErdaCSRFRouteSwitch:      false,
							mseCommon.MseErdaCSRFConfigUserCookie: []string{mseCommon.MseErdaCSRFDefaultUserCookie},
							mseCommon.MseErdaCSRFConfigExcludedMethod: []string{
								http.MethodGet,
								http.MethodHead,
								http.MethodOptions,
								http.MethodTrace,
							},
							mseCommon.MseErdaCSRFConfigTokenCookie:  mseCommon.MseErdaCSRFDefaultTokenName,
							mseCommon.MseErdaCSRFConfigTokenDomain:  mseCommon.MseErdaCSRFDefaultTokenDomain,
							mseCommon.MseErdaCSRFConfigCookieSecure: mseCommon.MseErdaCSRFDefaultCookieSecure,
							mseCommon.MseErdaCSRFConfigValidTTL:     mseCommon.MseErdaCSRFDefaultValidTTL,
							mseCommon.MseErdaCSRFConfigRefreshTTL:   mseCommon.MseErdaCSRFDefaultRefreshTTL,
							mseCommon.MseErdaCSRFConfigErrStatus:    mseCommon.MseErdaCSRFDefaultErrStatus,
							mseCommon.MseErdaCSRFConfigErrMsg:       mseCommon.MseErdaCSRFDefaultErrMsg,
							mseCommon.MseErdaCSRFConfigSecret:       mseCommon.MseErdaCSRFDefaultJWTSecret,
						},
						ZoneName: strings.ToLower(zone.Name),
					}
				}
				_, errr := gatewayAdapter.CreateOrUpdatePluginById(pluginReq)
				if errr != nil {
					err = errr
					return
				}
			}
		}

	case "":
		gatewayAdapter = kong.NewKongAdapter(kongInfo.KongAddr)
	default:
		logrus.Errorf("unknown gateway provider:%v\n", gatewayProvider)
		err = errors.Errorf("unknown gateway provider:%v\n", gatewayProvider)
		return
	}
	err = (*impl.ruleBiz).DeleteByPackageApi(pack, dao)
	if err != nil {
		return
	}
	if dao.Origin == string(gw.FROM_DICE) || dao.Origin == string(gw.FROM_SHADOW) {
		diceApiId := dao.DiceApiId
		var api *orm.GatewayApi
		var upApi *orm.GatewayUpstreamApi
		if diceApiId == "" {
			err = errors.Errorf("dice api id is empty, package api id:%s", apiId)
			return
		}
		err = impl.apiInPackageDb.Delete(packageId, diceApiId)
		if err != nil {
			return
		}
		api, err = impl.apiDb.GetById(diceApiId)
		if err != nil {
			return
		}
		if api != nil {
			if api.UpstreamApiId != "" {
				upApi, err = impl.upstreamApiDb.GetById(api.UpstreamApiId)
				if err != nil {
					return
				}
				if upApi != nil {
					err = (*impl.apiBiz).DeleteUpstreamBindApi(upApi)
					if err != nil {
						return
					}
					err = impl.upstreamApiDb.DeleteById(api.UpstreamApiId)
					if err != nil {
						return
					}
				}
			} else {
				(*impl.apiBiz).DeleteApi(api.Id)
			}
		}

	} else if dao.Origin == string(gw.FROM_CUSTOM) || dao.Origin == string(gw.FROM_DICEYML) {
		err = impl.DeleteKongApi(gatewayAdapter, apiId)
		if err != nil {
			return
		}
	}
	if dao.ZoneId != "" {
		if gatewayProvider == mseCommon.MseProviderName {
			impl.deleteMSEApi(dao, pack, ingressNamespace)
		}

		err = (*impl.zoneBiz).DeleteZone(dao.ZoneId, ingressNamespace)
		if err != nil {
			return
		}
	}
	err = impl.packageApiDb.Delete(apiId)
	if err != nil {
		return
	}
	result = true
	return
}

func (impl GatewayOpenapiServiceImpl) deleteMSEApi(dao *orm.GatewayPackageApi, pack *orm.GatewayPackage, ingressNamespace string) error {
	apiRedirectType := gw.RT_SERVICE
	var packageApi *orm.GatewayPackageApi
	packageApi, err := impl.packageApiDb.GetByAny(&orm.GatewayPackageApi{
		ZoneId: dao.ZoneId,
	})
	if err != nil {
		return err
	}
	if packageApi != nil {
		apiRedirectType = packageApi.RedirectType
	}

	if apiRedirectType == gw.RT_URL {
		// use mse for url redirect, need delete external Service
		if ingressNamespace == "" {
			ingressNamespace = "project-" + pack.DiceProjectId + "-" + strings.ToLower(pack.DiceEnv)
		}

		var zone *orm.GatewayZone
		zone, err = (*impl.zoneBiz).GetZone(dao.ZoneId)
		if err != nil {
			return err
		}
		if zone != nil {
			var k8sAdapter k8s.K8SAdapter
			k8sAdapter, err = k8s.NewAdapter(pack.DiceClusterName)
			if err != nil {
				return err
			}
			err = k8sAdapter.DeleteService(ingressNamespace, strings.ToLower(zone.Name))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) TouchRuntimePackageMeta(endpoint *orm.GatewayRuntimeService, session *db.SessionHelper) (string, bool, error) {
	var namespace, svcName string
	packSession, err := impl.packageDb.NewSession(session)
	if err != nil {
		return "", false, err
	}
	dao, err := packSession.GetByAny(&orm.GatewayPackage{
		RuntimeServiceId: endpoint.Id,
	})
	if err != nil {
		return "", false, err
	}
	if dao != nil {
		return dao.Id, false, nil
	}
	name := endpoint.AppName + "/" + endpoint.ServiceName + "/" + endpoint.RuntimeName
	namespace = endpoint.ProjectNamespace
	svcName = endpoint.ServiceName + "-" + endpoint.GroupName

	z, err := (*impl.zoneBiz).CreateZone(zone.ZoneConfig{
		Name:        regexp.MustCompile(`([^0-9a-zA-z]+|_+|\\+)`).ReplaceAllString("package-"+name, "-"),
		ProjectId:   endpoint.ProjectId,
		Env:         endpoint.Workspace,
		Az:          endpoint.ClusterName,
		Type:        db.ZONE_TYPE_PACKAGE_NEW,
		Namespace:   namespace,
		ServiceName: svcName,
	}, session)
	if err != nil {
		return "", false, err
	}

	orgId, _ := (*impl.globalBiz).GetOrgId(endpoint.ProjectId)
	pack := &orm.GatewayPackage{
		DiceOrgId:        orgId,
		DiceProjectId:    endpoint.ProjectId,
		DiceEnv:          endpoint.Workspace,
		DiceClusterName:  endpoint.ClusterName,
		PackageName:      name,
		Scene:            orm.WebapiScene,
		ZoneId:           z.Id,
		RuntimeServiceId: endpoint.Id,
	}
	unique, err := packSession.CheckUnique(pack)
	if err != nil {
		return "", false, err
	}
	if !unique {
		return "", false, errors.Errorf("package %s already exists", pack.PackageName)
	}
	err = packSession.Insert(pack)
	if err != nil {
		return "", false, err
	}
	return pack.Id, true, nil
}

func (impl GatewayOpenapiServiceImpl) RefreshRuntimePackage(packageId string, endpoint *orm.GatewayRuntimeService, session *db.SessionHelper) error {
	packSession, err := impl.packageDb.NewSession(session)
	if err != nil {
		return err
	}
	dao, err := packSession.Get(packageId)
	if err != nil {
		return err
	}
	if dao == nil {
		return nil
	}
	apiSession, err := impl.packageApiDb.NewSession(session)
	if err != nil {
		return err
	}
	apis, err := apiSession.SelectByAny(&orm.GatewayPackageApi{PackageId: packageId})
	if err != nil {
		return err
	}
	domains, err := (*impl.domainBiz).GetPackageDomains(packageId, session)
	if err != nil {
		return err
	}
	if len(domains) == 0 {
		return impl.TryClearRuntimePackage(endpoint, session, true)
	}
	wg := sync.WaitGroup{}
	for _, api := range apis {
		wg.Add(1)
		go func(apiRefer orm.GatewayPackageApi) {
			defer util.DoRecover()
			zoneId, aerr := impl.TouchPackageApiZone(endpoint_api.PackageApiInfo{&apiRefer, domains, dao.DiceProjectId, dao.DiceEnv, dao.DiceClusterName, false})
			if aerr != nil {
				err = aerr
			}
			if apiRefer.ZoneId == "" {
				apiRefer.ZoneId = zoneId
				aerr = apiSession.Update(&apiRefer)
				if aerr != nil {
					err = aerr
				}
			}
			wg.Done()
		}(api)
	}
	wg.Wait()
	if err != nil {
		return err
	}
	err = impl.updatePackageApiHost(dao, domains)
	if err != nil {
		return err
	}
	api, err := impl.packageApiDb.GetRawByAny(&orm.GatewayPackageApi{
		ApiPath:   "/",
		PackageId: packageId,
	})
	if err != nil {
		return err
	}
	if api == nil {
		_, _, _, err = impl.SessionCreatePackageApi(packageId, &gw.OpenapiDto{
			ApiPath:          "/",
			RedirectType:     gw.RT_SERVICE,
			RedirectPath:     "/",
			RedirectApp:      endpoint.AppName,
			RedirectService:  endpoint.ServiceName,
			RuntimeServiceId: endpoint.Id,
		}, session, true)
		if err != nil {
			return err
		}
	}

	err = (*impl.ruleBiz).SetPackageKongPolicies(dao, session)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) CreateUnityPackageZone(packageId string, session *db.SessionHelper) (*orm.GatewayZone, error) {
	packSession, err := impl.packageDb.NewSession(session)
	if err != nil {
		return nil, err
	}
	pack, err := packSession.Get(packageId)
	if err != nil {
		return nil, err
	}
	diceInfo := gw.DiceInfo{
		ProjectId: pack.DiceProjectId,
		Az:        pack.DiceClusterName,
		Env:       pack.DiceEnv,
	}
	outerHost, rewriteHost, err := (*impl.globalBiz).GenerateEndpoint(diceInfo)
	if err != nil {
		return nil, err
	}
	innerHost := (*impl.globalBiz).GetServiceAddr(diceInfo.Env)
	path, err := (*impl.globalBiz).GenerateDefaultPath(diceInfo.ProjectId)
	if err != nil {
		return nil, err
	}
	route := zone.RouteConfig{
		Hosts: []string{outerHost, innerHost},
		Path:  path + "/.*",
		RouteOptions: k8s.RouteOptions{
			BackendProtocol: &[]k8s.BackendProtocl{k8s.HTTPS}[0],
		},
	}
	route.RewriteHost = &rewriteHost
	route.UseRegex = true
	z, err := (*impl.zoneBiz).CreateZone(zone.ZoneConfig{
		ZoneRoute: &zone.ZoneRoute{Route: route},
		Name:      "unity",
		ProjectId: diceInfo.ProjectId,
		Env:       diceInfo.Env,
		Az:        diceInfo.Az,
		Type:      db.ZONE_TYPE_UNITY,
	}, session)
	if err != nil {
		return nil, err
	}
	pack.ZoneId = z.Id
	err = packSession.Update(pack)
	if err != nil {
		return nil, err
	}
	return z, nil
}

func (impl GatewayOpenapiServiceImpl) CreateTenantPackage(tenantId string, gatewayProvider string, session *db.SessionHelper) error {
	kongSession, err := impl.kongDb.NewSession(session)
	if err != nil {
		return err
	}
	kongInfo, err := kongSession.GetByAny(&orm.GatewayKongInfo{
		TenantId: tenantId,
	})
	if err != nil {
		return err
	}
	diceInfo := gw.DiceInfo{
		ProjectId: kongInfo.ProjectId,
		Az:        kongInfo.Az,
		Env:       kongInfo.Env,
	}
	outerHost, rewriteHost, err := (*impl.globalBiz).GenerateEndpoint(diceInfo, session)
	if err != nil {
		return err
	}
	innerHost := (*impl.globalBiz).GetServiceAddr(kongInfo.Env)
	path, err := (*impl.globalBiz).GenerateDefaultPath(kongInfo.ProjectId, session)
	if err != nil {
		return err
	}
	route := zone.RouteConfig{
		Hosts: []string{outerHost, innerHost},
		Path:  path + "/.*",
		RouteOptions: k8s.RouteOptions{
			BackendProtocol: &[]k8s.BackendProtocl{k8s.HTTPS}[0],
		},
	}
	route.RewriteHost = &rewriteHost
	route.UseRegex = true

	zoneConf := zone.ZoneConfig{
		ZoneRoute: &zone.ZoneRoute{Route: route},
		Name:      "unity",
		ProjectId: kongInfo.ProjectId,
		Env:       kongInfo.Env,
		Az:        kongInfo.Az,
		Type:      db.ZONE_TYPE_UNITY,
	}
	// 兼容 MSE 等网关，类似于做信息传递，后续处理逻辑中会恢复为  db.ZONE_TYPE_UNITY
	if gatewayProvider != "" {
		zoneConf.Type = db.ZONE_TYPE_UNITY_Provider
	}

	z, err := (*impl.zoneBiz).CreateZone(zoneConf, session)
	if err != nil {
		return err
	}

	var orgId string
	var pack *orm.GatewayPackage
	packageSession, err := impl.packageDb.NewSession(session)
	if err != nil {
		goto clear_route
	}
	orgId, err = (*impl.globalBiz).GetOrgId(kongInfo.ProjectId)
	if err != nil {
		goto clear_route
	}
	pack = &orm.GatewayPackage{
		DiceOrgId:       orgId,
		DiceProjectId:   kongInfo.ProjectId,
		DiceEnv:         kongInfo.Env,
		DiceClusterName: kongInfo.Az,
		PackageName:     "unity",
		Scene:           orm.UnityScene,
		ZoneId:          z.Id,
	}
	err = packageSession.Insert(pack)
	if err != nil {
		goto clear_route
	}

	_, err = (*impl.domainBiz).TouchPackageDomain(orgId, pack.Id, kongInfo.Az, []string{outerHost}, session)
	if err != nil {
		goto clear_route
	}
	switch gatewayProvider {
	case mseCommon.MseProviderName:
		logrus.Warnf("Mse gateway provider not support build-in plugin in kong")
	case "":
		_, _, err = (*impl.policyBiz).SetZonePolicyConfig(z, nil, apipolicy.Policy_Engine_Built_in, nil, session)
		if err != nil {
			goto clear_route
		}
	default:
		err = errors.Errorf("unknown gateway provider:%v\n", gatewayProvider)
		goto clear_route
	}
	{
		//TODO
		_ = session.Commit()
		switch gatewayProvider {
		case mseCommon.MseProviderName:
			logrus.Warnf("Mse gateway provider not support cors plugin in kong")
		case "":
			var policyEngine apipolicy.PolicyEngine
			var corsConfig apipolicy.PolicyDto
			var configByte []byte
			var azInfo *orm.GatewayAzInfo
			azInfo, _, err = impl.azDb.GetAzInfoByClusterName(kongInfo.Az)
			if err != nil {
				goto clear_route
			}
			if azInfo == nil {
				err = errors.New("az not found")
				goto clear_route
			}
			policyEngine, err = apipolicy.GetPolicyEngine(apipolicy.Policy_Engine_CORS)
			if err != nil {
				goto clear_route
			}
			corsConfig = policyEngine.CreateDefaultConfig(gatewayProvider, nil)
			corsConfig.SetEnable(true)
			configByte, err = json.Marshal(corsConfig)
			if err != nil {
				err = errors.WithStack(err)
				goto clear_route
			}
			_, err = (*impl.policyBiz).SetPackageDefaultPolicyConfig(apipolicy.Policy_Engine_CORS, pack.Id, azInfo, configByte)
			if err != nil {
				goto clear_route
			}
		default:
			err = errors.Errorf("unknown gateway provider:%v\n", gatewayProvider)
			goto clear_route
		}
	}
	return nil
clear_route:
	if gatewayProvider == "" {
		delErr := (*impl.zoneBiz).DeleteZoneRoute(z.Id, "", session)
		if delErr != nil {
			logrus.Errorf("delete zone failed, err:%+v", delErr)
		}
	}
	return err
}

func (impl GatewayOpenapiServiceImpl) CreateTenantHubPackages(ctx context.Context, tenantID string, session *db.SessionHelper) error {
	ctx = context1.WithLoggerIfWithout(ctx, logrus.StandardLogger())
	l := ctx.(*context1.LogContext).Entry().WithField("tenantID", tenantID)

	kongSession, err := impl.kongDb.NewSession(session)
	if err != nil {
		l.WithError(err).Errorln("failed to kongDB.NewSession")
		return err
	}
	kongInfo, err := kongSession.GetByAny(&orm.GatewayKongInfo{
		TenantId: tenantID,
	})
	if err != nil {
		l.WithError(err).Errorln("failed to kongSession.GetByAny")
		return err
	}
	orgID, err := (*impl.globalBiz).GetOrgId(kongInfo.ProjectId)
	if err != nil {
		l.WithError(err).WithField("projectID", kongInfo.ProjectId).Errorln("failed to GetOrgId")
		return err
	}
	return impl.createTenantHubPackages(ctx, orgID, kongInfo.ProjectId, kongInfo.Env, kongInfo.Az, session)
}

func (impl GatewayOpenapiServiceImpl) createTenantHubPackages(ctx context.Context, orgID, projectId, env, clusterName string, session *db.SessionHelper) error {
	ctx = context1.WithLoggerIfWithout(ctx, logrus.StandardLogger())
	l := ctx.(*context1.LogContext).Entry().WithFields(map[string]interface{}{
		"orgID":       orgID,
		"projectID":   projectId,
		"env":         env,
		"clusterName": clusterName,
	})

	// get tb_gateway_hub_info for this tenant
	hubInfoSession, err := impl.hubInfoDb.NewSession(session)
	if err != nil {
		return err
	}
	hubInfos, err := hubInfoSession.SelectByAny(&orm.GatewayHubInfo{
		OrgId:           orgID,
		DiceEnv:         env,
		DiceClusterName: clusterName,
	})
	if err != nil {
		l.WithError(err).Errorln("failed to hubInfoSession.SelectByAny")
		return err
	}
	if len(hubInfos) == 0 {
		l.WithError(err).Warnln("hub info not found")
		return nil
	}
	for _, hubInfo := range hubInfos {
		if _, err := impl.createOrGetTenantHubPackage(ctx, &hubInfo, orgID, projectId, env, clusterName, session); err != nil {
			return err
		}
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) createOrGetTenantHubPackage(ctx context.Context, hubInfo *orm.GatewayHubInfo, orgID, projectId, env,
	clusterName string, session *db.SessionHelper) (*orm.GatewayPackage, error) {
	var gatewayProvider string
	ctx = context1.WithLoggerIfWithout(ctx, logrus.StandardLogger())
	l := ctx.(*context1.LogContext).Entry().WithFields(map[string]interface{}{
		"orgID":       orgID,
		"projectID":   projectId,
		"env":         env,
		"clusterName": clusterName,
	})

	// check exist
	packageSession, err := impl.packageDb.NewSession(session)
	if err != nil {
		l.WithError(err).Errorln("failed to packageDb.NewSession")
		return nil, err
	}
	hubPkgs, err := impl.packageDb.SelectByAny(&orm.GatewayPackage{
		DiceProjectId: projectId,
		DiceEnv:       env,
		Scene:         orm.HubScene,
	})
	if err != nil {
		l.WithError(err).Errorln("failed to packageDb.SelectByAny")
		return nil, err
	}
	sort.Slice(hubPkgs, func(i, j int) bool {
		return hubPkgs[i].PackageName < hubPkgs[j].PackageName
	})
	hubDomains := strings.Split(hubInfo.BindDomain, ",")
	for _, pkg := range hubPkgs {
		pkgDomains := strings.Split(pkg.BindDomain, ",")
		for _, pkgDomain := range pkgDomains {
			for _, hubDomain := range hubDomains {
				if pkgDomain == hubDomain {
					l.WithField("current package domains", pkg.BindDomain).
						WithField("hub bind domains", hubInfo.BindDomain).
						Warnln("hub package with this domain exists")
					return &pkg, nil
				}
			}
		}
	}

	// construct zone route
	defaultPath, err := (*impl.globalBiz).GenerateDefaultPath(projectId, session)
	if err != nil {
		l.WithError(err).Errorln("failed to GenerateDefaultPath")
		return nil, err
	}
	route := zone.RouteConfig{
		Hosts: strings.Split(hubInfo.BindDomain, ","),
		Path:  defaultPath + "/*",
		RouteOptions: k8s.RouteOptions{
			BackendProtocol: &[]k8s.BackendProtocl{k8s.HTTPS}[0],
		},
	}
	rewriteHost := hubDomains[0]
	route.RewriteHost = &rewriteHost
	route.UseRegex = true

	// create zone
	z, err := (*impl.zoneBiz).CreateZone(zone.ZoneConfig{
		ZoneRoute: &zone.ZoneRoute{Route: route},
		Name:      orm.HubScene,
		ProjectId: projectId,
		Env:       env,
		Az:        clusterName,
		Type:      db.ZONE_TYPE_UNITY,
	})
	if err != nil {
		l.WithError(err).Errorln("failed to CreateZone")
		return nil, err
	}

	pkg := &orm.GatewayPackage{
		DiceOrgId:       orgID,
		DiceProjectId:   projectId,
		DiceEnv:         env,
		DiceClusterName: clusterName,
		ZoneId:          z.Id,
		PackageName:     orm.HubScene + "-" + strconv.Itoa(len(hubPkgs)),
		BindDomain:      hubInfo.BindDomain,
		Scene:           orm.HubScene,
	}
	if err := packageSession.Insert(pkg); err != nil {
		l.WithError(err).Errorln("failed to Insert package")
		goto clear_route
	}

	gatewayProvider, err = impl.GetGatewayProvider(clusterName)
	if err != nil {
		l.WithError(err).Errorf("get gateway provider failed for cluster %s: %v\n", clusterName, err)
		goto clear_route
	}
	if _, err = (*impl.domainBiz).TouchPackageDomain(orgID, pkg.Id, clusterName, route.Hosts, session); err != nil {
		l.WithError(err).Errorln("failed to TouchPackageDomain")
		goto clear_route
	}
	switch gatewayProvider {
	case mseCommon.MseProviderName:
		logrus.Warnf("mse gateway provider not support kong built-in policy yet\n")
	case "":
		if _, _, err = (*impl.policyBiz).SetZonePolicyConfig(z, nil, apipolicy.Policy_Engine_Built_in, nil, session); err != nil {
			l.WithError(err).Errorln("failed to SetZonePolicyConfig")
			goto clear_route
		}
	default:
		logrus.Errorf("unknown gateway provider:%v\n", gatewayProvider)
		err = errors.Errorf("unknown gateway provider:%v\n", gatewayProvider)
		goto clear_route
	}

	{
		var policyEngine apipolicy.PolicyEngine
		var corsConfig apipolicy.PolicyDto
		var configByte []byte
		var azInfo *orm.GatewayAzInfo
		azInfo, _, err = impl.azDb.GetAzInfoByClusterName(clusterName)
		if err != nil {
			goto clear_route
		}
		if azInfo == nil {
			err = errors.New("az not found")
			goto clear_route
		}
		policyEngine, err = apipolicy.GetPolicyEngine(apipolicy.Policy_Engine_CORS)
		if err != nil {
			goto clear_route
		}
		corsConfig = policyEngine.CreateDefaultConfig(gatewayProvider, nil)
		corsConfig.SetEnable(true)
		configByte, err = json.Marshal(corsConfig)
		if err != nil {
			err = errors.WithStack(err)
			goto clear_route
		}
		//TODO
		_ = session.Commit()
		switch gatewayProvider {
		case mseCommon.MseProviderName:
			logrus.Warnf("mse gateway provider not support kong cors policy yet\n")
		case "":
			_, err = (*impl.policyBiz).SetPackageDefaultPolicyConfig(apipolicy.Policy_Engine_CORS, pkg.Id, azInfo, configByte)
			if err != nil {
				goto clear_route
			}
		default:
			logrus.Errorf("unknown gateway provider:%v\n", gatewayProvider)
			err = errors.Errorf("unknown gateway provider:%v\n", gatewayProvider)
			goto clear_route
		}
	}

	return pkg, nil

clear_route:
	if delErr := (*impl.zoneBiz).DeleteZoneRoute(z.Id, "", session); delErr != nil {
		l.WithError(err).Errorln("failed to delete zone")
	}
	return nil, err
}

func (impl GatewayOpenapiServiceImpl) TouchPackageRootApi(packageId string, reqDto *gw.OpenapiDto) (result bool, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("error happened: %+v", err)
		}
	}()
	api, err := impl.packageApiDb.GetByAny(&orm.GatewayPackageApi{
		ApiPath:   "/",
		PackageId: packageId,
	})
	if err != nil {
		return
	}
	if !strings.HasPrefix(reqDto.RedirectAddr, "http://") && !strings.HasPrefix(reqDto.RedirectAddr, "https://") {
		reqDto.RedirectAddr = "http://" + reqDto.RedirectAddr
	}
	apiDto := &gw.OpenapiDto{
		ApiPath:      "/",
		RedirectType: gw.RT_URL,
		RedirectAddr: reqDto.RedirectAddr,
		RedirectPath: reqDto.RedirectPath,
	}
	if api == nil {
		_, _, err = impl.CreatePackageApi(packageId, apiDto)
		if err != nil {
			return
		}
		result = true
		return
	}
	_, _, err = impl.UpdatePackageApi(packageId, api.Id, apiDto)
	if err != nil {
		return
	}
	result = true
	return
}

func (impl GatewayOpenapiServiceImpl) endpointPadding(endpoint *diceyml.Endpoint, service *orm.GatewayRuntimeService) error {
	if strings.HasSuffix(endpoint.Domain, ".*") {
		azInfo, _, err := impl.azDb.GetAzInfoByClusterName(service.ClusterName)
		if err != nil {
			return err
		}
		if azInfo == nil {
			return errors.New("az not found")
		}
		endpoint.Domain = strings.Replace(endpoint.Domain, "*", azInfo.WildcardDomain, 1)
	}
	if endpoint.Path == "" {
		endpoint.Path = "/"
	}
	if endpoint.BackendPath == "" {
		endpoint.BackendPath = endpoint.Path
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) getPackageIdByDomain(domain, projectId, workspace string) (allowCreate bool, id string, err error) {
	var domains, allDomains []orm.GatewayDomain
	domains, err = (*impl.domainBiz).FindDomains(domain, projectId, workspace, orm.ExactMatch, orm.DT_PACKAGE)
	if err != nil {
		return
	}
	if len(domains) == 0 {
		allDomains, err = (*impl.domainBiz).FindDomains(domain, "", "", orm.ExactMatch)
		if err != nil {
			return
		}
		if len(allDomains) > 0 {
			return
		}
		allowCreate = true
		return
	}
	id = domains[0].PackageId
	return
}

func (impl GatewayOpenapiServiceImpl) createRuntimeEndpointPackage(ctx context.Context, domain string, service *orm.GatewayRuntimeService) (id string, err error) {
	packageName := domain
	namespace := service.ProjectNamespace
	svcName := service.ServiceName + "-" + service.GroupName
	orgId := ""
	orgId = apis.GetOrgID(ctx)
	if orgId == "" {
		orgId, _ = (*impl.globalBiz).GetOrgId(service.ProjectId)
	}

	dto, _, err := impl.CreatePackage(ctx, &gw.DiceArgsDto{
		OrgId:       orgId,
		ProjectId:   service.ProjectId,
		Env:         service.Workspace,
		Namespace:   namespace,
		ServiceName: svcName,
	}, &gw.PackageDto{
		Name:       packageName,
		BindDomain: []string{domain},
		Scene:      orm.WebapiScene,
	})
	if err != nil {
		err = errors.Errorf("create package failed, err:%v", err)
		return
	}
	id = dto.Id
	return
}

func (impl GatewayOpenapiServiceImpl) createRuntimeEndpointRoute(packageId string, endpoint diceyml.Endpoint, service *orm.GatewayRuntimeService) (id string, err error) {
	id, _, err = impl.CreatePackageApi(packageId, &gw.OpenapiDto{
		ApiPath:          endpoint.Path,
		RedirectPath:     endpoint.BackendPath,
		RedirectType:     gw.RT_SERVICE,
		RedirectApp:      service.AppName,
		RedirectService:  service.ServiceName,
		RuntimeServiceId: service.Id,
		Origin:           gw.FROM_DICEYML,
	})
	if err != nil {
		err = errors.Errorf("create package api failed, err:%v", err)
		return
	}
	return
}

func (impl GatewayOpenapiServiceImpl) updateRuntimeEndpointRoute(packageId, packageApiId string, endpoint diceyml.Endpoint, service *orm.GatewayRuntimeService) error {
	_, _, err := impl.UpdatePackageApi(packageId, packageApiId, &gw.OpenapiDto{
		ApiPath:          endpoint.Path,
		RedirectPath:     endpoint.BackendPath,
		RedirectType:     gw.RT_SERVICE,
		RedirectApp:      service.AppName,
		RedirectService:  service.ServiceName,
		RuntimeServiceId: service.Id,
		Origin:           gw.FROM_DICEYML,
	})
	if err != nil {
		return errors.Errorf("update package api failed, err:%v", err)
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) getPackageApiIdByRoute(packageId, path string, service *orm.GatewayRuntimeService) (allowChange bool, id string, err error) {
	api, err := impl.packageApiDb.GetByAny(&orm.GatewayPackageApi{
		ApiPath:   path,
		PackageId: packageId,
	})
	if err != nil {
		return
	}
	if api == nil {
		allowChange = true
		return
	}
	if api.RuntimeServiceId != service.Id {
		return
	}
	id = api.Id
	allowChange = true
	return
}

func (impl GatewayOpenapiServiceImpl) setRoutePolicy(category, packageId, packageApiId string, cf map[string]interface{}) error {
	engine, err := apipolicy.GetPolicyEngine(category)
	if err != nil {
		return err
	}
	config, err := engine.MergeDiceConfig(cf)
	if err != nil {
		return err
	}
	configByte, err := json.Marshal(config)
	if err != nil {
		return errors.Errorf("mashal failed, config:%v, err:%v",
			config, err)
	}
	_, err = (*impl.policyBiz).SetPolicyConfig(category, packageId, packageApiId, configByte)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) clearRoutePolicy(category, packageId, packageApiId string) error {
	engine, err := apipolicy.GetPolicyEngine(category)
	if err != nil {
		return err
	}
	pack, err := impl.packageDb.Get(packageId)
	if err != nil {
		return err
	}
	if pack == nil {
		return errors.New("endpoint not found")
	}

	gatewayProvider, err := impl.GetGatewayProvider(pack.DiceClusterName)
	if err != nil {
		errMsg := errors.Errorf("get gateway provider failed for cluster %s: %v\n", pack.DiceClusterName, err)
		logrus.Error(errMsg)
		return errMsg
	}

	config := engine.CreateDefaultConfig(gatewayProvider, map[string]interface{}{})
	configByte, err := json.Marshal(config)
	if err != nil {
		return errors.Errorf("mashal failed, config:%v, err:%v",
			config, err)
	}
	_, err = (*impl.policyBiz).SetPolicyConfig(category, packageId, packageApiId, configByte)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) setRuntimeEndpointRoutePolicy(packageId, packageApiId string, policies diceyml.EndpointPolicies) error {
	if policies.Cors != nil {
		err := impl.setRoutePolicy(apipolicy.Policy_Engine_CORS, packageId, packageApiId, *policies.Cors)
		if err != nil {
			return err
		}
	} else {
		// clear
		err := impl.clearRoutePolicy(apipolicy.Policy_Engine_CORS, packageId, packageApiId)
		if err != nil {
			return err
		}
	}
	if policies.RateLimit != nil {
		err := impl.setRoutePolicy("safety-server-guard", packageId, packageApiId, *policies.RateLimit)
		if err != nil {
			return err
		}
	} else {
		// clear
		err := impl.clearRoutePolicy("safety-server-guard", packageId, packageApiId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) clearRuntimeEndpointRoute(endpoints []endpointInfo, service *orm.GatewayRuntimeService) error {
	apis, err := impl.packageApiDb.SelectByAny(&orm.GatewayPackageApi{
		Origin:           string(gw.FROM_DICEYML),
		RuntimeServiceId: service.Id,
	})
	if err != nil {
		return nil
	}
	checkPackageEmpty := map[string]bool{}
	for _, api := range apis {
		exist := false
		for _, endpoint := range endpoints {
			if api.Id == endpoint.apiId {
				exist = true
				break
			}
		}
		if !exist {
			_, err := impl.DeletePackageApi(api.PackageId, api.Id)
			if err != nil {
				return err
			}
			checkPackageEmpty[api.PackageId] = true
		}
	}
	for pack := range checkPackageEmpty {
		apis, err := impl.packageApiDb.SelectByAny(&orm.GatewayPackageApi{
			PackageId: pack,
		})
		if err != nil {
			return err
		}
		// empty package, clear it
		if len(apis) == 0 {
			err = impl.deletePackage(pack, nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type endpointInfo struct {
	diceyml.Endpoint
	apiId string
}

func (impl GatewayOpenapiServiceImpl) SetRuntimeEndpoint(ctx context.Context, info runtime_service.RuntimeEndpointInfo) error {
	ctx = context1.WithLoggerIfWithout(ctx, logrus.StandardLogger())

	// get org locale
	var locale string
	if orgDTO, ok := orgCache.GetOrgByProjectID(info.RuntimeService.ProjectId); ok {
		locale = orgDTO.Locale
	}

	var endpoints []endpointInfo
	for _, endpoint := range info.Endpoints {
		err := impl.endpointPadding(&endpoint, info.RuntimeService)
		if err != nil {
			return err
		}
		allowCreate, packageId, err := impl.getPackageIdByDomain(endpoint.Domain, info.RuntimeService.ProjectId, info.RuntimeService.Workspace)
		if err != nil {
			return err
		}
		if !allowCreate && packageId == "" {
			rawLog := fmt.Sprintf("app:%s service:%s endpoint create failed,  domain:%s already used by other one",
				info.RuntimeService.AppName, info.RuntimeService.ServiceName, endpoint.Domain)
			humanLog := i18n.Sprintf(locale, "FailedToBindServiceDomain.Occupied", info.RuntimeService.ServiceName, endpoint.Domain)
			logrus.Error(rawLog)
			go common.AsyncRuntimeError(info.RuntimeService.RuntimeId, humanLog, rawLog)
			continue
		}
		if allowCreate {
			packageId, err := impl.createRuntimeEndpointPackage(ctx, endpoint.Domain, info.RuntimeService)
			if err != nil {
				return err
			}
			packageApiId, err := impl.createRuntimeEndpointRoute(packageId, endpoint, info.RuntimeService)
			if err != nil {
				return err
			}
			err = impl.setRuntimeEndpointRoutePolicy(packageId, packageApiId, endpoint.Policies)
			if err != nil {
				return err
			}
			endpoints = append(endpoints, endpointInfo{Endpoint: endpoint, apiId: packageApiId})
			continue
		}
		if packageId != "" {
			allowChange, packageApiId, err := impl.getPackageApiIdByRoute(packageId, endpoint.Path, info.RuntimeService)
			if err != nil {
				return err
			}
			if !allowChange {
				rawLog := fmt.Sprintf("app:%s service:%s endpoint route create failed,  domain:%s path:%s already used by other one",
					info.RuntimeService.AppName, info.RuntimeService.ServiceName, endpoint.Domain, endpoint.Path)
				humanLog := i18n.Sprintf(locale, "FailedToBindServiceDomainRoute.Occupied", info.RuntimeService.ServiceName,
					endpoint.Domain, endpoint.Path)
				go common.AsyncRuntimeError(info.RuntimeService.RuntimeId, humanLog, rawLog)
				continue
			}
			if packageApiId != "" {
				err = impl.updateRuntimeEndpointRoute(packageId, packageApiId, endpoint, info.RuntimeService)
			} else {
				packageApiId, err = impl.createRuntimeEndpointRoute(packageId, endpoint, info.RuntimeService)
			}
			if err != nil {
				return err
			}
			err = impl.setRuntimeEndpointRoutePolicy(packageId, packageApiId, endpoint.Policies)
			if err != nil {
				return err
			}
			endpoints = append(endpoints, endpointInfo{Endpoint: endpoint, apiId: packageApiId})
		}
	}
	err := impl.clearRuntimeEndpointRoute(endpoints, info.RuntimeService)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) ClearRuntimeRoute(id string) error {
	apis, err := impl.packageApiDb.SelectByAny(&orm.GatewayPackageApi{
		RuntimeServiceId: id,
	})
	if err != nil {
		return err
	}
	checkPackageEmpty := map[string]bool{}
	for _, api := range apis {
		_, err = impl.DeletePackageApi(api.PackageId, api.Id)
		if err != nil {
			return err
		}
		checkPackageEmpty[api.PackageId] = true
	}
	for pack := range checkPackageEmpty {
		apis, err := impl.packageApiDb.SelectByAny(&orm.GatewayPackageApi{
			PackageId: pack,
		})
		if err != nil {
			return err
		}
		// empty package, clear it
		if len(apis) == 0 {
			err = impl.deletePackage(pack, nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func hubExists(hubInfos []orm.GatewayHubInfo, bindDomain []string) (*orm.GatewayHubInfo, bool) {
	if len(hubInfos) == 0 {
		return nil, false
	}
	for _, bindDomain := range bindDomain {
		for _, hubInfo := range hubInfos {
			hubInfoDomains := strings.Split(hubInfo.BindDomain, ",")
			for _, hubDomain := range hubInfoDomains {
				if bindDomain == hubDomain {
					return &hubInfo, true
				}
			}
		}
	}
	return nil, false
}
