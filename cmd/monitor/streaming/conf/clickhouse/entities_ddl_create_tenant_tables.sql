// create entity table
CREATE TABLE IF NOT EXISTS <database>.<table_name> ON CLUSTER '{cluster}'
(
    `timestamp` DateTime64(9, 'Asia/Shanghai'),
    `update_timestamp` DateTime64(9, 'Asia/Shanghai'),
    `id` String,
    `type` LowCardinality(String),
    `key` String,
    `values` Map(String,String),
    `labels` Map(String,String)
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{cluster}--{shard}/entities', '{replica}')
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (timestamp)
TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;

// create distributed table
CREATE TABLE IF NOT EXISTS <database>.<table_name>_all ON CLUSTER '{cluster}'
AS <database>.entities
    ENGINE = Distributed('{cluster}', <database>, entities, rand());

CREATE TABLE IF NOT EXISTS <database>.<alias_table_name>_search ON CLUSTER '{cluster}'
AS <database>.entities
ENGINE = Merge(<database>, 'entities_all|<alias_table_name>.*_all$');