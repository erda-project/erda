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

package metricmeta

import (
	"embed"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda-infra/providers/i18n"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
)

type mockI18n struct {
}

type mockTranslator struct {
}

func (m mockTranslator) Get(lang i18n.LanguageCodes, key, def string) string {
	return key
}

func (m mockTranslator) Text(lang i18n.LanguageCodes, key string) string {
	return key
}

func (m mockTranslator) Sprintf(lang i18n.LanguageCodes, key string, args ...interface{}) string {
	return key
}

func (m mockI18n) Get(namespace string, lang i18n.LanguageCodes, key, def string) string {
	return key
}

func (m mockI18n) Text(namespace string, lang i18n.LanguageCodes, key string) string {
	return key
}

func (m mockI18n) Sprintf(namespace string, lang i18n.LanguageCodes, key string, args ...interface{}) string {
	return key
}

func (m mockI18n) Translator(namespace string) i18n.Translator {
	return mockTranslator{}
}

func (m mockI18n) RegisterFilesFromFS(fsPrefix string, rootFS embed.FS) error {
	return nil
}

type mockClickhouse struct {
	mchRow *mockClickhouseRow
	sql    string
}

type mockClickhouseRow struct {
	column []string
	err    error
	data   []ckMeta
	point  int
}

func (m *mockClickhouseRow) Next() bool {
	if m.point < len(m.data) {
		m.point = m.point + 1
		return true
	}
	return false
}

func (m *mockClickhouseRow) Scan(dest ...interface{}) error {
	return nil
}

func (m *mockClickhouseRow) ScanStruct(dest interface{}) error {
	meta := dest.(*ckMeta)
	data := m.data[m.point-1]
	meta.MetricGroup = data.MetricGroup
	meta.TagKeys = data.TagKeys
	meta.NumberKeys = data.NumberKeys
	meta.StringKeys = data.StringKeys
	return nil
}

func (m *mockClickhouseRow) ColumnTypes() []driver.ColumnType {
	return nil
}

func (m *mockClickhouseRow) Totals(dest ...interface{}) error {
	return nil
}

func (m *mockClickhouseRow) Columns() []string {
	return m.column
}

func (m mockClickhouseRow) Close() error {
	return nil
}

func (m *mockClickhouseRow) Err() error {
	return m.err
}

func (m *mockClickhouse) QueryRaw(orgName string, expr *goqu.SelectDataset) (driver.Rows, error) {
	if expr != nil {
		m.sql, _, _ = expr.ToSQL()
	}
	return m.mchRow, nil
}

func TestMock(t *testing.T) {
	row := mockClickhouseRow{}
	row.data = []ckMeta{
		ckMeta{
			MetricGroup: "",
			StringKeys:  []string{"11", "22", "33"},
			NumberKeys:  []string{"44", "55", "66"},
			TagKeys:     []string{"77", "88", "99"},
		},
		ckMeta{
			MetricGroup: "",
			StringKeys:  []string{"11", "22", "33"},
			NumberKeys:  []string{"44", "55", "66"},
			TagKeys:     []string{"77", "88", "99"},
		},
		ckMeta{
			MetricGroup: "",
			StringKeys:  []string{"11", "22", "33"},
			NumberKeys:  []string{"44", "55", "66"},
			TagKeys:     []string{"77", "88", "99"},
		},
	}
	mch := mockClickhouse{
		mchRow: &row,
	}
	resultRow, err := mch.QueryRaw("", nil)
	require.NoError(t, err)
	var cms []ckMeta

	for resultRow.Next() {
		var cm ckMeta
		err := resultRow.ScanStruct(&cm)
		require.NoError(t, err)
		require.ElementsMatch(t, []string{"11", "22", "33"}, cm.StringKeys)
		require.ElementsMatch(t, []string{"44", "55", "66"}, cm.NumberKeys)
		require.ElementsMatch(t, []string{"77", "88", "99"}, cm.TagKeys)
		cms = append(cms, cm)
	}
	require.Equal(t, 3, len(cms))
}

