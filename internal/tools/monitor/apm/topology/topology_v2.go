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
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/doug-martin/goqu/v9"

	"github.com/erda-project/erda-infra/providers/i18n"
	apm "github.com/erda-project/erda/internal/tools/monitor/apm/common"
	"github.com/erda-project/erda/internal/tools/monitor/apm/topology/clickhousesource"
	"github.com/erda-project/erda/pkg/ckhelper"
	pkgmath "github.com/erda-project/erda/pkg/math"
)

var (
	typeMetricGroupMap = map[string][]string{
		HttpRecMircoIndexType: {clickhousesource.GroupHttp, clickhousesource.GroupRpc, clickhousesource.GroupMicro},
		MQDBCacheIndexType:    {clickhousesource.GroupMq, clickhousesource.GroupDb, clickhousesource.GroupCache},
		ServiceNodeIndexType:  {clickhousesource.GroupServiceNode},
	}
)

type chNodeEdge struct {
	SourceID              uint32 `ch:"source_id"`
	SourceType            string `ch:"source_type"`
	SourceApplicationId   string `ch:"source_application_id"`
	SourceApplicationName string `ch:"source_application_name"`
	SourceRuntimeId       string `ch:"source_runtime_id"`
	SourceRuntimeName     string `ch:"source_runtime_name"`
	SourceServiceId       string `ch:"source_service_id"`
	SourceServiceName     string `ch:"source_service_name"`
	SourceAddonType       string `ch:"source_addon_type"`
	SourceAddonId         string `ch:"source_addon_id"`
	SourceAddonGroup      string `ch:"source_addon_group"`

	TargetID              uint32 `ch:"target_id"`
	TargetType            string `ch:"target_type"`
	TargetApplicationId   string `ch:"target_application_id"`
	TargetApplicationName string `ch:"target_application_name"`
	TargetRuntimeId       string `ch:"target_runtime_id"`
	TargetRuntimeName     string `ch:"target_runtime_name"`
	TargetServiceId       string `ch:"target_service_id"`
	TargetServiceName     string `ch:"target_service_name"`
	TargetAddonType       string `ch:"target_addon_type"`
	TargetAddonId         string `ch:"target_addon_id"`
	TargetAddonGroup      string `ch:"target_addon_group"`

	AddonId          string `ch:"addon_id"`
	AddonType        string `ch:"addon_type"`
	DBHost           string `ch:"db_host"`
	DBSystem         string `ch:"db_system"`
	Component        string `ch:"component"`
	Host             string `ch:"host"`
	HttpUrl          string `ch:"http_url"`
	PeerServiceScope string `ch:"peer_service_scope"`
	PeerAddress      string `ch:"peer_address"`

	// service-node related
	ServiceID       string `ch:"service_id"`
	ServiceName     string `ch:"service_name"`
	ApplicationId   string `ch:"application_id"`
	ApplicationName string `ch:"application_name"`
	RuntimeId       string `ch:"runtime_id"`
	RuntimeName     string `ch:"runtime_name"`

	ErrorsSum  float64 `ch:"errors_sum"`
	CountSum   float64 `ch:"count_sum"`
	ElapsedSum float64 `ch:"elapsed_sum"`
}

func (cn *chNodeEdge) ToTopologyNodeRelation() *TopologyNodeRelation {
	return &TopologyNodeRelation{
		Tags: Tag{
			Component:             cn.Component,
			DBType:                "",
			DBSystem:              cn.DBSystem,
			Host:                  cn.Host,
			HttpUrl:               cn.HttpUrl,
			PeerServiceScope:      cn.PeerServiceScope,
			PeerAddress:           cn.PeerAddress,
			PeerService:           "",
			DBHost:                cn.DBHost,
			SourceProjectId:       "",
			SourceProjectName:     "",
			SourceWorkspace:       "",
			SourceTerminusKey:     "",
			SourceApplicationId:   cn.SourceApplicationId,
			SourceApplicationName: cn.SourceApplicationName,
			SourceRuntimeId:       cn.SourceRuntimeId,
			SourceRuntimeName:     cn.SourceRuntimeName,
			SourceServiceName:     cn.SourceServiceName,
			SourceServiceId:       cn.SourceServiceId,
			SourceAddonID:         cn.SourceAddonId,
			SourceAddonType:       cn.SourceAddonType,
			TargetInstanceId:      "",
			TargetProjectId:       "",
			TargetProjectName:     "",
			TargetWorkspace:       "",
			TargetTerminusKey:     "",
			TargetApplicationId:   cn.TargetApplicationId,
			TargetApplicationName: cn.TargetApplicationName,
			TargetRuntimeId:       cn.TargetRuntimeId,
			TargetRuntimeName:     cn.TargetRuntimeName,
			TargetServiceName:     cn.TargetServiceName,
			TargetServiceId:       cn.TargetServiceId,
			TargetAddonID:         cn.TargetAddonId,
			TargetAddonType:       cn.TargetAddonType,
			TerminusKey:           "",
			ProjectId:             "",
			ProjectName:           "",
			Workspace:             "",
			ApplicationId:         cn.ApplicationId,
			ApplicationName:       cn.ApplicationName,
			RuntimeId:             cn.RuntimeId,
			RuntimeName:           cn.RuntimeName,
			ServiceName:           cn.ServiceName,
			ServiceId:             cn.ServiceID,
			ServiceInstanceId:     "",
			ServiceIp:             "",
			Type:                  cn.TargetType,
		},
		Fields: Field{},
		Metric: &Metric{
			Count:     int64(cn.CountSum),
			HttpError: int64(cn.ErrorsSum),
			Duration:  cn.ElapsedSum,
		},
	}
}

