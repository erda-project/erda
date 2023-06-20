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

package ip

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
	annotationscommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/common"
	gateway_providers "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers"
	providerDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	mse "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
)

type Policy struct {
	*apipolicy.BasePolicy
}

func (policy Policy) CreateDefaultConfig(map[string]interface{}) apipolicy.PolicyDto {
	dto := &PolicyDto{
		IpSource:  REMOTE_IP,
		IpAclType: ACL_BLACK,
	}
	dto.Switch = false
	return dto
}

func (policy Policy) UnmarshalConfig(config []byte) (apipolicy.PolicyDto, error, string) {
	policyDto := &PolicyDto{}
	err := json.Unmarshal(config, policyDto)
	if err != nil {
		return nil, errors.Wrapf(err, "json parse config failed, config:%s", config), "Invalid config"
	}
	ok, msg := policyDto.IsValidDto()
	if !ok {
		return nil, errors.Errorf("invalid policy dto, msg:%s", msg), msg
	}
	return policyDto, nil, ""
}

func (policy Policy) ParseConfig(dto apipolicy.PolicyDto, ctx map[string]interface{}, forValidate bool) (apipolicy.PolicyConfig, error) {
	res := apipolicy.PolicyConfig{}
	policyDto, ok := dto.(*PolicyDto)
	if !ok {
		return res, errors.Errorf("invalid config:%+v", dto)
	}

	gatewayAdapter, _, err := apipolicy.GetGatewayAdapter(ctx, apipolicy.Policy_Engine_IP)
	if err != nil {
		return res, err
	}

	if !policyDto.Switch {
		switch policyDto.IpSource {
		case X_REAL_IP, X_FORWARDED_FOR:
			mseAdapter, zone, err := getMSEAdapterAndZone(ctx, gatewayAdapter)
			if err != nil {
				return res, err
			}

			// 创建网关路由也会走到这里，但这个时候 MSE 那边应该还没有发现到新路由，因此调用此插件会失败，但创建过程又不能一直等待，因此这里做个异步处理
			if zone.Type == service.ZONE_TYPE_PACKAGE_API {
				resp, err := mseAdapter.CreateOrUpdatePluginById(policy.buildMSEPluginReq(policyDto, strings.ToLower(zone.Name)))
				if err != nil {
					go policy.nonSwitchUpdateMSEPluginConfig(mseAdapter, policyDto, zone.Name)
				} else {
					logrus.Infof("create or update mse erda-ip plugin with response: %+v", *resp)
				}
			}
		default:
		}

		emptyStr := ""
		// use empty str trigger regions update
		res.IngressAnnotation = &apipolicy.IngressAnnotation{
			Annotation: map[string]*string{
				string(annotationscommon.AnnotationWhiteListSourceRange): nil,
				string(mse.AnnotationMSEBlackListSourceRange):            nil,
				string(annotationscommon.AnnotationLimitConnections):     nil,
				//string(mse.AnnotationMSEDomainWhitelistSourceRange):     nil,
				//string(mse.AnnotationMSEDomainBlacklistSourceRange):     nil,
			},
			LocationSnippet: &emptyStr,
		}

		res.IngressController = &apipolicy.IngressController{
			HttpSnippet: &emptyStr,
		}
		if !forValidate {
			res.AnnotationReset = true
		}
		return res, nil
	}

	ipSource := ""
	if policyDto.IpSource != REMOTE_IP {
		ipSource += "set_real_ip_from 0.0.0.0/0;\n"
		if policyDto.IpSource == X_REAL_IP {
			ipSource += "real_ip_header X-Real-IP;\n"
		} else if policyDto.IpSource == X_FORWARDED_FOR {
			ipSource += "real_ip_header X-Forwarded-For;\n"
		}
	}
	acls := ""
	if len(policyDto.IpAclList) > 0 {
		var prefix string
		var bottom string
		if policyDto.IpAclType == ACL_BLACK {
			bottom = "allow all;\n"
			prefix = "deny "
		} else {
			bottom = "deny all;\n"
			prefix = "allow "
		}
		for _, ip := range policyDto.IpAclList {
			acls += prefix + ip + ";\n"
		}
		acls += bottom
	}

	if policyDto.IpRate != nil {
		logrus.Warningf("Current use MSE gateway, please set rate in policy %s", apipolicy.Policy_Engine_Service_Guard)
	}

	switch policyDto.IpSource {
	case X_REAL_IP, X_FORWARDED_FOR:
		mseAdapter, zone, err := getMSEAdapterAndZone(ctx, gatewayAdapter)
		if err != nil {
			return res, err
		}

		if zone.Type == service.ZONE_TYPE_PACKAGE_API {
			resp, err := mseAdapter.CreateOrUpdatePluginById(policy.buildMSEPluginReq(policyDto, strings.ToLower(zone.Name)))
			if err != nil {
				return res, err
			}
			logrus.Infof("create or update mse erda-ip plugin with response: %+v", *resp)
		}

		res.IngressAnnotation = &apipolicy.IngressAnnotation{
			Annotation: map[string]*string{
				string(annotationscommon.AnnotationWhiteListSourceRange): nil,
				string(mse.AnnotationMSEBlackListSourceRange):            nil,
				string(annotationscommon.AnnotationLimitConnections):     nil,
				//string(mse.AnnotationMSEDomainWhitelistSourceRange):     nil,
				//string(mse.AnnotationMSEDomainBlacklistSourceRange):     nil,
			},
		}

	default:
		res.IngressAnnotation = setMseIngressAnnotations(policyDto)
	}

	return res, nil
}

