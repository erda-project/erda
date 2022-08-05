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

package details_apis

import (
	"context"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/clickhouse"
	"github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/http/httputil"
)

const timeFormat = "2006-01-02 15:04:05"

type ClickhouseSource struct {
	Clickhouse clickhouse.Interface
	Log        logs.Logger
	DebugSQL   bool
	Loader     loader.Interface
	Org        org.ClientInterface
}

type podRow struct {
	TagKeys           []string  `ch:"tagKeys"`
	TagValues         []string  `ch:"tagValues"`
	NumberFieldKeys   []string  `ch:"numberFieldKeys"`
	NumberFieldValues []float64 `ch:"numberFieldValues"`
	TerminatedReason  string    `ch:"terminatedReason"`
}

type containerRow struct {
	ContainerID string `ch:"containerID"`
	HostIP      string `ch:"hostIP"`
}

func (chs *ClickhouseSource) GetPodInfo(ctx context.Context, clusterName, podName string, start, end int64) (*PodInfo, error) {
	orgID := apis.GetOrgID(ctx)
	if orgID == "" {
		return nil, errors.Errorf("orgID can not be empty")
	}
	ctx = transport.WithHeader(ctx, metadata.New(map[string]string{httputil.InternalHeader: "monitor"}))
	org, err := chs.Org.GetOrg(ctx, &pb.GetOrgRequest{IdOrName: orgID})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get org")
	}

	pod, err := chs.getPod(ctx, org.Data.Name, clusterName, podName, start, end)
	if err != nil {
		return nil, err
	}

	containers, err := chs.getContainers(ctx, org.Data.Name, clusterName, podName, start, end)
	if err != nil {
		return nil, err
	}
	return chs.parsePodInfo(pod, containers)
}

func (chs *ClickhouseSource) getPod(ctx context.Context, orgName, clusterName, podName string, start, end int64) (*podRow, error) {
	table, _ := chs.Loader.GetSearchTable(orgName)
	from := time.UnixMilli(start).Local().Format(timeFormat)
	to := time.UnixMilli(end).Local().Format(timeFormat)
	sql := goqu.From(table).Select(
		goqu.L("argMax(tag_keys, timestamp)").As("tagKeys"),
		goqu.L("argMax(tag_values, timestamp)").As("tagValues"),
		goqu.L("argMax(number_field_keys, timestamp)").As("numberFieldKeys"),
		goqu.L("argMax(number_field_values, timestamp)").As("numberFieldValues"),
		goqu.L("argMax(string_field_values[indexOf(string_field_keys,?)], timestamp)", "terminated_reason").As("terminatedReason"),
	).Where(
		goqu.Ex{
			"metric_group": "kubernetes_pod_container",
			"org_name":     orgName,
			"tenant_id":    orgName,
		},
		goqu.L("tag_values[indexOf(tag_keys,?)]", "pod_name").Eq(podName),
		goqu.L("tag_values[indexOf(tag_keys,?)]", "cluster_name").Eq(clusterName),
		goqu.L("timestamp").Between(goqu.Range(from, to)),
	).GroupBy(goqu.L("tag_values[indexOf(tag_keys,?)]", "pod_name"))

	sqlStr, err := chs.toSQL(sql)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert sql")
	}

	row, err := chs.Clickhouse.Client().Query(ctx, sqlStr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query pod row")
	}

	row.Next()
	pod := podRow{}
	if err := row.ScanStruct(&pod); err != nil {
		return nil, errors.Wrapf(err, "failed to scan row to podRow")
	}
	return &pod, nil
}

func (chs *ClickhouseSource) getContainers(ctx context.Context, orgName, clusterName, podName string, start, end int64) ([]*containerRow, error) {
	table, _ := chs.Loader.GetSearchTable(orgName)
	from := time.UnixMilli(start).Local().Format(timeFormat)
	to := time.UnixMilli(end).Local().Format(timeFormat)
	sql := goqu.From(table).Select(
		goqu.L("any(tag_values[indexOf(tag_keys,?)])", "container_id").As("containerID"),
		goqu.L("any(tag_values[indexOf(tag_keys,?)])", "host_ip").As("hostIP"),
	).Where(
		goqu.Ex{
			"metric_group": "docker_container_summary",
			"org_name":     orgName,
			"tenant_id":    orgName,
		},
		goqu.L("tag_values[indexOf(tag_keys,?)]", "cluster_name").Eq(clusterName),
		goqu.L("tag_values[indexOf(tag_keys,?)]", "pod_name").Eq(podName),
		goqu.L("tag_values[indexOf(tag_keys,?)]", "podsandbox").Neq("true"),
		goqu.L("timestamp").Between(goqu.Range(from, to)),
	).GroupBy(goqu.L("tag_values[indexOf(tag_keys,?)]", "container_id"))

	sqlStr, err := chs.toSQL(sql)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert sql")
	}

	rows, err := chs.Clickhouse.Client().Query(ctx, sqlStr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query container row")
	}

	var containers []*containerRow
	for rows.Next() {
		container := containerRow{}
		if err := rows.ScanStruct(&container); err != nil {
			return nil, errors.Wrapf(err, "failed to san containerRow")
		}
		containers = append(containers, &container)
	}
	return containers, nil
}

func (chs *ClickhouseSource) toSQL(sql *goqu.SelectDataset) (string, error) {
	str, _, err := sql.ToSQL()
	if err != nil {
		return "", errors.Errorf("failed to convert to sql: %v", err)
	}
	if chs.DebugSQL {
		chs.Log.Infof("Metrics clickhouse SQL: \n%s", str)
	}
	return str, nil
}

func (chs *ClickhouseSource) parsePodInfo(pod *podRow, containers []*containerRow) (*PodInfo, error) {
	if len(pod.TagKeys) != len(pod.TagValues) || len(pod.NumberFieldKeys) != len(pod.NumberFieldValues) {
		return nil, errors.New("invalid pod")
	}
	tags := make(map[string]string)
	for i := 0; i < len(pod.TagKeys); i++ {
		tags[pod.TagKeys[i]] = pod.TagValues[i]
	}

	numbers := make(map[string]float64)
	for i := 0; i < len(pod.NumberFieldKeys); i++ {
		numbers[pod.NumberFieldKeys[i]] = pod.NumberFieldValues[i]
	}
	podInfo := &PodInfo{
		Summary: PodInfoSummary{
			ClusterName:      tags["cluster_name"],
			NodeName:         tags["node_name"],
			HostIP:           tags["host_ip"],
			Namespace:        tags["namespace"],
			PodName:          tags["pod_name"],
			RestartTotal:     numbers["restarts_total"],
			StateCode:        numbers["state_code"],
			TerminatedReason: pod.TerminatedReason,
		},
	}

	for _, container := range containers {
		podInfo.Instances = append(podInfo.Instances, &PodInfoInstanse{
			ContainerID: container.ContainerID,
			HostIP:      container.HostIP,
		})
	}
	return podInfo, nil
}
