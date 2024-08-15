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

package addon

import (
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func TestSetlabelsFromOptions(t *testing.T) {
	type args struct {
		labels map[string]string
		opts   map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "test org label",
			args: args{
				labels: map[string]string{},
				opts: map[string]string{
					apistructs.LabelOrgName: "erda",
				},
			},
			want: map[string]string{
				apistructs.EnvDiceOrgName: "erda",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetlabelsFromOptions(tt.args.opts, tt.args.labels)
			if !reflect.DeepEqual(tt.args.labels, tt.want) {
				t.Errorf("SetlabelsFromOptions() = %v, want %v", tt.args.labels, tt.want)
			}
		})
	}
}

func TestBuildRocketMQOperaotrServiceItem(t *testing.T) {
	params := &apistructs.AddonHandlerCreateItem{
		Options: map[string]string{},
		Plan:    "basic",
	}
	addonIns := &dbclient.AddonInstance{
		ID: "1",
	}
	addonSpec := &apistructs.AddonExtension{
		Name: "rocketmq",
		Plan: map[string]apistructs.AddonPlanItem{
			"basic": {
				InsideMoudle: map[string]apistructs.AddonPlanItem{
					"rocketmq-namesrv": {
						CPU:   1,
						Mem:   2048,
						Nodes: 1,
					},
					"rocketmq-broker": {
						CPU:   1,
						Mem:   2048,
						Nodes: 2,
					},
					"rocketmq-console": {
						CPU:   0.5,
						Mem:   1024,
						Nodes: 1,
					},
				},
			},
		},
	}
	addonDice := &diceyml.Object{
		Services: diceyml.Services{
			"rocketmq-namesrv": {
				Labels: map[string]string{},
				Envs: map[string]string{
					"ADDON_TYPE": "rocketmq",
				},
			},
			"rocketmq-broker": {
				Labels: map[string]string{},
				Envs: map[string]string{
					"ADDON_TYPE": "rocketmq",
				},
			},
			"rocketmq-console": {
				Labels: map[string]string{},
				Envs: map[string]string{
					"ADDON_TYPE": "rocketmq",
				},
			},
		},
	}
	a := &Addon{}

	err := a.BuildRocketMQOperaotrServiceItem(params, addonIns, addonSpec, addonDice, nil, "5.0.0")
	assert.NoError(t, err)
	assert.Equal(t, 2, addonDice.Services["rocketmq-broker"].Deployments.Replicas)
}

func TestBuildRedisServiceItem(t *testing.T) {
	params := &apistructs.AddonHandlerCreateItem{
		Options: map[string]string{},
		Plan:    "basic",
	}
	addonIns := &dbclient.AddonInstance{
		ID: "1",
	}
	addonSpec := &apistructs.AddonExtension{
		Name: "redis",
		Plan: map[string]apistructs.AddonPlanItem{
			"basic": {
				CPU:   0.5,
				Mem:   256,
				Nodes: 1,
			},
		},
	}
	addonDice := &diceyml.Object{
		Services: diceyml.Services{
			"redis-master": {
				Labels: map[string]string{},
				Envs: map[string]string{
					"ADDON_TYPE":     "redis",
					"ADDON_GROUP_ID": "redis",
				},
			},
			"redis-slave": {
				Labels: map[string]string{},
				Envs: map[string]string{
					"ADDON_TYPE":     "redis",
					"ADDON_GROUP_ID": "redis",
				},
			},
			"redis-sentinel": {
				Labels: map[string]string{},
				Envs: map[string]string{
					"ADDON_TYPE":     "redis",
					"ADDON_GROUP_ID": "redis",
				},
			},
		},
	}
	bdl := bundle.New()

	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "KMSEncrypt", func(_ *bundle.Bundle, req apistructs.KMSEncryptRequest) (*kmstypes.EncryptResponse, error) {
		return &kmstypes.EncryptResponse{KeyID: "1", CiphertextBase64: "xxx"}, nil
	})
	defer pm1.Unpatch()

	db := &dbclient.DBClient{
		DBEngine: &dbengine.DBEngine{
			DB: &gorm.DB{},
		},
	}
	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "CreateAddonInstanceExtra", func(_ *dbclient.DBClient, addonInstanceExtra *dbclient.AddonInstanceExtra) error {
		return nil
	})
	defer pm2.Unpatch()

	pm3 := monkey.PatchInstanceMethod(reflect.TypeOf(&gorm.DB{}), "Create", func(_ *gorm.DB, value interface{}) *gorm.DB {
		return &gorm.DB{}
	})
	defer pm3.Unpatch()

	a := &Addon{db: db, bdl: bdl}
	err := a.BuildRedisServiceItem(params, addonIns, addonSpec, addonDice)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(addonDice.Services[apistructs.RedisMasterNamePrefix].SideCars))
}

