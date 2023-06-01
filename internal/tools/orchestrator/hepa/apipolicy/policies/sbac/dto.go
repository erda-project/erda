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

package sbac

import (
	"net/textproto"
	"net/url"
	"strings"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
	providerDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	mseCommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
)

type PolicyDto struct {
	apipolicy.BaseDto

	AccessControlAPI string   `json:"accessControlAPI"`
	Methods          []string `json:"methods"`
	Patterns         []string `json:"patterns"`
	WithHeaders      []string `json:"withHeaders"`
	WithCookie       bool     `json:"withCookie"`
	WithBody         bool     `json:"withBody"`
}

func (pc PolicyDto) IsValidDto(gatewayProvider string) error {
	if !pc.BaseDto.Switch {
		return nil
	}
	_, err := url.ParseRequestURI(pc.AccessControlAPI)
	return err
}

func (pc PolicyDto) ToPluginReqDto(gatewayProvider, zoneName string) *providerDto.PluginReqDto {
	var req = &providerDto.PluginReqDto{
		Name:    apipolicy.Policy_Engine_SBAC,
		Enabled: &pc.Switch,
		Config: map[string]interface{}{
			"access_control_api": pc.AccessControlAPI,
			"with_body":          pc.WithBody,
		},
	}

	if gatewayProvider == mseCommon.MseProviderName {
		req.Name = mseCommon.MsePluginSbac
		req.ZoneName = zoneName
	}

	// adjust "patterns"
	var patterns []string
	for _, pat := range pc.Patterns {
		if len(pat) > 0 {
			patterns = append(patterns, pat)
		}
	}
	if len(patterns) > 0 {
		req.Config["patterns"] = patterns
	}

	// adjust "methods"
	var methods = make(map[string]bool)
	for _, method := range pc.Methods {
		methods[strings.ToUpper(method)] = true
	}
	if len(methods) > 0 {
		req.Config["methods"] = methods
	}
	// adjust "with_headers"
	var headersKeys = make(map[string]struct{})
	if pc.WithCookie {
		headersKeys[textproto.CanonicalMIMEHeaderKey("cookie")] = struct{}{}
		if gatewayProvider == mseCommon.MseProviderName {
			req.Config["with_cookie"] = true
		}
	}
	for _, header := range pc.WithHeaders {
		if gatewayProvider == mseCommon.MseProviderName {
			headersKeys[strings.ToLower(header)] = struct{}{}
		} else {
			headersKeys[textproto.CanonicalMIMEHeaderKey(header)] = struct{}{}
		}
	}
	var headers []string
	for key := range headersKeys {
		headers = append(headers, key)
	}
	if len(headers) > 0 {
		req.Config["with_headers"] = headers
	}

	if gatewayProvider == mseCommon.MseProviderName {
		if pc.Switch {
			req.Config[mseCommon.MseErdaSBACRouteSwitch] = true
		} else {
			req.Config[mseCommon.MseErdaSBACRouteSwitch] = false
		}
	}

	return req
}
