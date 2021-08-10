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

package orchestrator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func Test_GenScaleMessage(t *testing.T) {
	oldService := &apistructs.RuntimeInspectServiceDTO{
		Resources: apistructs.RuntimeServiceResourceDTO{
			CPU: 1,
			Mem: 1024,
		},
		Deployments: apistructs.RuntimeServiceDeploymentsDTO{
			Replicas: 1,
		},
	}

	zh, en := genScaleMessage(oldService, oldService)
	assert.Equal(t, zh, "CPU: 未变, 内存: 未变, 实例数: 未变")
	assert.Equal(t, en, "CPU: no change, Mem: no change, Replicas: no change")

	newService := &apistructs.RuntimeInspectServiceDTO{}
	newService.Resources.CPU = 2
	newService.Resources.Mem = 1025
	newService.Deployments.Replicas = 2
	zh, en = genScaleMessage(oldService, newService)
	assert.Equal(t, zh, "CPU: 从 1核 变为 2核, 内存: 从 1024MB 变为 1025MB, 实例数: 从 1 变为 2")
	assert.Equal(t, en, "CPU: update from 1core to 2core, Mem: update from 1024MB to 1025MB, Replicas: update from 1 to 2")

	newService.Deployments.Replicas = 1
	zh, en = genScaleMessage(oldService, newService)
	assert.Equal(t, zh, "CPU: 从 1核 变为 2核, 内存: 从 1024MB 变为 1025MB, 实例数: 未变")
	assert.Equal(t, en, "CPU: update from 1core to 2core, Mem: update from 1024MB to 1025MB, Replicas: no change")
}
