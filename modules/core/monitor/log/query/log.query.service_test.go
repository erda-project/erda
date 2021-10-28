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

package query

import (
	"math"
	"reflect"
	"testing"

	"github.com/erda-project/erda/modules/core/monitor/log/storage"
)

func Test_toQuerySelector(t *testing.T) {
	tests := []struct {
		name    string
		req     Request
		want    *storage.Selector
		wantErr bool
	}{
		{
			req: &LogRequest{
				Start: 1,
				End:   math.MaxInt64,
			},
			wantErr: true,
		},
		{
			req: &LogRequest{
				Start: 10,
				End:   1,
			},
			wantErr: true,
		},
		{
			req: &LogRequest{
				ID:     "testid",
				Start:  1,
				End:    100,
				Count:  -200,
				Source: "container",
			},
			want: &storage.Selector{
				Start:  1,
				End:    100,
				Scheme: "container",
				Filters: []*storage.Filter{
					{
						Key:   "id",
						Op:    storage.EQ,
						Value: "testid",
					},
					{
						Key:   "source",
						Op:    storage.EQ,
						Value: "container",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toQuerySelector(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("toQuerySelector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toQuerySelector() = %v, want %v", got, tt.want)
			}
		})
	}
}