func TestMetricMetaWantSQL(t *testing.T) {
	tests := []struct {
		name    string
		scope   string
		scopeId string
		names   []string
		want    string
	}{
		{
			name:  "scope",
			scope: "org",
			want:  "SELECT \"metric_group\", string_field_keys AS \"sk\", number_field_keys AS \"nk\", tag_keys AS \"tk\" WHERE ((\"org_name\" = 'org') AND (\"timestamp\" >= fromUnixTimestamp64Nano(cast(1658201469067491000,'Int64'))) AND (\"timestamp\" < fromUnixTimestamp64Nano(cast(1658806269067491000,'Int64')))) GROUP BY \"metric_group\"",
		},
		{
			name:    "scope,scopeid",
			scope:   "org",
			scopeId: "13123",
			want:    "SELECT \"metric_group\", string_field_keys AS \"sk\", number_field_keys AS \"nk\", tag_keys AS \"tk\" WHERE ((\"org_name\" = 'org') AND (\"tenant_id\" = '13123') AND (\"timestamp\" >= fromUnixTimestamp64Nano(cast(1658201469067491000,'Int64'))) AND (\"timestamp\" < fromUnixTimestamp64Nano(cast(1658806269067491000,'Int64')))) GROUP BY \"metric_group\"",
		},
		{
			name:  "scope,names",
			scope: "org",
			names: []string{"metric1", "metric2"},
			want:  "SELECT \"metric_group\", string_field_keys AS \"sk\", number_field_keys AS \"nk\", tag_keys AS \"tk\" WHERE ((\"org_name\" = 'org') AND (\"metric_group\" IN ('metric1', 'metric2')) AND (\"timestamp\" >= fromUnixTimestamp64Nano(cast(1658201469067491000,'Int64'))) AND (\"timestamp\" < fromUnixTimestamp64Nano(cast(1658806269067491000,'Int64')))) GROUP BY \"metric_group\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := MetaClickhouseGroupProvider{}
			mockClickhouse := &mockClickhouse{
				mchRow: &mockClickhouseRow{},
			}
			p.clickhouse = mockClickhouse

			now = func() time.Time {
				return time.Unix(0, 1658806269067491000)
			}

			_, err := p.MetricMeta(nil, nil, tt.scope, tt.scopeId, tt.names...)
			require.NoError(t, err)
			require.Equal(t, tt.want, mockClickhouse.sql)
		})
	}
}

