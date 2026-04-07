CREATE TABLE IF NOT EXISTS `ai_proxy_setting`
(
    `id`         CHAR(36)     NOT NULL COMMENT 'primary key',
    `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `namespace`  VARCHAR(191) NOT NULL COMMENT '配置命名空间',
    `key`        VARCHAR(191) NOT NULL COMMENT '配置 key',
    `value`      TEXT         NOT NULL COMMENT '配置值',

    PRIMARY KEY (`id`),
    UNIQUE INDEX `uk_namespace_key` (`namespace`, `key`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT 'AI-Proxy 设置表';

INSERT IGNORE INTO `ai_proxy_setting` (`id`, `created_at`, `updated_at`, `deleted_at`, `namespace`, `key`, `value`)
VALUES (UUID(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '1970-01-01 00:00:00', 'blacklist_user_agent', 'client_token.blacklist', 'openclaw,coding-agent');

INSERT IGNORE INTO `ai_proxy_setting` (`id`, `created_at`, `updated_at`, `deleted_at`, `namespace`, `key`, `value`)
VALUES (UUID(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '1970-01-01 00:00:00', 'blacklist_user_agent', 'client.blacklist', 'openclaw,coding-agent');

INSERT IGNORE INTO `ai_proxy_setting` (`id`, `created_at`, `updated_at`, `deleted_at`, `namespace`, `key`, `value`)
VALUES (UUID(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '1970-01-01 00:00:00', 'blacklist_user_agent', 'general.headers', '');

INSERT IGNORE INTO `ai_proxy_setting` (`id`, `created_at`, `updated_at`, `deleted_at`, `namespace`, `key`, `value`)
VALUES (UUID(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '1970-01-01 00:00:00', 'blacklist_user_agent', 'general.prompts', '');