type graphTopo struct {
	adj   map[string]map[string]struct{}
	nodes map[string]Node
}

type relation struct {
	source, target Node
	stats          Metric
}

func (gt *graphTopo) addDirectEdge(rel *relation) {
	gt.addNode(rel.target, rel.stats)
	gt.addNode(rel.source, rel.stats)

	if rel.target.Valid() {
		if _, ok := gt.adj[rel.target.Id]; !ok {
			gt.adj[rel.target.Id] = map[string]struct{}{}
		}
		if rel.source.Valid() {
			gt.adj[rel.target.Id][rel.source.Id] = struct{}{}
		}
	}
}

func (gt *graphTopo) addNode(node Node, m Metric) {
	if !node.Valid() {
		return
	}
	if n, ok := gt.nodes[node.Id]; ok {
		if n.RuntimeId == "" {
			n.RuntimeId = node.RuntimeId
		}
		n.Metric.Count += m.Count
		n.Metric.HttpError += m.HttpError
		n.Metric.Duration += m.Duration
		n.Metric.RT += m.RT
		if n.RuntimeId == "" {
			n.RuntimeId = node.RuntimeId
		}
	} else {
		node.Metric = &m
		gt.nodes[node.Id] = node
	}
}

func (gt *graphTopo) toNodes() []*Node {
	nodes := make([]Node, 0, len(gt.adj))
	for id, n := range gt.nodes {
		adj := gt.adj[id]
		for pid := range adj {
			pnode := gt.nodes[pid]
			if pnode.Id == "" {
				fmt.Println()
			}
			n.Parents = append(n.Parents, &Node{
				Id:              pnode.Id,
				Name:            pnode.Name,
				Type:            pnode.Type,
				TypeDisplay:     pnode.TypeDisplay,
				AddonId:         pnode.AddonId,
				AddonType:       pnode.AddonType,
				ApplicationId:   pnode.ApplicationId,
				ApplicationName: pnode.ApplicationName,
				RuntimeId:       pnode.RuntimeId,
				RuntimeName:     pnode.RuntimeName,
				ServiceId:       pnode.ServiceId,
				ServiceName:     pnode.ServiceName,
				DashboardId:     pnode.DashboardId,
				Metric:          pnode.Metric,
				Parents:         []*Node{},
			})
		}
		nodes = append(nodes, n)
	}

	pnodes := make([]*Node, len(nodes))
	for i := 0; i < len(pnodes); i++ {
		pnodes[i] = &nodes[i]
	}
	return pnodes
}

func (topology *provider) GetTopologyV2(orgName string, lang i18n.LanguageCodes, param Vo) ([]*Node, error) {
	timeRange := (param.EndTime - param.StartTime) / 1e3 // second
	ctx := context.Background()
	tagInfo := parserTag(param)
	tg := &graphTopo{adj: map[string]map[string]struct{}{}, nodes: map[string]Node{}}
	table, _ := topology.Loader.GetSearchTable(orgName)
	var allNodeRelations []*relation
	for key, typeIndices := range typeMetricGroupMap {
		aggregationConditions, _ := clickhousesource.SelectRelation(key)
		for _, agg := range aggregationConditions {
			sd := goqu.From(table).Where(
				goqu.C("org_name").In(orgName, ""),
				goqu.C("metric_group").In(typeIndices),
			)
			sd = queryConditionsV2(sd, param, tagInfo, key)
			sd = sd.Select(agg.Select()).Where(agg.Where()).GroupBy(agg.GroupBy())

			sqlstr, _, err := sd.ToSQL()
			if err != nil {
				topology.Log.Errorf("SQL: %s", sqlstr)
				return nil, fmt.Errorf("failed tosql: %w", err)
			}
			// debug
			if param.Debug {
				topology.Log.Infof("key: %q, SQL: %s", key, sqlstr)
			}

			rows, err := topology.Clickhouse.Client().Query(ctx, sqlstr)
			if err != nil {
				return nil, fmt.Errorf("do query: %w", err)
			}

			allNodeRelations = append(allNodeRelations, topology.parseToTypologyNodeV2(lang, rows, tg)...)
		}
	}
	handleTargetOtherNodesByHttpUrl(allNodeRelations, topology.Cfg.TargetOtherNodeOptions)
	for _, rel := range allNodeRelations {
		tg.addDirectEdge(rel)
	}

	nodes := tg.toNodes()
	for _, node := range nodes {
		if node.Metric.Count != 0 { // by zero
			node.Metric.RT = pkgmath.DecimalPlacesWithDigitsNumber(node.Metric.Duration/float64(node.Metric.Count)/1e6, 2)
			node.Metric.ErrorRate = pkgmath.DecimalPlacesWithDigitsNumber(float64(node.Metric.HttpError)/float64(node.Metric.Count)*100, 2)
		}
		node.Metric.RPS = pkgmath.DecimalPlacesWithDigitsNumber(float64(node.Metric.Count)/float64(timeRange), 2)
	}

	if tagInfo.ServiceId != "" {
		return filterNodesByServiceId(tagInfo.ServiceId, nodes), nil
	}
	return nodes, nil
}

