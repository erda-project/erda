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

package pipeline

import (
	"testing"
	"time"

	"github.com/pyroscope-io/pyroscope/pkg/ingestion"
	"github.com/pyroscope-io/pyroscope/pkg/storage/segment"

	"github.com/erda-project/erda/internal/apps/msp/apm/trace"
	"github.com/erda-project/erda/internal/tools/monitor/core/log"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/erda-project/erda/internal/tools/monitor/core/profile"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model/odata"
)

func Test_startProcessors(t *testing.T) {
	p := &Pipeline{
		processors: []*model.RuntimeProcessor{
			{
				Name:      "profile-processor",
				Processor: &mockProfileProcessor{},
				Filter:    &model.DataFilter{},
			},
		},
		dtype: odata.ProfileType,
	}
	metricProvider := &Pipeline{
		processors: []*model.RuntimeProcessor{
			{
				Name:      "metric-processor",
				Processor: &mockProfileProcessor{},
				Filter:    &model.DataFilter{},
			},
		},
		dtype: odata.ExternalMetricType,
	}
	dataIn := make(chan odata.ObservableData, 10)
	dataOut := make(chan odata.ObservableData, 10)
	dataIn <- &profile.ProfileIngest{
		Metadata: ingestion.Metadata{
			Key: segment.NewKey(map[string]string{
				"DICE_WORKSPACE": "test",
				"DICE_ORG_NAME":  "erda",
			}),
		},
	}
	metricDataIn := make(chan odata.ObservableData, 10)
	metricDataOut := make(chan odata.ObservableData, 10)
	metricDataIn <- &metric.Metric{
		Name: "metric1",
	}
	go p.startProcessors(dataIn, dataOut)
	time.Sleep(time.Second)
	go metricProvider.startProcessors(metricDataIn, metricDataOut)
	time.Sleep(time.Second)
	profileOutput := <-dataOut
	metric1 := <-metricDataOut
	if _, ok := profileOutput.(*profile.Output); !ok {
		t.Errorf("profile processor failed")
	}
	if _, ok := metric1.(*metric.Metric); !ok {
		t.Errorf("metric processor failed")
	}
}

type mockProfileProcessor struct{}

func (m *mockProfileProcessor) ComponentConfig() interface{} { return nil }
func (m *mockProfileProcessor) ComponentClose() error        { return nil }
func (p *mockProfileProcessor) ProcessMetric(item *metric.Metric) (*metric.Metric, error) {
	return item, nil
}
func (p *mockProfileProcessor) ProcessLog(item *log.Log) (*log.Log, error)        { return item, nil }
func (p *mockProfileProcessor) ProcessSpan(item *trace.Span) (*trace.Span, error) { return item, nil }
func (p *mockProfileProcessor) ProcessRaw(item *odata.Raw) (*odata.Raw, error)    { return item, nil }
func (p *mockProfileProcessor) ProcessProfile(item *profile.ProfileIngest) (*profile.Output, error) {
	return &profile.Output{}, nil
}