func TestBuildEsServiceItem(t *testing.T) {
	params := &apistructs.AddonHandlerCreateItem{
		Options: map[string]string{},
		Plan:    "basic",
	}
	addonIns := &dbclient.AddonInstance{
		ID: "1",
	}
	addonSpec := &apistructs.AddonExtension{
		Name: "elasticsearch",
		Plan: map[string]apistructs.AddonPlanItem{
			"basic": {
				CPU:   0.5,
				Mem:   256,
				Nodes: 1,
			},
		},
	}
	addonDice := &diceyml.Object{
		Services: diceyml.Services{
			"elasticsearch": {
				Labels: map[string]string{},
				Envs: map[string]string{
					"ADDON_TYPE": "elasticsearch",
				},
			},
		},
	}

	a := &Addon{}
	err := a.BuildEsServiceItem(params, addonIns, addonSpec, addonDice, &apistructs.ClusterInfoData{})
	assert.NoError(t, err)
	assert.Equal(t, &[]int64{1000}[0], addonDice.Services["elasticsearch-1"].K8SSnippet.Container.SecurityContext.RunAsUser)
}

func Test_getKafkaExporterImage(t *testing.T) {
	testCases := []struct {
		name        string
		serviceItem diceyml.Service
		expect      string
	}{
		{
			name: "specify image in env",
			serviceItem: diceyml.Service{
				Envs: diceyml.EnvMap{
					EnvKafkaExporter: "kafka-exporter:1.0.0",
				},
			},
			expect: "kafka-exporter:1.0.0",
		},
		{
			name: "not specify image in env",
			serviceItem: diceyml.Service{
				Envs: diceyml.EnvMap{},
			},
			expect: DefaultKafkaExporterImage,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, getKafkaExporterImage(tc.serviceItem))
		})
	}
}

