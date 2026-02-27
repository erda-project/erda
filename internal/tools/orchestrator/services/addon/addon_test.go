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
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"bou.ke/monkey"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/cluster-manager/cluster"
	"github.com/erda-project/erda/internal/tools/orchestrator/components/runtime/mock"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/i18n"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/impl/servicegroup"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
	"github.com/erda-project/erda/pkg/mysqlhelper"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

const count = 20

func TestConcurrentReadWriteAppInfos(t *testing.T) {
	var keys = []string{"1", "2", "3", "4", "5"}

	var db *dbclient.DBClient
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetAttachmentsByInstanceID",
		func(*dbclient.DBClient, string) (*[]dbclient.AddonAttachment, error) {
			var addons []dbclient.AddonAttachment
			for _, v := range keys {
				addons = append(addons, dbclient.AddonAttachment{
					ProjectID:     v,
					ApplicationID: v,
				})
			}
			return &addons, nil
		},
	)
	defer monkey.UnpatchAll()

	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetAppsByProject",
		func(_ *bundle.Bundle, id uint64, _ uint64, _ string) (*apistructs.ApplicationListResponseData, error) {
			return &apistructs.ApplicationListResponseData{
				List: []apistructs.ApplicationDTO{
					{
						ID: id,
					},
				},
			}, nil
		},
	)

	var (
		wg         sync.WaitGroup
		orgID      uint64 = 1
		userID            = "1"
		instanceID        = "1"
	)
	wg.Add(count)
	for i := 0; i != count; i++ {
		go func() {
			a := Addon{}
			_, err := a.ListReferencesByInstanceID(orgID, userID, instanceID)
			if err != nil {
				t.Fatal(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	for _, v := range keys {
		_, ok := AppInfos.Load(v)
		assert.Equal(t, true, ok)
	}
}

func TestDeleteAddonUsed(t *testing.T) {
	var db *dbclient.DBClient
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetInstanceRouting",
		func(*dbclient.DBClient, string) (*dbclient.AddonInstanceRouting, error) {
			return &dbclient.AddonInstanceRouting{}, nil
		},
	)

	addon := Addon{}
	monkey.PatchInstanceMethod(reflect.TypeOf(&addon), "DeleteTenant",
		func(*Addon, string, string) error {
			return nil
		},
	)

	monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetAttachmentCountByRoutingInstanceID",
		func(*dbclient.DBClient, string) (int64, error) {
			return 1, nil
		},
	)
	defer monkey.UnpatchAll()

	err := addon.Delete("", "")
	if err.Error() != "addon is being referenced, can't delete" {
		t.Fatal("the err is not equal with expected")
	}
}

func Test_GetAddonConfig(t *testing.T) {
	cfgStr := `{"ADDON_HAS_ENCRIPY":"YES","MYSQL_DATABASE":"fake","MYSQL_HOST":"fake","MYSQL_PASSWORD":"fake1","MYSQL_PORT":"fake","MYSQL_USERNAME":"fake"}`
	cfg, err := GetAddonConfig(cfgStr)
	assert.NoError(t, err)
	assert.Equal(t, "fake1", cfg["MYSQL_PASSWORD"].(string))
}

func TestSetAddonVolumes(t *testing.T) {
	type args struct {
		service    *diceyml.Service
		options    map[string]string
		hostPath   string
		targetPath string
		readOnly   bool
	}

	service := &diceyml.Service{}
	options := map[string]string{
		"app_kind":             "deployment",
		"alibabacloud.com/eci": "true",
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "case_01",
			args: args{
				service:    service,
				options:    options,
				hostPath:   "/opt/data",
				targetPath: "/opt/data",
				readOnly:   false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func Test_addonCanScale(t *testing.T) {
	type args struct {
		addonName    string
		addonId      string
		action       string
		status       string
		addonPlan    string
		addonVersion string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Test_01",
			args: args{
				addonName:    "mysql",
				addonId:      "z44f5f6543f004d54ac2a2538efd4e9ec",
				action:       apistructs.ScaleActionDown,
				status:       string(apistructs.AddonAttached),
				addonPlan:    "basic",
				addonVersion: "5.7.29",
			},
			wantErr: false,
		},
		{
			name: "Test_02",
			args: args{
				addonName:    "mysql",
				addonId:      "z44f5f6543f004d54ac2a2538efd4e9ec",
				action:       apistructs.ScaleActionUp,
				status:       string(apistructs.AddonAttached),
				addonPlan:    "basic",
				addonVersion: "5.7.29",
			},
			wantErr: true,
		},
		{
			name: "Test_03",
			args: args{
				addonName:    "mysql",
				addonId:      "z44f5f6543f004d54ac2a2538efd4e9ec",
				action:       apistructs.ScaleActionUp,
				status:       string(apistructs.AddonAttached),
				addonPlan:    "basic",
				addonVersion: "5.7.29",
			},
			wantErr: true,
		},
		{
			name: "Test_04",
			args: args{
				addonName:    "mysql",
				addonId:      "z44f5f6543f004d54ac2a2538efd4e9ec",
				action:       apistructs.ScaleActionUp,
				status:       string(apistructs.AddonOffline),
				addonPlan:    "basic",
				addonVersion: "5.7.29",
			},
			wantErr: false,
		},
		{
			name: "Test_05",
			args: args{
				addonName:    "mysql",
				addonId:      "z44f5f6543f004d54ac2a2538efd4e9ec",
				action:       apistructs.ScaleActionUp,
				status:       string(apistructs.AddonOffline),
				addonPlan:    "professional",
				addonVersion: "5.7.29",
			},
			wantErr: false,
		},
		{
			name: "Test_06",
			args: args{
				addonName:    "redis",
				addonId:      "z44f5f6543f004d54ac2a2538efd4e9ec",
				action:       apistructs.ScaleActionUp,
				status:       string(apistructs.AddonOffline),
				addonPlan:    "professional",
				addonVersion: "5.7.29",
			},
			wantErr: false,
		},
		{
			name: "Test_07",
			args: args{
				addonName:    "terminus-elasticsearch",
				addonId:      "z44f5f6543f004d54ac2a2538efd4e9ec",
				action:       apistructs.ScaleActionUp,
				status:       string(apistructs.AddonOffline),
				addonPlan:    "professional",
				addonVersion: "5.7.29",
			},
			wantErr: false,
		},
		{
			name: "Test_08",
			args: args{
				addonName:    "terminus-elasticsearch",
				addonId:      "z44f5f6543f004d54ac2a2538efd4e9ec",
				action:       apistructs.ScaleActionUp,
				status:       string(apistructs.AddonOffline),
				addonPlan:    "professional",
				addonVersion: "6.8.9",
			},
			wantErr: true,
		},
		{
			name: "Test_09",
			args: args{
				addonName:    "mysql",
				addonId:      "z44f5f6543f004d54ac2a2538efd4e9ec",
				action:       apistructs.ScaleActionDown,
				status:       string(apistructs.AddonOffline),
				addonPlan:    "professional",
				addonVersion: "5.7.29",
			},
			wantErr: true,
		},
		{
			name: "Test_10",
			args: args{
				addonName:    "rocketmq",
				addonId:      "z44f5f6543f004d54ac2a2538efd4e9ec",
				action:       apistructs.ScaleActionDown,
				status:       string(apistructs.AddonOffline),
				addonPlan:    "basic",
				addonVersion: "4.2.0",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := addonCanScale(tt.args.addonName, tt.args.addonId, tt.args.addonPlan, tt.args.addonVersion, tt.args.status, tt.args.action); (err != nil) != tt.wantErr {
				t.Errorf("addonCanScale() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddon_doAddonScale(t *testing.T) {
	var db *dbclient.DBClient
	var bdl *bundle.Bundle
	var serviceGroupImpl *servicegroup.ServiceGroupImpl

	a := New(WithDBClient(db), WithBundle(bdl), WithServiceGroup(serviceGroupImpl))
	addonInstance := &dbclient.AddonInstance{
		ID:                  "z44f5f6543f004d54ac2a2538efd4e9ec",
		Name:                "mysql",
		AddonName:           "mysql",
		Category:            "database",
		Namespace:           "addon-mysql",
		ScheduleName:        "z44f5f6543f004d54ac2a2538efd4e9ec",
		Plan:                "professional",
		Version:             "5.7.29",
		Options:             "{\"applicationId\":\"21\",\"clusterName\":\"test\",\"orgId\":\"1\",\"projectId\":\"1\",\"runtimeId\":\"192\",\"runtimeName\":\"feature/develop\",\"version\":\"5.7.29\",\"workspace\":\"DEV\"}",
		Status:              "ATTACHED",
		ShareScope:          "PROJECT",
		OrgID:               "1",
		Cluster:             "test",
		ProjectID:           "1",
		ApplicationID:       "21",
		AppID:               "1",
		Workspace:           "DEV",
		Deleted:             "N",
		PlatformServiceType: 0,
		KmsKey:              "f2dcc7b3761d4244898303cb0104a584",
		CpuRequest:          0.4,
		CpuLimit:            4,
		MemRequest:          17204,
		MemLimit:            17204,
	}

	addonInstanceRoutings := make([]dbclient.AddonInstanceRouting, 0)

	monkey.PatchInstanceMethod(reflect.TypeOf(a), "GetAddonExtention",
		func(a *Addon, params *apistructs.AddonHandlerCreateItem) (*apistructs.AddonExtension, *diceyml.Object, error) {
			return &apistructs.AddonExtension{}, &diceyml.Object{}, nil
		},
	)

	monkey.PatchInstanceMethod(reflect.TypeOf(a), "BuildAddonScaleRequestGroup",
		func(a *Addon, params *apistructs.AddonHandlerCreateItem, addonIns *dbclient.AddonInstance, scaleAction string, addonSpec *apistructs.AddonExtension, addonDice *diceyml.Object) (*apistructs.ServiceGroup, error) {
			services := make(map[string]*diceyml.Service)
			services["mysql-1"] = &diceyml.Service{
				Image: "registry.erda.cloud/erda-addons-enterprise/addon-mysql:5.7.29-1.0.1-init",
				Ports: make([]diceyml.ServicePort, 0),
				Envs:  make(map[string]string),
				Resources: diceyml.Resources{
					CPU: 1,
					Mem: 4301,
				},
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
				Binds:       make([]string, 0),
				Deployments: diceyml.Deployments{
					Replicas: 1,
				},
			}

			services["mysql-1"].Ports = append(services["mysql-1"].Ports, diceyml.ServicePort{
				Port:       3306,
				Protocol:   "TCP",
				L4Protocol: "TCP",
				Expose:     false,
				Default:    false,
			})

			services["mysql-1"].Envs = diceyml.EnvMap{
				"ADDON_GROUPS":        "2",
				"ADDON_ID":            "z44f5f6543f004d54ac2a2538efd4e9ec",
				"ADDON_NODE_ID":       "f54fc4ff4197e4c4fa1cdc5b929ca5849",
				"ADDON_TYPE":          "mysql",
				"DICE_ADDON":          "z44f5f6543f004d54ac2a2538efd4e9ec",
				"DICE_ADDON_TYPE":     "mysql",
				"DICE_CLUSTER_NAME":   "test",
				"MYSQL_ROOT_PASSWORD": "cR7yf6zEBVFQ8WgE",
				"SERVER_ID":           "1",
				"SERVICE_TYPE":        "ADDONS",
			}

			services["mysql-1"].Labels = map[string]string{
				"ADDON_GROUP_ID": "mysql-master",
			}

			services["mysql-1"].Binds = []string{
				"/netdata/addon/mysql/backup/z44f5f6543f004d54ac2a2538efd4e9ec_1:/var/backup/mysql:rw",
				"z44f5f6543f004d54ac2a2538efd4e9ec_1:/var/lib/mysql:rw",
			}

			services["mysql-2"] = &diceyml.Service{
				Image: "registry.erda.cloud/erda-addons-enterprise/addon-mysql:5.7.29-1.0.1-init",
				Ports: make([]diceyml.ServicePort, 0),
				Envs:  make(map[string]string),
				Resources: diceyml.Resources{
					CPU: 1,
					Mem: 4301,
				},
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
				Binds:       make([]string, 0),
				Deployments: diceyml.Deployments{
					Replicas: 1,
				},
			}

			services["mysql-2"].Ports = append(services["mysql-2"].Ports, diceyml.ServicePort{
				Port:       3306,
				Protocol:   "TCP",
				L4Protocol: "TCP",
				Expose:     false,
				Default:    false,
			})

			services["mysql-2"].Envs = diceyml.EnvMap{
				"ADDON_GROUPS":        "2",
				"ADDON_ID":            "z44f5f6543f004d54ac2a2538efd4e9ec",
				"ADDON_NODE_ID":       "m6475b57e54884af59e4147382964f7ab",
				"ADDON_TYPE":          "mysql",
				"DICE_ADDON":          "z44f5f6543f004d54ac2a2538efd4e9ec",
				"DICE_ADDON_TYPE":     "mysql",
				"DICE_CLUSTER_NAME":   "test",
				"MYSQL_ROOT_PASSWORD": "cR7yf6zEBVFQ8WgE",
				"SERVER_ID":           "2",
				"SERVICE_TYPE":        "ADDONS",
			}

			services["mysql-2"].Labels = map[string]string{
				"ADDON_GROUP_ID": "mysql-slave",
			}

			services["mysql-2"].Binds = []string{
				"/netdata/addon/mysql/backup/z44f5f6543f004d54ac2a2538efd4e9ec_2:/var/backup/mysql:rw",
				"z44f5f6543f004d54ac2a2538efd4e9ec_2:/var/lib/mysql:rw",
			}

			req := &apistructs.ServiceGroupCreateV2Request{
				DiceYml: diceyml.Object{
					Version:  "2.0",
					Services: services,
				},
				ClusterName: "test",
				ID:          addonIns.ID,
				Type:        strings.Join([]string{"addon-", strings.Replace(strings.Replace(addonIns.AddonName, "terminus-", "", 1), "-operator", "", 1)}, ""),
				GroupLabels: make(map[string]string),
				Volumes:     make(map[string]apistructs.RequestVolumeInfo),
			}

			ret := &apistructs.UpdateServiceGroupScaleRequest{
				Namespace:   addonIns.Namespace,
				Name:        addonIns.ScheduleName,
				ClusterName: params.ClusterName,
				Labels:      make(map[string]string),
				Services:    make([]apistructs.Service, 0),
			}
			ret.Labels = req.GroupLabels

			for svcName, svc := range req.DiceYml.Services {
				scale := 0
				if scaleAction == apistructs.ScaleActionDown {
					scale = 0
				}
				if scaleAction == apistructs.ScaleActionUp {
					// TODO: 从数据库表获取前一次 scaleUp 成功之后的 replicas
					scale = svc.Deployments.Replicas
				}

				ret.Services = append(ret.Services, apistructs.Service{
					Name:  svcName,
					Scale: scale,
					Resources: apistructs.Resources{
						Cpu:  svc.Resources.CPU,
						Mem:  float64(svc.Resources.Mem),
						Disk: float64(svc.Resources.Disk),
					},
				})
			}
			return &apistructs.ServiceGroup{}, nil
		},
	)

	monkey.PatchInstanceMethod(reflect.TypeOf(serviceGroupImpl), "Scale",
		func(_ *servicegroup.ServiceGroupImpl, sg *apistructs.ServiceGroup) (interface{}, error) {
			return apistructs.ServiceGroup{}, nil
		},
	)

	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "UpdateAddonInstanceStatus",
		func(_ *dbclient.DBClient, ID, status string) error {
			return nil
		},
	)

	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "UpdateInstanceRouting",
		func(db *dbclient.DBClient, routing *dbclient.AddonInstanceRouting) error {
			return nil
		},
	)

	err := a.doAddonScale(addonInstance, &addonInstanceRoutings, apistructs.ScaleActionDown)
	assert.Equal(t, err, nil)
}

func TestAddon_insideAddonCanNotScale(t *testing.T) {
	type fields struct {
		db *dbclient.DBClient
	}
	type args struct {
		routingIns *dbclient.AddonInstanceRouting
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test_01",
			fields: fields{
				db: &dbclient.DBClient{},
			},
			args: args{
				routingIns: &dbclient.AddonInstanceRouting{
					AddonName:    "terminus-zookeeper",
					RealInstance: "y6f6485f7d9974c32b47c3a1ecd244109",
				},
			},
			wantErr: true,
		},
		{
			name: "test_02",
			fields: fields{
				db: &dbclient.DBClient{},
			},
			args: args{
				routingIns: &dbclient.AddonInstanceRouting{
					AddonName:    "mysql",
					RealInstance: "y6f6485f7d9974c32b47c3a1ecd244109",
				},
			},
			wantErr: true,
		},
		{
			name: "test_03",
			fields: fields{
				db: &dbclient.DBClient{},
			},
			args: args{
				routingIns: &dbclient.AddonInstanceRouting{
					AddonName:    "mysql",
					RealInstance: "y6f6485f7d9974c32b47c3a1ecd244109",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Addon{
				db: tt.fields.db,
			}

			monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "GetByInSideInstanceID",
				func(db *dbclient.DBClient, instanceID string) (*dbclient.AddonInstanceRelation, error) {
					if tt.name == "test_02" {
						return nil, nil
					}
					return &dbclient.AddonInstanceRelation{
						OutsideInstanceID: "c32a40074138a4910af97cff325f8bcd5",
						InsideInstanceID:  "y6f6485f7d9974c32b47c3a1ecd244109",
					}, nil
				},
			)

			monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "GetInstanceRoutingByRealInstance",
				func(db *dbclient.DBClient, realIns string) (*[]dbclient.AddonInstanceRouting, error) {
					if tt.name == "test_03" {
						return nil, nil
					}
					ars := make([]dbclient.AddonInstanceRouting, 0)
					ars = append(ars, dbclient.AddonInstanceRouting{
						Name: "kafka",
						ID:   "g3dd89da1f63245a3b2b174d9610661bd",
					})
					return &ars, nil
				},
			)

			if err := a.insideAddonCanNotScale(tt.args.routingIns); (err != nil) != tt.wantErr {
				t.Errorf("insideAddonCanNotScale() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddon(t *testing.T) {
	{
		WithResource(nil)
		WithKMSWrapper(nil)
		WithCap(nil)
		WithServiceGroup(nil)
		WithInstanceinfoImpl(nil)
		WithClusterInfoImpl(nil)
		WithClusterSvc(nil)
		WithTenantSvc(nil)
		md5V("")
		a := MicroServiceProjectData{&apistructs.MicroServiceProjectResponseData{ProjectID: "1"}, &apistructs.MicroServiceProjectResponseData{ProjectID: "2"}}
		a.Len()
		a.Swap(0, 1)
		a.Less(0, 1)
		buildMiddlewareFilter(apistructs.InstanceInfoDataList{
			{AddonID: "1"},
		})
	}

	defer monkey.UnpatchAll()

	var db *dbclient.DBClient
	var bdl *bundle.Bundle
	// var serviceGroupImpl *servicegroup.ServiceGroupImpl

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// bdlSvc := mock.NewMockBundleService(ctrl)
	// dbSvc := mock.NewMockDBService(ctrl)
	sgiSvc := mock.NewMockServiceGroup(ctrl)

	a := New(
		WithDBClient(db),
		WithBundle(bdl),
		WithServiceGroup(sgiSvc),
	)
	addonInstance := &dbclient.AddonInstance{
		ID:                  "z44f5f6543f004d54ac2a2538efd4e9ec",
		Name:                "mysql",
		AddonName:           "mysql",
		Category:            "database",
		Namespace:           "addon-mysql",
		ScheduleName:        "z44f5f6543f004d54ac2a2538efd4e9ec",
		Plan:                "professional",
		Version:             "5.7.29",
		Options:             "{\"applicationId\":\"21\",\"clusterName\":\"test\",\"orgId\":\"1\",\"projectId\":\"1\",\"runtimeId\":\"192\",\"runtimeName\":\"feature/develop\",\"version\":\"5.7.29\",\"workspace\":\"DEV\"}",
		Status:              "ATTACHED",
		ShareScope:          "PROJECT",
		OrgID:               "1",
		Cluster:             "test",
		ProjectID:           "1",
		ApplicationID:       "21",
		AppID:               "1",
		Workspace:           "DEV",
		Deleted:             "N",
		PlatformServiceType: 0,
		KmsKey:              "f2dcc7b3761d4244898303cb0104a584",
		CpuRequest:          0.4,
		CpuLimit:            4,
		MemRequest:          17204,
		MemLimit:            17204,
	}

	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "UpdateAddonInstanceResource",
		func(db *dbclient.DBClient, ID string, cpurequest, cpulimit float64, memrequest, memlimit int) error {
			return nil
		},
	)
	a.updateAddonInstanceResource(*addonInstance, apistructs.PodInfoDataList{
		apistructs.PodInfoData{},
	})

	// monkey.PatchInstanceMethod(reflect.TypeOf(a.bdl), "ListDopOrgs",
	// 	func(bdl *bundle.Bundle, req *apistructs.OrgSearchRequest) (*apistructs.PagingOrgDTO, error) {
	// 		return &apistructs.PagingOrgDTO{
	// 			List: []apistructs.OrgDTO{
	// 				{ID: 123},
	// 			},
	// 		}, nil
	// 	},
	// )
	// a.getAllOrgIDs()

	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "ListAddonInstanceByOrg",
		func(db *dbclient.DBClient, orgID uint64) (*[]dbclient.AddonInstance, error) {
			return &[]dbclient.AddonInstance{}, nil
		},
	)
	a.ListAddonInstanceByOrg(0)

	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "ListAttachmentIDRuntimeIDNotExist",
		func(db *dbclient.DBClient) ([]dbclient.AddonAttachment, error) {
			return nil, nil
		},
	)
	a.CleanRemainingAddonAttachment()
	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "ListAttachmentIDRuntimeIDNotExist",
		func(db *dbclient.DBClient) ([]dbclient.AddonAttachment, error) {
			return []dbclient.AddonAttachment{
				{ID: 1},
			}, nil
		},
	)
	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "DeleteAttachmentByIDs",
		func(db *dbclient.DBClient, id ...uint64) error {
			return nil
		},
	)
	a.CleanRemainingAddonAttachment()

	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "ListAttachmentIDRuntimeIDNotExist",
		func(db *dbclient.DBClient) ([]dbclient.AddonAttachment, error) {
			return nil, errors.New("123")
		},
	)
	a.CleanRemainingAddonAttachment()
	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "ListAttachmentIDRuntimeIDNotExist",
		func(db *dbclient.DBClient) ([]dbclient.AddonAttachment, error) {
			return []dbclient.AddonAttachment{
				{ID: 1},
			}, nil
		},
	)
	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "DeleteAttachmentByIDs",
		func(db *dbclient.DBClient, id ...uint64) error {
			return errors.New("123")
		},
	)
	a.CleanRemainingAddonAttachment()

	sg := apistructs.ServiceGroup{
		ClusterName:   "test",
		Force:         true,
		ScheduleInfo:  apistructs.ScheduleInfo{},
		ScheduleInfo2: apistructs.ScheduleInfo2{},
		Dice: apistructs.Dice{
			ID:     "z44f5f6543f004d54ac2a2538efd4e9ec",
			Type:   "addon-mysql",
			Labels: map[string]string{},
			Services: []apistructs.Service{
				{
					Name:          "mysql-1",
					Namespace:     "",
					Image:         "registry.erda.cloud/erda-addons-enterprise/addon-mysql:5.7.29-1.0.1-init",
					ImageUsername: "",
					ImagePassword: "",
					Cmd:           "",
					Ports:         []diceyml.ServicePort{{Port: 3306, Protocol: "TCP", L4Protocol: "TCP", Expose: false, Default: false}},
					Scale:         1,
					Resources:     apistructs.Resources{Cpu: 1, Mem: 4301, MaxCPU: 0, MaxMem: 0, Disk: 0},
					Env: map[string]string{
						"ADDON_GROUPS":        "2",
						"ADDON_ID":            "z44f5f6543f004d54ac2a2538efd4e9ec",
						"ADDON_NODE_ID":       "f54fc4ff4197e4c4fa1cdc5b929ca5849",
						"ADDON_TYPE":          "mysql",
						"DICE_ADDON":          "z44f5f6543f004d54ac2a2538efd4e9ec",
						"DICE_ADDON_TYPE":     "mysql",
						"DICE_CLUSTER_NAME":   "test",
						"MYSQL_ROOT_PASSWORD": "cR7yf6zEBVFQ8WgE",
						"SERVER_ID":           "1", "SERVICE_TYPE": "ADDONS",
					},
					Labels: map[string]string{"ADDON_GROUP_ID": "mysql-master"},
					Binds: []apistructs.ServiceBind{{
						Bind: apistructs.Bind{
							ContainerPath: "/var/backup/mysql",
							HostPath:      "/netdata/addon/mysql/backup/z44f5f6543f004d54ac2a2538efd4e9ec_1",
							ReadOnly:      false,
						},
					}, {
						Bind: apistructs.Bind{
							ContainerPath: "/var/lib/mysql",
							HostPath:      "z44f5f6543f004d54ac2a2538efd4e9ec_1",
							ReadOnly:      false,
						},
					},
					},
					Volumes: []apistructs.Volume{{
						ID:            "101",
						VolumePath:    "",
						VolumeType:    "local",
						Size:          10,
						ContainerPath: "/opt/test",
						Storage:       "",
					},
					},
					InstanceInfos: []apistructs.InstanceInfo{
						{Ip: "1.2.3.4"},
					},
				},
			},
		},
	}

	a.initMsAfterStart(&sg, "mysql-1", "", &apistructs.ClusterInfoData{
		apistructs.DICE_CLUSTER_NAME: "aaa",
	})
	a.initMsAfterStart(&sg, "mysql-2", "", &apistructs.ClusterInfoData{
		apistructs.DICE_CLUSTER_NAME: "aaa",
	})

	a.createDBs(&sg, nil, addonInstance, "mysql-1", "", &apistructs.ClusterInfoData{
		apistructs.DICE_CLUSTER_NAME: "aaa",
	})
	a.createDBs(&sg, &apistructs.ExistsMysqlExec{
		MysqlHost: "1222",
		Options: map[string]string{
			"create_dbs": "db1,db2",
			"init_sql":   "111",
		},
	}, addonInstance, "mysql-2", "", &apistructs.ClusterInfoData{
		apistructs.DICE_CLUSTER_NAME: "aaa",
	})

	a.checkMysqlHa(&sg, "mysql-1", "", &apistructs.ClusterInfoData{
		apistructs.DICE_CLUSTER_NAME: "aaa",
	})
	a.checkMysqlHa(&sg, "mysql-2", "", &apistructs.ClusterInfoData{
		apistructs.DICE_CLUSTER_NAME: "aaa",
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(a.bdl), "MySQLExecFile",
		func(bdl *bundle.Bundle, mysqlExec *apistructs.MysqlExec, soldierUrl string) error {
			return nil
		},
	)
	a.initSqlFile(&sg, nil, addonInstance, []string{"db1", "db2"}, "mysql-1", "", &apistructs.ClusterInfoData{
		apistructs.DICE_CLUSTER_NAME: "aaa",
	})
	a.initSqlFile(&sg, &apistructs.ExistsMysqlExec{
		MysqlHost: "1222",
		Options: map[string]string{
			"create_dbs": "db1,db2",
			"init_sql":   "111",
		},
	}, addonInstance, []string{"db1", "db2"}, "mysql-2", "", &apistructs.ClusterInfoData{
		apistructs.DICE_CLUSTER_NAME: "aaa",
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(a), "CreateAddonProvider",
		func(a *Addon, req *apistructs.AddonProviderRequest, addonName, providerDomain, userId string) (int, *apistructs.AddonProviderResponse, error) {
			return 0, nil, errors.New("error")
		},
	)
	monkey.PatchInstanceMethod(reflect.TypeOf(a), "DeleteAddonProvider",
		func(a *Addon, req *apistructs.AddonProviderRequest, uuid, addonName, providerDomain string) (*apistructs.AddonProviderDeleteResponse, error) {
			return nil, errors.New("error")
		},
	)
	monkey.PatchInstanceMethod(reflect.TypeOf(a), "UpdateAddonStatus",
		func(a *Addon, addonIns *dbclient.AddonInstance, addonStatus apistructs.AddonStatus) error {
			return nil
		},
	)

	a.providerAddonDeploy(addonInstance, nil, &apistructs.AddonHandlerCreateItem{}, &apistructs.AddonExtension{})

	monkey.UnpatchInstanceMethod(reflect.TypeOf(a), "CreateAddonProvider")
	monkey.PatchInstanceMethod(reflect.TypeOf(a), "CreateAddonProvider",
		func(a *Addon, req *apistructs.AddonProviderRequest, addonName, providerDomain, userId string) (int, *apistructs.AddonProviderResponse, error) {
			return 200, &apistructs.AddonProviderResponse{
				Data: apistructs.AddonProviderDataResp{
					Config: map[string]interface{}{"aaa": 123},
					Label:  map[string]string{"aaa": "123"},
				},
			}, nil
		},
	)
	a.providerAddonDeploy(addonInstance, nil, &apistructs.AddonHandlerCreateItem{}, &apistructs.AddonExtension{})

	monkey.UnpatchInstanceMethod(reflect.TypeOf(a), "CreateAddonProvider")
	monkey.UnpatchInstanceMethod(reflect.TypeOf(a), "DeleteAddonProvider")
	monkey.UnpatchInstanceMethod(reflect.TypeOf(a), "UpdateAddonStatus")
}

type cluster_info struct{}

func (cluster_info) Info(string) (apistructs.ClusterInfoData, error) {
	return apistructs.ClusterInfoData{}, nil
}
func (cluster_info) List([]string) (apistructs.ClusterInfoDataList, error) {
	return apistructs.ClusterInfoDataList{}, nil
}

func TestAddon2(t *testing.T) {
	defer monkey.UnpatchAll()

	var db *dbclient.DBClient
	var bdl *bundle.Bundle
	var serviceGroupImpl *servicegroup.ServiceGroupImpl

	a := New(
		WithDBClient(db),
		WithBundle(bdl),
		WithServiceGroup(serviceGroupImpl),
		WithHTTPClient(httpclient.New()),
		WithClusterInfoImpl(new(cluster_info)),
	)

	monkey.PatchInstanceMethod(reflect.TypeOf(a.bdl), "KMSCreateKey",
		func(bdl *bundle.Bundle, req apistructs.KMSCreateKeyRequest) (*kmstypes.CreateKeyResponse, error) {
			return nil, errors.New("error")
		},
	)

	monkey.Patch(i18n.OrgSprintf, func(orgId string, key string, args ...any) string {
		switch key {
		case AddonTypeDoseNoExist:
			return "AddonTypeDoseNoExist"
		case AddonVersionDoseNoExist:
			return "AddonVersionDoseNoExist"
		case AddonDefaultVersionDoseNoExist:
			return "AddonDefaultVersionDoseNoExist"
		case AddonPlanIllegal:
			return "AddonPlanIllegal"
		case AddonPlanNotSupport:
			return "AddonPlanNotSupport"
		}
		return "default"
	})

	a.buildAddonInstance(&apistructs.AddonExtension{}, &apistructs.AddonHandlerCreateItem{})

	a.CreateAddonProvider(&apistructs.AddonProviderRequest{}, "", "", "")
	a.DeleteAddonProvider(&apistructs.AddonProviderRequest{}, "", "", "")

	monkey.PatchInstanceMethod(reflect.TypeOf(a.bdl), "PushLog",
		func(bdl *bundle.Bundle, req *apistructs.LogPushRequest) error {
			return errors.New("error")
		},
	)
	a.pushLog("", &apistructs.AddonHandlerCreateItem{
		Options: map[string]string{"deploymentId": "id", "orgName": "org"},
	})

	ext := []apistructs.ExtensionVersion{}

	monkey.PatchInstanceMethod(reflect.TypeOf(a.bdl), "QueryExtensionVersions",
		func(bdl *bundle.Bundle, req apistructs.ExtensionVersionQueryRequest) ([]apistructs.ExtensionVersion, error) {
			return ext, nil
		},
	)
	a.GetAddonExtention(&apistructs.AddonHandlerCreateItem{})

	ext = append(ext, apistructs.ExtensionVersion{
		IsDefault: true,
		Version:   "1",
	})
	a.GetAddonExtention(&apistructs.AddonHandlerCreateItem{
		Options: map[string]string{"version": ""},
	})
	a.GetAddonExtention(&apistructs.AddonHandlerCreateItem{
		Options: map[string]string{"version": "1"},
	})
	ext = append(ext, apistructs.ExtensionVersion{
		IsDefault: true,
		Version:   "1",
	})
	a.GetAddonExtention(&apistructs.AddonHandlerCreateItem{
		Options: map[string]string{"version": ""},
	})
	a.GetAddonExtention(&apistructs.AddonHandlerCreateItem{
		Options: map[string]string{"version": "1"},
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(a), "GetAddonExtention",
		func(a *Addon, params *apistructs.AddonHandlerCreateItem) (*apistructs.AddonExtension, *diceyml.Object, error) {
			return nil, nil, errors.New("error")
		},
	)
	a.AddonCreate(apistructs.AddonDirectCreateRequest{
		Addons: diceyml.AddOns{
			"mysql": &diceyml.AddOn{
				Plan: "a:b",
			},
		},
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "GetByRuntimeIDAndRoutingInstanceID",
		func(db *dbclient.DBClient, runtimeID, routingInstanceID string) (*[]dbclient.AddonAttachment, error) {
			return nil, errors.New("error")
		},
	)
	a.existAttachAddon(&apistructs.AddonHandlerCreateItem{}, &apistructs.AddonExtension{}, &dbclient.AddonInstanceRouting{})

	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "GetByRuntimeIDAndRoutingInstanceID",
		func(db *dbclient.DBClient, runtimeID, routingInstanceID string) (*[]dbclient.AddonAttachment, error) {
			return &[]dbclient.AddonAttachment{
				{},
			}, nil
		},
	)

	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "GetAddonInstance",
		func(db *dbclient.DBClient, id string) (*dbclient.AddonInstance, error) {
			return &dbclient.AddonInstance{
				ID: "1",
			}, nil
		},
	)
	a.existAttachAddon(&apistructs.AddonHandlerCreateItem{}, &apistructs.AddonExtension{}, &dbclient.AddonInstanceRouting{})

	a.existAttachAddon(&apistructs.AddonHandlerCreateItem{}, &apistructs.AddonExtension{
		Name: string(apistructs.AddonMySQL),
	}, &dbclient.AddonInstanceRouting{
		Status: string(apistructs.AddonAttached),
	})

	a.buildAddonAttachments(&apistructs.AddonHandlerCreateItem{}, &dbclient.AddonInstanceRouting{})
	a.buildAddonInstanceRouting(&apistructs.AddonExtension{}, &apistructs.AddonHandlerCreateItem{}, &dbclient.AddonInstance{}, 1, 2, 3)

	monkey.PatchInstanceMethod(reflect.TypeOf(a.bdl), "KMSCreateKey",
		func(b *bundle.Bundle, req apistructs.KMSCreateKeyRequest) (*kmstypes.CreateKeyResponse, error) {
			return &kmstypes.CreateKeyResponse{}, nil
		},
	)
	monkey.PatchInstanceMethod(reflect.TypeOf(a.bdl), "KMSEncrypt",
		func(b *bundle.Bundle, req apistructs.KMSEncryptRequest) (*kmstypes.EncryptResponse, error) {
			return &kmstypes.EncryptResponse{}, nil
		},
	)
	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "GetAddonInstance",
		func(db *dbclient.DBClient, id string) (*dbclient.AddonInstance, error) {
			return &dbclient.AddonInstance{}, nil
		},
	)
	monkey.PatchInstanceMethod(reflect.TypeOf(a), "GetAddonConfig",
		func(a *Addon, ins *dbclient.AddonInstance) (*apistructs.AddonConfigRes, error) {
			return &apistructs.AddonConfigRes{
				Config: map[string]interface{}{
					"MYSQL_HOST":     "",
					"MYSQL_PASSWORD": "",
					"MYSQL_PORT":     "",
					"MYSQL_USERNAME": "",
				},
			}, nil
		},
	)

	monkey.PatchInstanceMethod(reflect.TypeOf(mysqlhelper.Request{}), "Exec",
		func(r mysqlhelper.Request) error {
			return nil
		},
	)
	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "CreateAddonInstanceTenant",
		func(db *dbclient.DBClient, tenant *dbclient.AddonInstanceTenant) error {
			return nil
		},
	)

	a.CreateMysqlTenant("", &dbclient.AddonInstanceRouting{}, &dbclient.AddonInstance{}, map[string]string{})

	monkey.PatchInstanceMethod(reflect.TypeOf(a.bdl), "CreateErrorLog",
		func(b *bundle.Bundle, errorLog *apistructs.ErrorLogCreateRequest) error {
			return nil
		},
	)
	a.ExportLogInfo(apistructs.InfoLevel, apistructs.AddonError, "", "", "")
	a.ExportLogInfoDetail(apistructs.InfoLevel, apistructs.AddonError, "", "", "")

	a.transPlan(nil)
	a.transPlan([]apistructs.AddonCreateItem{
		{Plan: "large"},
		{Plan: "medium"},
		{Plan: "small"},
		{Plan: ""},
	})

	a.AddonProvisionCallback("", &apistructs.AddonCreateCallBackResponse{})

	monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(a.bdl), "GetProjectWithSetter",
		func(b *bundle.Bundle, id uint64, requestSetter ...httpclient.RequestSetter) (*apistructs.ProjectDTO, error) {
			return &apistructs.ProjectDTO{}, nil
		},
	)
	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "GetRoutingInstanceByProjectAndName",
		func(db *dbclient.DBClient, projectID uint64, workspace, addonName, name, clusterName string) (*dbclient.AddonInstanceRouting, error) {
			return nil, nil
		},
	)
	monkey.PatchInstanceMethod(reflect.TypeOf(a.bdl), "KMSCreateKey",
		func(bdl *bundle.Bundle, req apistructs.KMSCreateKeyRequest) (*kmstypes.CreateKeyResponse, error) {
			return &kmstypes.CreateKeyResponse{}, errors.New("error")
		},
	)
	AddonInfos.Store("kkk", apistructs.Extension{})
	a.CreateCustom(&apistructs.CustomAddonCreateRequest{
		AddonName: "kkk",
	})

	_, _ = transCustomName2CloudName(apistructs.AddonCloudRedis)
	_, _ = transCustomName2CloudName(apistructs.AddonCloudRds)
	_, _ = transCustomName2CloudName(apistructs.AddonCloudOns)
	_, _ = transCustomName2CloudName(apistructs.AddonCloudOss)
	_, _ = transCustomName2CloudName(apistructs.AddonCloudGateway)
	_, _ = transCustomName2CloudName("")

	monkey.PatchInstanceMethod(reflect.TypeOf(a), "CreateCustom",
		func(a *Addon, req *apistructs.CustomAddonCreateRequest) (*map[string]string, error) {
			return &map[string]string{}, nil
		},
	)
	a.AddonYmlImport(0, diceyml.Object{
		AddOns: diceyml.AddOns{
			"a": {
				Plan: "custom:1",
				Options: map[string]string{
					"workspace": "DEV",
					"config":    "{}",
				},
			},
			"b": {
				Plan: "custom:2",
				Options: map[string]string{
					"workspace": "TEST",
					"config":    "{}",
				},
			},
		},
	}, "")
	a.ParseAddonFullPlan("")
	a.ParseAddonFullPlan("custom:1")

	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "ListAddonInstancesByProjectIDs",
		func(db *dbclient.DBClient, projectIDs []uint64, exclude ...string) (*[]dbclient.AddonInstance, error) {

			return &[]dbclient.AddonInstance{
				{AddonName: "custom"},
				{AddonName: "custom"},
			}, nil
		},
	)
	a.AddonYmlExport("1")

	monkey.PatchInstanceMethod(reflect.TypeOf(a.bdl), "GetProjectWithSetter",
		func(b *bundle.Bundle, id uint64, requestSetter ...httpclient.RequestSetter) (*apistructs.ProjectDTO, error) {
			return &apistructs.ProjectDTO{}, nil
		},
	)
	a.getProject("1")

	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "GetAddonInstance",
		func(db *dbclient.DBClient, id string) (*dbclient.AddonInstance, error) {
			return &dbclient.AddonInstance{
				Config: "{}",
			}, nil
		},
	)
	a.convert(&dbclient.AddonInstanceRouting{
		ProjectID:           "1",
		PlatformServiceType: 1,
	})
	a.convert(&dbclient.AddonInstanceRouting{
		ProjectID:           "1",
		PlatformServiceType: 1,
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "GetInstanceRoutingsByProjectIDs",
		func(db *dbclient.DBClient, platformServiceType int, projectIDs []string, az, env string) (*[]dbclient.AddonInstanceRouting, error) {
			return &[]dbclient.AddonInstanceRouting{
				{ProjectID: "1"},
			}, nil
		},
	)

	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "GetMicroAttachesByAddonNameAndProjectIDs",
		func(db *dbclient.DBClient, addonName string, projectIDs []string, env string) (*[]dbclient.AddonMicroAttach, error) {
			return &[]dbclient.AddonMicroAttach{
				{ID: 1},
			}, nil
		},
	)
	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "GetInstanceRoutingsByIDs",
		func(db *dbclient.DBClient, ids []string) (*[]dbclient.AddonInstanceRouting, error) {
			return &[]dbclient.AddonInstanceRouting{
				{ProjectID: "1"},
			}, nil
		},
	)

	a.ListMicroServiceProject([]string{"1", "2", "3"})

	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "GetInstancesByIDs",
		func(db *dbclient.DBClient, ids []string) (*[]dbclient.AddonInstance, error) {
			return &[]dbclient.AddonInstance{
				{ProjectID: "1"},
			}, nil
		},
	)

	a.ListMicroServiceMenu("1", "")

	a.appendMicroServiceProjectData(make(map[uint64]*apistructs.MicroServiceProjectResponseData), &apistructs.ProjectDTO{
		ID: 1,
	}, "")

	monkey.PatchInstanceMethod(reflect.TypeOf(a.db), "ListAddonInstancesByParamsWithoutPage",
		func(db *dbclient.DBClient, orgID uint64, params *apistructs.MiddlewareListRequest) ([]dbclient.AddonInstanceInfoExtra, error) {

			return []dbclient.AddonInstanceInfoExtra{
				{},
			}, nil
		},
	)

	a.MiddlewareListItem(1, 1, &apistructs.MiddlewareListRequest{}, []dbclient.AddonInstanceInfoExtra{{
		AddonInstance: dbclient.AddonInstance{ProjectID: "1"},
	}}, nil)

	a.GetMiddlewareAddonClassification(1, &apistructs.MiddlewareListRequest{})

	a.GetMiddlewareAddonDaily(1, &apistructs.MiddlewareListRequest{})

	monkey.PatchInstanceMethod(reflect.TypeOf(a), "ListReferencesByInstanceID",
		func(a *Addon, orgID uint64, userID, instanceID string) (*[]apistructs.AddonReferenceInfo, error) {

			return &[]apistructs.AddonReferenceInfo{
				{},
			}, nil
		},
	)

	a.GetMiddleware(1, "1", "1")

	a.InnerGetMiddleware("1")
}

