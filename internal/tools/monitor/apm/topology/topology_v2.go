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
			ApplicationId:         "",
			ApplicationName:       "",
			RuntimeId:             "",
			RuntimeName:           "",
			ServiceName:           "",
			ServiceId:             "",
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
	nodes map[string]*Node
}

func (gt *graphTopo) toNodes() []*Node {
	nodes := make([]*Node, 0, len(gt.adj))
	for id, set := range gt.adj {
		n := gt.nodes[id]
		for pid, _ := range set {
			pnode := gt.nodes[pid]
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
	return nodes
}

func (topology *provider) GetTopologyV2(orgName string, lang i18n.LanguageCodes, param Vo) ([]*Node, error) {
	timeRange := (param.EndTime - param.StartTime) / 1e3 // second
	ctx := context.Background()
	tagInfo := parserTag(param)
	tg := &graphTopo{adj: map[string]map[string]struct{}{}, nodes: map[string]*Node{}}
	table, _ := topology.Loader.GetSearchTable(orgName)
	for key, typeIndices := range typeMetricGroupMap {
		aggregationConditions, _ := clickhousesource.SelectRelation(key)
		for _, agg := range aggregationConditions {
			sd := goqu.From(table).Where(
				goqu.C("org_name").In(orgName, ""),
				goqu.C("metric_group").In(typeIndices),
			)
			sd = queryConditionsV2(sd, param, tagInfo)
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

			topology.parseToTypologyNodeV2(lang, timeRange, rows, tg)
		}
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

func (topology *provider) parseToTypologyNodeV2(lang i18n.LanguageCodes, timeRange int64, rows driver.Rows, tg *graphTopo) {
	for rows.Next() {
		cnode := chNodeEdge{}
		err := rows.ScanStruct(&cnode)
		if err != nil {
			topology.Log.Warnf("scan chNodeEdge failed: %s", err)
			continue
		}

		edge := cnode.ToTopologyNodeRelation()
		targetNode, sourceNode := columnsParser(cnode.TargetType, edge), columnsParser(cnode.SourceType, edge)

		if targetNode.Valid() {
			targetNode.TypeDisplay = topology.t.Text(lang, strings.ToLower(targetNode.Type))
			targetNode.Metric = edge.Metric
			if n, ok := tg.nodes[targetNode.Id]; ok { // merge metric
				n.Metric.Count += targetNode.Metric.Count
				n.Metric.HttpError += targetNode.Metric.HttpError
				n.Metric.Duration += targetNode.Metric.Duration
				n.Metric.RT += targetNode.Metric.RT
				if n.RuntimeId == "" {
					n.RuntimeId = targetNode.RuntimeId
				}
			} else {
				tg.nodes[targetNode.Id] = targetNode
			}

			if _, ok := tg.adj[targetNode.Id]; !ok {
				tg.adj[targetNode.Id] = map[string]struct{}{}
			}
		}

		if sourceNode.Valid() {
			sourceNode.TypeDisplay = topology.t.Text(lang, strings.ToLower(sourceNode.Type))
			sourceNode.Metric = edge.Metric
			tg.nodes[sourceNode.Id] = sourceNode

			tg.adj[targetNode.Id][sourceNode.Id] = struct{}{}
		}
	}
}

func queryConditionsV2(sd *goqu.SelectDataset, params Vo, tagInfo *TagInfo) *goqu.SelectDataset {
	where := sd
	where = where.Where(
		goqu.C("tenant_id").Eq(params.TerminusKey),
		goqu.C("timestamp").Gte(ckhelper.FromTimestampMilli(params.StartTime)),
		goqu.C("timestamp").Lte(ckhelper.FromTimestampMilli(params.EndTime)),
	)

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
