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
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
	gateway_providers "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers"
	providerDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	mse "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	db "github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
)

func init() {
	if err := apipolicy.RegisterPolicyEngine(apipolicy.ProviderMSE, apipolicy.Policy_Engine_SBAC, &Policy{BasePolicy: &apipolicy.BasePolicy{}}); err != nil {
		panic(err)
	}
}

type Policy struct {
	*apipolicy.BasePolicy
}

func (policy Policy) CreateDefaultConfig(ctx map[string]interface{}) apipolicy.PolicyDto {
	return new(PolicyDto)
}

func (policy Policy) UnmarshalConfig(config []byte) (apipolicy.PolicyDto, error, string) {
	var policyDto PolicyDto
	if err := json.Unmarshal(config, &policyDto); err != nil {
		return nil, errors.Wrapf(err, "failed to Unmarshal config: %s", string(config)), "Invalid config"
	}
	if err := policyDto.IsValidDto(); err != nil {
		return nil, errors.Wrap(err, "Invalid config"), "Invalid config"
	}
	return &policyDto, nil, ""
}

func (policy Policy) buildPluginReq(dto *PolicyDto, zoneName string) *providerDto.PluginReqDto {
	return dto.ToPluginReqDto(zoneName)
}

// forValidate 用于识别解析的目的，如果解析是用来做 nginx 配置冲突相关的校验，则关于数据表、调用 kong 接口的操作都不会执行
func (policy Policy) ParseConfig(dto apipolicy.PolicyDto, ctx map[string]interface{}, forValidate bool) (apipolicy.PolicyConfig, error) {
	l := logrus.WithField("pluginName", apipolicy.Policy_Engine_SBAC).WithField("func", "ParseConfig")
	l.Infof("dto: %+v", dto)
	res := apipolicy.PolicyConfig{}
	policyDto, ok := dto.(*PolicyDto)
	if !ok {
		return res, errors.Errorf("invalid config:%+v", dto)
	}
	logrus.Infof("sbac policyDto=%+v, ctx=%v", *policyDto, ctx)

	gatewayAdapter, _, err := apipolicy.GetGatewayAdapter(ctx, apipolicy.Policy_Engine_SBAC)
	if err != nil {
		return res, err
	}

	adapter, ok := gatewayAdapter.(gateway_providers.GatewayAdapter)
	if !ok {
		l.Infof("get gateway adpater for set sbac policy failed")
		return res, nil
	}

	gatewayVersion, err := adapter.GetVersion()
	if err != nil {
		return res, errors.Wrap(err, "failed to retrieve gateway version")
	}
	if !strings.HasPrefix(gatewayVersion, "2.") && !strings.HasPrefix(gatewayVersion, "mse-") {
		return res, errors.Errorf("the plugin %s is not supportted on the gateway version %s", apipolicy.Policy_Engine_SBAC, gatewayVersion)
	}

	zone, ok := ctx[apipolicy.CTX_ZONE].(*orm.GatewayZone)
	if !ok {
		return res, errors.Errorf("failed to get identify with %s: %+v", apipolicy.CTX_ZONE, ctx)
	}
	logrus.Infof("set sbac policy for zone=%+v", *zone)
	policyDb, _ := db.NewGatewayPolicyServiceImpl()
	exist, err := policyDb.GetByAny(&orm.GatewayPolicy{
		ZoneId:     zone.Id,
		PluginName: apipolicy.Policy_Engine_SBAC,
	})
	if err != nil {
		return res, err
	}
	if !policyDto.Switch && !forValidate {
		if exist != nil {
			// 创建网关路由也会走到这里，但这个时候 MSE 那边应该还没有发现到新路由，因此调用此插件会失败，但创建过程又不能一直等待，因此这里做个异步处理
			if zone.Type == db.ZONE_TYPE_PACKAGE_API {
				resp, err := adapter.CreateOrUpdatePluginById(policy.buildPluginReq(policyDto, strings.ToLower(zone.Name)))
				if err != nil {
					go apipolicy.NonSwitchUpdateMSEPluginConfig(adapter, policy.buildPluginReq(policyDto, strings.ToLower(zone.Name)), zone.Name, mse.MsePluginSbac)
				} else {
					logrus.Infof("create or update mse erda-sbac plugin with response: %+v", *resp)
				}
			}

			_ = policyDb.DeleteById(exist.Id)
			res.KongPolicyChange = true
		}
		return res, nil
	}

	req := policy.buildPluginReq(policyDto, strings.ToLower(zone.Name))

	packageAPIDB, _ := db.NewGatewayPackageApiServiceImpl()
	packageApi, err := packageAPIDB.GetByAny(&orm.GatewayPackageApi{ZoneId: zone.Id})
	if err != nil {
		l.WithError(err).Warnf("sbac: packageAPIDB.GetByAny(&orm.GatewayPackageApi{ZoneId: %s})", zone.Id)
		return res, err
	}
	if packageApi == nil {
		l.WithError(err).Warnf("sbac: not found packageAPIDB.GetByAny(&orm.GatewayPackageApi{ZoneId: %s})", zone.Id)
		return res, nil
	}

	if exist != nil {
		if !forValidate {
			req.Id = exist.PluginId
			resp, err := adapter.CreateOrUpdatePluginById(req)
			if err != nil {
				return res, err
			}
			configByte, err := json.Marshal(resp.Config)
			if err != nil {
				return res, err
			}
			exist.Config = configByte
			logrus.Infof("update exist sbac gateway policy=%+v", *exist)
			err = policyDb.Update(exist)
			if err != nil {
				return res, err
			}
		}
	} else {
		if !forValidate {
			var resp *providerDto.PluginRespDto
			// 如果是 MSE 网关，则仍然是 CreateOrUpdatePluginById
			if zone.Type == db.ZONE_TYPE_PACKAGE_API {
				resp, err = adapter.CreateOrUpdatePluginById(policy.buildPluginReq(policyDto, strings.ToLower(zone.Name)))
				if err != nil {
					return res, err
				}
			}

			configByte, err := json.Marshal(resp.Config)
			if err != nil {
				return res, err
			}
			policyDao := &orm.GatewayPolicy{
				ZoneId:     zone.Id,
				PluginName: apipolicy.Policy_Engine_SBAC,
				Category:   apipolicy.Policy_Category_Safety,
				PluginId:   resp.Id,
				Config:     configByte,
				Enabled:    1,
			}
			logrus.Infof("Insert sbac gateway policy=%+v", *policyDao)
			err = policyDb.Insert(policyDao)
			if err != nil {
				return res, err
			}
			res.KongPolicyChange = true
		}
	}
	return res, nil
}