func TestGetAddonExtention(t *testing.T) {
	defer monkey.UnpatchAll()

	bdl := &bundle.Bundle{}
	clusterSvc := &cluster.ClusterService{}
	a := New(WithBundle(bdl), WithClusterSvc(clusterSvc))
	ext := []apistructs.ExtensionVersion{
		{
			Name:    "custom",
			Version: "1.0",
			Spec: apistructs.AddonExtension{
				Type: "addon",
				Name: "custom",
				Plan: map[string]apistructs.AddonPlanItem{
					"custom": {},
				},
			},
			Dice: diceyml.Object{
				Version: "1.0",
				Services: map[string]*diceyml.Service{
					"custom": &diceyml.Service{},
				},
			},
		}, {
			Name:      "redis",
			Version:   "1.0",
			IsDefault: true,
			Spec: apistructs.AddonExtension{
				Type: "addon",
				Name: "redis",
				Plan: map[string]apistructs.AddonPlanItem{
					"redis": {},
				},
			},
			Dice: diceyml.Object{
				Version: "1.0",
				Services: map[string]*diceyml.Service{
					"redis": &diceyml.Service{},
				},
			},
		},
	}
	monkey.PatchInstanceMethod(reflect.TypeOf(a.bdl), "QueryExtensionVersions",
		func(bdl *bundle.Bundle, req apistructs.ExtensionVersionQueryRequest) ([]apistructs.ExtensionVersion, error) {
			return ext, nil
		},
	)

	monkey.Patch(i18n.OrgSprintf, func(orgID, key string, args ...interface{}) string {
		return ""
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(a.clusterSvc), "GetCluster",
		func(*cluster.ClusterService, context.Context, *pb.GetClusterRequest) (*pb.GetClusterResponse, error) {
			return &pb.GetClusterResponse{
				Data: &pb.ClusterInfo{
					Name: "erda",
					Cm: map[string]string{
						"ADDON_CUSTOM": "1.0",
					},
				},
			}, nil
		})

	_, _, err := a.GetAddonExtention(&apistructs.AddonHandlerCreateItem{
		AddonName: "custom",
		Options: map[string]string{
			"version": "1.0",
		},
		ClusterName: "erda",
	})
	assert.NoError(t, err)
	//assert.Equal(t, addon.Name, "custom")

	_, _, _ = a.GetAddonExtention(&apistructs.AddonHandlerCreateItem{
		AddonName: "reids",
		Options: map[string]string{
			"version": "1.2.0",
		},
		ClusterName: "erda",
	})
}

func mockGetUnDeletableAttachMentsByRuntimeID(db *dbclient.DBClient, orgId uint64, runtimeId uint64) (*[]dbclient.AddonAttachment, error) {
	// Add mock data
	ret := make([]dbclient.AddonAttachment, 0)
	return &ret, nil
}

func TestListByRuntime(t *testing.T) {
	// Create a new instance of your Addon struct and any required dependencies.
	addon := &Addon{}

	monkey.PatchInstanceMethod(reflect.TypeOf(addon.db), "GetUnDeletableAttachMentsByRuntimeID", mockGetUnDeletableAttachMentsByRuntimeID)
	defer monkey.UnpatchAll()

	// Define your test cases using args for-loop.
	testCases := []struct {
		name      string
		orgID     uint64
		runtimeID uint64
		projectID string
		workspace string
		// Add any other fields relevant for the test case.
		expected    *[]apistructs.AddonFetchResponseData
		expectedErr error
	}{
		// Test case 1
		{
			name:        "Empty Addons",
			orgID:       1,
			runtimeID:   1001,
			projectID:   "10",
			workspace:   "dev",
			expected:    &[]apistructs.AddonFetchResponseData{},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the ListByRuntime function with the test case inputs.
			actual, err := addon.ListByRuntime(tc.orgID, tc.runtimeID, tc.projectID, tc.workspace)

			// Assert the results.
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestDeploy(t *testing.T) {
	addon := &Addon{db: &dbclient.DBClient{DBEngine: &dbengine.DBEngine{}}}
	monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(addon.db), "FindTmcInstanceByNameAndCLuster", func(db *dbclient.DBClient, name, cluster string) ([]dbclient.TmcInstance, error) {
		return []dbclient.TmcInstance{{
			ID:      "1234",
			Engine:  "nacos",
			Version: "2.2",
		}}, nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(addon.db), "UpdatePrebuild", func(db *dbclient.DBClient, addonPrebuild *dbclient.AddonPrebuild) error {
		return nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(bundle.New()), "QueryExtensionVersions", func(b *bundle.Bundle, req apistructs.ExtensionVersionQueryRequest) ([]apistructs.ExtensionVersion, error) {
		switch req.Name {
		case RegisterCenterAddon:
			addonExtension := apistructs.AddonExtension{
				Type:        "addon",
				Name:        "registercenter",
				Desc:        "${{ i18n.desc }}",
				DisplayName: "${{ i18n.displayName }}",
				Category:    "distributed_cooperation",
				LogoUrl:     "//terminus-dice.oss-cn-hangzhou.aliyuncs.com/addon/ui/logo/terminus-roost.png",
				ImageURLs: []string{
					"//terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2019/06/27/3f052a37-3f35-4738-bdeb-50ee5d9fe2d7.png",
					"//terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2019/06/27/6ba281ea-f96c-4388-be6c-d60cfa88bc89.png",
				},
				Strategy:    map[string]interface{}{"parsingAddons": false},
				Version:     "3.0.0",
				SubCategory: "microservice",
				Domain:      "tmc.marathon.l4lb.thisdcos.directory:8050",
				Requires:    []string{},
				ConfigVars:  []string{"NACOS_ADDRESS", "NACOS_GRPC_ADDRESS", "NACOS_TENANT_ID", "ELASTICJOB_HOST"},
				Envs:        []string{},
				Plan:        map[string]apistructs.AddonPlanItem{},
				ShareScopes: []string{"PROJECT"},
				Similar:     []string{},
				Deprecated:  false,
			}

			// Object data
			object := &diceyml.Object{
				Version:  "2.0",
				Meta:     map[string]string{},
				Envs:     diceyml.EnvMap{},
				Services: diceyml.Services{},
				Jobs:     diceyml.Jobs{},
				AddOns: diceyml.AddOns{
					"zookeeper": {
						Plan: "zookeeper:basic",
					},
					"nacos": {
						Plan: "nacos:basic",
						Options: map[string]string{
							"version": "2.2",
						},
					},
				},
				Environments: diceyml.EnvObjects{},
				Values:       diceyml.ValueObjects{},
			}
			return []apistructs.ExtensionVersion{
				{Version: "3.0.0",
					Name:      RegisterCenterAddon,
					Type:      addonExtension.Type,
					Spec:      addonExtension,
					Dice:      object,
					Swagger:   nil,
					Readme:    "",
					IsDefault: true,
					Public:    true},
			}, nil
		case ConfigCenterAddon:
			addonExtension := apistructs.AddonExtension{
				Type:        "addon",
				Name:        "config-center",
				Desc:        "${{ i18n.desc }}",
				DisplayName: "${{ i18n.displayName }}",
				Category:    "content_management",
				LogoUrl:     "//terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2019/10/15/63628afc-b340-40d6-9e25-2c72568e2d42.png",
				ImageURLs:   []string{},
				Strategy:    map[string]interface{}{"parsingAddons": false},
				Version:     "3.0.0",
				SubCategory: "microservice",
				Domain:      "tmc.marathon.l4lb.thisdcos.directory:8050",
				Requires:    []string{},
				ConfigVars: []string{
					"CONFIGCENTER_ADDRESS",
					"CONFIGCENTER_GRPC_ADDRESS",
					"CONFIGCENTER_TENANT_ID",
					"CONFIGCENTER_TENANT_NAME",
					"CONFIGCENTER_GROUP_NAME",
				},
				Envs:        []string{},
				Plan:        map[string]apistructs.AddonPlanItem{},
				ShareScopes: []string{"APPLICATION"},
				Similar:     []string{},
				Deprecated:  false,
			}
			object := diceyml.Object{
				Version:  "2.0",
				Meta:     map[string]string{},
				Envs:     diceyml.EnvMap{},
				Services: diceyml.Services{},
				Jobs:     diceyml.Jobs{},
				AddOns: diceyml.AddOns{
					"nacos": {
						Plan: "nacos:basic",
						Options: map[string]string{
							"version": "2.2",
						},
					},
				},
				Environments: diceyml.EnvObjects{},
				Values:       diceyml.ValueObjects{},
			}
			return []apistructs.ExtensionVersion{
				{Name: addonExtension.Name,
					Version:   addonExtension.Version,
					Type:      addonExtension.Type,
					Spec:      addonExtension,
					Dice:      object,
					Swagger:   nil,
					Readme:    "",
					IsDefault: true,
					Public:    true},
			}, nil
		}
		return nil, errors.New("no such addon")
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(addon), "AttachAndCreate", func(a *Addon, params *apistructs.AddonHandlerCreateItem) (*apistructs.AddonInstanceRes, error) {
		return &apistructs.AddonInstanceRes{}, nil
	})

	defer func() {
		if err := recover(); err != nil {
			t.Log(err)
		}
	}()

	fmt.Println(addon.db == nil)
	AddonInfos.Store(RegisterCenterAddon, "")
	AddonInfos.Store(ConfigCenterAddon, "")
	_ = addon.deployAddons(&apistructs.AddonCreateRequest{
		ClusterName:   RegisterCenterAddon,
		OrgID:         232,
		ProjectID:     212,
		ApplicationID: 12321,
		RuntimeID:     23123,
	}, []dbclient.AddonPrebuild{
		{
			AddonName:         RegisterCenterAddon,
			InstanceName:      RegisterCenterAddon,
			Env:               "DEV",
			RuntimeID:         "100001",
			RoutingInstanceID: "100001",
			InstanceID:        "100001",
			Plan:              "basic",
			Options:           `{"version": "2.0.0"}`,
		}, {
			AddonName:         ConfigCenterAddon,
			InstanceName:      ConfigCenterAddon,
			Env:               "DEV",
			RuntimeID:         "100001",
			RoutingInstanceID: "100001",
			InstanceID:        "100001",
			Plan:              "basic",
		},
	})
}

func TestInjectVersion(t *testing.T) {
	defer monkey.UnpatchAll()

	InitCache(DefaultTtl, DefaultSize)

	monkey.PatchInstanceMethod(reflect.TypeOf(cache.bdl), "QueryExtensionVersions", func(_ *bundle.Bundle,
		req apistructs.ExtensionVersionQueryRequest) ([]apistructs.ExtensionVersion, error) {
		switch req.Name {
		case "registercenter":
			return []apistructs.ExtensionVersion{
				{Name: "registercenter", Version: "3.0.0", IsDefault: true},
				{Name: "registercenter", Version: "2.0.0"},
			}, nil
		case "config-center":
			return []apistructs.ExtensionVersion{
				{Name: "config-center", Version: "3.0.0", IsDefault: true},
				{Name: "config-center", Version: "2.0.0"},
			}, nil
		default:
			return []apistructs.ExtensionVersion{}, errors.New("not found")
		}
	})

	addon := &Addon{db: &dbclient.DBClient{DBEngine: &dbengine.DBEngine{}}}

	tests := []struct {
		Name    string
		Type    string
		Options map[string]string
	}{
		{
			Name:    "config-center",
			Type:    "config-center",
			Options: nil,
		},
		{
			Name: "register-center",
			Type: "registercenter",
			Options: map[string]string{
				"version": "2.0.0",
			},
		},
	}

	for _, tt := range tests {
		err := addon.injectVersion(&apistructs.AddonCreateItem{
			Name:    tt.Name,
			Type:    tt.Type,
			Options: tt.Options,
		})
		assert.NoError(t, err, "Expected no error, but got ", err)
	}

}

func TestAddon_preparePrebuildChanges_CreateNew(t *testing.T) {
	a := &Addon{}
	req := &apistructs.AddonCreateRequest{
		ApplicationID: 1,
		RuntimeID:     2,
		RuntimeName:   "master",
		Workspace:     "DEV",
		Addons: []apistructs.AddonCreateItem{
			{
				Name: "roost",
				Type: apistructs.AddonRoost,
				Plan: apistructs.AddonBasic,
			},
		},
	}

	toCreate, toUpdate, newPrebuildList, err := a.preparePrebuildChanges(req, map[string]dbclient.AddonPrebuild{}, false, false)
	assert.NoError(t, err)
	assert.Empty(t, toUpdate)
	assert.Len(t, toCreate, 1)
	assert.Len(t, newPrebuildList, 1)

	assert.Equal(t, "1", toCreate[0].ApplicationID)
	assert.Equal(t, "2", toCreate[0].RuntimeID)
	assert.Equal(t, "master", toCreate[0].GitBranch)
	assert.Equal(t, "DEV", toCreate[0].Env)
	assert.Equal(t, "roost", toCreate[0].InstanceName)
	assert.Equal(t, apistructs.AddonZKProxy, toCreate[0].AddonName)
	assert.Equal(t, apistructs.AddonBasic, toCreate[0].Plan)
}

func TestAddon_preparePrebuildChanges_UpdateDiceYmlDeleted(t *testing.T) {
	a := &Addon{}
	req := &apistructs.AddonCreateRequest{
		ApplicationID: 1,
		RuntimeID:     2,
		RuntimeName:   "master",
		Workspace:     "DEV",
		Addons: []apistructs.AddonCreateItem{
			{
				Name: "zk",
				Type: apistructs.AddonZookeeperAlias,
				Plan: apistructs.AddonProfessional,
			},
		},
	}

	existed := dbclient.AddonPrebuild{
		ID:           100,
		RuntimeID:    "2",
		InstanceName: "zk",
		AddonName:    apistructs.AddonZookeeper,
		Plan:         apistructs.AddonBasic,
		DeleteStatus: apistructs.AddonPrebuildDiceYmlDeleted,
		Options:      `{"k":"v"}`,
	}
	existBuildMap := map[string]dbclient.AddonPrebuild{
		strutil.Concat(existed.RuntimeID, existed.AddonName, existed.InstanceName): existed,
	}

	toCreate, toUpdate, newPrebuildList, err := a.preparePrebuildChanges(req, existBuildMap, false, false)
	assert.NoError(t, err)
	assert.Empty(t, toCreate)
	assert.Len(t, toUpdate, 1)
	assert.Len(t, newPrebuildList, 1)

	assert.Equal(t, uint64(100), toUpdate[0].ID)
	assert.Equal(t, apistructs.AddonPrebuildNotDeleted, toUpdate[0].DeleteStatus)
	assert.Equal(t, apistructs.AddonBasic, toUpdate[0].Plan)
	assert.Equal(t, `{"k":"v"}`, toUpdate[0].Options)
}

func TestAddon_preparePrebuildChanges_UpdateNotDeleted_PlanAndOptions(t *testing.T) {
	a := &Addon{}
	req := &apistructs.AddonCreateRequest{
		ApplicationID: 1,
		RuntimeID:     2,
		RuntimeName:   "master",
		Workspace:     "DEV",
		Addons: []apistructs.AddonCreateItem{
			{
				Name: "registercenter",
				Type: RegisterCenterAddon,
				Plan: apistructs.AddonProfessional,
				Options: map[string]string{
					"version": "2.0.0",
				},
			},
		},
	}

	existed := dbclient.AddonPrebuild{
		ID:           200,
		RuntimeID:    "2",
		InstanceName: "registercenter",
		AddonName:    RegisterCenterAddon,
		Plan:         apistructs.AddonBasic,
		DeleteStatus: apistructs.AddonPrebuildNotDeleted,
		Options:      `{"version":"1.0.0"}`,
	}
	existBuildMap := map[string]dbclient.AddonPrebuild{
		strutil.Concat(existed.RuntimeID, existed.AddonName, existed.InstanceName): existed,
	}

	toCreate, toUpdate, newPrebuildList, err := a.preparePrebuildChanges(req, existBuildMap, false, false)
	assert.NoError(t, err)
	assert.Empty(t, toCreate)
	assert.Len(t, toUpdate, 1)
	assert.Len(t, newPrebuildList, 1)

	assert.Equal(t, uint64(200), toUpdate[0].ID)
	assert.Equal(t, apistructs.AddonProfessional, toUpdate[0].Plan)
	assert.Equal(t, `{"version":"2.0.0"}`, toUpdate[0].Options)
}

func TestAddon_preparePrebuildChanges_NoUpdateWhenUnchanged(t *testing.T) {
	a := &Addon{}
	req := &apistructs.AddonCreateRequest{
		ApplicationID: 1,
		RuntimeID:     2,
		RuntimeName:   "master",
		Workspace:     "DEV",
		Addons: []apistructs.AddonCreateItem{
			{
				Name: "registercenter",
				Type: RegisterCenterAddon,
				Plan: apistructs.AddonBasic,
				Options: map[string]string{
					"version": "1.0.0",
				},
			},
		},
	}

	existed := dbclient.AddonPrebuild{
		ID:           300,
		RuntimeID:    "2",
		InstanceName: "registercenter",
		AddonName:    RegisterCenterAddon,
		Plan:         apistructs.AddonBasic,
		DeleteStatus: apistructs.AddonPrebuildNotDeleted,
		Options:      `{"version":"1.0.0"}`,
	}
	existBuildMap := map[string]dbclient.AddonPrebuild{
		strutil.Concat(existed.RuntimeID, existed.AddonName, existed.InstanceName): existed,
	}

	toCreate, toUpdate, newPrebuildList, err := a.preparePrebuildChanges(req, existBuildMap, false, false)
	assert.NoError(t, err)
	assert.Empty(t, toCreate)
	assert.Empty(t, toUpdate)
	assert.Len(t, newPrebuildList, 1)
	assert.Equal(t, uint64(300), newPrebuildList[0].ID)
}
