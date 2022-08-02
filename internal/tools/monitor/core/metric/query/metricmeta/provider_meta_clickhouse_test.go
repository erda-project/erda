package metricmeta

import (
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/require"
)

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
			want:  "SELECT \"metric_group\", groupUniqArray(arrayJoin(string_field_keys)) AS \"sk\", groupUniqArray(arrayJoin(number_field_keys)) AS \"nk\", groupUniqArray(arrayJoin(tag_keys)) AS \"tk\" WHERE ((\"org_name\" = 'org') AND (\"timestamp\" >= fromUnixTimestamp64Nano(cast(1658201469067491000,'Int64'))) AND (\"timestamp\" < fromUnixTimestamp64Nano(cast(1658806269067491000,'Int64')))) GROUP BY \"metric_group\"",
		},
		{
			name:    "scope,scopeid",
			scope:   "org",
			scopeId: "13123",
			want:    "SELECT \"metric_group\", groupUniqArray(arrayJoin(string_field_keys)) AS \"sk\", groupUniqArray(arrayJoin(number_field_keys)) AS \"nk\", groupUniqArray(arrayJoin(tag_keys)) AS \"tk\" WHERE ((\"org_name\" = 'org') AND (\"tenant_id\" = '13123') AND (\"timestamp\" >= fromUnixTimestamp64Nano(cast(1658201469067491000,'Int64'))) AND (\"timestamp\" < fromUnixTimestamp64Nano(cast(1658806269067491000,'Int64')))) GROUP BY \"metric_group\"",
		},
		{
			name:  "scope,names",
			scope: "org",
			names: []string{"metric1", "metric2"},
			want:  "SELECT \"metric_group\", groupUniqArray(arrayJoin(string_field_keys)) AS \"sk\", groupUniqArray(arrayJoin(number_field_keys)) AS \"nk\", groupUniqArray(arrayJoin(tag_keys)) AS \"tk\" WHERE ((\"org_name\" = 'org') AND (\"metric_group\" IN ('metric1', 'metric2')) AND (\"timestamp\" >= fromUnixTimestamp64Nano(cast(1658201469067491000,'Int64'))) AND (\"timestamp\" < fromUnixTimestamp64Nano(cast(1658806269067491000,'Int64')))) GROUP BY \"metric_group\"",
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
			require.Equal(t, mockClickhouse.sql, tt.want)
		})
	}
}
