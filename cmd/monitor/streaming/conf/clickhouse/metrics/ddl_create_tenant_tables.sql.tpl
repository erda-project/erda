CREATE TABLE IF NOT EXISTS <database>.<table_name> ON CLUSTER '{cluster}'
(
    `org_name`            LowCardinality(String),
    `tenant_id`           LowCardinality(String),
    `metric_group`        LowCardinality(String),
    `timestamp`           DateTime64(9,'Asia/Shanghai') CODEC (DoubleDelta),
    `number_field_keys`   Array(LowCardinality(String)),
    `number_field_values` Array(Float64),
    `string_field_keys`   Array(LowCardinality(String)),
    `string_field_values` Array(String),
    `tag_keys`            Array(LowCardinality(String)),
    `tag_values`          Array(LowCardinality(String))
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{cluster}-{shard}/{database}/<table_name>', '{replica}')
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (org_name, metric_group, timestamp)
TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;

// create distributed table
// notice: ddls to the <table> table should be synced to the <table>_all table
CREATE TABLE IF NOT EXISTS <database>.<table_name>_all ON CLUSTER '{cluster}' AS <database>.<table_name>
ENGINE = Distributed('{cluster}', <database>, <table_name>, rand());

// create merge table
CREATE TABLE IF NOT EXISTS <database>.<alias_table_name>_search ON CLUSTER '{cluster}' AS <database>.metrics
ENGINE = Merge(<database>, 'metrics_all|<alias_table_name>.*_all$');