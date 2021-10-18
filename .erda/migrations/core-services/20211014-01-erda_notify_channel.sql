CREATE TABLE `erda_notify_channel`
(
    `id`           varchar(100) NOT NULL COMMENT 'id',
    `name`         varchar(50)  NOT NULL COMMENT '渠道名称',
    `type`         varchar(20)  NOT NULL COMMENT '渠道类型',
    `config`       mediumtext   NOT NULL COMMENT '渠道配置',
    `scope_type`   varchar(20)  NOT NULL COMMENT '域类型',
    `scope_id`     varchar(20)  NOT NULL COMMENT '域id',
    `creator_id`   varchar(100) NOT NULL COMMENT '创建人id',
    `creator_name` varchar(100) NOT NULL COMMENT '创建人姓名',
    `create_at`    datetime     NOT NULL COMMENT '创建时间',
    `updated_at`   datetime     NOT NULL COMMENT '更新时间',
    `is_deleted`   tinyint(1) NOT NULL COMMENT '是否删除',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;