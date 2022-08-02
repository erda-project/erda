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
	"encoding/json"
	"net/http"
	"testing"

	ckdriver "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/clickhouse"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/i18n"
	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"
)

type MockLogger struct {
	logs.Logger
	t testing.T
}

func (log *MockLogger) Errorf(template string, args ...interface{}) {
	log.t.Errorf(template, args...)
}

func (log *MockLogger) Infof(template string, args ...interface{}) {
	logrus.Printf(template, args...)
}

func (log *MockLogger) Warnf(template string, args ...interface{}) {
	logrus.Warnf(template, args...)
}

type MockOrgChecker struct {
}

func (o *MockOrgChecker) checkOrgByClusters(ctx httpserver.Context, clusters []*resourceCluster) error {
	return nil
}

type MockOrgServiceServer struct {
	orgpb.OrgServiceServer
}

func (o *MockOrgServiceServer) GetOrg(context.Context, *orgpb.GetOrgRequest) (*orgpb.GetOrgResponse, error) {
	return &orgpb.GetOrgResponse{Data: &orgpb.Org{Name: "test_org"}}, nil
}

type MockLoader struct {
	loader.Interface
}

func (l *MockLoader) GetSearchTable(tenant string) (string, *loader.TableMeta) {
	return "monitor.metric_test_search", nil
}

type MockClickhouseInterface struct {
	clickhouse.Interface
}

func (c *MockClickhouseInterface) Client() ckdriver.Conn {
	return &MockClickhouseConn{}
}

type MockClickhouseConn struct {
	ckdriver.Conn
}

