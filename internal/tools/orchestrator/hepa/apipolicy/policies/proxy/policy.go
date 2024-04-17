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

package proxy

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
	annotationscommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/common"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/config"
	gatewayproviders "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers"
	providerDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	mseCommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	db "github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
)

type Policy struct {
	apipolicy.BasePolicy
}

func (policy Policy) CreateDefaultConfig(gatewayProvider string, ctx map[string]interface{}) apipolicy.PolicyDto {
	dto := &PolicyDto{
		ReqBuffer:         true,
		RespBuffer:        true,
		ClientReqLimit:    100,
		ClientReqTimeout:  60,
		ClientRespTimeout: 60,
		ProxyReqTimeout:   60,
		ProxyRespTimeout:  60,
		HostPassthrough:   config.APIProxyPolicyHostPassThrough(),
		SSLRedirect:       true,
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

// forValidate 用于识别解析的目的，如果解析是用来做 nginx 配置冲突相关的校验，则关于数据表、调用 kong 接口的操作都不会执行
func (policy Policy) ParseConfig(dto apipolicy.PolicyDto, ctx map[string]interface{}, forValidate bool) (apipolicy.PolicyConfig, error) {
	gatewayProvider := ""
	res := apipolicy.PolicyConfig{}
	policyDto, ok := dto.(*PolicyDto)
	if !ok {
		return res, errors.Errorf("invalid config:%+v", dto)
	}

	var adapter gatewayproviders.GatewayAdapter
	gatewayAdapter, gatewayProvider, err := policy.GetGatewayAdapter(ctx, apipolicy.Policy_Engine_Proxy)
	if err != nil {
		return res, err
	}

	adapter, ok = gatewayAdapter.(gatewayproviders.GatewayAdapter)
	if !ok {
		return res, errors.Errorf("convert failed:%+v", gatewayAdapter)
	}

	if !policyDto.Switch {
		switch gatewayProvider {
		case mseCommon.MseProviderName:
			res.IngressAnnotation = &apipolicy.IngressAnnotation{
				Annotation: map[string]*string{
					string(annotationscommon.AnnotationSSLRedirect): nil,
					string(mseCommon.AnnotationMSETimeOut):          nil,
				},
			}
		default:
			emptyStr := ""
			// use empty str trigger regions update
			res.IngressAnnotation = &apipolicy.IngressAnnotation{
				LocationSnippet: &emptyStr,
				Annotation: map[string]*string{
					string(annotationscommon.AnnotationProxyRequestBuffering): nil,
					string(annotationscommon.AnnotationProxyBuffering):        nil,
					string(annotationscommon.AnnotationProxyBodySize):         nil,
					string(annotationscommon.AnnotationProxySendTimeout):      nil,
					string(annotationscommon.AnnotationProxyReadTimeout):      nil,
					string(annotationscommon.AnnotationSSLRedirect):           nil,
				},
			}
		}
		return res, nil
	}

	annotation := map[string]*string{}
	switch gatewayProvider {
	case mseCommon.MseProviderName:
		if policyDto.SSLRedirect {
			value := "true"
			annotation[string(annotationscommon.AnnotationSSLRedirect)] = &value
		} else {
			value := "false"
			annotation[string(annotationscommon.AnnotationSSLRedirect)] = &value
		}

		if policyDto.ProxyReqTimeout > 0 {
			value := fmt.Sprintf("%d", policyDto.ProxyReqTimeout)
			annotation[string(mseCommon.AnnotationMSETimeOut)] = &value
		}

		res.IngressAnnotation = &apipolicy.IngressAnnotation{
			Annotation: annotation,
		}
	default:
		if policyDto.ReqBuffer {
			value := "on"
			annotation[string(annotationscommon.AnnotationProxyRequestBuffering)] = &value
		} else {
			value := "off"
			annotation[string(annotationscommon.AnnotationProxyRequestBuffering)] = &value
		}
		if policyDto.RespBuffer {
			value := "on"
			annotation[string(annotationscommon.AnnotationProxyBuffering)] = &value
		} else {
			value := "off"
			annotation[string(annotationscommon.AnnotationProxyBuffering)] = &value
		}
		if policyDto.SSLRedirect {
			value := "true"
			annotation[string(annotationscommon.AnnotationSSLRedirect)] = &value
		} else {
			value := "false"
			annotation[string(annotationscommon.AnnotationSSLRedirect)] = &value
		}
		limit := fmt.Sprintf("%dm", policyDto.ClientReqLimit)
		annotation[string(annotationscommon.AnnotationProxyBodySize)] = &limit
		reqTimeout := fmt.Sprintf("%d", policyDto.ProxyReqTimeout)
		annotation[string(annotationscommon.AnnotationProxySendTimeout)] = &reqTimeout
		respTimeout := fmt.Sprintf("%d", policyDto.ProxyRespTimeout)
		annotation[string(annotationscommon.AnnotationProxyReadTimeout)] = &respTimeout
		clientHeaderTimeout := fmt.Sprintf("client_header_timeout %ds;", policyDto.ClientReqTimeout)
		annotation[string(annotationscommon.AnnotationServerSnippet)] = &clientHeaderTimeout
		snippet := fmt.Sprintf(`
client_body_timeout %ds;
send_timeout %ds;
`, policyDto.ClientReqTimeout, policyDto.ClientRespTimeout)
		res.IngressAnnotation = &apipolicy.IngressAnnotation{
			LocationSnippet: &snippet,
			Annotation:      annotation,
		}
	}

	value, ok := ctx[apipolicy.CTX_ZONE]
	if !ok {
		return res, errors.Errorf("get identify failed:%+v", ctx)
	}
	zone, ok := value.(*orm.GatewayZone)
	if !ok {
		return res, errors.Errorf("convert failed:%+v", value)
	}
	policyDb, _ := db.NewGatewayPolicyServiceImpl()
	exist, err := policyDb.GetByAny(&orm.GatewayPolicy{
		ZoneId:     zone.Id,
		PluginName: "host-passthrough",
	})
	if err != nil {
		return res, err
	}
	if exist != nil && !policyDto.HostPassthrough {
		if !forValidate {
			err = adapter.RemovePlugin(exist.PluginId)
			if err != nil {
				return res, err
			}
			policyDb, _ := db.NewGatewayPolicyServiceImpl()
			_ = policyDb.DeleteById(exist.Id)
			res.KongPolicyChange = true
		}
	}
	if exist == nil && policyDto.HostPassthrough {
		if !forValidate {
			disable := false
			kongReq := &providerDto.PluginReqDto{
				Name:    "host-passthrough",
				Config:  map[string]interface{}{},
				Enabled: &disable,
			}
			resp, err := adapter.AddPlugin(kongReq)
			if err != nil {
				return res, err
			}
			configByte, err := json.Marshal(resp.Config)
			if err != nil {
				return res, err
			}
			policyDao := &orm.GatewayPolicy{
				ZoneId:     zone.Id,
				PluginName: "host-passthrough",
				Category:   apipolicy.Policy_Category_Proxy,
				PluginId:   resp.Id,
				Config:     configByte,
				Enabled:    1,
			}
			policyDb, _ := db.NewGatewayPolicyServiceImpl()
			err = policyDb.Insert(policyDao)
			if err != nil {
				return res, err
			}
			res.KongPolicyChange = true
		}
	}
	return res, nil
}

func init() {
	err := apipolicy.RegisterPolicyEngine(apipolicy.Policy_Engine_Proxy, &Policy{})
	if err != nil {
		panic(err)
	}
}
