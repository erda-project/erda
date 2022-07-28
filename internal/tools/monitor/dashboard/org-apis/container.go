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

package orgapis

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/olivere/elastic"

	"github.com/erda-project/erda-infra/providers/httpserver"

	"github.com/erda-project/erda/internal/tools/monitor/utils"
	api "github.com/erda-project/erda/pkg/common/httpapi"
)

func (p *provider) getContainers(ctx httpserver.Context, r *http.Request, params struct {
	InstanceType string `param:"instance_type" validate:"required"`
	Start        int64  `query:"start"`
	End          int64  `query:"end"`
}, res resourceRequest) interface{} {
	err := p.checkOrgByClusters(ctx, res.Clusters)
	if err != nil {
		return nil
	}
	now, timeRange := time.Now().UnixNano()/int64(time.Millisecond), 5*int64(time.Minute)/int64(time.Millisecond)
	if params.End < timeRange {
		params.End = now
	}
	if params.Start <= 0 {
		params.Start = params.End - timeRange
	}

	var (
		wg     sync.WaitGroup
		lock   sync.RWMutex
		result = make([]*containerData, 0, 16*len(res.Clusters))
	)
	wg.Add(len(res.Clusters))
	for _, cluster := range res.Clusters {
		go func(clusterName string, hostIPs []string) {
			defer wg.Done()
			containers := p.queryContainers(clusterName, hostIPs, params.InstanceType, res.Filters, params.Start, params.End)
			lock.Lock()
			defer lock.Unlock()
			result = append(result, containers...)
		}(cluster.ClusterName, cluster.HostIPs)
	}
	wg.Wait()
	return api.Success(result)
}

func (p *provider) queryContainers(cluster string, hostIPs []string, instanceType string, filters []*resourceFilter, start, end int64) []*containerData {
	query := elastic.NewBoolQuery().
		Filter(elastic.NewTermQuery(tagsClusterName, cluster)).
		Filter(elastic.NewRangeQuery(timestamp).Gte(start * int64(time.Millisecond)).Lt(end * int64(time.Millisecond))).
		Filter(elastic.NewTermsQuery(tagsHostIP, utils.ConvertStringArrToInterfaceArr(hostIPs)...)).
		MustNot(elastic.NewTermQuery("tags.container", "POD")).
		MustNot(elastic.NewTermQuery("tags.podsandbox", "true"))
	if instanceType != instanceTypeAll {
		query = query.Filter(elastic.NewTermQuery(tagsInstanceType, instanceType))
	}
	for _, filter := range filters {
		key := convertKey(filter.Key)
		if key == fieldsLabels {
			for _, v := range filter.Values {
				query.Filter(elastic.NewTermQuery(fieldsLabels, v))
			}
		} else {
			query.Filter(elastic.NewTermsQuery(key, utils.ConvertStringArrToInterfaceArr(filter.Values)...))
		}
	}

	topHitsAgg := elastic.NewTopHitsAggregation().Size(1).Sort(timestamp, false).
		FetchSourceContext(elastic.NewFetchSourceContext(true).Include(any))

	containerIDAgg := elastic.NewTermsAggregation().Field(tagsContainerID).Size(500).
		SubAggregation(cpuUsagePercent, elastic.NewMaxAggregation().Field(fieldsCPUUsagePercent)).
		SubAggregation(cpuLimit, elastic.NewMaxAggregation().Field(fieldsCPULimit)).
		SubAggregation(cpuRequest, elastic.NewMaxAggregation().Field(fieldsCPURequest)).
		SubAggregation(memUsage, elastic.NewMaxAggregation().Field(fieldsMemUsage)).
		SubAggregation(memLimit, elastic.NewMaxAggregation().Field(fieldsMemLimit)).
		SubAggregation(memRequest, elastic.NewMaxAggregation().Field(fieldsMemRequest)).
		SubAggregation(diskUsage, elastic.NewMaxAggregation().Field(fieldsDiskUsage)).
		SubAggregation(tagsTerminusVersion, elastic.NewTermsAggregation().Field(tagsTerminusVersion).Size(100).SubAggregation(topHits, topHitsAgg)).
		SubAggregation(topHits, topHitsAgg)

	searchSource := elastic.NewSearchSource().Query(query).Size(0).Aggregation(tagsContainerID, containerIDAgg)
	resp, apiErr := p.EsSearchRaw.QueryRaw([]string{nameContainerSummary, nameDockerContainerSummary}, []string{cluster}, start, end, searchSource)
	if apiErr != nil {
		return nil
	} else if resp == nil {
		return nil
	}
	return parseQueryContainer(resp.Aggregations)
}

