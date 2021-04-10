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
	"fmt"
	"regexp"
	"strings"

	"github.com/erda-project/erda/modules/hepa/apipolicy"
)

const (
	WAF_SWITCH_ON       = "on"
	WAF_SWITCH_OFF      = "off"
	WAF_SWITCH_WATCH    = "watch"
	MODSECURITY_SNIPPET = "nginx.ingress.kubernetes.io/modsecurity-snippet"
	MODSECURITY_ENABLE  = "nginx.ingress.kubernetes.io/enable-modsecurity"
)

type PolicyDto struct {
	apipolicy.BaseDto
	WafEnable   string `json:"wafEnable"`
	RemoveRules string `json:"removeRules"`
}

var matchRegex = regexp.MustCompile(`^"?[0-9]+\-?[0-9]*"?$`)

func (dto PolicyDto) IsValidDto() (bool, string) {
	if !dto.Switch {
		return true, ""
	}
	if dto.WafEnable != WAF_SWITCH_ON &&
		dto.WafEnable != WAF_SWITCH_OFF &&
		dto.WafEnable != WAF_SWITCH_WATCH {
		return false, "流量拦截开关参数错误"
	}
	dto.RemoveRules = strings.TrimSpace(dto.RemoveRules)
	if dto.RemoveRules != "" {
		rules := strings.Split(dto.RemoveRules, " ")
		for _, rule := range rules {
			if ok := matchRegex.MatchString(rule); !ok {
				return false, fmt.Sprintf("移除规则ID非法: %s", rule)
			}
		}
	}
	return true, ""
}
