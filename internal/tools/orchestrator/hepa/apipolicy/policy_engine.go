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

package apipolicy

import (
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	gateway_providers "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers"
	providerDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	db "github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
)

const (
	// ProviderMSE is micro-service-engine powered  by Aliyun
	ProviderMSE = "MSE"
	// ProviderNKE is nginx-kong-engine powered by Nginx and Kong, designed and composed by Erda
	ProviderNKE = "NKE"
)

const (
	CTX_IDENTIFY     = "id"
	CTX_K8S_CLIENT   = "k8s_client"
	CTX_KONG_ADAPTER = "kong_adapter"
	CTX_MSE_ADAPTER  = "mse_adapter"
	CTX_ZONE         = "zone"
	CTX_SERVICE_INFO = "service_info"

	Policy_Engine_Service_Guard = "safety-server-guard"
	Policy_Engine_Built_in      = "built-in"
	Policy_Engine_WAF           = "safety-waf"
	Policy_Engine_CORS          = "cors"
	Policy_Engine_Custom        = "custom"
	Policy_Engine_IP            = "safety-ip"
	Policy_Engine_Proxy         = "proxy"
	Policy_Engine_SBAC          = "sbac" // "sbac" is ServerBasedAccessControl
	Policy_Engine_CSRF          = "safety-csrf"

	Policy_Category_Basic   = "basic"
	Policy_Category_BuiltIn = "built-in"
	Policy_Category_Safety  = "safety"
	Policy_Category_Proxy   = "proxy"
	Policy_Category_Auth    = "auth"
)

var (
	_ PolicyEngineBasic = (*BasePolicy)(nil)
)

var (
	engines = struct{ MSE, NKE *sync.Map }{
		MSE: new(sync.Map),
		NKE: new(sync.Map),
	}
)

type PolicyEngine interface {
	PolicyEngineBasic
	PolicyEngineConfigurer
}

type PolicyEngineBasic interface {
	GetConfig(string, string, *orm.GatewayZone, map[string]interface{}) (PolicyDto, error)
	NeedResetAnnotation(PolicyDto) bool
	SetName(name string)
	GetName() string
	SetProvider(provider string)
	GetProvider() string
	NeedSerialUpdate() bool
}

type PolicyEngineConfigurer interface {
	MergeDiceConfig(map[string]interface{}) (PolicyDto, error)
	CreateDefaultConfig(map[string]interface{}) PolicyDto
	ParseConfig(PolicyDto, map[string]interface{}, bool) (PolicyConfig, error)
	UnmarshalConfig([]byte) (PolicyDto, error, string)
}

type PolicyDto interface {
	SetGlobal(bool)
	IsGlobal() bool
	Enable() bool
	SetEnable(bool)
}

type IngressAnnotation struct {
	Annotation      map[string]*string
	LocationSnippet *string
}

type IngressController struct {
	ConfigOption  map[string]*string
	MainSnippet   *string
	HttpSnippet   *string
	ServerSnippet *string
}

type PolicyConfig struct {
	KongPolicyChange  bool
	IngressAnnotation *IngressAnnotation
	IngressController *IngressController
	AnnotationReset   bool
}

type ServiceInfo struct {
	ProjectName string
	Env         string
}

type BaseDto struct {
	Switch bool `json:"switch"`
	Global bool `json:"global"`
}

func (dto *BaseDto) Enable() bool {
	return dto.Switch
}

func (dto *BaseDto) SetEnable(toggle bool) {
	dto.Switch = toggle
}

func (dto *BaseDto) SetGlobal(isGlobal bool) {
	dto.Global = isGlobal
}

func (dto *BaseDto) IsGlobal() bool {
	return dto.Global
}

type BasePolicy struct {
	Provider string
	Name     string
}

func (policy *BasePolicy) GetConfig(name, packageId string, zone *orm.GatewayZone, ctx map[string]interface{}) (PolicyDto, error) {
	engine, err := GetPolicyEngine(policy.Provider, name)
	if err != nil {
		return nil, err
	}
	policyDb, err := db.NewGatewayIngressPolicyServiceImpl()
	if err != nil {
		return nil, err
	}
	var policyConfig []byte
	useDefault := false
	if zone != nil {
		policyDao, err := policyDb.GetByAny(&orm.GatewayIngressPolicy{
			Name:   name,
			ZoneId: zone.Id,
			Az:     zone.DiceClusterName,
		})
		if err != nil {
			return nil, err
		}
		if policyDao != nil && len(policyDao.Config) > 0 {
			policyConfig = policyDao.Config
			if policy.Provider == "" {
				policy.Provider, _ = GetGatewayProvider(zone.DiceClusterName)
			}
			goto done
		}
	}
	{
		defaultPolicyDb, err := db.NewGatewayDefaultPolicyServiceImpl()
		if err != nil {
			return nil, err
		}
		defaultPolicy, err := defaultPolicyDb.GetByAny(&orm.GatewayDefaultPolicy{
			Level:     orm.POLICY_PACKAGE_LEVEL,
			Name:      name,
			PackageId: packageId,
		})
		if err != nil {
			return nil, err
		}
		if defaultPolicy == nil || len(defaultPolicy.Config) == 0 {
			dto := engine.CreateDefaultConfig(ctx)
			dto.SetGlobal(true)
			return dto, nil
		}
		policyConfig = defaultPolicy.Config
		useDefault = true
	}
done:
	dto, err, _ := engine.UnmarshalConfig(policyConfig)
	if err != nil {
		log.Errorf("unmarshal config failed, confg:%s, err:%+v", policyConfig, err)
		return nil, err
	}
	dto.SetGlobal(useDefault)
	return dto, nil
}

