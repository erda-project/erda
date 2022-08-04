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
	"encoding/json"
	"testing"

	ckdriver "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/clickhouse"
	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"
	"github.com/erda-project/erda/pkg/http/httputil"
)

type MockLogger struct {
	logs.Logger
	t testing.T
}

func (log *MockLogger) Infof(template string, args ...interface{}) {
	logrus.Printf(template, args...)
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
	case mockPodRow:
		return &MockPodRow{
			pods: []podRow{
				{
					TagKeys:          []string{"_meta", "_metric_scope", "_metric_scope_id", "cluster_name", "container_name", "host", "host_ip", "namespace", "node_name", "org_name", "pod_name", "state"},
					TagValues:        []string{"true", "org", "demo", "erda-test", "demo-242891120181430", "test-6wmj8", "2.2.2.2", "test-242891097387190", "cn-hangzhou.2.2.2.2", "demo", "demo-xxxx", "terminated"},
					RestartTotal:     2,
					StateCode:        1,
					TerminatedReason: "Completed",
				},
			},
		}, nil
	case mockContainerRow:
		return &MockContainer{
			containers: []containerRow{
				{
					ContainerID: "1",
					HostIP:      "1.1.1.1",
				},
				{
					ContainerID: "2",
					HostIP:      "2.2.2.2",
				},
			},
		}, nil
	}
	return nil, nil
}

type MockPodRow struct {
	ckdriver.Rows
	index int
	pods  []podRow
}

func (p *MockPodRow) Next() bool {
	if p.index+1 == len(p.pods) {
		return false
	}
	p.index++
	return true
}

func (p *MockPodRow) ScanStruct(dest interface{}) error {
	data, err := json.Marshal(p.pods[p.index])
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (p *MockPodRow) Close() error {
	return nil
}

type MockContainer struct {
	ckdriver.Rows
	index      int
	containers []containerRow
}

func (c *MockContainer) Next() bool {
	if c.index+1 == len(c.containers) {
		return false
	}
	c.index++
	return true
}

func (c *MockContainer) ScanStruct(dest interface{}) error {
	data, err := json.Marshal(c.containers[c.index])
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (c *MockContainer) Close() error {
	return nil
}

type (
	mockKey  string
	mockType string
)

var (
	key              mockKey  = "mockType"
	mockPodRow       mockType = "mockPodRow"
	mockContainerRow mockType = "mockContainerRow"
)

func TestClickhouseSource_GetPodInfo(t *testing.T) {
	chs := &ClickhouseSource{
		Clickhouse: &MockClickhouseInterface{},
		Log:        &MockLogger{},
		DebugSQL:   true,
		Loader:     &MockLoader{},
		Org:        &MockOrgServiceServer{},
	}

	ctx := transport.WithHeader(context.Background(), metadata.New(map[string]string{httputil.OrgHeader: "1"}))
	ctx = context.WithValue(ctx, key, mockContainerRow)
	if _, err := chs.GetPodInfo(ctx, "erda-test", "test", 0, 0); err != nil {
		t.Fatal(err)
	}
}
