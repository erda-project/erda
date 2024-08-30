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

package coordinator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	mutex "github.com/erda-project/erda-infra/providers/etcd-mutex"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/msp/instance/db"
	"github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers"
)

type Interface interface {
	CheckIfNeedRealDeploy(req handlers.ResourceDeployRequest) (bool, error)
	Deploy(req handlers.ResourceDeployRequest) (*handlers.ResourceDeployResult, error)
	UnDeploy(resourceId string) error
}

func (p *provider) findHandler(tmc *db.Tmc) handlers.ResourceDeployHandler {
	for _, handler := range p.handlers {
		if handler.IsMatch(tmc) {
			return handler
		}
	}
	return nil
}

func (p *provider) CheckIfNeedRealDeploy(req handlers.ResourceDeployRequest) (bool, error) {
	resourceInfo, err := p.defaultHandler.GetResourceInfo(&req)
	if err != nil {
		return false, err
	}

	handler := p.findHandler(resourceInfo.Tmc)
	if handler == nil {
		return false, fmt.Errorf("could not find deploy handler for %s", req.Engine)
	}

	// pre-check if need do further deploy
	_, needDeployInstance, err := handler.CheckIfNeedTmcInstance(&req, resourceInfo)
	if err != nil {
		return false, err
	}

	// if addon has no services or depend addons, no real deploy will perform
	hasServices := resourceInfo.Dice != nil && resourceInfo.Dice.Services != nil && len(resourceInfo.Dice.Services) > 0
	hasAddons := resourceInfo.Dice != nil && resourceInfo.Dice.AddOns != nil && len(resourceInfo.Dice.AddOns) > 0

	return needDeployInstance && (hasServices || hasAddons), err
}

func (p *provider) Deploy(req handlers.ResourceDeployRequest) (*handlers.ResourceDeployResult, error) {
	return p.mutexDeploy(p.deploy, req)
}

