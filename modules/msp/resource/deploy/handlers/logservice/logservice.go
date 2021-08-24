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

package logservice

import (
	"context"
	"fmt"

	"github.com/erda-project/erda/modules/msp/instance/db"
	"github.com/erda-project/erda/modules/msp/resource/deploy/handlers"
	"github.com/erda-project/erda/modules/msp/resource/utils"
)

const AddonLogIndexPrefix = "rlogs-"

func (p *provider) IsMatch(tmc *db.Tmc) bool {
	return tmc.Engine == handlers.ResourceLogService
}

func (p *provider) DoApplyTmcInstanceTenant(req *handlers.ResourceDeployRequest, resourceInfo *handlers.ResourceInfo,
	tmcInstance *db.Instance, tenant *db.InstanceTenant, clusterConfig map[string]string) (map[string]string, error) {

	err := p.createIndex(tenant.ID)
	if err != nil {
		return nil, err
	}

	config := map[string]string{}
	config["MSP_ENV"] = tenant.ID
	config["MSP_LOG_ATTACH"] = "true"

	params := map[string]string{
		"tenantId":    tenant.ID,
		"tenantGroup": tenant.TenantGroup,
		"logKey":      tenant.ID,
		"key":         "LogQuery",
	}
	paramstr, _ := utils.JsonConvertObjToString(params)
	config["PUBLIC_HOST"] = paramstr // params for microservice menu

	return config, nil
}

func (p *provider) DeleteTenant(tenant *db.InstanceTenant, tmcInstance *db.Instance, clusterConfig map[string]string) error {
	err := p.deleteIndex(tenant.ID)
	if err != nil {
		return err
	}

	return p.DefaultDeployHandler.DeleteTenant(tenant, tmcInstance, clusterConfig)
}

func (p *provider) createIndex(logKey string) error {
	// todo support multiple logServiceInstances
	logServiceInstance, err := p.LogServiceInstanceDB.GetFirst()
	if err != nil {
		return fmt.Errorf("failed to get log service instance: %s", err.Error())
	}

	if logServiceInstance == nil {
		return fmt.Errorf("no available log service instance")
	}

	esClient := utils.GetESClientFromLogService(logServiceInstance, "")

	if esClient == nil {
		return fmt.Errorf("can not build esclient")
	}

	indexPrefix := AddonLogIndexPrefix + logKey
	index := "<" + indexPrefix + "-{now/d{yyyy.ww}}-000001>"
	return esClient.CreateIndexWithAlias(index, indexPrefix)
}

func (p *provider) deleteIndex(logKey string) error {

	logServiceInstance, err := p.LogServiceInstanceDB.GetFirst()
	if err != nil {
		return fmt.Errorf("failed to get log service instance: %s", err.Error())
	}

	if logServiceInstance == nil {
		return fmt.Errorf("no available log service instance")
	}

	esClient := utils.GetESClientFromLogService(logServiceInstance, "")

	indexPrefix := AddonLogIndexPrefix + logKey
	indices := indexPrefix + "-*"

	resp, err := esClient.DeleteIndex(indices).Do(context.Background())
	if err != nil {
		return err
	}

	if !resp.Acknowledged {
		return fmt.Errorf("failed to delete index: response code error")
	}

	return nil
}
