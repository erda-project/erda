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
	kongDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/kong/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	db "github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
)

type Policy struct {
	apipolicy.BasePolicy
}

func (policy Policy) NeedSerialUpdate() bool {
	return true
}

func (policy Policy) CreateDefaultConfig(ctx map[string]interface{}) apipolicy.PolicyDto {
	return nil
}

func (policy Policy) UnmarshalConfig(config []byte) (apipolicy.PolicyDto, error, string) {
	return nil, nil, ""
}

// forValidate 用于识别解析的目的，如果解析是用来做鱼 nginx 配置冲突相关的校验，则关于数据表、调用 kong 接口的操作都不会执行
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

	value, ok := ctx[apipolicy.CTX_KONG_ADAPTER]
	if !ok {
		//TODO: MSE support built-in?
		log.Infof("Not use Kong Adapter, no need set built-in policy")
		return res, nil
	}
	gatewayAdapter, ok := value.(gateway_providers.GatewayAdapter)
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
		if !exist {
			if !forValidate {
				err = gatewayAdapter.RemovePlugin(plugin.PluginId)
				if err != nil {
					return res, err
				}
				_ = policyDb.DeleteById(plugin.Id)
				res.KongPolicyChange = true
			}
		}
	}
	if !forValidate {
		newPlugin, err := policy.touchPluginIfNeed(zone.Id, builtinPlugins, "spot-collector", map[string]interface{}{
			"send_port":          config.ServerConf.SpotSendPort,
			"addon_name":         config.ServerConf.SpotAddonName,
			"metric_name":        config.ServerConf.SpotMetricName,
			"tags_header_prefix": config.ServerConf.SpotTagsHeaderPrefix,
			"host_ip_key":        config.ServerConf.SpotHostIpKey,
			"instance_key":       config.ServerConf.SpotInstanceKey,
		}, gatewayAdapter)
		if err != nil {
			return res, err
		}
		if newPlugin {
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
	req := &kongDto.KongPluginReqDto{
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
	err := apipolicy.RegisterPolicyEngine(apipolicy.Policy_Engine_Built_in, &Policy{})
	if err != nil {
		panic(err)
	}
}
