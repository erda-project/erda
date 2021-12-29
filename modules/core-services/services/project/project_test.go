// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package project

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/core-services/dao"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/pkg/ucauth"
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

	patchProject(oldPrj, &body, 0)

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
	projectMap, err := p.GetModelProjectsMap(projectIDs, false)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(projectMap))
}

func TestWtihI18n(t *testing.T) {
	var translator i18n.Translator
	New(WithI18n(translator))
}

func TestWithDBClient(t *testing.T) {
	New(WithDBClient(new(dao.DBClient)))
}

func TestWithUCClient(t *testing.T) {
	New(WithUCClient(new(ucauth.UCClient)))
}

func TestWithBundle(t *testing.T) {
	var bdl bundle.Bundle
	New(WithBundle(&bdl))
}

func Test_convertAuditCreateReq2Model(t *testing.T) {
	var audit = apistructs.Audit{
		ID:           0,
		UserID:       "",
		ScopeType:    "",
		ScopeID:      0,
		FDPProjectID: "",
		OrgID:        0,
		ProjectID:    0,
		AppID:        0,
		Context:      nil,
		TemplateName: "",
		AuditLevel:   "",
		Result:       "",
		ErrorMsg:     "",
		StartTime:    "2006-01-02 15:04:05",
		EndTime:      "2006-01-02 15:04:05",
		ClientIP:     "",
		UserAgent:    "",
	}
	if _, err := convertAuditCreateReq2Model(audit); err != nil {
		t.Fatal(err)
	}

	audit.EndTime = "123456"
	if _, err := convertAuditCreateReq2Model(audit); err == nil {
		t.Fatal("err")
	}
	audit.StartTime = "123456"
	if _, err := convertAuditCreateReq2Model(audit); err == nil {
		t.Fatal("err")
	}
}

func Test_calcuRequestRate(t *testing.T) {
	var (
		prod = apistructs.ResourceConfigInfo{
			CPUQuota:            100,
			CPURequest:          50,
			CPURequestByAddon:   30,
			CPURequestByService: 10,
			MemQuota:            100,
			MemRequest:          500,
			MemRequestByAddon:   30,
			MemRequestByService: 10,
		}
		staging = prod
		test    = prod
		dev     = prod
	)
	var dto = &apistructs.ProjectDTO{
		ResourceConfig: &apistructs.ResourceConfigsInfo{
			PROD:    &prod,
			STAGING: &staging,
			TEST:    &test,
			DEV:     &dev,
		},
	}
	p := new(Project)
	p.calcuRequestRate(dto)
}

func Test_setQuotaFromResourceConfig(t *testing.T) {
	setQuotaFromResourceConfig(nil, nil)

	var (
		prodClusterName           = "the-prod"
		prodCPU            uint64 = 1 * 1000
		prodMem            uint64 = 5 * 1024 * 1024 * 1024
		stagingClusterName        = "the-staging"
		stagingCPU         uint64 = 2 * 1000
		stagingMem         uint64 = 6 * 1024 * 1024 * 1024
		testClusterName           = "the-test"
		testCPU            uint64 = 3 * 1000
		testMem            uint64 = 7 * 1024 * 1024 * 1024
		devClusterName            = "the-dev"
		devCPU             uint64 = 4 * 1000
		devMem             uint64 = 8 * 1024 * 1024 * 1024
	)
	var quota = new(apistructs.ProjectQuota)
	var resource = &apistructs.ResourceConfigs{
		PROD: &apistructs.ResourceConfig{
			ClusterName: prodClusterName,
			CPUQuota:    1,
			MemQuota:    5,
		},
		STAGING: &apistructs.ResourceConfig{
			ClusterName: stagingClusterName,
			CPUQuota:    2,
			MemQuota:    6,
		},
		TEST: &apistructs.ResourceConfig{
			ClusterName: testClusterName,
			CPUQuota:    3,
			MemQuota:    7,
		},
		DEV: &apistructs.ResourceConfig{
			ClusterName: devClusterName,
			CPUQuota:    4,
			MemQuota:    8,
		},
	}

	setQuotaFromResourceConfig(quota, resource)

	if !(quota.ProdClusterName == prodClusterName && quota.ProdCPUQuota == prodCPU && quota.ProdMemQuota == prodMem) {
		t.Fatal("sets error")
	}
	if !(quota.StagingClusterName == stagingClusterName && quota.StagingCPUQuota == stagingCPU && quota.StagingMemQuota == stagingMem) {
		t.Fatal("sets error")
	}
	if !(quota.TestClusterName == testClusterName && quota.TestCPUQuota == testCPU && quota.TestMemQuota == testMem) {
		t.Fatal("sets error")
	}
	if !(quota.DevClusterName == devClusterName && quota.DevCPUQuota == devCPU && quota.DevMemQuota == devMem) {
		t.Fatal("sets error")
	}
}

