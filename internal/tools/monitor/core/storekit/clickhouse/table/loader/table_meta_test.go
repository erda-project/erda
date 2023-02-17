package loader

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractTTL(t *testing.T) {
	meta := TableMeta{}
	meta.CreateTableSQL = "aaa TTL toDateTime(end_time) + toIntervalDay(1) TO VOLUME 'slow', toDateTime(end_time) + toIntervalDay(7) SETTINGS index_granularity = 8192 asdada"
	meta.extractTTLDays()

	tests := []struct {
		name           string
		createSQL      string
		wantTimeKey    string
		wantTTLDays    int64
		wantHotDDLDays int64
	}{
		{
			name:           "ttl + hot ttl",
			createSQL:      "TTL toDateTime(timestamp) + toIntervalDay(1) TO VOLUME 'slow', toDateTime(timestamp) + toIntervalDay(7) SETTINGS",
			wantTTLDays:    7,
			wantTimeKey:    "toDateTime(timestamp)",
			wantHotDDLDays: 1,
		},
		{
			name:           "ttl + hot ttl,ttl time key not equal ttl time key",
			createSQL:      "TTL toDateTime(end_time) + toIntervalDay(1) TO VOLUME 'slow', toDateTime(timestamp) + toIntervalDay(7) SETTINGS",
			wantTTLDays:    7,
			wantTimeKey:    "toDateTime(timestamp)",
			wantHotDDLDays: 1,
		},
		{
			name:           "only ttl",
			createSQL:      "TTL toDateTime(timestamp) + toIntervalDay(7) SETTINGS",
			wantTTLDays:    7,
			wantTimeKey:    "toDateTime(timestamp)",
			wantHotDDLDays: 0,
		},
		{
			name:           "none",
			createSQL:      "ENGINE = ReplicatedMergeTree('/clickhouse/tables/{cluster}-{shard}/monitor/metrics', '{replica}') PARTITION BY toYYYYMMDD(timestamp) ORDER BY (org_name, tenant_id, metric_group, timestamp)",
			wantTTLDays:    0,
			wantTimeKey:    "",
			wantHotDDLDays: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			meta := &TableMeta{}
			meta.CreateTableSQL = test.createSQL
			meta.extractTTLDays()
			require.Equal(t, test.wantTimeKey, meta.TimeKey)
			require.Equal(t, test.wantTTLDays, meta.TTLDays)
			require.Equal(t, test.wantHotDDLDays, meta.HotTTLDays)
		})
	}

}

func TestGetStringInBetween(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{
			text: "aaa TTL toDateTime(end_time) + toIntervalDay(1) TO VOLUME 'slow', toDateTime(end_time) + toIntervalDay(7) SETTINGS index_granularity = 8192 1231312",
			want: "toDateTime(end_time) + toIntervalDay(1) TO VOLUME 'slow', toDateTime(end_time) + toIntervalDay(7)",
		},
		{
			text: "",
			want: "",
		},
		{
			text: "TTL ",
			want: "",
		},
		{
			text: " SETTINGS",
			want: "",
		},
		{
			text: " SETTINGS",
			want: "",
		},
		{
			text: "CREATE TABLE monitor.metrics (     `org_name` LowCardinality(String),     `tenant_id` LowCardinality(String),     `metric_group` LowCardinality(String),     `timestamp` DateTime64(9, 'Asia/Shanghai') CODEC(DoubleDelta),     `number_field_keys` Array(LowCardinality(String)),     `number_field_values` Array(Float64),     `string_field_keys` Array(LowCardinality(String)),     `string_field_values` Array(String),     `tag_keys` Array(LowCardinality(String)),     `tag_values` Array(LowCardinality(String)),     INDEX idx_metric_service_id tag_values[indexOf(tag_keys, 'service_id')] TYPE bloom_filter GRANULARITY 1,     INDEX idx_metric_source_service_id tag_values[indexOf(tag_keys, 'source_service_id')] TYPE bloom_filter GRANULARITY 1,     INDEX idx_metric_target_service_id tag_values[indexOf(tag_keys, 'target_service_id')] TYPE bloom_filter GRANULARITY 1,     INDEX idx_metric_cluster_name tag_values[indexOf(tag_keys, 'cluster_name')] TYPE bloom_filter GRANULARITY 1,     INDEX idx_container_id tag_values[indexOf(tag_keys, 'container_id')] TYPE bloom_filter GRANULARITY 2,     INDEX idx_pod_name tag_values[indexOf(tag_keys, 'pod_name')] TYPE bloom_filter GRANULARITY 2,     INDEX idx_pod_uid tag_values[indexOf(tag_keys, 'pod_uid')] TYPE bloom_filter GRANULARITY 2,     INDEX idx_metric tag_values[indexOf(tag_keys, 'metric')] TYPE bloom_filter GRANULARITY 1,     INDEX idx_family_id tag_values[indexOf(tag_keys, 'family_id')] TYPE bloom_filter GRANULARITY 3,     INDEX idx_terminus_key tag_values[indexOf(tag_keys, 'terminus_key')] TYPE bloom_filter GRANULARITY 1 ) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{cluster}-{shard}/monitor/metrics', '{replica}') PARTITION BY toYYYYMMDD(timestamp) ORDER BY (org_name, tenant_id, metric_group, timestamp) TTL toDateTime(timestamp) + toIntervalHour(1) TO VOLUME 'slow', toDateTime(timestamp) + toIntervalDay(7) SETTINGS index_granularity = 8192",
			want: "toDateTime(timestamp) + toIntervalHour(1) TO VOLUME 'slow', toDateTime(timestamp) + toIntervalDay(7)",
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			got := GetStringInBetween(test.text, "TTL", "SETTINGS")
			require.Equalf(t, test.want, got, "GetStringInBetween(%s) = %s, want %s", test.text, got, test.want)
		})
	}
}
