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

package kuberneteslogs

import (
	"context"
	"fmt"
	"io"
	"os"
	reflect "reflect"
	"strings"
	"testing"
	"time"

	"github.com/erda-project/erda/modules/core/monitor/storage"
	"github.com/recallsong/go-utils/encoding/jsonx"
	v1 "k8s.io/api/core/v1"
)

type logTestItem struct {
	Timestamp int64
	Content   string
}

var (
	commonLogTestItems = []logTestItem{
		{Timestamp: 11, Content: "1111"},
		{Timestamp: 12, Content: "2222"},
		{Timestamp: 13, Content: "3333"},
		{Timestamp: 14, Content: "4444.0"},
		{Timestamp: 14, Content: "4444.1"},
		{Timestamp: 15, Content: "5555"},
		{Timestamp: 16, Content: "6666"},
	}
	commonTestPartitionKeys = []string{"n1", "p1", "c1"}
	commonTestLabels        = map[string]string{
		"namespace":      "n1",
		"pod_name":       "p1",
		"container_name": "c1",
	}
)

var stdout = os.Stdout

func printf(tmp string, args ...interface{}) {
	fmt.Fprintf(stdout, tmp, args...)
}

func queryFuncForTest(expectNamespace, expectPod, expectContainer string, items []logTestItem) func(ctx context.Context, namespace, pod string, opts *v1.PodLogOptions) (io.ReadCloser, error) {
	return func(ctx context.Context, namespace, pod string, opts *v1.PodLogOptions) (io.ReadCloser, error) {
		if expectNamespace != namespace || expectPod != pod || expectContainer != opts.Container {
			return nil, fmt.Errorf("want keys: [%q,%q,%q], got keys: [%q,%q,%q]",
				expectNamespace, expectPod, expectContainer, namespace, pod, opts.Container)
		}
		var lines []string
		if opts.SinceTime == nil {
			for _, item := range items {
				t := time.Unix(item.Timestamp/int64(time.Second), item.Timestamp%int64(time.Second))
				line := t.Format(time.RFC3339Nano) + " " + item.Content
				lines = append(lines, line)
			}
		} else {
			for i, item := range items {
				t := time.Unix(item.Timestamp/int64(time.Second), item.Timestamp%int64(time.Second))
				if t.After(opts.SinceTime.Time) || t.Equal(opts.SinceTime.Time) {
					for n := len(items); i < n; i++ {
						item := items[i]
						t := time.Unix(item.Timestamp/int64(time.Second), item.Timestamp%int64(time.Second))
						line := t.Format(time.RFC3339Nano) + " " + item.Content
						lines = append(lines, line)
					}
					break
				}
			}
		}
		printf("request since time: %v \n", opts.SinceTime)
		printf("response logs: \n%s\n", strings.Join(lines, "\n"))
		return io.NopCloser(strings.NewReader(strings.Join(lines, "\n"))), nil
	}
}

func Test_cStorage_Iterator_Next(t *testing.T) {
	tests := []struct {
		name        string
		source      []logTestItem
		bufferLines int64
		sel         *storage.Selector
		want        []*storage.Data
	}{
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     11,
				EndTime:       13,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: []*storage.Data{
				{
					Timestamp: 11,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "1111",
					},
				},
				{
					Timestamp: 12,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "2222",
					},
				},
			},
		},
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     12,
				EndTime:       13,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: []*storage.Data{
				{
					Timestamp: 12,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "2222",
					},
				},
			},
		},
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     13,
				EndTime:       16,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: []*storage.Data{
				{
					Timestamp: 13,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "3333",
					},
				},
				{
					Timestamp: 14,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "4444.0",
					},
				},
				{
					Timestamp: 14,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "4444.1",
					},
				},
				{
					Timestamp: 15,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "5555",
					},
				},
			},
		},
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     100,
				EndTime:       1000,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: nil,
		},
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     10,
				EndTime:       11,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: nil,
		},
		{
			source:      commonLogTestItems,
			bufferLines: defaultBufferLines,
			sel: &storage.Selector{
				StartTime:     11,
				EndTime:       13,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: []*storage.Data{
				{
					Timestamp: 11,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "1111",
					},
				},
				{
					Timestamp: 12,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "2222",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &cStorage{
				queryFunc:   queryFuncForTest(tt.sel.PartitionKeys[0], tt.sel.PartitionKeys[1], tt.sel.PartitionKeys[2], tt.source),
				bufferLines: tt.bufferLines,
			}
			it := s.Iterator(context.TODO(), tt.sel)
			var got []*storage.Data
			for it.Next() {
				got = append(got, it.Value())
			}
			if it.Error() != nil {
				t.Errorf("cStorage.Iterator() got error: %s", it.Error())
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cStorage.Iterator() \ngot %v, \nwant %v", jsonx.MarshalAndIndent(got), jsonx.MarshalAndIndent(tt.want))
			}
		})
	}
}

