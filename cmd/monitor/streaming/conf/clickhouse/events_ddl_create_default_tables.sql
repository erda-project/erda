// create event table
CREATE TABLE IF NOT EXISTS <database>.events ON CLUSTER '{cluster}'
(
    `timestamp` DateTime64(9, 'Asia/Shanghai'),
    `event_id` String,
    `content` String,
    `kind` LowCardinality(String),
    `tags` Map(String,String),

    `tags.display_url` String MATERIALIZED tags['display_url'],
    `tags.alert_title` String MATERIALIZED tags['alert_title'],
    `tags.trigger` String MATERIALIZED tags['trigger'],
    `tags.org_name` String MATERIALIZED tags['org_name'],
    `tags.dice_org_id` String MATERIALIZED tags['dice_org_id'],

    `relations` Nested (
                           `trace_id` String,
                           `res_id` String,
                           `res_type` LowCardinality(String)
    )
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{cluster}-{shard}/events', '{replica}')
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (timestamp)
TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;

// create distributed table
CREATE TABLE IF NOT EXISTS <database>.events_all ON CLUSTER '{cluster}'
AS <database>.events
    ENGINE = Distributed('{cluster}', <database>, events, rand());