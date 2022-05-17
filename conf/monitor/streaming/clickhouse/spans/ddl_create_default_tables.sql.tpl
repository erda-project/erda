CREATE TABLE IF NOT EXISTS <database>.spans_series ON CLUSTER '{cluster}'
(
  `org_name` LowCardinality(String),
  `series_id` UInt64,
  `trace_id` String,
  `span_id` String,
	`parent_span_id` String,
  `start_time` DateTime64(9,'Asia/Shanghai') CODEC(DoubleDelta),
  `end_time` DateTime64(9,'Asia/Shanghai') CODEC(DoubleDelta),
  INDEX idx_trace_id(trace_id) TYPE minmax GRANULARITY 1
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{cluster}-{shard}/{database}/spans_series', '{replica}')
PARTITION BY toYYYYMMDD(end_time)
ORDER BY (org_name, series_id, end_time)
TTL toDateTime(end_time) + INTERVAL 7 DAY;


CREATE TABLE IF NOT EXISTS <database>.spans_meta ON CLUSTER '{cluster}'
(
  `org_name` LowCardinality(String),
  `series_id` UInt64,
  `key` LowCardinality(String),
  `value` String,
  `create_at` DateTime64(9,'Asia/Shanghai')
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{cluster}-{shard}/{database}/spans_meta', '{replica}')
PARTITION BY toYYYYMM(create_at)
ORDER BY (org_name, series_id, key, value)
TTL toDateTime(create_at) + INTERVAL 14 DAY;