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
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/dao"
	"github.com/erda-project/erda/modules/core-services/model"
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

func TestCheckRollbackConfig(t *testing.T) {
	rollbackConfig := make(map[string]int, 0)
	assert.NoError(t, checkRollbackConfig(&rollbackConfig))
	rollbackConfig["DEV"] = 1
	rollbackConfig["TEST"] = 1
	rollbackConfig["STAGING"] = 1
	rollbackConfig["PROD"] = 1
	assert.NoError(t, checkRollbackConfig(&rollbackConfig))
}

func TestInitRollbackConfig(t *testing.T) {
	rollbackConfig := make(map[string]int, 0)
	err := initRollbackConfig(&rollbackConfig)
	assert.NoError(t, err)
	assert.Equal(t, 5, rollbackConfig["DEV"])
	assert.Equal(t, 5, rollbackConfig["TEST"])
	assert.Equal(t, 5, rollbackConfig["STAGING"])
	assert.Equal(t, 5, rollbackConfig["PROD"])
}

func TestGetModelProjectsMap(t *testing.T) {
	db := &dao.DBClient{}
	m := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetProjectsByIDs",
		func(db *dao.DBClient, projectIDs []uint64, params *apistructs.ProjectListRequest) (int, []model.Project, error) {
			return 3, []model.Project{
				{BaseModel: model.BaseModel{ID: 1}},
				{BaseModel: model.BaseModel{ID: 2}},
				{BaseModel: model.BaseModel{ID: 3}},
			}, nil
		})
	defer m.Unpatch()

	p := &Project{}
	projectIDs := []uint64{1, 2, 3}
	projectMap, err := p.GetModelProjectsMap(projectIDs)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(projectMap))
}
