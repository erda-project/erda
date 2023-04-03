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
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/common"
	. "github.com/erda-project/erda/internal/tools/orchestrator/hepa/common/vars"
	gconfig "github.com/erda-project/erda/internal/tools/orchestrator/hepa/config"
	gateway_providers "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers"
	providerDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/kong"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse"
	mseCommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	gw "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/k8s"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	db "github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/api_policy"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/zone"
)

type GatewayZoneServiceImpl struct {
	azDb         db.GatewayAzInfoService
	kongDb       db.GatewayKongInfoService
	packageDb    db.GatewayPackageService
	packageApiDb db.GatewayPackageApiService
	zoneDb       db.GatewayZoneService
	zoneInPackDb db.GatewayZoneInPackageService
	apiPolicyBiz *api_policy.GatewayApiPolicyService
}

var once sync.Once

func NewGatewayZoneServiceImpl() (e error) {
	once.Do(
		func() {
			azDb, err := db.NewGatewayAzInfoServiceImpl()
			if err != nil {
				e = err
				return
			}
			zoneDb, err := db.NewGatewayZoneServiceImpl()
			if err != nil {
				e = err
				return
			}
			packageDb, err := db.NewGatewayPackageServiceImpl()
			if err != nil {
				e = err
				return
			}
			packageApiDb, err := db.NewGatewayPackageApiServiceImpl()
			if err != nil {
				e = err
				return
			}
			kongDb, err := db.NewGatewayKongInfoServiceImpl()
			if err != nil {
				e = err
				return
			}
			zoneInPackDb, err := db.NewGatewayZoneInPackageServiceImpl()
			if err != nil {
				e = err
				return
			}
			zone.Service = &GatewayZoneServiceImpl{
				azDb:         azDb,
				zoneDb:       zoneDb,
				kongDb:       kongDb,
				packageDb:    packageDb,
				zoneInPackDb: zoneInPackDb,
				packageApiDb: packageApiDb,
				apiPolicyBiz: &api_policy.Service,
			}
		})
	return
}

