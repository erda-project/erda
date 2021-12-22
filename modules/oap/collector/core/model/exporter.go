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

type ExporterDescriber interface {
	Component
	Connect() error
	Close() error
}

type MetricExporter interface {
	ExporterDescriber
	ExportMetrics(metrics Metrics) error
}

type TraceExporter interface {
	ExporterDescriber
	ExportTraces(traces Traces) error
}

type LogExporter interface {
	ExporterDescriber
	ExportLogs(logs Logs) error
}

type NoopMetricExporter struct {
}

func (n *NoopMetricExporter) ComponentID() ComponentID {
	return "NoopMetricExporter"
}

func (n *NoopMetricExporter) Connect() error {
	return nil
}

func (n *NoopMetricExporter) Close() error {
	return nil
}

func (n *NoopMetricExporter) ExportMetrics(ms Metrics) error {
	return nil
}
