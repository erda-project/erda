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

package clickhousesource

import (
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"

	apm "github.com/erda-project/erda/internal/tools/monitor/apm/common"
	"github.com/erda-project/erda/pkg/ckhelper"
)

// metric catalog
const (
	HttpRecMircoIndexType = "http-rpc-mirco"
	MQDBCacheIndexType    = "mq-db-cache"
	ServiceNodeIndexType  = "service-node"
)

// node catalog
const (
	TargetServiceNode   = "TargetServiceNode"
	SourceServiceNode   = "SourceServiceNode"
	TargetAddonNode     = "TargetAddonNode"
	SourceAddonNode     = "SourceAddonNode"
	TargetComponentNode = "TargetComponentNode"
	TargetOtherNode     = "TargetOtherNode"
	SourceMQNode        = "SourceMQNode"
	TargetMQNode        = "TargetMQNode"
	TargetMQServiceNode = "TargetMQServiceNode"
	OtherNode           = "OtherNode"
)

const (
	GroupHttp        = "application_http_service"
	GroupRpc         = "application_rpc_service"
	GroupMicro       = "application_micro_service"
	GroupMq          = "application_mq_service"
	GroupDb          = "application_db_service"
	GroupCache       = "application_cache_service"
	GroupServiceNode = "service_node"
)

type NodeRelation struct {
	Source []*NodeType
	Target *NodeType
}

type NodeType struct {
	Type string
	// columns in group by
	GroupByField []string
	// columns in select
	ColumnFields []string
	// where condition
	Filter BoolQuery
	// aggregations
	Aggregation SumAggregationMap
}

type (
	SumAggregationMap map[string]exp.AliasedExpression
	BoolQuery         exp.LiteralExpression
)

var (
	TargetServiceNodeType   *NodeType
	SourceServiceNodeType   *NodeType
	TargetAddonNodeType     *NodeType
	SourceAddonNodeType     *NodeType
	TargetComponentNodeType *NodeType
	TargetOtherNodeType     *NodeType
	SourceMQNodeType        *NodeType
	TargetMQNodeType        *NodeType
	TargetMQServiceNodeType *NodeType
	OtherNodeType           *NodeType

	nodeAggregation        SumAggregationMap
	serviceNodeAggregation SumAggregationMap
)

var (
	NodeRelations map[string][]*NodeRelation
	Aggregations  map[string]AggregationCondition
)

type AggregationCondition []exp.SelectClauses

func SelectRelation(indexType string) (AggregationCondition, []*NodeRelation) {
	return Aggregations[indexType], NodeRelations[indexType]
}

// convert nodeRelations to aggregation expression
func toAggregation(nodeRelations []*NodeRelation) []exp.SelectClauses {
	ss := []exp.SelectClauses{}
	for _, relation := range nodeRelations {
		target := relation.Target
		if target == nil {
			continue
		}
		if len(relation.Source) == 0 {
			ss = append(ss, mergeNodeType(target, nil))
			continue
		}
		for _, source := range relation.Source {
			ss = append(ss, mergeNodeType(target, source))
		}
	}
	return ss
}

func mergeNodeType(target, source *NodeType) exp.SelectClauses {
	sc := exp.NewSelectClauses()
	// 针对target特殊处理
	if target != nil {
		sc = sc.SetSelect(exp.NewColumnListExpression(goqu.L("?", target.Type).As("target_type")))
		for _, g := range target.GroupByField {
			sc = sc.WhereAppend(goqu.L("has(tag_keys, ?) == 1", ckhelper.TrimTags(g)))
		}
	}

	// 针对source特殊处理
	if source != nil {
		sc = sc.SelectAppend(exp.NewColumnListExpression(goqu.L("?", source.Type).As("source_type")))
	}

	cmap, gmap, aggmap := map[string]struct{}{}, map[string]struct{}{}, map[string]struct{}{}
	for _, item := range []*NodeType{target, source} {
		if item == nil {
			continue
		}
		for k, v := range item.Aggregation {
			if _, ok := aggmap[k]; ok {
				continue
			}
			aggmap[k] = struct{}{}
			sc = sc.SelectAppend(exp.NewColumnListExpression(v))
		}
		// group
		for _, g := range item.GroupByField {
			if _, ok := gmap[g]; ok {
				continue
			}
			gmap[g] = struct{}{}
			sc = sc.GroupByAppend(exp.NewColumnListExpression(ckhelper.TrimTags(g)))

			// group column
			col := exp.NewColumnListExpression(ckhelper.FromTagsKey(g).As(ckhelper.TrimTags(g)))
			sc = sc.SelectAppend(col)
		}
		// filter
		sc = sc.WhereAppend(item.Filter)
		// column
		for _, c := range item.ColumnFields {
			if _, ok := cmap[c]; ok {
				continue
			}
			// deduplicated
			if _, ok := gmap[c]; ok {
				continue
			}
			cmap[c] = struct{}{}
			col := exp.NewColumnListExpression(goqu.L("argMax(tag_values[indexOf(tag_keys, ?)], timestamp)", ckhelper.TrimTags(c)).As(ckhelper.TrimTags(c)))

			sc = sc.SelectAppend(col)
		}
	}

	return sc
}

