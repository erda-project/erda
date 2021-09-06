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

package topology

import (
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/core/monitor/metric/query/metricq"
	"github.com/erda-project/erda/modules/monitor/common/db"
	"github.com/gocql/gocql"
	"log"
	"reflect"
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

func Test_provider_handleResult(t *testing.T) {
	itemResult := make(map[string]interface{})
	itemResult["operation"] = "test-topic"
	itemResult["type"] = "consumer"
	itemResult["component"] = "mq"
	itemResult["host"] = "xxx:8080"
	itemResult["call_count"] = 10
	itemResult["avg_elapsed"] = 1000
	itemResult["slow_elapsed_count"] = 2
	type fields struct {
		Cfg              *config
		Log              logs.Logger
		db               *db.DB
		es               *elastic.Client
		ctx              servicehub.Context
		metricq          metricq.Queryer
		t                i18n.Translator
		cassandraSession *gocql.Session
	}
	type args struct {
		r         []interface{}
		slowCount int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]interface{}
	}{
		{"case", fields{
			Cfg:              nil,
			Log:              nil,
			db:               nil,
			es:               nil,
			ctx:              nil,
			metricq:          nil,
			t:                nil,
			cassandraSession: nil,
		}, args{
			r: []interface{}{
				"test-topic",
				"consumer",
				"mq",
				"xxx:8080",
				10,
				1000,
			},
			slowCount: 2,
		}, itemResult,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topology := &provider{
				Cfg:              tt.fields.Cfg,
				Log:              tt.fields.Log,
				db:               tt.fields.db,
				es:               tt.fields.es,
				ctx:              tt.fields.ctx,
				metricq:          tt.fields.metricq,
				t:                tt.fields.t,
				cassandraSession: tt.fields.cassandraSession,
			}
			if got := topology.handleResult(tt.args.r, tt.args.slowCount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleResult() = %v, want %v", got, tt.want)
			}
		})
	}
}
