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
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	aliyun_errors "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cloudapi"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/hepa/apipolicy"
	"github.com/erda-project/erda/modules/hepa/bundle"
	"github.com/erda-project/erda/modules/hepa/common"
	"github.com/erda-project/erda/modules/hepa/common/util"
	"github.com/erda-project/erda/modules/hepa/common/vars"
	. "github.com/erda-project/erda/modules/hepa/common/vars"
	"github.com/erda-project/erda/modules/hepa/config"
	gw "github.com/erda-project/erda/modules/hepa/gateway/dto"
	"github.com/erda-project/erda/modules/hepa/k8s"
	"github.com/erda-project/erda/modules/hepa/kong"
	kongDto "github.com/erda-project/erda/modules/hepa/kong/dto"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
	db "github.com/erda-project/erda/modules/hepa/repository/service"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	"github.com/erda-project/erda/pkg/parser/diceyml"
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
	apiBiz          GatewayApiService
	zoneBiz         GatewayZoneService
	ruleBiz         GatewayOpenapiRuleService
	consumerBiz     GatewayOpenapiConsumerService
	globalBiz       GatewayGlobalService
	policyBiz       GatewayApiPolicyService
	runtimeDb       db.GatewayRuntimeServiceService
	domainBiz       GatewayDomainService
	ctx             context.Context
	ReqCtx          *gin.Context
}

func NewGatewayOpenapiServiceImpl() (*GatewayOpenapiServiceImpl, error) {
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
	apiBiz, _ := NewGatewayApiServiceImpl()
	zoneBiz, _ := NewGatewayZoneServiceImpl()
	consumerBiz, _ := NewGatewayOpenapiConsumerServiceImpl()
	ruleBiz, _ := NewGatewayOpenapiRuleServiceImpl()
	globalBiz, _ := NewGatewayGlobalServiceImpl()
	policyBiz, _ := NewGatewayApiPolicyServiceImpl()
	runtimeDb, _ := db.NewGatewayRuntimeServiceServiceImpl()
	domainBiz, _ := NewGatewayDomainServiceImpl()
	return &GatewayOpenapiServiceImpl{
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
		zoneBiz:         zoneBiz,
		apiBiz:          apiBiz,
		consumerBiz:     consumerBiz,
		ruleBiz:         ruleBiz,
		globalBiz:       globalBiz,
		policyBiz:       policyBiz,
		runtimeDb:       runtimeDb,
		domainBiz:       domainBiz,
	}, nil
}

func (impl GatewayOpenapiServiceImpl) createApiAclRule(aclType, packageId, apiId string) (*gw.OpenapiRule, error) {
	rule, err := impl.createAclRule(aclType, packageId)
	if err != nil {
		return nil, err
	}
	rule.PackageApiId = apiId
	rule.Region = gw.API_RULE
	return rule, nil
}

func (impl GatewayOpenapiServiceImpl) createAclRule(aclType, packageId string) (*gw.OpenapiRule, error) {
	var consumers []orm.GatewayConsumer
	var aclRule *gw.OpenapiRule
	consumers, err := impl.consumerBiz.GetConsumersOfPackage(packageId)
	if err != nil {
		return nil, err
	}
	aclRule = &gw.OpenapiRule{
		PackageId:  packageId,
		PluginName: gw.ACL,
		Category:   gw.ACL_RULE,
	}
	var buffer bytes.Buffer
	for _, consumer := range consumers {
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(impl.consumerBiz.GetKongConsumerName(&consumer))
	}
	wl := buffer.String()
	config := map[string]interface{}{}
	if wl == "" {
		wl = ","
	}
	config["whitelist"] = wl
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
	switch authType {
	case gw.AT_KEY_AUTH:
		authRule.Config = gw.KEYAUTH_CONFIG
	case gw.AT_OAUTH2:
		authRule.Config = gw.OAUTH2_CONFIG
	case gw.AT_SIGN_AUTH:
		authRule.Config = gw.SIGNAUTH_CONFIG
	case gw.AT_HMAC_AUTH:
		kongInfo, err := impl.kongDb.GetKongInfo(&orm.GatewayKongInfo{
			Az:        pack.DiceClusterName,
			ProjectId: pack.DiceProjectId,
			Env:       pack.DiceEnv,
		})
		if err != nil {
			return nil, err
		}
		kongAdapter := kong.NewKongAdapter(kongInfo.KongAddr)
		enabled, err := kongAdapter.CheckPluginEnabled(gw.AT_HMAC_AUTH)
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
	return authRule, nil
}

func (impl GatewayOpenapiServiceImpl) CreatePackage(args *gw.DiceArgsDto, dto *gw.PackageDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if args.ProjectId == "" || args.Env == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	dto.Name = strings.TrimSpace(dto.Name)
	for i := 0; i < len(dto.BindDomain); i++ {
		dto.BindDomain[i] = strings.TrimSpace(dto.BindDomain[i])
	}
	dto.BindDomain = util.UniqStringSlice(dto.BindDomain)
	var diceInfo DiceInfo
	var helper *db.SessionHelper
	var pack *orm.GatewayPackage
	var aclRule, authRule *gw.OpenapiRule
	var packSession db.GatewayPackageService
	var zone *orm.GatewayZone
	var unique bool
	var domains []string
	var err error
	auditCtx := map[string]interface{}{}
	auditCtx["endpoint"] = dto.Name
	auditCtx["domains"] = strings.Join(dto.BindDomain, ", ")
	defer func() {
		audit := common.MakeAuditInfo(impl.ReqCtx, common.ScopeInfo{
			ProjectId: args.ProjectId,
			Workspace: args.Env,
		}, apistructs.CreateEndpointTemplate, res, auditCtx)
		if audit != nil {
			err = bundle.Bundle.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: *audit})
			if err != nil {
				log.Errorf("create audit failed, err:%+v", err)
			}
		}
	}()
	az, err := impl.azDb.GetAz(&orm.GatewayAzInfo{
		Env:       args.Env,
		ProjectId: args.ProjectId,
	})
	if err != nil {
		goto failed
	}
	err = dto.CheckValid()
	if err != nil {
		goto failed
	}
	diceInfo = DiceInfo{
		ProjectId: args.ProjectId,
		Env:       args.Env,
		Az:        az,
	}
	helper, err = db.NewSessionHelper()
	if err != nil {
		goto failed
	}
	// create self zone
	zone, err = impl.zoneBiz.CreateZone(ZoneConfig{
		Name:      "package-" + dto.Name,
		ProjectId: args.ProjectId,
		Env:       args.Env,
		Az:        az,
		Type:      db.ZONE_TYPE_PACKAGE_NEW,
	}, helper)
	if err != nil {
		goto failed
	}
	// create package in db
	packSession, err = impl.packageDb.NewSession(helper)
	if err != nil {
		goto failed
	}
	pack = &orm.GatewayPackage{
		DiceProjectId:   args.ProjectId,
		DiceEnv:         args.Env,
		DiceClusterName: az,
		ZoneId:          zone.Id,
		PackageName:     dto.Name,
		AuthType:        dto.AuthType,
		AclType:         dto.AclType,
		Scene:           dto.Scene,
		Description:     dto.Description,
	}
	if dto.NeedBindCloudapi {
		pack.CloudapiNeedBind = 1
	}
	unique, err = packSession.CheckUnique(pack)
	if err != nil {
		goto failed
	}
	if !unique {
		err = errors.Errorf("package %s already exists", pack.PackageName)
		res.SetReturnCode(PACKAGE_EXIST)
		goto failed
	}
	err = packSession.Insert(pack)
	if err != nil {
		goto failed
	}
	if dto.NeedBindCloudapi {
		go func() {
			defer util.DoRecover()
			verr := impl.setCloudapiGroupBind(args.OrgId, pack, dto.BindDomain)
			if verr != nil {
				pack.CloudapiNeedBind = 0
				_ = impl.packageDb.Update(pack, "cloudapi_need_bind")
				log.Errorf("error happened:%+v", verr)
			}
		}()
	}
	domains, err = impl.domainBiz.TouchPackageDomain(pack.Id, az, dto.BindDomain, helper)
	if err != nil {
		goto failed
	}
	if dto.Scene == orm.OPENAPI_SCENE {
		// create auth, acl rule
		authRule, err = impl.createAuthRule(dto.AuthType, pack)
		if err != nil {
			goto failed
		}
		authRule.Region = gw.PACKAGE_RULE
		aclRule, err = impl.createAclRule(dto.AclType, pack.Id)
		if err != nil {
			goto failed
		}
		aclRule.Region = gw.PACKAGE_RULE
		err = impl.ruleBiz.CreateRule(diceInfo, authRule, helper)
		if err != nil {
			goto failed
		}
		err = impl.ruleBiz.CreateRule(diceInfo, aclRule, helper)
		if err != nil {
			goto failed
		}
		// update zone kong polices
		err = impl.ruleBiz.SetPackageKongPolicies(pack, helper)
		if err != nil {
			goto failed
		}
	}
	err = helper.Commit()
	if err != nil {
		goto failed
	}
	helper.Close()
	return res.SetSuccessAndData(impl.packageDto(pack, domains))
failed:
	// failed: delete self zone, session rollback
	log.Errorf("error happened, err:%+v", err)
	if helper != nil {
		_ = helper.Rollback()
		helper.Close()
	}
	return res.SetErrorInfo(&common.ErrInfo{Msg: errors.Cause(err).Error()})
}

func (impl GatewayOpenapiServiceImpl) GetPackage(id string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if id == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	dao, err := impl.packageDb.Get(id)
	var dto *gw.PackageInfoDto
	var domains []string
	if err != nil {
		goto failed
	}
	if dao == nil {
		err = errors.Errorf("package not found, packageId: %s", id)
		goto failed
	}
	domains, err = impl.domainBiz.GetPackageDomains(dao.Id)
	if err != nil {
		goto failed
	}
	dto = impl.packageDto(dao, domains)
	return res.SetSuccessAndData(dto)
failed:
	log.Errorf("error happened, err:%+v", err)
	return res
}

