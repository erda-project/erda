// create main table
CREATE TABLE IF NOT EXISTS <database>.main ON CLUSTER '{cluster}'
(
    `k` String,
    `v` String,
    `timestamp` DateTime64(9, 'Asia/Shanghai')
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/replicated/main', '{replica}')
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (timestamp)
TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;

// create trees table
CREATE TABLE IF NOT EXISTS <database>.trees ON CLUSTER '{cluster}'
(
    `k` String,
    `v` String,
    `timestamp` DateTime64(9, 'Asia/Shanghai')
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/replicated/trees', '{replica}')
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (timestamp)
TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;

// create segments table
CREATE TABLE IF NOT EXISTS <database>.segments ON CLUSTER '{cluster}'
(
    `k` String,
    `v` String,
    `timestamp` DateTime64(9, 'Asia/Shanghai')
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/replicated/segments', '{replica}')
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (timestamp)
TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;

// create dimensions table
CREATE TABLE IF NOT EXISTS <database>.dimensions ON CLUSTER '{cluster}'
(
    `k` String,
    `v` String,
    `timestamp` DateTime64(9, 'Asia/Shanghai')
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/replicated/dimensions', '{replica}')
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (timestamp)
TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;

// create profiles table
CREATE TABLE IF NOT EXISTS <database>.profiles ON CLUSTER '{cluster}'
(
    `k` String,
    `v` String,
    `timestamp` DateTime64(9, 'Asia/Shanghai')
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/replicated/profiles', '{replica}')
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (timestamp)
TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;

// create dicts table
CREATE TABLE IF NOT EXISTS <database>.dicts ON CLUSTER '{cluster}'
(
    `k` String,
    `v` String,
    `timestamp` DateTime64(9, 'Asia/Shanghai')
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/replicated/dicts', '{replica}')
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (timestamp)
TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;