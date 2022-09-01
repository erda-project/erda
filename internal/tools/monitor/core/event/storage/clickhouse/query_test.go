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

package clickhouse

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	ckdriver "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/clickhouse"
	"github.com/erda-project/erda/internal/tools/monitor/core/event"
	"github.com/erda-project/erda/internal/tools/monitor/core/event/storage"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"
)

type MockLogger struct {
	logs.Logger
	t testing.T
}

func (log *MockLogger) Infof(template string, args ...interface{}) {
	logrus.Printf(template, args...)
}

type MockLoader struct {
	loader.Interface
}

func (l *MockLoader) GetSearchTable(tenant string) (string, *loader.TableMeta) {
	return "monitor.events_all", nil
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

func (c *MockClickhouseConn) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	events := []event.Event{
		{
			EventID:   "7b8cf20b-d841-42bf-bd59-299f872d0c87",
			Name:      "",
			Kind:      "EVENT_KIND_ALERT",
			Content:   "Alert Content",
			Timestamp: 1662001331931000000,
			Tags: map[string]string{
				"display_url": "test.terminus.io/test/dataCenter/alarm/report/erda-test/cpu",
				"alert_title": "CPU Alert",
				"trigger":     "alert",
				"org_name":    "erda-test",
			},
			Relations: map[string]string{
				"res_id":   "1213_machine_cpu_zxjkys1qawnozw5nlte3mi4xni4xnzqunta",
				"res_type": "alert",
				"trace_id": "",
			},
		},
	}
	data, err := json.Marshal(events)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &dest)
}

func TestProvider_QueryPaged(t *testing.T) {
	p := &provider{
		Cfg:        &config{QueryTimeout: time.Second},
		Log:        &MockLogger{},
		Loader:     &MockLoader{},
		clickhouse: &MockClickhouseInterface{},
	}

	sel := &storage.Selector{
		Start: 1662001775935000000,
		End:   1662001788030000000,
		Filters: []*storage.Filter{
			{
				Key:   "event_id",
				Op:    storage.EQ,
				Value: "7b8cf20b-d841-42bf-bd59-299f872d0c87",
			},
			{
				Key:   "relations.res_type",
				Op:    storage.EQ,
				Value: "alert",
			},
			{
				Key:   "relations.res_id",
				Op:    storage.EQ,
				Value: "1213_machine_cpu_zxjkys1qawnozw5nlte3mi4xni4xnzqunta",
			},
			{
				Key:   "tags.org_name",
				Op:    storage.EQ,
				Value: "erda-test",
			},
		},
		Debug: true,
	}
	if _, err := p.QueryPaged(context.Background(), sel, 1, 200); err != nil {
		t.Fatal(err)
	}
}
