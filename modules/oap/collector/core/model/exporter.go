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

import (
	"context"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/modules/oap/collector/common"
	"github.com/erda-project/erda/modules/oap/collector/core/model/odata"
)

type RuntimeExporter struct {
	Name     string
	Logger   logs.Logger
	Exporter Exporter
	Filter   *DataFilter
	Timer    *common.RunningTimer
	Buffer   *odata.Buffer
}

func (re *RuntimeExporter) Start(ctx context.Context) {
	go re.Timer.Run(ctx)
	for {
		select {
		case <-ctx.Done():
			if err := re.flushOnce(); err != nil {
				re.Logger.Errorf("event done, but flush err: %s", err)
			}
			return
		case <-re.Timer.Elapsed():
			if re.Buffer.Empty() {
				continue
			}
			if err := re.flushOnce(); err != nil {
				re.Logger.Errorf("event elapsed, but flush err: %s", err)
			}
		}
	}
}

func (re *RuntimeExporter) flushOnce() error {
	return re.Exporter.Export(re.Buffer.FlushAll())
}

func (re *RuntimeExporter) Add(od odata.ObservableData) {
	if re.Buffer.Full() {
		if err := re.flushOnce(); err != nil {
			re.Logger.Errorf("event buffer-full, but flush err: %s", err)
		}
		re.Timer.Reset()
	}
	re.Buffer.Push(od)
}

type Exporter interface {
	Component
	Connect() error
	Export(ods []odata.ObservableData) error
}

type NoopExporter struct{}

func (n *NoopExporter) ComponentConfig() interface{} {
	return nil
}

func (n *NoopExporter) Connect() error {
	return nil
}

func (n *NoopExporter) Export(ods []odata.ObservableData) error {
	return nil
}
