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

const (
	CH_TABLE_SERIES = "spans_series"
	CH_TABLE_META   = "spans_meta"

	OrgNameKey = "org_name"
)

type Span struct {
	OrgName      string `json:"-" ch:"org_name"`
	TraceId      string `json:"trace_id" ch:"trace_id"`
	SpanId       string `json:"span_id" ch:"span_id"`
	ParentSpanId string `json:"parent_span_id" ch:"parent_span_id"`
	// Deprecated, move to tags
	OperationName string            `json:"operation_name" ch:"-"`
	StartTime     int64             `json:"start_time" ch:"start_time"`
	EndTime       int64             `json:"end_time" ch:"end_time"`
	Tags          map[string]string `json:"tags" ch:"tags"`
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

type Series struct {
	StartTime    int64  `json:"start_time" ch:"start_time"`
	EndTime      int64  `json:"end_time" ch:"end_time"`
	SeriesID     uint64 `json:"seriesId" ch:"series_id"`
	OrgName      string `json:"-" ch:"org_name"`
	TraceId      string `json:"trace_id" ch:"trace_id"`
	SpanId       string `json:"span_id" ch:"span_id"`
	ParentSpanId string `json:"parent_span_id" ch:"parent_span_id"`
}

type Meta struct {
	SeriesID uint64 `json:"seriesId" ch:"series_id"`
	CreateAt int64  `json:"createAt" ch:"create_at"`
	OrgName  string `json:"-" ch:"org_name"`
	Key      string `json:"key" ch:"key"`
	Value    string `json:"value" ch:"value"`
}
