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

package esinfluxql

import (
	"encoding/json"
	"testing"

	"github.com/olivere/elastic"
	"github.com/stretchr/testify/require"

	tsql "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/es-tsql"
)

func TestNormalSelectStmt(t *testing.T) {
	tests := []struct {
		name    string
		stm     string
		params  map[string]interface{}
		require func(*testing.T, []tsql.Query)
	}{
		/*
			storage no support function
			INFO[2022-06-30 12:02:39.445] [playback]ql:influxql, stm:SELECT rateps(elapsed_count::field) FROM application_http,application_rpc WHERE (target_terminus_key::tag=$terminus_key OR source_terminus_key::tag=$terminus_key) AND (target_service_id::tag=$service_id OR target_service_id::tag=$service_id)  GROUP BY time(),map:map[layer_path: service_id:pipeline terminus_key:3ac23db32e2590b65fd4e930d6760774],filter:[],opts:map[end:[1656561758284] start:[1656558158284]]  module=erda.core.monitor.metric
			,
		*/
		{
			name:   "group,where,scripts",
			stm:    "SELECT sum(if(eq(error::tag, 'true'),elapsed_count::field,0))/sum(elapsed_count::field) FROM application_http,application_rpc WHERE (target_terminus_key::tag=$terminus_key OR source_terminus_key::tag=$terminus_key) AND (target_service_id::tag=$service_id OR target_service_id::tag=$service_id)  GROUP BY time()",
			params: map[string]interface{}{"terminus_key": "123", "service_id": "456"},
			require: func(t *testing.T, queries []tsql.Query) {
				require.Len(t, queries, 1)
				require.Len(t, queries[0].Sources(), 2)

				require.Equal(t, "application_http", queries[0].Sources()[0].Name)
				require.Equal(t, "application_rpc", queries[0].Sources()[1].Name)

				require.Equal(t, "", queries[0].Sources()[0].Database)
				require.Equal(t, "", queries[0].Sources()[1].Database)

				searchSource := queries[0].SearchSource()

				//*elastic.SearchSource
				require.NotNil(t, searchSource)

				elasticsearchSearchSource, ok := searchSource.(*elastic.SearchSource)
				require.Truef(t, ok, "searchSource is not *elastic.SearchSource")
				queryPlan, err := elasticsearchSearchSource.Source()
				require.NoError(t, err)

				require.NotNilf(t, queryPlan, "query_plan is nil")
				plan, ok := queryPlan.(map[string]interface{})
				require.Truef(t, ok, "query_plan is not map[string]interface{}")

				require.Len(t, plan, 3)
				require.Equal(t, 0, plan["size"].(int))

				queryStm, ok := plan["query"].(map[string]interface{})
				require.Truef(t, ok, "query is not map[string]interface{}")
				require.NotNil(t, queryStm)

				filterStm, ok := queryStm["bool"].(map[string]interface{})["filter"].([]interface{})
				require.Truef(t, ok, "filter is by bool and filter")

				require.Len(t, filterStm, 2)
				require.Equal(t, map[string]interface{}{
					"range": map[string]interface{}{
						"timestamp": map[string]interface{}{
							"from":          int64(0),
							"to":            int64(0),
							"include_lower": true,
							"include_upper": true,
						},
					},
				}, filterStm[0])

				require.Equal(t, map[string]interface{}{
					"bool": map[string]interface{}{
						"filter": []interface{}{
							map[string]interface{}{
								"bool": map[string]interface{}{
									"should": []interface{}{
										map[string]interface{}{
											"bool": map[string]interface{}{
												"filter": map[string]interface{}{
													"term": map[string]interface{}{
														"tags.target_terminus_key": "123",
													},
												},
											},
										},
										map[string]interface{}{
											"bool": map[string]interface{}{
												"filter": map[string]interface{}{
													"term": map[string]interface{}{
														"tags.source_terminus_key": "123",
													},
												},
											},
										},
									},
								},
							},
							map[string]interface{}{
								"bool": map[string]interface{}{
									"should": []interface{}{
										map[string]interface{}{
											"bool": map[string]interface{}{
												"filter": map[string]interface{}{
													"term": map[string]interface{}{
														"tags.target_service_id": "456",
													},
												},
											},
										},
										map[string]interface{}{
											"bool": map[string]interface{}{
												"filter": map[string]interface{}{
													"term": map[string]interface{}{
														"tags.target_service_id": "456",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}, filterStm[1])

				aggsStm, ok := plan["aggregations"].(map[string]interface{})
				require.Truef(t, ok, "aggregations is not map[string]interface{}")

				zeroFloat64 := float64(0)
				scripts := json.RawMessage("\"(((((doc.containsKey('tags.error')?doc['tags.error'].value:''))==('true')))?((doc.containsKey('fields.elapsed_count')?doc['fields.elapsed_count'].value:'')):(0))\"")

				require.Equal(t, map[string]interface{}{
					"histogram": map[string]interface{}{
						"histogram": map[string]interface{}{
							"field":         "timestamp",
							"interval":      float64(60000000000),
							"offset":        float64(0),
							"min_doc_count": int64(0),
							"extended_bounds": map[string]interface{}{
								"min": &zeroFloat64,
								"max": &zeroFloat64,
							},
						},
						"aggregations": map[string]interface{}{
							"72e4961c054d5bb4": map[string]interface{}{
								"sum": map[string]interface{}{
									"script": map[string]interface{}{
										/*
											"(((((doc.containsKey('tags.error')?doc['tags.error'].value:''))==('true')))?((doc.containsKey('fields.elapsed_count')?doc['fields.elapsed_count'].value:'')):(0))"
										*/
										"source": &scripts,
									},
								},
							},
							"30c5a29cb248b5d9": map[string]interface{}{
								"sum": map[string]interface{}{
									"field": "fields.elapsed_count",
								},
							},
						},
					},
				}, aggsStm)
			},
		},
		{
			name:   "normal stmt",
			stm:    "select * from application_http",
			params: map[string]interface{}{"terminus_key": "123", "service_id": "456"},
			require: func(t *testing.T, queries []tsql.Query) {
				require.Len(t, queries, 1)
				require.Len(t, queries[0].Sources(), 1)

				require.Equal(t, "application_http", queries[0].Sources()[0].Name)
				require.Equal(t, "", queries[0].Sources()[0].Database)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := New(0, 0, test.stm, false)
			if test.params != nil {
				p.SetParams(test.params)
			}
			err := p.Build()
			require.NoError(t, err)
			queries, err := p.ParseQuery(ElasticsearchKind)
			require.NoError(t, err)
			test.require(t, queries)
		})
	}
}
