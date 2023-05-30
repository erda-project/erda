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

package trace

import (
	"time"
)

const (
	CH_TABLE = "spans"
	// distributed table
	CH_TABLE_ALL = "spans_all"
)

type Span struct {
	StartTime     int64             `json:"start_time"` // timestamp nano
	EndTime       int64             `json:"end_time"`   // timestamp nano
	OrgName       string            `json:"org_name"`
	TraceId       string            `json:"trace_id"`
	SpanId        string            `json:"span_id"`
	ParentSpanId  string            `json:"parent_span_id"`
	OperationName string            `json:"operation_name" `
	Tags          map[string]string `json:"tags"`
}

func (s *Span) Hash() uint64 {
	return 0
}

func (s *Span) GetTags() map[string]string {
	if s.Tags == nil {
		s.Tags = map[string]string{}
	}
	return s.Tags
}

type TableSpan struct {
	StartTime     time.Time `ch:"start_time"` // timestamp nano
	EndTime       time.Time `ch:"end_time"`   // timestamp nano
	OrgName       string    `ch:"org_name"`
	TenantId      string    `ch:"tenant_id"`
	TraceId       string    `ch:"trace_id"`
	SpanId        string    `ch:"span_id"`
	ParentSpanId  string    `ch:"parent_span_id"`
	OperationName string    `ch:"operation_name" `
	TagKeys       []string  `ch:"tag_keys"`
	TagValues     []string  `ch:"tag_values"`
}
