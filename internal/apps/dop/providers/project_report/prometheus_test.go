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

package project_report

import (
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/erda-project/erda/internal/apps/dop/dao"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type mockInfoProvider struct{}

func (m *mockInfoProvider) GetRequestedIterationsInfo() (map[uint64]*IterationInfo, error) {
	return map[uint64]*IterationInfo{
		1: {
			Iteration: &dao.Iteration{
				BaseModel: dbengine.BaseModel{
					ID: 1,
				},
				ProjectID: 1,
			},
			IterationMetricFields: &IterationMetricFields{
				RequirementTotal:            2,
				RequirementDoneTotal:        1,
				RequirementCompleteSchedule: 0.5,
			},
		},
	}, nil
}

func Test_collectIterationInfo(t *testing.T) {
	p := &provider{
		orgSet:       &orgCache{cache.New(cache.NoExpiration, cache.NoExpiration)},
		projectSet:   &projectCache{cache.New(cache.NoExpiration, cache.NoExpiration)},
		iterationSet: &iterationCache{cache.New(cache.NoExpiration, cache.NoExpiration)},
	}
	collector := &PrometheusCollector{
		infoProvider:        &mockInfoProvider{},
		iterationLabelsFunc: p.iterationLabelsFunc,
		iterationMetrics:    allIterationMetrics,
		errors: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "iteration",
			Name:      "scrape_error",
			Help:      "1 if there was an error while getting iteration metrics, 0 otherwise",
		}),
	}
	ch := make(chan prometheus.Metric, 100)
	defer close(ch)
	go func() {
		for m := range ch {
			t.Log(m.Desc().String())
		}
	}()
	collector.collectIterationInfo(ch)
	time.Sleep(time.Second)
}

func TestDescribe(t *testing.T) {
	p := &provider{
		orgSet:       &orgCache{cache.New(cache.NoExpiration, cache.NoExpiration)},
		projectSet:   &projectCache{cache.New(cache.NoExpiration, cache.NoExpiration)},
		iterationSet: &iterationCache{cache.New(cache.NoExpiration, cache.NoExpiration)},
	}
	collector := &PrometheusCollector{
		infoProvider:        &mockInfoProvider{},
		iterationLabelsFunc: p.iterationLabelsFunc,
		iterationMetrics:    allIterationMetrics,
		errors: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "iteration",
			Name:      "scrape_error",
			Help:      "1 if there was an error while getting iteration metrics, 0 otherwise",
		}),
	}
	ch := make(chan *prometheus.Desc, 100)
	defer close(ch)
	go func() {
		for m := range ch {
			t.Log(m.String())
		}
	}()
	collector.Describe(ch)
	time.Sleep(time.Second)
}