// deploy is the main logic for deploying addon
func (p *provider) deploy(req handlers.ResourceDeployRequest) (*handlers.ResourceDeployResult, error) {
	var result handlers.ResourceDeployResult

	// get resource info : tmc + extension info
	resourceInfo, err := p.defaultHandler.GetResourceInfo(&req)
	defer func() {
		// callback to orchestrator
		if len(req.Callback) == 0 {
			return
		}
		if err != nil {
			p.defaultHandler.Callback(req.Callback, req.Uuid, false, nil, req.Options, err.Error())
		} else {
			p.defaultHandler.Callback(req.Callback, result.ID, true, result.Config, result.Options, "")
		}
	}()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to defaultHandler.GetResourceInfo for %s:%s/%s", req.Engine, req.Options["version"], req.Az)
	}

	handler := p.findHandler(resourceInfo.Tmc)
	if handler == nil {
		return nil, fmt.Errorf("could not find deploy handler for %s", req.Engine)
	}

	// pre-check if it needs to further deploy
	tmcInstance, needDeployInstance, err := handler.CheckIfNeedTmcInstance(&req, resourceInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to CheckIfNeedTmcInstance for %s/%s", req.Engine, req.Az)
	}
	tenant, needApplyTenant, err := handler.CheckIfNeedTmcInstanceTenant(&req, resourceInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to CheckIfNeedTmcInstanceTenant for %s/%s", req.Engine, req.Az)
	}
	p.Log.Infof("[%s/%s] check if it needs to deploy tmc instance: %v,  needs to apply tenant: %v\n",
		req.Engine, req.Az, needDeployInstance, needApplyTenant)

	// create tmc_instance record if necessary
	var clusterConfig map[string]string
	clusterConfig, err = handler.GetClusterConfig(req.Az)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to GetClusterConfig(%s)", req.Az)
	}

	var subResults []*handlers.ResourceDeployResult
	// resolve dependency resources
	if needApplyTenant || needDeployInstance {
		// for some resource like monitor, do not has dice.yml definition
		if resourceInfo.Dice != nil && resourceInfo.Dice.AddOns != nil {
			defer func() {
				// delete related sub resources if error occur
				if err == nil {
					return
				}

				for _, subResult := range subResults {
					p.UnDeploy(subResult.ID)
					handler.DeleteRequestRelation(req.Uuid, subResult.ID)
				}
			}()

			// if use mse or other, it will reset addons to mse-nacos
			handler.ResetAddons(resourceInfo, clusterConfig)

			for name, addon := range resourceInfo.Dice.AddOns {
				if !isNeedDeployGatewayDependedAddon(resourceInfo.Spec.Type, resourceInfo.Spec.Name, clusterConfig) {
					// 部署的是 api-gateway 且网关用 MSE， 则无需部署 网关 api-gateway 对应的 dice.yml 中依赖的 addons (postgresql)
					break
				}

				// deploy dependency resource recursive
				subReq := handler.BuildSubResourceDeployRequest(name, addon, &req)
				if subReq == nil {
					continue
				}
				var subResult *handlers.ResourceDeployResult
				subResult, err = p.Deploy(*subReq)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to Deploy sub addon %s:%s/%s", subReq.Engine, subReq.Options["version"], subReq.Az)
				}
				subResults = append(subResults, subResult)
				handler.BuildRequestRelation(req.Uuid, subResult.ID)
			}
		}
	}

	if needDeployInstance {
		// initialize tmc_instance
		tmcInstance, err = handler.InitializeTmcInstance(&req, resourceInfo, subResults)
		if err != nil {
			return nil, errors.Wrap(err, "failed to InitializeTmcInstance")
		}
		defer func() {
			// delete instance if error occur,
			// if tmcInstance status is RUNNING skip delete even if error
			if err == nil || tmcInstance.Status == handlers.TmcInstanceStatusRunning {
				return
			}

			handler.DeleteTmcInstance(tmcInstance, handlers.TmcInstanceStatusError)
		}()

		// if is custom resource, do not real deploy, just update config and simply mark status as RUNNING
		customConfig, hasCustom := handler.CheckIfHasCustomConfig(clusterConfig)
		if hasCustom {
			handler.UpdateTmcInstanceOnCustom(tmcInstance, customConfig)
		} else {
			// do pre-deploy job if any
			if err = handler.DoPreDeployJob(resourceInfo, tmcInstance); err != nil {
				return nil, errors.Wrap(err, "failed to DoPreDeployJob")
			}

			// do deploy and wait for ready
			var sgDeployResult interface{}
			if resourceInfo.Dice == nil || resourceInfo.Dice.Services == nil || len(resourceInfo.Dice.Services) == 0 {
				// some resource do not need real deploy, e.g. configcenter.
				// this kind of resource do not have services section defined in dice.yml
				// just mock a success response
				sgDeployResult = &apistructs.ServiceGroup{
					StatusDesc: apistructs.StatusDesc{Status: apistructs.StatusReady},
				}
			} else {
				sgReq := handler.BuildServiceGroupRequest(resourceInfo, tmcInstance, clusterConfig)
				sgDeployResult, err = handler.DoDeploy(sgReq, resourceInfo, tmcInstance, clusterConfig)
				if err != nil {
					return nil, errors.Wrap(err, "failed to DoDeploy")
				}
			}

			// do post-deploy job if any
			additionalConfig := map[string]string{}
			additionalConfig, err = handler.DoPostDeployJob(tmcInstance, sgDeployResult, clusterConfig)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to DoPostDeployJob for tmc_instance %+v", tmcInstance)
			}

			// update tmc_instance config and status
			config := handler.BuildTmcInstanceConfig(tmcInstance, sgDeployResult, clusterConfig, additionalConfig)
			handler.UpdateTmcInstanceOnFinish(tmcInstance, config, handlers.TmcInstanceStatusRunning)
		}
	}

	if needApplyTenant {
		// create tmc_instance_tenant record
		tenant, err = handler.InitializeTmcInstanceTenant(&req, tmcInstance, subResults)
		if err != nil {
			return nil, errors.Wrap(err, "failed to InitializeTmcInstanceTenant")
		}
		defer func() {
			// delete tenant if error occur
			if err == nil {
				return
			}
			handler.DeleteTenant(tenant, tmcInstance, clusterConfig)
		}()

		// deploy tmc_instance_tenant
		var config map[string]string
		config, err = handler.DoApplyTmcInstanceTenant(&req, resourceInfo, tmcInstance, tenant, clusterConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to DoApplyTmcInstanceTenant for %+v", tmcInstance)
		}

		// update and persistent applied config
		tenant, err = handler.UpdateTmcInstanceTenantOnFinish(tenant, config)
		if err != nil {
			return nil, errors.Wrap(err, "failed to UpdateTmcInstanceTenantOnFinish")
		}
	}

	result = handler.BuildDeployResult(tmcInstance, tenant)

	return &result, nil
}

