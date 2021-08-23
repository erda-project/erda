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

package deployment

import (
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/events"
	"github.com/erda-project/erda/modules/orchestrator/services/log"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func TestConvertGroupLabels(t *testing.T) {
	app := apistructs.ApplicationDTO{
		OrgID:       1,
		OrgName:     "terminus",
		ProjectID:   2,
		ProjectName: "default",
		ID:          3,
		Name:        "pampas-blog",
		Workspaces: []apistructs.ApplicationWorkspace{
			{
				Workspace:       "DEV",
				ClusterName:     "terminus-dev",
				ConfigNamespace: "app-3-DEV",
			},
		},
	}
	runtime := dbclient.Runtime{
		BaseModel: dbengine.BaseModel{
			ID: 4,
		},
		Name:        "master",
		Workspace:   "DEV",
		ClusterName: "terminus-dev",
	}
	groupLabels := convertGroupLabels(&app, &runtime, 5)
	expect := map[string]string{
		"SERVICE_TYPE":              "STATELESS",
		"SERVICE_DISCOVERY_MODE":    "DEPEND", // default values are used for service discovery
		"DICE_ORG":                  "1",
		"DICE_ORG_ID":               "1",
		"DICE_ORG_NAME":             "terminus",
		"DICE_PROJECT":              "2",
		"DICE_PROJECT_ID":           "2",
		"DICE_PROJECT_NAME":         "default",
		"DICE_APPLICATION":          "3",
		"DICE_APPLICATION_ID":       "3",
		"DICE_APPLICATION_NAME":     "pampas-blog",
		"DICE_WORKSPACE":            "dev",
		"DICE_CLUSTER_NAME":         "terminus-dev",
		"DICE_RUNTIME":              "4",
		"DICE_RUNTIME_ID":           "4",
		"DICE_RUNTIME_NAME":         "master",
		"DICE_DEPLOYMENT":           "5",
		"DICE_DEPLOYMENT_ID":        "5",
		"DICE_APP_CONFIG_NAMESPACE": "app-3-DEV",
	}
	assert.Equal(t, expect, groupLabels)
}

func TestFSMTimeout(t *testing.T) {
	f := genFakeFSM()

	_ = recordUpdateDeployment()
	_ = recordEvent()
	_ = recordDLog()

	// do invoke
	f.Deployment.UpdatedAt = time.Now().Add(-61 * time.Minute)
	end, err := f.timeout()
	if assert.NoError(t, err) {
		assert.True(t, end)
	}

	f.Deployment.UpdatedAt = time.Now().Add(-59 * time.Minute)
	end, err = f.timeout()
	if assert.NoError(t, err) {
		assert.False(t, end)
	}
}

func TestFSMFailDeploy(t *testing.T) {
	f := genFakeFSM()

	defer monkey.UnpatchAll()
	updateC := recordUpdateDeployment()
	eventC := recordEvent()
	loggingC := recordDLog()

	// do invoke
	err := f.failDeploy(errors.Errorf("fake error"))

	if assert.NoError(t, err) {
		updates := collectUpdateDeployment(updateC)
		if assert.Len(t, updates, 1) {
			assert.Equal(t, "FAILED", string(updates[0].Status))
			assert.Equal(t, "INIT", string(updates[0].Phase))
		}

		es := collectEvent(eventC)
		if assert.Len(t, es, 1) {
			assert.Equal(t, "RuntimeDeployFailed", string(es[0].EventName))
			assert.Equal(t, "FAILED", string(es[0].Deployment.Status))
			assert.Equal(t, "INIT", string(es[0].Deployment.Phase))
			assert.Equal(t, uint64(101), es[0].Runtime.ID)
		}

		logging := collectDLog(loggingC)
		assert.Equal(t, []string{
			`deployment is fail, status: WAITING, phrase: INIT, (fake error)`,
		}, logging)
	}
}

func TestFSMPushOnPhase(t *testing.T) {
	f := genFakeFSM()

	defer monkey.UnpatchAll()
	updateC := recordUpdateDeployment()
	eventC := recordEvent()

	// do invoke
	err := f.pushOnPhase(apistructs.DeploymentPhaseAddon)

	if assert.NoError(t, err) {
		updates := collectUpdateDeployment(updateC)
		if assert.Len(t, updates, 1) {
			assert.Equal(t, "WAITING", string(updates[0].Status)) // although this is not legal
			assert.Equal(t, "ADDON_REQUESTING", string(updates[0].Phase))
		}

		es := collectEvent(eventC)
		if assert.Len(t, es, 1) {
			assert.Equal(t, "RuntimeDeployStatusChanged", string(es[0].EventName))
		}
	}
}

// func TestFSMRequestAddons(t *testing.T) {
// 	f := genFakeFSM()
//
// 	// patch methods
// 	var bdl *bundle.Bundle
// 	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetCluster",
// 		func(_ *bundle.Bundle, clusterName string) (*apistructs.ClusterInfo, error) {
// 			cluster := apistructs.ClusterInfo{
// 				ID:   rand.Int(),
// 				Name: clusterName,
// 			}
// 			return &cluster, nil
// 		},
// 	)
// 	raw := `version: 2.0
// services:
//   none:
//     image: nginx:latest
//     resources:
//       cpu: 0.01
//       mem: 64
//       disk: 33
//     deployments:
//       replicas: 1`
// 	rawYAML, err := diceyml.New([]byte(raw), true)
// 	if !assert.NoError(t, err) {
// 		return
// 	}
// 	cntFetchYml := 0
// 	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetDiceYAML",
// 		func(_ *bundle.Bundle, releaseID string, workspace ...string) (*diceyml.DiceYaml, error) {
// 			cntFetchYml++
// 			assert.Equal(t, "xxx-yyy", releaseID)
// 			return rawYAML, nil
// 		},
// 	)
// 	cntUnderlay := 0
// 	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "UpdateAddonUnderlay",
// 		func(_ *bundle.Bundle, uniqueId spec.RuntimeUniqueId, orgID uint64, projectID uint64,
// 			isSourceAbility bool, dice *spec.LegacyDice, diceYml *diceyml.DiceYaml) error {
// 			cntUnderlay++
// 			// check update underlay arguments
// 			assert.Equal(t, spec.RuntimeUniqueId{ApplicationId: 102, Workspace: "DEV", Name: "fake runtime"}, uniqueId)
// 			assert.Equal(t, uint64(104), orgID)
// 			assert.Equal(t, uint64(103), projectID)
// 			assert.False(t, isSourceAbility)
// 			return nil
// 		},
// 	)
// 	cntReqAddon := 0
// 	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CreateAddon",
// 		func(_ *bundle.Bundle, req *apistructs.AddonCreateRequest) error {
// 			cntReqAddon += 1
// 			// check req
// 			assert.Equal(t, uint64(104), req.OrgID)
// 			assert.Equal(t, uint64(103), req.ProjectID)
// 			assert.Equal(t, uint64(102), req.ApplicationID)
// 			assert.Equal(t, "DEV", req.Workspace)
// 			assert.Equal(t, uint64(101), req.RuntimeID)
// 			assert.Equal(t, "fake runtime", req.RuntimeName)
// 			assert.Equal(t, "terminus-test", req.ClusterName)
// 			assert.Equal(t, "fake user", req.Operator)
// 			assert.Equal(t, apistructs.AddonCreateOptions{
// 				OrgName:         "fake org",
// 				ProjectName:     "fake project",
// 				ApplicationName: "fake app",
// 				Workspace:       "DEV",
// 				RuntimeName:     "fake runtime",
// 				DeploymentID:    "100",
// 				ClusterName:     "terminus-test",
// 			}, req.Options)
// 			return nil
// 		},
// 	)
//
// 	// do invoke
// 	err = f.requestAddons()
//
// 	if assert.NoError(t, err) {
// 		assert.Equal(t, 1, cntUnderlay)
// 		assert.Equal(t, 1, cntFetchYml)
// 		assert.Equal(t, 1, cntReqAddon)
// 	}
// }

// func TestFSMGenerateDeployServiceRequest(t *testing.T) {
// 	input := "../../testdata/orc_pmp_release1.yml"
// 	output := "../../testdata/orc_pmp_req1.yml"
//
// 	f := genFakeFSM(input)
//
// 	// patch methods
// 	var bdl *bundle.Bundle
// 	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetRuntimeAddonConfig",
// 		func(_ *bundle.Bundle) (map[string]string, error) {
// 			return map[string]string{
// 				"FAKE_ADDON_ENV1": "it",
// 				"FAKE_ADDON_ENV2": "is",
// 				"FAKE_OVERLAP":    "addon",
// 			}, nil
// 		},
// 	)
// 	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "FetchDeploymentConfig",
// 		func(_ *bundle.Bundle, configNamespace string) (map[string]string, map[string]string, error) {
// 			return map[string]string{
// 					"FAKE_CONFIG_ENV1": "fake tan",
// 					"FAKE_CONFIG_ENV2": "fake ke",
// 					"FAKE_OVERLAP":     "config",
// 				}, map[string]string{
// 					"FAKE_CONFIG_FILE1": "fake file1",
// 					"FAKE_CONFIG_FILE2": "fake file2",
// 				}, nil
// 		},
// 	)
// 	var db *dbclient.DBClient
// 	monkey.PatchInstanceMethod(reflect.TypeOf(db), "FindDomainsByRuntimeIdAndServiceName",
// 		func(_ *dbclient.DBClient, runtimeId uint64, serviceName string) ([]dbclient.RuntimeDomain, error) {
// 			return []dbclient.RuntimeDomain{
// 				{
// 					RuntimeId: runtimeId,
// 					Domain:    fmt.Sprintf("%s-dev-%d.test.terminus.io", serviceName, runtimeId),
// 				},
// 			}, nil
// 		},
// 	)
//
// 	// do invoke
// 	group := apistructs.ServiceGroupCreateV2Request{}
// 	_, _, err := f.generateDeployServiceRequest(&group, []dbclient.AddonInstanceRouting{}, []dbclient.AddonInstanceTenant{})
// 	require.NoError(t, err)
//
// 	actualDice := group.DiceYml
//
// 	b, err := ioutil.ReadFile(output)
// 	require.NoError(t, err)
// 	var expectDice diceyml.Object
// 	err = yaml.Unmarshal(b, &expectDice)
// 	require.NoError(t, err)
//
// 	// 1. check dice.yml
// 	assert.Equal(t, expectDice, actualDice)
//
// 	// 2. check request
// 	group.DiceYml = diceyml.Object{} // diceYml is checked
// 	require.Equal(t, apistructs.ServiceGroupCreateV2Request{
// 		ID:          "schedule",
// 		Type:        "fake",
// 		ClusterName: "terminus-test",
// 	}, group)
// }

func TestFSMContinueCanceling(t *testing.T) {
	f := genFakeFSM()
	f.Deployment.Phase = apistructs.DeploymentPhaseAddon
	f.Deployment.Status = apistructs.DeploymentStatusCanceling
	f.Runtime.Status = apistructs.RuntimeStatusHealthy

	updateC := recordUpdateDeployment()
	emitC := recordEvent()
	loggingC := recordDLog()

	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CancelServiceGroup",
		func(_ *bundle.Bundle, namespace, name string) error {
			assert.Equal(t, "fake", namespace)
			assert.Equal(t, "schedule", name)
			return nil
		},
	)

	// do invoke
	err := f.continueCanceling()
	if assert.NoError(t, err) {
		updates := collectUpdateDeployment(updateC)
		assert.Len(t, updates, 1)

		es := collectEvent(emitC)
		assert.Len(t, es, 1)

		logging := collectDLog(loggingC)
		assert.Equal(t, []string{`deployment canceled`}, logging)
	}
}

