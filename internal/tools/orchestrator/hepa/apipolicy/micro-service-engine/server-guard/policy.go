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

package serverguard

import (
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
	mse "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
)

type Policy struct {
	*apipolicy.BasePolicy
}

func (policy Policy) NeedSerialUpdate() bool {
	return true
}

func (policy Policy) CreateDefaultConfig(map[string]interface{}) apipolicy.PolicyDto {
	dto := &PolicyDto{
		ExtraLatency:   500,
		RefuseCode:     429,
		RefuseResponse: "System is busy, please try it later.",
	}

	dto.RefuseCode = 503
	dto.RefuseResponse = "local_rate_limited"

	dto.Switch = false
	return dto
}

func (policy Policy) MergeDiceConfig(conf map[string]interface{}) (apipolicy.PolicyDto, error) {
	dto := &PolicyDto{
		ExtraLatency:   500,
		RefuseCode:     429,
		RefuseResponse: "System is busy, please try it later.",
	}
	if len(conf) == 0 {
		return dto, nil
	}
	dto.Switch = true
	if value, ok := conf["qps"]; ok {
		if vv, ok := value.(float64); ok {
			if vv != 0 {
				dto.MaxTps = int64(vv)
			} else {
				dto.Switch = false
			}
		}
	}
	if value, ok := conf["max_delay"]; ok {
		if vv, ok := value.(float64); ok {
			if vv != 0 {
				dto.ExtraLatency = int64(vv)
			}
		}
	}
	if value, ok := conf["deny_status"]; ok {
		if vv, ok := value.(float64); ok {
			if vv != 0 {
				dto.RefuseCode = int64(vv)
			}
		}
	}
	if value, ok := conf["deny_content"]; ok {
		if vv, ok := value.(string); ok {
			if vv != "" {
				dto.RefuseResponse = vv
			}
		}
	}
	dto.AdjustDto()
	return dto, nil
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

func (policy Policy) ParseConfig(dto apipolicy.PolicyDto, _ map[string]interface{}, forValidate bool) (apipolicy.PolicyConfig, error) {
	res := apipolicy.PolicyConfig{}
	policyDto, ok := dto.(*PolicyDto)
	if !ok {
		return res, errors.Errorf("invalid config:%+v", dto)
	}

	if !policyDto.Switch {
		// Kong
		emptyStr := ""
		// use empty str trigger regions update
		res.IngressAnnotation = &apipolicy.IngressAnnotation{
			LocationSnippet: &emptyStr,
		}
		annotations := make(map[string]*string)
		annotations[string(mse.AnnotationMSELimitRouteLimitRPS)] = nil
		annotations[string(mse.AnnotationMSELimitRouteLimitBurstMultiplier)] = nil
		res.IngressAnnotation.Annotation = annotations
		res.IngressController = &apipolicy.IngressController{
			HttpSnippet:   &emptyStr,
			ServerSnippet: &emptyStr,
		}
		if !forValidate {
			res.AnnotationReset = true
		}
		return res, nil
	}

	res.AnnotationReset = true
	if res.IngressAnnotation == nil {
		res.IngressAnnotation = &apipolicy.IngressAnnotation{}
	}
	setMSEIngressAnnotation(policyDto, res.IngressAnnotation)

	return res, nil
}
func setMSEIngressAnnotation(policyDto *PolicyDto, ingressAnnotations *apipolicy.IngressAnnotation) {
	annotations := make(map[string]*string)
	rateStr := strconv.FormatInt(policyDto.MaxTps, 10)
	multiplier := mse.MseBurstMultiplier2X
	if policyDto.Busrt > 0 {
		multiplier = strconv.FormatInt(policyDto.Busrt, 10)
	}

	switch {
	case policyDto.MaxTps > MSE_BURST_MULTIPLIER_THOUSAND:
		multiplier = mse.MseBurstMultiplier1X
	case policyDto.MaxTps > MSE_BURST_MULTIPLIER_HUNDRED:
		multiplier = mse.MseBurstMultiplier2X
	case policyDto.MaxTps > MSE_BURST_MULTIPLIER_TEN:
		multiplier = mse.MseBurstMultiplier3X
	default:
		multiplier = mse.MseBurstMultiplier4X
	}

	annotations[string(mse.AnnotationMSELimitRouteLimitRPS)] = &rateStr
	annotations[string(mse.AnnotationMSELimitRouteLimitBurstMultiplier)] = &multiplier
	if ingressAnnotations == nil {
		ingressAnnotations = &apipolicy.IngressAnnotation{}
	}
	if ingressAnnotations.Annotation == nil {
		ingressAnnotations.Annotation = annotations
	} else {
		for k, v := range annotations {
			ingressAnnotations.Annotation[k] = v
		}
	}
}

func init() {
	err := apipolicy.RegisterPolicyEngine(apipolicy.ProviderMSE, apipolicy.Policy_Engine_Service_Guard, &Policy{BasePolicy: &apipolicy.BasePolicy{}})
	if err != nil {
		panic(err)
	}
}
