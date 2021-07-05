/*
 * // Copyright (c) 2021 Terminus, Inc.
 * //
 * // This program is free software: you can use, redistribute, and/or modify
 * // it under the terms of the GNU Affero General Public License, version 3
 * // or later ("AGPL"), as published by the Free Software Foundation.
 * //
 * // This program is distributed in the hope that it will be useful, but WITHOUT
 * // ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * // FITNESS FOR A PARTICULAR PURPOSE.
 * //
 * // You should have received a copy of the GNU Affero General Public License
 * // along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package zkproxy

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/msp/instance/db"
	"github.com/erda-project/erda/modules/msp/resource/deploy/handlers"
	"github.com/erda-project/erda/modules/msp/resource/utils"
	"strings"
)

func (h *provider) IsMatch(tmc *db.Tmc) bool {
	return tmc.Engine == handlers.ResourceZKProxy
}

func (h *provider) BuildServiceGroupRequest(resourceInfo *handlers.ResourceInfo, tmcInstance *db.Instance, clusterConfig map[string]string) interface{} {
	req := h.DefaultDeployHandler.BuildServiceGroupRequest(resourceInfo, tmcInstance, clusterConfig).(*apistructs.ServiceGroupCreateV2Request)

	instanceOptions := map[string]string{}
	utils.JsonConvertObjToType(tmcInstance.Options, &instanceOptions)

	for name, service := range resourceInfo.Dice.Services {
		env := map[string]string{
			"ADDON_ID":       tmcInstance.ID,
			"ADDON_NODE_ID":  tmcInstance.ID + "_" + name,
			"ETCD_ENDPOINTS": instanceOptions["ETCD_ADDRESS"],
		}
		utils.AppendMap(service.Envs, env)
	}

	return req
}

func (h *provider) BuildTmcInstanceConfig(tmcInstance *db.Instance, serviceGroupDeployResult interface{},
	clusterConfig map[string]string, additionalConfig map[string]string) map[string]string {
	var vip string
	serviceGroup := serviceGroupDeployResult.(*apistructs.ServiceGroup)
	for _, service := range serviceGroup.Services {
		if service.Name == "zkproxy-1" {
			vip = service.Vip
			break
		}
	}

	instanceOptions := map[string]string{}
	utils.JsonConvertObjToType(tmcInstance.Options, &instanceOptions)

	config := map[string]string{
		"ELASTICJOB_HOST":     instanceOptions["ZOOKEEPER_ADDRESS"],
		"ZKPROXY_PUBLIC_HOST": "http://" + vip + ":8080",
		"ZOOKEEPER_DUBBO":     vip + ":2181",
		"ZKPROXY_HOST":        vip,
		"ZKPROXY_PORT":        "2181",
		"ETCD_ADDRESS":        instanceOptions["ETCD_ADDRESS"],
	}

	return config
}

func (h *provider) DoApplyTmcInstanceTenant(req *handlers.ResourceDeployRequest, resourceInfo *handlers.ResourceInfo,
	tmcInstance *db.Instance, tenant *db.InstanceTenant, clusterConfig map[string]string) (map[string]string, error) {
	tenantOptions := map[string]string{}
	utils.JsonConvertObjToType(tenant.Options, &tenantOptions)

	config := map[string]string{
		"DUBBO_TENANT_ID": tenantOptions["projectId"] + "_" + strings.ToLower(tenantOptions["env"]),
	}

	return config, nil
}
