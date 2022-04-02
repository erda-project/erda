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
	"sync"

	tpb "github.com/erda-project/erda-proto-go/oap/trace/pb"
)

// Spans
type Spans []*Span

type Span struct {
	Data map[string]interface{} `json:"data"`
	Meta *Metadata              `json:"meta"`
	sync.RWMutex
}

func (s *Span) HandleKeyValuePair(handler func(pairs map[string]interface{}) map[string]interface{}) {
	s.Data = handler(s.Data)
}

func (s *Span) Pairs() map[string]interface{} {
	return s.Data
}

func (s *Span) Name() string {
	return s.Data[NameKey].(string)
}

func NewSpan(item *tpb.Span) *Span {
	return &Span{Data: spanToMap(item), Meta: NewMetadata()}
}

func (s *Span) Metadata() *Metadata {
	return s.Meta
}

func (s *Span) Clone() ObservableData {
	res := make(map[string]interface{}, len(s.Data))
	for k, v := range s.Data {
		res[k] = v
	}
	return &Span{
		Data: res,
		Meta: s.Meta.Clone(),
	}
}

func (s *Span) Source() interface{} {
	return mapToSpan(s.Data)
}

func (s *Span) SourceCompatibility() interface{} {
	return mapToSpan(s.Data)
}

func (s *Span) SourceType() SourceType {
	return SpanType
}

func (s *Span) String() string {
	buf, _ := json.Marshal(s.Data)
	return fmt.Sprintf(string(buf))
}
