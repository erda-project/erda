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

	tpb "github.com/erda-project/erda-proto-go/oap/trace/pb"
)

// Spans
type Spans []*Span

type Span struct {
	Item *tpb.Span `json:"item"`
	Meta *Metadata `json:"meta"`
}

func NewSpan(item *tpb.Span) *Span {
	return &Span{Item: item, Meta: &Metadata{Data: map[string]string{}}}
}

func (s *Span) AddMetadata(key, value string) {
	if s.Meta == nil {
		s.Meta = &Metadata{Data: make(map[string]string)}
	}
	s.Meta.Add(key, value)
}

func (s *Span) GetMetadata(key string) (string, bool) {
	if s.Meta == nil {
		s.Meta = &Metadata{Data: make(map[string]string)}
	}
	return s.Meta.Get(key)
}

func (s *Span) HandleAttributes(handle func(attr map[string]string) map[string]string) {
	s.Item.Attributes = handle(s.Item.Attributes)
}

func (s *Span) HandleName(handle func(name string) string) {
	s.Item.Name = handle(s.Item.Name)
}

func (s *Span) Clone() ObservableData {
	item := &tpb.Span{
		TraceID:           s.Item.TraceID,
		SpanID:            s.Item.SpanID,
		ParentSpanID:      s.Item.ParentSpanID,
		StartTimeUnixNano: s.Item.StartTimeUnixNano,
		EndTimeUnixNano:   s.Item.EndTimeUnixNano,
		Name:              s.Item.Name,
		Relations:         s.Item.Relations,
		Attributes:        s.Item.Attributes,
	}
	return &Span{
		Item: item,
		Meta: s.Meta.Clone(),
	}
}

func (s *Span) Source() interface{} {
	return s.Item
}

func (s *Span) SourceCompatibility() interface{} {
	return s.Item
}

func (s *Span) SourceType() SourceType {
	return SpanType
}

func (s *Span) String() string {
	buf, _ := json.Marshal(s.Item)
	return fmt.Sprintf("Item => %s", string(buf))
}
