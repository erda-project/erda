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

package loganalytics

import (
	"context"
	"fmt"
	"time"

	"github.com/erda-project/erda/modules/msp/instance/db"
	"github.com/erda-project/erda/modules/msp/resource/deploy/handlers"
	"github.com/erda-project/erda/modules/msp/resource/utils"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

const AddonLogIndexPrefix = "rlogs-"

func (p *provider) IsMatch(tmc *db.Tmc) bool {
	return tmc.Engine == handlers.ResourceLogAnalytics
}

func (p *provider) BuildSubResourceDeployRequest(name string, addon *diceyml.AddOn, req *handlers.ResourceDeployRequest) *handlers.ResourceDeployRequest {
	deployment, _ := p.LogDeploymentDb.GetByClusterName(req.Az)
	if deployment != nil {
		// already deployed outside, no need to do further deploy
		orgId := req.Options["orgId"]
		orgDeployment, _ := p.LogDeploymentDb.GetByClusterNameAndOrgId(req.Az, orgId)
		if orgDeployment == nil {
			orgDeployment = deployment
			orgDeployment.ID = 0
			orgDeployment.OrgId = orgId
			p.LogDeploymentDb.Save(&orgDeployment)
		}

		return nil
	}

	return p.DefaultDeployHandler.BuildSubResourceDeployRequest(name, addon, req)
}

func (p *provider) BuildTmcInstanceConfig(tmcInstance *db.Instance, serviceGroupDeployResult interface{},
	clusterConfig map[string]string, additionalConfig map[string]string) map[string]string {
	config := map[string]string{}
	options := map[string]string{}
	utils.JsonConvertObjToType(tmcInstance.Options, &options)

	// if already deployed outside, just return ????
	orgId := options["orgId"]
	if len(orgId) == 0 {
		orgId = "0"
	}
	deployment, _ := p.LogDeploymentDb.GetByClusterName(tmcInstance.Az)
	if deployment != nil {
		// already deployed outside, no need to do further deploy
		orgDeployment, _ := p.LogDeploymentDb.GetByClusterNameAndOrgId(tmcInstance.Az, orgId)
		if orgDeployment == nil {
			orgDeployment = deployment
			orgDeployment.ID = 0
			orgDeployment.OrgId = orgId
			p.LogDeploymentDb.Save(&orgDeployment)
		}

		return config
	}

	if options["MONITOR_LOG_OUTPUT"] == "elasticsearch-rollover" {
		p.createLogDeployment(orgId, tmcInstance.Az, options["ES_URLS"], options["MONITOR_LOG_COLLECTOR"], clusterConfig)
		config["MONITOR_LOG_OUTPUT"] = "elasticsearch-rollover"
		config["MONITOR_LOG_OUTPUT_CONFIG"] = "-"
		config["MONITOR_LOG_COLLECTOR"] = options["MONITOR_LOG_COLLECTOR"]
	}

	return config
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

	collector := instanceOptions["MONITOR_LOG_COLLECTOR"]
	if len(collector) == 0 {
		dp, _ := p.LogDeploymentDb.GetByClusterName(tmcInstance.Az)
		if dp != nil {
			collector = dp.CollectorUrl
		}
	}

	config, err := p.createLogAnalytics(tenant.ID, tmcInstance.Az, options)
	if err != nil {
		return nil, err
	}
	config["MONITOR_LOG_KEY"] = tenant.ID
	config["MONITOR_LOG_OUTPUT"] = "elasticsearch-rollover"
	config["MONITOR_LOG_OUTPUT_CONFIG"] = "-"
	config["MONITOR_LOG_COLLECTOR"] = collector

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

	logInstance, _ := p.LogInstanceDb.GetLatestByLogKey(tenant.ID)
	if logInstance != nil && len(logInstance.ClusterName) > 0 {
		p.LogInstanceDb.Model(logInstance).Update("is_delete", 1)
		p.deleteIndex(tenant.ID, logInstance.ClusterName)
	}

	return p.DefaultDeployHandler.DeleteTenant(tenant, tmcInstance, clusterConfig)
}

func (p *provider) createLogDeployment(orgId string, clusterName string, esUrls string, collector string, clusterConfig map[string]string) {
	domain := clusterConfig["DICE_ROOT_DOMAIN"]
	clusterType := 0
	if clusterConfig["DICE_IS_EDGE"] == "true" {
		clusterType = 1
	}

	deploy := db.LogDeployment{
		OrgId:        orgId,
		ClusterName:  clusterName,
		ClusterType:  clusterType,
		EsUrl:        esUrls,
		EsConfig:     "{}",
		CollectorUrl: collector,
		Domain:       domain,
		Created:      time.Now(),
		Updated:      time.Now(),
	}

	p.LogDeploymentDb.Save(&deploy)
}

func (p *provider) createLogAnalytics(logKey string, clusterName string, options map[string]string) (map[string]string, error) {
	p.createIndex(clusterName, logKey)

	config := map[string]string{
		"TERMINUS_LOG_KEY": logKey,
	}

	instance, err := p.LogInstanceDb.GetLatestByLogKey(logKey)
	if err != nil {
		return nil, err
	}
	if instance != nil {
		return config, nil
	}

	configStr, _ := utils.JsonConvertObjToString(config)
	instance = &db.LogInstance{
		LogKey:          logKey,
		ClusterName:     clusterName,
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
	}
	if err := p.InstanceDb.Save(instance).Error; err != nil {
		return nil, err
	}

	return config, nil
}

func (p *provider) createIndex(clusterName string, logKey string) error {
	logDeployment, err := p.LogDeploymentDb.GetByClusterName(clusterName)
	if err != nil {
		return err
	}

	esClient := utils.GetESClientsFromLogAnalytics(logDeployment, logKey)
	if esClient == nil {
		return fmt.Errorf("can not build esclient")
	}

	indexPrefix := AddonLogIndexPrefix + logKey
	index := "<" + indexPrefix + "-{now/d{yyyy.ww}}-000001>"
	return esClient.CreateIndexWithAlias(index, indexPrefix)
}

func (p *provider) deleteIndex(logKey string, clusterName string) error {
	logDeployment, err := p.LogDeploymentDb.GetByClusterName(clusterName)
	if err != nil {
		return err
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
		return fmt.Errorf("response code error")
	}

	return nil
}
