// 创建日志表
CREATE TABLE IF NOT EXISTS <database>.logs ON CLUSTER '{cluster}'
(
    `_id` String,
    `timestamp` DateTime64(9,'Asia/Shanghai'),
    `source` String,
    `id` String,
    `org_name` String,
    `stream` String,
    `offset` Int64,
    `content` String,
    `tags` Map(String,String),
    INDEX idx__id(_id) TYPE minmax GRANULARITY 1
    )
    ENGINE = ReplicatedMergeTree('/clickhouse/tables/{cluster}-{shard}/logs', '{replica}')
    PARTITION BY toYYYYMMDD(timestamp)
    ORDER BY (org_name, timestamp, id)
    TTL toDateTime(timestamp) + INTERVAL 7 DAY;

// 将常用字段添加为物化列
ALTER TABLE <database>.logs ON CLUSTER '{cluster}'
    ADD COLUMN IF NOT EXISTS `tags.trace_id` String MATERIALIZED tags['trace_id'],
    ADD COLUMN IF NOT EXISTS `tags.level` String MATERIALIZED tags['level'],
    ADD COLUMN IF NOT EXISTS `tags.application_name` String MATERIALIZED tags['application_name'],
    ADD COLUMN IF NOT EXISTS `tags.service_name` String MATERIALIZED tags['service_name'],
    ADD COLUMN IF NOT EXISTS `tags.pod_name` String MATERIALIZED tags['pod_name'],
    ADD COLUMN IF NOT EXISTS `tags.pod_ip` String MATERIALIZED tags['pod_ip'],
    ADD COLUMN IF NOT EXISTS `tags.container_name` String MATERIALIZED tags['container_name'],
    ADD COLUMN IF NOT EXISTS `tags.container_id` String MATERIALIZED tags['container_id'];

// 对常用字段添加索引
ALTER TABLE <database>.logs ON CLUSTER '{cluster}' ADD INDEX IF NOT EXISTS idx_tace_id(tags.trace_id) TYPE bloom_filter GRANULARITY 1;
ALTER TABLE <database>.logs ON CLUSTER '{cluster}' ADD INDEX IF NOT EXISTS idx_id(id) TYPE bloom_filter GRANULARITY 1;

// 创建分布式表
// 注意: 如果对logs表结构新增列, 需要同步修改logs_all
CREATE TABLE IF NOT EXISTS <database>.logs_all ON CLUSTER '{cluster}'
AS <database>.logs
    ENGINE = Distributed('{cluster}', <database>, logs, rand());