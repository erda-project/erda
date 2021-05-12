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

package slsimport

import (
	"fmt"
	"testing"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/providers/elasticsearch"
	"github.com/erda-project/erda/modules/monitor/core/logs"
)

func TestRDSProcess_withFilter1(t *testing.T) {
	groups := rdsGroups()
	consumer := testConsumers()

	logFilterMap = map[string][]LogFilter{
		"rds": {&RDSLogFilter{SlowSQLThreshold: time.Microsecond * 600}},
	}
	consumer.rdsProcess(0, groups)
	assert.Equal(t, 2, consumer.outputs.kafka.(*fakeWriter).count)
	assert.Equal(t, 2, consumer.outputs.es.(*fakeWriter).count)
}

func TestRDSProcess_withFilter2(t *testing.T) {
	groups := rdsGroups()
	consumer := testConsumers()

	logFilterMap = map[string][]LogFilter{
		"rds": {&RDSLogFilter{SlowSQLThreshold: time.Microsecond * 100}},
	}
	consumer.rdsProcess(0, groups)
	assert.Equal(t, 3, consumer.outputs.kafka.(*fakeWriter).count)
	assert.Equal(t, 3, consumer.outputs.es.(*fakeWriter).count)
}

func TestRDSProcess_withFilter3(t *testing.T) {
	groups := rdsGroups()
	consumer := testConsumers()

	logFilterMap = map[string][]LogFilter{
		"rds": {&RDSLogFilter{SlowSQLThreshold: time.Microsecond * 100, ExcludeSQL: []string{"logout!"}}},
	}
	consumer.rdsProcess(0, groups)
	assert.Equal(t, 2, consumer.outputs.kafka.(*fakeWriter).count)
	assert.Equal(t, 2, consumer.outputs.es.(*fakeWriter).count)
}

type fakeWriter struct {
	count int
}

func (fw *fakeWriter) Write(data interface{}) error {
	defer func() {
		fw.count += 1
	}()
	switch data.(type) {
	case *elasticsearch.Document:
		fmt.Printf("%+v\n", data.(*elasticsearch.Document).Data.(*logs.Log).Content)
	default:
		fmt.Println(data)
	}
	return nil
}

func (fw *fakeWriter) WriteN(data ...interface{}) (int, error) {
	defer func() {
		fw.count += len(data)
	}()
	fmt.Println(len(data))
	return 0, nil
}

func (fw *fakeWriter) Close() error {
	return nil
}

func newStringPtr(s string) *string {
	return &s
}
func newu32intPtr(s uint32) *uint32 {
	return &s
}

func rdsGroups() *sls.LogGroupList {
	lg := sls.LogGroup{
		Logs: []*sls.Log{
			{
				Time: newu32intPtr(1597133902),
				Contents: []*sls.LogContent{
					{Key: newStringPtr("instance_id"), Value: newStringPtr("rm-bp17ar40w6824r8m0")},
					{Key: newStringPtr("thread_id"), Value: newStringPtr("4134281")},
					{Key: newStringPtr("origin_time"), Value: newStringPtr("1597133902400469")},
					{Key: newStringPtr("latency"), Value: newStringPtr("486")},
					{Key: newStringPtr("client_ip"), Value: newStringPtr("10.167.0.242")},
					{Key: newStringPtr("user"), Value: newStringPtr("dice")},
					{Key: newStringPtr("db"), Value: newStringPtr("db")},
					{Key: newStringPtr("fail"), Value: newStringPtr("0")},
					{Key: newStringPtr("return_rows"), Value: newStringPtr("0")},
					{Key: newStringPtr("update_rows"), Value: newStringPtr("0")},
					{Key: newStringPtr("check_rows"), Value: newStringPtr("0")},
					{Key: newStringPtr("sql"), Value: newStringPtr("COMMIT")},
				},
			},
			{
				Time: newu32intPtr(1597134902),
				Contents: []*sls.LogContent{
					{Key: newStringPtr("instance_id"), Value: newStringPtr("rm-bp17ar40w6824r8m0")},
					{Key: newStringPtr("thread_id"), Value: newStringPtr("4134281")},
					{Key: newStringPtr("origin_time"), Value: newStringPtr("1597133902400469")},
					{Key: newStringPtr("latency"), Value: newStringPtr("1000")},
					{Key: newStringPtr("client_ip"), Value: newStringPtr("10.167.0.242")},
					{Key: newStringPtr("user"), Value: newStringPtr("dice")},
					{Key: newStringPtr("db"), Value: newStringPtr("db")},
					{Key: newStringPtr("fail"), Value: newStringPtr("0")},
					{Key: newStringPtr("return_rows"), Value: newStringPtr("0")},
					{Key: newStringPtr("update_rows"), Value: newStringPtr("0")},
					{Key: newStringPtr("check_rows"), Value: newStringPtr("0")},
					{Key: newStringPtr("sql"), Value: newStringPtr("COMMIT")},
				},
			},
			{
				Time: newu32intPtr(1597134902),
				Contents: []*sls.LogContent{
					{Key: newStringPtr("instance_id"), Value: newStringPtr("rm-bp17ar40w6824r8m0")},
					{Key: newStringPtr("thread_id"), Value: newStringPtr("4134281")},
					{Key: newStringPtr("origin_time"), Value: newStringPtr("1597133902400469")},
					{Key: newStringPtr("latency"), Value: newStringPtr("1000")},
					{Key: newStringPtr("client_ip"), Value: newStringPtr("10.167.0.242")},
					{Key: newStringPtr("user"), Value: newStringPtr("dice")},
					{Key: newStringPtr("db"), Value: newStringPtr("db")},
					{Key: newStringPtr("fail"), Value: newStringPtr("0")},
					{Key: newStringPtr("return_rows"), Value: newStringPtr("0")},
					{Key: newStringPtr("update_rows"), Value: newStringPtr("0")},
					{Key: newStringPtr("check_rows"), Value: newStringPtr("0")},
					{Key: newStringPtr("sql"), Value: newStringPtr("logout!")},
				},
			},
		},
		Topic: newStringPtr("rds_audit_log"),
	}
	return &sls.LogGroupList{
		LogGroups: []*sls.LogGroup{&lg},
	}
}

func testConsumers() *Consumer {
	return &Consumer{
		ai: &AccountInfo{
			OrgName: "terminus",
			OrgID:   "1",
		},
		project:  "rds-test-xxxx",
		logStore: "slow-sql-test",
		outputs: &outputs{
			es:          &fakeWriter{},
			kafka:       &fakeWriter{},
			indexPrefix: "xxxxx-pre",
		},
	}
}