func (policy *BasePolicy) NeedResetAnnotation(dto PolicyDto) bool {
	return !dto.Enable()
}

func (policy *BasePolicy) SetName(name string) {
	policy.Name = name
}

func (policy *BasePolicy) GetName() string {
	return policy.Name
}

func (policy *BasePolicy) SetProvider(provider string) {
	policy.Provider = provider
}

func (policy *BasePolicy) GetProvider() string {
	return policy.Provider
}

func (policy *BasePolicy) MergeDiceConfig(map[string]interface{}) (PolicyDto, error) {
	return nil, nil
}

func (policy *BasePolicy) CreateDefaultConfig(map[string]interface{}) PolicyDto {
	return nil
}

func (policy *BasePolicy) NeedSerialUpdate() bool {
	return false
}

func (policy *BasePolicy) UnmarshalConfig([]byte) (PolicyDto, error, string) {
	return nil, nil, ""
}

func GetGatewayAdapter(ctx map[string]interface{}, policyName string) (gatewayAdapter interface{}, gatewayProvider string, err error) {
	gatewayAdapter, ok := ctx[CTX_KONG_ADAPTER]
	if !ok {
		gatewayAdapter, ok = ctx[CTX_MSE_ADAPTER]
		if !ok {
			errMsg := "convert failed: can not get gateway adapter from ctx"
			log.Errorf(errMsg)
			return nil, "", errors.Errorf(errMsg)
		}
		log.Debugf("use MSE gateway ParseConfig for policy %s", policyName)
		gatewayProvider = ProviderMSE
	} else {
		log.Debugf("use Kong gateway ParseConfig for policy %s", policyName)
	}
	return gatewayAdapter, gatewayProvider, nil
}

// NonSwitchUpdateMSEPluginConfig 初创路由或者关闭路由策略（PolicyDto.Switch == false）的时候,都会进入 ParseConfig 的同一段逻辑中， 但:
// 1. 如果是关闭路由策略，则对应的逻辑里需要清除已经配置的插件策略，一般直接就能处理了，因此进入不了 nonSwitchUpdateMSEPluginConfig() 的逻辑
// 2. 如果是新建路由，实际上是不需要进行处理的（但网关应用默认策略实际上还是会进入 ParseConfig），此时路由还没被 MSE 网关识别到，但可以延时等待拿到对应的新的路由信息，然后进行类似清除路由对应的策略配置的设置即可，但这个过程不能同步等待，因此异步执行，最多重试3次
func NonSwitchUpdateMSEPluginConfig(mseAdapter gateway_providers.GatewayAdapter, pluginReq *providerDto.PluginReqDto, zoneName string, msePluginName string) {
	for i := 0; i < 3; i++ {
		time.Sleep(10 * time.Second)
		//resp, err := mseAdapter.CreateOrUpdatePluginById(policy.buildPluginReq(policyDto, mseCommon.MseProviderName, strings.ToLower(zoneName)))
		resp, err := mseAdapter.CreateOrUpdatePluginById(pluginReq)
		if err != nil {
			if i == 2 {
				log.Errorf("can not update mse %s plugin for 4 times in 30s, err: %v", msePluginName, err)
				return
			}
			continue
		}
		log.Infof("create or update mse %s plugin for zonename=%s with response: %+v", msePluginName, strings.ToLower(zoneName), *resp)
		break
	}
	return
}

func GetGatewayProvider(clusterName string) (string, error) {
	if clusterName == "" {
		return "", errors.Errorf("clusterName is nil")
	}
	azDb, err := db.NewGatewayAzInfoServiceImpl()
	if err != nil {
		return "", err
	}
	_, azInfo, err := azDb.GetAzInfoByClusterName(clusterName)
	if err != nil {
		return "", err
	}

	if azInfo != nil && azInfo.GatewayProvider != "" {
		return azInfo.GatewayProvider, nil
	}

	return "", nil
}

func GetPolicyEngine(provider, name string) (policy PolicyEngine, err error) {
	var m = engines.MSE
	switch strings.ToUpper(provider) {
	case "", "NKE":
		m = engines.NKE
	}
	value, ok := m.Load(name)
	if !ok || value == nil {
		return nil, errors.Errorf("policy provider not registered, name:%s", name)
	}
	return value.(PolicyEngine), nil
}

func RegisterPolicyEngine(provider, name string, engine PolicyEngine) error {
	var m *sync.Map = engines.MSE
	if provider == "" || provider == "nke" {
		m = engines.NKE
	}
	if _, ok := m.Load(name); ok {
		return errors.Errorf("policy engine already registered, provider: %s, name:%s", provider, name)
	}
	engine.SetProvider(provider)
	engine.SetName(name)
	m.Store(name, engine)
	log.Infof("engine registerd, provider: %s, name:%s", provider, name)
	return nil
}
