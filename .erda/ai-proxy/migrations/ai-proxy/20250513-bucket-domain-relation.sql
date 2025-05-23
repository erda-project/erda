CREATE TABLE `ai_proxy_bucket_domain_relation`
(
    `id`           CHAR(36)                           NOT NULL primary key COMMENT 'primary key',
    created_at     datetime default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at     datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    deleted_at     datetime null comment '删除时间',

    `bucket_name`  varchar(64)                        not null comment 'bucket name',
    `domain`       varchar(1024)                      not null comment 'bucket domain',
    `area`         varchar(128)                       not null comment 'bucket area',
    `storage_type` varchar(32)                        not null comment '存储类型，如：oss, s3'
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT 'bucket和域名对应关系表';