func TestMetricMetaWantMeta(t *testing.T) {
	tests := []struct {
		name       string
		scope      string
		scopeId    string
		names      []string
		mockResult []ckMeta
		want       map[string]*metricpb.MetricMeta
	}{
		{
			name:       "nil meta",
			mockResult: nil,
			want:       map[string]*metricpb.MetricMeta{},
		},
		{
			name: "no meta",
			mockResult: []ckMeta{
				ckMeta{
					MetricGroup: "metric1",
				},
			},
			want: map[string]*metricpb.MetricMeta{
				"metric1": {
					Tags:   map[string]*metricpb.TagDefine{},
					Fields: map[string]*metricpb.FieldDefine{},
				},
			},
		},
		{
			name: "only tag",
			mockResult: []ckMeta{
				ckMeta{
					MetricGroup: "metric1",
					TagKeys:     []string{"tag", "tag1", "tag2"},
				},
			},
			want: map[string]*metricpb.MetricMeta{
				"metric1": {
					Tags: map[string]*metricpb.TagDefine{
						"tag": {
							Key:  "tag",
							Name: "tag",
						},
						"tag1": {
							Key:  "tag1",
							Name: "tag1",
						},
						"tag2": {
							Key:  "tag2",
							Name: "tag2",
						},
					},
				},
			},
		},
		{
			name: "tag,field",
			mockResult: []ckMeta{
				ckMeta{
					MetricGroup: "metric1",
					TagKeys:     []string{"tag", "tag1", "tag2"},
					StringKeys:  []string{"field", "field1", "field2"},
				},
			},
			want: map[string]*metricpb.MetricMeta{
				"metric1": {
					Tags: map[string]*metricpb.TagDefine{
						"tag": {
							Key:  "tag",
							Name: "tag",
						},
						"tag1": {
							Key:  "tag1",
							Name: "tag1",
						},
						"tag2": {
							Key:  "tag2",
							Name: "tag2",
						},
					},
					Fields: map[string]*metricpb.FieldDefine{
						"field": {
							Key:  "field",
							Name: "field",
							Type: "string",
						},
						"field1": {
							Key:  "field1",
							Name: "field1",
							Type: "string",
						},
						"field2": {
							Key:  "field2",
							Name: "field2",
							Type: "string",
						},
					},
				},
			},
		},
		{
			name: "string and number field",
			mockResult: []ckMeta{
				ckMeta{
					MetricGroup: "metric1",
					TagKeys:     []string{"tag", "tag1", "tag2"},
					StringKeys:  []string{"field", "field1", "field2"},
					NumberKeys:  []string{"field3", "field4", "field5"},
				},
			},
			want: map[string]*metricpb.MetricMeta{
				"metric1": {
					Tags: map[string]*metricpb.TagDefine{
						"tag": {
							Key:  "tag",
							Name: "tag",
						},
						"tag1": {
							Key:  "tag1",
							Name: "tag1",
						},
						"tag2": {
							Key:  "tag2",
							Name: "tag2",
						},
					},
					Fields: map[string]*metricpb.FieldDefine{
						"field": {
							Key:  "field",
							Name: "field",
							Type: "string",
						},
						"field1": {
							Key:  "field1",
							Name: "field1",
							Type: "string",
						},
						"field2": {
							Key:  "field2",
							Name: "field2",
							Type: "string",
						},
						"field3": {
							Key:  "field3",
							Name: "field3",
							Type: "number",
						},
						"field4": {
							Key:  "field4",
							Name: "field4",
							Type: "number",
						},
						"field5": {
							Key:  "field5",
							Name: "field5",
							Type: "number",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := MetaClickhouseGroupProvider{}
			mockClickhouse := &mockClickhouse{
				mchRow: &mockClickhouseRow{
					data: tt.mockResult,
				},
			}
			p.clickhouse = mockClickhouse
			i := mockI18n{}
			data, err := p.MetricMeta(nil, i, tt.scope, tt.scopeId, tt.names...)
			require.NoError(t, err)
			if tt.want == nil {
				require.Nil(t, data)
				return
			}

			for k, v := range tt.want {
				result, ok := data[k]
				require.True(t, ok)
				if v.Tags != nil {
					for _k, _v := range v.Tags {
						require.Contains(t, result.Tags, _k)
						require.Equal(t, _v, result.Tags[_k])
					}
				}

				if v.Labels != nil {
					for _k, _v := range v.Labels {
						require.Contains(t, result.Labels, _k)
						require.Equal(t, _v, result.Labels[_k])
					}
				}

				if v.Fields != nil {
					for _k, _v := range v.Fields {
						require.Contains(t, result.Fields, _k)
						require.Equal(t, _v, result.Fields[_k])
					}
				}
			}
		})
	}
}

func TestGroups(t *testing.T) {
	mockClickhouse := &mockClickhouse{
		mchRow: &mockClickhouseRow{},
	}
	p, err := NewMetaClickhouseGroupProvider(mockClickhouse)
	require.NoError(t, err)
	i := mockI18n{}
	ms := map[string]*metricpb.MetricMeta{
		"metric1": {
			Name: &metricpb.NameDefine{Key: "metric1", Name: "metric1"},
		},
	}

	group, err := p.Groups(nil, i.Translator(""), "", "", ms)
	require.NoError(t, err)
	require.NotNil(t, group)
	require.Equal(t, 1, len(group))
	require.Equal(t, "All Metrics", group[0].Name)
	require.Equal(t, "all", group[0].Id)

	require.Equal(t, 1, len(group[0].Children))
	require.Equal(t, "metric1", group[0].Children[0].Name)
	require.Equal(t, "all@metric1", group[0].Children[0].Id)
}

func TestMappingsByID(t *testing.T) {
	mockClickhouse := &mockClickhouse{
		mchRow: &mockClickhouseRow{},
	}

	p, err := NewMetaClickhouseGroupProvider(mockClickhouse)
	require.NoError(t, err)
	ms := map[string]*metricpb.MetricMeta{
		"metric1": {
			Name: &metricpb.NameDefine{Key: "metric1", Name: "metric1"},
		},
	}
	names := []string{"metric1"}

	t.Run("all", func(t *testing.T) {
		gmm, err := p.MappingsByID("all", "", "", names, ms)
		require.NoError(t, err)
		require.Equal(t, []*GroupMetricMap{
			{
				Name: "metric1",
			},
		}, gmm)
	})
	t.Run("no all", func(t *testing.T) {
		gmm, err := p.MappingsByID("metric", "", "", names, ms)
		require.NoError(t, err)
		require.Nil(t, gmm)
	})
}
