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

package apigateway

import (
	"fmt"
	"strings"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/msp/instance/db"
	"github.com/erda-project/erda/modules/msp/resource/deploy/handlers"
	"github.com/erda-project/erda/modules/msp/resource/utils"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func (p *provider) IsMatch(tmc *db.Tmc) bool {
	return tmc.Engine == handlers.ResourceApiGateway
}

func (p *provider) DoPreDeployJob(resourceInfo *handlers.ResourceInfo, tmcInstance *db.Instance) error {
	instanceOptions := map[string]string{}
	utils.JsonConvertObjToType(tmcInstance.Options, &instanceOptions)

	jobReq := apistructs.JobFromUser{
		Name:        tmcInstance.ID,
		Image:       resourceInfo.Dice.Services["api-gateway-0"].Image,
		CPU:         0.1,
		Memory:      512,
		Namespace:   "gateway",
		Cmd:         "kong migrations bootstrap",
		ClusterName: tmcInstance.Az,
		Env: map[string]string{
			"KONG_DATABASE":                 "postgres",
			"KONG_PG_HOST":                  instanceOptions["POSTGRESQL_HOST"],
			"KONG_CASSANDRA_CONTACT_POINTS": instanceOptions["POSTGRESQL_HOST"],
			"KONG_PG_PORT":                  instanceOptions["POSTGRESQL_PORT"],
			"KONG_PG_USER":                  instanceOptions["POSTGRESQL_USER"],
			"KONG_PG_PASSWORD":              instanceOptions["POSTGRESQL_PASSWORD"],
			"KONG_PG_DATABASE":              instanceOptions["POSTGRESQL_DATABASE"],
		},
		Labels: map[string]string{},
	}

	_, err := p.Bdl.CreateJob(jobReq)
	if err != nil {
		return err
	}

	_, err = p.Bdl.StartJob(jobReq.Namespace, jobReq.Name)
	if err != nil {
		return err
	}

	startTime := time.Now().Unix()
	status := apistructs.StatusUnknown
	for time.Now().Unix()-startTime < handlers.RuntimeMaxUpTimeoutSeconds {
		time.Sleep(10 * time.Second)

		status, err = p.Bdl.GetJobStatus(jobReq.Namespace, jobReq.Name)
		if err != nil {
			continue
		}

		if status == apistructs.StatusStoppedOnOK ||
			status == apistructs.StatusStoppedOnFailed ||
			status == apistructs.StatusStoppedByKilled {
			break
		}
	}

	p.Bdl.StopJob(jobReq.Namespace, jobReq.Name)
	p.Bdl.DeleteJob(jobReq.Namespace, jobReq.Name)

	switch status {
	case apistructs.StatusStoppedOnOK:
		return nil
	case apistructs.StatusStoppedOnFailed, apistructs.StatusStoppedByKilled:
		return fmt.Errorf("init job run failed")
	default:
		return fmt.Errorf("init job run timeout")
	}
}

func (p *provider) BuildServiceGroupRequest(resourceInfo *handlers.ResourceInfo, tmcInstance *db.Instance, clusterConfig map[string]string) interface{} {
	req := p.DefaultDeployHandler.BuildServiceGroupRequest(resourceInfo, tmcInstance, clusterConfig).(*apistructs.ServiceGroupCreateV2Request)

	delete(req.GroupLabels, "LOCATION-CLUSTER-SERVICE")

	instanceOptions := map[string]string{}
	utils.JsonConvertObjToType(tmcInstance.Options, &instanceOptions)

	for name, service := range resourceInfo.Dice.Services {
		delete(service.Envs, "LOCATION-CLUSTER-SERVICE")

		// envs
		nodeId := tmcInstance.ID + "_" + name
		env := map[string]string{
			"ADDON_ID":                      tmcInstance.ID,
			"ADDON_NODE_ID":                 nodeId,
			"DICE_CLUSTER":                  tmcInstance.Az,
			"KONG_PG_HOST":                  instanceOptions["POSTGRESQL_HOST"],
			"KONG_CASSANDRA_CONTACT_POINTS": instanceOptions["POSTGRESQL_HOST"],
			"KONG_PG_PORT":                  instanceOptions["POSTGRESQL_PORT"],
			"KONG_PG_USER":                  instanceOptions["POSTGRESQL_USER"],
			"KONG_PG_PASSWORD":              instanceOptions["POSTGRESQL_PASSWORD"],
			"KONG_PG_DATABASE":              instanceOptions["POSTGRESQL_DATABASE"],
		}
		utils.AppendMap(service.Envs, env)

		// labels
		service.Labels["HAPROXY_GROUP"] = "external"
		service.Labels["HAPROXY_0_VHOST"] = p.getHaproxyVHost(clusterConfig["DICE_ROOT_DOMAIN"])

		// volumes
		hostPath := tmcInstance.ID
		if p.IsNotDCOSCluster(clusterConfig["DICE_CLUSTER_TYPE"]) {
			service.Binds = diceyml.Binds{
				hostPath + ":/opt/backup:rw",
			}
		}
	}

	return req
}

func (p *provider) BuildTmcInstanceConfig(tmcInstance *db.Instance, serviceGroupDeployResult interface{},
	clusterConfig map[string]string, additionalConfig map[string]string) map[string]string {

	serviceGroup := serviceGroupDeployResult.(*apistructs.ServiceGroup)
	mainClusterName := p.Cfg.MainClusterInfo.Name
	mainClusterDomain := p.Cfg.MainClusterInfo.RootDomain
	mainClusterScheme := p.Cfg.MainClusterInfo.Protocol
	mainClusterHTTPPort := p.Cfg.MainClusterInfo.HttpPort
	mainClusterHTTPSPort := p.Cfg.MainClusterInfo.HttpsPort

	var vip string
	for _, service := range serviceGroup.Services {
		if service.Name == "api-gateway-0" {
			vip = service.Vip
			break
		}
	}

	config := map[string]string{}

	if tmcInstance.Az == mainClusterName {
		config["HEPA_GATEWAY_HOST"] = "http://hepa.default.svc.cluster.local" // todo may be not in the default namespace?
		config["HEPA_GATEWAY_PORT"] = "8080"
	} else {
		schema := mainClusterScheme
		var port string
		if strings.Contains(schema, "https") {
			schema = "https"
			port = mainClusterHTTPSPort
		} else if strings.Contains(schema, "http") {
			schema = "http"
			port = mainClusterHTTPPort
		}
		config["HEPA_GATEWAY_HOST"] = schema + "://hepa." + mainClusterDomain
		config["HEPA_GATEWAY_PORT"] = port
	}

	config["VIP_KONG_HOST"] = "http://" + vip
	config["PROXY_KONG_PORT"] = "8000"
	config["GATEWAY_ENDPOINT"] = "gateway." + clusterConfig["DICE_ROOT_DOMAIN"]
	config["GATEWAY_INSTANCE_ID"] = tmcInstance.ID
	config["ADMIN_ENDPOINT"] = clusterConfig["NETPORTAL_URL"] + "/" + vip + ":8001"
	config["KONG_HOST"] = vip

	return config
}

func (p *provider) DoApplyTmcInstanceTenant(req *handlers.ResourceDeployRequest, resourceInfo *handlers.ResourceInfo,
	tmcInstance *db.Instance, tenant *db.InstanceTenant, clusterConfig map[string]string) (map[string]string, error) {
	resultConfig := map[string]string{}

	tenantOptions := map[string]string{}
	utils.JsonConvertObjToType(tenant.Options, &tenantOptions)
	instanceConfig := map[string]string{}
	utils.JsonConvertObjToType(tmcInstance.Config, &instanceConfig)

	gatewayReq := apistructs.GatewayTenantRequest{
		ID:              tenant.ID,
		TenantGroup:     tenant.TenantGroup,
		Az:              tmcInstance.Az,
		ProjectId:       tenantOptions["projectId"],
		ProjectName:     tenantOptions["projectName"],
		Env:             tenantOptions["env"],
		ServiceName:     "api-gateway",
		InstanceId:      tmcInstance.ID,
		InnerAddr:       "http://" + instanceConfig["KONG_HOST"] + ":8000",
		AdminAddr:       instanceConfig["ADMIN_ENDPOINT"],
		GatewayEndpoint: instanceConfig["GATEWAY_ENDPOINT"],
	}
	err := p.Bdl.CreateGatewayTenant(&gatewayReq)
	success := err == nil // for debug purpose
	if !success {
		return nil, err
	}

	key, _ := p.TmcIniDb.GetMicroServiceEngineJumpKey(tmcInstance.Engine)

	console := map[string]string{
		"tenantGroup": tenant.TenantGroup,
		"tenantId":    tenant.ID,
		"key":         key,
	}

	str, _ := utils.JsonConvertObjToString(console)
	resultConfig["PUBLIC_HOST"] = str

	return resultConfig, nil
}

func (p *provider) getHaproxyVHost(domain string) string {
	domainPrefixes := []string{
		"dev-gateway.",
		"test-gateway.",
		"staging-gateway.",
		"gateway.",
	}

	for i, prefix := range domainPrefixes {
		domainPrefixes[i] = prefix + domain
	}

	return strings.Join(domainPrefixes, ",")
}
