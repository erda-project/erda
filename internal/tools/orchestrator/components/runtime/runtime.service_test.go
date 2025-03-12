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

package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	clusterpb "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	releasepb "github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda-proto-go/orchestrator/runtime/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/dicehub/release"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/pkg/diceworkspace"
	"github.com/erda-project/erda/internal/pkg/gitflowutil"
	org_mock "github.com/erda-project/erda/internal/pkg/mock"
	"github.com/erda-project/erda/internal/pkg/user"
	"github.com/erda-project/erda/internal/tools/cluster-manager/cluster"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/events"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/impl/clusterinfo"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/addon"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func TestStartRuntime(t *testing.T) {
	assert := require.New(t)
	runtimeSvc := NewRuntimeService()
	_, err := runtimeSvc.StartRuntime(context.Background(), &pb.StartRuntimeRequest{RuntimeID: "1"})
	assert.Nil(err)
}

func TestRestartRuntime(t *testing.T) {
	assert := require.New(t)
	runtimeSvc := NewRuntimeService()
	_, err := runtimeSvc.RestartRuntime(context.Background(), &pb.RestartRuntimeRequest{RuntimeID: "1"})
	assert.Nil(err)
}

func TestServiceRedeploy(t *testing.T) {
	assert := require.New(t)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	gormDB, err := gorm.Open("mysql", db)
	gormDB = gormDB.Debug()
	dbSvc := dbServiceImpl{db: &dbclient.DBClient{
		DBEngine: &dbengine.DBEngine{
			DB: gormDB,
		},
	}}

	row := sqlmock.NewRows([]string{"id", "application_id", "workspace", "name"}).AddRow(2, 2, "DEV", "master")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `ps_v2_project_runtimes` WHERE (id = ?)")).
		WithArgs(uint64(2)).
		WillReturnRows(row)

	row = sqlmock.NewRows([]string{"release_id", "status"}).AddRow(1, apistructs.DeploymentStatusOK)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `ps_v2_deployments` WHERE (runtime_id = ?) AND (status = ?) ORDER BY id desc LIMIT 1")).
		WithArgs(2, apistructs.DeploymentStatusOK).
		WillReturnRows(row)

	bdl := bundle.New(bundle.WithErdaServer(), bundle.WithClusterManager(), bundle.WithScheduler())
	runtimeSvc := NewRuntimeService(WithBundleService(bdl), WithDBService(&dbSvc))

	gomonkey.ApplyMethod(reflect.TypeOf(&bundle.Bundle{}), "GetApp", func(s *bundle.Bundle, id uint64) (*apistructs.ApplicationDTO, error) {
		return &apistructs.ApplicationDTO{
			ID: 1,
		}, nil
	})

	gomonkey.ApplyMethod(reflect.TypeOf(&bundle.Bundle{}), "CheckPermission", func(s *bundle.Bundle, req *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
		return &apistructs.PermissionCheckResponseData{
			Access: true,
		}, nil
	})

	gomonkey.ApplyPrivateMethod(reflect.TypeOf(&Service{}), "doDeployRuntime", func(s *Service, ctx *DeployContext) (*apistructs.DeploymentCreateResponseDTO, error) {
		return nil, nil
	})

	_, err = runtimeSvc.Redeploy(user.ID("1"), 1, uint64(2))
	assert.Nil(err)
}

func generateMultiAddons(t *testing.T, randCount int) string {
	assert := require.New(t)
	var addonsCount int
	if randCount != 0 {
		addonsCount = rand.Intn(randCount)
	}

	addons := make(diceyml.AddOns)
	for i := 0; i < addonsCount; i++ {
		name := fmt.Sprintf("existAddon%d", i)
		plan := fmt.Sprintf("existAddon%d:%s", i, apistructs.AddonUltimate)
		addons[name] = &diceyml.AddOn{
			Plan: plan,
			Options: map[string]string{
				"version": "1.0.0",
			},
		}
	}

	if addonsCount != 0 {
		addons["nonExistAddon1"] = &diceyml.AddOn{
			Plan: fmt.Sprintf("nonExistAddon1:%s", apistructs.AddonBasic),
			Options: map[string]string{
				"version": "1.0.0",
			},
		}
	}

	diceObj := diceyml.Object{
		Version: "2.0",
		AddOns:  addons,
	}

	diceYaml, err := yaml.Marshal(diceObj)
	assert.Nil(err)

	return string(diceYaml)
}

