// create entity table
CREATE TABLE IF NOT EXISTS <database>.entities ON CLUSTER '{cluster}'
(
    `timestamp` DateTime64(9, 'Asia/Shanghai'),
    `update_timestamp` DateTime64(9, 'Asia/Shanghai'),
    `id` String,
    `type` LowCardinality(String),
    `key` String,
    `values` Map(String,String),
    `labels` Map(String,String),
    `labels.terminusKey` LowCardinality(String) MATERIALIZED labels['terminusKey'],

    INDEX idx_type(type) TYPE bloom_filter GRANULARITY 1,
    INDEX idx_key(key) TYPE bloom_filter GRANULARITY 1,
    INDEX idx_terminus_key(labels.terminusKey) TYPE bloom_filter GRANULARITY 1
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{cluster}--{shard}/entities', '{replica}')
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (timestamp, type, key)
TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;

// create distributed table
CREATE TABLE IF NOT EXISTS <database>.entities_all ON CLUSTER '{cluster}'
AS <database>.entities
    ENGINE = Distributed('{cluster}', <database>, entities, rand());