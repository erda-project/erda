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

package kafka

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model/odata"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/common/unmarshalwork"
)

func TestParseSpotMetricDropsFutureData(t *testing.T) {
	previousDatadrop := datadrop
	datadrop = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "test_receiver_dropped",
	}, []string{"receiver"})
	t.Cleanup(func() { datadrop = previousDatadrop })
	unmarshalwork.Start()
	t.Cleanup(unmarshalwork.Stop)

	consumed := false
	p := &provider{
		Cfg: &config{FutureDiscardDate: 24 * time.Hour},
		consumer: func(data odata.ObservableData) error {
			consumed = true
			return nil
		},
	}
	value, err := json.Marshal(&metric.Metric{
		Name:      "future_metric",
		Timestamp: time.Now().Add(25 * time.Hour).UnixNano(),
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.parseSpotMetric()(&sarama.ConsumerMessage{Value: value}); err != nil {
		t.Fatal(err)
	}
	if consumed {
		t.Fatal("future metric should not be consumed")
	}
	if got := testutil.ToFloat64(datadrop.WithLabelValues(string(spotMetric))); got != 1 {
		t.Fatalf("dropped metric count = %v, want 1", got)
	}
}