func (impl GatewayOpenapiServiceImpl) GetPackages(args *gw.GetPackagesDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if args.ProjectId == "" || args.Env == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	var domainInfos []orm.GatewayDomain
	var packageIds []string
	var err error
	if args.Domain != "" {
		domainInfos, err = impl.domainBiz.FindDomains(args.Domain, args.ProjectId, args.Env, orm.PrefixMatch, orm.DT_PACKAGE)
		if err != nil {
			log.Errorf("err:%+v", err)
			return res.SetReturnCode(PARAMS_IS_NULL)
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
		return res.SetSuccessAndData(common.NewPages(nil, 0))
	}
	var list []gw.PackageInfoDto
	var daos []orm.GatewayPackage
	var ok bool
	page, err := impl.packageDb.GetPage(options, pageInfo)
	if err != nil {
		goto failed
	}
	daos, ok = page.Result.([]orm.GatewayPackage)
	if !ok {
		err = errors.New("type convert failed")
		goto failed
	}

	for _, dao := range daos {
		var dto *gw.PackageInfoDto
		domains, err := impl.domainBiz.GetPackageDomains(dao.Id)
		if err != nil {
			goto failed
		}
		dto = impl.packageDto(&dao, domains)
		list = append(list, *dto)
	}
	return res.SetSuccessAndData(common.NewPages(list, pageInfo.TotalNum))
failed:
	log.Errorf("error happened, err:%+v", err)
	return res
}

func (impl GatewayOpenapiServiceImpl) packageDto(dao *orm.GatewayPackage, domains []string) *gw.PackageInfoDto {
	res := &gw.PackageInfoDto{
		Id:       dao.Id,
		CreateAt: dao.CreateTime.Format("2006-01-02T15:04:05"),
		PackageDto: gw.PackageDto{
			Name:        dao.PackageName,
			BindDomain:  domains,
			AuthType:    dao.AuthType,
			AclType:     dao.AclType,
			Description: dao.Description,
			Scene:       dao.Scene,
		},
	}
	if dao.CloudapiNeedBind == 1 {
		res.PackageDto.NeedBindCloudapi = true
	}
	return res
}

func (impl GatewayOpenapiServiceImpl) GetPackagesName(args *gw.GetPackagesDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if args.ProjectId == "" || args.Env == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	list := []gw.PackageInfoDto{}
	var daos []orm.GatewayPackage
	az, err := impl.azDb.GetAz(&orm.GatewayAzInfo{
		Env:       args.Env,
		ProjectId: args.ProjectId,
	})
	if err != nil {
		goto failed
	}
	daos, err = impl.packageDb.SelectByAny(&orm.GatewayPackage{
		DiceProjectId:   args.ProjectId,
		DiceEnv:         args.Env,
		DiceClusterName: az,
	})
	if err != nil {
		goto failed
	}
	for _, dao := range daos {
		var dto *gw.PackageInfoDto
		domains, err := impl.domainBiz.GetPackageDomains(dao.Id)
		if err != nil {
			goto failed
		}
		dto = impl.packageDto(&dao, domains)
		list = append(list, *dto)
	}
	return res.SetSuccessAndData(list)
failed:
	log.Errorf("error happened, err:%+v", err)
	return res
}

func (impl GatewayOpenapiServiceImpl) updatePackageApiHost(pack *orm.GatewayPackage, hosts []string) error {
	kongInfo, err := impl.kongDb.GetKongInfo(&orm.GatewayKongInfo{
		Az:        pack.DiceClusterName,
		ProjectId: pack.DiceProjectId,
		Env:       pack.DiceEnv,
	})
	if err != nil {
		return err
	}
	kongAdapter := kong.NewKongAdapter(kongInfo.KongAddr)
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
		req := &kongDto.KongRouteReqDto{
			RouteId: route.RouteId,
			Hosts:   hosts,
		}
		resp, err := kongAdapter.UpdateRoute(req)
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

func (impl GatewayOpenapiServiceImpl) UpdatePackage(id string, dto *gw.PackageDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if id == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	needUpdate := false
	needRefreshCloudapi := false
	var diceInfo DiceInfo
	var dao *orm.GatewayPackage
	var zone *orm.GatewayZone
	var apiSession db.GatewayPackageApiService
	var domains []string
	var apis []orm.GatewayPackageApi
	var err error
	auditCtx := map[string]interface{}{}
	auditCtx["endpoint"] = dto.Name
	auditCtx["domains"] = strings.Join(dto.BindDomain, ", ")
	defer func() {
		audit := common.MakeAuditInfo(impl.ReqCtx, common.ScopeInfo{
			ProjectId: dao.DiceProjectId,
			Workspace: dao.DiceEnv,
		}, apistructs.UpdateEndpointTemplate, res, auditCtx)
		if audit != nil {
			err = bundle.Bundle.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: *audit})
			if err != nil {
				log.Errorf("create audit failed, err:%+v", err)
			}
		}
	}()
	domainHasDiff := false
	session, err := db.NewSessionHelper()
	if err != nil {
		goto failed
	}
	err = dto.CheckValid()
	if err != nil {
		goto failed
	}
	for i := 0; i < len(dto.BindDomain); i++ {
		dto.BindDomain[i] = strings.TrimSpace(dto.BindDomain[i])
	}
	dto.BindDomain = util.UniqStringSlice(dto.BindDomain)
	dao, err = impl.packageDb.Get(id)
	if err != nil {
		goto failed
	}
	diceInfo = DiceInfo{
		ProjectId: dao.DiceProjectId,
		Env:       dao.DiceEnv,
		Az:        dao.DiceClusterName,
	}
	domainHasDiff, err = impl.domainBiz.IsPackageDomainsDiff(id, dao.DiceClusterName, dto.BindDomain, session)
	if err != nil {
		goto failed
	}
	if domainHasDiff {
		needUpdate = true
		domains, err = impl.domainBiz.TouchPackageDomain(id, dao.DiceClusterName, dto.BindDomain, session)
		if err != nil {
			goto failed
		}

		zone, err = impl.zoneBiz.GetZone(dao.ZoneId, session)
		if err != nil {
			goto failed
		}
		//老类型兼容
		if zone.Type == db.ZONE_TYPE_PACKAGE {
			_, err = impl.zoneBiz.UpdateZoneRoute(dao.ZoneId, ZoneRoute{
				RouteConfig{
					Hosts: domains,
					Path:  "/",
				},
			})
			if err != nil {
				goto failed
			}
		}
		apiSession, err = impl.packageApiDb.NewSession(session)
		if err != nil {
			goto failed
		}
		apis, err = apiSession.SelectByAny(&orm.GatewayPackageApi{PackageId: id})
		if err != nil {
			goto failed
		}
		wg := sync.WaitGroup{}
		for _, api := range apis {
			wg.Add(1)
			go func(apiRefer orm.GatewayPackageApi) {
				defer doRecover()
				log.Debugf("begin touch pacakge api zone, api:%+v", apiRefer)
				zoneId, aerr := impl.TouchPackageApiZone(PackageApiInfo{&apiRefer, domains, dao.DiceProjectId, dao.DiceEnv, dao.DiceClusterName, false})
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
				log.Debugf("complete touch pacakge api zone, api:%+v", apiRefer)
			}(api)
		}
		wg.Wait()
		if err != nil {
			goto failed
		}
		err = impl.updatePackageApiHost(dao, domains)
		if err != nil {
			goto failed
		}
	}
	if dao.AuthType != dto.AuthType {
		if dao.AuthType == gw.AT_ALIYUN_APP || dto.AuthType == gw.AT_ALIYUN_APP {
			needRefreshCloudapi = true
		}
		needUpdate = true
		var oldAuthRule []gw.OpenapiRuleInfo
		oldAuthRule, err = impl.ruleBiz.GetPackageRules(id, session, gw.AUTH_RULE)
		if err != nil {
			goto failed
		}
		for _, rule := range oldAuthRule {
			err = impl.ruleBiz.DeleteRule(rule.Id, session)
			if err != nil {
				goto failed
			}
		}
		if dto.AuthType != "" {
			var authRule *gw.OpenapiRule
			authRule, err = impl.createAuthRule(dto.AuthType, dao)
			if err != nil {
				goto failed
			}
			authRule.Region = gw.PACKAGE_RULE
			err = impl.ruleBiz.CreateRule(diceInfo, authRule, session)
			if err != nil {
				goto failed
			}
		}
	}
	if dao.AclType != dto.AclType {
		needUpdate = true
		var oldAclRule []gw.OpenapiRuleInfo
		oldAclRule, err = impl.ruleBiz.GetPackageRules(id, nil, gw.ACL_RULE)
		if err != nil {
			goto failed
		}
		for _, rule := range oldAclRule {
			err = impl.ruleBiz.DeleteRule(rule.Id, session)
			if err != nil {
				goto failed
			}
		}
		if dto.AclType != "" {
			var aclRule *gw.OpenapiRule
			aclRule, err = impl.createAclRule(dto.AclType, id)
			if err != nil {
				goto failed
			}
			aclRule.Region = gw.PACKAGE_RULE
			err = impl.ruleBiz.CreateRule(diceInfo, aclRule, session)
			if err != nil {
				goto failed
			}
		}
	}
	dao.AuthType = dto.AuthType
	dao.AclType = dto.AclType
	dao.Description = dto.Description
	err = impl.packageDb.Update(dao)
	if err != nil {
		goto failed
	}
	if needUpdate {
		// update zone kong polices
		err = impl.ruleBiz.SetPackageKongPolicies(dao, session)
		if err != nil {
			goto failed
		}
	}
	err = session.Commit()
	if err != nil {
		goto failed
	}
	session.Close()
	if needRefreshCloudapi {
		var apis []orm.GatewayPackageApi
		apis, err = impl.packageApiDb.SelectByAny(&orm.GatewayPackageApi{
			PackageId: id,
		})
		if err != nil {
			goto failed
		}
		go func() {
			defer util.DoRecover()
			for _, api := range apis {
				verr := impl.touchCloudapiApi(dao, api)
				if verr != nil {
					log.Errorf("error happened: %+v", verr)
				}
			}
		}()
	}
	return res.SetSuccessAndData(impl.packageDto(dao, dto.BindDomain))
failed:
	log.Errorf("error happened, err:%+v", err)
	if session != nil {
		_ = session.Rollback()
		session.Close()
	}
	return res.SetErrorInfo(&common.ErrInfo{
		Msg: errors.Cause(err).Error(),
	})

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
	if dao.Scene == orm.UNITY_SCENE {
		return errors.New("unity scene can't delete")
	}
	// remove consumer check
	// consumers, err = impl.consumerBiz.GetConsumersOfPackage(id)
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
		res := impl.DeletePackageApi(id, api.Id)
		if !res.Success {
			return errors.Errorf("delete package api failed, err:%+v", res.Err)
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
	err = impl.ruleBiz.DeleteByPackage(dao)
	if err != nil {
		return err
	}
	if dao.ZoneId != "" {
		err = impl.zoneBiz.DeleteZone(dao.ZoneId)
		if err != nil {
			return err
		}
	}
	oldDomains, err = impl.domainBiz.GetPackageDomains(id, session)
	if err != nil {
		return err
	}
	_, err = impl.domainBiz.TouchPackageDomain(id, dao.DiceClusterName, nil, session)
	if err != nil {
		return err
	}
	if dao.CloudapiGroupId != "" {
		orgId, err := impl.globalBiz.GetOrgId(dao.DiceProjectId)
		if err != nil {
			return err
		}
		req := cloudapi.CreateDeleteApiGroupRequest()
		resp := cloudapi.CreateDeleteApiGroupResponse()
		req.GroupId = dao.CloudapiGroupId
		req.SecurityToken = uuid.UUID()
		go func() {
			wg.Wait()
			defer util.DoRecover()
			verr := bundle.Bundle.DoRemoteAliyunAction(orgId, dao.DiceClusterName, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, req, resp)
			if verr != nil {
				log.Errorf("error happened: %+v", verr)
			}
		}()
	}
	err = packSession.Delete(id, true)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) DeletePackage(id string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	var err error
	auditCtx := map[string]interface{}{}
	defer func() {
		if auditCtx["projectId"] == nil || auditCtx["workspace"] == nil {
			return
		}
		audit := common.MakeAuditInfo(impl.ReqCtx, common.ScopeInfo{
			ProjectId: auditCtx["projectId"].(string),
			Workspace: auditCtx["workspace"].(string),
		}, apistructs.DeleteEndpointTemplate, res, auditCtx)
		if audit != nil {
			err = bundle.Bundle.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: *audit})
			if err != nil {
				log.Errorf("create audit failed, err:%+v", err)
			}
		}
	}()
	if id == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	session, _ := db.NewSessionHelper()
	err = impl.deletePackage(id, session, auditCtx)
	if err != nil {
		goto failed
	}
	err = session.Commit()
	if err != nil {
		goto failed
	}
	session.Close()
	return res.SetSuccessAndData(true)