func TestPrecheck(t *testing.T) {
	fsm := genFakeFSM()

	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CreateErrorLog",
		func(_ *bundle.Bundle, errorLog *apistructs.ErrorLogCreateRequest) error {
			return nil
		},
	)

	fsm.bdl = bdl
	fsm.Spec = &diceyml.Object{
		Services: map[string]*diceyml.Service{"fakesvC1-12": nil, "fakesvc2": nil},
	}

	// do invoke
	err := fsm.precheck()
	assert.Error(t, err)
}

func recordUpdateDeployment() chan dbclient.Deployment {
	var db *dbclient.DBClient
	c := make(chan dbclient.Deployment, 1000)
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "UpdateDeployment",
		func(_ *dbclient.DBClient, toUpdate *dbclient.Deployment) error {
			c <- *toUpdate
			return nil
		},
	)
	return c
}

func collectUpdateDeployment(c chan dbclient.Deployment) []dbclient.Deployment {
	close(c)
	updates := make([]dbclient.Deployment, 0)
	for u := range c {
		updates = append(updates, u)
	}
	return updates
}

func recordEvent() chan events.RuntimeEvent {
	var event *events.EventManager
	c := make(chan events.RuntimeEvent, 1000)
	monkey.PatchInstanceMethod(reflect.TypeOf(event), "EmitEvent",
		func(_ *events.EventManager, e *events.RuntimeEvent) {
			c <- *e
		},
	)
	return c
}

