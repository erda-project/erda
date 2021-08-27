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
	"time"

	"github.com/erda-project/erda/modules/msp/instance/db"
	"github.com/erda-project/erda/modules/msp/resource/deploy/handlers"
	"github.com/erda-project/erda/modules/msp/resource/utils"
)

const AddonLogIndexPrefix = "rlogs-"

func (p *provider) IsMatch(tmc *db.Tmc) bool {
	return tmc.Engine == handlers.ResourceLogService
}

func (p *provider) DoPostDeployJob(tmcInstance *db.Instance, serviceGroupDeployResult interface{}, clusterConfig map[string]string) (map[string]string, error) {
	config := map[string]string{}
	options := map[string]string{}
	utils.JsonConvertObjToType(tmcInstance.Options, &options)

	// reuse if already assign service_instance
	orgId := options["orgId"]
	if len(orgId) == 0 {
		return nil, fmt.Errorf("orgId is empty")
	}
	deployment, _ := p.LogDeploymentDb.GetByClusterNameAndOrgId(tmcInstance.Az, orgId, db.LogTypeLogService)
	if deployment != nil {
		return config, nil
	}

	deployment, _ = p.LogDeploymentDb.GetByOrgId(orgId, db.LogTypeLogService)
	if deployment != nil {
		deployment.ID = 0
		deployment.ClusterName = tmcInstance.Az
		if err := p.LogDeploymentDb.Save(&deployment).Error; err != nil {
			return nil, err
		}
		return config, nil
	}

	// todo: algorithm to decide which logServiceInstance to use
	logServiceInstance, err := p.LogServiceInstanceDB.GetFirst()
	if err != nil {
		return nil, fmt.Errorf("failed to get log service instance: %s", err.Error())
	}

	deployment = &db.LogDeployment{
		OrgId:       orgId,
		ClusterName: tmcInstance.Az,
		EsUrl:       logServiceInstance.EsUrls,
		EsConfig:    logServiceInstance.EsConfig,
		Created:     time.Now(),
		Updated:     time.Now(),
		LogType:     string(db.LogTypeLogService),
	}

	if err = p.LogDeploymentDb.Save(deployment).Error; err != nil {
		return nil, err
	}

	return config, nil
}

func (p *provider) DoApplyTmcInstanceTenant(req *handlers.ResourceDeployRequest, resourceInfo *handlers.ResourceInfo,
	tmcInstance *db.Instance, tenant *db.InstanceTenant, clusterConfig map[string]string) (map[string]string, error) {

	options := map[string]string{}

	instanceOptions := map[string]string{}
	utils.JsonConvertObjToType(tmcInstance.Options, &instanceOptions)
	utils.AppendMap(options, instanceOptions)

	tenantOptions := map[string]string{}
	utils.JsonConvertObjToType(tenant.Options, &tenantOptions)
	utils.AppendMap(options, tenantOptions)

	logKey := tenant.ID
	config := map[string]string{}
	config["MSP_ENV"] = logKey
	config["MSP_LOG_ATTACH"] = "true"
	configStr, _ := utils.JsonConvertObjToString(config)

	instance, err := p.LogInstanceDb.GetLatestByLogKey(logKey, db.LogTypeLogService)
	if err != nil {
		return nil, err
	}

	// create if not exists
	if instance == nil {
		err = p.createIndex(options["orgId"], tmcInstance.Az, tenant.ID)
		if err != nil {
			return nil, err
		}

		instance = &db.LogInstance{
			LogKey:          logKey,
			ClusterName:     tmcInstance.Az,
			OrgId:           options["orgId"],
			OrgName:         options["orgName"],
			ProjectId:       options["projectId"],
			ProjectName:     options["projectName"],
			Workspace:       options["workspace"],
			ApplicationId:   options["applicationId"],
			ApplicationName: options["applicationName"],
			RuntimeId:       options["runtimeId"],
			RuntimeName:     options["runtimeName"],
			Plan:            "",
			IsDelete:        0,
			Version:         options["version"],
			Config:          configStr,
			Created:         time.Now(),
			Updated:         time.Now(),
			LogType:         string(db.LogTypeLogService),
		}
		if err = p.InstanceDb.Save(instance).Error; err != nil {
			return nil, err
		}
	}

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
	logInstance, _ := p.LogInstanceDb.GetLatestByLogKey(tenant.ID, db.LogTypeLogService)
	if logInstance != nil && len(logInstance.ClusterName) > 0 {
		p.LogInstanceDb.Model(logInstance).Update("is_delete", 1)
		p.deleteIndex(logInstance.OrgId, tenant.ID, logInstance.ClusterName)
	}

	return p.DefaultDeployHandler.DeleteTenant(tenant, tmcInstance, clusterConfig)
}

func (p *provider) createIndex(orgId string, clusterName string, logKey string) error {
	logDeployment, err := p.LogDeploymentDb.GetByClusterNameAndOrgId(clusterName, orgId, db.LogTypeLogService)
	if err != nil {

		return err
	}

	if logDeployment == nil {
		return fmt.Errorf("could not found logDeployment for cluster: %s, orgId: %s", clusterName, orgId)
	}

	esClient := utils.GetESClientsFromLogAnalytics(logDeployment, logKey)
	if esClient == nil {
		return fmt.Errorf("can not build esclient")
	}

	indexPrefix := AddonLogIndexPrefix + logKey
	index := "<" + indexPrefix + "-{now/d{yyyy.ww}}-000001>"
	var orgAlias string
	if len(orgId) > 0 {
		orgAlias = AddonLogIndexPrefix + orgId
		esClient.CreateIndexTemplate(indexPrefix, indexPrefix+"-*", orgAlias)
	}

	return esClient.CreateIndexWithAlias(index, indexPrefix, orgAlias)
}

func (p *provider) deleteIndex(orgId string, logKey string, clusterName string) error {
	logDeployment, err := p.LogDeploymentDb.GetByClusterNameAndOrgId(clusterName, orgId, db.LogTypeLogService)
	if err != nil {
		return err
	}

	if logDeployment == nil {
		return fmt.Errorf("could not found logDeployment for cluster: %s, orgId: %s", clusterName, orgId)
	}

	esClient := utils.GetESClientsFromLogAnalytics(logDeployment, logKey)
	if esClient == nil {
		return fmt.Errorf("can not build esclient")
	}

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
