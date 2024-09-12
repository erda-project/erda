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
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/policies/custom"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/bundle"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/common"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/config"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/kong"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse"
	mseCommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/k8s"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/api_policy"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/domain"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/endpoint_api"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/global"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/openapi_rule"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/zone"
)

const (
	mutexBucketSize = 512
)

var azMutex []*sync.Mutex

// Nginx 配置 llocations 部分，more_set_headers、proxy_set_header、set、limit_req、limit_conn、error_page、deny、allow、return、add_header 等允许多次设置
var skipKeys = map[string]bool{
	"more_set_headers": true,
	"proxy_set_header": true,
	"set":              true,
	"limit_req":        true,
	"limit_conn":       true,
	"error_page":       true,
	"deny":             true,
	"allow":            true,
	"return":           true,
	"add_header":       true,
}

func init() {
	for i := 0; i < mutexBucketSize; i++ {
		azMutex = append(azMutex, &sync.Mutex{})
	}
}

func strHash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func getAzMutex(tag string) *sync.Mutex {
	index := strHash(tag) % mutexBucketSize
	return azMutex[index]
}

type GatewayApiPolicyServiceImpl struct {
	azDb            service.GatewayAzInfoService
	kongDb          service.GatewayKongInfoService
	ingressPolicyDb service.GatewayIngressPolicyService
	kongPolicyDb    service.GatewayPolicyService
	packageDb       service.GatewayPackageService
	packageApiDb    service.GatewayPackageApiService
	routeDB         service.GatewayRouteService
	defaultPolicyDb service.GatewayDefaultPolicyService
	openapiRuleBiz  *openapi_rule.GatewayOpenapiRuleService
	zoneBiz         *zone.GatewayZoneService
	zoneDb          service.GatewayZoneService
	globalBiz       *global.GatewayGlobalService
	engine          *orm.OrmEngine
	domainBiz       *domain.GatewayDomainService
	packageBiz      *endpoint_api.GatewayOpenapiService
	reqCtx          context.Context
	runtimeDb       service.GatewayRuntimeServiceService
}

var once sync.Once

func NewGatewayApiPolicyServiceImpl() error {
	once.Do(
		func() {
			azDb, _ := service.NewGatewayAzInfoServiceImpl()
			ingressPolicyDb, _ := service.NewGatewayIngressPolicyServiceImpl()
			kongPolicyDb, _ := service.NewGatewayPolicyServiceImpl()
			defaultPolicyDb, _ := service.NewGatewayDefaultPolicyServiceImpl()
			kongDb, _ := service.NewGatewayKongInfoServiceImpl()
			zoneDb, _ := service.NewGatewayZoneServiceImpl()
			packageDb, _ := service.NewGatewayPackageServiceImpl()
			packageApiDb, _ := service.NewGatewayPackageApiServiceImpl()
			routeDB, _ := service.NewGatewayRouteServiceImpl()
			engine, _ := orm.GetSingleton()
			runtimeDb, _ := service.NewGatewayRuntimeServiceServiceImpl()
			api_policy.Service = &GatewayApiPolicyServiceImpl{
				azDb:            azDb,
				kongDb:          kongDb,
				ingressPolicyDb: ingressPolicyDb,
				kongPolicyDb:    kongPolicyDb,
				packageDb:       packageDb,
				packageApiDb:    packageApiDb,
				routeDB:         routeDB,
				defaultPolicyDb: defaultPolicyDb,
				openapiRuleBiz:  &openapi_rule.Service,
				zoneBiz:         &zone.Service,
				zoneDb:          zoneDb,
				globalBiz:       &global.Service,
				engine:          engine,
				domainBiz:       &domain.Service,
				packageBiz:      &endpoint_api.Service,
				runtimeDb:       runtimeDb,
			}
		})
	return nil
}

func (impl GatewayApiPolicyServiceImpl) Clone(ctx context.Context) api_policy.GatewayApiPolicyService {
	newService := impl
	newService.reqCtx = ctx
	return &newService
}

func (impl GatewayApiPolicyServiceImpl) GetPolicyConfig(category, packageId, packageApiId string) (result interface{}, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("error happened, err:%+v", err)
		}
	}()
	var gatewayProvider string
	var policyEngine apipolicy.PolicyEngine
	var projectName string
	var api *orm.GatewayPackageApi
	if category == "" || packageId == "" {
		// keep compatible
		result = map[string]interface{}{
			"category":    "",
			"description": "",
			"policyList":  nil,
		}
		return
	}
	pack, err := impl.packageDb.Get(packageId)
	if err != nil {
		return
	}

	gatewayProvider, err = impl.GetGatewayProvider(pack.DiceClusterName)
	if err != nil {
		logrus.Errorf("get gateway provider failed for cluster %s: %v\n", pack.DiceClusterName, err)
		return
	}

	switch gatewayProvider {
	case mseCommon.MseProviderName:
		if category != apipolicy.Policy_Engine_Service_Guard && category != apipolicy.Policy_Engine_CORS && category != apipolicy.Policy_Engine_IP {
			//TODO: 不同的 gateway provider 的插件配置情况是否需要调整,还是旧使用以前默认的 Kong 的方式，对于 mse 的情况返回结果一样，但有些插件不生效?
			logrus.Warnf("gateway provider %s no policy config for policy %s\n", gatewayProvider, category)
		}
	case "":
	default:
		err = errors.Errorf("unknown gateway provider:%v\n", gatewayProvider)
		return
	}
	policyEngine, err = apipolicy.GetPolicyEngine(category)
	if err != nil {
		return
	}
	info := dto.DiceInfo{
		ProjectId: pack.DiceProjectId,
	}
	projectName, err = (*impl.globalBiz).GetProjectName(info)
	if err != nil || projectName == "" {
		err = errors.Errorf("get projectName failed, info:%+v, err:%+v", info, err)
		return
	}
	serviceInfo := apipolicy.ServiceInfo{
		ProjectName: projectName,
		Env:         pack.DiceProjectId,
	}
	ctx := map[string]interface{}{}
	ctx[apipolicy.CTX_SERVICE_INFO] = serviceInfo
	if packageApiId == "" {
		var dto interface{}
		dto, err = policyEngine.GetConfig(gatewayProvider, category, packageId, nil, ctx)
		if err != nil {
			return
		}
		result = dto
		return
	}
	api, err = impl.packageApiDb.Get(packageApiId)
	if err != nil {
		return
	}
	if api == nil {
		return nil, errors.Errorf("package api not found, packageID: %s, apiID: %s", packageId, packageApiId)
	}
	var zone *orm.GatewayZone
	if api.ZoneId != "" {
		zone, err = (*impl.zoneBiz).GetZone(api.ZoneId)
		if err != nil || zone == nil {
			err = errors.Errorf("get zone failed, info:%+v, err:%+v", info, err)
			return
		}
	}
	var dto interface{}
	dto, err = policyEngine.GetConfig(gatewayProvider, category, packageId, zone, ctx)
	if err != nil {
		return
	}
	result = dto
	return
}

