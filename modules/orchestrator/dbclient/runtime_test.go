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

package dbclient

import (
	"fmt"
	"strings"
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/stretchr/testify/assert"
)

// func TestDBClient_CreateRuntime(t *testing.T) {
// 	client := initDb(t)
//
// 	err := client.CreateRuntime(&Runtime{
// 		ApplicationID: 1,
// 		Creator:       "55",
// 		Workspace:     "TEST",
// 		Name:          "feature/test3",
// 		Status:        "Healthy",
// 		LegacyStatus:  "DELETING",
// 		Deleting:      false,
// 		Version:       "1",
// 		ScheduleName:  ScheduleName{Namespace: "services", Name: "prod-123"},
// 		Source:        "PIPELINE",
// 		DiceVersion:   "2",
// 	})
// 	assert.NoError(t, err)
// }
//
// func TestDBClient_FindRuntimeServices(t *testing.T) {
// 	client := initDb(t)
//
// 	service := RuntimeService{
// 		RuntimeId: 1,
// 		Errors:    `[{"code":"InstanceFailed","msg":"实例 runtimes_v1_services_dev-4_web.c1ba9ec8-eee3-11e8-93c0-02420b542c69 启动失败, Failed/BeforeHealthCheckTimeout","ctx":null}]`,
// 	}
// 	err := client.CreateOrUpdateRuntimeService(&service, true)
// 	if !assert.NoError(t, err) {
// 		return
// 	}
//
// 	rss, err := client.FindRuntimeServices(1)
// 	if assert.NoError(t, err) {
// 		assert.Equal(t, 1, len(rss))
// 	}
//
// 	var errs []apistructs.ErrorResponse
// 	err = json.Unmarshal([]byte(rss[0].Errors), &errs)
// 	if assert.NoError(t, err) {
// 		if assert.Equal(t, 1, len(errs)) {
// 			assert.Equal(t, "InstanceFailed", errs[0].Code)
// 		}
// 	}
// }
//
// func TestDBClient_GetInstanceByTaskId(t *testing.T) {
// 	client := initDb(t)
//
// 	instance := RuntimeInstance{
// 		InstanceId: "111",
// 		Status:     "test",
// 		Stage:      "test",
// 	}
// 	client.CreateInstance(&instance)
//
// 	found, err := client.GetInstanceByTaskId("111")
// 	if assert.NoError(t, err) && assert.NotNil(t, found) {
// 		assert.Equal(t, "test", found.Stage)
// 	}
// }
//
// func TestDBClient_FindRuntimesByAppId(t *testing.T) {
// 	client := initDb(t)
//
// 	rs, err := client.FindRuntimesByAppId(1)
// 	require.NoError(t, err)
//
// 	fmt.Println(rs)
// }
//
// func initDb(t *testing.T) *DBClient {
// 	os.Setenv("MYSQL_HOST", "127.0.0.1")
// 	os.Setenv("MYSQL_PORT", "3306")
// 	os.Setenv("MYSQL_DATABASE", "orchestrator")
// 	os.Setenv("MYSQL_USERNAME", "root")
// 	client, err := Open()
// 	if assert.NoError(t, err) {
// 		return client
// 	}
// 	t.FailNow()
// 	return nil
// }

func TestInitScheduleName(t *testing.T) {
	r := Runtime{
		ApplicationID: 1,
		Workspace:     "dev",
		Name:          "feature/dev",
	}
	clusterType := "dcos"
	name := md5V(fmt.Sprintf("%d-%s-%s", r.ApplicationID, r.Workspace, r.Name))
	if clusterType == apistructs.EDAS {
		r.ID = 1111
		name = fmt.Sprintf("%s-%d", strings.ToLower(r.Workspace), r.ID)
	}
	fmt.Println(name)
}

func TestFnvV(t *testing.T) {
	s := "5-DEV-Srm"
	str1 := fnvV(s)
	str2 := fnvV(s)
	assert.Equal(t, str1, str2)
	assert.Equal(t, 10, len(str1))
}
