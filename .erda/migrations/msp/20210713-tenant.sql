CREATE TABLE IF NOT EXISTS `msp_tenant`
(
    `id`                 varchar(100) NOT NULL COMMENT '域id',
    `type`               varchar(10) DEFAULT NULL COMMENT '类型（dop 、msp）',
    `related_project_id` bigint(20)  DEFAULT NULL COMMENT '项目id',
    `related_workspace`  varchar(10) DEFAULT NULL COMMENT '环境（ DEV、TEST、STAGING、PROD、DEFAULT）',
    `create_time`        datetime    DEFAULT NULL COMMENT '创建时间',
    `update_time`        datetime    DEFAULT NULL COMMENT '更新时间',
    `is_deleted`         int(1)      DEFAULT NULL COMMENT '是否删除',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;