func (impl GatewayApiPolicyServiceImpl) RefreshZoneIngress(zone orm.GatewayZone, az orm.GatewayAzInfo, namespace string, useKong bool) error {
	exist, err := (*impl.zoneBiz).GetZone(zone.Id)
	if err != nil {
		return err
	}
	if exist == nil {
		logrus.Infof("zone not exist, maybe session rollback, id:%s", zone.Id)
		return nil
	}
	if exist.Type != service.ZONE_TYPE_PACKAGE_API && !useKong {
		logrus.Infof("zone type is %s, and gateway provider is MSE, no need refresh ingress for zone: %+v", exist.Type, zone)
		return nil
	}
	var zoneRegions []string
	zoneRegions = append(zoneRegions, service.ZONE_REGIONS...)
	zoneRegions = append(zoneRegions, service.GLOBAL_REGIONS...)

	changes, err := impl.ingressPolicyDb.GetChangesByRegions(az.Az, strings.Join(zoneRegions, "|"), zone.Id)
	if err != nil {
		return err
	}
	k8sAdapter, err := k8s.NewAdapter(az.Az)
	if err != nil {
		return err
	}
	if useKong {
		namespace, _, err = impl.kongDb.GetK8SInfo(&orm.GatewayKongInfo{
			ProjectId: zone.DiceProjectId,
			Az:        zone.DiceClusterName,
			Env:       zone.DiceEnv,
		})
		if err != nil {
			return err
		}
	}
	if namespace == "" {
		runtimes, err := impl.runtimeDb.SelectByAny(&orm.GatewayRuntimeService{
			ProjectId:   zone.DiceProjectId,
			Workspace:   zone.DiceEnv,
			ClusterName: az.Az,
		})
		if err != nil {
			logrus.Errorf("RefreshZoneIngress try to get namespace form tb_gateway_runtime_service failed: %v", err)
			return errors.Errorf("RefreshZoneIngress try to get namespace form tb_gateway_runtime_service failed: %v", err)
		}

		if len(runtimes) == 0 {
			logrus.Errorf("RefreshZoneIngress no records found from tb_gateway_runtime_service")
			return errors.Errorf("RefreshZoneIngress no records found from tb_gateway_runtime_service")
		} else {
			namespace = runtimes[0].ProjectNamespace
		}
	}
	if namespace == "" {
		// 走到这里表示 是按 url 转发，因此，namespace 是固定的
		namespace = "project-" + zone.DiceProjectId + "-" + strings.ToLower(zone.DiceEnv)
	}
	err = impl.deployIngressChanges(k8sAdapter, namespace, useKong, zone, *changes)
	if err != nil {
		logrus.Errorf("deployIngressChanges in RefreshZoneIngress for zone=%+v with useKong=%v namespace=%s error: %v\n", zone, useKong, namespace, err)
		return err
	}
	return nil
}

func (impl GatewayApiPolicyServiceImpl) patchGlobalHttpSnippet(k8sAdapter k8s.K8SAdapter, httpSnippet *[]string) error {
	count, err := k8sAdapter.CountIngressController()
	if err != nil {
		count = 1
		logrus.Errorf("get ingress controller count failed, err:%+v", err)
	}
	qps := int64(math.Ceil(float64(config.ServerConf.OfflineQps) / float64(count)))
	limitReqZone := fmt.Sprintf(`limit_req_zone 1 zone=offline-limit:1m rate=%dr/s;
map $http_referer                   $cors_use_referer {
    ~(^https?://[^/]*)/.*           $1;
}
map $http_origin-$cors_use_referer $from_request_origin_or_referer {
    "-"                             "*";
    "~^-"                           $cors_use_referer;
    default                         $http_origin;
}
`, qps)
	*httpSnippet = append(*httpSnippet, limitReqZone)
	return nil
}

