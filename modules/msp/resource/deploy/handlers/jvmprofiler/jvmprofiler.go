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

package jvmprofiler

import (
	"github.com/erda-project/erda/modules/msp/instance/db"
	"github.com/erda-project/erda/modules/msp/resource/deploy/handlers"
	"github.com/erda-project/erda/modules/msp/resource/utils"
)

func (p *provider) IsMatch(tmc *db.Tmc) bool {
	return tmc.Engine == handlers.ResourceJvmProfiler
}

func (p *provider) DoApplyTmcInstanceTenant(req *handlers.ResourceDeployRequest, resourceInfo *handlers.ResourceInfo,
	tmcInstance *db.Instance, tenant *db.InstanceTenant, clusterConfig map[string]string) (map[string]string, error) {

	key, _ := p.TmcIniDb.GetMicroServiceEngineJumpKey(tmcInstance.Engine)
	console := map[string]string{
		"tenantId":                tenant.ID,
		"tenantGroup":             tenant.TenantGroup,
		"monitor_jvm_profiler_id": tenant.ID,
		"key":                     key,
	}

	consoleStr, _ := utils.JsonConvertObjToString(console)
	config := map[string]string{
		"MONITOR_JVM_PROFILER_ID": tenant.ID,
		"PUBLIC_HOST":             consoleStr,
	}

	return config, nil
}
