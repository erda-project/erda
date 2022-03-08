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
	"github.com/erda-project/erda/modules/oap/collector/core/model/odata"
)

type RuntimeProcessor struct {
	Name      string
	Processor Processor
}

type Processor interface {
	Component
	Process(in odata.ObservableData) (odata.ObservableData, error)
}

type RunningProcessor interface {
	Processor
	StartProcessor(consumer ObservableDataConsumerFunc)
}

type NoopProcessor struct {
}

func (n *NoopProcessor) ComponentID() ComponentID {
	return "NoopProcessor"
}

func (n *NoopProcessor) Process(in odata.ObservableData) (odata.ObservableData, error) {
	return in, nil
}
