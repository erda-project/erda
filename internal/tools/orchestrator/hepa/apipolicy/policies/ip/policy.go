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

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
	annotationscommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/common"
	providerDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	mseCommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/k8s"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
)

type Policy struct {
	apipolicy.BasePolicy
}

func (policy Policy) CreateDefaultConfig(ctx map[string]interface{}) apipolicy.PolicyDto {
	dto := &PolicyDto{
		IpSource:  REMOTE_IP,
		IpAclType: ACL_BLACK,
	}
	dto.Switch = false
	return dto
}

func (policy Policy) UnmarshalConfig(config []byte, gatewayProvider string) (apipolicy.PolicyDto, error, string) {
	policyDto := &PolicyDto{}
	err := json.Unmarshal(config, policyDto)
	if err != nil {
		return nil, errors.Wrapf(err, "json parse config failed, config:%s", config), "Invalid config"
	}
	ok, msg := policyDto.IsValidDto(gatewayProvider)
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

	value, ok = ctx[apipolicy.CTX_ZONE]
	if !ok {
		return res, errors.Errorf("get identify failed:%+v", ctx)
	}
	zone, ok := value.(*orm.GatewayZone)
	if !ok {
		return res, errors.Errorf("convert failed:%+v", value)
	}

	gatewayProvider, err := policy.GetGatewayProvider(zone.DiceClusterName)
	if err != nil {
		return res, errors.Errorf("get gateway provider failed for cluster %s:%v\n", zone.DiceClusterName, err)
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

	switch gatewayProvider {
	case mseCommon.Mse_Provider_Name:
		if policyDto.IpRate != nil {
			logrus.Warningf("Current use MSE gateway, please set rate in policy %s", apipolicy.Policy_Engine_Service_Guard)
		}

		if policyDto.IpSource != REMOTE_IP {
			return res, errors.Errorf("current use MSE gateway with this proxy for IPSource=%s, untile mse plugin %s ready\n", policyDto.IpSource, mseCommon.MsePluginIP)
			res.IngressAnnotation = setMseIngressAnnotations(policyDto)
		}
		res.IngressAnnotation = setMseIngressAnnotations(policyDto)

	case "":
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
	default:
		return res, errors.Errorf("unknown gateway provider:%v\n", gatewayProvider)
	}
	return res, nil
}

func setMseIngressAnnotations(policyDto *PolicyDto) *apipolicy.IngressAnnotation {
	ingressAnnotations := &apipolicy.IngressAnnotation{
		Annotation: map[string]*string{
			string(annotationscommon.AnnotationWhiteListSourceRange): nil,
			string(mseCommon.AnnotationMSEBlackListSourceRange):      nil,
		},
	}

	switch policyDto.IpAclType {
	case ACL_BLACK:
		bl := strings.Join(policyDto.IpAclList, ",")
		ingressAnnotations.Annotation[string(mseCommon.AnnotationMSEBlackListSourceRange)] = &bl
	default:
		wl := strings.Join(policyDto.IpAclList, ",")
		ingressAnnotations.Annotation[string(annotationscommon.AnnotationWhiteListSourceRange)] = &wl
	}

	return ingressAnnotations
}

func (policy Policy) buildPluginReq(dto *PolicyDto) *providerDto.PluginReqDto {
	req := &providerDto.PluginReqDto{
		Name:    mseCommon.MsePluginIP,
		Config:  map[string]interface{}{},
		Enabled: &dto.Switch,
	}

	return req
}

func init() {
	err := apipolicy.RegisterPolicyEngine(apipolicy.Policy_Engine_IP, &Policy{})
	if err != nil {
		panic(err)
	}
}