func init() {
	serviceNodeAggregation = map[string]exp.AliasedExpression{
		apm.FieldsCountSum:  goqu.SUM(ckhelper.FromFieldNumberKey(apm.FieldsCountSum)).As(ckhelper.TrimFields(apm.FieldsCountSum)),
		apm.FieldElapsedSum: goqu.SUM(ckhelper.FromFieldNumberKey(apm.FieldElapsedSum)).As(ckhelper.TrimFields(apm.FieldElapsedSum)),
		apm.FieldsErrorsSum: goqu.SUM(ckhelper.FromFieldNumberKey(apm.FieldsErrorsSum)).As(ckhelper.TrimFields(apm.FieldsErrorsSum)),
	}
	nodeAggregation = map[string]exp.AliasedExpression{
		apm.FieldsCountSum:  goqu.SUM(ckhelper.FromFieldNumberKey(apm.FieldsCountSum)).As(ckhelper.TrimFields(apm.FieldsCountSum)),
		apm.FieldElapsedSum: goqu.SUM(ckhelper.FromFieldNumberKey(apm.FieldElapsedSum)).As(ckhelper.TrimFields(apm.FieldElapsedSum)),
	}

	TargetServiceNodeType = &NodeType{
		Type:         TargetServiceNode,
		GroupByField: []string{apm.TagsTargetServiceId, apm.TagsTargetServiceName},
		ColumnFields: []string{apm.TagsTargetApplicationId, apm.TagsTargetRuntimeName, apm.TagsTargetServiceName, apm.TagsTargetServiceId, apm.TagsTargetApplicationName, apm.TagsTargetRuntimeId},
		Filter:       goqu.L(fmt.Sprintf("has(tag_keys, '%s') == 0", ckhelper.TrimTags(apm.TagsTargetAddonType))),
		Aggregation:  serviceNodeAggregation,
	}
	SourceServiceNodeType = &NodeType{
		Type:         SourceServiceNode,
		GroupByField: []string{apm.TagsSourceServiceId, apm.TagsSourceServiceName},
		ColumnFields: []string{apm.TagsSourceApplicationName, apm.TagsSourceApplicationId, apm.TagsSourceServiceName, apm.TagsSourceServiceId, apm.TagsSourceRuntimeName, apm.TagsSourceRuntimeId},
		Filter:       goqu.L(fmt.Sprintf("has(tag_keys, '%s') == 0", ckhelper.TrimTags(apm.TagsSourceAddonType))),
		Aggregation:  serviceNodeAggregation,
	}
	TargetAddonNodeType = &NodeType{
		Type:         TargetAddonNode,
		GroupByField: []string{apm.TagsTargetAddonType, apm.TagsTargetAddonId},
		ColumnFields: []string{apm.TagsTargetAddonType, apm.TagsTargetAddonId, apm.TagsTargetAddonGroup, apm.TagsComponent},
		Filter:       goqu.L(fmt.Sprintf("has(tag_keys, '%s') == 1", ckhelper.TrimTags(apm.TagsTargetAddonType))),
		Aggregation:  nodeAggregation,
	}
	SourceAddonNodeType = &NodeType{
		Type:         SourceAddonNode,
		GroupByField: []string{apm.TagsSourceAddonType, apm.TagsSourceAddonId},
		ColumnFields: []string{apm.TagsSourceAddonType, apm.TagsSourceAddonId, apm.TagsSourceAddonGroup},
		Filter:       goqu.L(fmt.Sprintf("has(tag_keys, '%s') == 1", ckhelper.TrimTags(apm.TagsSourceAddonType))),
		Aggregation:  nodeAggregation,
	}
	TargetComponentNodeType = &NodeType{
		Type:         TargetComponentNode,
		GroupByField: []string{apm.TagsDBHost, apm.TagsDBSystem},
		ColumnFields: []string{apm.TagsComponent, apm.TagsHost, apm.TagsTargetAddonGroup, apm.TagsDBSystem, apm.TagsDBHost},
		Filter:       goqu.L(fmt.Sprintf("has(tag_keys, '%s') == 0", ckhelper.TrimTags(apm.TagsTargetAddonType))),
		Aggregation:  nodeAggregation,
	}
	TargetOtherNodeType = &NodeType{
		Type:         TargetOtherNode,
		GroupByField: []string{apm.TagsHttpUrl, apm.TagsPeerServiceScope},
		ColumnFields: []string{apm.TagsPeerServiceScope, apm.TagsHttpUrl, apm.TagsComponent},
		Filter:       goqu.L(fmt.Sprintf("has(tag_keys, '%s') == 0", ckhelper.TrimTags(apm.TagsTargetAddonType))),
		Aggregation:  nodeAggregation,
	}
	SourceMQNodeType = &NodeType{
		Type:         SourceMQNode,
		GroupByField: []string{apm.TagsComponent, apm.TagsPeerAddress},
		ColumnFields: []string{apm.TagsComponent, apm.TagsHost, apm.TagsPeerAddress},
		Filter:       goqu.L(fmt.Sprintf("metric_group == '%s' AND has(tag_keys, '%s') == 0", GroupMq, ckhelper.TrimTags(apm.TagsTargetAddonType))),
		Aggregation:  nodeAggregation,
	}
	TargetMQNodeType = &NodeType{
		Type:         TargetMQNode,
		GroupByField: []string{apm.TagsComponent, apm.TagsPeerAddress},
		ColumnFields: []string{apm.TagsComponent, apm.TagsHost, apm.TagsPeerAddress},
		Filter:       goqu.L(fmt.Sprintf("metric_group == '%s' AND has(tag_keys, '%s') == 0", GroupMq, ckhelper.TrimTags(apm.TagsTargetAddonType))),
		Aggregation:  nodeAggregation,
	}
	TargetMQServiceNodeType = &NodeType{
		Type:         TargetMQServiceNode,
		GroupByField: []string{apm.TagsTargetServiceId, apm.TagsTargetServiceName},
		ColumnFields: []string{apm.TagsTargetApplicationId, apm.TagsTargetRuntimeName, apm.TagsTargetServiceName, apm.TagsTargetServiceId, apm.TagsTargetApplicationName, apm.TagsTargetRuntimeId},
		Filter:       goqu.L(fmt.Sprintf("has(tag_keys, '%s') == 0", ckhelper.TrimTags(apm.TagsTargetAddonType))),
	}
	OtherNodeType = &NodeType{
		Type:         OtherNode,
		GroupByField: []string{apm.TagsServiceId, apm.TagsServiceName},
		ColumnFields: []string{apm.TagsApplicationId, apm.TagsRuntimeName, apm.TagsServiceName, apm.TagsServiceId, apm.TagsApplicationName, apm.TagsRuntimeId, apm.TagsComponent},
		Filter:       goqu.L(fmt.Sprintf("has(tag_keys, '%s') == 1", ckhelper.TrimTags(apm.TagsServiceId))),
	}

	NodeRelations = map[string][]*NodeRelation{
		HttpRecMircoIndexType: {
			// Topology Relation (Addon: ApiGateway...)
			// 1.SourceService -> TargetService 2.SourceAddon -> TargetService
			// SourceService -> TargetAddon
			// SourceService -> TargetOther
			{Source: []*NodeType{SourceServiceNodeType, SourceAddonNodeType}, Target: TargetServiceNodeType},
			{Source: []*NodeType{SourceServiceNodeType}, Target: TargetAddonNodeType},
			{Source: []*NodeType{SourceServiceNodeType}, Target: TargetOtherNodeType},
		},
		MQDBCacheIndexType: {
			// Topology Relation (Component: Mysql Redis MQ)
			// SourceMQService  -> TargetMQService (consumer)
			// SourceService  -> TargetMQ (producer)
			// SourceService    -> TargetComponent
			{Source: []*NodeType{SourceMQNodeType}, Target: TargetMQServiceNodeType},
			{Source: []*NodeType{SourceServiceNodeType}, Target: TargetMQNodeType},
			{Source: []*NodeType{SourceServiceNodeType}, Target: TargetComponentNodeType},
		},
		ServiceNodeIndexType: {
			// Topology Relation
			// OtherNode
			{Target: OtherNodeType},
		},
	}

	// mapping between type and aggregation condition
	Aggregations = map[string]AggregationCondition{
		HttpRecMircoIndexType: toAggregation(NodeRelations[HttpRecMircoIndexType]),
		MQDBCacheIndexType:    toAggregation(NodeRelations[MQDBCacheIndexType]),
		ServiceNodeIndexType:  toAggregation(NodeRelations[ServiceNodeIndexType]),
	}
}