func (conn *MockClickhouseConn) Query(ctx context.Context, query string, args ...interface{}) (ckdriver.Rows, error) {
	typ, ok := ctx.Value(key).(mockType)
	if !ok {
		return nil, nil
	}
	switch typ {
	case mockContainerRow:
		return &MockContainerRow{containerRows: []containerRow{
			{
				ContainerID: "1d817720053c059cdea3de6e5785c34cab6bfed0e73e2c33a78a1c85c776e24b",
				TagKeys:     []string{"_meta", "_metric_scope", "_metric_scope_id", "cluster_name", "container", "container_id", "container_name", "host_ip", "id", "image", "interface", "node", "org_name", "pod_ip", "pod_name", "pod_namespace", "pod_source", "pod_uid", "service_instance_id"},
				TagValues:   []string{"true", "org", "erda", "terminus-test", "POD", "1d817720053c059cdea3de6e5785c34cab6bfed0e73e2c33a78a1c85c776e24b", "POD", "10.112.2.8", "/kubepods/burstable/pod07056557-7669-49ff-bb08-357653a5bf3c/1d817720053c059cdea3de6e5785c34cab6bfed0e73e2c33a78a1c85c776e24b", "registry.cn-hangzhou.aliyuncs.com/acs/test:1.0", "eth0", "virtual-kubelet-cn-hangzhou-k", "erda", "10.0.6.138", "sourcecov-operator-564f84c599-cm6zn", "default", "eci", "07056557-7669-49ff-bb08-357653axxxxxx", "07056557-7669-49ff-bb08-357653xxxxx"},
				CpuUsage:    0.036069506,
				CpuRequest:  0.01,
				CpuLimit:    0.1,
				CpuOrigin:   0,
				MemUsage:    228823040,
				MemRequest:  314572800,
				MemLimit:    524288000,
				MemOrigin:   0,
				DiskUsage:   0,
				DiskLimit:   0,
			},
			{
				ContainerID: "1",
				TagKeys:     []string{"_meta", "_metric_scope", "_metric_scope_id", "cluster_name", "container", "container_id", "device", "host_ip", "id", "image", "node", "org_name", "pod_ip", "pod_name", "pod_namespace", "pod_uid", "service_instance_id"},
				TagValues:   []string{"true", "org", "erda", "terminus-dev", "config-reloader", "8c32a9817993904c9f353e0e3c34ad607f2e57412ae6db8be356afefe7a81c4d", "/dev/vdb1", "10.0.6.216", "/kubepods/burstable/pod005e729b-c330-44e1-b4a4-f0c9799efb02/8c32a9817993904c9f353e0e3c34ad607f2e57412ae6db8be356afefe7a81c4d", "sha256:18cec637a88f705f4010ed953e1949bede93b1a970d1cb08b854b9353c450c26", "node-010000006216", "erda", "10.112.1.9", "prometheus-prometheus-0", "default", "005e729b-c330-44e1-b4a4-f0c9799efb02", "005e729b-c330-44e1-b4a4-f0c9799efb02"},
				CpuUsage:    0.049031729,
				CpuRequest:  0.5,
				CpuLimit:    0.5,
				CpuOrigin:   0,
				MemUsage:    228823040,
				MemRequest:  536870912,
				MemLimit:    536870912,
				MemOrigin:   0,
				DiskUsage:   0,
				DiskLimit:   0,
			},
		}}, nil
	case mockHostType:
		return &MockHostTypeRow{hostTypeRows: []hostTypeRow{
			{
				ClusterName: "cluster1",
				CPUs:        "8",
				Mem:         "32",
				HostIP:      "1.1.1.1",
				Labels:      "test",
			},
			{
				ClusterName: "cluster2",
				CPUs:        "16",
				Mem:         "32",
				HostIP:      "2.2.2.2",
				Labels:      "test",
			},
		}}, nil
	case mockHosts:
		return &MockHostRow{hostRows: []hostRow{
			{
				TagKeys:           []string{"_meta", "_metric_scope", "_metric_scope_id", "cluster_name", "host", "host_ip", "hostname", "kernel_version", "mem", "n_cpus", "org_name", "os"},
				TagValues:         []string{"true", "org", "erda", "terminus-test", "telegraf-6fc3aaf6af-2gbqx", "10.0.6.111", "node-010000006111", "4.18.0-305.3.1.el8.x86_64", "31", "8", "erda", "centos 8.4.2105"},
				Labels:            "location-cluster-service,minio_single,workspace-prod,lb,bigdata-job,pack-job,benchmark-test,job,stateful-service,workspace-staging,platform,stateless-service,workspace-test,workspace-dev,org-erda",
				TaskContainers:    54,
				CPUCoresUsage:     7.99,
				CPURequestTotal:   6.78,
				CPULimitTotal:     44.55,
				CPUOriginTotal:    6.78,
				CPUTotal:          8,
				CPUAllocatable:    8,
				MemUsed:           29547917312,
				MemRequestTotal:   22648193024,
				MemLimitTotal:     83152807936,
				MemOriginTotal:    22648193024,
				MemTotal:          32965337088,
				MemAllocatable:    32831119360,
				DiskUsed:          11181203456,
				DiskTotal:         42938118144,
				Load1:             576.28,
				Load5:             479.46,
				Load15:            251.66,
				CPUUsageActive:    99.98643094854494,
				MemUsedPercent:    89.6332933988289,
				DiskUsedPercent:   26.040273629370542,
				CPURequestPercent: 84.74999999999997,
				MemRequestPercent: 68.98391972463043,
			},
			{
				TagKeys:           []string{"_meta", "_metric_scope", "_metric_scope_id", "cluster_name", "host", "host_ip", "hostname", "kernel_version", "mem", "n_cpus", "org_name", "os"},
				TagValues:         []string{"true", "org", "erda", "terminus-test", "telegraf-6fc3aaf6af-2gbqc", "10.0.6.222", "node-010000006222", "4.18.0-305.3.1.el8.x86_64", "31", "8", "erda", "centos 8.4.2105"},
				Labels:            "stateful-service,bigdata-job,benchmark-test,location-cluster-service,job,workspace-staging,workspace-dev,lb,pack-job,org-erda,minio_single,org-terminus,workspace-test,stateless-service,hugegraph,workspace-prod,platform,xxx",
				TaskContainers:    78,
				CPUCoresUsage:     6.034839090727826,
				CPURequestTotal:   3.9800000000000004,
				CPULimitTotal:     23.200000000000003,
				CPUOriginTotal:    3.9800000000000004,
				CPUTotal:          8,
				CPUAllocatable:    8,
				MemUsed:           28685352960,
				MemRequestTotal:   11210326016,
				MemLimitTotal:     45013270528,
				MemOriginTotal:    11210326016,
				MemTotal:          32965337088,
				MemAllocatable:    32831119360,
				DiskUsed:          61075689472,
				DiskTotal:         107362627584,
				Load1:             14.7,
				Load5:             10.42,
				Load15:            7.82,
				CPUUsageActive:    75.43548863409782,
				MemUsedPercent:    87.01671359654321,
				DiskUsedPercent:   56.88729015524017,
				CPURequestPercent: 49.75000000000001,
				MemRequestPercent: 34.14542737052752,
			},
		}}, nil
	}
	return nil, nil
}

type (
	mockKey  string
	mockType string
)

var (
	key              mockKey  = "mockType"
	mockContainerRow mockType = "mockContainerRow"
	mockHostType     mockType = "mockHostType"
	mockHosts        mockType = "mockHosts"
)

type MockContainerRow struct {
	ckdriver.Rows
	index         int
	containerRows []containerRow
}

func (c *MockContainerRow) Next() bool {
	if c.index+1 == len(c.containerRows) {
		return false
	}
	c.index++
	return true
}

