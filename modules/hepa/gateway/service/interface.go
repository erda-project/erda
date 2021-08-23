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
	"github.com/gin-gonic/gin"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/hepa/apipolicy"
	"github.com/erda-project/erda/modules/hepa/common"
	"github.com/erda-project/erda/modules/hepa/common/vars"
	"github.com/erda-project/erda/modules/hepa/endpoint"
	gw "github.com/erda-project/erda/modules/hepa/gateway/dto"
	"github.com/erda-project/erda/modules/hepa/gateway/exdto"
	. "github.com/erda-project/erda/modules/hepa/k8s"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
	db "github.com/erda-project/erda/modules/hepa/repository/service"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

type DiceInfo struct {
	OrgId       string
	ProjectId   string
	Env         string
	Az          string
	AppName     string
	ServiceName string
}

type EndpointDto struct {
	orm.GatewayRuntimeService
	EndpointDomains []gw.EndpointDomainDto
}

type PackageApiInfo struct {
	*orm.GatewayPackageApi
	Hosts               []string
	ProjectId           string
	Env                 string
	Az                  string
	InjectRuntimeDomain bool
}

type RuntimeEndpointInfo struct {
	RuntimeService *orm.GatewayRuntimeService
	Endpoints      []diceyml.Endpoint
}

type GatewayOpenapiService interface {
	ClearRuntimeRoute(id string) error
	SetRuntimeEndpoint(RuntimeEndpointInfo) error
	TouchPackageRootApi(packageId string, reqDto *gw.OpenapiDto) *common.StandardResult
	TryClearRuntimePackage(*orm.GatewayRuntimeService, *db.SessionHelper, ...bool) error
	TouchRuntimePackageMeta(*orm.GatewayRuntimeService, *db.SessionHelper) (string, bool, error)
	RefreshRuntimePackage(string, *orm.GatewayRuntimeService, *db.SessionHelper) error
	//用于补偿迁移数据生成的unity package
	CreateUnityPackageZone(string, *db.SessionHelper) (*orm.GatewayZone, error)
	CreateTenantPackage(string, *db.SessionHelper) error
	CreatePackage(*gw.DiceArgsDto, *gw.PackageDto) *common.StandardResult
	GetPackages(*gw.GetPackagesDto) *common.StandardResult
	GetPackage(string) *common.StandardResult
	GetPackagesName(*gw.GetPackagesDto) *common.StandardResult
	UpdatePackage(string, *gw.PackageDto) *common.StandardResult
	DeletePackage(string) *common.StandardResult
	CreatePackageApi(string, *gw.OpenapiDto) *common.StandardResult
	GetPackageApis(string, *gw.GetOpenapiDto) *common.StandardResult
	UpdatePackageApi(string, string, *gw.OpenapiDto) *common.StandardResult
	DeletePackageApi(string, string) *common.StandardResult
	TouchPackageApiZone(info PackageApiInfo, session ...*db.SessionHelper) (string, error)
	// 获取阿里云API网关信息
	GetCloudapiInfo(projectId, env string) *common.StandardResult
	// 获取绑定的阿里云API网关分组域名
	GetCloudapiGroupBind(string) *common.StandardResult
	// 自动绑定阿里云API网关分组
	SetCloudapiGroupBind(orgId, projectId string) *common.StandardResult
}

type GatewayOrgClientService interface {
	Create(orgId, name string) *common.StandardResult
	Delete(id string) *common.StandardResult
	GetCredentials(id string) *common.StandardResult
	UpdateCredentials(id string, secret ...string) *common.StandardResult
	GrantPackage(id, packageId string) *common.StandardResult
	RevokePackage(id, packageId string) *common.StandardResult
	CreateOrUpdateLimit(id, packageId string, limits exdto.ChangeLimitsReq) *common.StandardResult
}