func setMseIngressAnnotations(policyDto *PolicyDto) *apipolicy.IngressAnnotation {
	ingressAnnotations := &apipolicy.IngressAnnotation{
		Annotation: map[string]*string{
			string(annotationscommon.AnnotationWhiteListSourceRange): nil,
			string(mse.AnnotationMSEBlackListSourceRange):            nil,
		},
	}

	switch policyDto.IpAclType {
	case ACL_BLACK:
		bl := strings.Join(policyDto.IpAclList, ",")
		ingressAnnotations.Annotation[string(mse.AnnotationMSEBlackListSourceRange)] = &bl
	default:
		wl := strings.Join(policyDto.IpAclList, ",")
		ingressAnnotations.Annotation[string(annotationscommon.AnnotationWhiteListSourceRange)] = &wl
	}

	return ingressAnnotations
}

func (policy Policy) buildMSEPluginReq(dto *PolicyDto, zoneName string) *providerDto.PluginReqDto {
	req := &providerDto.PluginReqDto{
		Name:     mse.MsePluginIP,
		Config:   map[string]interface{}{},
		Enabled:  &dto.Switch,
		ZoneName: zoneName,
	}

	switch dto.IpSource {
	case X_REAL_IP:
		req.Config[mse.MseErdaIpIpSource] = mse.MseErdaIpSourceXRealIP
	case X_FORWARDED_FOR:
		req.Config[mse.MseErdaIpIpSource] = mse.MseErdaIpSourceXForwardedFor
	default:
		req.Config[mse.MseErdaIpIpSource] = mse.MseErdaIpSourceRemoteIP
	}

	req.Config[mse.MseErdaIpAclType] = string(dto.IpAclType)
	req.Config[mse.MseErdaIpAclList] = dto.IpAclList

	if dto.Switch {
		req.Config[mse.MseErdaIpRouteSwitch] = true
	} else {
		req.Config[mse.MseErdaIpRouteSwitch] = false
	}

	return req
}

// 初创路由或者关闭路由策略（PolicyDto.Switch == false）的时候,都会进入 ParseConfig 的同一段逻辑中， 但:
// 1. 如果是关闭路由策略，则对应的逻辑里需要清除已经配置的插件策略，一般直接就能处理了，因此进入不了 nonSwitchUpdateMSEPluginConfig() 的逻辑
// 2. 如果是新建路由，实际上是不需要进行处理的（但网关应用默认策略实际上还是会进入 ParseConfig），此时路由还没被 MSE 网关识别到，但可以延时等待拿到对应的新的路由信息，然后进行类似清除路由对应的策略配置的设置即可，但这个过程不能同步等待，因此异步执行，最多重试3次
func (policy Policy) nonSwitchUpdateMSEPluginConfig(mseAdapter gateway_providers.GatewayAdapter, policyDto *PolicyDto, zoneName string) {
	for i := 0; i < 3; i++ {
		time.Sleep(10 * time.Second)
		resp, err := mseAdapter.CreateOrUpdatePluginById(policy.buildMSEPluginReq(policyDto, strings.ToLower(zoneName)))
		if err != nil {
			if i == 2 {
				logrus.Errorf("can not update mse erda-ip plugin for 4 times in 30s, err: %v", err)
				return
			}
			continue
		}
		logrus.Infof("create or update mse erda-ip plugin for zonename=%s with response: %+v", strings.ToLower(zoneName), *resp)
		break
	}
	return
}

func getMSEAdapterAndZone(ctx map[string]interface{}, gatewayAdapter interface{}) (gateway_providers.GatewayAdapter, *orm.GatewayZone, error) {
	mseAdapter, ok := gatewayAdapter.(gateway_providers.GatewayAdapter)
	if !ok {
		return nil, nil, errors.Errorf("convert gatewayAdapter failed:%+v", gatewayAdapter)
	}

	value, ok := ctx[apipolicy.CTX_ZONE]
	if !ok {
		return nil, nil, errors.Errorf("get gatewayzone failed:%+v", ctx)
	}
	zone, ok := value.(*orm.GatewayZone)
	if !ok {
		return nil, nil, errors.Errorf("convert gatewayzone failed:%+v", value)
	}
	return mseAdapter, zone, nil
}

func init() {
	err := apipolicy.RegisterPolicyEngine(apipolicy.ProviderMSE, apipolicy.Policy_Engine_IP, &Policy{BasePolicy: &apipolicy.BasePolicy{}})
	if err != nil {
		panic(err)
	}
}
