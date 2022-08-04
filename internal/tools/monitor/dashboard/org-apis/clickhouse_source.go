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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"

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
	orgChecker orgChecker
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
	err := chs.orgChecker.checkOrgByClusters(ctx, res.Clusters)
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

	org, err := chs.p.Org.GetOrg(apis.WithInternalClientContext(api.GetContextHeader(r), "monitor"),
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
		goqu.L("org_name").Eq(orgName),
		goqu.L("tenant_id").Eq(orgName),
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
		sql = chs.getFilterByKey(sql, filter.Key, filter.Values...)
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
			chs.Log.Errorf("failed to scan ch row to containerData, %v", err)
			continue
		}
		cd := chs.parseContainer(&row)
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

	from := time.Now().Add(-5 * time.Minute).Format(timeFormat)
	to := time.Now().Format(timeFormat)
	sql := goqu.From(table).Select(
		goqu.L("tag_values[indexOf(tag_keys,?)]", clusterName).As("clusterName"),
		goqu.L("tag_values[indexOf(tag_keys,?)]", ncpus).As("cpus"),
		goqu.L("tag_values[indexOf(tag_keys,?)]", mem).As("mem"),
		goqu.L("tag_values[indexOf(tag_keys,?)]", hostIP).As("hostIP"),
		goqu.L("string_field_values[indexOf(string_field_keys,?)]", labels).As("labels"),
	).Where(
		goqu.L("metric_group").Eq(groupHostSummary),
		goqu.L("org_name").Eq(params.OrgName),
		goqu.L("tenant_id").Eq(params.OrgName),
		goqu.L("tag_values[indexOf(tag_keys,?)]", clusterName).In(clusterNames),
		goqu.L("string_field_values[indexOf(string_field_keys,?)]", labels).NotLike("%offline%"),
		goqu.L("timestamp").Between(goqu.Range(from, to)))

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
			chs.Log.Errorf("failed to scan ch row to hostTypeRow, %v", err)
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
			Key:    host,
			Values: mapToSlice(hostIPList),
		},
		{
			Key:    labels,
			Values: mapToSlice(labelList),
		},
		{
			Key:    cpuUsageActive,
			Name:   "cpu used",
			Values: percents,
			Prefix: "cpu used",
		},
		{
			Key:    memUsedPercent,
			Name:   "mem used",
			Values: percents,
			Prefix: "mem used",
		},
		{
			Key:    diskUsedPercent,
			Name:   "disk used",
			Values: percents,
			Prefix: "disk used",
		},
		{
			Key:    cpuRequestPercent,
			Name:   "cpu dispatch",
			Values: percents,
			Prefix: "cpu dispatch",
		},
		{
			Key:    memRequestPercent,
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

func (chs *ClickhouseSource) GetGroupHosts(req *http.Request, params struct {
	OrgName string `query:"orgName" validate:"required" json:"-"`
}, res resourceRequest) interface{} {
	var clusterNames []interface{}
	for _, cluster := range res.Clusters {
		clusterNames = append(clusterNames, cluster.ClusterName)
	}

	table, _ := chs.Loader.GetSearchTable(params.OrgName)

	from := time.Now().Add(-5 * time.Minute).Format(timeFormat)
	to := time.Now().Format(timeFormat)
	sql := goqu.From(table).Select(
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", cpuCoresUsage).As("cpuCoresUsage"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", cpuRequestTotal).As("cpuRequestTotal"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", cpuLimitTotal).As("cpuLimitTotal"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", cpuOriginTotal).As("cpuOriginTotal"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", ncpus).As("cpuTotal"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", cpuAllocatable).As("cpuAllocatable"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", memUsed).As("memUsed"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", memRequestTotal).As("memRequestTotal"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", memLimitTotal).As("memLimitTotal"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", memOriginTotal).As("memOriginTotal"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", memTotal).As("memTotal"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", memAllocatable).As("memAllocatable"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", diskUsed).As("diskUsed"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", diskTotal).As("diskTotal"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", load1).As("load1"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", load5).As("load5"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", load15).As("load15"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", cpuUsageActive).As("cpuUsageActive"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", memUsedPercent).As("memUsedPercent"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", diskUsedPercent).As("diskUsedPercent"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", cpuRequestPercent).As("cpuRequestPercent"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", memRequestPercent).As("memRequestPercent"),
		goqu.L("MAX(number_field_values[indexOf(number_field_keys, ?)])", taskContainers).As("taskContainers"),
		goqu.L("any(string_field_values[indexOf(string_field_keys,?)])", labels).As("labels"),
		goqu.L("any(tag_keys)").As("tagKeys"),
		goqu.L("any(tag_values)").As("tagValues"),
	).Where(
		goqu.L("metric_group").Eq("host_summary"),
		goqu.L("tag_values[indexOf(tag_keys,?)]", clusterName).In(clusterNames),
		goqu.L("org_name").Eq(params.OrgName),
		goqu.L("tenant_id").Eq(params.OrgName),
		goqu.L("string_field_values[indexOf(string_field_keys,?)]", labels).NotLike("%offline%"),
		goqu.L("timestamp").Between(goqu.Range(from, to)),
	).GroupBy(goqu.L("tag_values[indexOf(tag_keys,?)]", hostIP))

	sql = chs.wrapGroupHostFilter(res.Filters, sql)
	sqlStr, err := chs.toSQL(sql)
	if err != nil {
		chs.Log.Errorf("failed to convert sql to str, %v", err)
		return nil
	}

	rows, err := chs.Clickhouse.Client().Query(req.Context(), sqlStr)
	if err != nil {
		chs.Log.Errorf("failed to query clickhouse, %v", err)
		return nil
	}

	var hosts []*hostData
	for rows.Next() {
		host := hostRow{}
		if err := rows.ScanStruct(&host); err != nil {
			chs.Log.Errorf("failed to scan row to hostRow, %v", err)
			continue
		}
		hd := parseHostRow(&host)
		if hd != nil {
			hosts = append(hosts, hd)
		}
	}

	groups := chs.parseGroupHost(hosts, res.Groups, 0)
	chs.p.updateClusterStatus(api.GetContextHeader(req), groups)
	return api.Success(groups)
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

func (chs *ClickhouseSource) parseContainer(row *containerRow) *containerData {
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

func (chs *ClickhouseSource) wrapGroupHostFilter(filters []*resourceFilter, sql *goqu.SelectDataset) *goqu.SelectDataset {
	for _, filter := range filters {
		switch filter.Key {
		case cpuCoresUsage, memUsedPercent, diskUsedPercent, cpuRequestPercent, memRequestPercent:
			var exps []exp.Expression
			for _, value := range filter.Values {
				if strings.HasPrefix(value, ">=") {
					val := value[2:]
					from, err := convertFilterPairValue(val)
					if err != nil {
						continue
					}
					exps = append(exps, goqu.L("number_field_values[indexOf(number_field_keys,?)]", filter.Key).Gte(from))
				} else if strings.Contains(value, "-") {
					vs := strings.Split(value, "-")
					val := vs[0]
					from, err := convertFilterPairValue(val)
					if err != nil {
						continue
					}
					to := 0.0
					if len(vs) > 1 {
						val = vs[1]
						to, err = convertFilterPairValue(val)
						if err != nil {
							continue
						}
					}
					exps = append(exps, goqu.L("number_field_values[indexOf(number_field_keys,?)]", filter.Key).Between(goqu.Range(from, to)))
				}
			}
			sql = sql.Where(goqu.Or(exps...))
		default:
			sql = chs.getFilterByKey(sql, filter.Key, filter.Values...)
		}
	}
	return sql
}

func (chs *ClickhouseSource) getFilterByKey(sql *goqu.SelectDataset, key string, values ...string) *goqu.SelectDataset {
	if len(values) == 0 {
		return sql
	}
	key = chs.convertKey(key)
	var exp exp.LiteralExpression
	switch key {
	case ncpus, mem, clusterName, hostIP:
		exp = goqu.L("tag_values[indexOf(tag_keys,?)]", key)
	case labels:
		exp = goqu.L("string_field_values[indexOf(string_field_keys,?)]", key)
	default:
		chs.Log.Errorf("unsupported filter key: %v", key)
		return sql
	}

	if len(values) == 1 {
		return sql.Where(exp.Eq(values[0]))
	}
	return sql.Where(exp.In(values))
}

func (chs *ClickhouseSource) parseGroupHost(hosts []*hostData, groups []string, index int) *groupHostData {
	group := new(groupHostData)
	if index == len(groups) {
		group.Machines = hosts
		metric := new(groupHostMetric)
		for _, hostData := range hosts {
			metric.Machines++
			metric.CPUUsage += hostData.CPUUsage
			metric.CPURequest += hostData.CPURequest
			metric.CPULimit += hostData.CPULimit
			metric.CPUOrigin += hostData.CPUOrigin
			metric.CPUTotal += hostData.CPUTotal
			metric.CPUAllocatable += hostData.CPUAllocatable
			metric.MemUsage += hostData.MemUsage
			metric.MemRequest += hostData.MemRequest
			metric.MemLimit += hostData.MemLimit
			metric.MemOrigin += hostData.MemOrigin
			metric.MemTotal += hostData.MemTotal
			metric.MemAllocatable += hostData.MemAllocatable
			metric.DiskUsage += hostData.DiskUsage
			metric.DiskTotal += hostData.DiskTotal
		}
		group.Metric = metric
	} else {
		var innerGroups []*groupHostData
		groupHosts := make(map[string][]*hostData)
		key := groups[index]
		for _, host := range hosts {
			val := getFieldValue(key, host)
			groupHosts[val] = append(groupHosts[val], host)
		}

		for val, g := range groupHosts {
			innerGroup := chs.parseGroupHost(g, groups, index+1)
			if innerGroup == nil {
				continue
			}
			innerGroup.Name = val
			innerGroups = append(innerGroups, innerGroup)
		}
		if len(innerGroups) == 0 {
			return nil
		}
		group.Groups = innerGroups

		metric := new(groupHostMetric)
		for _, innerGroup := range innerGroups {
			metric.Machines += innerGroup.Metric.Machines
			metric.CPUUsage += innerGroup.Metric.CPUUsage
			metric.CPURequest += innerGroup.Metric.CPURequest
			metric.CPULimit += innerGroup.Metric.CPULimit
			metric.CPUOrigin += innerGroup.Metric.CPUOrigin
			metric.CPUTotal += innerGroup.Metric.CPUTotal
			metric.CPUAllocatable += innerGroup.Metric.CPUAllocatable
			metric.MemUsage += innerGroup.Metric.MemUsage
			metric.MemRequest += innerGroup.Metric.MemRequest
			metric.MemLimit += innerGroup.Metric.MemLimit
			metric.MemOrigin += innerGroup.Metric.MemOrigin
			metric.MemTotal += innerGroup.Metric.MemTotal
			metric.MemAllocatable += innerGroup.Metric.MemAllocatable
			metric.DiskUsage += innerGroup.Metric.DiskUsage
			metric.DiskTotal += innerGroup.Metric.DiskTotal
		}
		group.Metric = metric
	}
	return group
}

func (chs *ClickhouseSource) convertKey(key string) string {
	ckKeys := map[string]string{
		cpus:    ncpus,
		cluster: clusterName,
		host:    hostIP,
	}
	if ckKey, ok := ckKeys[key]; ok {
		return ckKey
	}
	return key
}

func interfaceToString(obj interface{}) string {
	str, ok := obj.(string)
	if ok {
		return str
	}
	return fmt.Sprintf("%s", obj)
}

func mapToSlice(mp map[string]struct{}) []string {
	var res []string
	for key := range mp {
		res = append(res, key)
	}
	return res
}

func parseHostRow(row *hostRow) *hostData {
	tags := make(map[string]string)
	if len(row.TagKeys) != len(row.TagValues) {
		return nil
	}
	for i := 0; i < len(row.TagKeys); i++ {
		tags[row.TagKeys[i]] = row.TagValues[i]
	}

	loadPercent := 0.0
	cpu, err := strconv.ParseFloat(tags[ncpus], 64)
	if err == nil && cpu > 1e-6 {
		loadPercent = row.Load5 * 100 / cpu
	}
	return &hostData{
		ClusterName:      tags[clusterName],
		IP:               tags[hostIP],
		Hostname:         tags[hostName],
		OS:               tags[os],
		KernelVersion:    tags[kernelVersion],
		Labels:           row.Labels,
		Tasks:            row.TaskContainers,
		CPU:              tags[ncpus],
		Mem:              tags[mem],
		CPUUsage:         row.CPUCoresUsage,
		CPURequest:       row.CPURequestTotal,
		CPULimit:         row.CPULimitTotal,
		CPUOrigin:        row.CPUOriginTotal,
		CPUTotal:         row.CPUTotal,
		CPUAllocatable:   row.CPUAllocatable,
		MemUsage:         row.MemUsed,
		MemRequest:       row.MemRequestTotal,
		MemLimit:         row.MemLimitTotal,
		MemOrigin:        row.MemOriginTotal,
		MemTotal:         row.MemTotal,
		MemAllocatable:   row.MemAllocatable,
		DiskUsage:        row.DiskUsed,
		DiskTotal:        row.DiskTotal,
		Load1:            row.Load1,
		Load5:            row.Load5,
		Load15:           row.Load15,
		CPUUsagePercent:  row.CPUUsageActive,
		MemUsagePercent:  row.MemUsedPercent,
		DiskUsagePercent: row.DiskUsedPercent,
		LoadPercent:      loadPercent,
		CPUDispPercent:   row.CPURequestPercent,
		MemDispPercent:   row.MemRequestPercent,
	}
}

func getFieldValue(field string, host *hostData) string {
	switch field {
	case mem:
		return host.Mem
	case ncpus, cpus:
		return host.CPU
	case cluster, clusterName:
		return host.ClusterName
	}
	return ""
}