func (p *provider) UnDeploy(resourceId string) error {
	// judge if is tenant
	// call scheduler if need
	// mark tmc_instance as deleted
	// delete sub resources recursive if any
	// delete request relation record

	tmcInstance, tenant, tmc, dice, err := p.defaultHandler.GetInstanceById(resourceId)
	if err != nil {
		return err
	}
	if tmcInstance == nil && tenant == nil {
		return fmt.Errorf("resource not found")
	}

	handler := p.findHandler(tmc)
	if handler == nil {
		return fmt.Errorf("handler not found")
	}

	if tmc.DeployMode == handlers.DeployModePAAS {
		if dice != nil && dice.Services != nil && len(dice.Services) > 0 {
			handler.UnDeploy(tmcInstance)
		}
		handler.DeleteTmcInstance(tmcInstance, handlers.TmcInstanceStatusDeleted)
	} else {
		clusterConfig, err := handler.GetClusterConfig(tmcInstance.Az)
		if err != nil {
			return err
		}
		handler.DeleteTenant(tenant, tmcInstance, clusterConfig)
	}

	parentId := tmcInstance.ID
	if tenant != nil {
		parentId = tenant.ID
	}

	childIds := handler.GetRelationResourceIds(parentId)
	for _, childId := range childIds {
		p.UnDeploy(childId)
		_ = handler.DeleteRequestRelation(parentId, childId)
	}

	return nil
}

const (
	ResourceRegisterCenter = "registercenter"
	ResourceConfigCenter   = "config-center"

	RegisterAndConfigMutexKey = "ConfigLoveRegister"
)

// mutexDeploy wraps a global distributed lock around the function deploy to ensure that only one instance of an addon of a certain type of engine is being
// deployed within a cluster, to prevent duplicate instances from being pulled up.
// deploy is the main logic for deployment.
func (p *provider) mutexDeploy(deploy func(request handlers.ResourceDeployRequest) (*handlers.ResourceDeployResult, error), req handlers.ResourceDeployRequest) (*handlers.ResourceDeployResult, error) {
	// get the lock at the "<addon engine>/<cluster>" granularity before deploying to avoid duplicate instances
	p.Log.Infof("[%s/%s] to get the ETCD Mutex", req.Engine, req.Az)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*12)
	defer cancel()

	var err error
	var mu mutex.Mutex

	if req.Engine == ResourceRegisterCenter || req.Engine == ResourceConfigCenter {
		mu, err = p.Mutex.New(ctx, strings.Join([]string{RegisterAndConfigMutexKey, req.Az}, "/"))
	} else {
		mu, err = p.Mutex.New(ctx, strings.Join([]string{req.Engine, req.Az}, "/"))
	}
	if err != nil {
		p.Log.Errorf("[%s/%s] failed to New a global distributed lock (ETCD Mutex) before deploying: %v\n", req.Engine, req.Az, err)
		return nil, errors.Wrapf(err, "failed to New ETCD Mutex for %s/%s", req.Engine, req.Az)
	}
	defer func() {
		if mu != nil {
			p.Log.Infof("[%s/%s] to release ETCD Mutex\n", req.Engine, req.Az)
			if err := mu.Unlock(ctx); err != nil {
				p.Log.Errorf("[%s/%s] failed to Unlock the global distributed lock (ETCD Mutex) after deployed: %v\n", req.Engine, req.Az, err)
			}
			if err := mu.Close(); err != nil {
				p.Log.Errorf("[%s/%s] failed to Close the global distributed lock (ETCD Mutex) after deployed: %v\n", req.Engine, req.Az, err)
			}
		}
	}()
	if err = mu.Lock(ctx); err != nil {
		p.Log.Errorf("[%s/%s] failed to Lock: %v\n", req.Engine, req.Az, err)
		return nil, errors.Wrapf(err, "failed to Lock for %s/%s", req.Engine, req.Az)
	}
	p.Log.Infof("[%s/%s] to deploy: %+v\n", req.Engine, req.Az, req)
	return deploy(req)
}

func isNeedDeployGatewayDependedAddon(resourceSpecType, resourceSpecName string, clusterConfig map[string]string) bool {
	if resourceSpecType == "addon" && resourceSpecName == apistructs.AddonApiGateway {
		gatewayProvider, ok := clusterConfig[handlers.GatewayProviderVendorKey]
		if ok && gatewayProvider != "" {
			// 部署的是 api-gateway 且网关用 MSE， 则无需部署 网关 api-gateway 对应的 dice.yml 中依赖的 addons (postgresql)
			return false
		}
	}
	return true
}
