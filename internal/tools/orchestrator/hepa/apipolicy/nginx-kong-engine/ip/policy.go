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
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
	annotationscommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/common"
	gateway_providers "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers"
	providerDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	mseCommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/k8s"
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

	if !policyDto.Switch {
		emptyStr := ""
		// use empty str trigger regions update
		res.IngressAnnotation = &apipolicy.IngressAnnotation{
			Annotation: map[string]*string{
				string(annotationscommon.AnnotationWhiteListSourceRange): nil,
				string(mseCommon.AnnotationMSEBlackListSourceRange):      nil,
				string(annotationscommon.AnnotationLimitConnections):     nil,
				//string(mseCommon.AnnotationMSEDomainWhitelistSourceRange):     nil,
				//string(mseCommon.AnnotationMSEDomainBlacklistSourceRange):     nil,
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
	value, ok := ctx[apipolicy.CTX_IDENTIFY]
	if !ok {
		return res, errors.Errorf("get identify failed:%+v", ctx)
	}
	id, ok := value.(string)
	if !ok {
		return res, errors.Errorf("convert failed:%+v", value)
	}
	value, ok = ctx[apipolicy.CTX_K8S_CLIENT]
	if !ok {
		return res, errors.Errorf("get k8s client failed:%+v", ctx)
	}
	adapter, ok := value.(k8s.K8SAdapter)
	if !ok {
		return res, errors.Errorf("convert failed:%+v", value)
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

	limitConnZone := ""
	limitConn := ""
	limitReqZone := ""
	limitReq := ""
	count, err := adapter.CountIngressController()
	if err != nil {
		count = 1
		logrus.Errorf("get ingress controller count failed, err:%+v", err)
	}

	if policyDto.IpMaxConnections > 0 {
		limitConnZone = fmt.Sprintf("limit_conn_zone $binary_remote_addr zone=ip-conn-%s:10m;\n", id)
		maxConn := int64(math.Ceil(float64(policyDto.IpMaxConnections) / float64(count)))
		limitConn = fmt.Sprintf("limit_conn ip-conn-%s %d;\n", id, maxConn)
	} else {
		limitConnZone = fmt.Sprintf("limit_conn_zone $binary_remote_addr zone=ip-conn-%s:10m;\n", id)
	}

	if policyDto.IpRate != nil {
		unit := "r/s"
		if policyDto.IpRate.Unit == MINUTE {
			unit = "r/m"
		}
		maxReq := int64(math.Ceil(float64(policyDto.IpRate.Rate) / float64(count)))
		limitReqZone = fmt.Sprintf("limit_req_zone $binary_remote_addr zone=ip-req-%s:10m rate=%d%s;\n",
			id, maxReq, unit)
		limitReq = fmt.Sprintf("limit_req zone=ip-req-%s nodelay;\n", id)
	} else {
		limitReqZone = fmt.Sprintf("limit_req_zone $binary_remote_addr zone=ip-req-%s:10m rate=%d%s;\n",
			id, 10000, "r/s")
	}
	locationSnippet := ipSource + acls + limitConn + limitReq
	res.IngressAnnotation = &apipolicy.IngressAnnotation{
		LocationSnippet: &locationSnippet,
	}
	httpSnippet := limitConnZone + limitReqZone
	res.IngressController = &apipolicy.IngressController{
		HttpSnippet: &httpSnippet,
	}

	return res, nil
}

func (policy Policy) buildMSEPluginReq(dto *PolicyDto, zoneName string) *providerDto.PluginReqDto {
	req := &providerDto.PluginReqDto{
		Name:     mseCommon.MsePluginIP,
		Config:   map[string]interface{}{},
		Enabled:  &dto.Switch,
		ZoneName: zoneName,
	}

	switch dto.IpSource {
	case X_REAL_IP:
		req.Config[mseCommon.MseErdaIpIpSource] = mseCommon.MseErdaIpSourceXRealIP
	case X_FORWARDED_FOR:
		req.Config[mseCommon.MseErdaIpIpSource] = mseCommon.MseErdaIpSourceXForwardedFor
	default:
		req.Config[mseCommon.MseErdaIpIpSource] = mseCommon.MseErdaIpSourceRemoteIP
	}

	req.Config[mseCommon.MseErdaIpAclType] = string(dto.IpAclType)
	req.Config[mseCommon.MseErdaIpAclList] = dto.IpAclList

	if dto.Switch {
		req.Config[mseCommon.MseErdaIpRouteSwitch] = true
	} else {
		req.Config[mseCommon.MseErdaIpRouteSwitch] = false
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

func init() {
	err := apipolicy.RegisterPolicyEngine(apipolicy.ProviderNKE, apipolicy.Policy_Engine_IP, &Policy{BasePolicy: &apipolicy.BasePolicy{}})
	if err != nil {
		panic(err)
	}
}
