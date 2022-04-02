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

// Logs
type Logs []*Log

type Log struct {
	Meta *Metadata              `json:"meta"`
	Data map[string]interface{} `json:"data"`
}

func NewLog(item *lpb.Log) *Log {
	return &Log{
		Meta: NewMetadata(),
		Data: logToMap(item),
	}
}

func (l *Log) HandleKeyValuePair(handler func(pairs map[string]interface{}) map[string]interface{}) {
	l.Data = handler(l.Data)
}

func (l *Log) Pairs() map[string]interface{} {
	return l.Data
}

func (l *Log) Name() string {
	return l.Data[NameKey].(string)
}

func (l *Log) Metadata() *Metadata {
	return l.Meta
}

func (l *Log) Clone() ObservableData {
	res := make(map[string]interface{}, len(l.Data))
	for k, v := range l.Data {
		res[k] = v
	}
	return &Log{
		Data: res,
		Meta: l.Meta.Clone(),
	}
}

func (l *Log) Source() interface{} {
	return mapToLog(l.Data)
}

func (l *Log) SourceCompatibility() interface{} {
	// TODO
	return nil
}

func (l *Log) SourceType() SourceType {
	return LogType
}

func (l *Log) String() string {
	buf, _ := json.Marshal(l.Data)
	return fmt.Sprintf(string(buf))
}
