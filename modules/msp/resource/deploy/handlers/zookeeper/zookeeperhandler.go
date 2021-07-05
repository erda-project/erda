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

package zookeeper

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/msp/instance/db"
	"github.com/erda-project/erda/modules/msp/resource/deploy/handlers"
	"github.com/erda-project/erda/modules/msp/resource/utils"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func (p *provider) IsMatch(tmc *db.Tmc) bool {
	return tmc.Engine == handlers.ResourceZookeeper
}

func (p *provider) BuildServiceGroupRequest(resourceInfo *handlers.ResourceInfo, tmcInstance *db.Instance, clusterConfig map[string]string) interface{} {
	req := p.DefaultDeployHandler.BuildServiceGroupRequest(resourceInfo, tmcInstance, clusterConfig).(*apistructs.ServiceGroupCreateV2Request)

	for _, service := range resourceInfo.Dice.Services {
		nodeId := tmcInstance.ID + "_" + service.Envs["ZOO_MY_ID"]
		env := map[string]string{
			"ADDON_ID":      tmcInstance.ID,
			"ADDON_NODE_ID": nodeId,
		}
		utils.AppendMap(service.Envs, env)

		if p.IsNotDCOSCluster(clusterConfig["DICE_CLUSTER_TYPE"]) {
			service.Binds = diceyml.Binds{
				nodeId + "_data:/data:rw",
				nodeId + "_datalog:/datalog:rw",
			}
		} else {
			service.Binds = diceyml.Binds{
				clusterConfig["DICE_STORAGE_MOUNTPOINT"] + "/addon/zookeeper/data/" + nodeId + ":/data:rw",
				clusterConfig["DICE_STORAGE_MOUNTPOINT"] + "/addon/zookeeper/datalog/" + nodeId + ":/datalog:rw",
			}
		}
	}

	return req
}

func (p *provider) BuildTmcInstanceConfig(tmcInstance *db.Instance, serviceGroupDeployResult interface{},
	clusterConfig map[string]string, additionalConfig map[string]string) map[string]string {
	serviceGroup := serviceGroupDeployResult.(*apistructs.ServiceGroup)
	var vip string
	for _, service := range serviceGroup.Services {
		if service.Name == "zookeeper-1" {
			vip = service.Vip
			break
		}
	}

	config := map[string]string{
		"ZOOKEEPER_HOST":    vip,
		"ZOOKEEPER_PORT":    "2181",
		"ZOOKEEPER_ADDRESS": vip + ":2181",
	}

	return config
}
