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

package entity

import "time"

type Entity struct {
	Timestamp          time.Time         `json:"-" ch:"timestamp"`
	UpdateTimestamp    time.Time         `json:"-" ch:"update_timestamp"`
	ID                 string            `json:"id" ch:"id"`
	Type               string            `json:"type" ch:"type"`
	Key                string            `json:"key" ch:"key"`
	Values             map[string]string `json:"values" ch:"values"`
	Labels             map[string]string `json:"labels" ch:"labels"`
	CreateTimeUnixNano int64             `json:"createTimeUnixNano"`
	UpdateTimeUnixNano int64             `json:"UpdateTimeUnixNano"`
}

type GroupedEntity struct {
	Timestamp       time.Time         `ch:"_timestamp"`
	UpdateTimestamp time.Time         `ch:"_update_timestamp"`
	ID              string            `ch:"id"`
	Type            string            `ch:"_type"`
	Key             string            `ch:"key"`
	Values          map[string]string `ch:"_values"`
	Labels          map[string]string `ch:"_labels"`
}
