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

package mysql

import (
	"encoding/json"
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/msp/instance/db"
	"github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers"
	"github.com/erda-project/erda/internal/apps/msp/resource/utils"
)

func TestTryReadFile(t *testing.T) {
	p := &provider{}
	sql, err := p.tryReadFile("file://tmc/nacos.tar.gz")
	if err != nil {
		t.Errorf("with exists filepath, should not return error")
	}
	if len(sql) == 0 {
		t.Errorf("with exists filepath, should not return empty content")
	}
}

func TestParseResp2MySQLDtoMap(t *testing.T) {
	var ins = new(db.Instance)
	ins.Options = `{"MYSQL_ROOT_PASSWORD":"this-is-a-mocked-password"}`
	var sgStr = `{"created_time":1680161481,"last_modified_time":1680161481,"executor":"this-is-a-mocked-executor","clusterName":"local-cluster","force":true,"scheduleInfo":{"Likes":null,"UnLikes":null,"LikePrefixs":null,"UnLikePrefixs":null,"ExclusiveLikes":null,"InclusiveLikes":null,"Flag":false,"HostUnique":false,"HostUniqueInfo":null,"SpecificHost":null,"IsPlatform":false,"IsUnLocked":false,"Location":null},"scheduleInfo2":{"HasHostUnique":false,"HostUnique":null,"SpecificHost":null,"IsPlatform":false,"IsDaemonset":false,"IsUnLocked":false,"Location":null,"HasOrg":false,"Org":"","HasWorkSpace":false,"WorkSpaces":null,"Job":false,"PreferJob":false,"Pack":false,"PreferPack":false,"Stateful":false,"PreferStateful":false,"Stateless":false,"PreferStateless":false,"BigData":false,"BigDataLabels":null,"HasProject":false,"Project":""},"name":"this-is-a-mocked-name","namespace":"addon-mysql","labels":{"ADDON_GROUPS":"2","ADDON_ID":"this-is-a-mocked-addon-id","ADDON_TYPE":"this-is-a-mocked-addon-type","DICE_ADDON":"this-is-a-mocked-dice-addon","DICE_ADDON_TYPE":"mysql","LOCATION-CLUSTER-SERVICE":"","PASSWORD":"this-is-a-mocked-password","SERVICE_TYPE":"ADDONS","USE_OPERATOR":"mysql"},"services":[{"name":"mysql","namespace":"group-addon-mysql--a-mocked-namespace","image":"this-is-a-mocked-image","image_username":"","image_password":"","Ports":[{"port":3306}],"vip":"this-is-a-mocked-vip","scale":2,"resources":{"cpu":1,"mem":2048,"max_cpu":2,"emptydir_size":0,"ephemeral_storage_size":0},"env":{"ADDON_GROUPS":"1","ADDON_GROUP_ID":"mysql","ADDON_ID":"this-is-a-mocked-addon-id","ADDON_NODE_ID":"this-is-a-mocked-addon-node-id","ADDON_TYPE":"this-is-a-mocked-addon-type","DICE_ADDON":"this-is-a-mocked-dice-addon","DICE_ADDON_TYPE":"mysql","LOCATION-CLUSTER-SERVICE":"","MYSQL_ROOT_PASSWORD":"this-is-a-mocked-password","MYSQL_VERSION":"8.0","SERVICE_TYPE":"ADDONS"},"labels":{"ADDON_GROUP_ID":"mysql"},"selectors":null,"volumes":[{"volumeID":"0","volumePath":"","volumeTp":"nas","storage":20,"containerPath":"/var/backup/mysql","scVolume":{"type":"DICE-NAS","storageClassName":"dice-local-volume","size":20,"targetPath":"/var/backup/mysql","snapshot":{}}},{"volumeID":"1","volumePath":"","volumeTp":"nas","storage":20,"containerPath":"/var/lib/mysql","scVolume":{"type":"DICE-NAS","storageClassName":"dice-local-volume","size":20,"targetPath":"/var/lib/mysql","snapshot":{}}}],"healthCheck":null,"health_check":{"exec":{"cmd":"mysql -uroot -p'this-is-a-mocked-password'  -e 'select 1'"}},"traffic_security":{},"status":"Healthy","reason":"","unScheduledReasons":{},"desiredReplicas":0,"readyReplicas":0}],"serviceDiscoveryKind":"","projectNamespace":"","status":"Healthy","reason":"","unScheduledReasons":{},"desiredReplicas":0,"readyReplicas":0}`
	var sg apistructs.ServiceGroup
	if err := json.Unmarshal([]byte(sgStr), &sg); err != nil {
		t.Errorf("failed to yaml.Unmarshal: %v", err)
	}
	t.Logf("%+v", sg)
	dtoMap := ParseResp2MySQLDtoMap(ins, &sg)
	data, err := json.MarshalIndent(dtoMap, "", "  ")
	if err != nil {
		t.Errorf("failed to json.MarshalIndent: %v", err)
	}
	t.Log(string(data))
	mysql, ok := dtoMap["mysql"]
	if !ok || mysql == nil {
		t.Fatal("failed to parse")
	}
	if len(mysql.Options) == 0 {
		t.Fatal("failed to parse")
	}
	if mysql.Options["MYSQL_ROOT_PASSWORD"] != "this-is-a-mocked-password" {
		t.Fatal("failed to parse")
	}
}

func TestCheckIfNeedTmcInstance(t *testing.T) {
	p := &provider{
		DefaultDeployHandler: &handlers.DefaultDeployHandler{
			InstanceDb: &db.InstanceDB{},
		},
	}
	req := &handlers.ResourceDeployRequest{
		Engine: "mysql",
		Uuid:   utils.GetRandomId(),
		Az:     "test-cluster",
	}
	info := &handlers.ResourceInfo{
		Tmc: &db.Tmc{},
		TmcVersion: &db.TmcVersion{
			Engine: "mysql",
		},
	}

	monkey.PatchInstanceMethod(reflect.TypeOf(p.InstanceDb), "First", func(DB *db.InstanceDB, where map[string]any) (*db.Instance, bool, error) {
		return &db.Instance{
			Engine:    "mysql",
			Version:   "9.0",
			ReleaseID: "i am release id!",
			Status:    "RUNNING",
			Az:        "test-cluster",
			Config:    "",
		}, false, nil
	})

	_, _, _ = p.CheckIfNeedTmcInstance(req, info)
}