func collectEvent(c chan events.RuntimeEvent) []events.RuntimeEvent {
	defer close(c)
	es := make([]events.RuntimeEvent, 0)
	for {
		select {
		case e := <-c:
			es = append(es, e)
		default:
			return es
		}
	}
}

func recordDLog() chan string {
	var logger *log.DeployLogHelper
	c := make(chan string, 1000)
	monkey.PatchInstanceMethod(reflect.TypeOf(logger), "Log",
		func(_ *log.DeployLogHelper, content string) {
			c <- content
		},
	)
	return c
}

func collectDLog(c chan string) []string {
	defer close(c)
	lo := make([]string, 0)
	for {
		select {
		case l := <-c:
			lo = append(lo, l)
		default:
			return lo
		}
	}
}

func genFakeFSM(specPath ...string) *DeployFSMContext {
	fsm := DeployFSMContext{
		Deployment: &dbclient.Deployment{
			BaseModel: dbengine.BaseModel{
				ID: 100,
			},
			RuntimeId: 101,
			Status:    apistructs.DeploymentStatusWaiting,
			Phase:     apistructs.DeploymentPhaseInit,
			ReleaseId: "xxx-yyy",
			Operator:  "fake user",
		},
		Runtime: &dbclient.Runtime{
			BaseModel: dbengine.BaseModel{
				ID: 101,
			},
			Name:          "fake runtime",
			ApplicationID: 102,
			Workspace:     "DEV",
			ClusterName:   "terminus-test",
			ClusterId:     999,
			ScheduleName:  dbclient.ScheduleName{Namespace: "fake", Name: "schedule"},
			GitRepoAbbrev: "fake/runtime",
		},
		App: &apistructs.ApplicationDTO{
			ID:          102,
			Name:        "fake app",
			ProjectID:   103,
			ProjectName: "fake project",
			OrgID:       104,
			OrgName:     "fake org",
			Workspaces: []apistructs.ApplicationWorkspace{
				{
					Workspace:       "DEV",
					ClusterName:     "terminus-test",
					ConfigNamespace: "test-config-namespace",
				},
			},
		},
		Cluster: &apistructs.ClusterInfo{
			Name:           "terminus-test",
			OrgID:          104,
			WildcardDomain: "test.terminus.io",
		},
	}
	if len(specPath) > 0 {
		b, err := ioutil.ReadFile(specPath[0])
		if err != nil {
			panic(err)
		}
		y, err := diceyml.New(b, true)
		if err != nil {
			panic(err)
		}
		err = y.MergeEnv("DEV")
		if err != nil {
			panic(err)
		}
		fsm.Spec = y.Obj()
	}
	return &fsm
}
