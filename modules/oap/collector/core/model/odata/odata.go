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
	"sort"
	"strings"

	"github.com/cespare/xxhash"

	"github.com/erda-project/erda-proto-go/oap/common/pb"
)

type SourceType string

const (
	MetricType SourceType = "METRIC"
	SpanType   SourceType = "SPAN"
	LogType    SourceType = "LOG"
	RawType    SourceType = "RAW"
)

type ObservableData interface {
	Attributes() map[string]string
	HandleAttributes(handle func(attr map[string]string) map[string]string)
	Name() string
	HandleName(handle func(name string) string)
	Metadata() *Metadata
	Clone() ObservableData
	Source() interface{}
	SourceCompatibility() interface{}
	SourceType() SourceType
	String() string
}

type SourceItem interface {
	GetName() string
	GetAttributes() map[string]string
	GetRelations() *pb.Relation
}

// fieldKey is field of measurement
// certain metric if fields if sourceItem is Metric
// empty string if sourceItem is Log
// ...
func HashSourceItem(fieldKey string, sourceItem SourceItem) uint64 {
	var sb strings.Builder
	keys := make([]string, 0, len(sourceItem.GetAttributes()))
	for k := range sourceItem.GetAttributes() {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sb.WriteString(sourceItem.GetName() + "\n")
	for _, k := range keys {
		sb.WriteString(k + sourceItem.GetAttributes()[k] + "\n")
	}
	sb.WriteString(fieldKey)

	return xxhash.Sum64String(sb.String())
}
