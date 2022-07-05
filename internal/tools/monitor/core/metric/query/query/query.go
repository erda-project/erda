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
	"context"
	"net/url"
	"time"

	"github.com/olivere/elastic"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"

	"github.com/erda-project/erda-infra/providers/i18n"

	indexloader "github.com/erda-project/erda/internal/tools/monitor/core/storekit/elasticsearch/index/loader"
)

// Queryer .
type Queryer interface {
	Query(tsql, statement string, params map[string]interface{}, options url.Values) (*model.ResultSet, error)
	QueryWithFormat(tsql, statement, format string, langCodes i18n.LanguageCodes, params map[string]interface{}, filters []*model.Filter, options url.Values) (*model.ResultSet, interface{}, error)

	QueryRaw(metrics, clusters []string, start, end int64, searchSource *elastic.SearchSource) (*elastic.SearchResult, error)
	SearchRaw(indices []string, searchSource *elastic.SearchSource) (*elastic.SearchResult, error)
}

// IndexLoader .
type IndexLoader interface {
	indexloader.Interface
	GetIndices(metrics, clusters []string, start, end int64) []string
}

// MetricIndexLoader .
type MetricIndexLoader struct {
	indexloader.Interface
}

func (mi *MetricIndexLoader) GetIndices(metrics, clusters []string, start, end int64) []string {
	keys := make([]indexloader.KeyPath, len(metrics)+1)
	for i, item := range metrics {
		keys[i] = indexloader.KeyPath{
			Keys:      []string{item},
			Recursive: true,
		}
	}
	keys[len(metrics)] = indexloader.KeyPath{}
	start = start * int64(time.Millisecond)
	end = end*int64(time.Millisecond) + (int64(time.Millisecond) - 1)
	return mi.Interface.Indices(context.Background(), start, end, keys...)
}
