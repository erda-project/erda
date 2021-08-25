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

package topology

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/olivere/elastic"
	"github.com/stretchr/testify/assert"
)

func TestSpliceIndexByTime(t *testing.T) {
	startTimeMs := 1603148400544 // 2020-10-20 07:00:00
	endTimeMs := 1603753200543   // 2020-10-27 07:00:00

	indices := createTypologyIndices(int64(startTimeMs), int64(endTimeMs))
	for _, index := range indices {
		fmt.Println(index)
	}
}

func TestJsonStrToStruct(t *testing.T) {
	str := `{"name":"application_micro_service","timestamp":1603671150000000000,"tags":{"_meta":"true","_metric_scope":"micro_service","_metric_scope_id":"z341b9c025b914180877ad7dbb9d80d9f","cluster_name":"terminus-dev","host":"node-010000006205","host_ip":"10.0.6.205","org_name":"terminus","source_application_id":"4","source_application_name":"apm-demo","source_org_id":"1","source_project_id":"1","source_project_name":"test","source_runtime_id":"48","source_runtime_name":"feature/simple_with_nacos","source_service_id":"4_feature/simple_with_nacos_apm-demo-dubbo","source_service_instance_id":"fae63126-78f8-4ddd-9756-c9d363211e5f","source_service_name":"apm-demo-dubbo","source_terminus_key":"z341b9c025b914180877ad7dbb9d80d9f","source_workspace":"DEV","target_addon_id":"registerCenter","target_addon_type":"registerCenter"},"fields":{"elapsed_count":1,"elapsed_max":13945,"elapsed_mean":13945,"elapsed_min":13945,"elapsed_sum":13945},"@timestamp":1603671150000}`
	tnr := TopologyNodeRelation{}
	err := json.Unmarshal([]byte(str), &tnr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tnr)
}

func TestRegexp(t *testing.T) {
	re := regexp.MustCompile("^" + ServiceNodeIndex + "(.*)$")
	matchString := re.MatchString("spot-service_node-*-1603065600000")
	if !matchString {
		log.Fatal("not match")
	}
}

func TestCreateAggregation(t *testing.T) {

	NodeRelations := map[string][]*NodeRelation{}

	NodeRelations["mq-db-cache"] = []*NodeRelation{
		// Topology Relation (Component: Mysql Redis MQ)
		// SourceMQService  -> TargetMQService
		// SourceService    -> TargetComponent
		{Source: []*NodeType{SourceMQNodeType}, Target: TargetMQServiceNodeType},
		{Source: []*NodeType{SourceServiceNodeType}, Target: TargetComponentNodeType},
	}

	//aggregation := elastic.NewFilterAggregation()
	//allQuery := elastic.NewMatchAllQuery()
	aggregation := elastic.NewFilterAggregation().Filter(elastic.NewMatchAllQuery()) // 1级索引过滤
	for _, relation := range NodeRelations["http-rpc-mirco"] {
		// target
		if relation.Target != nil {
			uuid, _ := uuid.NewV4()
			childAggregation := elastic.NewFilterAggregation()
			aggregation.SubAggregation(uuid.String(), childAggregation)
			if relation.Target.Filter != nil {
				not := elastic.NewBoolQuery().MustNot(relation.Target.Filter)
				childAggregation.Filter(not)
			}
		}
	}

	source, _ := aggregation.Source()
	marshal, _ := json.Marshal(source)
	fmt.Println(string(marshal))

}

func TestToEsAggregation(t *testing.T) {
	NodeRelations := map[string][]*NodeRelation{}

	NodeRelations["mq-db-cache"] = []*NodeRelation{
		// Topology Relation (Component: Mysql Redis MQ)
		// SourceMQService  -> TargetMQService
		// SourceService    -> TargetComponent
		//{Source: []*NodeType{SourceMQNodeType}, Target: TargetMQServiceNodeType},
		{Source: []*NodeType{SourceServiceNodeType}, Target: TargetComponentNodeType},
	}
}

func TestFloat64ToString(t *testing.T) {
	f := 12.33
	float := strconv.FormatFloat(f, 'f', 2, 64)
	fmt.Println(float)
}

func Test_filterInstance(t *testing.T) {
	type args struct {
		instanceList          []*InstanceInfo
		instanceListForStatus []*InstanceInfo
	}
	var instanceListCase []*InstanceInfo
	var instanceListForStatusCase []*InstanceInfo
	for i := 0; i < 100; i++ {
		info := InstanceInfo{
			Id:     fmt.Sprintf("instance-%d", i),
			Ip:     "127.0.0.1",
			Status: false,
		}
		instanceListCase = append(instanceListCase, &info)
		infoForStatus := InstanceInfo{
			Id:     fmt.Sprintf("instance-%d", i),
			Ip:     "127.0.0.1",
			Status: false,
		}
		if i%2 == 0 {
			infoForStatus.Status = true
		}
		instanceListForStatusCase = append(instanceListForStatusCase, &infoForStatus)
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "case1", args: args{instanceList: instanceListCase, instanceListForStatus: nil}},
		{name: "case2", args: args{instanceList: nil, instanceListForStatus: instanceListForStatusCase}},
		{name: "case3", args: args{instanceList: instanceListCase, instanceListForStatus: instanceListForStatusCase}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterInstance(tt.args.instanceList, tt.args.instanceListForStatus)
			count := 0
			for _, info := range tt.args.instanceList {
				if info.Status == true {
					count++
				}
			}
			if tt.name == "case1" {
				assert.Equal(t, 0, count)
			}
			if tt.name == "case2" {
				assert.Equal(t, 0, count)
			}
			if tt.name == "case3" {
				assert.Equal(t, 50, count)
			}
		})
	}
}

func Test_getDashboardId(t *testing.T) {

	type args struct {
		nodeType string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: TypeService, args: args{TypeService}, want: "topology_node_service"},
		{name: TypeGateway, args: args{TypeGateway}, want: "topology_node_gateway"},
		{name: TypeMysql, args: args{TypeMysql}, want: "topology_node_db"},
		{name: TypeRedis, args: args{TypeRedis}, want: "topology_node_cache"},
		{name: TypeRocketMQ, args: args{TypeRocketMQ}, want: "topology_node_mq"},
		{name: TypeHttp, args: args{TypeHttp}, want: "topology_node_other"},
		{name: JavaProcessType, args: args{JavaProcessType}, want: "process_analysis_java"},
		{name: NodeJsProcessType, args: args{NodeJsProcessType}, want: "process_analysis_nodejs"},
		{name: "not", args: args{"not"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDashboardId(tt.args.nodeType); got != tt.want {
				t.Errorf("getDashboardId() = %v, want %v", got, tt.want)
			}
		})
	}
}
