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

package monitorcollector

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/msp/instance/db"
	"github.com/erda-project/erda/modules/msp/resource/deploy/handlers"
	"github.com/erda-project/erda/modules/msp/resource/utils"
)

func (p *provider) IsMatch(tmc *db.Tmc) bool {
	return tmc.Engine == handlers.ResourceMonitorCollector
}

func (p *provider) BuildServiceGroupRequest(resourceInfo *handlers.ResourceInfo, tmcInstance *db.Instance, clusterConfig map[string]string) interface{} {
	req := p.DefaultDeployHandler.BuildServiceGroupRequest(resourceInfo, tmcInstance, clusterConfig).(*apistructs.ServiceGroupCreateV2Request)

	instanceOptions := map[string]string{}
	utils.JsonConvertObjToType(tmcInstance.Options, &instanceOptions)

	for _, service := range resourceInfo.Dice.Services {
		nodeId := utils.GetRandomId()

		// envs
		env := map[string]string{
			"ADDON_ID":          tmcInstance.ID,
			"ADDON_NODE_ID":     nodeId,
			"BOOTSTRAP_SERVERS": instanceOptions["KAFKA_HOSTS"],
			"ZOOKEEPER_ADDR":    instanceOptions["ZK_HOSTS"],
		}
		utils.AppendMap(service.Envs, env)
	}

	return req
}

func (p *provider) BuildTmcInstanceConfig(tmcInstance *db.Instance, serviceGroupDeployResult interface{},
	clusterConfig map[string]string, additionalConfig map[string]string) map[string]string {
	instanceOptions := map[string]string{}
	utils.JsonConvertObjToType(tmcInstance.Options, &instanceOptions)

	var vip string
	serviceGroup := serviceGroupDeployResult.(*apistructs.ServiceGroup)
	for _, service := range serviceGroup.Services {
		if service.Name == "monitor-collector" {
			vip = service.Vip
			break
		}
	}

	config := map[string]string{
		"BOOTSTRAP_SERVERS":     instanceOptions["KAFKA_HOSTS"],
		"ZOOKEEPER_ADDR":        "ZK_HOSTS",
		"MONITOR_LOG_COLLECTOR": "http://" + vip + ":7076/collect/logs/container",
	}

	return config
}
