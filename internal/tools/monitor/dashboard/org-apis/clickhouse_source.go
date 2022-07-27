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
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/doug-martin/goqu/v9"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/clickhouse"
	"github.com/erda-project/erda-infra/providers/httpserver"
	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"
	"github.com/erda-project/erda/pkg/common/apis"
	api "github.com/erda-project/erda/pkg/common/httpapi"
)

type ClickhouseSource struct {
	p          *provider
	Clickhouse clickhouse.Interface
	Log        logs.Logger
	DebugSQL   bool
	Loader     loader.Interface
}

func (chs *ClickhouseSource) GetContainers(ctx httpserver.Context, r *http.Request, params struct {
	InstanceType string `param:"instance_type" validate:"required"`
	Start        int64  `query:"start"`
	End          int64  `query:"end"`
}, res resourceRequest) interface{} {
	err := chs.p.checkOrgByClusters(ctx, res.Clusters)
	if err != nil {
		return nil
	}
	now, timeRange := time.Now().Unix(), 5*int64(time.Minute)/int64(time.Second)
	if params.End < timeRange {
		params.End = now
	}
	if params.Start <= 0 {
		params.Start = params.End - timeRange
	}

	org, err := chs.p.Org.GetOrg(apis.WithInternalClientContext(context.Background(), "monitor"),
		&orgpb.GetOrgRequest{IdOrName: api.OrgID(ctx.Request())})
	if err != nil {
		chs.Log.Errorf("failed to get org, %v", err)
		return nil
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
			containers := chs.queryContainers(ctx.Request().Context(), org.Data.Name, clusterName, hostIPs, params.InstanceType, res.Filters, params.Start, params.End)
			lock.Lock()
			defer lock.Unlock()
			result = append(result, containers...)
		}(cluster.ClusterName, cluster.HostIPs)
	}
	wg.Wait()
	return api.Success(result)
}

func (chs *ClickhouseSource) queryContainers(ctx context.Context, orgName, cluster string, hostIPs []string, instanceType string,
	filters []*resourceFilter, start, end int64) []*containerData {
	table, _ := chs.Loader.GetSearchTable(orgName)
	sql := goqu.From(table).Select(
		goqu.L("tag_values[indexOf(tag_keys,?)]", containerID).As("containerID"),
		goqu.L("any(tag_keys)").As("tagKeys"),
		goqu.L("any(tag_values)").As("tagValues"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys,?)])", cpuUsage).As("cpuUsage"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys,?)])", cpuRequest).As("cpuRequest"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys,?)])", cpuLimit).As("cpuLimit"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys,?)])", cpuOrigin).As("cpuOrigin"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys,?)])", memUsage).As("memUsage"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys,?)])", memRequest).As("memRequest"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys,?)])", memLimit).As("memLimit"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys,?)])", memOrigin).As("memOrigin"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys,?)])", diskUsage).As("diskUsage"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys,?)])", diskLimit).As("diskLimit"),
	).Where(
		goqu.L("metric_group").In([]string{nameContainerSummary, nameDockerContainerSummary}),
		goqu.L("tag_values[indexOf(tag_keys,?)]", "container").Neq("POD"),
		goqu.L("tag_values[indexOf(tag_keys,?)]", "podsandbox").Neq("true"),
		goqu.L("tag_values[indexOf(tag_keys,?)]", hostIP).In(hostIPs),
		goqu.L("tag_values[indexOf(tag_keys,?)]", clusterName).Eq(cluster),
		goqu.L("toUnixTimestamp(timestamp)").Gte(start),
		goqu.L("toUnixTimestamp(timestamp)").Lt(end),
	)

	if instanceType != instanceTypeAll {
		sql = sql.Where(goqu.L("tag_values[indexOf(tag_keys,?)]", instanceType).Eq(instanceType))
	}
	for _, filter := range filters {
		key := convertKeyV2(filter.Key)
		if len(filter.Values) == 0 {
			continue
		} else if len(filter.Values) == 1 {
			sql.Where(goqu.L("tag_values[indexOf(tag_keys,?)]", key).Eq(filter.Values[0]))
		} else {
			sql.Where(goqu.L("tag_values[indexOf(tag_keys,?)]", key).In(filter.Values))
		}
	}
	sql = sql.GroupBy("containerID")

	sqlStr, err := chs.toSQL(sql)
	if err != nil {
		chs.Log.Error(err)
		return nil
	}

	rows, err := chs.Clickhouse.Client().Query(ctx, sqlStr)
	if err != nil {
		chs.Log.Error(err)
		return nil
	}
	defer rows.Close()

	var res []*containerData
	for rows.Next() {
		row := containerRow{}
		if err := rows.ScanStruct(&row); err != nil {
			chs.Log.Errorf("failed to scan ch row, %v", err)
			return nil
		}
		cd := chs.convert(&row)
		if cd != nil {
			res = append(res, cd)
		}
	}
	return res
}

func (chs *ClickhouseSource) toSQL(sql *goqu.SelectDataset) (string, error) {
	str, _, err := sql.ToSQL()
	if err != nil {
		return "", fmt.Errorf("failed to convert to sql: %v", err)
	}
	if chs.DebugSQL {
		chs.Log.Infof("Metrics clickhouse SQL: \n%s", str)
	}
	return str, nil
}

func (chs *ClickhouseSource) convert(row *containerRow) *containerData {
	tagKeys := row.TagKeys
	tagValues := row.TagValues
	if len(tagKeys) != len(tagValues) {
		chs.Log.Warnf("length of tagKeys and tagValues are not equal, continue")
		return nil
	}

	tags := make(map[string]string)
	for i := 0; i < len(tagKeys); i++ {
		key := interfaceToString(tagKeys[i])
		tags[key] = interfaceToString(tagValues[i])
	}
	return &containerData{
		ClusterName:     tags[clusterName],
		HostIP:          tags[hostIP],
		ContainerID:     tags[containerID],
		InstanceType:    tags[instanceType],
		InstanceID:      tags[instanceID],
		Image:           tags[image],
		OrgID:           tags[orgID],
		OrgName:         tags[orgName],
		ProjectID:       tags[projectID],
		ProjectName:     tags[projectName],
		ApplicationID:   tags[applicationID],
		ApplicationName: tags[applicationName],
		Workspace:       tags[workspace],
		RuntimeID:       tags[runtimeID],
		RuntimeName:     tags[runtimeName],
		ServiceID:       tags[serviceID],
		ServiceName:     tags[serviceName],
		JobID:           tags[jobID],
		CpuUsage:        row.CpuUsage,
		CpuRequest:      row.CpuRequest,
		CpuLimit:        row.CpuLimit,
		CpuOrigin:       row.CpuOrigin,
		MemUsage:        row.MemUsage,
		MemRequest:      row.MemRequest,
		MemLimit:        row.MemLimit,
		MemOrigin:       row.MemOrigin,
		DiskUsage:       row.DiskUsage,
		DiskLimit:       row.DiskLimit,
		Status:          tags["status"],
		Container:       tags["container"],
		PodUid:          tags["pod_uid"],
		PodName:         tags[podName],
		PodNamespace:    tags["pod_namespace"],
	}
}

func interfaceToString(obj interface{}) string {
	str, ok := obj.(string)
	if ok {
		return str
	}
	return fmt.Sprintf("%s", obj)
}

func convertKeyV2(key string) string {
	if key == host {
		return hostIP
	}
	if key == cluster {
		return clusterName
	}
	return key
}
