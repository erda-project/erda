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
	"github.com/erda-project/erda/internal/tools/monitor/core/entity"
	"github.com/erda-project/erda/internal/tools/monitor/core/entity/storage"
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
	entities := []entity.GroupedEntity{
		{
			Timestamp:       time.Now(),
			UpdateTimestamp: time.Now(),
			Type:            "err_exception",
			Key:             "8f8cb1453e5165e280d34c4a5a49d5ce",
			Values: map[string]string{
				"service_instance_id": "'string_value:\"f6bc8bd3-0485-4b79-a25f-0f31a05cxxxx\"'",
				"runtime_id":          "'string_value:\"12345\"'",
				"org_name":            "'string_value:\"test\"'",
			},
			Labels: map[string]string{
				"applicationId": "1234",
				"serviceName":   "testSvc",
				"terminusKey":   "5030d0d1a505db773fecf5049f67xxxx",
			},
		},
	}

	data, err := json.Marshal(entities)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &dest)
}

func TestProvider_ListEntities(t *testing.T) {
	p := &provider{
		Cfg:        &config{QueryTimeout: time.Second},
		Log:        &MockLogger{},
		Loader:     &MockLoader{},
		clickhouse: &MockClickhouseInterface{},
	}

	if _, _, err := p.ListEntities(context.Background(), &storage.ListOptions{
		Type: "err_exception",
		Labels: map[string]string{
			"applicationId": "1234",
		},
		Limit:                 200,
		UpdateTimeUnixNanoMin: 1662002666968000000,
		UpdateTimeUnixNanoMax: 1662002666968000000,
		Debug:                 true,
	}); err != nil {
		t.Fatal(err)
	}
}
