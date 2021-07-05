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

package servicemesh

import (
	"fmt"
	"github.com/erda-project/erda/modules/msp/instance/db"
	"github.com/erda-project/erda/modules/msp/resource/deploy/handlers"
	"github.com/erda-project/erda/modules/msp/resource/utils"
)

func (p *provider) IsMatch(tmc *db.Tmc) bool {
	return tmc.Engine == handlers.ResourceServiceMesh
}

func (p *provider) DoPostDeployJob(tmcInstance *db.Instance, serviceGroupDeployResult interface{}, clusterConfig map[string]string) (map[string]string, error) {
	if clusterConfig["ISTIO_INSTALLED"] != "true" {
		return nil, fmt.Errorf("istio not installed")
	}
	if clusterConfig["ISTIO_VERSION"] == "1.1.4" {
		return nil, fmt.Errorf("istio need upgrade")
	}

	return nil, nil
}

func (p *provider) BuildTmcInstanceConfig(tmcInstance *db.Instance, serviceGroupDeployResult interface{},
	clusterConfig map[string]string, additionalConfig map[string]string) map[string]string {

	config := map[string]string{
		"ADDON_SERVICE_MESH": "on",
	}

	return config
}

func (p *provider) DoApplyTmcInstanceTenant(req *handlers.ResourceDeployRequest, resourceInfo *handlers.ResourceInfo,
	tmcInstance *db.Instance, tenant *db.InstanceTenant, clusterConfig map[string]string) (map[string]string, error) {

	key, _ := p.TmcIniDb.GetMicroServiceEngineJumpKey(handlers.ResourceMonitor)

	tk := ""
	var monitorInstance, _ = p.InstanceDb.GetByEngineAndTenantGroup(handlers.ResourceMonitor, tenant.TenantGroup)
	if monitorInstance != nil {
		monitorConfig := map[string]string{}
		utils.JsonConvertObjToType(monitorInstance.Config, &monitorConfig)
		tk = monitorConfig["TERMINUS_KEY"]
	}

	console := map[string]string{
		"tenantGroup": tenant.TenantGroup,
		"tenantId":    tenant.ID,
		"key":         key,
		"terminusKey": tk,
	}

	str, _ := utils.JsonConvertObjToString(console)
	config := map[string]string{
		"PUBLIC_HOST": str,
	}

	return config, nil
}