failed:
	log.Errorf("error happened, err:%+v", err)
	if session != nil {
		_ = session.Rollback()
		session.Close()
	}
	return res.SetErrorInfo(&common.ErrInfo{Msg: errors.Cause(err).Error()})
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

func (impl GatewayOpenapiServiceImpl) createKongServiceReq(dto *gw.OpenapiDto, serviceId ...string) *kongDto.KongServiceReqDto {
	i := 0
	reqDto := &kongDto.KongServiceReqDto{
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

func (impl GatewayOpenapiServiceImpl) kongServiceDao(dto *kongDto.KongServiceRespDto, apiId string) *orm.GatewayService {
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

func (impl GatewayOpenapiServiceImpl) touchKongService(adapter kong.KongAdapter, dto *gw.OpenapiDto, apiId string,
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
	var req *kongDto.KongServiceReqDto
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

func (impl GatewayOpenapiServiceImpl) deleteKongService(adapter kong.KongAdapter, apiId string) error {
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

func (impl GatewayOpenapiServiceImpl) createKongRouteReq(dto *gw.OpenapiDto, serviceId string, routeId ...string) *kongDto.KongRouteReqDto {
	reqDto := &kongDto.KongRouteReqDto{}
	if dto.Method != "" {
		reqDto.Methods = []string{dto.Method}
	}
	if dto.ApiPath != "" {
		reqDto.Paths = []string{dto.AdjustPath}
	}
	reqDto.Hosts = dto.Hosts
	reqDto.Service = &kongDto.Service{}
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

func (impl GatewayOpenapiServiceImpl) kongRouteDao(dto *kongDto.KongRouteRespDto, serviceId, apiId string) *orm.GatewayRoute {
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

func (impl GatewayOpenapiServiceImpl) touchKongRoute(adapter kong.KongAdapter, dto *gw.OpenapiDto, apiId string,
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
	var req *kongDto.KongRouteReqDto
	if route == nil {
		req = impl.createKongRouteReq(dto, dto.ServiceId)
	} else {
		req = impl.createKongRouteReq(dto, dto.ServiceId, route.RouteId)
	}
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

func (impl GatewayOpenapiServiceImpl) deleteKongRoute(adapter kong.KongAdapter, apiId string) error {
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

func (impl GatewayOpenapiServiceImpl) deleteKongApi(adapter kong.KongAdapter, apiId string) error {
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

func (impl GatewayOpenapiServiceImpl) createOrUpdatePlugins(adapter kong.KongAdapter, dto *gw.OpenapiDto) error {
	if dto.ServiceRewritePath != "" {
		reqDto := &kongDto.KongPluginReqDto{
			Name: "path-variable",
			// ServiceId: dto.ServiceId,
			RouteId: dto.RouteId,
			Config: map[string]interface{}{
				"request_regex": dto.AdjustPath,
				"rewrite_path":  dto.ServiceRewritePath,
			},
		}
		_, err := adapter.CreateOrUpdatePlugin(reqDto)
		if err != nil {
			return err
		}
	}
	if config.ServerConf.HasRouteInfo {
		reqDto := &kongDto.KongPluginReqDto{
			Name:    "set-route-info",
			RouteId: dto.RouteId,
			Config: map[string]interface{}{
				"project_id": dto.ProjectId,
				"workspace":  strings.ToLower(dto.Env),
				"api_path":   dto.ApiPath,
			},
		}
		_, err := adapter.CreateOrUpdatePlugin(reqDto)
		if err != nil {
			return err
		}
	}
	return nil
}

type cloudapiRequestConfig struct {
	RequestProtocol     string `json:"RequestProtocol"`
	RequestHttpMethod   string `json:"RequestHttpMethod"`
	RequestPath         string `json:"RequestPath"`
	BodyFormat          string `json:"BodyFormat"`
	PostBodyDescription string `json:"PostBodyDescription"`
	RequestMode         string `json:"RequestMode"`
	BodyModel           string `json:"BodyModel"`
}

type cloudapiServiceConfig struct {
	ServiceProtocol       string            `json:"ServiceProtocol"`
	ServiceHttpMethod     string            `json:"ServiceHttpMethod"`
	ServiceAddress        string            `json:"ServiceAddress"`
	ServiceTimeout        requests.Integer  `json:"ServiceTimeout"`
	ServicePath           string            `json:"ServicePath"`
	Mock                  requests.Boolean  `json:"Mock"`
	MockResult            string            `json:"MockResult"`
	ServiceVpcEnable      requests.Boolean  `json:"ServiceVpcEnable"`
	VpcConfig             cloudapiVpcConfig `json:"VpcConfig"`
	FunctionComputeConfig interface{}       `json:"FunctionComputeConfig"`
	ContentTypeCatagory   string            `json:"ContentTypeCatagory"`
	ContentTypeValue      string            `json:"ContentTypeValue"`
	InnerServiceType      string            `json:"InnerServiceType"`
	AoneAppName           string            `json:"AoneAppName"`
}

type cloudapiVpcConfig struct {
	VpcScheme string `json:"VpcScheme"`
	Name      string `json:"Name"`
}

func (impl GatewayOpenapiServiceImpl) makeCloudapiCreateApiRequest(pack *orm.GatewayPackage, api orm.GatewayPackageApi, authType, codeAuthType string) (*cloudapi.CreateApiRequest, error) {
	apiName := fmt.Sprintf("API_%s", api.Id)
	req := cloudapi.CreateCreateApiRequest()
	req.Visibility = "PRIVATE"
	req.GroupId = pack.CloudapiGroupId
	req.ApiName = apiName
	req.ForceNonceCheck = requests.NewBoolean(false)
	req.DisableInternet = requests.NewBoolean(false)
	req.AuthType = authType
	req.AppCodeAuthType = codeAuthType
	req.ResultType = "PASSTHROUGH"
	req.AllowSignatureMethod = "HmacSHA256"
	req.WebSocketApiType = "COMMON"
	req.SecurityToken = uuid.UUID()
	var requestConfig cloudapiRequestConfig
	requestConfig.RequestProtocol = "HTTP,HTTPS"
	if api.Method == "" {
		requestConfig.RequestHttpMethod = "ANY"
	} else {
		requestConfig.RequestHttpMethod = api.Method
	}
	requestConfig.RequestPath = fmt.Sprintf("%s/*", strings.TrimSuffix(api.ApiPath, "/"))
	requestConfig.RequestMode = "PASSTHROUGH"
	configStr, err := json.Marshal(requestConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.RequestConfig = string(configStr)
	var serviceConfig cloudapiServiceConfig
	var vpcConfig cloudapiVpcConfig
	vpcConfig.Name = pack.CloudapiVpcGrant
	serviceConfig.VpcConfig = vpcConfig
	serviceConfig.ServiceProtocol = "HTTP"
	serviceConfig.ServiceHttpMethod = requestConfig.RequestHttpMethod
	serviceConfig.ServiceTimeout = requests.NewInteger(10000)
	serviceConfig.ServicePath = requestConfig.RequestPath
	serviceConfig.Mock = requests.NewBoolean(false)
	serviceConfig.ServiceVpcEnable = requests.NewBoolean(true)
	// if outter user, this is an empty string
	serviceConfig.AoneAppName = config.ServerConf.AoneAppName
	configStr, err = json.Marshal(serviceConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.ServiceConfig = string(configStr)
	return req, nil
}

func (impl GatewayOpenapiServiceImpl) makeCloudapiModifyApiRequest(pack *orm.GatewayPackage, api orm.GatewayPackageApi, authType, codeAuthType string, desc *cloudapi.DescribeApiResponse) (*cloudapi.ModifyApiRequest, error) {
	req := cloudapi.CreateModifyApiRequest()
	req.SecurityToken = uuid.UUID()
	req.WebSocketApiType = desc.WebSocketApiType
	var err error
	var b []byte
	if len(desc.ErrorCodeSamples.ErrorCodeSample) > 0 {
		b, err = json.Marshal(desc.ErrorCodeSamples)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		req.ErrorCodeSamples = string(b)
	}
	req.AppCodeAuthType = codeAuthType
	req.AuthType = authType
	req.Description = desc.Description
	req.DisableInternet = requests.NewBoolean(desc.DisableInternet)
	if len(desc.ConstantParameters.ConstantParameter) > 0 {
		b, err = json.Marshal(desc.ConstantParameters)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		req.ConstantParameters = string(b)
	}
	if len(desc.ServiceParameters.ServiceParameter) > 0 {
		req.AllowSignatureMethod = desc.AllowSignatureMethod
		b, err = json.Marshal(desc.ServiceParameters)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		req.ServiceParameters = string(b)
	}
	req.FailResultSample = desc.FailResultSample
	if len(desc.SystemParameters.SystemParameter) > 0 {
		b, err = json.Marshal(desc.SystemParameters)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		req.SystemParameters = string(b)
	}
	if len(desc.ServiceParametersMap.ServiceParameterMap) > 0 {
		b, err = json.Marshal(desc.ServiceParametersMap)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		req.ServiceParametersMap = string(b)
	}
	b, err = json.Marshal(desc.OpenIdConnectConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.OpenIdConnectConfig = string(b)
	if len(desc.RequestParameters.RequestParameter) > 0 {
		b, err = json.Marshal(desc.RequestParameters)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		req.RequestParameters = string(b)
	}
	if len(desc.ResultDescriptions.ResultDescription) > 0 {
		b, err = json.Marshal(desc.ResultDescriptions)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		req.ResultDescriptions = string(b)
	}
	req.Visibility = desc.Visibility
	req.GroupId = desc.GroupId
	req.ResultType = desc.ResultType
	req.ApiName = desc.ApiName
	req.ResultSample = desc.ResultSample
	req.ForceNonceCheck = requests.NewBoolean(desc.ForceNonceCheck)
	desc.RequestConfig.RequestPath = fmt.Sprintf("%s/*", strings.TrimSuffix(api.ApiPath, "/"))
	if api.Method == "" {
		desc.RequestConfig.RequestHttpMethod = "ANY"
	} else {
		desc.RequestConfig.RequestHttpMethod = api.Method
	}
	b, err = json.Marshal(desc.RequestConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.RequestConfig = string(b)
	desc.ServiceConfig.ServiceHttpMethod = desc.RequestConfig.RequestHttpMethod
	desc.ServiceConfig.ServicePath = desc.RequestConfig.RequestPath
	b, err = json.Marshal(desc.ServiceConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.ServiceConfig = string(b)
	req.ResultBodyModel = desc.ResultBodyModel
	req.ApiId = api.CloudapiApiId
	return req, nil
}

func (impl GatewayOpenapiServiceImpl) touchCloudapiApi(pack *orm.GatewayPackage, api orm.GatewayPackageApi) error {
	if pack.CloudapiGroupId == "" {
		return nil
	}
	// TODO: support path variable
	if strings.Contains(api.ApiPath, "{") {
		return nil
	}
	orgId, err := impl.globalBiz.GetOrgId(pack.DiceProjectId)
	if err != nil {
		return err
	}
	cloudapiAuthType := "ANONYMOUS"
	cloudapiAppCodeAuthType := ""
	if pack.AuthType == gw.AT_ALIYUN_APP && api.AclType != gw.ACL_OFF {
		cloudapiAuthType = "APP"
		cloudapiAppCodeAuthType = "DEFAULT"
	}
	if api.CloudapiApiId != "" {
		req := cloudapi.CreateDescribeApiRequest()
		req.GroupId = pack.CloudapiGroupId
		req.ApiId = api.CloudapiApiId
		resp := cloudapi.CreateDescribeApiResponse()
		err = bundle.Bundle.DoRemoteAliyunAction(orgId, pack.DiceClusterName, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, req, resp)
		if err != nil {
			rootErr := errors.Cause(err)
			if serverErr, ok := rootErr.(*aliyun_errors.ServerError); ok {
				if serverErr.HttpStatus() == 404 {
					goto new_api
				}
			}
		}
		if err != nil {
			return err
		}
		{
			modifyReq, err := impl.makeCloudapiModifyApiRequest(pack, api, cloudapiAuthType, cloudapiAppCodeAuthType, resp)
			if err != nil {
				return err
			}
			modifyResp := cloudapi.CreateModifyApiResponse()
			err = bundle.Bundle.DoRemoteAliyunAction(orgId, pack.DiceClusterName, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, modifyReq, modifyResp)
			if err != nil {
				return err
			}
			goto done
		}
	}

new_api:
	{
		req, err := impl.makeCloudapiCreateApiRequest(pack, api, cloudapiAuthType, cloudapiAppCodeAuthType)
		if err != nil {
			return err
		}
		resp := cloudapi.CreateCreateApiResponse()
		err = bundle.Bundle.DoRemoteAliyunAction(orgId, pack.DiceClusterName, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, req, resp)
		if err != nil {
			return err
		}
		api.CloudapiApiId = resp.ApiId
	}
done:
	req := cloudapi.CreateDeployApiRequest()
	req.StageName = "RELEASE"
	req.GroupId = pack.CloudapiGroupId
	req.Description = "new deploy"
	req.SecurityToken = uuid.UUID()
	req.ApiId = api.CloudapiApiId
	resp := cloudapi.CreateDeployApiResponse()
	err = bundle.Bundle.DoRemoteAliyunAction(orgId, pack.DiceClusterName, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, req, resp)
	if err != nil {
		return err
	}
	_ = impl.packageApiDb.Update(&api, "cloudapi_api_id")
	return nil
}

func (impl GatewayOpenapiServiceImpl) SessionCreatePackageApi(id string, dto *gw.OpenapiDto, session *db.SessionHelper, injectRuntimeDomain bool) (bool, string, *apistructs.Audit, error) {
	exist := false
	var pack *orm.GatewayPackage
	var kongInfo *orm.GatewayKongInfo
	var kongAdapter kong.KongAdapter
	var dao *orm.GatewayPackageApi
	var unique bool
	var serviceId, routeId string
	var apiSession db.GatewayPackageApiService
	var packSession db.GatewayPackageService
	var err error
	var zoneId string
	var domains []string
	var audit *apistructs.Audit
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
	audit = common.MakeAuditInfo(impl.ReqCtx, common.ScopeInfo{
		ProjectId: pack.DiceProjectId,
		Workspace: pack.DiceEnv,
	}, apistructs.CreateRouteTemplate, nil, map[string]interface{}{
		"endpoint": pack.PackageName,
		"method":   method,
		"path":     dto.ApiPath,
	})
	dto.ProjectId = pack.DiceProjectId
	dto.Env = pack.DiceEnv
	if pack.Scene == orm.UNITY_SCENE {
		defaultPath, err := impl.globalBiz.GenerateDefaultPath(pack.DiceProjectId)
		if err != nil {
			goto failed
		}
		if !strings.HasPrefix(dto.ApiPath, defaultPath) {
			dto.ApiPath = defaultPath + dto.ApiPath
		}
	}
	// get package domain
	domains, err = impl.domainBiz.GetPackageDomains(id, session)
	if err != nil {
		goto failed
	}
	if len(domains) == 0 {
		err = errors.New("you need set the endpoint's domain first")
		goto failed
	}
	dto.Hosts = domains
	log.Debugf("api dto before adjust %+v", dto) // output for debug
	err = dto.Adjust()
	if err != nil {
		goto failed
	}
	log.Debugf("api dto after adjust %+v", dto) // output for debug
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
		kongAdapter = kong.NewKongAdapter(kongInfo.KongAddr)
		serviceId, err = impl.touchKongService(kongAdapter, dto, dao.Id, session)
		if err != nil {
			goto failed
		}
		dto.ServiceId = serviceId
		routeId, err = impl.touchKongRoute(kongAdapter, dto, dao.Id, session)
		if err != nil {
			goto failed
		}
		dto.RouteId = routeId
		err = impl.createOrUpdatePlugins(kongAdapter, dto)
		if err != nil {
			goto failed
		}
	} else if dto.RedirectType == gw.RT_SERVICE {
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
		dao.RuntimeServiceId = runtimeService.Id
		err = apiSession.Insert(dao)
		if err != nil {
			goto failed
		}
	}
	zoneId, err = impl.TouchPackageApiZone(PackageApiInfo{dao, domains, pack.DiceProjectId, pack.DiceEnv, pack.DiceClusterName, injectRuntimeDomain}, session)
	if err != nil {
		goto failed
	}
	if dao.ZoneId == "" {
		dao.ZoneId = zoneId
		err = apiSession.Update(dao)
		if err != nil {
			goto failed
		}
	}
	err = apiSession.Update(dao)
	if err != nil {
		goto failed
	}
	if dto.AllowPassAuth {
		var authRule, aclRule *gw.OpenapiRule
		diceInfo := DiceInfo{
			ProjectId: pack.DiceProjectId,
			Env:       pack.DiceEnv,
			Az:        pack.DiceClusterName,
		}
		aclRule, err = impl.createApiAclRule(gw.ACL_OFF, dao.PackageId, dao.Id)
		if err != nil {
			goto failed
		}
		authRule, err = impl.createApiAuthRule(dao.PackageId, dao.Id, false)
		if err != nil {
			goto failed
		}
		err = impl.ruleBiz.CreateRule(diceInfo, aclRule, session)
		if err != nil {
			goto failed
		}
		err = impl.ruleBiz.CreateRule(diceInfo, authRule, session)
		if err != nil {
			goto failed
		}
		needUpdateDomainPolicy = true
	}
	if needUpdateDomainPolicy {
		err = impl.ruleBiz.SetPackageKongPolicies(pack, session)
		if err != nil {
			goto failed
		}
	}
	go func() {
		defer util.DoRecover()
		verr := impl.touchCloudapiApi(pack, *dao)
		if verr != nil {
			log.Errorf("error happened: %+v", verr)
		}
	}()
	return false, dao.Id, audit, nil
failed:
	if serviceId != "" {
		_ = kongAdapter.DeleteService(serviceId)
	}
	if routeId != "" {
		_ = kongAdapter.DeleteRoute(routeId)
	}
	if zoneId != "" {
		zerr := impl.zoneBiz.DeleteZoneRoute(zoneId, session)
		if zerr != nil {
			log.Errorf("delete zone route failed, err:%+v", zerr)
		}
	}
	return exist, "", audit, err
}

func (impl GatewayOpenapiServiceImpl) TouchPackageApiZone(info PackageApiInfo, session ...*db.SessionHelper) (string, error) {
	var runtimeSession db.GatewayRuntimeServiceService
	var err error
	if len(session) > 0 {
		runtimeSession, err = impl.runtimeDb.NewSession(session[0])
		if err != nil {
			return "", err
		}
	} else {
		runtimeSession = impl.runtimeDb
	}
	az, err := impl.azDb.GetAzInfoByClusterName(info.Az)
	if err != nil {
		return "", err
	}
	// set route config
	routeConfig := RouteConfig{
		Hosts: info.Hosts,
		Path:  info.ApiPath,
	}
	routeConfig.InjectRuntimeDomain = info.InjectRuntimeDomain
	if info.RedirectType == gw.RT_SERVICE {
		runtimeService, err := runtimeSession.Get(info.RuntimeServiceId)
		if err != nil {
			return "", err
		}
		if runtimeService == nil {
			return "", errors.Errorf("runtime service not found, err:%+v", err)
		}
		if runtimeService.IsSecurity == 1 {
			routeConfig.BackendProtocol = &[]k8s.BackendProtocl{k8s.HTTPS}[0]
		}
		_, innerHost, err := impl.globalBiz.GenerateEndpoint(DiceInfo{
			Az:        info.Az,
			ProjectId: info.ProjectId,
			Env:       info.Env,
		})
		if err != nil {
			return "", err
		}
		routeConfig.UseRegex = true
		routeConfig.RewriteHost = &innerHost
		rewritePath, err := impl.globalBiz.GetRuntimeServicePrefix(runtimeService)
		if err != nil {
			return "", err
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
	} else {
		routeConfig.BackendProtocol = &[]k8s.BackendProtocl{k8s.HTTPS}[0]
		routeConfig.UseRegex = true
		routeConfig.Path += ".*"
		routeConfig.Annotations = map[string]*string{}
		routeConfig.Annotations[k8s.REWRITE_HOST_KEY] = nil
		routeConfig.Annotations[k8s.REWRITE_PATH_KEY] = nil
	}
	if info.ZoneId == "" {
		//create zone
		zone, err := impl.zoneBiz.CreateZoneWithoutIngress(ZoneConfig{
			ZoneRoute: &ZoneRoute{routeConfig},
			Name:      "api-" + info.Id,
			ProjectId: info.ProjectId,
			Env:       info.Env,
			Az:        info.Az,
			Type:      db.ZONE_TYPE_PACKAGE_API,
		}, session...)
		if err != nil {
			return "", err
		}
		transSucc := false
		annotations, locationSnippet, retSession, err := impl.policyBiz.SetZoneDefaultPolicyConfig(info.PackageId, zone, az, session...)
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
			return "", err
		}
		routeConfig.Annotations = annotations
		routeConfig.LocationSnippet = locationSnippet
		_, err = impl.zoneBiz.UpdateZoneRoute(zone.Id, ZoneRoute{routeConfig}, session...)
		if err != nil {
			return "", err
		}
		transSucc = true
		return zone.Id, nil
	}
	//update zone route
	exist, err := impl.zoneBiz.UpdateZoneRoute(info.ZoneId, ZoneRoute{routeConfig})
	if err != nil {
		return "", err
	}
	if !exist {
		zone, err := impl.zoneBiz.GetZone(info.ZoneId, session...)
		if err != nil {
			return "", err
		}
		err = impl.policyBiz.RefreshZoneIngress(*zone, *az)
		if err != nil {
			return "", err
		}
	}
	return info.ZoneId, nil
}

func (impl GatewayOpenapiServiceImpl) CreatePackageApi(id string, dto *gw.OpenapiDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if id == "" || (dto.ApiPath == "" && dto.Method == "") {
		log.Errorf("invalid argument, %+v", dto)
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	var exist bool
	var err error
	var audit *apistructs.Audit
	var apiId string
	defer func() {
		if audit != nil {
			if res.Success {
				audit.Result = apistructs.SuccessfulResult
			} else {
				//audit.Result = apistructs.FailureResult
				return
			}
			if res.Err != nil {
				audit.ErrorMsg = res.Err.Msg
			}
			err = bundle.Bundle.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: *audit})
			if err != nil {
				log.Errorf("create audit failed, err:%+v", err)
			}
		}
	}()
	helper, err := db.NewSessionHelper()
	if err != nil {
		goto failed
	}
	if exist, apiId, audit, err = impl.SessionCreatePackageApi(id, dto, helper, false); err != nil {
		if exist {
			res.SetReturnCode(API_EXIST)
			goto failed
		}
		res.SetErrorInfo(&common.ErrInfo{
			Msg: errors.Cause(err).Error(),
		})
		goto failed
	}
	_ = helper.Commit()
	helper.Close()
	return res.SetSuccessAndData(apiId)
failed:
	if helper != nil {
		_ = helper.Rollback()
		helper.Close()
	}
	log.Errorf("errorHappened, err:%+v", err)
	return res
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
			log.Errorf("get runtimeservice failed, err:%+v", err)
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
			log.Errorf("invalide RedirectAddr %s", dto.RedirectAddr)
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
	rules, err := impl.ruleBiz.GetApiRules(apiId, gw.AUTH_RULE)
	if err != nil {
		return err
	}
	if len(rules) > 0 {
		err = impl.ruleBiz.DeleteRule(rules[0].Id, nil)
		if err != nil {

			return err
		}
	}
	rules, err = impl.ruleBiz.GetApiRules(apiId, gw.ACL_RULE)
	if err != nil {
		return err
	}
	if len(rules) > 0 {
		err = impl.ruleBiz.DeleteRule(rules[0].Id, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) createApiPassAuthRules(pack *orm.GatewayPackage, apiId string) error {
	var authRule, aclRule *gw.OpenapiRule
	diceInfo := DiceInfo{
		ProjectId: pack.DiceProjectId,
		Env:       pack.DiceEnv,
		Az:        pack.DiceClusterName,
	}
	aclRule, err := impl.createApiAclRule(gw.ACL_OFF, pack.Id, apiId)
	if err != nil {
		return err
	}
	authRule, err = impl.createApiAuthRule(pack.Id, apiId, false)
	if err != nil {
		return err
	}
	err = impl.ruleBiz.CreateRule(diceInfo, aclRule, nil)
	if err != nil {
		return err
	}
	err = impl.ruleBiz.CreateRule(diceInfo, authRule, nil)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) GetPackageApis(id string, args *gw.GetOpenapiDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
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
	page, err := impl.packageApiDb.GetPage(options, pageInfo)
	if err != nil {
		goto failed
	}
	daos, ok = page.Result.([]orm.GatewayPackageApi)
	if !ok {
		err = errors.New("type convert failed")
		goto failed
	}
	for _, dao := range daos {
		dto := impl.openapiDto(&dao)
		list = append(list, *dto)
	}
	return res.SetSuccessAndData(common.NewPages(list, pageInfo.TotalNum))
failed:
	log.Errorf("error happened, err:%+v", err)
	return res
}

func (impl GatewayOpenapiServiceImpl) UpdatePackageApi(packageId, apiId string, dto *gw.OpenapiDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if packageId == "" || apiId == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	var dao, updateDao *orm.GatewayPackageApi
	var pack *orm.GatewayPackage
	var kongInfo *orm.GatewayKongInfo
	var kongAdapter kong.KongAdapter
	var unique bool
	var serviceId, routeId string
	var err error
	var domains []string
	needUpdateDomainPolicy := false
	auditCtx := map[string]interface{}{}
	defer func() {
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
		audit := common.MakeAuditInfo(impl.ReqCtx, common.ScopeInfo{
			ProjectId: pack.DiceProjectId,
			Workspace: pack.DiceEnv,
		}, apistructs.UpdateRouteTemplate, res, auditCtx)
		if audit != nil {
			err = bundle.Bundle.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: *audit})
			if err != nil {
				log.Errorf("create audit failed, err:%+v", err)
			}
		}
	}()
	if valid, msg := dto.CheckValid(); !valid {
		res.SetErrorInfo(&common.ErrInfo{
			Code: "格式错误",
			Msg:  msg,
		})
		goto failed
	}
	// get package domain
	domains, err = impl.domainBiz.GetPackageDomains(packageId)
	if err != nil {
		goto failed
	}
	dto.Hosts = domains
	err = dto.Adjust()
	if err != nil {
		goto failed
	}
	dao, err = impl.packageApiDb.Get(apiId)
	if err != nil {
		goto failed
	}
	if dao == nil {
		err = errors.New("api not exist")
		goto failed
	}
	pack, err = impl.packageDb.Get(packageId)
	if err != nil {
		goto failed
	}
	if pack == nil {
		goto failed
	}
	auditCtx["endpoint"] = pack.PackageName
	if pack.Scene == orm.UNITY_SCENE {
		var defaultPath string
		defaultPath, err = impl.globalBiz.GenerateDefaultPath(pack.DiceProjectId)
		if err != nil {
			goto failed
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
		goto failed
	}
	kongAdapter = kong.NewKongAdapter(kongInfo.KongAddr)
	updateDao = impl.packageApiDao(dto)
	updateDao.ZoneId = dao.ZoneId
	updateDao.CloudapiApiId = dao.CloudapiApiId
	//mutable check
	if dao.Origin == string(gw.FROM_DICE) || dao.Origin == string(gw.FROM_SHADOW) {
		if dto.RedirectAddr != dao.RedirectAddr {
			err = errors.Errorf("redirectaddr not same, old:%s, new:%s", dao.RedirectAddr,
				dto.RedirectAddr)
			res.SetReturnCode(DICE_API_NOT_MUTABLE)
			goto failed
		}
		if dto.Method != dao.Method {
			err = errors.Errorf("method not same, old:%s, new:%s", dao.Method,
				dto.Method)
			res.SetReturnCode(DICE_API_NOT_MUTABLE)
			goto failed
		}
	} else if dao.Origin == string(gw.FROM_CUSTOM) || dao.Origin == string(gw.FROM_DICEYML) {
		var zoneId string
		updateDao.PackageId = packageId
		updateDao.Id = apiId
		unique, err = impl.packageApiDb.CheckUnique(updateDao)
		if err != nil {
			goto failed
		}
		if !unique {
			err = errors.Errorf("package api already exist, dao:%+v", updateDao)
			res.SetReturnCode(API_EXIST)
			goto failed
		}
		dto.ProjectId = pack.DiceProjectId
		dto.Env = pack.DiceEnv
		if updateDao.RedirectType == gw.RT_URL {
			updateDao.RuntimeServiceId = ""
			updateDao.DiceApp = ""
			updateDao.DiceService = ""
			if dto.RedirectAddr != dao.RedirectAddr || dto.ApiPath != dao.ApiPath || dto.Method != dao.Method {
				serviceId, err = impl.touchKongService(kongAdapter, dto, apiId)
				if err != nil {
					goto failed
				}
				dto.ServiceId = serviceId
				routeId, err = impl.touchKongRoute(kongAdapter, dto, apiId)
				if err != nil {
					goto failed
				}
				dto.RouteId = routeId
			} else {
				var route *orm.GatewayRoute
				route, err = impl.routeDb.GetByApiId(apiId)
				if err != nil {
					goto failed
				}
				if route == nil {
					err = errors.Errorf("route of api:%s not exist", apiId)
					goto failed
				}
				dto.RouteId = route.RouteId
			}
			err = impl.createOrUpdatePlugins(kongAdapter, dto)
			if err != nil {
				goto failed
			}
		} else if updateDao.RedirectType == gw.RT_SERVICE {
			var runtimeService *orm.GatewayRuntimeService
			if dao.RedirectType == gw.RT_URL {
				err = impl.deleteKongApi(kongAdapter, apiId)
				if err != nil {
					goto failed
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
				goto failed
			}
			if runtimeService == nil {
				err = errors.New("find runtime service failed")
				goto failed
			}
			updateDao.RuntimeServiceId = runtimeService.Id
		}
		zoneId, err = impl.TouchPackageApiZone(PackageApiInfo{updateDao, domains, pack.DiceProjectId, pack.DiceEnv, pack.DiceClusterName, false})
		if err != nil {
			goto failed
		}
		updateDao.ZoneId = zoneId
	}
	err = impl.packageApiDb.Update(updateDao)
	if err != nil {
		goto failed
	}
	if dao.AclType != updateDao.AclType {
		needUpdateDomainPolicy = true
		if updateDao.AclType == gw.ACL_NONE {
			err = impl.deleteApiPassAuthRules(apiId)
			if err != nil {
				goto failed
			}
		} else if updateDao.AclType == gw.ACL_OFF {
			err = impl.createApiPassAuthRules(pack, apiId)
			if err != nil {
				goto failed
			}
		}
	} else if dao.RedirectType != updateDao.RedirectType {
		needUpdateDomainPolicy = true
		if updateDao.AclType == gw.ACL_OFF {
			err = impl.deleteApiPassAuthRules(apiId)
			if err != nil {
				goto failed
			}
			err = impl.createApiPassAuthRules(pack, apiId)
			if err != nil {
				goto failed
			}
		}
	}
	if needUpdateDomainPolicy {
		err = impl.ruleBiz.SetPackageKongPolicies(pack, nil)
		if err != nil {
			goto failed
		}
	}
	go func() {
		defer util.DoRecover()
		verr := impl.touchCloudapiApi(pack, *updateDao)
		if verr != nil {
			log.Errorf("error happened: %+v", verr)
		}
	}()
	return res.SetSuccessAndData(impl.openapiDto(dao))
failed:
	log.Errorf("error happened, err:%+v", err)
	return res
}

func (impl *GatewayOpenapiServiceImpl) DeletePackageApi(packageId, apiId string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	var pack *orm.GatewayPackage
	var kongInfo *orm.GatewayKongInfo
	var kongAdapter kong.KongAdapter
	auditCtx := map[string]interface{}{}
	dao, err := impl.packageApiDb.Get(apiId)
	if err != nil {
		goto failed
	}
	if dao == nil {
		return res.SetSuccessAndData(true)
	}
	defer func() {
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
		audit := common.MakeAuditInfo(impl.ReqCtx, common.ScopeInfo{
			ProjectId: pack.DiceProjectId,
			Workspace: pack.DiceEnv,
		}, apistructs.DeleteRouteTemplate, res, auditCtx)
		if audit != nil {
			err = bundle.Bundle.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: *audit})
			if err != nil {
				log.Errorf("create audit failed, err:%+v", err)
			}
		}
	}()

	pack, err = impl.packageDb.Get(packageId)
	if err != nil {
		goto failed
	}
	kongInfo, err = impl.kongDb.GetKongInfo(&orm.GatewayKongInfo{
		Az:        pack.DiceClusterName,
		ProjectId: pack.DiceProjectId,
		Env:       pack.DiceEnv,
	})
	if err != nil {
		goto failed
	}
	kongAdapter = kong.NewKongAdapter(kongInfo.KongAddr)
	err = impl.ruleBiz.DeleteByPackageApi(pack, dao)
	if err != nil {
		goto failed
	}
	if dao.Origin == string(gw.FROM_DICE) || dao.Origin == string(gw.FROM_SHADOW) {
		diceApiId := dao.DiceApiId
		var api *orm.GatewayApi
		var upApi *orm.GatewayUpstreamApi
		if diceApiId == "" {
			err = errors.Errorf("dice api id is empty, package api id:%s", apiId)
			goto failed
		}
		err = impl.apiInPackageDb.Delete(packageId, diceApiId)
		if err != nil {
			goto failed
		}
		api, err = impl.apiDb.GetById(diceApiId)
		if err != nil {
			goto failed
		}
		if api != nil {
			if api.UpstreamApiId != "" {
				upApi, err = impl.upstreamApiDb.GetById(api.UpstreamApiId)
				if err != nil {
					goto failed
				}
				if upApi != nil {
					err = impl.apiBiz.DeleteUpstreamBindApi(upApi)
					if err != nil {
						goto failed
					}
					err = impl.upstreamApiDb.DeleteById(api.UpstreamApiId)
					if err != nil {
						goto failed
					}
				}
			} else {
				impl.apiBiz.DeleteApi(api.Id)
			}
		}

	} else if dao.Origin == string(gw.FROM_CUSTOM) || dao.Origin == string(gw.FROM_DICEYML) {
		err = impl.deleteKongApi(kongAdapter, apiId)
		if err != nil {
			goto failed
		}
	}
	if dao.ZoneId != "" {
		err = impl.zoneBiz.DeleteZone(dao.ZoneId)
		if err != nil {
			goto failed
		}
	}
	if dao.CloudapiApiId != "" {
		orgId, err := impl.globalBiz.GetOrgId(pack.DiceProjectId)
		if err != nil {
			goto failed
		}
		abolishReq := cloudapi.CreateAbolishApiRequest()
		abolishResp := cloudapi.CreateAbolishApiResponse()
		abolishReq.GroupId = pack.CloudapiGroupId
		abolishReq.ApiId = dao.CloudapiApiId
		abolishReq.StageName = "RELEASE"
		req := cloudapi.CreateDeleteApiRequest()
		resp := cloudapi.CreateDeleteApiResponse()
		req.GroupId = pack.CloudapiGroupId
		req.ApiId = dao.CloudapiApiId
		req.SecurityToken = uuid.UUID()
		go func() {
			wgi := impl.ctx.Value(ctxWg{})
			if wgi == nil {
				return
			}
			wg := wgi.(*sync.WaitGroup)
			defer func() {
				wg.Done()
				util.DoRecover()
			}()
			verr := bundle.Bundle.DoRemoteAliyunAction(orgId, pack.DiceClusterName, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, abolishReq, abolishResp)
			if verr != nil {
				log.Errorf("error happened: %+v", verr)
			}
			verr = bundle.Bundle.DoRemoteAliyunAction(orgId, pack.DiceClusterName, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, req, resp)
			if verr != nil {
				log.Errorf("error happened: %+v", verr)
			}
		}()
	}
	err = impl.packageApiDb.Delete(apiId)
	if err != nil {
		goto failed
	}
	return res.SetSuccessAndData(true)
failed:
	log.Errorf("error happened, err:%+v", err)
	return res.SetErrorInfo(&common.ErrInfo{Msg: errors.Cause(err).Error()})
}

func (impl GatewayOpenapiServiceImpl) TouchRuntimePackageMeta(endpoint *orm.GatewayRuntimeService, session *db.SessionHelper) (string, bool, error) {
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
	zone, err := impl.zoneBiz.CreateZone(ZoneConfig{
		Name:      regexp.MustCompile(`([^0-9a-zA-z]+|_+|\\+)`).ReplaceAllString("package-"+name, "-"),
		ProjectId: endpoint.ProjectId,
		Env:       endpoint.Workspace,
		Az:        endpoint.ClusterName,
		Type:      db.ZONE_TYPE_PACKAGE_NEW,
	}, session)
	if err != nil {
		return "", false, err
	}
	pack := &orm.GatewayPackage{
		DiceProjectId:    endpoint.ProjectId,
		DiceEnv:          endpoint.Workspace,
		DiceClusterName:  endpoint.ClusterName,
		PackageName:      name,
		Scene:            orm.WEBAPI_SCENE,
		ZoneId:           zone.Id,
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
	domains, err := impl.domainBiz.GetPackageDomains(packageId, session)
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
			defer doRecover()
			zoneId, aerr := impl.TouchPackageApiZone(PackageApiInfo{&apiRefer, domains, dao.DiceProjectId, dao.DiceEnv, dao.DiceClusterName, false})
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

	err = impl.ruleBiz.SetPackageKongPolicies(dao, session)
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
	diceInfo := DiceInfo{
		ProjectId: pack.DiceProjectId,
		Az:        pack.DiceClusterName,
		Env:       pack.DiceEnv,
	}
	outerHost, rewriteHost, err := impl.globalBiz.GenerateEndpoint(diceInfo)
	if err != nil {
		return nil, err
	}
	innerHost := impl.globalBiz.GetServiceAddr(diceInfo.Env)
	path, err := impl.globalBiz.GenerateDefaultPath(diceInfo.ProjectId)
	if err != nil {
		return nil, err
	}
	route := RouteConfig{
		Hosts: []string{outerHost, innerHost},
		Path:  path + "/.*",
		RouteOptions: k8s.RouteOptions{
			BackendProtocol: &[]k8s.BackendProtocl{k8s.HTTPS}[0],
		},
	}
	route.RewriteHost = &rewriteHost
	route.UseRegex = true
	zone, err := impl.zoneBiz.CreateZone(ZoneConfig{
		ZoneRoute: &ZoneRoute{Route: route},
		Name:      "unity",
		ProjectId: diceInfo.ProjectId,
		Env:       diceInfo.Env,
		Az:        diceInfo.Az,
		Type:      db.ZONE_TYPE_UNITY,
	}, session)
	if err != nil {
		return nil, err
	}
	pack.ZoneId = zone.Id
	err = packSession.Update(pack)
	if err != nil {
		return nil, err
	}
	return zone, nil
}

func (impl GatewayOpenapiServiceImpl) CreateTenantPackage(tenantId string, session *db.SessionHelper) error {
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
	diceInfo := DiceInfo{
		ProjectId: kongInfo.ProjectId,
		Az:        kongInfo.Az,
		Env:       kongInfo.Env,
	}
	outerHost, rewriteHost, err := impl.globalBiz.GenerateEndpoint(diceInfo, session)
	if err != nil {
		return err
	}
	innerHost := impl.globalBiz.GetServiceAddr(kongInfo.Env)
	path, err := impl.globalBiz.GenerateDefaultPath(kongInfo.ProjectId, session)
	if err != nil {
		return err
	}
	route := RouteConfig{
		Hosts: []string{outerHost, innerHost},
		Path:  path + "/.*",
		RouteOptions: k8s.RouteOptions{
			BackendProtocol: &[]k8s.BackendProtocl{k8s.HTTPS}[0],
		},
	}
	route.RewriteHost = &rewriteHost
	route.UseRegex = true
	zone, err := impl.zoneBiz.CreateZone(ZoneConfig{
		ZoneRoute: &ZoneRoute{Route: route},
		Name:      "unity",
		ProjectId: kongInfo.ProjectId,
		Env:       kongInfo.Env,
		Az:        kongInfo.Az,
		Type:      db.ZONE_TYPE_UNITY,
	}, session)
	if err != nil {
		return err
	}
	var pack *orm.GatewayPackage
	packageSession, err := impl.packageDb.NewSession(session)
	if err != nil {
		goto clear_route
	}
	pack = &orm.GatewayPackage{
		DiceProjectId:   kongInfo.ProjectId,
		DiceEnv:         kongInfo.Env,
		DiceClusterName: kongInfo.Az,
		PackageName:     "unity",
		Scene:           orm.UNITY_SCENE,
		ZoneId:          zone.Id,
	}
	err = packageSession.Insert(pack)
	if err != nil {
		goto clear_route
	}
	_, err = impl.domainBiz.TouchPackageDomain(pack.Id, kongInfo.Az, []string{outerHost}, session)
	if err != nil {
		goto clear_route
	}
	_, _, err = impl.policyBiz.SetZonePolicyConfig(zone, "built-in", nil, session)
	if err != nil {
		goto clear_route
	}
	{
		var policyEngine apipolicy.PolicyEngine
		var corsConfig apipolicy.PolicyDto
		var configByte []byte
		var azInfo *orm.GatewayAzInfo
		azInfo, err = impl.azDb.GetAzInfoByClusterName(kongInfo.Az)
		if err != nil {
			goto clear_route
		}
		if azInfo == nil {
			err = errors.New("az not found")
			goto clear_route
		}
		policyEngine, err = apipolicy.GetPolicyEngine("cors")
		if err != nil {
			goto clear_route
		}
		corsConfig = policyEngine.CreateDefaultConfig(nil)
		corsConfig.SetEnable(true)
		configByte, err = json.Marshal(corsConfig)
		if err != nil {
			err = errors.WithStack(err)
			goto clear_route
		}
		//TODO
		_ = session.Commit()
		_, err = impl.policyBiz.SetPackageDefaultPolicyConfig("cors", pack.Id, azInfo, configByte)
		if err != nil {
			goto clear_route
		}
	}
	return nil
clear_route:
	delErr := impl.zoneBiz.DeleteZoneRoute(zone.Id, session)
	if delErr != nil {
		log.Errorf("delete zone failed, err:%+v", delErr)
	}
	return err
}

// 获取绑定的阿里云API网关分组域名
func (impl GatewayOpenapiServiceImpl) GetCloudapiGroupBind(packageId string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	bindInfo := gw.CloudapiBindInfo{}
	pack, err := impl.packageDb.Get(packageId)
	if err != nil {
		goto failed
	}
	if pack == nil {
		err = errors.New("package not exists")
		goto failed
	}
	bindInfo.Domain = pack.CloudapiDomain
	return res.SetSuccessAndData(bindInfo)
failed:
	log.Errorf("error happened, err:%+v", err)
	return res

}

// 绑定阿里云API网关分组
func (impl GatewayOpenapiServiceImpl) setCloudapiGroupBind(orgId string, pack *orm.GatewayPackage, domains []string) error {
	if pack.CloudapiDomain != "" {
		return nil
	}
	addonResp, err := bundle.Bundle.ListByAddonName(apistructs.AddonCloudGateway, pack.DiceProjectId, pack.DiceEnv)
	if err != nil {
		return errors.WithStack(err)
	}
	if !addonResp.Success {
		return errors.Errorf("list addon failed, err:%s", addonResp.Error.Msg)
	}
	if len(addonResp.Data) == 0 {
		return errors.New("api instance not found")
	}
	config := addonResp.Data[0].Config
	var instanceId, vpcGrant string
	if instanceIdI, ok := config["ALIYUN_GATEWAY_INSTANCE_ID"]; !ok {
		return errors.Errorf("invalid config: %+v", config)

	} else if instanceId, ok = instanceIdI.(string); !ok {
		return errors.New("convert failed")
	}
	if vpcGrantI, ok := config["ALIYUN_GATEWAY_VPC_GRANT"]; !ok {
		return errors.Errorf("invalid config: %+v", config)
	} else if vpcGrant, ok = vpcGrantI.(string); !ok {
		return errors.New("convert failed")
	}
	bindInfo, err := impl.createCloudapiGroup(*pack, orgId, instanceId, vpcGrant)
	if err != nil {
		return err
	}
	pack.CloudapiDomain = bindInfo.Domain
	pack.CloudapiGroupId = bindInfo.GroupId
	pack.CloudapiInstanceId = instanceId
	pack.CloudapiVpcGrant = vpcGrant
	err = impl.packageDb.Update(pack, "cloudapi_instance_id", "cloudapi_group_id", "cloudapi_domain", "cloudapi_vpc_grant")
	if err != nil {
		return err
	}
	err = impl.domainBiz.SetCloudapiDomain(pack, domains)
	if err != nil {
		return err
	}
	return nil
}

// 获取阿里云API网关信息
func (impl GatewayOpenapiServiceImpl) GetCloudapiInfo(projectId, env string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	addonResp, err := bundle.Bundle.ListByAddonName(apistructs.AddonCloudGateway, projectId, env)
	if err != nil {
		err = errors.WithStack(err)
		goto failed
	}
	if !addonResp.Success {
		err = errors.Errorf("list addon failed, err:%s", addonResp.Error.Msg)
		goto failed
	}
	if len(addonResp.Data) == 0 {
		return res.SetSuccessAndData(&gw.CloudapiInfo{CloudapiExists: false})
	}
	return res.SetSuccessAndData(&gw.CloudapiInfo{CloudapiExists: true})
failed:
	log.Errorf("error happened, err:%+v", err)
	return res.SetErrorInfo(&common.ErrInfo{
		Msg: errors.Cause(err).Error(),
	})
}

// 自动绑定阿里云API网关分组
func (impl GatewayOpenapiServiceImpl) SetCloudapiGroupBind(orgId, packageId string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	var domains []string
	pack, err := impl.packageDb.Get(packageId)
	if err != nil {
		goto failed
	}
	if pack == nil {
		err = errors.New("package not exists")
		goto failed
	}
	if pack.CloudapiNeedBind == 1 {
		log.Warnf("cloudapi already binding or binded, pack:%+v", pack)
		return res.SetSuccessAndData(&gw.CloudapiBindInfo{})
	}
	domains, _ = impl.domainBiz.GetPackageDomains(packageId)
	pack.CloudapiNeedBind = 1
	_ = impl.packageDb.Update(pack, "cloudapi_need_bind")
	go func() {
		defer util.DoRecover()
		verr := impl.setCloudapiGroupBind(orgId, pack, domains)
		if verr != nil {
			pack.CloudapiNeedBind = 0
			_ = impl.packageDb.Update(pack, "cloudapi_need_bind")
			log.Errorf("error happended: %+v", verr)
		}
		apis, verr := impl.packageApiDb.SelectByAny(&orm.GatewayPackageApi{
			PackageId: packageId,
		})
		if verr != nil {
			log.Errorf("error happended: %+v", verr)
			return
		}
		for _, api := range apis {
			verr = impl.touchCloudapiApi(pack, api)
			if verr != nil {
				log.Errorf("error happended: %+v", verr)
			}
		}
	}()
	return res.SetSuccessAndData(&gw.CloudapiBindInfo{})
failed:
	log.Errorf("error happened, err:%+v", err)
	return res

}

func (impl GatewayOpenapiServiceImpl) createCloudapiGroup(pack orm.GatewayPackage, orgId, instanceId, vpcGrant string) (*gw.CloudapiGroupInfo, error) {
	endpointType := cloudapi.GetEndpointType()
	endpointMap := cloudapi.GetEndpointMap()
	descGroupsReq := cloudapi.CreateDescribeApiGroupsRequest()
	groupName := strings.ToLower(fmt.Sprintf("dice_%s_%s_%s", strings.ReplaceAll(pack.PackageName, "-", "_"), pack.DiceProjectId, pack.DiceEnv))
	descGroupsReq.GroupName = groupName
	descGroupsReq.InstanceId = instanceId
	descGroupResp := cloudapi.CreateDescribeApiGroupsResponse()
	err := bundle.Bundle.DoRemoteAliyunAction(orgId, pack.DiceClusterName, endpointType, endpointMap, descGroupsReq, descGroupResp)
	if err != nil {
		return nil, err
	}
	if descGroupResp.TotalCount > 0 {
		return &gw.CloudapiGroupInfo{
			Domain:  descGroupResp.ApiGroupAttributes.ApiGroupAttribute[0].SubDomain,
			GroupId: descGroupResp.ApiGroupAttributes.ApiGroupAttribute[0].GroupId,
		}, nil
	}
	createGroupReq := cloudapi.CreateCreateApiGroupRequest()
	createGroupReq.InstanceId = instanceId
	createGroupReq.SecurityToken = uuid.UUID()
	createGroupReq.GroupName = groupName
	createGroupReq.Description = fmt.Sprintf("project:%s workspace:%s endpoint:%s", pack.DiceProjectId, pack.DiceEnv, pack.PackageName)
	createGroupResp := cloudapi.CreateCreateApiGroupResponse()
	err = bundle.Bundle.DoRemoteAliyunAction(orgId, pack.DiceClusterName, endpointType, endpointMap, createGroupReq, createGroupResp)
	if err != nil {
		return nil, err
	}
	// 设置请求头Host透传
	modifyGroupReq := cloudapi.CreateModifyApiGroupRequest()
	modifyGroupReq.PassthroughHeaders = "host"
	modifyGroupReq.GroupId = createGroupResp.GroupId
	modifyGroupReq.CustomTraceConfig = `{"parameterLocation":"HEADER","parameterName":"EagleEye-TraceId"}`
	modifyGroupReq.SecurityToken = uuid.UUID()
	modifyGroupResp := cloudapi.CreateModifyApiGroupResponse()
	err = bundle.Bundle.DoRemoteAliyunAction(orgId, pack.DiceClusterName, endpointType, endpointMap, modifyGroupReq, modifyGroupResp)
	if err != nil {
		return nil, err
	}
	return &gw.CloudapiGroupInfo{
		Domain:  createGroupResp.SubDomain,
		GroupId: createGroupResp.GroupId,
	}, nil
}

func (impl GatewayOpenapiServiceImpl) TouchPackageRootApi(packageId string, reqDto *gw.OpenapiDto) (res *common.StandardResult) {
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
		midRes := impl.CreatePackageApi(packageId, apiDto)
		if !midRes.Success {
			err = errors.Errorf("err:%+v", midRes.Err)
			return
		}
		res.SetSuccessAndData(true)
		return
	}
	midRes := impl.UpdatePackageApi(packageId, api.Id, apiDto)
	if !midRes.Success {
		err = errors.Errorf("err:%+v", midRes.Err)
		return
	}
	res.SetSuccessAndData(true)
	return
}

func (impl GatewayOpenapiServiceImpl) endpointPadding(endpoint *diceyml.Endpoint, service *orm.GatewayRuntimeService) error {
	if strings.HasSuffix(endpoint.Domain, ".*") {
		azInfo, err := impl.azDb.GetAzInfoByClusterName(service.ClusterName)
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
	domains, err = impl.domainBiz.FindDomains(domain, projectId, workspace, orm.ExactMatch, orm.DT_PACKAGE)
	if err != nil {
		return
	}
	if len(domains) == 0 {
		allDomains, err = impl.domainBiz.FindDomains(domain, "", "", orm.ExactMatch)
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

func (impl GatewayOpenapiServiceImpl) createRuntimeEndpointPackage(domain string, service *orm.GatewayRuntimeService) (id string, err error) {
	packageName := domain
	res := impl.CreatePackage(&gw.DiceArgsDto{
		ProjectId: service.ProjectId,
		Env:       service.Workspace,
	}, &gw.PackageDto{
		Name:       packageName,
		BindDomain: []string{domain},
		Scene:      gw.WEBAPI_SCENE,
	})
	if !res.Success {
		err = errors.Errorf("create package failed, err:%v", res.Err)
		return
	}
	if dto, ok := res.Data.(*gw.PackageInfoDto); !ok {
		err = errors.New("convert failed")
		return
	} else {
		id = dto.Id
		return
	}
}

func (impl GatewayOpenapiServiceImpl) createRuntimeEndpointRoute(packageId string, endpoint diceyml.Endpoint, service *orm.GatewayRuntimeService) (id string, err error) {
	res := impl.CreatePackageApi(packageId, &gw.OpenapiDto{
		ApiPath:          endpoint.Path,
		RedirectPath:     endpoint.BackendPath,
		RedirectType:     gw.RT_SERVICE,
		RedirectApp:      service.AppName,
		RedirectService:  service.ServiceName,
		RuntimeServiceId: service.Id,
		Origin:           gw.FROM_DICEYML,
	})
	if !res.Success {
		err = errors.Errorf("create package api failed, err:%v", res.Err)
		return
	}
	if apiId, ok := res.Data.(string); !ok {
		err = errors.New("convert failed")
		return
	} else {
		id = apiId
		return
	}
}

func (impl GatewayOpenapiServiceImpl) updateRuntimeEndpointRoute(packageId, packageApiId string, endpoint diceyml.Endpoint, service *orm.GatewayRuntimeService) error {
	res := impl.UpdatePackageApi(packageId, packageApiId, &gw.OpenapiDto{
		ApiPath:          endpoint.Path,
		RedirectPath:     endpoint.BackendPath,
		RedirectType:     gw.RT_SERVICE,
		RedirectApp:      service.AppName,
		RedirectService:  service.ServiceName,
		RuntimeServiceId: service.Id,
		Origin:           gw.FROM_DICEYML,
	})
	if !res.Success {
		return errors.Errorf("update package api failed, err:%v", res.Err)
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
	res := impl.policyBiz.SetPolicyConfig(category, packageId, packageApiId, configByte)
	if !res.Success {
		return errors.Errorf("set policy:%s failed, err: %v", category, res.Err)
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) clearRoutePolicy(category, packageId, packageApiId string) error {
	engine, err := apipolicy.GetPolicyEngine(category)
	if err != nil {
		return err
	}
	config := engine.CreateDefaultConfig(map[string]interface{}{})
	configByte, err := json.Marshal(config)
	if err != nil {
		return errors.Errorf("mashal failed, config:%v, err:%v",
			config, err)
	}
	res := impl.policyBiz.SetPolicyConfig(category, packageId, packageApiId, configByte)
	if !res.Success {
		return errors.Errorf("set policy:%s failed, err: %v", category, res.Err)
	}
	return nil
}

func (impl GatewayOpenapiServiceImpl) setRuntimeEndpointRoutePolicy(packageId, packageApiId string, policies diceyml.EndpointPolicies) error {
	if policies.Cors != nil {
		err := impl.setRoutePolicy("cors", packageId, packageApiId, *policies.Cors)
		if err != nil {
			return err
		}
	} else {
		// clear
		err := impl.clearRoutePolicy("cors", packageId, packageApiId)
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
			res := impl.DeletePackageApi(api.PackageId, api.Id)
			if !res.Success {
				return errors.Errorf("delete package api failed, err:%v", res.Err)
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

func (impl GatewayOpenapiServiceImpl) SetRuntimeEndpoint(info RuntimeEndpointInfo) error {
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
			humanLog := fmt.Sprintf("服务: %s 的域名: %s 绑定失败, 已经被占用", info.RuntimeService.ServiceName, endpoint.Domain)
			log.Error(rawLog)
			go common.AsyncRuntimeError(info.RuntimeService.RuntimeId, humanLog, rawLog)
			continue
		}
		if allowCreate {
			packageId, err := impl.createRuntimeEndpointPackage(endpoint.Domain, info.RuntimeService)
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
				humanLog := fmt.Sprintf("服务: %s 的域名路由: %s%s 绑定失败，已经被占用", info.RuntimeService.ServiceName,
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
		res := impl.DeletePackageApi(api.PackageId, api.Id)
		if !res.Success {
			return errors.Errorf("delete api failed, err:%v", res.Err)
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