func Test_cStorage_Iterator_Prev(t *testing.T) {
	tests := []struct {
		name        string
		source      []logTestItem
		bufferLines int64
		sel         *storage.Selector
		want        []*storage.Data
	}{
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     11,
				EndTime:       13,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: []*storage.Data{
				{
					Timestamp: 12,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "2222",
					},
				},
				{
					Timestamp: 11,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "1111",
					},
				},
			},
		},
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     12,
				EndTime:       13,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: []*storage.Data{
				{
					Timestamp: 12,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "2222",
					},
				},
			},
		},
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     13,
				EndTime:       16,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: []*storage.Data{
				{
					Timestamp: 15,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "5555",
					},
				},
				{
					Timestamp: 14,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "4444.1",
					},
				},
				{
					Timestamp: 14,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "4444.0",
					},
				},
				{
					Timestamp: 13,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "3333",
					},
				},
			},
		},
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     100,
				EndTime:       1000,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: nil,
		},
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     0,
				EndTime:       11,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: nil,
		},
		{
			source:      commonLogTestItems,
			bufferLines: defaultBufferLines,
			sel: &storage.Selector{
				StartTime:     10,
				EndTime:       13,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: []*storage.Data{
				{
					Timestamp: 12,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "2222",
					},
				},
				{
					Timestamp: 11,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "1111",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &cStorage{
				queryFunc:   queryFuncForTest(tt.sel.PartitionKeys[0], tt.sel.PartitionKeys[1], tt.sel.PartitionKeys[2], tt.source),
				bufferLines: tt.bufferLines,
			}
			it := s.Iterator(context.TODO(), tt.sel)
			var got []*storage.Data
			for it.Prev() {
				got = append(got, it.Value())
			}
			if it.Error() != nil {
				t.Errorf("cStorage.Iterator() got error: %s", it.Error())
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cStorage.Iterator() \ngot %v, \nwant %v", jsonx.MarshalAndIndent(got), jsonx.MarshalAndIndent(tt.want))
			}
		})
	}
}

func Test_cStorage_Iterator_FirstNext(t *testing.T) {
	tests := []struct {
		name        string
		source      []logTestItem
		bufferLines int64
		sel         *storage.Selector
		want        []*storage.Data
	}{
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     13,
				EndTime:       15,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: []*storage.Data{
				{
					Timestamp: 13,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "3333",
					},
				},
				{
					Timestamp: 14,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "4444.0",
					},
				},
				{
					Timestamp: 14,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "4444.1",
					},
				},
			},
		},
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     100,
				EndTime:       1000,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: nil,
		},
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     10,
				EndTime:       11,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &cStorage{
				queryFunc:   queryFuncForTest(tt.sel.PartitionKeys[0], tt.sel.PartitionKeys[1], tt.sel.PartitionKeys[2], tt.source),
				bufferLines: tt.bufferLines,
			}
			it := s.Iterator(context.TODO(), tt.sel)
			var got []*storage.Data
			if it.First() {
				got = append(got, it.Value())
			}
			for it.Next() {
				got = append(got, it.Value())
			}
			if it.Error() != nil {
				t.Errorf("cStorage.Iterator() got error: %s", it.Error())
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cStorage.Iterator() \ngot %v, \nwant %v", jsonx.MarshalAndIndent(got), jsonx.MarshalAndIndent(tt.want))
			}
		})
	}
}

func Test_cStorage_Iterator_LastPrev(t *testing.T) {
	tests := []struct {
		name        string
		source      []logTestItem
		bufferLines int64
		sel         *storage.Selector
		want        []*storage.Data
	}{
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     13,
				EndTime:       15,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: []*storage.Data{
				{
					Timestamp: 14,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "4444.1",
					},
				},
				{
					Timestamp: 14,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "4444.0",
					},
				},
				{
					Timestamp: 13,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "3333",
					},
				},
			},
		},
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     100,
				EndTime:       1000,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: nil,
		},
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     10,
				EndTime:       11,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &cStorage{
				queryFunc:   queryFuncForTest(tt.sel.PartitionKeys[0], tt.sel.PartitionKeys[1], tt.sel.PartitionKeys[2], tt.source),
				bufferLines: tt.bufferLines,
			}
			it := s.Iterator(context.TODO(), tt.sel)
			var got []*storage.Data
			if it.Last() {
				got = append(got, it.Value())
			}
			for it.Prev() {
				got = append(got, it.Value())
			}
			if it.Error() != nil {
				t.Errorf("cStorage.Iterator() got error: %s", it.Error())
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cStorage.Iterator() \ngot %v, \nwant %v", jsonx.MarshalAndIndent(got), jsonx.MarshalAndIndent(tt.want))
			}
		})
	}
}

func Test_cStorage_Query(t *testing.T) {
	tests := []struct {
		name        string
		source      []logTestItem
		bufferLines int64
		sel         *storage.Selector
		want        []*storage.Data
	}{
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     13,
				EndTime:       15,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: []*storage.Data{
				{
					Timestamp: 13,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "3333",
					},
				},
				{
					Timestamp: 14,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "4444.0",
					},
				},
				{
					Timestamp: 14,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "4444.1",
					},
				},
			},
		},
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     13,
				EndTime:       15,
				PartitionKeys: commonTestPartitionKeys,
				Limit:         2,
			},
			want: []*storage.Data{
				{
					Timestamp: 13,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "3333",
					},
				},
				{
					Timestamp: 14,
					Labels:    commonTestLabels,
					Fields: map[string]interface{}{
						"content": "4444.0",
					},
				},
			},
		},
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     100,
				EndTime:       1000,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: nil,
		},
		{
			source:      commonLogTestItems,
			bufferLines: 1,
			sel: &storage.Selector{
				StartTime:     10,
				EndTime:       11,
				PartitionKeys: commonTestPartitionKeys,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &cStorage{
				queryFunc:   queryFuncForTest(tt.sel.PartitionKeys[0], tt.sel.PartitionKeys[1], tt.sel.PartitionKeys[2], tt.source),
				bufferLines: tt.bufferLines,
			}
			result, err := s.Query(context.TODO(), tt.sel)
			if err != nil {
				t.Errorf("cStorage.Query() got error: %s", err)
				return
			}
			if !reflect.DeepEqual(result.Values, tt.want) {
				t.Errorf("cStorage.Iterator() \ngot %v, \nwant %v", jsonx.MarshalAndIndent(result.Values), jsonx.MarshalAndIndent(tt.want))
			}
		})
	}
}