func (impl GatewayZoneServiceImpl) GetRegisterAppInfo(orgId, projectId, env string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if projectId == "" || env == "" {
		log.Errorf("projectId, env is empty")
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	az, err := impl.azDb.GetAz(&orm.GatewayAzInfo{
		OrgId:     orgId,
		Env:       env,
		ProjectId: projectId,
	})
	if err != nil {
		log.Errorf("get az failed, err:%+v", err)
		return res
	}
	cond := &orm.GatewayZone{
		DiceProjectId:   projectId,
		DiceEnv:         env,
		DiceClusterName: az,
	}
	zones, err := impl.zoneDb.SelectByAny(cond)
	if err != nil {
		log.Errorf("get zone failed, cond:%+v, err:%+v", cond, err)
		return res
	}
	appMap := map[string]map[string]bool{}
	for _, zone := range zones {
		appName := zone.DiceApp
		svcName := zone.DiceService
		if appName == "" || svcName == "" {
			continue
		}
		serviceMap, exist := appMap[appName]
		if !exist {
			appMap[appName] = map[string]bool{}
			appMap[appName][svcName] = true
			continue
		}
		serviceMap[svcName] = true
	}
	resDto := gw.RegisterAppsDto{
		Apps: []gw.RegisterApp{},
	}
	for appName, services := range appMap {
		app := gw.RegisterApp{
			Name: appName,
		}
		for svcName := range services {
			app.Services = append(app.Services, svcName)
		}
		resDto.Apps = append(resDto.Apps, app)
	}
	return res.SetSuccessAndData(resDto)
}

func (impl GatewayZoneServiceImpl) UpdateBuiltinPolicies(zoneId string) error {
	zone, err := impl.zoneDb.GetById(zoneId)
	if err != nil {
		return err
	}
	_, _, err = (*impl.apiPolicyBiz).SetZonePolicyConfig(zone, nil, apipolicy.Policy_Engine_Built_in, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayZoneServiceImpl) GetGatewayProvider(clusterName string) (string, error) {
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

func (impl GatewayZoneServiceImpl) UpdateKongDomainPolicy(az, projectId, env string, helper *db.SessionHelper) error {
	var zoneDbService db.GatewayZoneService
	var packDbService db.GatewayPackageService
	var packageApiService db.GatewayPackageApiService
	var kongService db.GatewayKongInfoService
	var gatewayAdapter gateway_providers.GatewayAdapter
	var err error
	if helper != nil {
		zoneDbService, err = impl.zoneDb.NewSession(helper)
		if err != nil {
			return err
		}
		packDbService, err = impl.packageDb.NewSession(helper)
		if err != nil {
			return err
		}
		packageApiService, err = impl.packageApiDb.NewSession(helper)
		if err != nil {
			return err
		}
		kongService, err = impl.kongDb.NewSession(helper)
		if err != nil {
			return err
		}
	} else {
		zoneDbService = impl.zoneDb
		packDbService = impl.packageDb
		packageApiService = impl.packageApiDb
		kongService = impl.kongDb
	}

	gatewayProvider := ""
	gatewayProvider, err = impl.GetGatewayProvider(az)
	if err != nil {
		return err
	}

	kongInfo, err := kongService.GetKongInfo(&orm.GatewayKongInfo{
		Az:        az,
		ProjectId: projectId,
		Env:       env,
	})
	if err != nil {
		return err
	}
	// tuples, err := kongService.GetBelongTuples(kongInfo.AddonInstanceId)
	// if err != nil {
	// 	return err
	// }
	regexs := []string{}
	ids := []string{}
	enables := []string{}
	disables := []string{}
	allows := []string{}
	packs := []string{}
	dpids := []string{}
	denvs := []string{}
	zones, err := zoneDbService.SelectPolicyZones(az)
	if err != nil {
		return err
	}
	var domainPolicyList []gw.ZoneKongPolicies
	for _, zone := range zones {
		if len(zone.KongPolicies) == 0 {
			continue
		}
		var kongPolicies gw.ZoneKongPolicies
		err = json.Unmarshal(zone.KongPolicies, &kongPolicies)
		if err != nil {
			return errors.WithStack(err)
		}
		if kongPolicies.PackageName == "" {
			var pack *orm.GatewayPackage
			if zone.Type == db.ZONE_TYPE_PACKAGE || zone.Type == db.ZONE_TYPE_PACKAGE_NEW || zone.Type == db.ZONE_TYPE_UNITY {
				pack, err = packDbService.GetByAny(&orm.GatewayPackage{ZoneId: zone.Id})
				if err != nil {
					return err
				}
			} else if zone.Type == db.ZONE_TYPE_PACKAGE_API {
				packageApi, err := packageApiService.GetByAny(&orm.GatewayPackageApi{ZoneId: zone.Id})
				if err != nil {
					return err
				}
				pack, err = packDbService.Get(packageApi.PackageId)
				if err != nil {
					return err
				}
			}
			if pack != nil {
				kongPolicies.PackageName = pack.PackageName
			}
		}
		kongPolicies.ProjectId = zone.DiceProjectId
		kongPolicies.Env = strings.ToLower(zone.DiceEnv)
		domainPolicyList = append(domainPolicyList, kongPolicies)
	}
	sortList := gw.SortByRegexList(domainPolicyList)
	sort.Sort(sortList)
	for _, kongPolicies := range sortList {
		ids = append(ids, kongPolicies.Id)
		regexs = append(regexs, kongPolicies.Regex)
		enables = append(enables, kongPolicies.Enables)
		disables = append(disables, kongPolicies.Disables)
		allows = append(allows, "1")
		packs = append(packs, kongPolicies.PackageName)
		dpids = append(dpids, kongPolicies.ProjectId)
		denvs = append(denvs, kongPolicies.Env)
	}

	switch gatewayProvider {
	case mseCommon.Mse_Provider_Name:
		gatewayAdapter, err = mse.NewMseAdapter(az)
		if err != nil {
			return err
		}
	case "":
		gatewayAdapter = kong.NewKongAdapter(kongInfo.KongAddr)
		_, err = gatewayAdapter.CreateOrUpdatePlugin(&providerDto.PluginReqDto{
			Name: "domain-policy",
			Config: map[string]interface{}{
				"regexs":   regexs,
				"ids":      ids,
				"enables":  enables,
				"disables": disables,
				"allows":   allows,
				"packs":    packs,
				"dpids":    dpids,
				"denvs":    denvs,
			},
		})
		if err != nil {
			return err
		}
	default:
		return errors.Errorf("unknown gateway provider:%v\n", gatewayProvider)
	}
	return nil
}

func (impl GatewayZoneServiceImpl) SetZoneKongPoliciesWithoutDomainPolicy(zoneId string, policies *gw.ZoneKongPolicies, helper *db.SessionHelper) error {
	var zoneDbService db.GatewayZoneService
	var err error
	if helper != nil {
		zoneDbService, err = impl.zoneDb.NewSession(helper)
		if err != nil {
			return err
		}
	} else {
		zoneDbService = impl.zoneDb
	}
	zone, err := zoneDbService.GetById(zoneId)
	if err != nil {
		return err
	}
	if zone == nil {
		return errors.New("zone not find")
	}
	if policies != nil {
		policyStr, err := json.Marshal(policies)
		if err != nil {
			return errors.WithStack(err)
		}
		zone.KongPolicies = policyStr
	} else {
		zone.KongPolicies = []byte{}
	}
	err = zoneDbService.Update(zone)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayZoneServiceImpl) SetZoneKongPolicies(zoneId string, policies gw.ZoneKongPolicies, helper *db.SessionHelper) error {
	var zoneDbService db.GatewayZoneService
	var err error
	if helper != nil {
		zoneDbService, err = impl.zoneDb.NewSession(helper)
		if err != nil {
			return err
		}
	} else {
		zoneDbService = impl.zoneDb
	}
	zone, err := zoneDbService.GetById(zoneId)
	if err != nil {
		return err
	}
	if zone == nil {
		return errors.New("zone not find")
	}
	err = impl.SetZoneKongPoliciesWithoutDomainPolicy(zoneId, &policies, helper)
	if err != nil {
		return err
	}
	err = impl.UpdateKongDomainPolicy(zone.DiceClusterName, zone.DiceProjectId, zone.DiceEnv, helper)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayZoneServiceImpl) DeleteZone(zoneId, namespace string) error {
	zone, err := impl.zoneDb.GetById(zoneId)
	if err != nil {
		return err
	}
	if zone == nil {
		return nil
	}
	err = impl.DeleteZoneRoute(zoneId, namespace)
	if err != nil {
		return err
	}
	err = impl.zoneDb.DeleteById(zoneId)
	if err != nil {
		return err
	}
	err = impl.UpdateKongDomainPolicy(zone.DiceClusterName, zone.DiceProjectId, zone.DiceEnv, nil)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayZoneServiceImpl) CreateZoneWithoutIngress(config zone.ZoneConfig, session ...*db.SessionHelper) (*orm.GatewayZone, error) {
	var zoneSession db.GatewayZoneService
	var err error
	if len(session) > 0 {
		zoneSession, err = impl.zoneDb.NewSession(session[0])
		if err != nil {
			return nil, err
		}
	} else {
		zoneSession = impl.zoneDb
	}
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	id := strings.Replace(uuid.String(), "-", "", -1)
	name := fmt.Sprintf("dice-%s-%s-%s-%s%s", config.Env, config.ProjectId, config.Name, id[:3], id[len(id)-3:])
	name = strings.Replace(name, "_", "-", -1)
	zone := &orm.GatewayZone{
		BaseRow:         orm.BaseRow{Id: id},
		Name:            name,
		DiceProjectId:   config.ProjectId,
		DiceEnv:         config.Env,
		DiceClusterName: config.Az,
		Type:            config.Type,
	}
	err = zoneSession.Insert(zone)
	if err != nil {
		return nil, err
	}
	return zone, nil
}

func (impl GatewayZoneServiceImpl) CreateZone(config zone.ZoneConfig, session ...*db.SessionHelper) (*orm.GatewayZone, error) {
	var kongSession db.GatewayKongInfoService
	var err error
	var azInfo *db.ClusterInfoDto
	var gatewayProvider string
	var namespace, svcName string
	if len(session) > 0 {
		kongSession, err = impl.kongDb.NewSession(session[0])
		if err != nil {
			return nil, err
		}
	} else {
		kongSession = impl.kongDb
	}
	// 对接 MSE 这个字段用来识别是不是要创建后续的 Nginx-Kong 的 Ingress 对象，处理之前进行还原操作
	if config.Type == db.ZONE_TYPE_UNITY_Provider {
		config.Type = db.ZONE_TYPE_UNITY
		zone, err := impl.CreateZoneWithoutIngress(config, session...)
		if err != nil {
			return nil, err
		}
		return zone, nil
	}
	zone, err := impl.CreateZoneWithoutIngress(config, session...)
	if err != nil {
		return nil, err
	}
	az, azInfo, err := impl.azDb.GetAzInfoByClusterName(config.Az)
	if err != nil {
		return nil, err
	}
	if az == nil {
		return nil, errors.New("find cluster failed")
	}

	if azInfo == nil {
		return nil, errors.New("find clusterInfo failed")
	}
	gatewayProvider = azInfo.GatewayProvider

	if (az.Type != orm.AT_K8S && az.Type != orm.AT_EDAS) || gconfig.ServerConf.UseAdminEndpoint {
		return zone, nil
	}
	if config.ZoneRoute == nil {
		return zone, nil
	}
	adapter, err := k8s.NewAdapter(config.Az)
	if err != nil {
		return nil, err
	}
	switch gatewayProvider {
	case mseCommon.Mse_Provider_Name:
		//TODO: get svcName
		namespace = config.Namespace
		svcName = config.ServiceName
		if config.Type == db.ZONE_TYPE_PACKAGE_NEW || config.Type == db.ZONE_TYPE_UNITY {
			// 走到这里表示 mse 网关创建网关入口，无需创建对应的 ingress
			return zone, nil
		}
	case "":
		namespace, svcName, err = kongSession.GetK8SInfo(&orm.GatewayKongInfo{
			ProjectId: config.ProjectId,
			Az:        config.Az,
			Env:       config.Env,
		})
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.Errorf("unknown gateway provider:%v\n", gatewayProvider)
	}

	if namespace == "" || svcName == "" {
		// 走到这里表示 mse 网关需要创建 路由对应的 ingress， 但是无法获取目标服务对应的 k8s Service 名称和 Namespace
		log.Errorf("use mse gateway but can not get Service name or namespace: [namespace=%s] [svcName=%s] \n", namespace, svcName)
		return zone, errors.Errorf("use mse gateway but can not get Service name or namespace: [namespace=%s] [svcName=%s] \n", namespace, svcName)
	}

	_, err = createOrUpdateZoneIngress(adapter, namespace, svcName, zone.Name, config.Route, gatewayProvider)
	if err != nil {
		goto clear_ingress
	}
	return zone, nil
clear_ingress:
	clearErr := clearZoneIngress(adapter, namespace, zone.Name)
	if clearErr != nil {
		log.Errorf("clear zone ingress failed, err:%+v", clearErr)
	}
	return nil, err
}

func (impl GatewayZoneServiceImpl) UpdateZoneRoute(zoneId string, route zone.ZoneRoute, runtimeService *orm.GatewayRuntimeService, redirectType string, session ...*db.SessionHelper) (bool, error) {
	var zoneService db.GatewayZoneService
	var err error
	if len(session) > 0 {
		zoneService, err = impl.zoneDb.NewSession(session[0])
		if err != nil {
			return false, err
		}
	} else {
		zoneService = impl.zoneDb
	}
	zone, err := zoneService.GetById(zoneId)
	if err != nil {
		return false, err
	}
	if zone == nil {
		return false, errors.Errorf("zone not find, id:%s", zoneId)
	}

	az, azInfo, err := impl.azDb.GetAzInfoByClusterName(zone.DiceClusterName)
	if err != nil {
		return false, err
	}

	useKong := true
	if azInfo != nil && azInfo.GatewayProvider != "" {
		useKong = false
	}
	if (az.Type != orm.AT_K8S && az.Type != orm.AT_EDAS) || gconfig.ServerConf.UseAdminEndpoint {
		return false, errors.Errorf("clusterType:%s, not support api route config", az.Type)
	}
	adapter, err := k8s.NewAdapter(zone.DiceClusterName)
	if err != nil {
		return false, err
	}
	namespace := ""
	svcName := ""
	if useKong {
		namespace, svcName, err = impl.kongDb.GetK8SInfo(&orm.GatewayKongInfo{
			ProjectId: zone.DiceProjectId,
			Az:        zone.DiceClusterName,
			Env:       zone.DiceEnv,
		})
		if err != nil {
			return false, err
		}
	} else {
		// 从 tb_gateway_runtime_service 获取到 namespace
		if runtimeService != nil {
			if runtimeService.ProjectNamespace != "" {
				namespace = runtimeService.ProjectNamespace
			}
			svcName = runtimeService.ServiceName + "-" + runtimeService.GroupName
		}

		// 按地址转发， namespace 和 svcName 固定
		if (namespace == "" || svcName == "") && redirectType == gw.RT_URL {
			namespace = "project-" + zone.DiceProjectId + "-" + strings.ToLower(zone.DiceEnv)
			svcName = strings.ToLower(zone.Name)
		}
	}

	if namespace == "" || svcName == "" {
		if err != nil {
			return false, errors.Errorf("can not get namespace or svcName")
		}
	}

	gatewayProvider := ""
	if !useKong {
		gatewayProvider = azInfo.GatewayProvider
	}
	exist, err := createOrUpdateZoneIngress(adapter, namespace, svcName, zone.Name, route.Route, gatewayProvider)
	if err != nil {
		return exist, err
	}
	return exist, nil
}

func (impl GatewayZoneServiceImpl) DeleteZoneRoute(zoneId string, namespace string, session ...*db.SessionHelper) error {
	var zoneSession db.GatewayZoneService
	var kongSession db.GatewayKongInfoService
	var err error
	if len(session) > 0 {
		zoneSession, err = impl.zoneDb.NewSession(session[0])
		if err != nil {
			return err
		}
		kongSession, err = impl.kongDb.NewSession(session[0])
		if err != nil {
			return err
		}
	} else {
		zoneSession = impl.zoneDb
		kongSession = impl.kongDb
	}
	zone, err := zoneSession.GetById(zoneId)
	if err != nil {
		return err
	}
	if zone == nil {
		return errors.Errorf("zone not find, id:%s", zoneId)
	}
	// need not delete ingress if the zone dose not hava ingress
	if !zone.HasIngress() {
		return nil
	}

	gatewayProvider := ""
	gatewayProvider, err = impl.GetGatewayProvider(zone.DiceClusterName)
	if err != nil {
		return err
	}

	adapter, err := k8s.NewAdapter(zone.DiceClusterName)
	if err != nil {
		return err
	}

	switch gatewayProvider {
	case mseCommon.Mse_Provider_Name:
		if namespace == "" {
			// 走到这里，应该是按地址转发，设置固定 namespace
			namespace = "project-" + zone.DiceProjectId + "-" + strings.ToLower(zone.DiceEnv)
		}
	case "":
		namespace, _, err = kongSession.GetK8SInfo(&orm.GatewayKongInfo{
			ProjectId: zone.DiceProjectId,
			Az:        zone.DiceClusterName,
			Env:       zone.DiceEnv,
		})
		if err != nil {
			return err
		}
	default:
		return errors.Errorf("unknown gateway provider:%v\n", gatewayProvider)
	}

	return clearZoneIngress(adapter, namespace, zone.Name)
}

func clearZoneIngress(adapter k8s.K8SAdapter, namespace, zoneName string) error {
	err := adapter.DeleteIngress(namespace, zoneName)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateZoneIngress(adapter k8s.K8SAdapter, namespace, svcName, ingName string, config zone.RouteConfig, gatewayProvider string) (bool, error) {
	var routes []k8s.IngressRoute
	for _, host := range config.Hosts {
		routes = append(routes, k8s.IngressRoute{
			Domain: host,
			Path:   config.Path,
		})
	}
	//TODO: 支持配置
	config.EnableTLS = true

	kongBackend := k8s.IngressBackend{}

	switch gatewayProvider {
	case mseCommon.Mse_Provider_Name:
		kongBackend.ServiceName = svcName
		services, err := adapter.ListAllServices(namespace)
		if err != nil {
			return false, err
		}
		if len(services) == 0 {
			return false, errors.Errorf("not found service in namespace %s", namespace)
		}
		kongBackend.ServicePort = DEFALUT_SERVICE_PORT
		for _, svc := range services {
			if svc.Name == svcName {
				kongBackend.ServicePort = int(svc.Spec.Ports[0].Port)
				break
			}
		}
	case "":
		kongBackend.ServiceName = svcName
		kongBackend.ServicePort = KONG_SERVICE_PORT
		if config.BackendProtocol != nil && *config.BackendProtocol == k8s.HTTPS {
			supportHttps, err := adapter.IsGatewaySupportHttps(namespace, svcName)
			if err != nil {
				return false, err
			}
			if supportHttps {
				kongBackend.ServicePort = KONG_HTTPS_SERVICE_PORT
			} else {
				// use http if not support https
				config.BackendProtocol = nil
			}
		}
	default:
		log.Errorf("unknown gateway provider:%v\n", gatewayProvider)
		return false, errors.Errorf("unknown gateway provider:%v\n", gatewayProvider)
	}

	exist, err := adapter.CreateOrUpdateIngress(namespace, ingName, routes, kongBackend, gatewayProvider, config.RouteOptions)
	if err != nil {
		return exist, err
	}
	return exist, nil
}

func (impl GatewayZoneServiceImpl) GetZone(id string, session ...*db.SessionHelper) (*orm.GatewayZone, error) {
	var zoneSession db.GatewayZoneService
	var err error
	if len(session) > 0 {
		zoneSession, err = impl.zoneDb.NewSession(session[0])
		if err != nil {
			return nil, err
		}
	} else {
		zoneSession = impl.zoneDb
	}
	return zoneSession.GetById(id)
}
