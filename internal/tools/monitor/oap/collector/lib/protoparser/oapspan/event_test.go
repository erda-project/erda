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

package oapspan

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
)

func Test_span_event_unmarshal(t *testing.T) {
	type fields struct {
		buf []byte
		err error
	}

	tests := []struct {
		name    string
		fields  fields
		want    []metric.Metric
		wantErr bool
	}{
		{
			name: "normal trace event",
			fields: fields{
				buf: []byte(`{
    "traceID": "0ad6bfb29fb19c13a052c5d8f5f15cb6",
    "spanID": "f7ae4819e2f9d85f",
    "parentSpanID": "bda0d9faf27c1a19",
    "startTimeUnixNano": 1659682795021103000,
    "endTimeUnixNano": 1659682795314559958,
    "name": "metric.clickhouse",
    "relations": null,
    "attributes": {
        "instrument": "executive",
        "metrics": "apm_span_event",
        "org_name": "",
        "result": "total: 5, interval: 0, columns: tttt,_metric_scope::tag,host_ip::tag,org_name::tag,span_id::tag,terminus_key::tag,_meta::tag,event::tag,host::tag,cluster_name::tag,trace_id::tag,name,timestamp,_metric_scope_id::tag,field_count::field, rows: [0x140012e3590 0x14001484f00 0x14001484f50 0x14001484fa0 0x14001484ff0 0x14001485090 0x140014850e0 0x14001485130 0x14001485180 0x140014851d0 0x14001485220 0x14001485270 0x140014852c0 0x14001485310 0x14001485360]",
        "result_total": "5",
        "settings_max_execution_time": "65",
        "span_kind": "local",
        "sql": "SELECT *, tag_values[indexOf(tag_keys,'event')] AS \"event\" FROM \"monitor\".\"metrics_all\" WHERE ((\"timestamp\" \u003e= fromUnixTimestamp64Nano(cast(0,'Int64'))) AND (\"timestamp\" \u003c fromUnixTimestamp64Nano(cast(1659682795020000000,'Int64'))) AND (tag_values[indexOf(tag_keys,'span_id')] = '4a15deb5-7231-4b64-905a-c772f6e2fe33') AND (\"metric_group\" IN ('apm_span_event'))) ORDER BY \"timestamp\" ASC LIMIT 10",
        "table": "monitor.metrics_all"
    },
    "events": [
        {
            "timeUnixNano": 1659682795305042000,
            "name": "profile_info",
            "attributes": {
                "metrics": "apm_span_event",
                "org_name": "",
                "result": "total: 5, interval: 0, columns: tttt,_metric_scope::tag,host_ip::tag,org_name::tag,span_id::tag,terminus_key::tag,_meta::tag,event::tag,host::tag,cluster_name::tag,trace_id::tag,name,timestamp,_metric_scope_id::tag,field_count::field, rows: [0x140012e3590 0x14001484f00 0x14001484f50 0x14001484fa0 0x14001484ff0 0x14001485090 0x140014850e0 0x14001485130 0x14001485180 0x140014851d0 0x14001485220 0x14001485270 0x140014852c0 0x14001485310 0x14001485360]",
                "result_total": "5",
                "settings.max_execution_time": "65",
                "sql": "SELECT *, tag_values[indexOf(tag_keys,'event')] AS \"event\" FROM \"monitor\".\"metrics_all\" WHERE ((\"timestamp\" \u003e= fromUnixTimestamp64Nano(cast(0,'Int64'))) AND (\"timestamp\" \u003c fromUnixTimestamp64Nano(cast(1659682795020000000,'Int64'))) AND (tag_values[indexOf(tag_keys,'span_id')] = '4a15deb5-7231-4b64-905a-c772f6e2fe33') AND (\"metric_group\" IN ('apm_span_event'))) ORDER BY \"timestamp\" ASC LIMIT 10",
                "table": "monitor.metrics_all",
                "tenant_id": ""
            },
            "droppedAttributesCount": 0
        }
    ]
}`),
			},
			want: []metric.Metric{
				{
					Name:      "profile_info",
					Timestamp: int64(1659682795305042000),
					Tags: map[string]string{
						"instrument":                  "executive",
						"metrics":                     "apm_span_event",
						"org_name":                    "",
						"result":                      "total: 5, interval: 0, columns: tttt,_metric_scope::tag,host_ip::tag,org_name::tag,span_id::tag,terminus_key::tag,_meta::tag,event::tag,host::tag,cluster_name::tag,trace_id::tag,name,timestamp,_metric_scope_id::tag,field_count::field, rows: [0x140012e3590 0x14001484f00 0x14001484f50 0x14001484fa0 0x14001484ff0 0x14001485090 0x140014850e0 0x14001485130 0x14001485180 0x140014851d0 0x14001485220 0x14001485270 0x140014852c0 0x14001485310 0x14001485360]",
						"result_total":                "5",
						"settings_max_execution_time": "65",
						"span_kind":                   "local",
						"settings.max_execution_time": "65",
						"sql":                         "SELECT *, tag_values[indexOf(tag_keys,'event')] AS \"event\" FROM \"monitor\".\"metrics_all\" WHERE ((\"timestamp\" \u003e= fromUnixTimestamp64Nano(cast(0,'Int64'))) AND (\"timestamp\" \u003c fromUnixTimestamp64Nano(cast(1659682795020000000,'Int64'))) AND (tag_values[indexOf(tag_keys,'span_id')] = '4a15deb5-7231-4b64-905a-c772f6e2fe33') AND (\"metric_group\" IN ('apm_span_event'))) ORDER BY \"timestamp\" ASC LIMIT 10",
						"table":                       "monitor.metrics_all",
						"tenant_id":                   "",
						"trace_id":                    "0ad6bfb29fb19c13a052c5d8f5f15cb6",
						"span_id":                     "f7ae4819e2f9d85f",
					},
					Fields:  nil,
					OrgName: "",
				},
			},
			wantErr: false,
		},
		{
			name: "",
			fields: fields{
				buf: []byte(`"traceID":"bbb","spanID":"aaa","parentSpanID":"","startTimeUnixNano":1652756014793553000,"endTimeUnixNano":1652756014793553000,"name":"GET /","relations":null,"attributes":{"hello":"world","org_name":"erda"}}`),
			},
			want:    []metric.Metric{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uw := &unmarshalEventWork{
				buf: tt.fields.buf,
				err: tt.fields.err,
				callback: func(metrics []metric.Metric) error {
					require.Equal(t, len(tt.want), len(metrics))
					for i, want := range tt.want {
						require.Equal(t, want.OrgName, metrics[i].OrgName)
						require.Equal(t, want.Timestamp, metrics[i].Timestamp)

						for k, v := range want.Tags {
							require.Equal(t, v, metrics[i].Tags[k])
						}
						if want.Fields != nil {
							for k, v := range want.Fields {
								require.Equal(t, v, metrics[i].Fields[k])
							}
						}

					}
					return nil
				},
			}
			uw.wg.Add(1)
			uw.Unmarshal()
			uw.wg.Wait()
			if !tt.wantErr {
				require.Nil(t, uw.err)
			} else {
				require.Error(t, uw.err)
			}
		})
	}
}
