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

package elasticsearch

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/olivere/elastic"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
	indexloader "github.com/erda-project/erda/internal/tools/monitor/core/storekit/elasticsearch/index/loader"
)

func TestSearch(t *testing.T) {
	tests := []struct {
		name          string
		query         func(*MockQueryMockRecorder)
		want          *model.Data
		mock          func(*MockInterfaceMockRecorder)
		mockExecution func(context.Context, *elastic.Client, []string, *elastic.SearchSource) (*elastic.SearchResult, error)
	}{
		{
			name: "normal",
			mock: func(expect *MockInterfaceMockRecorder) {
				expect.Indices(context.Background(), int64(0), int64(1999999), []indexloader.KeyPath{
					{
						Keys:      []string{"name"},
						Recursive: true,
					}, {
						Keys:      nil,
						Recursive: false,
					},
				}).Return([]string{"test-index"})
				expect.RequestTimeout().Return(time.Second)
				expect.Client().Return(&elastic.Client{})
			},
			query: func(expect *MockQueryMockRecorder) {
				expect.Sources().Return([]*model.Source{{Database: "db", Name: "name"}})
				expect.Timestamp().Return(int64(0), int64(1))
				expect.AppendBoolFilter(gomock.Any(), gomock.Any()).Return()
				expect.Return(false)
				expect.SearchSource().Return(elastic.NewSearchSource().Query(
					elastic.NewBoolQuery().Filter()))
				expect.ParseResult(context.Background(), gomock.Any()).Return(&model.Data{}, nil)
			},
			mockExecution: func(ctx context.Context, client *elastic.Client, strings []string, source *elastic.SearchSource) (*elastic.SearchResult, error) {
				return &elastic.SearchResult{}, nil
			},
			want: &model.Data{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			loader := NewMockInterface(ctrl)
			query := NewMockQuery(ctrl)
			defer ctrl.Finish()

			if tt.mock != nil {
				tt.mock(loader.EXPECT())
			}
			if tt.query != nil {
				tt.query(query.EXPECT())
			}
			if tt.mockExecution != nil {
				execution = tt.mockExecution
			}
			p := &provider{
				Loader: loader,
			}

			got, err := p.Query(context.Background(), query)
			if err != nil {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.want)
			}
			require.Equal(t, tt.want, got.Data)
		})
	}
}
