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
	"encoding/json"
	"fmt"

	lpb "github.com/erda-project/erda-proto-go/oap/logs/pb"
)

// the representation of certain field key in Fields
const (
	// reference to lpb.Log.Content
	TagKeyLogContent = "__log_content"
	// reference to lpb.Log.Severity
	TagKeyLogSeverity = "__log_severity"
)

// Logs
type Logs struct {
	Logs []*lpb.Log `json:"logs"`
}

func (l *Logs) String() string {
	buf, _ := json.Marshal(l.Logs)
	return fmt.Sprintf("logs => %s", string(buf))
}

func (l *Logs) RangeFunc(_ handleFunc) {}

func (l *Logs) RangeNameFunc(handle func(name string) string) {
	for _, item := range l.Logs {
		item.Name = handle(item.Name)
	}
}

func (l *Logs) RangeTagsFunc(handle func(tags map[string]string) map[string]string) {
	for _, item := range l.Logs {
		item.Attributes = handle(item.Attributes)
	}
}

func (l *Logs) SourceData() interface{} {
	return l.Logs
}

func (l *Logs) CompatibilitySourceData() interface{} {
	// todo
	return nil
}

func (l *Logs) Clone() ObservableData {
	data := make([]*lpb.Log, len(l.Logs))
	copy(data, l.Logs)
	return &Logs{Logs: data}
}
