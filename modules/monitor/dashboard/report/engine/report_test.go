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

package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func setup() {
	logrus.SetLevel(logrus.DebugLevel)
	os.Setenv("ACTION_REPORT_ID", "1")
	// os.Setenv("DICE_ORG_ID", "2")
	os.Setenv("ACTION_DOMAIN_ADDR", "http://terminus-org.dev.terminus.io")
	// os.Setenv("ACTION_SCOPE", "org")
	// os.Setenv("ACTION_SCOPE_ID", "2")
	os.Setenv("ACTION_ORG_NAME", "terminus")
}

func shutdown() {
	os.Clearenv()
}

func TestReport_FetchAndConvert(t *testing.T) {
	var mul interface{}
	json.Unmarshal([]byte(`[{
  "w": 24,
  "h": 9,
  "x": 0,
  "y": 37,
  "i": "d51",
  "view": {
    "title": "服务OOM次数Top10",
    "description": "",
    "chartType": "table",
    "dataSourceType": "static",
    "api": {
      "url": "/api/telemetry/docker_container_summary",
      "query": {
        "start": "before_24h",
        "end": "now",
        "filter_instance_type": "service",
        "field_eq_oomkilled":"b:false",
        "filter_org_name": "",
        "group": "tags.service_id",
        "cardinality":"tags.container_id",
        "sort": "cardinality_tags.container_id",
        "last": ["tags.service_name","tags.project_name"],
        "format":"chartv2",
        "limit": 10
      },
      "body": {},
      "method": "GET",
      "header": {}
    }
  }
}]`), &mul)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	type fields struct {
		Block             *blockEntity
		ReportTask        *reportTaskEntity
		DecodedViewConfig string
		Metrics           map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"multiple",
			fields{
				&blockEntity{ViewConfig: mul},
				&reportTaskEntity{baseEntity: baseEntity{ScopeID: "2", Scope: "org"}},
				"",
				nil,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{}
			r := &Report{
				resource: &Resource{
					Block:      tt.fields.Block,
					ReportTask: tt.fields.ReportTask,
				},
				cfg: cfg,
			}
			if err := r.CurrentFetchAndConvert(ctx); (err != nil) != tt.wantErr {
				t.Errorf("FetchAndConvert() error = %v, wantErr %v", err, tt.wantErr)
			}
			d, _ := json.Marshal(r.DataConfig)
			t.Log(string(d))
		})
	}
}

// func TestReport_createEventbox(t *testing.T) {
// 	// prepare
// 	cfg := &config{}
// 	r := New(cfg)
// 	r.resource = &Resource{
// 		ReportTask: &reportTaskEntity{
// 			Type:       "weekly",
// 			baseEntity: baseEntity{ID: 1},
// 		},
// 	}
// 	history := &historyEntity{baseEntity: baseEntity{ID: 2}}
//
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()
// 	res, err := r.createEventbox(ctx, nil, history)
// 	assert.Nil(t, err)
// 	t.Log(res)
// }

func TestRenderContent(t *testing.T) {
	params := &tmplParams{
		HistoryURL:  "xxx.io",
		ReportType:  "周报",
		DateDisplay: "04月5日-04月12日",
		OrgName:     "新华书店",
	}
	renderContent("## {{.OrgName}}{{.ReportType}}\n{{.DateDisplay}} 监控{{.ReportType}}已生成，[点击查看]({{.HistoryURL}})", params)
}

func TestXXX(t *testing.T) {
	m := map[string]interface{}{
		"a": []int{1, 2},
		"b": 1,
		"c": map[string]string{"hello": "world"},
		"d": "hello",
	}
	for _, v := range m {
		t.Log(fmt.Sprintf("%v", v))
	}
}
