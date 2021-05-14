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

package builtin

import (
	"encoding/json"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/hepa/apipolicy"
	"github.com/erda-project/erda/modules/hepa/config"
	"github.com/erda-project/erda/modules/hepa/kong"
	kongDto "github.com/erda-project/erda/modules/hepa/kong/dto"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
	db "github.com/erda-project/erda/modules/hepa/repository/service"
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

func (policy Policy) ParseConfig(dto apipolicy.PolicyDto, ctx map[string]interface{}) (apipolicy.PolicyConfig, error) {
	res := apipolicy.PolicyConfig{}
	annotation := map[string]*string{}
	annotation["nginx.ingress.kubernetes.io/proxy-next-upstream"] = &config.ServerConf.NextUpstreams
	nextTries := "4"
	annotation["nginx.ingress.kubernetes.io/proxy-next-upstream-tries"] = &nextTries
	nextTimeout := "5"
	annotation["nginx.ingress.kubernetes.io/proxy-next-upstream-timeout"] = &nextTimeout
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
		return res, errors.Errorf("get identify failed:%+v", ctx)
	}
	kongAdapter, ok := value.(kong.KongAdapter)
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
		Category: "built-in",
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
			err = kongAdapter.RemovePlugin(plugin.PluginId)
			if err != nil {
				return res, err
			}
			_ = policyDb.DeleteById(plugin.Id)
			res.KongPolicyChange = true
		}
	}
	newPlugin, err := policy.touchPluginIfNeed(zone.Id, builtinPlugins, "spot-collector", map[string]interface{}{
		"send_port":          config.ServerConf.SpotSendPort,
		"addon_name":         config.ServerConf.SpotAddonName,
		"metric_name":        config.ServerConf.SpotMetricName,
		"tags_header_prefix": config.ServerConf.SpotTagsHeaderPrefix,
		"host_ip_key":        config.ServerConf.SpotHostIpKey,
		"instance_key":       config.ServerConf.SpotInstanceKey,
	}, kongAdapter)
	if err != nil {
		return res, err
	}
	if newPlugin {
		res.KongPolicyChange = true
	}
	return res, nil

}

func (policy Policy) touchPluginIfNeed(zoneId string, builtinPlugins []string, pluginName string, config map[string]interface{}, adapter kong.KongAdapter) (bool, error) {
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
		Category:   "built-in",
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
	err := apipolicy.RegisterPolicyEngine("built-in", &Policy{})
	if err != nil {
		panic(err)
	}
}
