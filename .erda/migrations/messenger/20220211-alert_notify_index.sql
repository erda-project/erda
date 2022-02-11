CREATE TABLE `alert_notify_index`
(
    `id`          bigint(20)   NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `notify_id`   bigint(20)   NOT NULL COMMENT '关联通知的notify_id',
    `notify_name` varchar(150) NOT NULL COMMENT '告警列表中的标题',
    `status`      varchar(150) NOT NULL COMMENT '通知发送状态',
    `channel`     varchar(150) NOT NULL COMMENT '通知发送方式',
    `alert_id`    int(11)      NOT NULL COMMENT '告警策略关联id',
    `created_at`  timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `send_time`   timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发送时间',
    `scope_type`  varchar(150) NOT NULL,
    `scope_id`    varchar(150) NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE InnoDB DEFAULT CHARSET=utf8mb4 COMMENT = '告警通知索引表'
