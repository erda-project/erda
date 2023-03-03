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
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/msp/instance/db"
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
	ins.Options = `{"MYSQL_ROOT_PASSWORD":"444c6564e96b420c8f2cf933838f0700"}`
	var sgStr = `{"created_time":1680161481,"last_modified_time":1680161481,"executor":"MARATHONFORLOCALCLUSTER","clusterName":"local-cluster","force":true,"scheduleInfo":{"Likes":null,"UnLikes":null,"LikePrefixs":null,"UnLikePrefixs":null,"ExclusiveLikes":null,"InclusiveLikes":null,"Flag":false,"HostUnique":false,"HostUniqueInfo":null,"SpecificHost":null,"IsPlatform":false,"IsUnLocked":false,"Location":null},"scheduleInfo2":{"HasHostUnique":false,"HostUnique":null,"SpecificHost":null,"IsPlatform":false,"IsDaemonset":false,"IsUnLocked":false,"Location":null,"HasOrg":false,"Org":"","HasWorkSpace":false,"WorkSpaces":null,"Job":false,"PreferJob":false,"Pack":false,"PreferPack":false,"Stateful":false,"PreferStateful":false,"Stateless":false,"PreferStateless":false,"BigData":false,"BigDataLabels":null,"HasProject":false,"Project":""},"name":"qa4bb396ca4b3474ca1d953e6f98aa7d4","namespace":"addon-mysql","labels":{"ADDON_GROUPS":"2","ADDON_ID":"qa4bb396ca4b3474ca1d953e6f98aa7d4","ADDON_TYPE":"qa4bb396ca4b3474ca1d953e6f98aa7d4","DICE_ADDON":"qa4bb396ca4b3474ca1d953e6f98aa7d4","DICE_ADDON_TYPE":"mysql","LOCATION-CLUSTER-SERVICE":"","PASSWORD":"444c6564e96b420c8f2cf933838f0700","SERVICE_TYPE":"ADDONS","USE_OPERATOR":"mysql"},"services":[{"name":"mysql","namespace":"group-addon-mysql--qa4bb396ca4b3474ca1d953e6f98aa7d4","image":"registry.erda.cloud/erda-addons/mylet:v8.0","image_username":"","image_password":"","Ports":[{"port":3306}],"vip":"mysql-qa4bb396ca-write.group-addon-mysql--qa4bb396ca4b3474ca1d953e6f98aa7d4.svc.cluster.local","scale":2,"resources":{"cpu":1,"mem":2048,"max_cpu":2,"emptydir_size":0,"ephemeral_storage_size":0},"env":{"ADDON_GROUPS":"1","ADDON_GROUP_ID":"mysql","ADDON_ID":"qa4bb396ca4b3474ca1d953e6f98aa7d4","ADDON_NODE_ID":"ae72b1255add94a63a244809f4fd64381","ADDON_TYPE":"qa4bb396ca4b3474ca1d953e6f98aa7d4","DICE_ADDON":"qa4bb396ca4b3474ca1d953e6f98aa7d4","DICE_ADDON_TYPE":"mysql","LOCATION-CLUSTER-SERVICE":"","MYSQL_ROOT_PASSWORD":"444c6564e96b420c8f2cf933838f0700","MYSQL_VERSION":"8.0","SERVICE_TYPE":"ADDONS"},"labels":{"ADDON_GROUP_ID":"mysql"},"selectors":null,"volumes":[{"volumeID":"0","volumePath":"","volumeTp":"nas","storage":20,"containerPath":"/var/backup/mysql","scVolume":{"type":"DICE-NAS","storageClassName":"dice-local-volume","size":20,"targetPath":"/var/backup/mysql","snapshot":{}}},{"volumeID":"1","volumePath":"","volumeTp":"nas","storage":20,"containerPath":"/var/lib/mysql","scVolume":{"type":"DICE-NAS","storageClassName":"dice-local-volume","size":20,"targetPath":"/var/lib/mysql","snapshot":{}}}],"healthCheck":null,"health_check":{"exec":{"cmd":"mysql -uroot -p444c6564e96b420c8f2cf933838f0700  -e 'select 1'"}},"traffic_security":{},"status":"Healthy","reason":"","unScheduledReasons":{},"desiredReplicas":0,"readyReplicas":0}],"serviceDiscoveryKind":"","projectNamespace":"","status":"Healthy","reason":"","unScheduledReasons":{},"desiredReplicas":0,"readyReplicas":0}`
	var sg apistructs.ServiceGroup
	if err := json.Unmarshal([]byte(sgStr), &sg); err != nil {
		t.Errorf("failure to yaml.Unmarshal: %v", err)
	}
	t.Logf("%+v", sg)
	dtoMap := ParseResp2MySQLDtoMap(ins, &sg)
	data, err := json.MarshalIndent(dtoMap, "", "  ")
	if err != nil {
		t.Errorf("failure to json.MarshalIndent: %v", err)
	}
	t.Log(string(data))
}