func TestBuildCanalServiceItem(t *testing.T) {
	a := &Addon{}
	ins := &dbclient.AddonInstance{
		ID:                  "a40d0fd095bd1428ba3e2193d38b01a09",
		Name:                "canal",
		AddonID:             "",
		AddonName:           "canal",
		Category:            "database",
		Namespace:           "",
		ScheduleName:        "",
		Plan:                "basic",
		Version:             "1.1.5",
		Options:             `{"applicationId":"7234","applicationName":"ttt-apm-demo","clusterName":"erda-jicheng","deploymentId":"26348","env":"TEST","logSource":"deploy","orgId":"633","orgName":"erda-development","projectId":"1904","projectName":"erda-development","runtimeId":"16070","runtimeName":"develop","tenantGroup":"5030d0d1a505db773fecf5049f679628","version":"1.1.5","workspace":"TEST"}`,
		Config:              `{"CANAL_HOST":"canal-1904-test-x.project-1904-test.svc.cluster.local","CANAL_PORT":"11111"}`,
		Label:               "DETACHED",
		Status:              "",
		ShareScope:          "PROJECT",
		OrgID:               "633",
		Cluster:             "erda-jicheng",
		ProjectID:           "1904",
		ApplicationID:       "7234",
		AppID:               "0",
		Workspace:           "TEST",
		Deleted:             "N",
		PlatformServiceType: 0,
		KmsKey:              "",
		CreatedAt:           time.Date(2023, 4, 27, 14, 56, 14, 0, time.UTC),
		UpdatedAt:           time.Date(2023, 6, 25, 10, 40, 33, 0, time.UTC),
		CpuRequest:          0,
		CpuLimit:            0,
		MemRequest:          0,
		MemLimit:            0,
	}
	spec := &apistructs.AddonExtension{
		Type:        "addon",
		Name:        "canal",
		Desc:        "${{ i18n.desc }}",
		DisplayName: "Canal",
		Category:    "database",
		LogoUrl:     "//terminus-dice.oss-cn-hangzhou.aliyuncs.com/addon/ui/logo/canal.png",
		ImageURLs: []string{
			"//terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2019/06/28/c4973926-9ba5-4e7a-b51c-3439ac4736a7.png",
			"//terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2019/06/28/a849135f-9d6e-4641-9e0a-d9ab54f7d03b.png",
			"//terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2019/06/28/09094324-f16b-4159-b66e-5a132e5e6da8.png",
		},
		Strategy: map[string]interface{}{
			"supportClusterType": 5,
		},
		Version:     "1.1.6",
		SubCategory: "middleware",
		Domain:      "http://middleware.terminus.io",
		Requires: []string{
			"many_per_app",
			"attachable",
			"plan_change",
		},
		ConfigVars: []string{
			"CANAL_HOST",
			"CANAL_PORT",
		},
		Envs: []string{
			"canal.admin.manager",
			"canal.instance.master.address",
			"canal.instance.dbUsername",
			"canal.instance.dbPassword",
			"admin.spring.datasource.address",
			"admin.spring.datasource.username",
			"admin.spring.datasource.password",
			"admin.spring.datasource.database",
		},
		Plan: map[string]apistructs.AddonPlanItem{
			"basic": {
				CPU:   1,
				Mem:   2048,
				Nodes: 1,
				Offerings: []string{
					"${{ i18n.plan.basic.offerings-0 }}",
				},
			},
			"professional": {
				CPU:   2,
				Mem:   4096,
				Nodes: 1,
				Offerings: []string{
					"${{ i18n.plan.professional.offerings-0 }}",
				},
			},
			"ultimate": {
				CPU:   2,
				Mem:   4096,
				Nodes: 2,
				Offerings: []string{
					"${{ i18n.plan.ultimate.offerings-0 }}",
				},
			},
		},
		ShareScopes: []string{
			"PROJECT",
			"ORG",
		},
		Similar:    []string{},
		Deprecated: false,
	}
	object := &diceyml.Object{
		Version: "2.0",
		Meta:    map[string]string{},
		Envs:    diceyml.EnvMap{},
		Services: diceyml.Services{
			"canal": &diceyml.Service{
				Image: "registry.erda.cloud/erda-addons/canal:1.1.6",
				Resources: diceyml.Resources{
					CPU: 1,
					Mem: 2048,
				},
				Deployments: diceyml.Deployments{
					Replicas: 1,
				},
				Ports: []diceyml.ServicePort{{Port: 8089}, {Port: 11110}},
				Envs: map[string]string{
					"ADDON_GROUP_ID": "canal",
					"ADDON_TYPE":     "canal",
				},
			},
		},
		Jobs:         diceyml.Jobs{},
		Environments: diceyml.EnvObjects{},
		Values:       diceyml.ValueObjects{},
	}

	config := map[string]string{
		"CANAL_HOST": "canal-1904-test-x.project-1904-test.svc.cluster.local",
		"CANAL_PORT": "11111",
	}

	options := map[string]string{
		"applicationId":                 "7234",
		"applicationName":               "ttt-apm-demo",
		"clusterName":                   "erda-jicheng",
		"deploymentId":                  "26348",
		"env":                           "TEST",
		"logSource":                     "deploy",
		"orgId":                         "633",
		"orgName":                       "erda-development",
		"projectId":                     "1904",
		"projectName":                   "erda-development",
		"runtimeId":                     "16070",
		"runtimeName":                   "develop",
		"tenantGroup":                   "5030d0d1a505db773fecf5049f679628",
		"version":                       "1.1.5",
		"workspace":                     "TEST",
		"canal.instance.master.address": "canal",
		"canal.instance.dbUsername":     "mysql",
		"canal.instance.dbPassword":     "pwd",
	}

	test := []*apistructs.AddonHandlerCreateItem{
		{InstanceName: "canal",
			AddonName:     "canal",
			Plan:          "basic",
			Tag:           "DETACHED",
			ClusterName:   "erda-jicheng",
			Workspace:     "TEST",
			OrgID:         "633",
			ProjectID:     "1904",
			ApplicationID: "7234",
			OperatorID:    "", // 无法从addonInstance中获取
			Config:        config,
			Options:       options,
			RuntimeID:     "16070",
			RuntimeName:   "develop",
			InsideAddon:   "", // 无法从addonInstance中获取
			ShareScope:    "PROJECT"},

		{InstanceName: "canal",
			AddonName:     "canal",
			Plan:          "basic",
			Tag:           "DETACHED",
			ClusterName:   "erda-jicheng",
			Workspace:     "TEST",
			OrgID:         "633",
			ProjectID:     "1904",
			ApplicationID: "7234",
			OperatorID:    "", // 无法从addonInstance中获取
			Config:        config,
			Options:       options,
			RuntimeID:     "16070",
			RuntimeName:   "develop",
			InsideAddon:   "", // 无法从addonInstance中获取
			ShareScope:    "APPLICATION"},
	}

	for _, tt := range test {
		err := a.BuildCanalServiceItem(true, tt, ins, spec, object)
		if err != nil {
			t.Fatal(err)
		}
	}
}