func (impl GatewayApiPolicyServiceImpl) deployControllerChanges(k8sAdapter k8s.K8SAdapter, changes service.IngressChanges) error {
	configmap := map[string]*string{}
	for _, options := range changes.ConfigmapOptions {
		for key, value := range options {
			if old, exist := configmap[key]; exist {
				logrus.Debugf("invalid changes:%+v", changes)
				return errors.Errorf("config map key duplicated, key:%s, value:%+v, new value:%+v",
					key, old, value)
			}
			configmap[key] = value
		}
	}
	luaDictConfig := "configuration_data: 100"
	configmap["lua-shared-dicts"] = &luaDictConfig
	emptyStr := ""
	var mainSnippet *string
	if changes.MainSnippets != nil {
		if len(*changes.MainSnippets) == 0 {
			mainSnippet = &emptyStr
		} else {
			mergeStr := strings.Join(*changes.MainSnippets, "\n###HEPA-AUTO-CONFIG###\n")
			mainSnippet = &mergeStr
		}
	}
	var httpSnippet *string
	if changes.HttpSnippets != nil {
		err := impl.patchGlobalHttpSnippet(k8sAdapter, changes.HttpSnippets)
		if err != nil {
			return err
		}
		mergeStr := strings.Join(*changes.HttpSnippets, "\n###HEPA-AUTO-CONFIG###\n")
		httpSnippet = &mergeStr
	}
	var serverSnippet *string
	if changes.ServerSnippets != nil {
		if len(*changes.ServerSnippets) == 0 {
			serverSnippet = &emptyStr
		} else {
			mergeStr := strings.Join(*changes.ServerSnippets, "\n###HEPA-AUTO-CONFIG###\n")
			serverSnippet = &mergeStr
		}
	}
	if len(configmap) > 0 || mainSnippet != nil || httpSnippet != nil || serverSnippet != nil {
		err := k8sAdapter.UpdateIngressController(configmap, mainSnippet, httpSnippet, serverSnippet)
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl GatewayApiPolicyServiceImpl) getAnnotationChanges(changes service.IngressChanges) (map[string]*string, *string, error) {
	emptyStr := ""
	annotation := map[string]*string{}
	for _, anno := range changes.Annotations {
		for key, value := range anno {
			if old, exist := annotation[key]; exist {
				logrus.Debugf("invalid changes:%+v", changes)
				return nil, nil, errors.Errorf("annotation key duplicated, key:%s, value:%+v, new value:%+v",
					key, old, value)
			}
			annotation[key] = value
		}
	}
	var locationSnippet *string
	if changes.LocationSnippets != nil {
		if len(*changes.LocationSnippets) == 0 {
			locationSnippet = &emptyStr
		} else {
			mergeStr := strings.Join(*changes.LocationSnippets, "\n###HEPA-AUTO-CONFIG###\n")
			locationSnippet = &mergeStr
		}

	}
	return annotation, locationSnippet, nil
}

func (impl GatewayApiPolicyServiceImpl) deployAnnotationChanges(k8sAdapter k8s.K8SAdapter, zone orm.GatewayZone, changes service.IngressChanges, namespace string) error {
	annotation, locationSnippet, err := impl.getAnnotationChanges(changes)
	if err != nil {
		return err
	}
	if len(annotation) > 0 || locationSnippet != nil {
		err = k8sAdapter.UpdateIngressAnnotation(namespace, zone.Name, annotation, locationSnippet)
		if err != nil {
			cause := err.Error()
			lines := strings.Split(cause, "\n")
			for i := len(lines) - 1; i >= 0; i-- {
				// move the useless warn info of nginx
				if strings.Contains(lines[i], "[warn]") {
					lines = append(lines[:i], lines[i+1:]...)
				}
			}
			return errors.New(strings.Join(lines, "\n"))
		}
	}
	return nil
}

func (impl GatewayApiPolicyServiceImpl) deployIngressChanges(k8sAdapter k8s.K8SAdapter, namespace string, useKong bool, zone orm.GatewayZone, changes service.IngressChanges, annotationReset ...bool) error {
	logrus.Debugf("deploy ingress zone:%+v, changes:%+v", zone, changes)
	var err error
	if len(annotationReset) > 0 && annotationReset[0] {
		err = impl.deployAnnotationChanges(k8sAdapter, zone, changes, namespace)
		if err != nil {
			return err
		}
		if useKong {
			err = impl.deployControllerChanges(k8sAdapter, changes)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if useKong {
		err = impl.deployControllerChanges(k8sAdapter, changes)
		if err != nil {
			return err
		}
	}
	err = impl.deployAnnotationChanges(k8sAdapter, zone, changes, namespace)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayApiPolicyServiceImpl) setIngressPolicyDaoConfig(policyDao *orm.GatewayIngressPolicy, policyConfig apipolicy.PolicyConfig) error {
	annotation := policyConfig.IngressAnnotation
	controller := policyConfig.IngressController
	var regions []string
	if annotation != nil {
		if annotation.Annotation != nil {
			regions = append(regions, "annotation")
			annoByte, err := json.Marshal(annotation.Annotation)
			if err != nil {
				return errors.WithStack(err)
			}
			policyDao.Annotations = annoByte
		}
		if annotation.LocationSnippet != nil {
			regions = append(regions, "location")
			policyDao.LocationSnippet = []byte(*annotation.LocationSnippet)
		}
	}
	if controller != nil {
		if controller.ConfigOption != nil {
			regions = append(regions, "option")
			optionByte, err := json.Marshal(controller.ConfigOption)
			if err != nil {
				return errors.WithStack(err)
			}
			policyDao.ConfigmapOption = optionByte
		}
		if controller.MainSnippet != nil {
			regions = append(regions, "main")
			policyDao.MainSnippet = []byte(*controller.MainSnippet)
		}
		if controller.HttpSnippet != nil {
			regions = append(regions, "http")
			policyDao.HttpSnippet = []byte(*controller.HttpSnippet)
		}
		if controller.ServerSnippet != nil {
			regions = append(regions, "server")
			policyDao.ServerSnippet = []byte(*controller.ServerSnippet)
		}
	}
	policyDao.Regions = strings.Join(regions, "|")
	return nil
}

func (impl GatewayApiPolicyServiceImpl) executePolicyEngine(zone *orm.GatewayZone, runtimeService *orm.GatewayRuntimeService, category string, engine apipolicy.PolicyEngine, config []byte, dto apipolicy.PolicyDto, ctx map[string]interface{}, policyService service.GatewayIngressPolicyService, k8sAdapter k8s.K8SAdapter, helper *service.SessionHelper, needDeployTag ...bool) error {
	var apiService service.GatewayPackageApiService
	var packService service.GatewayPackageService
	var kongService service.GatewayKongInfoService
	var err error
	if helper != nil {
		apiService, err = impl.packageApiDb.NewSession(helper)
		if err != nil {
			return err
		}
		packService, err = impl.packageDb.NewSession(helper)
		if err != nil {
			return err
		}
		kongService, err = impl.kongDb.NewSession(helper)
		if err != nil {
			return err
		}
	} else {
		apiService = impl.packageApiDb
		packService = impl.packageDb
		kongService = impl.kongDb
	}
	needDeployIngress := true
	if len(needDeployTag) > 0 {
		needDeployIngress = needDeployTag[0]
	}
	policyConfig, err := engine.ParseConfig(dto, ctx, false)
	if err != nil {
		return err
	}
	policyDao := &orm.GatewayIngressPolicy{
		Name:   category,
		Az:     zone.DiceClusterName,
		ZoneId: zone.Id,
	}
	err = impl.setIngressPolicyDaoConfig(policyDao, policyConfig)
	if err != nil {
		return err
	}
	err = policyService.CreateOrUpdate(policyDao)
	if err != nil {
		return err
	}

	namespace := ""
	useKong := true
	if _, ok := ctx[apipolicy.CTX_KONG_ADAPTER]; !ok {
		useKong = false
	}
	if useKong {
		namespace, _, err = kongService.GetK8SInfo(&orm.GatewayKongInfo{
			ProjectId: zone.DiceProjectId,
			Az:        zone.DiceClusterName,
			Env:       zone.DiceEnv,
		})
		if err != nil {
			return err
		}
	} else {
		if runtimeService != nil && runtimeService.ProjectNamespace != "" {
			namespace = runtimeService.ProjectNamespace
		} else {
			// 走到这里表示 是按 url 转发，因此，namespace 是固定的
			namespace = "project-" + zone.DiceProjectId + "-" + strings.ToLower(zone.DiceEnv)
		}
	}

	if needDeployIngress {
		changes, err := policyService.GetChangesByRegionsImpl(zone.DiceClusterName, policyDao.Regions, zone.Id)
		if err != nil {
			return err
		}
		err = impl.deployIngressChanges(k8sAdapter, namespace, useKong, *zone, *changes, policyConfig.AnnotationReset)
		if err != nil {
			return err
		}
	}
	if policyConfig.KongPolicyChange {
		if zone.Type == service.ZONE_TYPE_PACKAGE_API {
			api, err := apiService.GetByAny(&orm.GatewayPackageApi{
				ZoneId: zone.Id,
			})
			if err != nil {
				return err
			}
			if api != nil {
				err = (*impl.openapiRuleBiz).SetPackageApiKongPolicies(api, useKong, helper)
				if err != nil {
					return err
				}
			}
		} else if zone.Type == service.ZONE_TYPE_UNITY {
			pack, err := packService.GetByAny(&orm.GatewayPackage{
				ZoneId: zone.Id,
			})
			if err != nil {
				return err
			}
			if pack != nil {
				err = (*impl.openapiRuleBiz).SetPackageKongPolicies(pack, helper)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (impl GatewayApiPolicyServiceImpl) GetGatewayProvider(clusterName string) (string, error) {
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

func (impl GatewayApiPolicyServiceImpl) SetZonePolicyConfig(zone *orm.GatewayZone, runtimeService *orm.GatewayRuntimeService, category string, config []byte, helper *service.SessionHelper, needDeployTag ...bool) (apipolicy.PolicyDto, string, error) {
	needDeployIngress := true
	if len(needDeployTag) > 0 {
		needDeployIngress = needDeployTag[0]
	}
	var policyService service.GatewayIngressPolicyService
	var kongService service.GatewayKongInfoService
	var err error
	if helper != nil {
		policyService, err = impl.ingressPolicyDb.NewSession(helper)
		if err != nil {
			return nil, "", err
		}
		kongService, err = impl.kongDb.NewSession(helper)
		if err != nil {
			return nil, "", err
		}
	} else {
		policyService = impl.ingressPolicyDb
		kongService = impl.kongDb
	}

	// 部署ingress需要加锁
	if needDeployIngress {
		mutex := getAzMutex(zone.DiceClusterName)
		mutex.Lock()
		defer mutex.Unlock()
	}
	gatewayProvider, err := impl.GetGatewayProvider(zone.DiceClusterName)
	if err != nil {
		return nil, "", err
	}

	policyEngine, err := apipolicy.GetPolicyEngine(category)
	if err != nil {
		return nil, "", err
	}
	dto, err, msg := policyEngine.UnmarshalConfig(config, gatewayProvider)
	if err != nil {
		return nil, msg, err
	}
	k8sAdapter, err := k8s.NewAdapter(zone.DiceClusterName)
	if err != nil {
		return nil, "", err
	}
	kongInfo, err := kongService.GetKongInfo(&orm.GatewayKongInfo{
		Az:        zone.DiceClusterName,
		ProjectId: zone.DiceProjectId,
		Env:       zone.DiceEnv,
	})
	if err != nil {
		return nil, "", err
	}

	ctx := map[string]interface{}{
		apipolicy.CTX_K8S_CLIENT: k8sAdapter,
		apipolicy.CTX_IDENTIFY:   zone.Name,
		apipolicy.CTX_ZONE:       zone,
	}
	switch gatewayProvider {
	case mseCommon.MseProviderName:
		gatewayAdapter, err := mse.NewMseAdapter(zone.DiceClusterName)
		if err != nil {
			return nil, "", errors.Errorf("init mse gateway adpter failed:%v\n", err)
		}
		ctx[apipolicy.CTX_MSE_ADAPTER] = gatewayAdapter
	case "":
		gatewayAdapter := kong.NewKongAdapter(kongInfo.KongAddr)
		ctx[apipolicy.CTX_KONG_ADAPTER] = gatewayAdapter
	default:
		logrus.Errorf("unknown gateway provider:%v\n", gatewayProvider)
		return nil, "", errors.Errorf("unknown gateway provider:%v\n", gatewayProvider)
	}

	err = impl.executePolicyEngine(zone, runtimeService, category, policyEngine, config, dto, ctx, policyService, k8sAdapter, helper, needDeployTag...)
	if err != nil {
		return nil, fmt.Sprintf("执行策略失败, 失败原因:\n%s", errors.Cause(err)), err
	}
	if category != apipolicy.Policy_Engine_Built_in {
		builtinEngine, err := apipolicy.GetPolicyEngine(apipolicy.Policy_Engine_Built_in)
		if err != nil {
			return nil, "", err
		}
		err = impl.executePolicyEngine(zone, runtimeService, apipolicy.Policy_Engine_Built_in, builtinEngine, nil, nil, ctx, policyService, k8sAdapter, helper, needDeployTag...)
		if err != nil {
			return nil, fmt.Sprintf("更新内置策略失败, 失败原因:\n%s", errors.Cause(err)), err
		}
	}
	return dto, "", nil
}

func (impl GatewayApiPolicyServiceImpl) SetZoneDefaultPolicyConfig(packageId string, runtimeService *orm.GatewayRuntimeService, zone *orm.GatewayZone, az *orm.GatewayAzInfo, helper ...*service.SessionHelper) (map[string]*string, *string, *service.SessionHelper, error) {
	mutex := getAzMutex(az.Az)
	mutex.Lock()
	defer mutex.Unlock()
	var session *service.SessionHelper
	var err error
	if len(helper) > 0 {
		session = helper[0]
	} else {
		session, err = service.NewSessionHelper()
		if err != nil {
			return nil, nil, nil, err
		}
	}
	policies, err := impl.defaultPolicyDb.SelectByAny(&orm.GatewayDefaultPolicy{
		Level:     orm.POLICY_PACKAGE_LEVEL,
		PackageId: packageId,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	for _, policy := range policies {
		_, _, err = impl.SetZonePolicyConfig(zone, runtimeService, policy.Name, policy.Config, session, false)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	if len(policies) == 0 {
		_, _, err = impl.SetZonePolicyConfig(zone, runtimeService, apipolicy.Policy_Engine_Built_in, nil, session, false)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	policyService, err := impl.ingressPolicyDb.NewSession(session)
	if err != nil {
		return nil, nil, nil, err
	}

	zoneChange, err := policyService.GetChangesByRegions(az.Az, strings.Join(service.ZONE_REGIONS, "|"), zone.Id)
	if err != nil {
		return nil, nil, nil, err
	}

	useKong := true
	_, azInfo, err := impl.azDb.GetAzInfoByClusterName(az.Az)
	if err != nil {
		logrus.Errorf("get ClusterInfoDto for cluster %s failed: %v", az.Az, err)
		return nil, nil, nil, err
	}

	if azInfo != nil && azInfo.GatewayProvider != "" {
		useKong = false
	}

	if useKong {
		k8sAdapter, err := k8s.NewAdapter(az.Az)
		if err != nil {
			return nil, nil, nil, err
		}

		controllerChange, err := policyService.GetChangesByRegions(az.Az, strings.Join(service.GLOBAL_REGIONS, "|"))
		if err != nil {
			return nil, nil, nil, err
		}

		err = impl.deployControllerChanges(k8sAdapter, *controllerChange)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	annotations, locationSnippet, err := impl.getAnnotationChanges(*zoneChange)
	if err != nil {
		return nil, nil, nil, err
	}
	return annotations, locationSnippet, session, nil
}

func doRecover() {
	if r := recover(); r != nil {
		logrus.Errorf("recovered from: %+v ", r)
		debug.PrintStack()
	}
}

func (impl GatewayApiPolicyServiceImpl) SetPackageDefaultPolicyConfig(category, packageId string, az *orm.GatewayAzInfo, config []byte, helperOption ...*service.SessionHelper) (string, error) {
	mutex := getAzMutex(az.Az)
	mutex.Lock()
	defer mutex.Unlock()
	var err error
	transSucc := false
	useKongGateWay := true
	var helper *service.SessionHelper
	if len(helperOption) == 0 {
		helper, err = service.NewSessionHelper()
		if err != nil {
			return "", err
		}
		defer func() {
			if transSucc {
				_ = helper.Commit()
			} else {
				_ = helper.Rollback()
			}
			helper.Close()
		}()
	} else {
		helper = helperOption[0]
	}
	var zones []orm.GatewayZone
	defer func() {
		if !transSucc {
			_ = helper.Rollback()
			wg := sync.WaitGroup{}
			for _, zone := range zones {
				wg.Add(1)
				go func(z orm.GatewayZone) {
					defer doRecover()
					err = impl.RefreshZoneIngress(z, *az, "", useKongGateWay)
					if err != nil {
						logrus.Errorf("refresh zone ingress failed, %+v", err)
					}
					wg.Done()
				}(zone)
			}
			wg.Wait()
			logrus.Info("zone ingress rollback done")
		}
	}()
	packageDb, err := impl.packageDb.NewSession(helper)
	if err != nil {
		return "", err
	}
	pack, err := packageDb.Get(packageId)
	if err != nil {
		return "", err
	}

	gatewayProvider, err := impl.GetGatewayProvider(pack.DiceClusterName)
	if err != nil {
		return "", err
	}

	if gatewayProvider == mseCommon.MseProviderName {
		useKongGateWay = false
	}

	if pack.Scene == orm.UnityScene || pack.Scene == orm.HubScene {
		zone, err := (*impl.zoneBiz).GetZone(pack.ZoneId, helper)
		if err != nil {
			return "", err
		}
		if zone == nil {
			zone, err = (*impl.packageBiz).CreateUnityPackageZone(packageId, helper)
			if err != nil {
				return "", err
			}
		}
		zones = append(zones, *zone)
	}
	packageApiDb, err := impl.packageApiDb.NewSession(helper)
	if err != nil {
		return "", err
	}
	apis, err := packageApiDb.SelectByAny(&orm.GatewayPackageApi{
		PackageId: packageId,
	})
	if err != nil {
		return "", err
	}
	for _, api := range apis {
		if api.ZoneId != "" {
			zone, err := (*impl.zoneBiz).GetZone(api.ZoneId, helper)
			if err != nil {
				return "", err
			}
			if zone != nil {
				zones = append(zones, *zone)
			}
		}
	}
	policyEngine, err := apipolicy.GetPolicyEngine(category)
	if err != nil {
		return "", err
	}
	dto, err, msg := policyEngine.UnmarshalConfig(config, gatewayProvider)
	if err != nil {
		return msg, err
	}
	needResetAnnotaion := policyEngine.NeedResetAnnotation(dto)
	policyService, err := impl.ingressPolicyDb.NewSession(helper)
	if err != nil {
		return "", err
	}
	for _, zone := range zones {
		logrus.Debugf("Set zone_id=%s  zone_name=%s for policy  %s", zone.Id, zone.Name, category)
		policy, err := policyService.GetByAny(&orm.GatewayIngressPolicy{
			Name:   category,
			ZoneId: zone.Id,
		})
		if err != nil {
			return "", err
		}
		if policy == nil || len(policy.Config) == 0 {
			_, msg, err = impl.SetZonePolicyConfig(&zone, nil, category, config, helper, false)
			if err != nil {
				return msg, err
			}
		}
	}
	controllerChanges, err := policyService.GetChangesByRegions(az.Az, strings.Join(service.GLOBAL_REGIONS, "|"))
	if err != nil {
		return "", err
	}
	var zoneChanges []service.IngressChanges
	for _, zone := range zones {
		zoneChange, err := policyService.GetChangesByRegions(az.Az, strings.Join(service.ZONE_REGIONS, "|"), zone.Id)
		if err != nil {
			return "", err
		}
		zoneChanges = append(zoneChanges, *zoneChange)
	}
	k8sAdapter, err := k8s.NewAdapter(az.Az)
	if err != nil {
		return "", err
	}
	kongDb, err := impl.kongDb.NewSession(helper)
	if err != nil {
		return "", err
	}
	namespace, _, err := kongDb.GetK8SInfo(&orm.GatewayKongInfo{
		ProjectId: pack.DiceProjectId,
		Env:       pack.DiceEnv,
		Az:        pack.DiceClusterName,
	})
	if gatewayProvider == mseCommon.MseProviderName {
		//使用 MSE 网关，ingress 的 namespace 获取方式不一样
		runtimes, err := impl.runtimeDb.SelectByAny(&orm.GatewayRuntimeService{
			ProjectId:   pack.DiceProjectId,
			Workspace:   pack.DiceEnv,
			ClusterName: az.Az,
		})
		if err != nil {
			logrus.Warnf("RefreshZoneIngress try to get namespace form tb_gateway_runtime_service failed: %v", err)
		}

		if len(runtimes) > 0 {
			namespace = runtimes[0].ProjectNamespace
		}

		if namespace == "" {
			// 走到这里表示 是按 url 转发，因此，namespace 是固定的
			namespace = "project-" + pack.DiceProjectId + "-" + strings.ToLower(pack.DiceEnv)
		}
	}
	if err != nil {
		return "", err
	}
	deployZoneFunc := func() error {
		wg := sync.WaitGroup{}
		var err error
		for i := 0; i < len(zones); i++ {
			if zones[i].Type != service.ZONE_TYPE_PACKAGE_API && !useKongGateWay {
				logrus.Infof("zone type is %s and gateway provider is MSE, no need refresh ingress for zone=%+v", zones[i].Type, zones[i])
				continue
			}
			wg.Add(1)
			go func(index int) {
				defer doRecover()
				gerr := impl.deployAnnotationChanges(k8sAdapter, zones[index], zoneChanges[index], namespace)
				if gerr != nil {
					err = gerr
				}
				wg.Done()
			}(i)
		}
		wg.Wait()
		return err
	}
	if needResetAnnotaion {
		err = deployZoneFunc()
		if err != nil {
			return "", err
		}
		if useKongGateWay {
			err = impl.deployControllerChanges(k8sAdapter, *controllerChanges)
			if err != nil {
				return "", err
			}
		}
	} else {
		if useKongGateWay {
			err = impl.deployControllerChanges(k8sAdapter, *controllerChanges)
			if err != nil {
				return "", err
			}
		}
		err = deployZoneFunc()
		if err != nil {
			return "", err
		}
	}
	err = impl.defaultPolicyDb.CreateOrUpdate(&orm.GatewayDefaultPolicy{
		Config:    config,
		Level:     orm.POLICY_PACKAGE_LEVEL,
		Name:      category,
		PackageId: packageId,
	})
	if err != nil {
		return "", err
	}
	transSucc = true
	return "", nil
}

func (impl GatewayApiPolicyServiceImpl) SetPolicyConfig(category, packageId, packageApiId string, config []byte) (result interface{}, rerr error) {
	auditCtx := map[string]interface{}{}
	var pack *orm.GatewayPackage
	var gatewayProvider string
	logrus.Infof("Set policy config for %s, packageId=%s, packageApiId=%s config=[%s]\n", category, packageId, packageApiId, string(config))
	defer func() {
		if rerr != nil {
			logrus.Errorf("error happened, err:%+v", rerr)
		}
		if pack == nil {
			return
		}
		engine, err := apipolicy.GetPolicyEngine(category)
		if err != nil {
			return
		}
		dto, err, _ := engine.UnmarshalConfig(config, gatewayProvider)
		if err != nil {
			return
		}
		if dto.Enable() {
			auditCtx["switch"] = "on"
		} else {
			auditCtx["switch"] = "off"
		}
		auditCtx["endpoint"] = pack.PackageName
		auditKey := apistructs.UpdateGlobalRoutePolicyTemplate
		if packageApiId != "" {
			auditKey = apistructs.UpdateRoutePolicyTemplate
		}
		auditCtx["policy"] = category
		audit := common.MakeAuditInfo(impl.reqCtx, common.ScopeInfo{
			ProjectId: pack.DiceProjectId,
			Workspace: pack.DiceEnv,
		}, auditKey, err, auditCtx)
		if audit != nil {
			err = bundle.Bundle.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: *audit})
			if err != nil {
				logrus.Errorf("create audit failed, err:%+v", err)
			}
		}
	}()
	if category == "" || packageId == "" {
		rerr = errors.New("invalid argument")
		return
	}

	useKong := true
	pack, rerr = impl.packageDb.Get(packageId)
	if rerr != nil {
		return
	}
	if pack == nil {
		rerr = errors.New("endpoint not found")
		return
	}

	gatewayProvider, rerr = impl.GetGatewayProvider(pack.DiceClusterName)
	if rerr != nil {
		logrus.Errorf("get gateway provider failed for cluster %s: %v\n", pack.DiceClusterName, rerr)
		return
	}

	switch gatewayProvider {
	case mseCommon.MseProviderName:
		useKong = false
		switch category {
		case apipolicy.Policy_Engine_Service_Guard:
		case apipolicy.Policy_Engine_CORS:
		case apipolicy.Policy_Engine_IP:
		case apipolicy.Policy_Engine_Proxy:
		case apipolicy.Policy_Engine_SBAC:
		case apipolicy.Policy_Engine_CSRF:
		default:
			rerr = errors.Errorf("gateway provider %s not support set policy %s", gatewayProvider, category)
			return
		}
	case "":
		useKong = true
	default:
		logrus.Errorf("unknown gateway provider:%v\n", gatewayProvider)
		rerr = errors.Errorf("unknown gateway provider:%v\n", gatewayProvider)
		return
	}

	az, _, rerr := impl.azDb.GetAzInfoByClusterName(pack.DiceClusterName)
	if rerr != nil {
		return
	}

	// 校验 custom 配置语法有效性
	if category == apipolicy.Policy_Engine_Custom {
		policyDto := &custom.PolicyDto{}
		rerr = json.Unmarshal(config, policyDto)
		if rerr != nil {
			return
		}
		rerr = validateCustomNginxConf(category, policyDto.Config)
		if rerr != nil {
			return
		}
	}

	if packageApiId == "" {
		var msg string
		msg, rerr = impl.SetPackageDefaultPolicyConfig(category, packageId, az, config)
		if rerr != nil {
			logrus.Errorf("set default policy config failed, err:%+v", rerr)
			if msg != "" {
				rerr = errors.Errorf("update config failed: %s", msg)
			}
			return
		}
		result = true
		return
	}
	api, err := impl.packageApiDb.Get(packageApiId)
	if err != nil {
		rerr = err
		return
	}
	if api == nil {
		rerr = errors.New("api not found")
		return
	}
	auditCtx["path"] = api.ApiPath
	auditCtx["method"] = api.Method
	var zone *orm.GatewayZone
	if api.ZoneId == "" {
		var domains []string
		domains, rerr = (*impl.domainBiz).GetPackageDomains(packageId)
		if rerr != nil {
			return
		}
		var zoneId string
		zoneId, rerr = (*impl.packageBiz).TouchPackageApiZone(endpoint_api.PackageApiInfo{api, domains, pack.DiceProjectId, pack.DiceEnv, pack.DiceClusterName, false})
		if rerr != nil {
			return
		}
		zone, rerr = (*impl.zoneBiz).GetZone(zoneId)
		if rerr != nil || zone == nil {
			rerr = errors.Errorf("get zone failed, err:%+v", rerr)
			return
		}
		api.ZoneId = zoneId
		rerr = impl.packageApiDb.Update(api)
		if rerr != nil {
			return
		}
	} else {
		zone, rerr = (*impl.zoneBiz).GetZone(api.ZoneId)
		if rerr != nil || zone == nil {
			rerr = errors.Errorf("get zone failed, api:%+v, err:%+v", api, rerr)
			return
		}
	}
	helper, err := service.NewSessionHelper()
	if err != nil {
		rerr = err
		return
	}
	if gatewayProvider != mseCommon.MseProviderName {
		err = impl.checkDuplicatedPolicyConfig(gatewayProvider, packageId, packageApiId, category, config, zone, helper)
		if err != nil {
			logrus.Errorf("has deplicated policy config for policy %s: %v\n", category, err)
			rerr = err
			return
		}
	}

	transSucc := false
	defer func() {
		if transSucc {
			_ = helper.Commit()
		} else {
			_ = helper.Rollback()
			err := impl.RefreshZoneIngress(*zone, *az, "", useKong)
			if err != nil {
				logrus.Errorf("refresh zone ingress failed, %+v", err)
			}
			logrus.Info("zone ingress rollback done")
		}
		helper.Close()
	}()

	dto, msg, err := impl.SetZonePolicyConfig(zone, nil, category, config, helper)
	if err != nil {
		logrus.Errorf("set zone policy config failed, err:%+v", err)
		if msg != "" {
			rerr = errors.Errorf("update config failed: %s", msg)
		} else {
			rerr = errors.Errorf("set zone policy config failed, err:%+v", err)
		}
		return
	}
	ingressPolicyService, err := impl.ingressPolicyDb.NewSession(helper)
	if err != nil {
		rerr = err
		return
	}
	if dto.IsGlobal() {
		config = []byte{}
	}
	rerr = ingressPolicyService.UpdatePartial(&orm.GatewayIngressPolicy{
		Az:     az.Az,
		ZoneId: zone.Id,
		Name:   category,
		Config: config,
	}, "config")
	if rerr != nil {
		return
	}
	transSucc = true
	result = dto
	return
}

// checkDuplicatedPolicyConfig 检查各种路由策略之间是否彼此有造成 nginx 无法重新加载的配置项
func (impl GatewayApiPolicyServiceImpl) checkDuplicatedPolicyConfig(gatewayProvider, packageId, packageApiId, category string, category_config []byte, zone *orm.GatewayZone, helper *service.SessionHelper) error {
	var kongService service.GatewayKongInfoService
	var err error
	if helper != nil {
		kongService, err = impl.kongDb.NewSession(helper)
		if err != nil {
			return err
		}
	} else {
		kongService = impl.kongDb
	}

	k8sAdapter, err := k8s.NewAdapter(zone.DiceClusterName)
	if err != nil {
		return err
	}
	kongInfo, err := kongService.GetKongInfo(&orm.GatewayKongInfo{
		Az:        zone.DiceClusterName,
		ProjectId: zone.DiceProjectId,
		Env:       zone.DiceEnv,
	})
	if err != nil {
		return err
	}

	ctx := map[string]interface{}{
		apipolicy.CTX_K8S_CLIENT: k8sAdapter,
		apipolicy.CTX_IDENTIFY:   zone.Name,
		apipolicy.CTX_ZONE:       zone,
	}

	switch gatewayProvider {
	case mseCommon.MseProviderName:
		gatewayAdapter, err := mse.NewMseAdapter(zone.DiceClusterName)
		if err != nil {
			return errors.Errorf("init mse gateway adpter failed:%v\n", err)
		}
		ctx[apipolicy.CTX_MSE_ADAPTER] = gatewayAdapter
	case "":
		gatewayAdapter := kong.NewKongAdapter(kongInfo.KongAddr)
		ctx[apipolicy.CTX_KONG_ADAPTER] = gatewayAdapter
	default:
		return errors.Errorf("unknown gateway provider:%v\n", gatewayProvider)
	}

	policyEngine, err := apipolicy.GetPolicyEngine(category)
	if err != nil {
		return errors.Errorf("apipolicy.GetPolicyEngine error:%v\n", err)
	}

	category_dto, err, _ := policyEngine.UnmarshalConfig(category_config, gatewayProvider)
	if err != nil {
		logrus.Errorf("policyEngine.UnmarshalConfig for %s error:%v\n", category, err)
		return errors.Errorf("policyEngine.UnmarshalConfig for %s error:%v\n", category, err)
	}

	policyConfig, err := policyEngine.ParseConfig(category_dto, ctx, true)
	if err != nil {
		logrus.Errorf("parseConfig error:%v\n", err)
		return errors.Errorf("parseConfig error:%v\n", err)
	}

	categories := make([]string, 0)

	switch category {
	case apipolicy.Policy_Engine_Custom:
		// custom 策略需要跟其他所有策略校验冲突 (csrf 和 sbac 无需检查，因为没有配置会更新到 nginx conf)
		categories = []string{apipolicy.Policy_Engine_Built_in, apipolicy.Policy_Engine_WAF, apipolicy.Policy_Engine_Service_Guard, apipolicy.Policy_Engine_CORS,
			apipolicy.Policy_Engine_IP, apipolicy.Policy_Engine_Proxy}
	default:
		categories = []string{apipolicy.Policy_Engine_Custom}
	}

	policyConfigs := make(map[string]apipolicy.PolicyConfig)
	for _, ca := range categories {
		if ca == category {
			// 主要是与其他插件策略对比冲突，因此忽略自己当前的设置
			continue
		}
		if ca != apipolicy.Policy_Engine_Built_in {
			dto, err := impl.GetPolicyConfig(ca, packageId, packageApiId)
			if err != nil {
				return errors.Errorf("GetPolicyConfig for policy %s error:%v\n", ca, err)
			}

			sDto, ok := dto.(apipolicy.PolicyDto)
			if !ok {
				return errors.Errorf("parse policy %s config failed: GetPolicyConfig() return invalid config %v\n", ca, dto)
			}

			pEngine, err := apipolicy.GetPolicyEngine(ca)
			if err != nil {
				logrus.Errorf("GetPolicyEngine for %s error:%v\n", ca, err)
				return err
			}

			pc, err := pEngine.ParseConfig(sDto, ctx, true)
			if err != nil {
				logrus.Errorf("parseConfig for policy %s error:%v\n", ca, err)
				return errors.Errorf("parseConfig for policy %s error:%v\n", ca, err)
			}
			policyConfigs[ca] = pc
		} else {
			var sDto apipolicy.PolicyDto
			pEngine, err := apipolicy.GetPolicyEngine(ca)
			if err != nil {
				return errors.Errorf("GetPolicyEngine for policy %s error:%v\n", ca, err)
			}

			pc, err := pEngine.ParseConfig(sDto, ctx, true)
			if err != nil {
				logrus.Errorf("parseConfig for policy %s error:%v\n", ca, err)
				return err
			}
			policyConfigs[ca] = pc
		}
	}

	for ca, pc := range policyConfigs {
		logrus.Infof("check duplicated config for policy: p1=%s   p2=%s\n", category, ca)
		duplicated, err := hasDuplicatedConfig(policyConfig, pc)
		if err != nil {
			return errors.Errorf("check duplicated config for policy [%s, %s] error:%v\n", category, ca, err)
		}
		if duplicated != "" {
			logrus.Errorf("duplicated config [%s] have set in policy %s\n", duplicated, ca)
			return errors.Errorf("duplicated config [%s] have set in policy %s\n", duplicated, ca)
		}
	}
	return nil
}

// 判断两个策略配置是否有冲突，如果有，则返回发现的第一个冲突的配置项
func hasDuplicatedConfig(p1, p2 apipolicy.PolicyConfig) (string, error) {
	if p2.IngressAnnotation == nil && p2.IngressController == nil {
		return "", nil
	}
	p1_Nginx, err := getNginxConfFromPolicyConfig(p1)
	if err != nil {
		return "", err
	}
	p2_Nginx, err := getNginxConfFromPolicyConfig(p2)
	if err != nil {
		return "", err
	}

	for k := range p1_Nginx {
		if val, ok := p2_Nginx[k]; ok {
			// more_set_headers、proxy_set_header、set、limit_req、limit_conn、error_page、deny、allow、return 等允许多次设置
			if !skipKeys[k] {
				return k + fmt.Sprintf("=%s", val), nil
			}
		}
	}
	return "", nil
}

// 提取单个策略对应的 key/value 格式配置用于判断两个策略的 key 是否冲突
func getNginxConfFromPolicyConfig(p apipolicy.PolicyConfig) (map[string]string, error) {
	ret := make(map[string]string)

	if p.IngressAnnotation != nil {
		for k, v := range p.IngressAnnotation.Annotation {
			keys := strings.Split(k, "/")
			logrus.Infof("p.IngressAnnotation.Annotation k=%s    keys=%v\n   p.IngressAnnotation.Annotation[%s]=%v", k, keys, k, p.IngressAnnotation.Annotation[k])
			l := len(keys)

			key := strings.ReplaceAll(keys[l-1], "-", "_")
			if _, ok := ret[key]; ok {
				// more_set_headers、proxy_set_header、set、limit_req、limit_conn、error_page、deny、allow、return 等允许多次设置
				if !skipKeys[key] {
					return ret, errors.Errorf("Annotation nginx conf %s duplicated", key)
				}
			}
			if v == nil {
				ret[key] = ""
			} else {
				ret[key] = *v
			}
		}

		if p.IngressAnnotation.LocationSnippet != nil {
			src := *(p.IngressAnnotation.LocationSnippet)
			err := extractConfigFromString("p.IngressAnnotation.LocationSnippet", src, ret)
			if err != nil {
				return ret, err
			}
		}
	}

	return ret, nil
}

func extractConfigFromString(kind, src string, ret map[string]string) error {
	if src != "" {
		kvs := strings.Split(trimAllPrefixShift(src), "\n")
		for idx, v := range kvs {
			v = strings.TrimSpace(v)
			if hasSpecialSubstr(v) {
				continue
			}
			logrus.Infof("%s:  index=%d  val=%v\n", kind, idx, v)
			kv := strings.Split(strings.TrimSuffix(strings.TrimSuffix(strings.TrimSpace(v), "\n"), ";"), " ")
			if len(kv) >= 2 {
				key := strings.ReplaceAll(kv[0], "-", "_")
				if _, ok := ret[key]; ok {
					// more_set_headers、proxy_set_header、set、limit_req、limit_conn、error_page、deny、allow、return 等允许多次设置
					if !skipKeys[key] {
						return errors.Errorf("%s nginx conf %s duplicated", kind, key)
					}
				}
				ret[key] = strings.Join(kv[1:], " ")
			}
		}
	}
	return nil
}

// 去掉 /n 前缀及前后空格
func trimAllPrefixShift(src string) string {
	src = strings.TrimSpace(src)
	for strings.HasPrefix(src, "\n") {
		src = strings.TrimPrefix(src, "\n")
		src = strings.TrimSpace(src)
	}
	return src
}

// 特殊语法，不校验冲突
func hasSpecialSubstr(v string) bool {
	return strings.HasPrefix(v, "lua_ingress.rewrite") ||
		strings.HasPrefix(v, "force_ssl_redirect") ||
		strings.HasPrefix(v, "use_port_in_redirects") ||
		strings.HasPrefix(v, "balancer.rewrite()") ||
		strings.HasPrefix(v, "plugins.run()") ||
		strings.HasPrefix(v, "monitor.call()") ||
		strings.HasPrefix(v, "balancer.log()") ||
		strings.HasPrefix(v, "location") ||
		strings.HasPrefix(v, "if") ||
		strings.HasPrefix(v, "{") ||
		strings.HasPrefix(v, "}") ||
		v == "\n"
}

// 校验 nginx 自定义 路由策略(custom) 的语义有效性
func validateCustomNginxConf(category, config string) error {
	if category != apipolicy.Policy_Engine_Custom {
		return nil
	}

	conf := fmt.Sprintf(`
load_module modules/ndk_http_module.so;
load_module modules/ngx_http_dav_ext_module.so;
load_module modules/ngx_http_geoip2_module.so;
load_module modules/ngx_http_image_filter_module.so;
load_module modules/ngx_http_subs_filter_module.so;
load_module modules/ngx_http_xslt_filter_module.so;
load_module modules/ngx_http_auth_pam_module.so;
load_module modules/ngx_http_echo_module.so;
load_module modules/ngx_http_geoip_module.so;
load_module modules/ngx_http_lua_module.so;
load_module modules/ngx_http_uploadprogress_module.so;
load_module modules/ngx_mail_module.so;
load_module modules/ngx_http_cache_purge_module.so;
load_module modules/ngx_http_fancyindex_module.so;
load_module modules/ngx_http_headers_more_filter_module.so;
load_module modules/ngx_http_perl_module.so;
load_module modules/ngx_http_upstream_fair_module.so;
load_module modules/ngx_stream_module.so;
events {
	accept_mutex on;   
	multi_accept on;  
	worker_connections  1024;
}

http {  
    keepalive_time 1h;  # nginx 1.19.10  引入
	server {
		listen 80;
		server_name example.com;
		root /var/www/example;
		keepalive_time 1h;  # nginx 1.19.10  引入

		location / {
			keepalive_time 1h; # nginx 1.19.10  引入
			%s
		}
	}
}
`, config)

	logrus.Infof("custom Nginx config:\n%s\n", conf)
	fileName := fmt.Sprintf("/tmp/nginx-%v", time.Now().UnixNano())
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return errors.Errorf("validate custom nginx config failed, open file %s failed,err:", err)
	}

	defer os.Remove(fileName)
	defer file.Close()

	_, err = file.WriteString(conf)
	if err != nil {
		return errors.Errorf("validate custom nginx config failed, Write file %s Error: %v\n", fileName, err)
	}

	cmd := exec.Command("nginx", "-t", "-c", fileName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Errorf("validate custom nginx config failed, validate output: %s\n with error: %v", string(out), err)
	}

	return nil
}
