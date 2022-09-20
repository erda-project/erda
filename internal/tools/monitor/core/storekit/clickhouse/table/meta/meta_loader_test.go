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

package meta

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/pkg/mock"
)

func Test(t *testing.T) {
	p := provider{}
	p.Clickhouse = &mockClickhouse{}

}

func TestMetricMetaWantSQL(t *testing.T) {
	mockUpdateChn := make(chan *updateMetricsRequest, 1)
	defer close(mockUpdateChn)
	go func() {
		for {
			select {
			case req, _ := <-mockUpdateChn:
				req.Done <- struct{}{}
				close(req.Done)
				return
			default:
			}

		}

	}()

	p := provider{
		Cfg: &config{
			MetaStartTime: time.Hour * -2,
			MetaTable:     "metric_meta",
		},
		updateMetricsCh: mockUpdateChn,
	}

	for _, test := range []struct {
		name       string
		want       string
		mockResult mockResult
	}{
		{
			name:       "test",
			want:       "SELECT \"metric_group\", \"org_name\", \"tenant_id\", groupUniqArray(arrayJoin(if(empty(string_field_keys),[null],string_field_keys))) AS \"sk\", groupUniqArray(arrayJoin(if(empty(number_field_keys),[null],number_field_keys))) AS \"nk\", groupUniqArray(arrayJoin(if(empty(tag_keys),[null],tag_keys))) AS \"tk\" FROM \"metric_meta\" WHERE ((\"timestamp\" >= fromUnixTimestamp64Nano(cast(1658795469067491000,'Int64'))) AND (\"timestamp\" < fromUnixTimestamp64Nano(cast(1658802669067491000,'Int64')))) GROUP BY \"metric_group\", \"org_name\", \"tenant_id\"",
			mockResult: mockResult{},
		},
	} {
		t.Run(test.name, func(t *testing.T) {

			now = func() time.Time {
				return time.Unix(0, 1658806269067491000)
			}

			p.Clickhouse = &mockClickhouse{
				mockResult: &test.mockResult,
				verify: func(sql string) {
					require.Equal(t, test.want, sql)
				},
			}

			err := p.reloadMetaFromClickhouse(context.Background())
			require.NoError(t, err)

		})
	}
}

func TestClickhouseMetaLoader(t *testing.T) {
	tests := []struct {
		name       string
		mockResult mockResult
		want       map[MetricUniq]*MetricMeta
	}{
		{
			name: "one_metric",
			mockResult: mockResult{data: []*ckMeta{
				{
					MetricGroup: "metric",
					OrgName:     "org",
					TenantId:    "",
					StringKeys:  []string{"field"},
					NumberKeys:  []string{"field2"},
					TagKeys:     []string{"tag"},
				},
			}},
			want: map[MetricUniq]*MetricMeta{
				MetricUniq{
					MetricGroup: "metric",
					Scope:       "org",
					ScopeId:     "",
				}: {
					MetricGroup: "metric",
					Scope:       "org",
					ScopeId:     "",
					StringKeys:  []string{"field"},
					NumberKeys:  []string{"field2"},
					TagKeys:     []string{"tag"},
				},
			},
		},
		{
			name: "two_metric",
			mockResult: mockResult{data: []*ckMeta{
				{
					MetricGroup: "metric",
					OrgName:     "org",
					TenantId:    "",
					StringKeys:  []string{"field"},
					NumberKeys:  []string{"field2"},
					TagKeys:     []string{"tag"},
				},
				{
					MetricGroup: "metric",
					OrgName:     "org",
					TenantId:    "1",
					StringKeys:  []string{"field"},
					NumberKeys:  []string{"field2"},
					TagKeys:     []string{"tag"},
				},
			}},
			want: map[MetricUniq]*MetricMeta{
				MetricUniq{
					MetricGroup: "metric",
					Scope:       "org",
					ScopeId:     "",
				}: {
					MetricGroup: "metric",
					Scope:       "org",
					ScopeId:     "",
					StringKeys:  []string{"field"},
					NumberKeys:  []string{"field2"},
					TagKeys:     []string{"tag"},
				},
				MetricUniq{
					MetricGroup: "metric",
					Scope:       "org",
					ScopeId:     "1",
				}: {
					MetricGroup: "metric",
					Scope:       "org",
					ScopeId:     "1",
					StringKeys:  []string{"field"},
					NumberKeys:  []string{"field2"},
					TagKeys:     []string{"tag"},
				},
			},
		},
		{
			name: "merge_metric",
			mockResult: mockResult{data: []*ckMeta{
				{
					MetricGroup: "metric",
					OrgName:     "org",
					TenantId:    "",
					StringKeys:  []string{"field"},
					NumberKeys:  []string{"field2"},
					TagKeys:     []string{"tag"},
				},
				{
					MetricGroup: "metric",
					OrgName:     "org",
					TenantId:    "",
					StringKeys:  []string{"field3"},
					NumberKeys:  []string{"field4"},
					TagKeys:     []string{"tag"},
				},
			}},
			want: map[MetricUniq]*MetricMeta{
				MetricUniq{
					MetricGroup: "metric",
					Scope:       "org",
					ScopeId:     "",
				}: {
					MetricGroup: "metric",
					Scope:       "org",
					ScopeId:     "",
					StringKeys:  []string{"field", "field3"},
					NumberKeys:  []string{"field2", "field4"},
					TagKeys:     []string{"tag", "tag"},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockLogger := mock.NewMockLogger(ctrl)
			mockLogger.EXPECT().Info(gomock.Any()).Return()
			mockLogger.EXPECT().Info(gomock.Any()).Return()
			mockLogger.EXPECT().Info(gomock.Any()).Return()

			mockUpdateChn := make(chan *updateMetricsRequest, 1)
			defer close(mockUpdateChn)

			p := provider{
				updateMetricsCh: mockUpdateChn,
				Cfg: &config{
					MetaStartTime:  time.Hour * -2,
					MetaTable:      "metric_meta",
					ReloadInterval: time.Hour,
					Once:           true,
				},
				Log: mockLogger,
			}

			go func() {
				for {
					select {
					case req, _ := <-mockUpdateChn:
						p.Meta.Store(req.Metas)
						req.Done <- struct{}{}
						close(req.Done)
						return
					default:
					}

				}
			}()

			p.Clickhouse = &mockClickhouse{
				mockResult: &test.mockResult,
			}

			err := p.runClickhouseMetaLoader(context.Background())
			require.NoError(t, err)

			metas, ok := p.Meta.Load().(map[MetricUniq]*MetricMeta)
			require.True(t, ok)

			for k, v := range test.want {
				got, ok := metas[k]
				require.True(t, ok)
				require.Equal(t, v.MetricGroup, got.MetricGroup)
				require.Equal(t, v.Scope, got.Scope)
				require.Equal(t, v.ScopeId, got.ScopeId)
				require.ElementsMatch(t, v.NumberKeys, got.NumberKeys, "number keys not match")
				require.ElementsMatch(t, v.TagKeys, got.TagKeys, "tag keys not match")
				require.ElementsMatch(t, v.StringKeys, got.StringKeys, "string keys not match")
			}
		})
	}
}
