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

package cassandra

import (
	reflect "reflect"
	"testing"
)

func Test_mergeSavedLog(t *testing.T) {
	tests := []struct {
		name string
		a    []*SavedLog
		b    []*SavedLog
		less func(a, b *SavedLog) bool
		want []*SavedLog
	}{
		{
			a: []*SavedLog{
				{
					Timestamp: 100,
					Offset:    10,
				},
				{
					Timestamp: 102,
					Offset:    12,
				},
			},
			b: []*SavedLog{
				{
					Timestamp: 101,
					Offset:    11,
				},
				{
					Timestamp: 103,
					Offset:    13,
				},
			},
			less: lessSavedLog,
			want: []*SavedLog{
				{
					Timestamp: 100,
					Offset:    10,
				},
				{
					Timestamp: 101,
					Offset:    11,
				},
				{
					Timestamp: 102,
					Offset:    12,
				},
				{
					Timestamp: 103,
					Offset:    13,
				},
			},
		},
		{
			a: []*SavedLog{
				{
					Timestamp: 100,
					Offset:    10,
				},
				{
					Timestamp: 102,
					Offset:    12,
				},
			},
			b: []*SavedLog{
				{
					Timestamp: 100,
					Offset:    11,
				},
				{
					Timestamp: 102,
					Offset:    9,
				},
			},
			less: lessSavedLog,
			want: []*SavedLog{
				{
					Timestamp: 100,
					Offset:    10,
				},
				{
					Timestamp: 100,
					Offset:    11,
				},
				{
					Timestamp: 102,
					Offset:    9,
				},
				{
					Timestamp: 102,
					Offset:    12,
				},
			},
		},
		{
			a: []*SavedLog{
				{
					Timestamp: 100,
					Offset:    10,
				},
				{
					Timestamp: 102,
					Offset:    12,
				},
			},
			b:    nil,
			less: lessSavedLog,
			want: []*SavedLog{
				{
					Timestamp: 100,
					Offset:    10,
				},
				{
					Timestamp: 102,
					Offset:    12,
				},
			},
		},
		{
			a: nil,
			b: []*SavedLog{
				{
					Timestamp: 100,
					Offset:    11,
				},
				{
					Timestamp: 102,
					Offset:    9,
				},
			},
			less: lessSavedLog,
			want: []*SavedLog{
				{
					Timestamp: 100,
					Offset:    11,
				},
				{
					Timestamp: 102,
					Offset:    9,
				},
			},
		},
		{
			a: []*SavedLog{
				{
					Timestamp: 102,
					Offset:    12,
				},
				{
					Timestamp: 100,
					Offset:    10,
				},
			},
			b: []*SavedLog{
				{
					Timestamp: 103,
					Offset:    13,
				},
				{
					Timestamp: 101,
					Offset:    11,
				},
			},
			less: reverseLessSavedLog,
			want: []*SavedLog{
				{
					Timestamp: 103,
					Offset:    13,
				},
				{
					Timestamp: 102,
					Offset:    12,
				},
				{
					Timestamp: 101,
					Offset:    11,
				},
				{
					Timestamp: 100,
					Offset:    10,
				},
			},
		},
		{
			a: []*SavedLog{
				{
					Timestamp: 102,
					Offset:    12,
				},
				{
					Timestamp: 100,
					Offset:    10,
				},
			},
			b: []*SavedLog{
				{
					Timestamp: 102,
					Offset:    9,
				},
				{
					Timestamp: 100,
					Offset:    11,
				},
			},
			less: reverseLessSavedLog,
			want: []*SavedLog{
				{
					Timestamp: 102,
					Offset:    12,
				},
				{
					Timestamp: 102,
					Offset:    9,
				},
				{
					Timestamp: 100,
					Offset:    11,
				},
				{
					Timestamp: 100,
					Offset:    10,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeSavedLog(tt.a, tt.b, tt.less); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeSavedLog() = %v, want %v", got, tt.want)
			}
		})
	}
}
