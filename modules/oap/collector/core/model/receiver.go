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

type (
	MetricReceiverConsumeFunc func(data Metrics)
	TraceReceiverConsumeFunc  func(data Traces)
	LogReceiverConsumeFunc    func(data Logs)

	ObservableDataReceiverFunc func(data ObservableData)
)

type Receiver interface {
	Component
	RegisterConsumeFunc(consumer ObservableDataReceiverFunc)
}

type NoopReceiver struct {
}

func (n *NoopReceiver) ComponentID() ComponentID {
	return "NoopReceiver"
}

func (n *NoopReceiver) RegisterConsumeFunc(consumer ObservableDataReceiverFunc) {}

// type MetricReceiver interface {
// 	Component
// 	RegisterConsumeFunc(consumer MetricReceiverConsumeFunc)
// }
//
// type TraceReceiver interface {
// 	Component
// 	RegisterConsumeFunc(consumer TraceReceiverConsumeFunc)
// }
//
// type LogReceiver interface {
// 	Component
// 	RegisterConsumeFunc(consumer LogReceiverConsumeFunc)
// }
