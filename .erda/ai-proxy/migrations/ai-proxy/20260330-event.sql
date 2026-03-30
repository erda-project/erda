CREATE TABLE IF NOT EXISTS `ai_proxy_event`
(
    `id`         BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'primary key',
    `created_at` DATETIME(3)         NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` DATETIME(3)         NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',

    `event`      VARCHAR(191)        NOT NULL COMMENT '事件类型',
    `detail`     VARCHAR(255)        NOT NULL DEFAULT '' COMMENT '事件详情',

    PRIMARY KEY (`id`),
    INDEX `idx_event_created_at` (`event`, `created_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT 'AI-Proxy 事件表';
