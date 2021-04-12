// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package query

import (
	"net/url"
	"time"

	"github.com/erda-project/erda-infra/providers/i18n"
	tsql "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql"
	"github.com/olivere/elastic"
)

// metric keys
const (
	TimestampKey   = tsql.TimestampKey
	NameKey        = tsql.NameKey
	TagKey         = "tags"
	FieldKey       = "fields"
	ClusterNameKey = TagKey + ".cluster_name"
)

// ResultSet .
type ResultSet struct {
	*tsql.ResultSet
	Details interface{}
	Elapsed struct {
		Search time.Duration `json:"search"`
	} `json:"elapsed"`
}

// Queryer .
type Queryer interface {
	Query(tsql, statement string, params map[string]interface{}, options url.Values) (*ResultSet, error)
	QueryWithFormat(tsql, statement, format string, langCodes i18n.LanguageCodes, params map[string]interface{}, options url.Values) (*ResultSet, interface{}, error)
	Export(tsql, statement string, params map[string]interface{}, options url.Values, handle func(id string, source []byte) error) error // 暂时没用到, 待移除

	QueryRaw(metrics, clusters []string, start, end int64, searchSource *elastic.SearchSource) (*elastic.SearchResult, error)
	SearchRaw(indices []string, searchSource *elastic.SearchSource) (*elastic.SearchResult, error)
}
