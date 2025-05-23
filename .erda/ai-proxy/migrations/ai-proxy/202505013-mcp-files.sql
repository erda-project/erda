CREATE TABLE `ai_proxy_mcp_files`
(
    `id`           CHAR(36)                           NOT NULL primary key COMMENT 'primary key',
    created_at     datetime default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at     datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    deleted_at     datetime null comment '删除时间',

    `storage_type` varchar(32)                        not null comment '存储类型，如：oss, s3',
    `bucket`       varchar(128)                       not null comment '存储桶名称',
    `object_key`   varchar(512)                       not null comment '存储对象 key',
    `file_name`    varchar(512)                       not null comment '文件名称',
    `file_size`    bigint                             not null comment '文件大小，单位：字节',
    `file_md5`     varchar(32)                        not null comment '文件 md5 值',
    `version_id`   varchar(128)                       not null comment '文件版本id',
    `keep`         char(1)  default 'N' comment '是否保留文件，Y-保留，N-不保留',
    `e_tag`        varchar(512)                       not null comment '文件 eTag 值',
    `is_deleted`   char(1)  default 'N' comment '是否删除，Y-删除，N-不删除'
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT 'mcp工具文件表';