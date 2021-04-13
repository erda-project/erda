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
	"github.com/gofrs/uuid"
	"github.com/olivere/elastic"
	"log"
	"regexp"
	"strconv"
	"testing"
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
