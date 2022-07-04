CREATE TABLE IF NOT EXISTS <database>.spans ON CLUSTER '{cluster}'
(
    `org_name`       LowCardinality(String),
    `tenant_id`      LowCardinality(String),
    `trace_id`       String,
    `span_id`        String,
    `parent_span_id` String,
    `operation_name` LowCardinality(String),
    `start_time`     DateTime64(9,'Asia/Shanghai') CODEC (DoubleDelta),
    `end_time`       DateTime64(9,'Asia/Shanghai') CODEC (DoubleDelta),
    `tag_keys`       Array(LowCardinality(String)),
    `tag_values`     Array(LowCardinality(String)),
    INDEX idx_trace_id(trace_id) TYPE bloom_filter GRANULARITY 1
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{cluster}-{shard}/{database}/spans', '{replica}')
PARTITION BY toYYYYMMDD(end_time)
ORDER BY (org_name, tenant_id, start_time, end_time)
TTL toDateTime(end_time) + INTERVAL <ttl_in_days> DAY;

// create distributed table
// notice: ddls to the <table> table should be synced to the <table>_all table
CREATE TABLE IF NOT EXISTS <database>.spans_all ON CLUSTER '{cluster}' AS <database>.spans
ENGINE = Distributed('{cluster}', <database>, spans, rand());