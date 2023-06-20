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

package builtin

import (
	"encoding/json"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
	annotationscommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/common"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/config"
	gateway_providers "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers"
	providerDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	db "github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
)

type Policy struct {
	*apipolicy.BasePolicy
}

func (policy Policy) NeedSerialUpdate() bool {
	return true
}

// ParseConfig forValidate 用于识别解析的目的，如果解析是用来做 nginx 配置冲突相关的校验，则关于数据表、调用 kong 接口的操作都不会执行
func (policy Policy) ParseConfig(dto apipolicy.PolicyDto, ctx map[string]interface{}, forValidate bool) (apipolicy.PolicyConfig, error) {
	res := apipolicy.PolicyConfig{}
	annotation := map[string]*string{}
	annotation[string(annotationscommon.AnnotationProxyNextUpstream)] = &config.ServerConf.NextUpstreams
	nextTries := "4"
	annotation[string(annotationscommon.AnnotationProxyNextUpstreamRetries)] = &nextTries
	nextTimeout := "5"
	annotation[string(annotationscommon.AnnotationProxyNextUpstreamTimeOut)] = &nextTimeout
	snippet := `
proxy_intercept_errors on;
`
	res.IngressAnnotation = &apipolicy.IngressAnnotation{
		Annotation:      annotation,
		LocationSnippet: &snippet,
	}
	emptyStr := ""
	// trigger httpsnippet update
	res.IngressController = &apipolicy.IngressController{
		HttpSnippet: &emptyStr,
	}

	builtinPlugins := config.ServerConf.BuiltinPlugins
	gatewayAdapter, _, err := apipolicy.GetGatewayAdapter(ctx, apipolicy.Policy_Engine_Built_in)
	if err != nil {
		return res, err
	}

	adapter, ok := gatewayAdapter.(gateway_providers.GatewayAdapter)
	if !ok {
		return res, errors.Errorf("convert failed:%+v", adapter)
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
	plugins, err := policyDb.SelectByAny(&orm.GatewayPolicy{
		ZoneId:   zone.Id,
		Category: apipolicy.Policy_Category_BuiltIn,
	})
	if err != nil {
		return res, err
	}
	for _, plugin := range plugins {
		exist := false
		for _, name := range builtinPlugins {
			if plugin.PluginName == name {
				exist = true
			}
		}
		if !exist && !forValidate {
			err = adapter.RemovePlugin(plugin.PluginId)
			if err != nil {
				return res, err
			}
			_ = policyDb.DeleteById(plugin.Id)
			res.KongPolicyChange = true
		}
	}

	return res, nil
}

func (policy Policy) touchPluginIfNeed(zoneId string, builtinPlugins []string, pluginName string, config map[string]interface{}, adapter gateway_providers.GatewayAdapter) (bool, error) {
	enable := false
	for _, name := range builtinPlugins {
		if name == pluginName {
			enable = true
		}
	}
	if !enable {
		log.Infof("plugin not enable: %s", pluginName)
		return false, nil
	}
	disable := false
	req := &providerDto.PluginReqDto{
		Name:    pluginName,
		Config:  config,
		Enabled: &disable,
	}
	policyDb, _ := db.NewGatewayPolicyServiceImpl()
	exist, err := policyDb.GetByAny(&orm.GatewayPolicy{
		ZoneId:     zoneId,
		PluginName: pluginName,
	})
	if err != nil {
		return false, err
	}
	if exist != nil {
		req.Id = exist.PluginId
		resp, err := adapter.CreateOrUpdatePluginById(req)
		if err != nil {
			return false, err
		}
		configByte, err := json.Marshal(resp.Config)
		if err != nil {
			return false, err
		}
		exist.Config = configByte
		err = policyDb.Update(exist)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	resp, err := adapter.AddPlugin(req)
	if err != nil {
		return false, err
	}
	configByte, err := json.Marshal(resp.Config)
	if err != nil {
		return false, err
	}
	policyDao := &orm.GatewayPolicy{
		ZoneId:     zoneId,
		PluginName: pluginName,
		Category:   apipolicy.Policy_Category_BuiltIn,
		PluginId:   resp.Id,
		Config:     configByte,
		Enabled:    1,
	}
	err = policyDb.Insert(policyDao)
	if err != nil {
		return false, err
	}
	return true, nil
}

func init() {
	err := apipolicy.RegisterPolicyEngine(apipolicy.ProviderMSE, apipolicy.Policy_Engine_Built_in, &Policy{BasePolicy: &apipolicy.BasePolicy{}})
	if err != nil {
		panic(err)
	}
}
