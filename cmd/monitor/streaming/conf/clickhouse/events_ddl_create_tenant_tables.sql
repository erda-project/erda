// create event table
CREATE TABLE IF NOT EXISTS <database>.<table_name> ON CLUSTER '{cluster}'
(
    `timestamp` DateTime64(9, 'Asia/Shanghai'),
    `event_id` String,
    `content` String,
    `kind` LowCardinality(String),
    `relations` Map(String,String),
    `tags` Map(String,String),

    `relations.res_id` String MATERIALIZED relations['res_id'],
    `relations.res_type` String MATERIALIZED relations['res_type'],
    `relations.trace_id` String MATERIALIZED relations['trace_id']
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{cluster}-{shard}/events', '{replica}')
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (timestamp)
TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;

// create distributed table
CREATE TABLE IF NOT EXISTS <database>.<table_name>_all ON CLUSTER '{cluster}'
AS <database>.events
    ENGINE = Distributed('{cluster}', <database>, events, rand());

CREATE TABLE IF NOT EXISTS <database>.<alias_table_name>_search ON CLUSTER '{cluster}'
AS <database>.events
ENGINE = Merge(<database>, 'events_all|<alias_table_name>.*_all$');