func TestPreCheck(t *testing.T) {
	assert := require.New(t)

	r := NewRuntimeService()
	a := addon.New()
	type args struct {
		diceYaml  string
		workspace string
	}

	gomonkey.ApplyMethod(reflect.TypeOf(a), "GetAddonExtention", func(a *addon.Addon, params *apistructs.AddonHandlerCreateItem) (*apistructs.AddonExtension, *diceyml.Object, error) {
		addonName := params.AddonName
		if strings.Contains(addonName, "nonExistAddon") {
			return nil, nil, errors.New("not found")
		}

		if addonName == apistructs.AddonCustomCategory {
			return &apistructs.AddonExtension{
				SubCategory: apistructs.BasicAddon,
				Category:    apistructs.AddonCustomCategory,
			}, nil, nil
		}

		return &apistructs.AddonExtension{
			SubCategory: apistructs.BasicAddon,
		}, nil, nil
	})

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "basic addon deploy to no-prod environment",
			args: args{
				diceYaml: `
version: 2
addons:
  rds:
    plan: custom:basic
    options:
      version: 1.0.0
  addon-1:
    plan: redis:basic
    options:
      version: 3.2.12
`,
				workspace: apistructs.WORKSPACE_TEST,
			},
			wantErr: false,
		},
		{
			name: "basic addon deploy to prod environment",
			args: args{
				diceYaml: `
version: 2
addons:
  rds:
    plan: custom:basic
    options:
      version: 1.0.0
  addon-1:
    plan: redis:basic
    options:
      version: 3.2.12
`,
				workspace: apistructs.WORKSPACE_PROD,
			},
			// TODO: precheck should return error
			wantErr: false,
		},
		{
			name: "professional addon deploy to prod environment",
			args: args{
				diceYaml: `
version: 2
addons:
  rds:
    plan: custom
    options:
      version: 1.0.0
  addon-1:
    plan: redis:professional
    options:
      version: 3.2.12
`,
				workspace: apistructs.WORKSPACE_PROD,
			},
			wantErr: false,
		},
		{
			name: "multi addons deploy to prod environment, had non-exist addon and plan error addon",
			args: args{
				diceYaml:  generateMultiAddons(t, 200),
				workspace: apistructs.WORKSPACE_PROD,
			},
			// TODO: precheck should return error
			wantErr: false,
		},
		{
			name: "non addons deploy to prod environment",
			args: args{
				diceYaml:  generateMultiAddons(t, 0),
				workspace: apistructs.WORKSPACE_PROD,
			},
			wantErr: false,
		},
		{
			name: "illegal addon plan format",
			args: args{
				diceYaml: `
version: 2
addons:
  rds:
    plan: custom:basic:err
    options:
      version: 1.0.0
`,
				workspace: apistructs.WORKSPACE_PROD,
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dice, err := diceyml.New([]byte(test.args.diceYaml), false)
			if err != nil {
				assert.Nil(err)
			}

			if err := r.PreCheck(dice, test.args.workspace); (err != nil) != test.wantErr {
				t.Errorf("PreCheck error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestUpdateStatusToDisplay(t *testing.T) {
	assert := require.New(t)

	runtime := &pb.RuntimeInspect{Status: "Unknown",
		Services: map[string]*pb.Service{
			"test": {
				Status: "Stopped",
			},
		}}

	updateStatusToDisplay(runtime)
	assert.Equal("Unknown", runtime.Status)
	for _, s := range runtime.Services {
		assert.Equal("Stopped", s.Status)
	}
}

func TestFillRuntimeDataWithServiceGroup(t *testing.T) {
	var (
		assert        = require.New(t)
		data          apistructs.RuntimeInspectDTO
		targetService diceyml.Services
		targetJob     diceyml.Jobs
		sg            apistructs.ServiceGroup
		domainMap     = make(map[string][]string, 0)
		status        string
	)

	fakeData := `{"id":1,"name":"develop","serviceGroupName":"1f1a1k1e11","serviceGroupNamespace":"services","source":"PIPELINE","status":"Healthy","deployStatus":"CANCELED","deleteStatus":"","releaseId":"11f1a1k1e11111111111111111111111","clusterId":1,"clusterName":"fake-cluster","clusterType":"k8s","resources":{"cpu":0.6,"mem":3072,"disk":0},"extra":{"applicationId":1,"buildId":0,"workspace":"DEV"},"projectID":1,"services":{"fake-service":{"status":"Healthy","deployments":{"replicas":1},"resources":{"cpu":0.3,"mem":1536,"disk":0},"envs":{"fakeEnv":"1"},"addrs":["fake-service.services--1f1a1k1e11.svc.cluster.local:8060"],"expose":["http://fake-service-dev-1-app.dev.fake.io"],"errors":null}},"lastMessage":{},"timeCreated":"2021-06-11T16:40:49+08:00","createdAt":"2021-06-11T16:40:49+08:00","updatedAt":"2021-06-17T13:53:49+08:00","errors":null}`
	err := json.Unmarshal([]byte(fakeData), &data)
	assert.Nil(err)

	fakeTargetService := `{"fake-service":{"image":"registry.fake.com/dice/fake-service:fake","image_username":"","image_password":"","cmd":"","ports":[{"port":8060,"protocol":"TCP","l4_protocol":"TCP","expose":true,"default":false}],"envs":{"fakeEnv":"1"},"resources":{"cpu":0.3,"mem":1536,"max_cpu":1,"max_mem":1536,"disk":0,"network":{"mode":"container"}},"deployments":{"replicas":2,"policies":""},"health_check":{"http":{"port":8060,"path":"/fake/health","duration":200},"exec":{}},"traffic_security":{}}}`
	err = json.Unmarshal([]byte(fakeTargetService), &targetService)
	assert.Nil(err)

	fakeSG := `{"created_time":1623400858,"last_modified_time":1623908973,"executor":"fake","clusterName":"fake-cluster","force":true,"name":"1f1a1k1e11","namespace":"services","services":[{"name":"fake-service","namespace":"services--1f1a1k1e11","image":"registry.fake.com/dice/fake-service:fake","image_username":"","image_password":"","Ports":[{"port":8060,"protocol":"TCP","l4_protocol":"TCP","expose":true,"default":false}],"proxyPorts":[8060],"vip":"fake-service.services--1f1a1k1e11.svc.cluster.local","shortVIP":"192.168.1.1","proxyIp":"192.168.1.1","scale":2,"resources":{"cpu":0.8,"mem":1537},"health_check":{"http":{"port":8060,"path":"/fake","duration":200}},"traffic_security":{},"status":"Healthy","reason":"","unScheduledReasons":{}}],"serviceDiscoveryKind":"","serviceDiscoveryMode":"DEPEND","projectNamespace":"","status":"Healthy","reason":"","unScheduledReasons":{}}`
	err = json.Unmarshal([]byte(fakeSG), &sg)
	assert.Nil(err)

	domainMap["fake-service"] = []string{"http://fake-services-dev-1-app.fake.io"}
	status = "CANCELED"

	fillRuntimeDataWithServiceGroup(&data, targetService, targetJob, &sg, domainMap, status)
	assert.Equal("", data.ModuleErrMsg["fake-service"]["Msg"])
	assert.Equal("", data.ModuleErrMsg["fake-service"]["Reason"])
	assert.Equal(1.6, data.Resources.CPU)
	assert.Equal(3074, data.Resources.Mem)
	assert.Equal(0, data.Resources.Disk)

	assert.Equal(0.8, data.Services["fake-service"].Resources.CPU)
	assert.Equal(1537, data.Services["fake-service"].Resources.Mem)
	assert.Equal(0, data.Services["fake-service"].Resources.Disk)
	assert.Equal("Healthy", data.Services["fake-service"].Status)
	assert.Equal(2, data.Services["fake-service"].Deployments.Replicas)
}

func TestGetRollbackConfig(t *testing.T) {
	var bdl *bundle.Bundle
	assert := require.New(t)
	gomonkey.ApplyMethod(reflect.TypeOf(bdl), "GetAllProjects",
		func(*bundle.Bundle) ([]apistructs.ProjectDTO, error) {
			return []apistructs.ProjectDTO{
				{ID: 1, RollbackConfig: map[string]int{"DEV": 3, "TEST": 5, "STAGING": 4, "PROD": 6}},
				{ID: 2, RollbackConfig: map[string]int{"DEV": 4, "TEST": 6, "STAGING": 5, "PROD": 7}},
				{ID: 3, RollbackConfig: map[string]int{"DEV": 5, "TEST": 7, "STAGING": 6, "PROD": 8}},
			}, nil
		},
	)

	r := NewRuntimeService(WithBundleService(bdl))
	cfg, err := r.getRollbackConfig()
	assert.Nil(err)
	assert.Equal(3, cfg[1]["DEV"])
	assert.Equal(6, cfg[2]["TEST"])
	assert.Equal(6, cfg[3]["STAGING"])
	assert.Equal(8, cfg[3]["PROD"])
}

func Test_getRedeployPipelineYmlName(t *testing.T) {
	type args struct {
		runtime dbclient.Runtime
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases
		{
			name: "Filled in the space and scene set",
			args: args{
				runtime: dbclient.Runtime{
					ApplicationID: 1,
					Workspace:     "PORD",
					Name:          "master",
				},
			},
			want: "1/PORD/master/pipeline.yml",
		},
		{
			name: "Filled in the space and scene set",
			args: args{
				runtime: dbclient.Runtime{
					ApplicationID: 4,
					Workspace:     "TEST",
					Name:          "master",
				},
			},
			want: "4/TEST/master/pipeline.yml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRedeployPipelineYmlName(tt.args.runtime); got != tt.want {
				t.Errorf("getRedeployPipelineYmlName() = %v, want %v", got, tt.want)
			}
		})
	}
}

var diceYml = `version: 2.0
services:
  web:
    ports:
      - 8080
      - port: 20880
      - port: 1234
        protocol: "UDP"
      - port: 4321
        protocol: "HTTP"
      - port: 53
        protocol: "DNS"
        l4_protocol: "UDP"
        default: true
    deployments:
      replicas: 1
    resources:
      cpu: 0.1
      mem: 512
    k8s_snippet:
      container:
        name: abc
        stdin: true
        workingDir: aaa
        imagePullPolicy: Always
        securityContext:
          privileged: true
`

func TestGetServicesNames(t *testing.T) {
	assert := require.New(t)
	name, err := getServicesNames(diceYml)
	if err != nil {
		assert.Error(err)
		return
	}
	assert.Equal([]string{"web"}, name)
}

func Test_setClusterName(t *testing.T) {
	assert := require.New(t)
	var bdl *bundle.Bundle
	var clusterinfoImpl *clusterinfo.ClusterInfoImpl
	m1 := gomonkey.ApplyMethod(reflect.TypeOf(clusterinfoImpl), "Info", func(_ *clusterinfo.ClusterInfoImpl, clusterName string) (apistructs.ClusterInfoData, error) {
		if clusterName == "erda-edge" {
			return apistructs.ClusterInfoData{apistructs.JOB_CLUSTER: "erda-center", apistructs.DICE_IS_EDGE: "true"}, nil
		}
		return apistructs.ClusterInfoData{apistructs.DICE_IS_EDGE: "false"}, nil
	})
	defer m1.Reset()
	runtimeSvc := NewRuntimeService(WithBundleService(bdl), WithClusterInfoImpl(clusterinfoImpl))
	rt := &dbclient.Runtime{
		ClusterName: "erda-edge",
	}
	runtimeSvc.setClusterName(rt)
	assert.Equal("erda-center", rt.ClusterName)
}

func Test_generateListGroupAppResult(t *testing.T) {
	var bdl *bundle.Bundle
	var db *dbclient.DBClient
	assert := require.New(t)
	runtime := NewRuntimeService(WithBundleService(bdl), WithDBService(db))
	var result = struct {
		sync.RWMutex
		m map[uint64][]*apistructs.RuntimeSummaryDTO
	}{m: make(map[uint64][]*apistructs.RuntimeSummaryDTO)}
	var wg sync.WaitGroup
	r := dbclient.Runtime{
		BaseModel: dbengine.BaseModel{
			ID: 111,
		},
		Name:          "master",
		ApplicationID: 1,
		Workspace:     "",
	}
	d := dbclient.Deployment{
		BaseModel: dbengine.BaseModel{
			ID: 3,
		},
		RuntimeId: 111,
		ReleaseId: "aaaa-bbbbb-cccc",
		Operator:  "erda",
		Status:    "OK",
	}
	wg.Add(1)
	runtime.generateListGroupAppResult(&result, 1, &r, d, &wg)
	assert.Equal(apistructs.DeploymentStatus("OK"), result.m[1][0].DeployStatus)
}

func Test_listGroupByApps(t *testing.T) {
	var bdl *bundle.Bundle
	var db *dbclient.DBClient
	assert := require.New(t)
	m1 := gomonkey.ApplyMethod(reflect.TypeOf(db), "FindRuntimesInApps", func(_ *dbclient.DBClient, appIDs []uint64, env string) (map[uint64][]*dbclient.Runtime, []uint64, error) {
		a := make(map[uint64][]*dbclient.Runtime)
		a[1] = []*dbclient.Runtime{{
			BaseModel: dbengine.BaseModel{
				ID: 1,
			},
			Name:          "master",
			Workspace:     "DEV",
			ApplicationID: 1,
		}}
		return a, []uint64{1}, nil
	})
	defer m1.Reset()

	m2 := gomonkey.ApplyMethod(reflect.TypeOf(db), "FindLastDeploymentIDsByRutimeIDs", func(_ *dbclient.DBClient, runtimeIDs []uint64) ([]uint64, error) {
		return []uint64{5}, nil
	})
	defer m2.Reset()

	m3 := gomonkey.ApplyMethod(reflect.TypeOf(db), "FindDeploymentsByIDs", func(_ *dbclient.DBClient, ids []uint64) (map[uint64]dbclient.Deployment, error) {
		a := make(map[uint64]dbclient.Deployment)
		a[1] = dbclient.Deployment{
			BaseModel: dbengine.BaseModel{
				ID: 5,
			},
			RuntimeId: 1,
			Status:    "OK",
		}
		return a, nil
	})
	defer m3.Reset()

	runtime := NewRuntimeService(WithBundleService(bdl), WithDBService(db))
	result, _ := runtime.ListGroupByApps([]uint64{1}, "DEV")
	assert.Equal(apistructs.DeploymentStatus("OK"), result[1][0].DeployStatus)
}

type orgMock struct {
	org_mock.OrgMock
}

func (m orgMock) GetOrg(ctx context.Context, request *orgpb.GetOrgRequest) (*orgpb.GetOrgResponse, error) {
	if request.IdOrName == "" {
		return nil, fmt.Errorf("the IdOrName is empty")
	}
	return &orgpb.GetOrgResponse{Data: &orgpb.Org{}}, nil
}

func TestRuntimeGetOrg(t *testing.T) {
	type fields struct {
		org org.ClientInterface
	}
	type args struct {
		orgID uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *orgpb.Org
		wantErr bool
	}{
		{
			name: "test with error",
			fields: fields{
				org: orgMock{},
			},
			args:    args{orgID: 0},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test with no error",
			fields: fields{
				org: orgMock{},
			},
			args:    args{orgID: 1},
			want:    &orgpb.Org{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Service{
				org: tt.fields.org,
			}
			got, err := r.GetOrg(tt.args.orgID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOrg() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBatchRuntimeReDeploy(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	gormDB, err := gorm.Open("mysql", db)
	gormDB = gormDB.Debug()
	dbSvc := dbServiceImpl{db: &dbclient.DBClient{
		DBEngine: &dbengine.DBEngine{
			DB: gormDB,
		},
	}}

	s := NewRuntimeService(WithDBService(&dbSvc))

	userID := user.ID("2")

	runtime1 := dbclient.Runtime{
		BaseModel: dbengine.BaseModel{
			ID: 129,
		},
		Name:          "feature/develop",
		ApplicationID: 21,
		Workspace:     "DEV",
		GitBranch:     "feature/develop",
		ProjectID:     1,
		Env:           "DEV",
		ClusterName:   "test",
		ClusterId:     1,
		Creator:       "2",
		ScheduleName: dbclient.ScheduleName{
			Namespace: "services",
			Name:      "3dbfa5bf4c2",
		},
		Status:           "Healthy",
		LegacyStatus:     "INIT",
		Deployed:         true,
		Version:          "1",
		Source:           "IPELINE",
		DiceVersion:      "2",
		CPU:              0.10,
		Mem:              128.00,
		ReadableUniqueId: "dice-orchestrator",
		GitRepoAbbrev:    "xxx-test/test02",
		OrgID:            1,
	}

	runtime2 := dbclient.Runtime{
		BaseModel: dbengine.BaseModel{
			ID: 128,
		},
		Name:          "feature/develop",
		ApplicationID: 1,
		Workspace:     "DEV",
		GitBranch:     "feature/develop",
		ProjectID:     1,
		Env:           "DEV",
		ClusterName:   "test",
		ClusterId:     1,
		Creator:       "2",
		ScheduleName: dbclient.ScheduleName{
			Namespace: "services",
			Name:      "302615dbf0",
		},
		Status:           "Healthy",
		LegacyStatus:     "INIT",
		Deployed:         true,
		Version:          "1",
		Source:           "IPELINE",
		DiceVersion:      "2",
		CPU:              0.10,
		Mem:              128.00,
		ReadableUniqueId: "dice-orchestrator",
		GitRepoAbbrev:    "xxx-test/test01",
		OrgID:            1,
	}

	runtime3 := dbclient.Runtime{
		BaseModel: dbengine.BaseModel{
			ID: 130,
		},
		Name:          "feature/develop",
		ApplicationID: 22,
		Workspace:     "DEV",
		GitBranch:     "feature/develop",
		ProjectID:     1,
		Env:           "DEV",
		ClusterName:   "test",
		ClusterId:     1,
		Creator:       "2",
		ScheduleName: dbclient.ScheduleName{
			Namespace: "services",
			Name:      "302615dbf1",
		},
		Status:           "Healthy",
		LegacyStatus:     "INIT",
		Deployed:         true,
		Version:          "1",
		Source:           "IPELINE",
		DiceVersion:      "2",
		CPU:              0.10,
		Mem:              128.00,
		ReadableUniqueId: "dice-orchestrator",
		GitRepoAbbrev:    "xxx-test/test03",
		OrgID:            1,
	}
	runtimes := []dbclient.Runtime{runtime1, runtime2}
	runtimeScaleRecords1 := apistructs.RuntimeScaleRecords{
		IDs: []uint64{128, 129},
	}

	rsr1 := apistructs.RuntimeScaleRecord{
		ApplicationId: 1,
		Workspace:     "DEV",
		Name:          "feature/develop",
		PayLoad: apistructs.PreDiceDTO{
			Services: make(map[string]*apistructs.RuntimeInspectServiceDTO),
		},
	}
	rsr1.PayLoad.Services["go-demo"] = &apistructs.RuntimeInspectServiceDTO{
		Deployments: apistructs.RuntimeServiceDeploymentsDTO{
			Replicas: 1,
		},
		Resources: apistructs.RuntimeServiceResourceDTO{
			CPU: 0.1,
			Mem: 128,
		},
	}

	rsr2 := apistructs.RuntimeScaleRecord{
		ApplicationId: 21,
		Workspace:     "DEV",
		Name:          "feature/develop",
		PayLoad: apistructs.PreDiceDTO{
			Services: make(map[string]*apistructs.RuntimeInspectServiceDTO),
		},
	}
	rsr1.PayLoad.Services["go-demo"] = &apistructs.RuntimeInspectServiceDTO{
		Deployments: apistructs.RuntimeServiceDeploymentsDTO{
			Replicas: 1,
		},
		Resources: apistructs.RuntimeServiceResourceDTO{
			CPU: 0.1,
			Mem: 128,
		},
	}

	rsr3 := apistructs.RuntimeScaleRecord{
		ApplicationId: 22,
		Workspace:     "DEV",
		Name:          "feature/develop",
		PayLoad: apistructs.PreDiceDTO{
			Services: make(map[string]*apistructs.RuntimeInspectServiceDTO),
		},
	}
	rsr3.PayLoad.Services["xxxyyyy"] = &apistructs.RuntimeInspectServiceDTO{
		Deployments: apistructs.RuntimeServiceDeploymentsDTO{
			Replicas: 1,
		},
		Resources: apistructs.RuntimeServiceResourceDTO{
			CPU: 0.1,
			Mem: 128,
		},
	}

	runtimeScaleRecords2 := apistructs.RuntimeScaleRecords{
		Runtimes: []apistructs.RuntimeScaleRecord{rsr1, rsr2},
	}

	runtimeScaleRecords3 := apistructs.RuntimeScaleRecords{
		Runtimes: []apistructs.RuntimeScaleRecord{rsr3},
		IDs:      []uint64{130},
	}

	gomonkey.ApplyMethod(reflect.TypeOf(&dbServiceImpl{}), "FindRuntime", func(db *dbServiceImpl, id spec.RuntimeUniqueId) (*dbclient.Runtime, error) {
		uniqeKey := fmt.Sprintf("%s-%s-%s", strconv.Itoa(int(id.ApplicationId)), id.Name, id.Workspace)
		if uniqeKey == "21-feature/develop-DEV" {
			runtime := &dbclient.Runtime{
				BaseModel: dbengine.BaseModel{
					ID: 129,
				},
				Name:          "feature/develop",
				ApplicationID: 21,
				Workspace:     "DEV",
				GitBranch:     "feature/develop",
				ProjectID:     1,
				Env:           "DEV",
				ClusterName:   "test",
				ClusterId:     1,
				Creator:       "2",
				ScheduleName: dbclient.ScheduleName{
					Namespace: "services",
					Name:      "3dbfa5bf4c2",
				},
				Status:           "Healthy",
				LegacyStatus:     "INIT",
				Deployed:         true,
				Version:          "1",
				Source:           "IPELINE",
				DiceVersion:      "2",
				CPU:              0.10,
				Mem:              128.00,
				ReadableUniqueId: "dice-orchestrator",
				GitRepoAbbrev:    "xxx-test/test02",
				OrgID:            1,
			}
			return runtime, nil
		}
		if uniqeKey == "1-feature/develop-DEV" {
			runtime := &dbclient.Runtime{
				BaseModel: dbengine.BaseModel{
					ID: 128,
				},
				Name:          "feature/develop",
				ApplicationID: 1,
				Workspace:     "DEV",
				GitBranch:     "feature/develop",
				ProjectID:     1,
				Env:           "DEV",
				ClusterName:   "test",
				ClusterId:     1,
				Creator:       "2",
				ScheduleName: dbclient.ScheduleName{
					Namespace: "services",
					Name:      "302615dbf0",
				},
				Status:           "Healthy",
				LegacyStatus:     "INIT",
				Deployed:         true,
				Version:          "1",
				Source:           "IPELINE",
				DiceVersion:      "2",
				CPU:              0.10,
				Mem:              128.00,
				ReadableUniqueId: "dice-orchestrator",
				GitRepoAbbrev:    "xxx-test/test01",
				OrgID:            1,
			}
			return runtime, nil
		} else {
			runtime := &dbclient.Runtime{
				BaseModel: dbengine.BaseModel{
					ID: 130,
				},
				Name:          "feature/develop",
				ApplicationID: 22,
				Workspace:     "DEV",
				GitBranch:     "feature/develop",
				ProjectID:     1,
				Env:           "DEV",
				ClusterName:   "test",
				ClusterId:     1,
				Creator:       "2",
				ScheduleName: dbclient.ScheduleName{
					Namespace: "services",
					Name:      "302615dbf1",
				},
				Status:           "Healthy",
				LegacyStatus:     "INIT",
				Deployed:         true,
				Version:          "1",
				Source:           "IPELINE",
				DiceVersion:      "2",
				CPU:              0.10,
				Mem:              128.00,
				ReadableUniqueId: "dice-orchestrator",
				GitRepoAbbrev:    "xxx-test/test03",
				OrgID:            1,
			}
			return runtime, nil
		}
	})

	gomonkey.ApplyMethod(reflect.TypeOf(s), "RedeployPipeline", func(rt *Service, ctx context.Context, operator user.ID, orgID uint64, runtimeID uint64) (*apistructs.RuntimeDeployDTO, error) {
		if runtimeID == 128 {
			ret := &apistructs.RuntimeDeployDTO{
				PipelineID:      10000260,
				ApplicationID:   1,
				ApplicationName: "test01",
				ProjectID:       1,
				ProjectName:     "test",
				OrgID:           1,
				OrgName:         "xxx",
				ServicesNames:   []string{"go-demo"},
			}
			return ret, nil
		}
		if runtimeID == 129 {
			ret := &apistructs.RuntimeDeployDTO{
				PipelineID:      10000259,
				ApplicationID:   21,
				ApplicationName: "test02",
				ProjectID:       1,
				ProjectName:     "test",
				OrgID:           1,
				OrgName:         "xxx",
				ServicesNames:   []string{"xxx"},
			}
			return ret, nil
		} else {
			ret := &apistructs.RuntimeDeployDTO{
				ApplicationID:   22,
				ApplicationName: "test03",
				ProjectID:       1,
				ProjectName:     "test",
				OrgID:           1,
				OrgName:         "xxx",
				ServicesNames:   []string{"xxxyyy"},
			}
			return ret, errors.New("failed")
		}
	})

	want1 := apistructs.BatchRuntimeReDeployResults{
		Total:   0,
		Success: 2,
		Failed:  0,
		ReDeployed: []apistructs.RuntimeDeployDTO{
			{
				PipelineID:      10000260,
				ApplicationID:   1,
				ApplicationName: "test01",
				ProjectID:       1,
				ProjectName:     "test",
				OrgID:           1,
				OrgName:         "xxx",
				ServicesNames:   []string{"go-demo"},
			},
			{
				PipelineID:      10000259,
				ApplicationID:   21,
				ApplicationName: "test02",
				ProjectID:       1,
				ProjectName:     "test",
				OrgID:           1,
				OrgName:         "xxx",
				ServicesNames:   []string{"xxx"},
			},
		},
		ReDeployedIds:   []uint64{128, 129},
		UnReDeployed:    []apistructs.RuntimeDTO{},
		UnReDeployedIds: []uint64{},
		ErrMsg:          []string{},
	}
	want2 := want1
	want2.Total = 2

	want3 := apistructs.BatchRuntimeReDeployResults{
		Total:         0,
		Success:       0,
		Failed:        1,
		ReDeployed:    []apistructs.RuntimeDeployDTO{},
		ReDeployedIds: []uint64{},
		UnReDeployed: []apistructs.RuntimeDTO{
			{
				ID:              130,
				Name:            "xxxyyy",
				GitBranch:       "feature/develop",
				Workspace:       "DEV",
				ClusterName:     "test",
				ClusterId:       1,
				Status:          "",
				ApplicationID:   22,
				ApplicationName: "xxxyyy",
				ProjectID:       1,
				ProjectName:     "test",
				OrgID:           1,
			},
		},
		UnReDeployedIds: []uint64{130},
		ErrMsg:          []string{"failed"},
	}
	ctx := context.Background()
	got := s.batchRuntimeReDeploy(ctx, userID, runtimes, runtimeScaleRecords1)
	if len(got.ReDeployedIds) != len(want1.ReDeployedIds) {
		t.Errorf("batchRuntimeReDeploy() = %v, want %v", got, want1)
	}

	gotMap := make(map[uint64]bool)
	for _, id := range got.ReDeployedIds {
		gotMap[id] = true
	}
	for _, id := range want1.ReDeployedIds {
		if _, ok := gotMap[id]; !ok {
			t.Errorf("batchRuntimeReDeploy() = %v, want %v", got, want1)
		}
	}

	var rts []dbclient.Runtime
	got = s.batchRuntimeReDeploy(ctx, userID, rts, runtimeScaleRecords2)
	if len(got.ReDeployedIds) != len(want2.ReDeployedIds) {
		t.Errorf("batchRuntimeReDeploy() = %v, want %v", got, want2)
	}
	gotMap1 := make(map[uint64]bool)
	for _, id := range got.ReDeployedIds {
		gotMap1[id] = true
	}
	for _, id := range want2.ReDeployedIds {
		if _, ok := gotMap1[id]; !ok {
			t.Errorf("batchRuntimeReDeploy() = %v, want %v", got, want2)
		}
	}

	runtimes3 := []dbclient.Runtime{runtime3}
	got = s.batchRuntimeReDeploy(ctx, userID, runtimes3, runtimeScaleRecords3)
	if len(got.UnReDeployedIds) != len(want3.UnReDeployedIds) {
		t.Errorf("batchRuntimeReDeploy() = %v, want %v", got, want3)
	}

	gotMap = make(map[uint64]bool)
	for _, id := range got.UnReDeployedIds {
		gotMap[id] = true
	}
	for _, id := range want3.UnReDeployedIds {
		if _, ok := gotMap[id]; !ok {
			t.Errorf("batchRuntimeReDeploy() = %v, want %v", got, want3)
		}
	}

}

func TestGenOverlayDataForAudit(t *testing.T) {
	assert := require.New(t)
	oldServiceData := &diceyml.Service{
		Resources: diceyml.Resources{
			CPU:  1,
			Mem:  1024,
			Disk: 0,
		},
		Deployments: diceyml.Deployments{
			Replicas: 1,
		},
	}

	auditData := genOverlayDataForAudit(oldServiceData)

	assert.Equal(float64(1), auditData.Resources.CPU)
	assert.Equal(1024, auditData.Resources.Mem)
	assert.Equal(0, auditData.Resources.Disk)
	assert.Equal(1, auditData.Deployments.Replicas)
}

func TestGetRuntimeScaleRecordByRuntimeIds(t *testing.T) {

	s := &Service{
		db: &dbclient.DBClient{},
	}

	ids := []uint64{128, 129}

	gomonkey.ApplyMethod(reflect.TypeOf(s.db), "FindRuntimesByIds", func(db *dbclient.DBClient, ids []uint64) ([]dbclient.Runtime, error) {
		runtimes := make([]dbclient.Runtime, 0)
		runtimes = append(runtimes, dbclient.Runtime{
			BaseModel: dbengine.BaseModel{
				ID: 128,
			},
			Name:          "feature/develop",
			ApplicationID: 1,
			Workspace:     "DEV",
			GitBranch:     "feature/develop",
			ProjectID:     1,
			Env:           "DEV",
			ClusterName:   "test",
			ClusterId:     1,
			Creator:       "2",
			ScheduleName: dbclient.ScheduleName{
				Namespace: "services",
				Name:      "302615dbf0",
			},
			Status:           "Healthy",
			LegacyStatus:     "INIT",
			Deployed:         true,
			Version:          "1",
			Source:           "IPELINE",
			DiceVersion:      "2",
			CPU:              0.10,
			Mem:              128.00,
			ReadableUniqueId: "dice-orchestrator",
			GitRepoAbbrev:    "xxx-test/test01",
			OrgID:            1,
		})
		runtimes = append(runtimes, dbclient.Runtime{
			BaseModel: dbengine.BaseModel{
				ID: 129,
			},
			Name:          "feature/develop",
			ApplicationID: 21,
			Workspace:     "DEV",
			GitBranch:     "feature/develop",
			ProjectID:     1,
			Env:           "DEV",
			ClusterName:   "test",
			ClusterId:     1,
			Creator:       "2",
			ScheduleName: dbclient.ScheduleName{
				Namespace: "services",
				Name:      "3dbfa5bf4c2",
			},
			Status:           "Healthy",
			LegacyStatus:     "INIT",
			Deployed:         true,
			Version:          "1",
			Source:           "IPELINE",
			DiceVersion:      "2",
			CPU:              0.10,
			Mem:              128.00,
			ReadableUniqueId: "dice-orchestrator",
			GitRepoAbbrev:    "xxx-test/test02",
			OrgID:            1,
		})

		return runtimes, nil
	})

	gomonkey.ApplyMethod(reflect.TypeOf(s.db), "FindPreDeployment", func(db *dbclient.DBClient, uniqueId spec.RuntimeUniqueId) (*dbclient.PreDeployment, error) {

		uniqeKey := fmt.Sprintf("%s-%s-%s", strconv.Itoa(int(uniqueId.ApplicationId)), uniqueId.Name, uniqueId.Workspace)
		if uniqeKey == "21-feature/develop-DEV" {
			dice := &diceyml.Object{
				Version: "2.0",
				Services: diceyml.Services{
					"xxx": &diceyml.Service{
						Image:         "addon-registry.default.svc.cluster.local:5000/xxx-test/test02:test02-1641494278825921631",
						ImageUsername: "",
						ImagePassword: "",
						Cmd:           "",
						Ports: []diceyml.ServicePort{
							{
								Port:   8080,
								Expose: true,
							},
						},
						Resources: diceyml.Resources{
							CPU:     0.1,
							Mem:     128,
							Network: map[string]string{"mode": "caontainer"},
						},
						Deployments: diceyml.Deployments{
							Replicas: 1,
						},
					},
				},
				AddOns: diceyml.AddOns{
					"mysql": &diceyml.AddOn{
						Plan:    "mysql:basic",
						Options: map[string]string{"create_dbs": "testdb1,testdb2", "version": "5.7.29"},
					},
				},
			}

			b, _ := json.Marshal(dice)
			diceJson := string(b)

			dice_overlay := diceyml.Object{
				Services: diceyml.Services{
					"xxx": &diceyml.Service{
						Resources: diceyml.Resources{
							CPU: 0.1,
							Mem: 128,
						},
						Deployments: diceyml.Deployments{
							Replicas: 1,
						},
					},
				},
			}

			b, _ = json.Marshal(&dice_overlay)
			DiceOverlayJson := string(b)

			return &dbclient.PreDeployment{
				BaseModel: dbengine.BaseModel{
					ID: 23,
				},
				ApplicationId: 21,
				Workspace:     "DEV",
				RuntimeName:   "feature/develop",
				Dice:          diceJson,
				DiceOverlay:   DiceOverlayJson,
				DiceType:      1,
			}, nil
		} else {
			dice := &diceyml.Object{
				Version: "2.0",
				Services: diceyml.Services{
					"go-demo": &diceyml.Service{
						Image:         "addon-registry.default.svc.cluster.local:5000/xxx-test/test01:go-demo-1641494267330770612",
						ImageUsername: "",
						ImagePassword: "",
						Cmd:           "",
						Ports: []diceyml.ServicePort{
							{
								Port:   8080,
								Expose: true,
							},
						},
						Resources: diceyml.Resources{
							CPU:     0.1,
							Mem:     128,
							Network: map[string]string{"mode": "caontainer"},
						},
						Deployments: diceyml.Deployments{
							Replicas: 1,
						},
					},
				},
				AddOns: diceyml.AddOns{
					"kafka": &diceyml.AddOn{
						Plan:    "kafka:basic",
						Options: map[string]string{"version": "2.0.0"},
					},
				},
			}

			b, _ := json.Marshal(dice)
			diceJson := string(b)

			dice_overlay := diceyml.Object{
				Services: diceyml.Services{
					"go-demo": &diceyml.Service{
						Resources: diceyml.Resources{
							CPU: 0.1,
							Mem: 128,
						},
						Deployments: diceyml.Deployments{
							Replicas: 1,
						},
					},
				},
			}

			b, _ = json.Marshal(&dice_overlay)
			DiceOverlayJson := string(b)

			return &dbclient.PreDeployment{
				BaseModel: dbengine.BaseModel{
					ID: 20,
				},
				ApplicationId: 1,
				Workspace:     "DEV",
				RuntimeName:   "feature/develop",
				Dice:          diceJson,
				DiceOverlay:   DiceOverlayJson,
				DiceType:      1,
			}, nil
		}
	})

	_, _, err := s.getRuntimeScaleRecordByRuntimeIds(ids)
	assert := require.New(t)
	assert.Equal(err, nil)
}

func TestBatchRuntimeDelete(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	gormDB, err := gorm.Open("mysql", db)
	gormDB = gormDB.Debug()
	dbSvc := dbServiceImpl{db: &dbclient.DBClient{
		DBEngine: &dbengine.DBEngine{
			DB: gormDB,
		},
	}}
	s := NewRuntimeService(WithDBService(&dbSvc))

	userID := user.ID("2")

	runtime1 := dbclient.Runtime{
		BaseModel: dbengine.BaseModel{
			ID: 129,
		},
		Name:          "feature/develop",
		ApplicationID: 21,
		Workspace:     "DEV",
		GitBranch:     "feature/develop",
		ProjectID:     1,
		Env:           "DEV",
		ClusterName:   "test",
		ClusterId:     1,
		Creator:       "2",
		ScheduleName: dbclient.ScheduleName{
			Namespace: "services",
			Name:      "3dbfa5bf4c2",
		},
		Status:           "Healthy",
		LegacyStatus:     "INIT",
		Deployed:         true,
		Version:          "1",
		Source:           "IPELINE",
		DiceVersion:      "2",
		CPU:              0.10,
		Mem:              128.00,
		ReadableUniqueId: "dice-orchestrator",
		GitRepoAbbrev:    "xxx-test/test02",
		OrgID:            1,
	}

	runtime2 := dbclient.Runtime{
		BaseModel: dbengine.BaseModel{
			ID: 128,
		},
		Name:          "feature/develop",
		ApplicationID: 1,
		Workspace:     "DEV",
		GitBranch:     "feature/develop",
		ProjectID:     1,
		Env:           "DEV",
		ClusterName:   "test",
		ClusterId:     1,
		Creator:       "2",
		ScheduleName: dbclient.ScheduleName{
			Namespace: "services",
			Name:      "302615dbf0",
		},
		Status:           "Healthy",
		LegacyStatus:     "INIT",
		Deployed:         true,
		Version:          "1",
		Source:           "IPELINE",
		DiceVersion:      "2",
		CPU:              0.10,
		Mem:              128.00,
		ReadableUniqueId: "dice-orchestrator",
		GitRepoAbbrev:    "xxx-test/test01",
		OrgID:            1,
	}

	runtimes := []dbclient.Runtime{runtime1, runtime2}
	runtimeScaleRecords1 := apistructs.RuntimeScaleRecords{
		IDs: []uint64{128, 129},
	}

	rsr1 := apistructs.RuntimeScaleRecord{
		ApplicationId: 1,
		Workspace:     "DEV",
		Name:          "feature/develop",
		PayLoad: apistructs.PreDiceDTO{
			Services: make(map[string]*apistructs.RuntimeInspectServiceDTO),
		},
	}
	rsr1.PayLoad.Services["go-demo"] = &apistructs.RuntimeInspectServiceDTO{
		Deployments: apistructs.RuntimeServiceDeploymentsDTO{
			Replicas: 1,
		},
		Resources: apistructs.RuntimeServiceResourceDTO{
			CPU: 0.1,
			Mem: 128,
		},
	}

	rsr2 := apistructs.RuntimeScaleRecord{
		ApplicationId: 21,
		Workspace:     "DEV",
		Name:          "feature/develop",
		PayLoad: apistructs.PreDiceDTO{
			Services: make(map[string]*apistructs.RuntimeInspectServiceDTO),
		},
	}
	rsr1.PayLoad.Services["go-demo"] = &apistructs.RuntimeInspectServiceDTO{
		Deployments: apistructs.RuntimeServiceDeploymentsDTO{
			Replicas: 1,
		},
		Resources: apistructs.RuntimeServiceResourceDTO{
			CPU: 0.1,
			Mem: 128,
		},
	}

	runtimeScaleRecords2 := apistructs.RuntimeScaleRecords{
		Runtimes: []apistructs.RuntimeScaleRecord{rsr1, rsr2},
	}

	gomonkey.ApplyMethod(reflect.TypeOf(s), "DeleteRuntime", func(rt *Service, operator user.ID, orgID uint64, runtimeID uint64) (*apistructs.RuntimeDTO, error) {
		if runtimeID == 128 {
			ret := &apistructs.RuntimeDTO{
				ID:              128,
				Name:            "feature/develop",
				GitBranch:       "feature/develop",
				Workspace:       "DEV",
				ClusterName:     "test",
				ClusterId:       1,
				Status:          "Healthy",
				ApplicationID:   1,
				ApplicationName: "test01",
				ProjectID:       1,
				ProjectName:     "test",
				OrgID:           1,
				Errors:          []apistructs.ErrorResponse{},
			}
			return ret, nil
		} else {
			ret := &apistructs.RuntimeDTO{
				ID:              129,
				Name:            "feature/develop",
				GitBranch:       "feature/develop",
				Workspace:       "DEV",
				ClusterName:     "test",
				ClusterId:       1,
				Status:          "Healthy",
				ApplicationID:   21,
				ApplicationName: "test02",
				ProjectID:       1,
				ProjectName:     "test",
				OrgID:           1,
				Errors:          []apistructs.ErrorResponse{},
			}
			return ret, nil
		}
	})

	gomonkey.ApplyMethod(reflect.TypeOf(&dbServiceImpl{}), "FindRuntime", func(db *dbServiceImpl, uniqueId spec.RuntimeUniqueId) (*dbclient.Runtime, error) {
		uniqeKey := fmt.Sprintf("%s-%s-%s", strconv.Itoa(int(uniqueId.ApplicationId)), uniqueId.Name, uniqueId.Workspace)
		if uniqeKey == "21-feature/develop-DEV" {
			runtime := &dbclient.Runtime{
				BaseModel: dbengine.BaseModel{
					ID: 129,
				},
				Name:          "feature/develop",
				ApplicationID: 21,
				Workspace:     "DEV",
				GitBranch:     "feature/develop",
				ProjectID:     1,
				Env:           "DEV",
				ClusterName:   "test",
				ClusterId:     1,
				Creator:       "2",
				ScheduleName: dbclient.ScheduleName{
					Namespace: "services",
					Name:      "3dbfa5bf4c2",
				},
				Status:           "Healthy",
				LegacyStatus:     "INIT",
				Deployed:         true,
				Version:          "1",
				Source:           "IPELINE",
				DiceVersion:      "2",
				CPU:              0.10,
				Mem:              128.00,
				ReadableUniqueId: "dice-orchestrator",
				GitRepoAbbrev:    "xxx-test/test02",
				OrgID:            1,
			}
			return runtime, nil
		} else {
			runtime := &dbclient.Runtime{
				BaseModel: dbengine.BaseModel{
					ID: 128,
				},
				Name:          "feature/develop",
				ApplicationID: 1,
				Workspace:     "DEV",
				GitBranch:     "feature/develop",
				ProjectID:     1,
				Env:           "DEV",
				ClusterName:   "test",
				ClusterId:     1,
				Creator:       "2",
				ScheduleName: dbclient.ScheduleName{
					Namespace: "services",
					Name:      "302615dbf0",
				},
				Status:           "Healthy",
				LegacyStatus:     "INIT",
				Deployed:         true,
				Version:          "1",
				Source:           "IPELINE",
				DiceVersion:      "2",
				CPU:              0.10,
				Mem:              128.00,
				ReadableUniqueId: "dice-orchestrator",
				GitRepoAbbrev:    "xxx-test/test01",
				OrgID:            1,
			}
			return runtime, nil
		}
	})

	want1 := apistructs.BatchRuntimeDeleteResults{
		Total:   0,
		Success: 2,
		Failed:  0,
		Deleted: []apistructs.RuntimeDTO{
			{
				ID:              128,
				Name:            "",
				GitBranch:       "feature/develop",
				Workspace:       "DEV",
				ClusterName:     "test",
				ClusterId:       1,
				Status:          "",
				ApplicationID:   1,
				ApplicationName: "test01",
				ProjectID:       1,
				ProjectName:     "test",
				OrgID:           1,
			},
			{
				ID:              129,
				Name:            "",
				GitBranch:       "feature/develop",
				Workspace:       "DEV",
				ClusterName:     "test",
				ClusterId:       1,
				Status:          "",
				ApplicationID:   1,
				ApplicationName: "test02",
				ProjectID:       1,
				ProjectName:     "test",
				OrgID:           1,
			},
		},
		DeletedIds:   []uint64{128, 129},
		UnDeleted:    []apistructs.RuntimeDTO{},
		UnDeletedIds: []uint64{},
		ErrMsg:       []string{},
	}
	want2 := want1
	want2.Total = 2

	got := s.batchRuntimeDelete(userID, runtimes, runtimeScaleRecords1)
	if len(got.DeletedIds) != len(want1.DeletedIds) {
		t.Errorf("batchRuntimeReDeploy() = %v, want %v", got, want1)
	}

	gotMap := make(map[uint64]bool)
	for _, id := range got.DeletedIds {
		gotMap[id] = true
	}
	for _, id := range want1.DeletedIds {
		if _, ok := gotMap[id]; !ok {
			t.Errorf("batchRuntimeReDeploy() = %v, want %v", got, want1)
		}
	}

	var rts []dbclient.Runtime
	got = s.batchRuntimeDelete(userID, rts, runtimeScaleRecords2)
	if len(got.DeletedIds) != len(want2.DeletedIds) {
		t.Errorf("batchRuntimeReDeploy() = %v, want %v", got, want2)
	}
	gotMap1 := make(map[uint64]bool)
	for _, id := range got.DeletedIds {
		gotMap1[id] = true
	}
	for _, id := range want2.DeletedIds {
		if _, ok := gotMap1[id]; !ok {
			t.Errorf("batchRuntimeReDeploy() = %v, want %v", got, want2)
		}
	}

}

func TestService_CreateByReleaseID(t *testing.T) {
	ctx := context.Background()
	assert := require.New(t)

	releaseReq := &apistructs.RuntimeReleaseCreateRequest{
		ProjectID:     1,
		ApplicationID: 12,
		Workspace:     "dev",
	}

	releaseSvc := &release.ReleaseService{}
	service := NewRuntimeService(WithReleaseSvc(releaseSvc), WithBundleService(bundle.New()))

	m1 := gomonkey.ApplyMethod(reflect.TypeOf(releaseSvc), "GetRelease", func(rt *release.ReleaseService, ctx context.Context, req *releasepb.ReleaseGetRequest) (*releasepb.ReleaseGetResponse, error) {
		return &releasepb.ReleaseGetResponse{
			Data: &releasepb.ReleaseGetResponseData{
				ProjectID:     1,
				ApplicationID: 12,
				ClusterName:   "test",
			},
			UserIDs: nil,
		}, nil
	})
	defer m1.Reset()

	m2 := gomonkey.ApplyMethod(reflect.TypeOf(service.bundle), "GetAllValidBranchWorkspace", func(bundle *bundle.Bundle, appId uint64, userID string) ([]apistructs.ValidBranch, error) {
		return []apistructs.ValidBranch{
			{
				Workspace: "DEV",
			},
		}, nil
	})
	defer m2.Reset()

	m3 := gomonkey.ApplyMethod(reflect.TypeOf(service.bundle), "GetProjectWithSetter", func(bundle *bundle.Bundle, id uint64) (*apistructs.ProjectDTO, error) {
		return &apistructs.ProjectDTO{
			ClusterConfig: map[string]string{
				"dev": "test",
			},
		}, nil
	})
	defer m3.Reset()

	m4 := gomonkey.ApplyFunc(gitflowutil.IsValidBranchWorkspace, func(_ []apistructs.ValidBranch, _ apistructs.DiceWorkspace) (bool, bool) {
		// 这里直接返回你想要的测试值
		return true, true // 第一个返回值，第二个返回值
	})
	defer m4.Reset()

	m5 := gomonkey.ApplyMethod(reflect.TypeOf(service), "Create", func(rt *Service, operator user.ID, req *apistructs.RuntimeCreateRequest) (*apistructs.DeploymentCreateResponseDTO, error) {
		return nil, nil
	})
	defer m5.Reset()

	_, err := service.CreateByReleaseID(ctx, user.ID("1"), releaseReq)
	assert.Nil(err)
}

func TestService_checkRuntimeCreateReq(t *testing.T) {
	assert := require.New(t)
	testcases := []struct {
		name string
		req  *apistructs.RuntimeCreateRequest
		want error
	}{
		{
			name: "not name",
			req: &apistructs.RuntimeCreateRequest{
				Name:        "",
				ReleaseID:   "1",
				Operator:    "1",
				ClusterName: "jicheng-newb",
				Source:      apistructs.PIPELINE,
				Extra: apistructs.RuntimeCreateRequestExtra{
					OrgID:         1,
					Workspace:     "master",
					ProjectID:     1,
					ApplicationID: 1,
				},
			},
			want: errors.New("runtime name is not specified"),
		},
		{
			name: "not releaseId",
			req: &apistructs.RuntimeCreateRequest{
				Name:        "test",
				ReleaseID:   "",
				Operator:    "1",
				ClusterName: "jicheng-newb",
				Source:      apistructs.PIPELINE,
				Extra: apistructs.RuntimeCreateRequestExtra{
					OrgID:         1,
					Workspace:     "master",
					ProjectID:     1,
					ApplicationID: 1,
				},
			},
			want: errors.New("releaseId is not specified"),
		},
		{
			name: "not operator",
			req: &apistructs.RuntimeCreateRequest{
				Name:        "test",
				ReleaseID:   "1",
				Operator:    "",
				ClusterName: "jicheng-newb",
				Source:      apistructs.PIPELINE,
				Extra: apistructs.RuntimeCreateRequestExtra{
					OrgID:         1,
					Workspace:     "master",
					ProjectID:     1,
					ApplicationID: 1,
				},
			},
			want: errors.New("operator is not specified"),
		},
		{
			name: "not clusterName",
			req: &apistructs.RuntimeCreateRequest{
				Name:        "test",
				ReleaseID:   "1",
				Operator:    "1",
				ClusterName: "",
				Source:      apistructs.PIPELINE,
				Extra: apistructs.RuntimeCreateRequestExtra{
					OrgID:         1,
					Workspace:     "master",
					ProjectID:     1,
					ApplicationID: 1,
				},
			},
			want: errors.New("clusterName is not specified"),
		},
		{
			name: "not source",
			req: &apistructs.RuntimeCreateRequest{
				Name:        "test",
				ReleaseID:   "1",
				Operator:    "1",
				ClusterName: "jicheng-newb",
				Source:      "default",
				Extra: apistructs.RuntimeCreateRequestExtra{
					OrgID:         1,
					Workspace:     "master",
					ProjectID:     1,
					ApplicationID: 1,
				},
			},
			want: errors.New("source is unknown"),
		},
		{
			name: "PIPELINE",
			req: &apistructs.RuntimeCreateRequest{
				Name:        "test",
				ReleaseID:   "1",
				Operator:    "1",
				ClusterName: "jicheng-newb",
				Source:      apistructs.PIPELINE,
				Extra: apistructs.RuntimeCreateRequestExtra{
					OrgID:         1,
					Workspace:     "master",
					ProjectID:     1,
					ApplicationID: 1,
				},
			},
			want: nil,
		},
		{
			name: "RUNTIMEADDON",
			req: &apistructs.RuntimeCreateRequest{
				Name:        "test",
				ReleaseID:   "1",
				Operator:    "1",
				ClusterName: "jicheng-newb",
				Source:      apistructs.RUNTIMEADDON,
				Extra: apistructs.RuntimeCreateRequestExtra{
					OrgID:         1,
					Workspace:     "master",
					ProjectID:     1,
					ApplicationID: 1,
					InstanceID:    "1",
				},
			},
			want: nil,
		},
		{
			name: "ABILITY",
			req: &apistructs.RuntimeCreateRequest{
				Name:        "test",
				ReleaseID:   "1",
				Operator:    "1",
				ClusterName: "jicheng-newb",
				Source:      apistructs.ABILITY,
				Extra: apistructs.RuntimeCreateRequestExtra{
					OrgID:           1,
					Workspace:       "master",
					ProjectID:       1,
					ApplicationID:   1,
					ApplicationName: "jicheng-newb",
				},
			},
			want: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkRuntimeCreateReq(tc.req)
			if tc.want != nil {
				assert.EqualError(tc.want, err.Error())
			} else {
				assert.Nil(err)
			}
		})
	}

}

func TestService_Create(t *testing.T) {
	assert := require.New(t)
	req := &apistructs.RuntimeCreateRequest{
		Name:        "test",
		ReleaseID:   "1",
		Operator:    "1",
		ClusterName: "jicheng-newb",
		Source:      apistructs.PIPELINE,
		Extra: apistructs.RuntimeCreateRequestExtra{
			OrgID:         1,
			Workspace:     "master",
			ProjectID:     1,
			ApplicationID: 1,
		},
	}

	mdb, mock, err := dbclient.InitMysqlMock()
	mdb = mdb.Debug()
	dbSvc := dbServiceImpl{db: &dbclient.DBClient{
		DBEngine: &dbengine.DBEngine{
			DB: mdb,
		},
	}}

	eventSvc := &events.EventManager{}
	clusterSvc := &cluster.ClusterService{}
	service := NewRuntimeService(WithBundleService(bundle.New()), WithClusterSvc(clusterSvc), WithDBService(&dbSvc), WithEventManagerService(eventSvc))

	m1 := gomonkey.ApplyMethod(reflect.TypeOf(service.bundle), "GetApp", func(bundle *bundle.Bundle, id uint64) (*apistructs.ApplicationDTO, error) {
		return &apistructs.ApplicationDTO{
			ProjectID: 1,
		}, nil
	})
	defer m1.Reset()

	m2 := gomonkey.ApplyMethod(reflect.TypeOf(service.bundle), "GetProjectBranchRules", func(bundle *bundle.Bundle, projectId uint64) ([]*apistructs.BranchRule, error) {
		return []*apistructs.BranchRule{
			{},
		}, nil
	})
	defer m2.Reset()

	m3 := gomonkey.ApplyFunc(diceworkspace.GetValidBranchByGitReference, func(ref string, branchRules []*apistructs.BranchRule) *apistructs.ValidBranch {
		return &apistructs.ValidBranch{IsProtect: true}
	})
	defer m3.Reset()

	m4 := gomonkey.ApplyMethod(reflect.TypeOf(service.bundle), "CheckPermission", func(bundle *bundle.Bundle, req *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
		return &apistructs.PermissionCheckResponseData{Access: true}, nil
	})
	defer m4.Reset()

	m5 := gomonkey.ApplyMethod(reflect.TypeOf(service.clusterSvc), "GetCluster", func(clusterSvc *cluster.ClusterService, ctx context.Context,
		req *clusterpb.GetClusterRequest) (*clusterpb.GetClusterResponse, error) {
		return &clusterpb.GetClusterResponse{
			Data: &clusterpb.ClusterInfo{
				Id: 1,
			},
		}, nil
	})
	defer m5.Reset()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta("")).WithArgs().WillReturnRows(rows)

	m6 := gomonkey.ApplyMethod(reflect.TypeOf(&dbServiceImpl{}), "FindRuntimeOrCreate", func(db *dbServiceImpl, uniqueId spec.RuntimeUniqueId, operator string, source apistructs.RuntimeSource,
		clusterName string, clusterId uint64, gitRepoAbbrev string, projectID, orgID uint64, deploymentOrderId,
		releaseVersion, extraParams string) (*dbclient.Runtime, bool, error) {
		return &dbclient.Runtime{
			Name: "test",
		}, true, nil
	})
	defer m6.Reset()

	m7 := gomonkey.ApplyMethod(reflect.TypeOf(&dbServiceImpl{}), "FindLastDeployment", func(db *dbServiceImpl, id uint64) (*dbclient.Deployment, error) {
		return &dbclient.Deployment{
			Status: apistructs.DeploymentStatusOK,
		}, nil
	})
	defer m7.Reset()

	m8 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&Service{}), "doDeployRuntime", func(s *Service, ctx *DeployContext) (*apistructs.DeploymentCreateResponseDTO, error) {
		return nil, nil
	})
	defer m8.Reset()

	_, err = service.Create(user.ID("1"), req)
	assert.Nil(err)
}

func TestService_doDeployRuntime(t *testing.T) {
	assert := require.New(t)

	deployCtx := &DeployContext{
		Runtime: &dbclient.Runtime{
			ApplicationID: 1,
			Workspace:     "DEV",
			Name:          "test",
		},
		App: &apistructs.ApplicationDTO{
			ProjectID: 1,
		},
	}

	mdb, _, err := dbclient.InitMysqlMock()
	mdb = mdb.Debug()
	dbSvc := dbServiceImpl{db: &dbclient.DBClient{
		DBEngine: &dbengine.DBEngine{
			DB: mdb,
		},
	}}

	eventSvc := &events.EventManager{}

	m1 := gomonkey.ApplyMethod(reflect.TypeOf(&bundle.Bundle{}), "GetDiceYAML", func(bundle *bundle.Bundle, releaseID string, workspace ...string) (*diceyml.DiceYaml, error) {
		return &diceyml.DiceYaml{}, nil
	})
	defer m1.Reset()

	m2 := gomonkey.ApplyMethod(reflect.TypeOf(&Service{}), "PreCheck", func(service *Service, dice *diceyml.DiceYaml, workspace string) error {
		return nil
	})
	defer m2.Reset()

	m3 := gomonkey.ApplyMethod(reflect.TypeOf(&dbServiceImpl{}), "FindPreDeploymentOrCreate", func(db *dbServiceImpl, uniqueId spec.RuntimeUniqueId, dice *diceyml.DiceYaml) (*dbclient.PreDeployment, error) {
		return &dbclient.PreDeployment{}, nil
	})
	defer m3.Reset()

	m4 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&Service{}), "syncRuntimeServices", func(service *Service, runtimeID uint64, dice *diceyml.DiceYaml) error {
		return nil
	})
	defer m4.Reset()

	m5 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&Service{}), "checkOrgDeployBlocked", func(service *Service, orgID uint64, runtime *dbclient.Runtime) (bool, error) {
		return false, nil
	})
	defer m5.Reset()

	m6 := gomonkey.ApplyMethod(reflect.TypeOf(&bundle.Bundle{}), "GetProjectBranchRules", func(bundle *bundle.Bundle, projectId uint64) ([]*apistructs.BranchRule, error) {
		return []*apistructs.BranchRule{}, nil
	})
	defer m6.Reset()

	m7 := gomonkey.ApplyFunc(diceworkspace.GetValidBranchByGitReference, func(ref string, branchRules []*apistructs.BranchRule) *apistructs.ValidBranch {
		return &apistructs.ValidBranch{NeedApproval: true}
	})
	defer m7.Reset()

	m8 := gomonkey.ApplyMethod(reflect.TypeOf(&bundle.Bundle{}), "ListMembers", func(bundle *bundle.Bundle, req apistructs.MemberListRequest) ([]apistructs.Member, error) {
		return []apistructs.Member{}, nil
	})
	defer m8.Reset()

	m9 := gomonkey.ApplyMethod(reflect.TypeOf(&dbServiceImpl{}), "CreateDeployment", func(db *dbServiceImpl, deployment *dbclient.Deployment) error {
		return nil
	})
	defer m9.Reset()

	m10 := gomonkey.ApplyMethod(reflect.TypeOf(&bundle.Bundle{}), "ListUsers", func(bundle *bundle.Bundle, req apistructs.UserListRequest) (*apistructs.UserListResponseData, error) {
		return &apistructs.UserListResponseData{}, nil
	})
	defer m10.Reset()

	m11 := gomonkey.ApplyMethod(reflect.TypeOf(&events.EventManager{}), "EmitEvent", func(eventManager *events.EventManager, e *events.RuntimeEvent) {

	})
	defer m11.Reset()

	service := NewRuntimeService(WithBundleService(bundle.New()), WithDBService(&dbSvc), WithEventManagerService(eventSvc))
	_, err = service.doDeployRuntime(deployCtx)
	assert.Nil(err)
}
