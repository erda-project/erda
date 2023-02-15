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

package cors

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
	annotationscommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/common"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
)

type Policy struct {
	apipolicy.BasePolicy
}

func (policy Policy) CreateDefaultConfig(ctx map[string]interface{}) apipolicy.PolicyDto {
	dto := &PolicyDto{
		Methods:     "GET, PUT, POST, DELETE, PATCH, OPTIONS",
		Headers:     "$http_access_control_request_headers",
		Origin:      "$from_request_origin_or_referer",
		Credentials: true,
		MaxAge:      86400,
	}
	dto.Switch = false
	return dto
}

func (policy Policy) MergeDiceConfig(conf map[string]interface{}) (apipolicy.PolicyDto, error) {
	dto := &PolicyDto{
		Methods:     "GET, PUT, POST, DELETE, PATCH, OPTIONS",
		Headers:     "$http_access_control_request_headers",
		Origin:      "$from_request_origin_or_referer",
		Credentials: true,
		MaxAge:      86400,
	}
	if len(conf) == 0 {
		return dto, nil
	}
	dto.Switch = true
	if value, ok := conf["allow_origins"]; ok {
		if vv, ok := value.(string); ok {
			if vv != "any" && vv != "*" {
				dto.Origin = vv
			}
		}
	}
	if value, ok := conf["allow_methods"]; ok {
		if vv, ok := value.(string); ok {
			if vv != "any" && vv != "*" {
				dto.Methods = vv
			}
		}
	}
	if value, ok := conf["allow_headers"]; ok {
		if vv, ok := value.(string); ok {
			if vv != "any" && vv != "*" {
				dto.Headers = vv
			}
		}
	}
	if value, ok := conf["allow_credentials"]; ok {
		if vv, ok := value.(bool); ok {
			if !vv {
				dto.Credentials = false
			}
		}
	}
	if value, ok := conf["max_age"]; ok {
		if vv, ok := value.(float64); ok {
			if vv != 0 {
				dto.MaxAge = int64(vv)
			}
		}
	}
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

func (policy Policy) ParseConfig(dto apipolicy.PolicyDto, ctx map[string]interface{}) (apipolicy.PolicyConfig, error) {
	res := apipolicy.PolicyConfig{}
	policyDto, ok := dto.(*PolicyDto)
	if !ok {
		return res, errors.Errorf("invalid config:%+v", dto)
	}

	value, ok := ctx[apipolicy.CTX_ZONE]
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

	if !policyDto.Switch {
		emptyStr := ""
		res.IngressAnnotation = &apipolicy.IngressAnnotation{
			Annotation: map[string]*string{
				string(annotationscommon.AnnotationEnableCORS):           nil,
				string(annotationscommon.AnnotationCORSAllowMethods):     nil,
				string(annotationscommon.AnnotationCORSAllowHeaders):     nil,
				string(annotationscommon.AnnotationCORSAllowOrigin):      nil,
				string(annotationscommon.AnnotationCORSAllowCredentials): nil,
				string(annotationscommon.AnnotationCORSMaxAge):           nil,
			},
			LocationSnippet: &emptyStr,
		}
		return res, nil
	}
	coreSnippet := fmt.Sprintf(`more_set_headers 'Access-Control-Allow-Origin: %s';
more_set_headers 'Access-Control-Allow-Methods: %s';
more_set_headers 'Access-Control-Allow-Headers: %s';
`, policyDto.Origin, policyDto.Methods, policyDto.Headers)
	if policyDto.Credentials {
		coreSnippet += `more_set_headers 'Access-Control-Allow-Credentials: true';
`
	}
	locationSnippet := fmt.Sprintf(`if ($request_method = 'OPTIONS') {
   %s
   more_set_headers 'Access-Control-Max-Age: %d';
   more_set_headers 'Content-Type: text/plain charset=UTF-8';
   more_set_headers 'Content-Length: 0';
   return 200;
}
%s
`, coreSnippet, policyDto.MaxAge, coreSnippet)

	switch gatewayProvider {
	case mse.Mse_Provider_Name:
		res.IngressAnnotation = policy.setIngressAnnotations(gatewayProvider, policyDto, locationSnippet)
	case "":
		res.IngressAnnotation = policy.setIngressAnnotations(gatewayProvider, policyDto, locationSnippet)
	default:
		return res, errors.Errorf("unknown gateway provider:%v\n", gatewayProvider)
	}
	emptyStr := ""
	// trigger httpsnippet update
	res.IngressController = &apipolicy.IngressController{
		HttpSnippet: &emptyStr,
	}
	return res, nil
}

func (policy Policy) setIngressAnnotations(gatewayProvider string, policyDto *PolicyDto, locationSnippet string) *apipolicy.IngressAnnotation {
	var ret *apipolicy.IngressAnnotation
	switch gatewayProvider {
	case mse.Mse_Provider_Name:
		corsEnable := "true"
		corsMethods := policyDto.Methods
		corsHeaders := policyDto.Headers
		corsOrigin := policyDto.Origin
		corsCredentials := fmt.Sprintf("%v", policyDto.Credentials)
		corsMaxAge := "86400"
		if policyDto.MaxAge > 0 {
			corsMaxAge = fmt.Sprintf("%v", policyDto.MaxAge)
		}
		ret = &apipolicy.IngressAnnotation{
			Annotation: map[string]*string{
				string(annotationscommon.AnnotationEnableCORS):           &corsEnable,
				string(annotationscommon.AnnotationCORSAllowMethods):     &corsMethods,
				string(annotationscommon.AnnotationCORSAllowHeaders):     &corsHeaders,
				string(annotationscommon.AnnotationCORSAllowOrigin):      &corsOrigin,
				string(annotationscommon.AnnotationCORSAllowCredentials): &corsCredentials,
				string(annotationscommon.AnnotationCORSMaxAge):           &corsMaxAge,
			},
			LocationSnippet: &locationSnippet,
		}
	default:
		ret = &apipolicy.IngressAnnotation{
			Annotation: map[string]*string{
				string(annotationscommon.AnnotationEnableCORS):           nil,
				string(annotationscommon.AnnotationCORSAllowMethods):     nil,
				string(annotationscommon.AnnotationCORSAllowHeaders):     nil,
				string(annotationscommon.AnnotationCORSAllowOrigin):      nil,
				string(annotationscommon.AnnotationCORSAllowCredentials): nil,
				string(annotationscommon.AnnotationCORSMaxAge):           nil,
			},
			LocationSnippet: &locationSnippet,
		}
	}
	return ret
}

func init() {
	err := apipolicy.RegisterPolicyEngine(apipolicy.Policy_Engine_CORS, &Policy{})
	if err != nil {
		panic(err)
	}
}