func Test_setProjectDtoQuotaFromModel(t *testing.T) {
	setProjectDtoQuotaFromModel(nil, nil)

	var quota = apistructs.ProjectQuota{
		ProdClusterName:    "the-prod",
		StagingClusterName: "the-staging",
		TestClusterName:    "the-test",
		DevClusterName:     "the-dev",
		ProdCPUQuota:       1 * 1000,
		ProdMemQuota:       1 * 1024 * 1024 * 1024,
		StagingCPUQuota:    2 * 1000,
		StagingMemQuota:    2 * 1024 * 1024 * 1024,
		TestCPUQuota:       3 * 1000,
		TestMemQuota:       3 * 1024 * 1024 * 1024,
		DevCPUQuota:        4 * 1000,
		DevMemQuota:        4 * 1024 * 1024 * 1024,
	}
	var dto = new(apistructs.ProjectDTO)
	setProjectDtoQuotaFromModel(dto, &quota)
	rc := dto.ResourceConfig
	if !(rc.PROD.ClusterName == quota.ProdClusterName && rc.PROD.CPUQuota == float64(quota.ProdCPUQuota)/1000 && rc.PROD.MemQuota == float64(quota.ProdMemQuota)/1024/1024/1024) {
		t.Fatal("setProjectDtoQuotaFromModel error")
	}
	if !(rc.STAGING.ClusterName == quota.StagingClusterName && rc.STAGING.CPUQuota == float64(quota.StagingCPUQuota)/1000 && rc.STAGING.MemQuota == float64(quota.StagingMemQuota)/1024/1024/1024) {
		t.Fatal("setProjectDtoQuotaFromModel error")
	}
	if !(rc.TEST.ClusterName == quota.TestClusterName && rc.TEST.CPUQuota == float64(quota.TestCPUQuota)/1000 && rc.TEST.MemQuota == float64(quota.TestMemQuota)/1024/1024/1024) {
		t.Fatal("setProjectDtoQuotaFromModel error")
	}
	if !(rc.DEV.ClusterName == quota.DevClusterName && rc.DEV.CPUQuota == float64(quota.DevCPUQuota)/1000 && rc.DEV.MemQuota == float64(quota.DevMemQuota)/1024/1024/1024) {
		t.Fatal("setProjectDtoQuotaFromModel error")
	}

}

func initOldNewQutoa() (old, new_ apistructs.ProjectQuota) {
	return apistructs.ProjectQuota{
			ProdClusterName:    "prod",
			StagingClusterName: "staging",
			TestClusterName:    "test",
			DevClusterName:     "dev",
			ProdCPUQuota:       10,
			ProdMemQuota:       20,
			StagingCPUQuota:    30,
			StagingMemQuota:    40,
			TestCPUQuota:       50,
			TestMemQuota:       60,
			DevCPUQuota:        70,
			DevMemQuota:        80,
		}, apistructs.ProjectQuota{
			ProdClusterName:    "prod",
			StagingClusterName: "staging",
			TestClusterName:    "test",
			DevClusterName:     "dev",
			ProdCPUQuota:       10,
			ProdMemQuota:       20,
			StagingCPUQuota:    30,
			StagingMemQuota:    40,
			TestCPUQuota:       50,
			TestMemQuota:       60,
			DevCPUQuota:        70,
			DevMemQuota:        80,
		}
}

func Test_isQuotaChanged(t *testing.T) {
	if isChanged := isQuotaChanged(initOldNewQutoa()); isChanged {
		t.Fatal("error")
	}
}

func Test_isQuotaChangedOnTheWorkspace(t *testing.T) {
	var changedRecord map[string]bool
	old, new_ := initOldNewQutoa()
	isQuotaChangedOnTheWorkspace(changedRecord, old, new_)
	if len(changedRecord) > 0 {
		t.Fatal("error")
	}

	changedRecord = make(map[string]bool)
	isQuotaChangedOnTheWorkspace(changedRecord, old, new_)
	if len(changedRecord) != 4 {
		t.Fatal("error")
	}
	for _, w := range []string{"PROD", "STAGING", "TEST", "DEV"} {
		if changedRecord[w] {
			t.Fatal("error")
		}
	}

	new_.ProdCPUQuota += 10
	isQuotaChangedOnTheWorkspace(changedRecord, old, new_)
	if !changedRecord["PROD"] {
		t.Fatal("error")
	}
}

func Test_getFirstValidOwnerOrLead(t *testing.T) {
	var members = []model.Member{
		{
			UserID: "1",
			Roles:  []string{"Developer"},
		}, {
			UserID: "2",
			Roles:  []string{"Lead"},
		}, {
			UserID: "3",
			Roles:  []string{"Lead", "Owner"},
		}, {
			UserID: "4",
			Roles:  []string{"Owner"},
		},
	}
	var member *model.Member
	hitFirstValidOwnerOrLead(member, members)

	member = new(model.Member)
	hitFirstValidOwnerOrLead(member, members)
	if member.UserID != "3" {
		t.Fatal("hit error")
	}
}

func Test_defaultResourceConfig(t *testing.T) {
	var dto = new(apistructs.ProjectDTO)
	defaultResourceConfig(dto)

	dto.ClusterConfig = make(map[string]string)
	defaultResourceConfig(dto)

	dto.ClusterConfig["PROD"] = "prod-cluster"
	dto.ClusterConfig["STAGING"] = "staging-cluster"
	defaultResourceConfig(dto)

	dto.ClusterConfig["TEST"] = "test-cluster"
	dto.ClusterConfig["DEV"] = "dev-cluster"
	defaultResourceConfig(dto)
	if dto.ResourceConfig == nil {
		t.Fatal("dto.ResourceConfig should not be nil")
	}
	if dto.ResourceConfig.PROD.ClusterName != "prod-cluster" ||
		dto.ResourceConfig.STAGING.ClusterName != "staging-cluster" ||
		dto.ResourceConfig.TEST.ClusterName != "test-cluster" ||
		dto.ResourceConfig.DEV.ClusterName != "dev-cluster" {
		t.Fatal("clusters names error in ResourceConfig")
	}
}