func (topology *provider) parseToTypologyNodeV2(lang i18n.LanguageCodes, rows driver.Rows, tg *graphTopo) []*relation {
	allRelations := make([]*relation, 0)
	defer rows.Close()
	for rows.Next() {
		cnode := chNodeEdge{}
		err := rows.ScanStruct(&cnode)
		if err != nil {
			topology.Log.Warnf("scan chNodeEdge failed: %s", err)
			continue
		}

		edge := cnode.ToTopologyNodeRelation()
		targetNode, sourceNode := columnsParser(cnode.TargetType, edge), columnsParser(cnode.SourceType, edge)
		if cnode.TargetType == TargetOtherNode && targetNode.Type == TypeInternal {
			continue
		}
		targetNode.TypeDisplay = topology.t.Text(lang, strings.ToLower(targetNode.Type))
		sourceNode.TypeDisplay = topology.t.Text(lang, strings.ToLower(sourceNode.Type))
		rel := &relation{target: *targetNode, source: *sourceNode, stats: *edge.Metric}
		allRelations = append(allRelations, rel)
	}
	return allRelations
}

func handleTargetOtherNodesByHttpUrl(allRels []*relation, opts targetOtherNodeOptions) {
	// get other nodes number
	targetOtherNodes := make([]*Node, 0, len(allRels)*2)
	httpUrlCount := map[string]struct{}{}
	httpUrlWithoutQueryParamsCount := map[string]struct{}{}
	for _, rel := range allRels {
		nodes := []*Node{&rel.source, &rel.target}
		for _, node := range nodes {
			if node.Type == TypeExternal {
				// when node type is TargetOtherNode, node name is http url
				httpUrlCount[node.Name] = struct{}{}
				targetOtherNodes = append(targetOtherNodes, node)
			}
		}
	}
	if len(httpUrlCount) <= opts.MaxNum {
		return
	}
	// re calculate node id by name
	defer func() {
		for _, node := range targetOtherNodes {
			node.Id = calculateTargetOtherNodeId(node)
		}
	}()
	// ignore query params firstly
	for _, node := range targetOtherNodes {
		// http url remove query params
		node.Name = strings.SplitN(node.Name, "?", 2)[0]
		httpUrlWithoutQueryParamsCount[node.Name] = struct{}{}
		if len(httpUrlWithoutQueryParamsCount) > opts.MaxNum {
			// need to handle host, fast break
			break
		}
	}
	if len(httpUrlWithoutQueryParamsCount) <= opts.MaxNum {
		return
	}
	// only host
	for _, node := range targetOtherNodes {
		u, err := url.Parse(node.Name)
		if err == nil {
			node.Name = u.Host
		}
	}
}

func queryConditionsV2(sd *goqu.SelectDataset, params Vo, tagInfo *TagInfo, indexType string) *goqu.SelectDataset {
	where := sd
	where = where.Where(
		goqu.C("tenant_id").In(params.TerminusKey, ""),
		goqu.C("timestamp").Gte(ckhelper.FromTimestampMilli(params.StartTime)),
		goqu.C("timestamp").Lte(ckhelper.FromTimestampMilli(params.EndTime)),
	)
	// in-case some metric don't have tenant_id
	if ServiceNodeIndexType == indexType {
		where = where.Where(ckhelper.FromTagsKey(apm.TagsTerminusKey).Eq(params.TerminusKey))
	} else {
		where = where.Where(goqu.Or(
			ckhelper.FromTagsKey(apm.TagsTargetTerminusKey).Eq(params.TerminusKey),
			ckhelper.FromTagsKey(apm.TagsSourceTerminusKey).Eq(params.TerminusKey),
		))
	}

	//filter: RegisterCenter ConfigCenter NoticeCenter
	where = where.Where(
		ckhelper.FromTagsKey(apm.TagsComponent).NotIn("registerCenter", "configCenter", "noticeCenter"),
		ckhelper.FromTagsKey(apm.TagsTargetAddonType).NotIn("registerCenter", "configCenter", "noticeCenter"),
	)

	if tagInfo.ApplicationName != "" {
		where = where.Where(goqu.Or(
			ckhelper.FromTagsKey(apm.TagsApplicationName).Eq(tagInfo.ApplicationName),
			ckhelper.FromTagsKey(apm.TagsTargetApplicationName).Eq(tagInfo.ApplicationName),
			ckhelper.FromTagsKey(apm.TagsSourceApplicationName).Eq(tagInfo.ApplicationName),
		))
	}
	return where
}
