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

package details_apis

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/olivere/elastic"

	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq/query"
	"github.com/erda-project/erda/modules/monitor/utils"
)

// PodInfoSummary .
type PodInfoSummary struct {
	ClusterName      interface{} `json:"clusterName"`
	NodeName         interface{} `json:"nodeName"`
	HostIP           interface{} `json:"hostIP"`
	Namespace        interface{} `json:"namespace"`
	PodName          interface{} `json:"podName"`
	RestartTotal     interface{} `json:"restartTotal"`
	StateCode        interface{} `json:"stateCode"`
	TerminatedReason interface{} `json:"terminatedReason"`
}

// PodInfoInstanse .
type PodInfoInstanse struct {
	ContainerID interface{} `json:"containerId"`
	HostIP      interface{} `json:"hostIP"`
	StartedAt   interface{} `json:"startedAt"`
	FinishedAt  interface{} `json:"finishedAt"`
	ExitCode    interface{} `json:"exitCode"`
	OomKilled   interface{} `json:"oomKilled"`
}

// PodInfo .
type PodInfo struct {
	Summary   PodInfoSummary     `json:"summary"`
	Instances []*PodInfoInstanse `json:"instances"`
}

func (p *provider) getPodInfo(clusterName, podName string, start, end int64) (*PodInfo, error) {
	boolQuery := elastic.NewBoolQuery().
		Filter(elastic.NewRangeQuery(query.TimestampKey).Gte(start * 1000000).Lte(end * 1000000)).
		Filter(elastic.NewTermQuery(query.ClusterNameKey, clusterName)).
		Filter(elastic.NewTermQuery(query.TagKey+".pod_name", podName))
	searchSource := elastic.NewSearchSource().Query(boolQuery).Size(0)
	searchSource.Aggregation("pod", elastic.NewFilterAggregation().
		Filter(elastic.NewBoolQuery().Filter(elastic.NewTermQuery("name", "kubernetes_pod_container"))).
		SubAggregation("pod", elastic.NewTopHitsAggregation().Size(1).Sort(query.TimestampKey, false)),
	)
	const containerIDKey = query.TagKey + ".container_id"
	searchSource.Aggregation("containers", elastic.NewFilterAggregation().
		Filter(elastic.NewBoolQuery().Filter(elastic.NewTermQuery("name", "docker_container_status")).MustNot(elastic.NewTermQuery("tags.podsandbox", "true"))).
		SubAggregation(containerIDKey, elastic.NewTermsAggregation().Field(containerIDKey).
			SubAggregation("container", elastic.NewTopHitsAggregation().Size(1).Sort(query.TimestampKey, false).Sort("fields.finished_at", false)),
		),
	)

	resp, err := p.metricq.QueryRaw([]string{"kubernetes_pod_container", "docker_container_status"},
		[]string{clusterName}, start, end, searchSource)
	if err != nil {
		return nil, fmt.Errorf("fail to query pod and containers: %s", err)
	}
	var info PodInfo
	if resp.Aggregations == nil {
		return &info, nil
	}
	pod, ok := resp.Aggregations.Filter("pod")
	if ok && pod != nil && pod.Aggregations != nil {
		topHis, ok := pod.Aggregations.TopHits("pod")

		if ok && topHis != nil && topHis.Hits != nil && len(topHis.Hits.Hits) > 0 {

			var source map[string]interface{}
			err := json.Unmarshal([]byte(*topHis.Hits.Hits[0].Source), &source)
			if err == nil {
				info.Summary.ClusterName = utils.GetMapValue("tags.cluster_name", source)
				info.Summary.NodeName = utils.GetMapValue("tags.node_name", source)
				info.Summary.HostIP = utils.GetMapValue("tags.host_ip", source)
				info.Summary.Namespace = utils.GetMapValue("tags.namespace", source)
				info.Summary.PodName = utils.GetMapValue("tags.pod_name", source)
				info.Summary.StateCode = utils.GetMapValue("fields.state_code", source)
				info.Summary.RestartTotal = utils.GetMapValue("fields.restarts_total", source)
				info.Summary.TerminatedReason = utils.GetMapValue("fields.terminated_reason", source)
			}
		}
	}
	containers, ok := resp.Aggregations.Filter("containers")
	if ok && containers != nil && containers.Aggregations != nil {
		terms, ok := containers.Aggregations.Terms(containerIDKey)
		if ok && terms != nil {
			for _, b := range terms.Buckets {
				if b.Aggregations == nil {
					continue
				}
				topHis, ok := b.Aggregations.TopHits("container")
				if ok && topHis != nil && topHis.Hits != nil && len(topHis.Hits.Hits) > 0 {
					var source map[string]interface{}
					err := json.Unmarshal([]byte(*topHis.Hits.Hits[0].Source), &source)
					if err == nil {
						inst := &PodInfoInstanse{
							ContainerID: b.Key,
							HostIP:      utils.GetMapValue("tags.host_ip", source),
							StartedAt:   formatDate(utils.GetMapValue("fields.started_at", source)),
							FinishedAt:  formatDate(utils.GetMapValue("fields.finished_at", source)),
							ExitCode:    utils.GetMapValue("fields.exitcode", source),
							OomKilled:   utils.GetMapValue("fields.oomkilled", source),
						}
						info.Instances = append(info.Instances, inst)
					}
				}
			}
		}
	}
	return &info, nil
}

func formatDate(date interface{}) interface{} {
	val, ok := utils.ConvertInt64(date)
	if !ok {
		return date
	}
	return time.Unix(val/int64(time.Second), val%int64(time.Second)).Format("2006-01-02 15:04:05")
}
