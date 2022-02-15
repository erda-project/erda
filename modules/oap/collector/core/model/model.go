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
	"sort"
	"strings"

	"github.com/cespare/xxhash"
	structpb "github.com/golang/protobuf/ptypes/struct"

	tpb "github.com/erda-project/erda-proto-go/oap/trace/pb"
)

type DataType string

const (
	MetricDataType DataType = "metric"
	TraceDataType  DataType = "trace"
	LogDataType    DataType = "log"
)

type handleFunc func(item *DataItem) (bool, *DataItem)

type ObservableData interface {
	Clone() ObservableData
	RangeFunc(handle handleFunc)
	RangeTagsFunc(handle func(tags map[string]string) map[string]string)
	RangeNameFunc(handle func(name string) string)
	SourceData() interface{}
	String() string
	CompatibilitySourceData() interface{}
}

// DataItem as middle object to store common data
type DataItem struct {
	Name          string
	TimestampNano uint64
	Tags          map[string]string
	// same as DataPoints when Type is metric
	// nothing when Type is trace&log
	Fields map[string]*structpb.Value
	Type   DataType
}

func (di DataItem) HashDataItem(fieldKey string) uint64 {
	var sb strings.Builder
	keys := make([]string, 0, len(di.Tags))
	for k := range di.Tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sb.WriteString(di.Name + "\n")
	for _, k := range keys {
		sb.WriteString(k + di.Tags[k] + "\n")
	}
	sb.WriteString(fieldKey)

	return xxhash.Sum64String(sb.String())
}

// Traces
type Traces struct {
	Spans []*tpb.Span `json:"spans"`
}

func (t *Traces) String() string {
	buf, _ := json.Marshal(t.Spans)
	return fmt.Sprintf("spans => %s", string(buf))
}

func (t *Traces) RangeFunc(handle handleFunc) {
}

func (t *Traces) RangeNameFunc(handle func(name string) string) {
	for _, item := range t.Spans {
		item.Name = handle(item.Name)
	}
}

func (t *Traces) SourceData() interface{} {
	return t.Spans
}

func (t *Traces) CompatibilitySourceData() interface{} {
	// todo
	return nil
}

func (t *Traces) Clone() ObservableData {
	data := make([]*tpb.Span, len(t.Spans))
	copy(data, t.Spans)
	return &Traces{Spans: data}
}

func (t *Traces) RangeTagsFunc(handle func(tags map[string]string) map[string]string) {
	for _, item := range t.Spans {
		item.Attributes = handle(item.Attributes)
	}
}
