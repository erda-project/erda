// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package waf

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/modules/hepa/apipolicy"

	"github.com/pkg/errors"
)

type Policy struct {
	apipolicy.BasePolicy
}

func (policy Policy) NeedSerialUpdate() bool {
	return true
}

func (policy Policy) CreateDefaultConfig(ctx map[string]interface{}) apipolicy.PolicyDto {
	return &PolicyDto{
		WafEnable:   WAF_SWITCH_WATCH,
		RemoveRules: "",
	}
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
	res.IngressAnnotation = &apipolicy.IngressAnnotation{
		Annotation: map[string]*string{},
	}
	var modSnippet *string
	if !policyDto.Switch || policyDto.WafEnable == WAF_SWITCH_OFF {
		modSnippet = nil
	} else if policyDto.WafEnable == WAF_SWITCH_WATCH {
		snippet := `Include /etc/nginx/owasp-modsecurity-crs/nginx-modsecurity.conf
SecAuditEngine RelevantOnly
SecAuditLogRelevantStatus "^(?:5|4(?!04))"
SecAuditLogParts ABH
SecRuleEngine DetectionOnly
SecAuditLog /dev/stderr
`
		modSnippet = &snippet
	} else if policyDto.WafEnable == WAF_SWITCH_ON {
		snippet := `Include /etc/nginx/owasp-modsecurity-crs/nginx-modsecurity.conf
SecAuditEngine RelevantOnly
SecAuditLogRelevantStatus "^(?:5|4(?!04))"
SecAuditLogParts ABH
SecRuleEngine On
SecAuditLog /dev/stderr
`
		modSnippet = &snippet
	}
	if policyDto.RemoveRules != "" && modSnippet != nil {
		*modSnippet = *modSnippet + fmt.Sprintf("SecRuleRemoveById %s\n", policyDto.RemoveRules)
	}
	res.IngressAnnotation.Annotation[MODSECURITY_SNIPPET] = modSnippet
	if modSnippet != nil {
		enable := "true"
		res.IngressAnnotation.Annotation[MODSECURITY_ENABLE] = &enable
	} else {
		res.IngressAnnotation.Annotation[MODSECURITY_ENABLE] = nil
	}
	return res, nil

}

func init() {
	err := apipolicy.RegisterPolicyEngine("safety-waf", &Policy{})
	if err != nil {
		panic(err)
	}
}
