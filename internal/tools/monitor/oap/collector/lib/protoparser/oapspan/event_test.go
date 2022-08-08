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
	"fmt"
	"strings"
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
						"message":                     "table=monitor.metrics_all;tenant_id=;metrics=apm_span_event;org_name=;result=total: 5, interval: 0, columns: tttt,_metric_scope::tag,host_ip::tag,org_name::tag,span_id::tag,terminus_key::tag,_meta::tag,event::tag,host::tag,cluster_name::tag,trace_id::tag,name,timestamp,_metric_scope_id::tag,field_count::field, rows: [0x140012e3590 0x14001484f00 0x14001484f50 0x14001484fa0 0x14001484ff0 0x14001485090 0x140014850e0 0x14001485130 0x14001485180 0x140014851d0 0x14001485220 0x14001485270 0x140014852c0 0x14001485310 0x14001485360];result_total=5;settings.max_execution_time=65;sql=SELECT *, tag_values[indexOf(tag_keys,'event')] AS \"event\" FROM \"monitor\".\"metrics_all\" WHERE ((\"timestamp\" >= fromUnixTimestamp64Nano(cast(0,'Int64'))) AND (\"timestamp\" < fromUnixTimestamp64Nano(cast(1659682795020000000,'Int64'))) AND (tag_values[indexOf(tag_keys,'span_id')] = '4a15deb5-7231-4b64-905a-c772f6e2fe33') AND (\"metric_group\" IN ('apm_span_event'))) ORDER BY \"timestamp\" ASC LIMIT 10;",
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
		{
			name: "progress event",
			fields: fields{
				buf: []byte(`{
    "traceID": "01d56cc2630983060966d988c2d569e2",
    "spanID": "edce319ace93a8cb",
    "parentSpanID": "e78a9d647bb703ae",
    "startTimeUnixNano": 1659940670682674000,
    "endTimeUnixNano": 1659940670886532000,
    "name": "metric.clickhouse",
    "relations": null,
    "attributes": {
        "env_id": "",
        "instrument": "executive",
        "instrument_version": "",
        "metrics": "apm_span_event",
        "org_name": "",
        "result": "total: 5, interval: 0, columns: tttt,_meta::tag,host_ip::tag,timestamp,terminus_key::tag,trace_id::tag,field_count::field,_metric_scope_id::tag,span_id::tag,name,_metric_scope::tag,cluster_name::tag,event::tag,host::tag,org_name::tag, rows: [0x14001ce8a50 0x140019f7a90 0x140019f7ae0 0x140019f7b30 0x140019f7b80 0x140019f7c20 0x140019f7c70 0x140019f7cc0 0x140019f7d10 0x140019f7d60 0x140019f7db0 0x140019f7e00 0x140019f7e50 0x140019f7ea0 0x140019f7ef0]",
        "result_total": "5",
        "service_id": "monitor",
        "service_name": "monitor",
        "settings_max_execution_time": "65",
        "span_kind": "local",
        "span_layer": "local",
        "sql": "SELECT *, tag_values[indexOf(tag_keys,'event')] AS \"event\" FROM \"monitor\".\"metrics_all\" WHERE ((\"timestamp\" \u003e= fromUnixTimestamp64Nano(cast(0,'Int64'))) AND (\"timestamp\" \u003c fromUnixTimestamp64Nano(cast(1659940670682000000,'Int64'))) AND (tag_values[indexOf(tag_keys,'span_id')] = '4a15deb5-7231-4b64-905a-c772f6e2fe33') AND (\"metric_group\" IN ('apm_span_event'))) ORDER BY \"timestamp\" ASC LIMIT 10",
        "table": "monitor.metrics_all",
        "tenant_id": "",
        "terminus_key": ""
    },
    "events": [
        {
            "timeUnixNano": 1659940670885873000,
            "name": "profile_info",
            "attributes": {
                "applied.limit": "true",
                "blocks": "1",
                "bytes": "713706",
                "rows": "5",
                "rows.before.limit": "5"
            },
            "droppedAttributesCount": 0
        },
        {
            "timeUnixNano": 1659940670885882000,
            "name": "progress",
            "attributes": {
                "bytes": "158080328",
                "rows": "3102064",
                "total_rows": "3102064",
                "wrote_bytes": "0"
            },
            "droppedAttributesCount": 0
        },
        {
            "timeUnixNano": 1659940670885891000,
            "name": "progress",
            "attributes": {
                "bytes": "0",
                "rows": "0",
                "total_rows": "0",
                "wrote_bytes": "0"
            },
            "droppedAttributesCount": 0
        }
    ]
}`),
			},
			want: []metric.Metric{
				{
					Name:      "profile_info",
					Timestamp: int64(1659940670885873000),
					Tags: map[string]string{
						"env_id":                      "",
						"instrument":                  "executive",
						"instrument_version":          "",
						"metrics":                     "apm_span_event",
						"org_name":                    "",
						"result":                      "total: 5, interval: 0, columns: tttt,_meta::tag,host_ip::tag,timestamp,terminus_key::tag,trace_id::tag,field_count::field,_metric_scope_id::tag,span_id::tag,name,_metric_scope::tag,cluster_name::tag,event::tag,host::tag,org_name::tag, rows: [0x14001ce8a50 0x140019f7a90 0x140019f7ae0 0x140019f7b30 0x140019f7b80 0x140019f7c20 0x140019f7c70 0x140019f7cc0 0x140019f7d10 0x140019f7d60 0x140019f7db0 0x140019f7e00 0x140019f7e50 0x140019f7ea0 0x140019f7ef0]",
						"result_total":                "5",
						"service_id":                  "monitor",
						"service_name":                "monitor",
						"settings_max_execution_time": "65",
						"span_kind":                   "local",
						"span_layer":                  "local",
						"sql":                         "SELECT *, tag_values[indexOf(tag_keys,'event')] AS \"event\" FROM \"monitor\".\"metrics_all\" WHERE ((\"timestamp\" \u003e= fromUnixTimestamp64Nano(cast(0,'Int64'))) AND (\"timestamp\" \u003c fromUnixTimestamp64Nano(cast(1659940670682000000,'Int64'))) AND (tag_values[indexOf(tag_keys,'span_id')] = '4a15deb5-7231-4b64-905a-c772f6e2fe33') AND (\"metric_group\" IN ('apm_span_event'))) ORDER BY \"timestamp\" ASC LIMIT 10",
						"table":                       "monitor.metrics_all",
						"tenant_id":                   "",
						"terminus_key":                "",
						"applied.limit":               "true",
						"blocks":                      "1",
						"bytes":                       "713706",
						"rows":                        "5",
						"rows.before.limit":           "5",
						"event":                       "profile_info",
					},
					Fields:  nil,
					OrgName: "",
				},
				{
					Name:      "progress",
					Timestamp: int64(1659940670885882000),
					Tags: map[string]string{
						"env_id":                      "",
						"instrument":                  "executive",
						"instrument_version":          "",
						"metrics":                     "apm_span_event",
						"org_name":                    "",
						"result":                      "total: 5, interval: 0, columns: tttt,_meta::tag,host_ip::tag,timestamp,terminus_key::tag,trace_id::tag,field_count::field,_metric_scope_id::tag,span_id::tag,name,_metric_scope::tag,cluster_name::tag,event::tag,host::tag,org_name::tag, rows: [0x14001ce8a50 0x140019f7a90 0x140019f7ae0 0x140019f7b30 0x140019f7b80 0x140019f7c20 0x140019f7c70 0x140019f7cc0 0x140019f7d10 0x140019f7d60 0x140019f7db0 0x140019f7e00 0x140019f7e50 0x140019f7ea0 0x140019f7ef0]",
						"result_total":                "5",
						"service_id":                  "monitor",
						"service_name":                "monitor",
						"settings_max_execution_time": "65",
						"span_kind":                   "local",
						"span_layer":                  "local",
						"sql":                         "SELECT *, tag_values[indexOf(tag_keys,'event')] AS \"event\" FROM \"monitor\".\"metrics_all\" WHERE ((\"timestamp\" \u003e= fromUnixTimestamp64Nano(cast(0,'Int64'))) AND (\"timestamp\" \u003c fromUnixTimestamp64Nano(cast(1659940670682000000,'Int64'))) AND (tag_values[indexOf(tag_keys,'span_id')] = '4a15deb5-7231-4b64-905a-c772f6e2fe33') AND (\"metric_group\" IN ('apm_span_event'))) ORDER BY \"timestamp\" ASC LIMIT 10",
						"table":                       "monitor.metrics_all",
						"tenant_id":                   "",
						"terminus_key":                "",
						"bytes":                       "158080328",
						"rows":                        "3102064",
						"total_rows":                  "3102064",
						"wrote_bytes":                 "0",
						"event":                       "progress",
					},
					Fields:  nil,
					OrgName: "",
				},
				{
					Name:      "progress",
					Timestamp: int64(1659940670885891000),
					Tags: map[string]string{
						"env_id":                      "",
						"instrument":                  "executive",
						"instrument_version":          "",
						"metrics":                     "apm_span_event",
						"org_name":                    "",
						"result":                      "total: 5, interval: 0, columns: tttt,_meta::tag,host_ip::tag,timestamp,terminus_key::tag,trace_id::tag,field_count::field,_metric_scope_id::tag,span_id::tag,name,_metric_scope::tag,cluster_name::tag,event::tag,host::tag,org_name::tag, rows: [0x14001ce8a50 0x140019f7a90 0x140019f7ae0 0x140019f7b30 0x140019f7b80 0x140019f7c20 0x140019f7c70 0x140019f7cc0 0x140019f7d10 0x140019f7d60 0x140019f7db0 0x140019f7e00 0x140019f7e50 0x140019f7ea0 0x140019f7ef0]",
						"result_total":                "5",
						"service_id":                  "monitor",
						"service_name":                "monitor",
						"settings_max_execution_time": "65",
						"span_kind":                   "local",
						"span_layer":                  "local",
						"sql":                         "SELECT *, tag_values[indexOf(tag_keys,'event')] AS \"event\" FROM \"monitor\".\"metrics_all\" WHERE ((\"timestamp\" \u003e= fromUnixTimestamp64Nano(cast(0,'Int64'))) AND (\"timestamp\" \u003c fromUnixTimestamp64Nano(cast(1659940670682000000,'Int64'))) AND (tag_values[indexOf(tag_keys,'span_id')] = '4a15deb5-7231-4b64-905a-c772f6e2fe33') AND (\"metric_group\" IN ('apm_span_event'))) ORDER BY \"timestamp\" ASC LIMIT 10",
						"table":                       "monitor.metrics_all",
						"tenant_id":                   "",
						"terminus_key":                "",
						"bytes":                       "0",
						"rows":                        "0",
						"total_rows":                  "0",
						"wrote_bytes":                 "0",
						"event":                       "progress",
					},
					Fields:  nil,
					OrgName: "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uw := &unmarshalEventWork{
				buf: tt.fields.buf,
				err: tt.fields.err,
				callback: func(metrics []*metric.Metric) error {
					require.Equalf(t, len(tt.want), len(metrics), "length no should")
					for i, want := range tt.want {
						require.Equal(t, want.OrgName, metrics[i].OrgName)
						require.Equal(t, want.Timestamp, metrics[i].Timestamp)

						for k, v := range want.Tags {
							t.Run(fmt.Sprintf("metric[%v] check tags_%s", i, k), func(t *testing.T) {
								if k == "message" {
									require.ElementsMatch(t, strings.Split(v, ";"), strings.Split(metrics[i].Tags[k], ";"))
									return
								}
								require.Equalf(t, v, metrics[i].Tags[k], "tags no should")
							})

						}
						if want.Fields != nil {
							for k, v := range want.Fields {
								t.Run(fmt.Sprintf("metric[%v] check fields_%s", i, k), func(t *testing.T) {
									require.Equalf(t, v, metrics[i].Fields[k], "fields no should")
								})
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
