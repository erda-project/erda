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
	"strings"
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

const timeFormat = "2006-01-02 15:04:05"

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
		goqu.L("timestamp").Gte(time.Unix(start, 0).Local().Format(timeFormat)),
		goqu.L("timestamp").Lt(time.Unix(end, 0).Local().Format(timeFormat)),
	)

	if instanceType != instanceTypeAll {
		sql = sql.Where(goqu.L("tag_values[indexOf(tag_keys,?)]", "instance_type").Eq(instanceType))
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
		chs.Log.Errorf("failed to convert sql to string, %v", err)
		return nil
	}

	rows, err := chs.Clickhouse.Client().Query(ctx, sqlStr)
	if err != nil {
		chs.Log.Errorf("failed to query clickhouse, %v", err)
		return nil
	}
	defer rows.Close()

	var res []*containerData
	for rows.Next() {
		row := containerRow{}
		if err := rows.ScanStruct(&row); err != nil {
			chs.Log.Errorf("failed to scan ch row, %v", err)
			continue
		}
		cd := chs.convert(&row)
		if cd != nil {
			res = append(res, cd)
		}
	}
	return res
}

func (chs *ClickhouseSource) GetHostTypes(req *http.Request, params struct {
	ClusterName string `query:"clusterName" validate:"required"`
	OrgName     string `query:"orgName" validate:"required"`
}) interface{} {
	var clusterNames []interface{}
	for _, v := range strings.Split(params.ClusterName, ",") {
		clusterNames = append(clusterNames, v)
	}

	table, _ := chs.Loader.GetSearchTable(params.OrgName)
	sql := goqu.From(table).Select(
		goqu.L("tag_values[indexOf(tag_keys,?)]", clusterName).As("clusterName"),
		goqu.L("tag_values[indexOf(tag_keys,?)]", "n_cpus").As("cpus"),
		goqu.L("tag_values[indexOf(tag_keys,?)]", mem).As("mem"),
		goqu.L("tag_values[indexOf(tag_keys,?)]", hostIP).As("hostIP"),
		goqu.L("string_field_values[indexOf(string_field_keys,?)]", labels).As("labels"),
	).Where(
		goqu.L("metric_group").Eq(groupHostSummary),
		goqu.L("org_name").Eq(params.OrgName),
		goqu.L("tag_values[indexOf(tag_keys,?)]", clusterName).In(clusterNames),
		goqu.L("string_field_values[indexOf(string_field_keys,?)]", labels).NotLike("%offline%")).Limit(500)

	sqlStr, err := chs.toSQL(sql)
	if err != nil {
		chs.Log.Errorf("failed to convert sql to string, %v", err)
		return nil
	}

	rows, err := chs.Clickhouse.Client().Query(req.Context(), sqlStr)
	if err != nil {
		chs.Log.Errorf("failed to query clickhouse, %v", err)
		return nil
	}

	var hostTypes []*hostTypeRow
	for rows.Next() {
		var hostType hostTypeRow
		if err = rows.ScanStruct(&hostType); err != nil {
			chs.Log.Errorf("failed to scan ch row, %v", err)
			continue
		}
		hostTypes = append(hostTypes, &hostType)
	}

	var (
		cpuList         = make(map[string]struct{})
		memList         = make(map[string]struct{})
		clusterNameList = make(map[string]struct{})
		hostIPList      = make(map[string]struct{})
		labelList       = make(map[string]struct{})
	)
	for _, hostType := range hostTypes {
		if _, ok := cpuList[hostType.CPUs]; !ok {
			cpuList[hostType.CPUs] = struct{}{}
		}
		if _, ok := memList[hostType.Mem]; !ok {
			memList[hostType.Mem] = struct{}{}
		}
		if _, ok := clusterNameList[hostType.ClusterName]; !ok {
			clusterNameList[hostType.ClusterName] = struct{}{}
		}
		if _, ok := hostIPList[hostType.HostIP]; !ok {
			hostIPList[hostType.HostIP] = struct{}{}
		}
		labels := strings.Split(hostType.Labels, ",")
		for _, label := range labels {
			if _, ok := labelList[label]; !ok {
				labelList[label] = struct{}{}
			}
		}
	}

	res := []*groupHostTypeData{
		{
			Key:    cpus,
			Values: mapToSlice(cpuList),
		},
		{
			Key:    mem,
			Values: mapToSlice(memList),
		},
		{
			Key:    cluster,
			Values: mapToSlice(clusterNameList),
		},
		{
			Key:    hostIP,
			Values: mapToSlice(hostIPList),
		},
		{
			Key:    labels,
			Values: mapToSlice(labelList),
		},
		{
			Key:    cpuUsagePercent,
			Name:   "cpu used",
			Values: percents,
			Prefix: "cpu used",
		},
		{
			Key:    memUsagePercent,
			Name:   "mem used",
			Values: percents,
			Prefix: "mem used",
		},
		{
			Key:    diskUsagePercent,
			Name:   "disk used",
			Values: percents,
			Prefix: "disk used",
		},
		{
			Key:    cpuDispPercent,
			Name:   "cpu dispatch",
			Values: percents,
			Prefix: "cpu dispatch",
		},
		{
			Key:    memDispPercent,
			Name:   "mem dispatch",
			Values: percents,
			Prefix: "mem dispatch",
		},
	}

	lang := api.Language(req)
	for _, typ := range res {
		if len(typ.Name) <= 0 {
			typ.Name = chs.p.t.Text(lang, typ.Key)
		} else {
			typ.Name = chs.p.t.Text(lang, typ.Name)
		}
		if len(typ.Prefix) >= 0 {
			typ.Prefix = chs.p.t.Text(lang, typ.Prefix)
		}
	}
	return api.Success(res)
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

func mapToSlice(mp map[string]struct{}) []string {
	var res []string
	for key, _ := range mp {
		res = append(res, key)
	}
	return res
}
