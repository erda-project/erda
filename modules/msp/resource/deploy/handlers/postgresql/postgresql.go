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

package postgresql

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/msp/instance/db"
	"github.com/erda-project/erda/modules/msp/resource/deploy/handlers"
	"github.com/erda-project/erda/modules/msp/resource/utils"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func (p *provider) IsMatch(tmc *db.Tmc) bool {
	return tmc.Engine == handlers.ResourcePostgresql
}

func (p *provider) CheckIfHasCustomConfig(clusterConfig map[string]string) (map[string]string, bool) {
	// try find if aliyun mse instance exists, reuse it if any
	pgsqlHost, ok := clusterConfig["POSTGRESQL_HOST"]
	if !ok {
		return nil, false
	}

	pgsqlPort, ok := clusterConfig["POSTGRESQL_PORT"]
	if !ok {
		return nil, false
	}

	pgsqlUser, ok := clusterConfig["POSTGRESQL_USER"]
	if !ok {
		return nil, false
	}

	pgsqlPassword, ok := clusterConfig["POSTGRESQL_PASSWORD"]
	if !ok {
		return nil, false
	}
	decryptPassword := utils.AesDecrypt(pgsqlPassword)

	config := map[string]string{
		"POSTGRESQL_HOST":     pgsqlHost,
		"POSTGRESQL_PORT":     pgsqlPort,
		"POSTGRESQL_USER":     pgsqlUser,
		"POSTGRESQL_PASSWORD": decryptPassword,
		"POSTGRESQL_DATABASE": "kong",
		"POSTGRESQL_ADDRESS":  pgsqlHost + ":" + pgsqlPort,
	}

	return config, true
}

func (p *provider) BuildServiceGroupRequest(resourceInfo *handlers.ResourceInfo, tmcInstance *db.Instance, clusterConfig map[string]string) interface{} {
	req := p.DefaultDeployHandler.BuildServiceGroupRequest(resourceInfo, tmcInstance, clusterConfig).(*apistructs.ServiceGroupCreateV2Request)

	delete(req.GroupLabels, "LOCATION-CLUSTER-SERVICE")

	for name, service := range resourceInfo.Dice.Services {
		delete(service.Envs, "LOCATION-CLUSTER-SERVICE")

		// envs
		nodeId := tmcInstance.ID + "_" + name
		env := map[string]string{
			"ADDON_ID":      tmcInstance.ID,
			"ADDON_NODE_ID": nodeId,
			"DICE_CLUSTER":  tmcInstance.Az,
		}
		utils.AppendMap(service.Envs, env)

		// volumes
		hostPath := tmcInstance.ID
		if p.IsNotDCOSCluster(clusterConfig["DICE_CLUSTER_TYPE"]) {
			service.Binds = diceyml.Binds{
				hostPath + ":/var/lib/postgresql/data:rw",
				clusterConfig["DICE_STORAGE_MOUNTPOINT"] + "/addon/postgresql/backup/" + hostPath + ":/var/backup/pg:rw",
			}
		} else {
			service.Binds = diceyml.Binds{
				clusterConfig["DICE_STORAGE_MOUNTPOINT"] + "/addon/postgresql/data/" + hostPath + ":/var/lib/postgresql/data:rw",
				clusterConfig["DICE_STORAGE_MOUNTPOINT"] + "/addon/postgresql/backup/" + nodeId + ":/var/backup/pg:rw",
			}
		}

		// health check
		service.HealthCheck = diceyml.HealthCheck{
			Exec: &diceyml.ExecCheck{Cmd: "psql -d " + service.Envs["POSTGRES_DB"] + " -U " + service.Envs["POSTGRES_USER"] + " -h localhost -c 'select 1';"},
		}
	}

	return req
}

func (p *provider) BuildTmcInstanceConfig(tmcInstance *db.Instance, serviceGroupDeployResult interface{},
	clusterConfig map[string]string, additionalConfig map[string]string) map[string]string {
	var vip string
	serviceGroup := serviceGroupDeployResult.(*apistructs.ServiceGroup)
	for _, service := range serviceGroup.Services {
		if service.Name == "postgresql" {
			vip = service.Vip
			break
		}
	}

	config := map[string]string{
		"POSTGRESQL_HOST":     vip,
		"POSTGRESQL_PORT":     "5432",
		"POSTGRESQL_USER":     "kong",
		"POSTGRESQL_PASSWORD": "",
		"POSTGRESQL_DATABASE": "kong",
		"POSTGRESQL_ADDRESS":  vip + ":5432",
	}

	return config
}