func (c *MockContainerRow) ScanStruct(dest interface{}) error {
	data, err := json.Marshal(c.containerRows[c.index])
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (c *MockContainerRow) Close() error {
	return nil
}

type MockHostTypeRow struct {
	ckdriver.Rows
	index        int
	hostTypeRows []hostTypeRow
}

func (h *MockHostTypeRow) Next() bool {
	if h.index+1 == len(h.hostTypeRows) {
		return false
	}
	h.index++
	return true
}

func (h *MockHostTypeRow) ScanStruct(dest interface{}) error {
	data, err := json.Marshal(h.hostTypeRows[h.index])
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (h *MockHostTypeRow) Close() error {
	return nil
}

type MockHostRow struct {
	ckdriver.Rows
	index    int
	hostRows []hostRow
}

func (g *MockHostRow) Next() bool {
	if g.index+1 == len(g.hostRows) {
		return false
	}
	g.index++
	return true
}

func (g *MockHostRow) ScanStruct(dest interface{}) error {
	data, err := json.Marshal(g.hostRows[g.index])
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (g *MockHostRow) Close() error {
	return nil
}

type MockContext struct {
	httpserver.Context
	req *http.Request
}

func (c *MockContext) Request() *http.Request {
	return c.req
}

type MockTran struct {
	i18n.Translator
}

func (t *MockTran) Text(lang i18n.LanguageCodes, key string) string {
	return key
}

type MockQueryServiceImpl struct {
}

func (q *MockQueryServiceImpl) queryStatus(clusterName string) (statuses []*statusDTO, err error) {
	return []*statusDTO{{
		Name:        "component1",
		DisplayName: "component1",
		Status:      0,
	}}, nil
}

func TestClickhouseSource_GetContainers(t *testing.T) {
	chs := ClickhouseSource{
		p:          &provider{Org: &MockOrgServiceServer{}},
		orgChecker: &MockOrgChecker{},
		Clickhouse: &MockClickhouseInterface{},
		Log:        &MockLogger{},
		DebugSQL:   true,
		Loader:     &MockLoader{},
	}

	req, err := http.NewRequestWithContext(context.WithValue(context.Background(), key, mockContainerRow), http.MethodPost, "unit.test", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := &MockContext{req: req}
	chs.GetContainers(ctx, req, struct {
		InstanceType string `param:"instance_type" validate:"required"`
		Start        int64  `query:"start"`
		End          int64  `query:"end"`
	}(struct {
		InstanceType string
		Start        int64
		End          int64
	}{InstanceType: "all"}), resourceRequest{
		Clusters: []*resourceCluster{
			{
				ClusterName: "cluster1",
				HostIPs:     []string{"1.1.1.1", "2.2.2.2"},
			},
		},
	})
}

func TestClickhouseSource_GetHostTypes(t *testing.T) {
	chs := ClickhouseSource{
		p:          &provider{t: &MockTran{}},
		Clickhouse: &MockClickhouseInterface{},
		Log:        &MockLogger{},
		DebugSQL:   true,
		Loader:     &MockLoader{},
	}
	req, err := http.NewRequestWithContext(context.WithValue(context.Background(), key, mockHostType), http.MethodPost, "unit.test", nil)
	if err != nil {
		t.Fatal(err)
	}

	chs.GetHostTypes(req, struct {
		ClusterName string `query:"clusterName" validate:"required"`
		OrgName     string `query:"orgName" validate:"required"`
	}(struct {
		ClusterName string
		OrgName     string
	}{ClusterName: "cluster", OrgName: "org"}))
}

func TestClickhouseSource_GetGroupHosts(t *testing.T) {
	chs := ClickhouseSource{
		p:          &provider{service: &MockQueryServiceImpl{}},
		Clickhouse: &MockClickhouseInterface{},
		Log:        &MockLogger{},
		DebugSQL:   true,
		Loader:     &MockLoader{},
	}
	req, err := http.NewRequestWithContext(context.WithValue(context.Background(), key, mockHosts), http.MethodPost, "unit.test", nil)
	if err != nil {
		t.Fatal(err)
	}

	chs.GetGroupHosts(req, struct {
		OrgName string `query:"orgName" validate:"required" json:"-"`
	}(struct{ OrgName string }{OrgName: "erda"}), resourceRequest{
		Clusters: []*resourceCluster{
			{
				ClusterName: "cluster1",
				HostIPs:     []string{"1.1.1.1", "2.2.2.2"},
			},
		},
		Filters: []*resourceFilter{
			{
				Key:    cpus,
				Values: []string{"8"},
			},
			{
				Key:    cpuCoresUsage,
				Values: []string{">=90%"},
			},
			{
				Key:    memUsedPercent,
				Values: []string{"40%-70%"},
			},
		},
		Groups: []string{"cluster", "cpus"},
	})
}
