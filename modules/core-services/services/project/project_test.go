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

package project

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/stretchr/testify/assert"
)

func TestClass_genProjectNamespace(t *testing.T) {
	namespaces := genProjectNamespace("1")
	expected := map[string]string{"DEV": "project-1-dev", "TEST": "project-1-test", "STAGING": "project-1-staging", "PROD": "project-1-prod"}
	for env, expectedNS := range expected {
		actuallyNS, ok := namespaces[env]
		assert.True(t, ok)
		assert.Equal(t, expectedNS, actuallyNS)
	}
}

func TestClaas_patchProject(t *testing.T) {
	oldPrj := &model.Project{
		DisplayName:    "displayName",
		Desc:           "desc",
		Logo:           "log",
		DDHook:         "ddhook",
		ClusterConfig:  `{"DEV":"dev","PROD":"prod","STAGING":"staging","TEST":"test"}`,
		RollbackConfig: `{"DEV":5,"PROD":5,"STAGING":5,"TEST":5}`,
		CpuQuota:       1,
		MemQuota:       1,
		ActiveTime:     time.Now(),
		IsPublic:       false,
	}

	b := `{"name":"test997","logo":"newLogo","desc":"newDesc","ddHook":"","clusterConfig":{"DEV":"newDev","TEST":"newTest","STAGING":"newStaging","PROD":"newProd"},"memQuota":2,"cpuQuota":2}`
	var body apistructs.ProjectUpdateBody
	err := json.Unmarshal([]byte(b), &body)

	patchProject(oldPrj, &body)

	assert.NoError(t, err)
	assert.Equal(t, oldPrj.DisplayName, "displayName")
	assert.Equal(t, oldPrj.Desc, "newDesc")
	assert.Equal(t, oldPrj.Logo, "newLogo")
	assert.Equal(t, oldPrj.DDHook, "")
	assert.Equal(t, oldPrj.CpuQuota, float64(2))
	assert.Equal(t, oldPrj.MemQuota, float64(2))
	assert.NotEqual(t, oldPrj.ClusterConfig, "{}")
	assert.NotEqual(t, oldPrj.RollbackConfig, "{}")
}