type GatewayOpenapiConsumerService interface {
	GrantPackageToConsumer(consumerId, packageId string) error
	RevokePackageFromConsumer(consumerId, packageId string) error
	CreateClientConsumer(clientName, clientId, clientSecret, clusterName string) (*orm.GatewayConsumer, error)
	CreateConsumer(*gw.DiceArgsDto, *gw.OpenConsumerDto) *common.StandardResult
	GetConsumers(*gw.GetOpenConsumersDto) *common.StandardResult
	GetConsumersName(*gw.GetOpenConsumersDto) *common.StandardResult
	UpdateConsumer(string, *gw.OpenConsumerDto) *common.StandardResult
	DeleteConsumer(string) *common.StandardResult
	GetConsumerCredentials(string) *common.StandardResult
	UpdateConsumerCredentials(string, *gw.ConsumerCredentialsDto) *common.StandardResult
	GetConsumerAcls(string) *common.StandardResult
	UpdateConsumerAcls(string, *gw.ConsumerAclsDto) *common.StandardResult
	GetConsumersOfPackage(string) ([]orm.GatewayConsumer, error)
	GetKongConsumerName(consumer *orm.GatewayConsumer) string
	GetPackageAcls(string) *common.StandardResult
	UpdatePackageAcls(string, *gw.PackageAclsDto) *common.StandardResult
	GetPackageApiAcls(string, string) *common.StandardResult
	UpdatePackageApiAcls(string, string, *gw.PackageAclsDto) *common.StandardResult
	// 获取阿里云APP凭证
	GetCloudapiAppCredential(string) *common.StandardResult
	// 生成阿里云APP凭证
	SetCloudapiAppCredential(string, ...bool) *common.StandardResult
	// 删除阿里云APP凭证
	DeleteCloudapiAppCredential(string) *common.StandardResult
}

type GatewayOpenapiRuleService interface {
	CreateOrUpdateLimitRule(consumerId, packageId string, limits []exdto.LimitType) error
	CreateLimitRule(*gw.DiceArgsDto, *gw.OpenLimitRuleDto) *common.StandardResult
	UpdateLimitRule(string, *gw.OpenLimitRuleDto) *common.StandardResult
	GetLimitRules(*gw.GetOpenLimitRulesDto) *common.StandardResult
	DeleteLimitRule(string) *common.StandardResult
	CreateRule(DiceInfo, *gw.OpenapiRule, *db.SessionHelper) error
	UpdateRule(string, *gw.OpenapiRule) (*orm.GatewayPackageRule, error)
	// use session if helper not nil
	GetPackageRules(string, *db.SessionHelper, ...gw.RuleCategory) ([]gw.OpenapiRuleInfo, error)
	GetApiRules(string, ...gw.RuleCategory) ([]gw.OpenapiRuleInfo, error)
	DeleteRule(string, *db.SessionHelper) error
	// recycle plugins
	DeleteByPackage(*orm.GatewayPackage) error
	DeleteByPackageApi(*orm.GatewayPackage, *orm.GatewayPackageApi) error
	SetPackageKongPolicies(*orm.GatewayPackage, *db.SessionHelper) error
	SetPackageApiKongPolicies(packageApi *orm.GatewayPackageApi, session *db.SessionHelper) error
}

type GatewayApiService interface {
	GetRuntimeApis(runtimeServiceId string, registerType ...string) ([]gw.ApiDto, error)
	CreateRuntimeApi(dto *gw.ApiDto, session ...*db.SessionHelper) (string, vars.StandardErrorCode, error)
	CreateApi(*gw.ApiReqDto) *common.StandardResult
	GetApiInfos(*gw.GetApisDto) *common.StandardResult
	DeleteApi(string) *common.StandardResult
	UpdateApi(string, *gw.ApiReqDto) *common.StandardResult
	CreateUpstreamBindApi(*orm.GatewayConsumer, string, string, string, *orm.GatewayUpstreamApi, string) (string, error)
	UpdateUpstreamBindApi(*orm.GatewayConsumer, string, string, *orm.GatewayUpstreamApi, string) error
	DeleteUpstreamBindApi(*orm.GatewayUpstreamApi) error
	TouchRuntimeApi(*orm.GatewayRuntimeService, *db.SessionHelper, bool) error
	ClearRuntimeApi(*orm.GatewayRuntimeService) error
}

