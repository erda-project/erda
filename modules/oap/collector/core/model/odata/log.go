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

package odata

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
type Logs []*Log

type Log struct {
	Item *lpb.Log  `json:"item"`
	Meta *Metadata `json:"meta"`
}

func NewLog(item *lpb.Log) *Log {
	return &Log{Item: item, Meta: &Metadata{data: map[string]string{}}}
}

func (l *Log) AddMetadata(key, value string) {
	l.Meta.Add(key, value)
}

func (l *Log) GetMetadata(key string) (string, bool) {
	return l.Meta.Get(key)
}

func (l *Log) HandleAttributes(handle func(attr map[string]string) map[string]string) {
	l.Item.Attributes = handle(l.Item.Attributes)
}

func (l *Log) HandleName(handle func(name string) string) {
	l.Item.Name = handle(l.Item.Name)
}

func (l *Log) Clone() ObservableData {
	item := &lpb.Log{
		TimeUnixNano: l.Item.TimeUnixNano,
		Name:         l.Item.Name,
		Attributes:   l.Item.Attributes,
		Relations:    l.Item.Relations,
		Severity:     l.Item.Severity,
		Content:      l.Item.Content,
	}
	return &Log{
		Item: item,
		Meta: l.Meta.Clone(),
	}
}

func (l *Log) Source() interface{} {
	return l.Item
}

func (l *Log) SourceCompatibility() interface{} {
	// TODO
	return nil
}

func (l *Log) SourceType() SourceType {
	return LogType
}

func (l *Log) String() string {
	buf, _ := json.Marshal(l.Item)
	return fmt.Sprintf("Item => %s", string(buf))
}
