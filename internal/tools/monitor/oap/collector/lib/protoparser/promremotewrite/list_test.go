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

package promremotewrite

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"testing"
	"time"
)

var metricsJson = `
  [
    {
      "name": "host_summary_node_exporter",
      "timestamp": 1741830321335000000,
      "tags": {
        "collector_group": "node_mem",
        "container": "node-exporter",
        "endpoint": "metrics",
        "instance": "10.0.1.226",
        "namespace": "default",
        "pod": "node-exporter-prometheus-node-exporter-6gkpm",
        "prometheus": "default/prometheus",
        "prometheus_replica": "prometheus-prometheus-0"
      },
      "fields": {
        "node_memory_MemFree_bytes": 6249947136
      }
    },
    {
      "name": "host_summary_node_exporter",
      "timestamp": 1741830321335000000,
      "tags": {
        "collector_group": "node_mem",
        "container": "node-exporter",
        "endpoint": "metrics",
        "instance": "10.0.1.226",
        "namespace": "default",
        "pod": "node-exporter-prometheus-node-exporter-6gkpm",
        "prometheus": "default/prometheus",
        "prometheus_replica": "prometheus-prometheus-0"
      },
      "fields": {
        "node_memory_MemTotal_bytes": 32843362304
      }
    },
    {
      "name": "host_summary_node_exporter",
      "timestamp": 1741830321335000000,
      "tags": {
        "collector_group": "node_mem",
        "container": "node-exporter",
        "endpoint": "metrics",
        "instance": "10.0.1.226",
        "namespace": "default",
        "pod": "node-exporter-prometheus-node-exporter-6gkpm",
        "prometheus": "default/prometheus",
        "prometheus_replica": "prometheus-prometheus-0"
      },
      "fields": {
        "node_memory_Buffers_bytes": 2119184384
      }
    },
    {
      "name": "host_summary_node_exporter",
      "timestamp": 1741830321335000000,
      "tags": {
        "collector_group": "node_mem",
        "container": "node-exporter",
        "endpoint": "metrics",
        "instance": "10.0.1.226",
        "namespace": "default",
        "pod": "node-exporter-prometheus-node-exporter-6gkpm",
        "prometheus": "default/prometheus",
        "prometheus_replica": "prometheus-prometheus-0"
      },
      "fields": {
        "node_memory_Cached_bytes": 14316683264
      }
    }
  ]
`

func TestList(t *testing.T) {
	//marshal, _ := json.Marshal(list)
	var ch = make(chan *metric.Metric, 1000)
	go DealGroupMetrics(context.Background(), GroupMetricsOptions{
		MinSize:        0,
		RetentionRatio: 0,
		Callback: func(record *metric.Metric) error {
			return nil
		},
		MetricsChan: ch,
	})
	var ms = make([]*metric.Metric, 0)
	err := json.Unmarshal([]byte(metricsJson), &ms)
	fmt.Printf("%v", err)
	for i := 0; i < len(ms)-1; i++ {
		ch <- ms[i]
	}
	ch <- nil
	time.Sleep(65 * time.Second)
	go func() {
		ch <- ms[len(ms)-1]
	}()
	time.Sleep(10 * time.Second)
}
