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

package events

// func TestDeployTimeCollector_collectDeployTimes(t *testing.T) {
// 	db := initDb(t)
// 	c := NewDeployTimeCollector(nil, db)
//
// 	now := time.Now()
// 	deployment := &dbclient.Deployment{
// 		Status: apistructs.DeploymentStatusDeploying,
// 		Phase:  apistructs.DeploymentPhaseInit,
// 	}
// 	err := db.CreateDeployment(deployment)
// 	if !assert.NoError(t, err) {
// 		return
// 	}
// 	e := RuntimeEvent{
// 		EventName: RuntimeDeployStatusChanged,
// 		Deployment: &apistructs.Deployment{
// 			ID:     deployment.ID,
// 			Status: apistructs.DeploymentStatusDeploying,
// 			Phase:  apistructs.DeploymentPhaseInit,
// 		},
// 	}
// 	// init
// 	(*c).OnEvent(&e)
// 	deployment, err = db.GetDeployment(deployment.ID)
// 	if !assert.NoError(t, err) {
// 		return
// 	}
// 	assert.True(t, deployment.Extra.AddonPhaseStartAt == nil)
// 	assert.True(t, deployment.Extra.AddonPhaseEndAt == nil)
// 	assert.True(t, deployment.Extra.ServicePhaseStartAt == nil)
// 	assert.True(t, deployment.Extra.ServicePhaseEndAt == nil)
//
// 	// leap to addon
// 	deployment.Phase = apistructs.DeploymentPhaseAddon
// 	err = db.UpdateDeployment(deployment)
// 	if !assert.NoError(t, err) {
// 		return
// 	}
// 	e.Deployment.Phase = apistructs.DeploymentPhaseAddon
// 	(*c).OnEvent(&e)
// 	deployment, err = db.GetDeployment(deployment.ID)
// 	if !assert.NoError(t, err) {
// 		return
// 	}
// 	assert.True(t, deployment.Extra.AddonPhaseStartAt.After(now))
// 	assert.True(t, deployment.Extra.AddonPhaseEndAt == nil)
// 	assert.True(t, deployment.Extra.ServicePhaseStartAt == nil)
// 	assert.True(t, deployment.Extra.ServicePhaseEndAt == nil)
//
// 	// leap to service
// 	deployment.Phase = apistructs.DeploymentPhaseScript
// 	err = db.UpdateDeployment(deployment)
// 	if !assert.NoError(t, err) {
// 		return
// 	}
// 	e.Deployment.Phase = apistructs.DeploymentPhaseScript
// 	(*c).OnEvent(&e)
// 	deployment, err = db.GetDeployment(deployment.ID)
// 	if !assert.NoError(t, err) {
// 		return
// 	}
// 	assert.True(t, deployment.Extra.AddonPhaseStartAt.After(now))
// 	assert.True(t, deployment.Extra.AddonPhaseEndAt.After(now))
// 	assert.True(t, deployment.Extra.ServicePhaseStartAt.After(now))
// 	assert.True(t, deployment.Extra.ServicePhaseEndAt == nil)
//
// 	// all completed
// 	deployment.Phase = apistructs.DeploymentPhaseCompleted
// 	err = db.UpdateDeployment(deployment)
// 	if !assert.NoError(t, err) {
// 		return
// 	}
// 	e.Deployment.Phase = apistructs.DeploymentPhaseCompleted
// 	(*c).OnEvent(&e)
// 	deployment, err = db.GetDeployment(deployment.ID)
// 	if !assert.NoError(t, err) {
// 		return
// 	}
// 	assert.True(t, deployment.Extra.AddonPhaseStartAt.After(now))
// 	assert.True(t, deployment.Extra.AddonPhaseEndAt.After(now))
// 	assert.True(t, deployment.Extra.ServicePhaseStartAt.After(now))
// 	assert.True(t, deployment.Extra.ServicePhaseEndAt.After(now))
// }
//
// func initDb(t *testing.T) *dbclient.DBClient {
// 	os.Setenv("MYSQL_HOST", "127.0.0.1")
// 	os.Setenv("MYSQL_PORT", "3306")
// 	os.Setenv("MYSQL_DATABASE", "orchestrator")
// 	os.Setenv("MYSQL_USERNAME", "root")
// 	client, err := dbclient.Open()
// 	if assert.Nil(t, err) {
// 		return client
// 	} else {
// 		panic(err)
// 	}
// }
