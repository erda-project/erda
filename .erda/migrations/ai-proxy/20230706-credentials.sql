CREATE TABLE `ai_proxy_credentials`
(
    `id`            CHAR(36)     NOT NULL COMMENT 'primary key',
    `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`    DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `access_key_id` CHAR(36)     NOT NULL COMMENT '平台接入 AI 服务的 AK',
    `secret_key_id` CHAR(36)     NOT NULL COMMENT '平台接入 AI 服务的 SK',
    `name`          VARCHAR(64)  NOT NULL COMMENT '凭证名称',
    `platform`      VARCHAR(128) NOT NULL COMMENT '接入 AI 服务的平台',
    `description`   VARCHAR(512) NOT NULL COMMENT '凭证描述',
    `enabled`       BOOLEAN      NOT NULL DEFAULT true COMMENT '是否启用该凭证',
    `expired_at`    DATETIME     NOT NULL DEFAULT '2099-01-01 00:00:00' COMMENT '凭证过期时间',

    PRIMARY KEY (`id`),
    INDEX `idx_access_key_id` (`access_key_id`),
    INDEX `idx_name` (`name`),
    INDEX `idx_platform` (`platform`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
    COMMENT 'ai-proxy 凭证';
