CREATE TABLE `ai_proxy_bucket_domain_relation`
(
    `id`           CHAR(36)                           NOT NULL primary key COMMENT 'primary key',
    `created_at`   DATETIME default CURRENT_TIMESTAMP not null comment '创建时间',
    `updated_at`   DATETIME default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    `deleted_at`   DATETIME                           NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `bucket_name`  varchar(64)                        not null comment 'bucket name',
    `domain`       varchar(1024)                      not null comment 'bucket domain',
    `area`         varchar(128)                       not null comment 'bucket area',
    `storage_type` varchar(32)                        not null comment '存储类型，如：oss, s3'
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT 'bucket和域名对应关系表';