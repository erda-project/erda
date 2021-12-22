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

package model

type MetricProcessor interface {
	Component
	ProcessMetrics(metrics Metrics) (Metrics, error)
}

type TraceProcessor interface {
	Component
	ProcessTraces(spans Traces) (Traces, error)
}

type LogProcessor interface {
	Component
	ProcessLogs(logs Logs) (Logs, error)
}

type NoopMetricProcessor struct {
}

func (n *NoopMetricProcessor) ComponentID() ComponentID {
	return "NoopMetricProcessor"
}

func (n *NoopMetricProcessor) ProcessMetrics(ms Metrics) (Metrics, error) {
	return ms, nil
}