func parseQueryContainer(agg elastic.Aggregations) []*containerData {
	if agg == nil {
		return nil
	}
	containerIDAgg, ok := agg.Terms(tagsContainerID)
	if !ok {
		return nil
	}
	var containersData []*containerData
	for _, item := range containerIDAgg.Buckets {
		topHitsAgg, _ := item.TopHits(topHits)
		if versionAgg, ok := item.Terms(tagsTerminusVersion); ok {
			var maxVersion int64
			for _, versionItem := range versionAgg.Buckets {
				if version, err := versionItem.KeyNumber.Int64(); err != nil {
					continue
				} else if version > maxVersion {
					topHitsAgg, ok = versionItem.Aggregations.TopHits(topHits)
					if !ok {
						continue
					}
					maxVersion = version
				}
			}
		}

		data := &containerData{}
		data.CpuUsage = rawToFloat64(item.Aggregations[cpuUsage])
		data.CpuRequest = rawToFloat64(item.Aggregations[cpuRequest])
		data.CpuLimit = rawToFloat64(item.Aggregations[cpuLimit])
		data.CpuOrigin = rawToFloat64(item.Aggregations[cpuOrigin])
		data.MemUsage = rawToFloat64(item.Aggregations[memUsage])
		data.MemRequest = rawToFloat64(item.Aggregations[memRequest])
		data.MemLimit = rawToFloat64(item.Aggregations[memLimit])
		data.MemOrigin = rawToFloat64(item.Aggregations[memOrigin])
		data.DiskUsage = rawToFloat64(item.Aggregations[diskUsage])
		data.DiskLimit = rawToFloat64(item.Aggregations[diskLimit])
		wrapContainerData(data, topHitsAgg)
		containersData = append(containersData, data)
	}
	return containersData
}

func rawToFloat64(raw *json.RawMessage) float64 {
	if raw == nil || len(*raw) == 0 {
		return 0
	}
	var res struct {
		Value interface{} `json:"value"`
	}
	err := json.Unmarshal(*raw, &res)
	if err != nil {
		return 0
	}
	switch data := res.Value.(type) {
	case float64:
		return data
	case int64:
		return float64(data)
	default:
		return 0
	}
}

func wrapContainerData(src *containerData, topHits *elastic.AggregationTopHitsMetric) {
	if topHits == nil || topHits.Hits == nil {
		return
	}

	for _, hit := range topHits.Hits.Hits {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(*hit.Source), &m); err != nil {
			continue
		}

		tags, ok := utils.GetMapValueMap(m, tags)
		if !ok {
			continue
		}
		// fields, ok := utils.GetMapValueMap(m, fields)
		// if !ok {
		// 	continue
		// }

		isDeleted, _ := utils.GetMapValueString(tags, isDeleted)
		if isDeleted == "true" {
			continue
		}
		if _, ok := utils.GetMapValueString(tags, image, "container_image"); !ok {
			continue
		}

		if val, ok := utils.GetMapValueString(tags, image, "container_image"); ok {
			src.Image = val
		}

		src.ClusterName, _ = utils.GetMapValueString(tags, clusterName)
		src.HostIP, _ = utils.GetMapValueString(tags, hostIP)
		src.ContainerID, _ = utils.GetMapValueString(tags, containerID)
		src.InstanceType, _ = utils.GetMapValueString(tags, instanceType)
		src.InstanceID, _ = utils.GetMapValueString(tags, instanceID)
		src.OrgID, _ = utils.GetMapValueString(tags, orgID)
		src.OrgName, _ = utils.GetMapValueString(tags, orgName)
		src.ProjectID, _ = utils.GetMapValueString(tags, projectID)
		src.ProjectName, _ = utils.GetMapValueString(tags, projectName)
		src.ApplicationID, _ = utils.GetMapValueString(tags, applicationID)
		src.ApplicationName, _ = utils.GetMapValueString(tags, applicationName)
		src.Workspace, _ = utils.GetMapValueString(tags, workspace)
		src.RuntimeID, _ = utils.GetMapValueString(tags, runtimeID)
		src.RuntimeName, _ = utils.GetMapValueString(tags, runtimeName)
		src.ServiceID, _ = utils.GetMapValueString(tags, serviceID)
		src.ServiceName, _ = utils.GetMapValueString(tags, serviceName)
		src.JobID, _ = utils.GetMapValueString(tags, jobID)

		src.Container, _ = utils.GetMapValueString(tags, "container")
		src.PodUid, _ = utils.GetMapValueString(tags, "pod_uid")
		src.PodName, _ = utils.GetMapValueString(tags, "pod_name")
		src.PodNamespace, _ = utils.GetMapValueString(tags, "pod_namespace")
	}
	return
}

func (p *provider) mergeResourceChart(list []*resourceChart) *resourceChart {
	var res *resourceChart
	for _, item := range list {
		if res == nil || len(res.Results) == 0 || len(res.Results[0].Data) == 0 {
			res = item
			continue
		}
		if len(item.Results) == 0 || len(item.Results[0].Data) == 0 {
			continue
		}
		resData := res.Results[0].Data[0]
		itemData := item.Results[0].Data[0]
		for k, v := range itemData {
			val, ok := resData[k]
			if !ok {
				resData[k] = v
				continue
			}
			for i, item := range v.Data {
				if len(v.Data) > i {
					val.Data[i] += item
				} else {
					val.Data = append(val.Data, item)
				}
			}
		}
	}
	return res
}
