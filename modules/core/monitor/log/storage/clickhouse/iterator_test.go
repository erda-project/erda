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

package clickhouse

import (
	"context"
	"testing"

	"gotest.tools/assert"

	"github.com/erda-project/erda/modules/core/monitor/log/storage"
)

func Test_Iterator_Should_Success(t *testing.T) {
	p := &provider{
		Cfg: &config{
			ReadPageSize: 100,
		},
		Loader:     MockLoader{},
		Creator:    MockCreator{},
		clickhouse: MockClickhouse{},
	}

	tests := []struct {
		name         string
		sel          *storage.Selector
		wantPageSize int
		wantErr      bool
	}{
		{
			sel: &storage.Selector{
				Start: 1,
				End:   10,
				Filters: []*storage.Filter{
					{
						Key:   "tags.trace_id",
						Op:    storage.EQ,
						Value: "trace_id_1",
					},
				},
				Meta: storage.QueryMeta{
					OrgNames: []string{"", "erda"},
				},
			},
			wantPageSize: 100,
			wantErr:      false,
		},
		{
			sel: &storage.Selector{
				Start: 1,
				End:   10,
				Filters: []*storage.Filter{
					{
						Key:   "tags.trace_id",
						Op:    storage.EQ,
						Value: "trace_id_1",
					},
				},
				Meta: storage.QueryMeta{
					OrgNames:            []string{"", "erda"},
					PreferredBufferSize: 200,
				},
			},
			wantPageSize: 200,
			wantErr:      false,
		},
		{
			sel: &storage.Selector{
				Start: 1,
				End:   10,
				Filters: []*storage.Filter{
					{
						Key:   "tags.trace_id",
						Op:    storage.EQ,
						Value: "trace_id_1",
					},
				},
				Meta: storage.QueryMeta{
					OrgNames:            []string{"", "erda"},
					PreferredBufferSize: 200,
				},
				Options: map[string]interface{}{
					storage.SelectorKeyCount: int64(300),
				},
			},
			wantPageSize: 300,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		it, err := p.Iterator(context.Background(), tt.sel)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("expect error but got nil error")
			}
		} else {
			assert.NilError(t, err)
		}

		typedIt := it.(*clickhouseIterator)
		assert.Equal(t, tt.wantPageSize, typedIt.pageSize)
	}
}
