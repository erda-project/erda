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

package analysis

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	subSystem       = "log_service"
	Namespace       = "erda_system"
	valueLogService = "log-service"

	keyMetric          = "metric"
	keyPattern         = "pattern"
	keyContent         = "content"
	keyCostTime        = "cost_time"
	keyCostTimeNanoSec = "cost_time_nano_sec"
	keyDiceComponent   = "dice_component"
)

type selfMetrics struct {
	slowAnalysis *prometheus.CounterVec
	consumedNum  prometheus.Counter
}

// scope: org / cluster / micro_service
func initSelfMetrics(scope string) *selfMetrics {
	s := &selfMetrics{
		slowAnalysis: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:      "slow_analysis_" + scope,
			Subsystem: subSystem,
			Help:      "log-service self analysis metric: slow analysis",
			Namespace: Namespace,
			ConstLabels: map[string]string{
				keyDiceComponent: valueLogService,
			},
		}, []string{keyMetric, keyPattern, keyContent, keyCostTime, keyCostTimeNanoSec}),
		consumedNum: prometheus.NewCounter(prometheus.CounterOpts{
			Name:      "consumed_num_" + scope,
			Subsystem: subSystem,
			Help:      "log-service self analysis metric: consumed number",
			Namespace: Namespace,
			ConstLabels: map[string]string{
				keyDiceComponent: valueLogService,
			},
		}),
	}
	prometheus.MustRegister(s.slowAnalysis)
	prometheus.MustRegister(s.consumedNum)
	return s
}
