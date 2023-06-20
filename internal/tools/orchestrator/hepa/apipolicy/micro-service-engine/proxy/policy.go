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
	gatewayproviders "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers"
	providerDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	mse "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	db "github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
)

type Policy struct {
	*apipolicy.BasePolicy
}

func (policy Policy) CreateDefaultConfig(map[string]interface{}) apipolicy.PolicyDto {
	dto := &PolicyDto{
		ReqBuffer:         true,
		RespBuffer:        true,
		ClientReqLimit:    100,
		ClientReqTimeout:  60,
		ClientRespTimeout: 60,
		ProxyReqTimeout:   60,
		ProxyRespTimeout:  60,
		HostPassthrough:   false,
		SSLRedirect:       true,
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

// forValidate 用于识别解析的目的，如果解析是用来做 nginx 配置冲突相关的校验，则关于数据表、调用 kong 接口的操作都不会执行
func (policy Policy) ParseConfig(dto apipolicy.PolicyDto, ctx map[string]interface{}, forValidate bool) (apipolicy.PolicyConfig, error) {
	res := apipolicy.PolicyConfig{}
	policyDto, ok := dto.(*PolicyDto)
	if !ok {
		return res, errors.Errorf("invalid config:%+v", dto)
	}

	var adapter gatewayproviders.GatewayAdapter
	gatewayAdapter, _, err := apipolicy.GetGatewayAdapter(ctx, apipolicy.Policy_Engine_Proxy)
	if err != nil {
		return res, err
	}

	adapter, ok = gatewayAdapter.(gatewayproviders.GatewayAdapter)
	if !ok {
		return res, errors.Errorf("convert failed:%+v", gatewayAdapter)
	}

	if !policyDto.Switch {
		res.IngressAnnotation = &apipolicy.IngressAnnotation{
			Annotation: map[string]*string{
				string(annotationscommon.AnnotationSSLRedirect): nil,
				string(mse.AnnotationMSETimeOut):                nil,
			},
		}
		return res, nil
	}

	annotation := map[string]*string{}
	if policyDto.SSLRedirect {
		value := "true"
		annotation[string(annotationscommon.AnnotationSSLRedirect)] = &value
	} else {
		value := "false"
		annotation[string(annotationscommon.AnnotationSSLRedirect)] = &value
	}

	if policyDto.ProxyReqTimeout > 0 {
		value := fmt.Sprintf("%d", policyDto.ProxyReqTimeout)
		annotation[string(mse.AnnotationMSETimeOut)] = &value
	}

	res.IngressAnnotation = &apipolicy.IngressAnnotation{
		Annotation: annotation,
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
	err := apipolicy.RegisterPolicyEngine(apipolicy.ProviderMSE, apipolicy.Policy_Engine_Proxy, &Policy{BasePolicy: &apipolicy.BasePolicy{}})
	if err != nil {
		panic(err)
	}
}
