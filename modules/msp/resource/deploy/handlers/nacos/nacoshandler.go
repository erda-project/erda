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

package nacos

import (
	"fmt"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/msp/instance/db"
	"github.com/erda-project/erda/modules/msp/resource/deploy/handlers"
	"github.com/erda-project/erda/modules/msp/resource/utils"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"strings"
)

func (p *provider) IsMatch(tmc *db.Tmc) bool {
	return tmc.Engine == handlers.ResourceNacos
}

func (p *provider) CheckIfHasCustomConfig(clusterConfig map[string]string) (map[string]string, bool) {
	// try find if aliyun mse instance exists, reuse it if any
	nacosHost, ok := clusterConfig["MS_NACOS_HOST"]
	if !ok {
		return nil, false
	}

	nacosPort, ok := clusterConfig["MS_NACOS_PORT"]
	if !ok {
		return nil, false
	}

	config := map[string]string{
		"NACOS_HOST":     nacosHost,
		"NACOS_PORT":     nacosPort,
		"NACOS_ADDRESS":  nacosHost + ":" + nacosPort,
		"NACOS_USER":     "nacos",
		"NACOS_PASSWORD": "nacos",
	}

	return config, true
}

func (p *provider) BuildServiceGroupRequest(resourceInfo *handlers.ResourceInfo, tmcInstance *db.Instance, clusterConfig map[string]string) interface{} {
	req := p.DefaultDeployHandler.BuildServiceGroupRequest(resourceInfo, tmcInstance, clusterConfig).(*apistructs.ServiceGroupCreateV2Request)

	options := map[string]string{}
	utils.JsonConvertObjToType(tmcInstance.Options, &options)
	envs := map[string]string{
		"MYSQL_DATABASE_NUM":            "2",
		"MYSQL_MASTER_SERVICE_HOST":     options["MYSQL_HOST"],
		"MYSQL_MASTER_SERVICE_PORT":     options["MYSQL_PORT"],
		"MYSQL_MASTER_SERVICE_DB_NAME":  options["MYSQL_DATABASES"],
		"MYSQL_SLAVE_SERVICE_HOST":      options["MYSQL_SLAVE_HOST"],
		"MYSQL_SLAVE_SERVICE_PORT":      options["MYSQL_SLAVE_PORT"],
		"MYSQL_MASTER_SERVICE_USER":     options["MYSQL_USERNAME"],
		"MYSQL_MASTER_SERVICE_PASSWORD": options["MYSQL_PASSWORD"],
	}

	healthCheck := diceyml.HealthCheck{
		HTTP: &diceyml.HTTPCheck{
			Port: 8848,
			Path: "/nacos/v1/console/health/liveness",
		},
	}

	for _, service := range req.DiceYml.Services {
		utils.AppendMap(service.Envs, envs)
		service.HealthCheck = healthCheck
	}

	return req
}

func (p *provider) BuildTmcInstanceConfig(tmcInstance *db.Instance, serviceGroupDeployResult interface{},
	clusterConfig map[string]string, additionalConfig map[string]string) map[string]string {
	var vip string
	serviceGroup := serviceGroupDeployResult.(*apistructs.ServiceGroup)
	for _, service := range serviceGroup.Services {
		if service.Name == "nacos" {
			vip = service.Vip
			break
		}
	}

	config := map[string]string{
		"NACOS_USER":     "nacos",
		"NACOS_PASSWORD": "nacos",
		"NACOS_HOST":     vip,
		"NACOS_PORT":     "8848",
		"NACOS_ADDRESS":  vip + ":8848",
	}

	return config
}

func (p *provider) DoApplyTmcInstanceTenant(req *handlers.ResourceDeployRequest, resourceInfo *handlers.ResourceInfo,
	tmcInstance *db.Instance, tenant *db.InstanceTenant, clusterConfig map[string]string) (map[string]string, error) {

	instanceConfig := map[string]string{}
	_ = utils.JsonConvertObjToType(tmcInstance.Config, &instanceConfig)
	addr := instanceConfig["NACOS_ADDRESS"]
	user := instanceConfig["NACOS_USER"]
	pwd := instanceConfig["NACOS_PASSWORD"]
	namespace := tenant.TenantGroup

	if len(namespace) == 0 {
		return nil, fmt.Errorf("namespace is nil")
	}

	netportalUrl := clusterConfig["NETPORTAL_URL"]
	if len(netportalUrl) > 0 {
		addr = netportalUrl + "/" + strings.TrimPrefix(addr, "http://")
	} else if !strings.HasPrefix(addr, "http://") {
		addr = "http://" + addr
	}

	namespaceId, err := getOrCreateNacosNamespace(addr, user, pwd, namespace)
	if err != nil {
		return nil, err
	}

	config := map[string]string{
		"NACOS_TENANT_ID": namespaceId,
	}

	return config, nil
}

func getOrCreateNacosNamespace(addr string, user string, pwd string, namespace string) (string, error) {
	nacosClient := utils.NewNacosClient(addr, user, pwd)
	namespaceId, _ := nacosClient.GetNamespaceId(namespace)
	if len(namespaceId) > 0 {
		return namespaceId, nil
	}
	namespaceId, err := nacosClient.CreateNamespace(namespace)
	if len(namespaceId) > 0 {
		return namespaceId, nil
	}
	return "", err
}