type GatewayCategoryService interface {
	CreatePolicy(string, *gw.PolicyCreateDto) *common.StandardResult
	UpdatePolicy(string, string, *gw.PolicyCreateDto) *common.StandardResult
	DeletePolicy(string) *common.StandardResult
	GetCategoryInfo(string, string, string, string) *common.StandardResult
}

type GatewayConsumerService interface {
	CreateDefaultConsumer(string, string, string, string) (*orm.GatewayConsumer, *orm.ConsumerAuthConfig, vars.StandardErrorCode, error)
	CreateConsumer(*gw.ConsumerCreateDto) *common.StandardResult
	GetProjectConsumerInfo(string, string, string) *common.StandardResult
	GetConsumerInfo(string) *common.StandardResult
	UpdateConsumerInfo(string, *gw.ConsumerDto) *common.StandardResult
	UpdateConsumerApi(*gw.ConsumerEditDto) *common.StandardResult
	GetConsumerList(string, string, string) *common.StandardResult
	DeleteConsumer(string) *common.StandardResult
}

type GatewayConsumerApiService interface {
	Create(string, string) (string, error)
	Delete(string) error
	UpdateConsumerApi(*gw.ConsumerApiReqDto) *common.StandardResult
}

type GatewayApiPolicyService interface {
	SetPackageDefaultPolicyConfig(category, packageId string, az *orm.GatewayAzInfo, config []byte, helper ...*db.SessionHelper) (string, error)
	GetPolicyConfig(category, packageId, packageApiId string) *common.StandardResult
	SetPolicyConfig(category, packageId, packageApiId string, config []byte) *common.StandardResult
	RefreshZoneIngress(zone orm.GatewayZone, az orm.GatewayAzInfo) error
	SetZonePolicyConfig(zone *orm.GatewayZone, category string, config []byte, helper *db.SessionHelper, needDeployTag ...bool) (apipolicy.PolicyDto, string, error)
	SetZoneDefaultPolicyConfig(packageId string, zone *orm.GatewayZone, az *orm.GatewayAzInfo, session ...*db.SessionHelper) (map[string]*string, *string, *db.SessionHelper, error)
}
type RouteConfig struct {
	Hosts []string
	Path  string
	RouteOptions
}

type ZoneRoute struct {
	Route RouteConfig
	//	InnerRoute *RouteConfig
}

type ZoneConfig struct {
	*ZoneRoute
	Name      string
	ProjectId string
	Env       string
	Az        string
	Type      string
}
type GatewayZoneService interface {
	CreateZoneWithoutIngress(ZoneConfig, ...*db.SessionHelper) (*orm.GatewayZone, error)
	CreateZone(ZoneConfig, ...*db.SessionHelper) (*orm.GatewayZone, error)
	DeleteZoneRoute(string, ...*db.SessionHelper) error
	UpdateZoneRoute(string, ZoneRoute, ...*db.SessionHelper) (bool, error)
	// trigger domain-policy update
	SetZoneKongPolicies(string, gw.ZoneKongPolicies, *db.SessionHelper) error
	SetZoneKongPoliciesWithoutDomainPolicy(zoneId string, policies *gw.ZoneKongPolicies, helper *db.SessionHelper) error
	UpdateKongDomainPolicy(az, projectId, env string, helper *db.SessionHelper) error
	DeleteZone(string) error
	UpdateBuiltinPolicies(string) error
	GetZone(string, ...*db.SessionHelper) (*orm.GatewayZone, error)
}

