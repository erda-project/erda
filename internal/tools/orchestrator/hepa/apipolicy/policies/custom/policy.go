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

package custom

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
)

type Policy struct {
	apipolicy.BasePolicy
}

// 如下参数在 nginx-template 中有全局设置，因此不能在 cunstom 中再次设置
var UNSETABLE_KEYS = map[string]string{"client_max_body_size": "100m",
	"proxy_connect_timeout": "5s",
	"proxy_buffer_size":     "256k",
	"proxy_buffers":         "4 256k",
	"proxy_http_version":    "1.1",
	"proxy_cookie_domain":   "off",
	"proxy_cookie_path":     "off",
	"port_in_redirect":      "off",
}

func (policy Policy) CreateDefaultConfig(ctx map[string]interface{}) apipolicy.PolicyDto {
	dto := &PolicyDto{}
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
		empty := ""
		res.IngressAnnotation = &apipolicy.IngressAnnotation{
			LocationSnippet: &empty,
		}
		return res, nil
	}

	// 如下参数在 nginx-template 中有全局设置，因此不能在 custom 中再次设置
	src := strings.TrimSpace(policyDto.Config)
	for strings.HasPrefix(src, "\n") {
		src = strings.TrimPrefix(src, "\n")
		src = strings.TrimSpace(src)
	}
	kvs := strings.Split(src, "\n")
	for _, kv := range kvs {
		kv = strings.TrimSpace(kv)
		for strings.HasPrefix(kv, "\n") {
			kv = strings.TrimPrefix(kv, "\n")
			kv = strings.TrimSpace(kv)
		}

		kvPair := strings.Split(kv, " ")
		if len(kvPair) >= 1 {
			if _, ok := UNSETABLE_KEYS[kvPair[0]]; ok {
				if kvPair[0] != "client_max_body_size" {
					return res, errors.Errorf("invalid config: %s not allowed set in custom policy, it have set in configmap nginx-template", kvPair[0])
				} else {
					return res, errors.Errorf("invalid config: %s not allowed set in custom policy, it have set in configmap nginx-template, please set it in policy proxy or do not set it ", kvPair[0])
				}
			}
		}
	}

	if policyDto.Config != "" {
		res.IngressAnnotation = &apipolicy.IngressAnnotation{
			LocationSnippet: &policyDto.Config,
		}
	}
	return res, nil
}

func init() {
	err := apipolicy.RegisterPolicyEngine(apipolicy.Policy_Engine_Custom, &Policy{})
	if err != nil {
		panic(err)
	}
}
