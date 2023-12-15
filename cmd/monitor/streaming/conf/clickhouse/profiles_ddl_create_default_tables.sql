// create main table
CREATE TABLE IF NOT EXISTS <database>.main ON CLUSTER '{cluster}'
(
    `k` String,
    `v` String,
    `timestamp` DateTime64(9, 'Asia/Shanghai')
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{cluster}-{shard}/main', '{replica}')
PRIMARY KEY (k)
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (k)
TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;

// create main distributed table
CREATE TABLE IF NOT EXISTS <database>.main_all ON CLUSTER '{cluster}'
AS <database>.main
    ENGINE = Distributed('{cluster}', <database>, main, rand());

// create trees table
CREATE TABLE IF NOT EXISTS <database>.trees ON CLUSTER '{cluster}'
(
    `k` String,
    `v` String,
    `timestamp` DateTime64(9, 'Asia/Shanghai')
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{cluster}-{shard}/trees', '{replica}')
PRIMARY KEY (k)
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (k)
TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;

// create trees distributed table
CREATE TABLE IF NOT EXISTS <database>.trees_all ON CLUSTER '{cluster}'
AS <database>.trees
    ENGINE = Distributed('{cluster}', <database>, trees, rand());

// create segments table
CREATE TABLE IF NOT EXISTS <database>.segments ON CLUSTER '{cluster}'
(
    `k` String,
    `v` String,
    `timestamp` DateTime64(9, 'Asia/Shanghai')
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{cluster}-{shard}/segments', '{replica}')
PRIMARY KEY (k)
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (k)
TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;

// create segments distributed table
CREATE TABLE IF NOT EXISTS <database>.segments_all ON CLUSTER '{cluster}'
AS <database>.segments
    ENGINE = Distributed('{cluster}', <database>, segments, rand());

// create dimensions table
CREATE TABLE IF NOT EXISTS <database>.dimensions ON CLUSTER '{cluster}'
(
    `k` String,
    `v` String,
    `timestamp` DateTime64(9, 'Asia/Shanghai')
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{cluster}-{shard}/dimensions', '{replica}')
PRIMARY KEY (k)
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (k)
TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;

// create dimensions distributed table
CREATE TABLE IF NOT EXISTS <database>.dimensions_all ON CLUSTER '{cluster}'
AS <database>.dimensions
    ENGINE = Distributed('{cluster}', <database>, dimensions, rand());

// create profiles table
CREATE TABLE IF NOT EXISTS <database>.profiles ON CLUSTER '{cluster}'
(
    `k` String,
    `v` String,
    `timestamp` DateTime64(9, 'Asia/Shanghai')
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{cluster}-{shard}/profiles', '{replica}')
PRIMARY KEY (k)
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (k)
TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;

// create profiles distributed table
CREATE TABLE IF NOT EXISTS <database>.profiles_all ON CLUSTER '{cluster}'
AS <database>.profiles
    ENGINE = Distributed('{cluster}', <database>, profiles, rand());

// create dicts table
CREATE TABLE IF NOT EXISTS <database>.dicts ON CLUSTER '{cluster}'
(
    `k` String,
    `v` String,
    `timestamp` DateTime64(9, 'Asia/Shanghai')
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{cluster}-{shard}/dicts', '{replica}')
PRIMARY KEY (k)
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (k)
TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;

// create dicts distributed table
CREATE TABLE IF NOT EXISTS <database>.dicts_all ON CLUSTER '{cluster}'
AS <database>.dicts
    ENGINE = Distributed('{cluster}', <database>, dicts, rand());