type GatewayGlobalService interface {
	GetDiceHealth() gw.DiceHealthDto
	GenTenantGroup(projectId, env, clusterName string) string
	GetGatewayFeatures(clusterName string) *common.StandardResult
	GetTenantGroup(projectId, env string) *common.StandardResult
	GetOrgId(string) (string, error)
	GetClusterUIType(string, string, string) *common.StandardResult
	GenerateEndpoint(DiceInfo, ...*db.SessionHelper) (string, string, error)
	GenerateDefaultPath(string, ...*db.SessionHelper) (string, error)
	GetProjectName(DiceInfo, ...*db.SessionHelper) (string, error)
	GetProjectNameFromCmdb(string) (string, error)
	GetClustersByOrg(string) ([]string, error)
	GetServiceAddr(string) string
	GetRuntimeServicePrefix(*orm.GatewayRuntimeService) (string, error)
	CreateTenant(*gw.TenantDto) *common.StandardResult
}

type GatewayMockService interface {
	RegisterMockApi(*gw.MockInfoDto) *common.StandardResult
	CallMockApi(string, string, string) *common.StandardResult
}

type GatewayUpstreamService interface {
	UpstreamValidAsync(*gin.Context, *gw.UpstreamRegisterDto) *common.StandardResult
	UpstreamRegister(*gw.UpstreamRegisterDto) *common.StandardResult
	UpstreamRegisterAsync(*gin.Context, *gw.UpstreamRegisterDto) *common.StandardResult
}

type GatewayUpstreamLbService interface {
	UpstreamTargetOnline(*gw.UpstreamLbDto) *common.StandardResult
	UpstreamTargetOffline(*gw.UpstreamLbDto) *common.StandardResult
}

type GatewayRuntimeServiceService interface {
	TouchRuntimeComplete(*gin.Context, *gw.RuntimeServiceReqDto) *common.StandardResult
	GetRegisterAppInfo(string, string) *common.StandardResult
	TouchRuntime(*gin.Context, *gw.RuntimeServiceReqDto) *common.StandardResult
	DeleteRuntime(string) *common.StandardResult
	GetServiceRuntimes(projectId, env, app, service string) *common.StandardResult
	// 获取指定服务的API前缀
	GetServiceApiPrefix(*gw.ApiPrefixReqDto) *common.StandardResult
}

type GatewayDomainService interface {
	FindDomains(domain, projectId, workspace string, matchType orm.OptionType, domainType ...string) ([]orm.GatewayDomain, error)
	GetOrgDomainInfo(gw.DiceArgsDto, *gw.ManageDomainReq) *common.StandardResult
	UpdateRuntimeServicePort(runtimeService *orm.GatewayRuntimeService, releaseInfo *diceyml.Object) error
	RefreshRuntimeDomain(runtimeService *orm.GatewayRuntimeService, session *db.SessionHelper) error
	GiveRuntimeDomainToPackage(runtimeService *orm.GatewayRuntimeService, session *db.SessionHelper) (bool, error)
	TouchRuntimeDomain(runtimeService *orm.GatewayRuntimeService, material endpoint.EndpointMaterial, domains []gw.EndpointDomainDto, audits *[]apistructs.Audit, session *db.SessionHelper) (string, error)
	TouchPackageDomain(packageId, clusterName string, domains []string, session *db.SessionHelper) ([]string, error)
	GetPackageDomains(packageId string, session ...*db.SessionHelper) ([]string, error)
	IsPackageDomainsDiff(packageId, clusterName string, domains []string, session *db.SessionHelper) (bool, error)
	GetTenantDomains(projectId, env string) *common.StandardResult
	GetRuntimeDomains(runtimeId string) *common.StandardResult
	UpdateRuntimeServiceDomain(orgId, runtimeId, serviceName string, reqDto *gw.ServiceDomainReqDto) *common.StandardResult
	SetCloudapiDomain(pack *orm.GatewayPackage, domains []string) error
	CreateOrUpdateComponentIngress(apistructs.ComponentIngressUpdateRequest) *common.StandardResult
}
